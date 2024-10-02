package chain_pusher

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type StorkContractInterfacer struct {
	logger zerolog.Logger

	contract   *StorkContract
	privateKey *ecdsa.PrivateKey
	chainID    *big.Int

	pollingFrequencySec int
	verifyPublishers    bool
}

func NewStorkContractInterfacer(rpcUrl, contractAddr, mnemonicFile string, pollingFreqSec int, verifyPublishers bool, logger zerolog.Logger) *StorkContractInterfacer {
	logger.With().Str("component", "stork-contract-interfacer").Logger()

	privateKey, err := loadPrivateKey(mnemonicFile)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to load private key")
	}

	client, err := ethclient.Dial(rpcUrl)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to connect to Ethereum client")
	}

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to get chain ID")
	}

	contractAddress := common.HexToAddress(contractAddr)
	contract, err := NewStorkContract(contractAddress, client)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize contract")
	}

	return &StorkContractInterfacer{
		logger: logger,

		contract:   contract,
		privateKey: privateKey,
		chainID:    chainID,

		pollingFrequencySec: pollingFreqSec,
		verifyPublishers:    verifyPublishers,
	}
}

func (sci *StorkContractInterfacer) ListenContractEvents(ch chan map[InternalEncodedAssetId]StorkStructsTemporalNumericValue) {
	watchOpts := &bind.WatchOpts{
		Context: context.Background(),
	}

	eventCh := make(chan *StorkContractValueUpdate)
	sub, err := sci.contract.WatchValueUpdate(watchOpts, eventCh, nil)
	if err != nil {
		log.Error().Err(err).Msg("Failed to watch contract events. Is the RPC URL a WebSocket endpoint?")
		return
	}

	sci.logger.Info().Msg("Listening for contract events")
	for {
		select {
		case err := <-sub.Err():
			// TODO - handle restart
			log.Fatal().Err(err).Msg("Error watching contract events")
		case vLog := <-eventCh:
			tv := StorkStructsTemporalNumericValue{
				QuantizedValue: vLog.QuantizedValue,
				TimestampNs:    vLog.TimestampNs,
			}
			ch <- map[InternalEncodedAssetId]StorkStructsTemporalNumericValue{vLog.Id: tv}
		}
	}
}

func (sci *StorkContractInterfacer) Poll(encodedAssetIds []InternalEncodedAssetId, ch chan map[InternalEncodedAssetId]StorkStructsTemporalNumericValue) {
	sci.logger.Info().Msgf("Polling contract for new values for %d assets", len(encodedAssetIds))
	for _ = range time.Tick(time.Duration(sci.pollingFrequencySec) * time.Second) {
		polledVals, err := sci.PullValues(encodedAssetIds)
		if err != nil {
			sci.logger.Error().Err(err).Msg("Failed to poll contract")
			continue
		}
		if len(polledVals) > 0 {
			ch <- polledVals
		}
	}
}

func (sci *StorkContractInterfacer) PullValues(encodedAssetIds []InternalEncodedAssetId) (map[InternalEncodedAssetId]StorkStructsTemporalNumericValue, error) {
	polledVals := make(map[InternalEncodedAssetId]StorkStructsTemporalNumericValue)
	for _, encodedAssetId := range encodedAssetIds {
		storkStructsTemporalNumericValue, err := sci.contract.GetTemporalNumericValueV1(nil, encodedAssetId)
		if err != nil {
			if strings.Contains(err.Error(), "NotFound()") {
				sci.logger.Debug().Str("assetId", hex.EncodeToString(encodedAssetId[:])).Msg("No value found")
			} else {
				sci.logger.Debug().Str("assetId", hex.EncodeToString(encodedAssetId[:])).Msg("Failed to get latest value")
			}
			continue
		}
		polledVals[encodedAssetId] = storkStructsTemporalNumericValue
	}
	return polledVals, nil
}

func getUpdatePayload(priceUpdates map[InternalEncodedAssetId]AggregatedSignedPrice) ([]StorkStructsTemporalNumericValueInput, error) {
	updates := make([]StorkStructsTemporalNumericValueInput, len(priceUpdates))
	i := 0
	for _, priceUpdate := range priceUpdates {

		quantizedPriceBigInt := new(big.Int)
		quantizedPriceBigInt.SetString(string(priceUpdate.StorkSignedPrice.QuantizedPrice), 10)

		encodedAssetId, err := stringToByte32(string(priceUpdate.StorkSignedPrice.EncodedAssetId))
		if err != nil {
			return nil, err
		}

		rBytes, err := stringToByte32(priceUpdate.StorkSignedPrice.TimestampedSignature.Signature.R)
		if err != nil {
			return nil, err
		}

		sBytes, err := stringToByte32(priceUpdate.StorkSignedPrice.TimestampedSignature.Signature.S)
		if err != nil {
			return nil, err
		}

		publisherMerkleRoot, err := stringToByte32(priceUpdate.StorkSignedPrice.PublisherMerkleRoot)
		if err != nil {
			return nil, err
		}

		checksum, err := stringToByte32(priceUpdate.StorkSignedPrice.StorkCalculationAlg.Checksum)
		if err != nil {
			return nil, err
		}

		vInt, err := strconv.ParseInt(priceUpdate.StorkSignedPrice.TimestampedSignature.Signature.V[2:], 16, 8)
		if err != nil {
			return nil, err
		}

		updates[i] = StorkStructsTemporalNumericValueInput{
			TemporalNumericValue: StorkStructsTemporalNumericValue{
				TimestampNs:    uint64(priceUpdate.StorkSignedPrice.TimestampedSignature.Timestamp),
				QuantizedValue: quantizedPriceBigInt,
			},
			Id:                  encodedAssetId,
			PublisherMerkleRoot: publisherMerkleRoot,
			ValueComputeAlgHash: checksum,
			R:                   rBytes,
			S:                   sBytes,
			V:                   uint8(vInt),
		}
		i++
	}

	return updates, nil
}

type VerifyPayload struct {
	pubSigs    []StorkStructsPublisherSignature
	merkleRoot [32]byte
}

func getVerifyPublishersPayloads(priceUpdates map[InternalEncodedAssetId]AggregatedSignedPrice) ([]VerifyPayload, error) {
	payloads := make([]VerifyPayload, len(priceUpdates))
	i := 0
	for _, priceUpdate := range priceUpdates {
		merkleRootBytes, err := stringToByte32(priceUpdate.StorkSignedPrice.PublisherMerkleRoot)
		if err != nil {
			return nil, err
		}

		payloads[i] = VerifyPayload{
			pubSigs:    make([]StorkStructsPublisherSignature, len(priceUpdate.SignedPrices)),
			merkleRoot: merkleRootBytes,
		}
		j := 0
		for _, signedPrice := range priceUpdate.SignedPrices {
			pubKeyBytes, err := stringToByte20(string(signedPrice.PublisherKey))
			if err != nil {
				return nil, err
			}

			quantizedPriceBigInt := new(big.Int)
			quantizedPriceBigInt.SetString(string(signedPrice.QuantizedPrice), 10)

			rBytes, err := stringToByte32(signedPrice.TimestampedSignature.Signature.R)
			if err != nil {
				return nil, err
			}

			sBytes, err := stringToByte32(signedPrice.TimestampedSignature.Signature.S)
			if err != nil {
				return nil, err
			}

			vInt, err := strconv.ParseInt(signedPrice.TimestampedSignature.Signature.V[2:], 16, 8)
			if err != nil {
				return nil, err
			}

			payloads[i].pubSigs[j] = StorkStructsPublisherSignature{
				PubKey:         pubKeyBytes,
				AssetPairId:    signedPrice.ExternalAssetId,
				Timestamp:      uint64(signedPrice.TimestampedSignature.Timestamp) / 1000000000,
				QuantizedValue: quantizedPriceBigInt,
				R:              rBytes,
				S:              sBytes,
				V:              uint8(vInt),
			}
			j++
		}
		i++
	}

	return payloads, nil
}

func (sci *StorkContractInterfacer) BatchPushToContract(priceUpdates map[InternalEncodedAssetId]AggregatedSignedPrice) error {
	if sci.verifyPublishers {
		publisherVerifyPayloads, err := getVerifyPublishersPayloads(priceUpdates)
		if err != nil {
			return err
		}
		for i, _ := range publisherVerifyPayloads {
			verified, err := sci.contract.VerifyPublisherSignaturesV1(nil, publisherVerifyPayloads[i].pubSigs, publisherVerifyPayloads[i].merkleRoot)
			if err != nil {
				sci.logger.Error().Err(err).Msg("Failed to verify publisher signatures")
				return err
			}
			if !verified {
				sci.logger.Error().Msg("Publisher signatures not verified, skipping update")
				return nil
			}
		}
	}

	updatePayload, err := getUpdatePayload(priceUpdates)
	if err != nil {
		return err
	}

	fee, err := sci.contract.GetUpdateFeeV1(nil, updatePayload)
	if err != nil {
		return err
	}

	auth, err := bind.NewKeyedTransactorWithChainID(sci.privateKey, sci.chainID)
	if err != nil {
		return err
	}

	// let the library auto-estimate the gas price
	auth.GasLimit = 0
	auth.Value = fee

	tx, err := sci.contract.UpdateTemporalNumericValuesV1(auth, updatePayload)
	if err != nil {
		return err
	}

	sci.logger.Info().
		Str("txHash", tx.Hash().Hex()).
		Int("numUpdates", len(updatePayload)).
		Uint64("gasPrice", tx.GasPrice().Uint64()).
		Msg("Pushed new values to contract")
	return nil
}
