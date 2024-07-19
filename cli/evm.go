package main

import (
	"context"
	"crypto/ecdsa"
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

type ContractConfig struct {
	RpcUrl         string
	ContractAddr   string
	MnemonicFile   string
	PollingFreqSec int
	GasPrice       *big.Int
}

type StorkContractInterfacer struct {
	logger zerolog.Logger

	contract   *StorkContract
	privateKey *ecdsa.PrivateKey
	chainID    *big.Int

	pollingFrequencySec int
}

func NewStorkContractInterfacer(rpcUrl, contractAddr, mnemonicFile string, pollingFreqSec int, logger zerolog.Logger) *StorkContractInterfacer {
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
	}
}

func (sci *StorkContractInterfacer) ListenContractEvents(ch chan map[InternalEncodedAssetId]StorkStructsTemporalNumericValue) {
	watchOpts := &bind.WatchOpts{
		Context: context.Background(),
	}

	eventCh := make(chan *StorkContractValueUpdate)
	sub, err := sci.contract.WatchValueUpdate(watchOpts, eventCh, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to watch contract events")
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
				sci.logger.Debug().Bytes("assetId", encodedAssetId[:]).Msg("Asset not found")
			} else {
				sci.logger.Error().Err(err).Bytes("assetId", encodedAssetId[:]).Msg("Failed to get latest value")
			}
			continue
		}
		polledVals[encodedAssetId] = storkStructsTemporalNumericValue
	}
	return polledVals, nil
}

func (sci *StorkContractInterfacer) BatchPushToContract(priceUpdates map[InternalEncodedAssetId]AggregatedSignedPrice) error {
	updates := make([]StorkStructsTemporalNumericValueInput, len(priceUpdates))
	i := 0
	for _, priceUpdate := range priceUpdates {

		quantizedPriceBigInt := new(big.Int)
		quantizedPriceBigInt.SetString(string(priceUpdate.StorkSignedPrice.QuantizedPrice), 10)

		// remove the 0x prefix
		encodedAssetId, err := stringToByte32(string(priceUpdate.StorkSignedPrice.EncodedAssetId))
		if err != nil {
			return err
		}

		rBytes, err := stringToByte32(priceUpdate.StorkSignedPrice.TimestampedSignature.Signature.R)
		if err != nil {
			return err
		}

		sBytes, err := stringToByte32(priceUpdate.StorkSignedPrice.TimestampedSignature.Signature.S)
		if err != nil {
			return err
		}

		publisherMerkleRoot, err := stringToByte32(priceUpdate.StorkSignedPrice.PublisherMerkleRoot)
		if err != nil {
			return err
		}

		checksum, err := stringToByte32(priceUpdate.StorkSignedPrice.StorkCalculationAlg.Checksum)
		if err != nil {
			return err
		}

		vInt, err := strconv.ParseInt(priceUpdate.StorkSignedPrice.TimestampedSignature.Signature.V[2:], 16, 8)
		if err != nil {
			return err
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

	fee, err := sci.contract.GetUpdateFeeV1(nil, updates)
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

	tx, err := sci.contract.UpdateTemporalNumericValuesV1(auth, updates)
	if err != nil {
		return err
	}

	sci.logger.Info().
		Str("txHash", tx.Hash().Hex()).
		Int("numUpdates", len(updates)).
		Uint64("gasPrice", tx.GasPrice().Uint64()).
		Msg("Pushed new values to contract")
	return nil
}
