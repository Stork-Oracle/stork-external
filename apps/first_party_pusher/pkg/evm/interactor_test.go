package first_party_evm

import (
	"math/big"
	"testing"

	"github.com/Stork-Oracle/stork-external/apps/first_party_pusher/pkg/types"
	publisher_agent "github.com/Stork-Oracle/stork-external/apps/publisher_agent/pkg"
	"github.com/Stork-Oracle/stork-external/shared"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// FirstPartyPriceCase represents a test case for FirstPartyStork price updates.
// This is similar to testutil.PriceCase but for FirstParty-specific data structures.
type FirstPartyPriceCase struct {
	Name         string
	Update       publisher_agent.SignedPriceUpdate[*shared.EvmSignature]
	AssetEntry   types.AssetEntry
	ExpectedData *ExpectedUpdateData
	WantError    bool
}

type ExpectedUpdateData struct {
	PubKey          common.Address
	AssetPairID     string
	StoreHistorical bool
	TimestampNs     uint64
	QuantizedValue  *big.Int
	R               [32]byte
	S               [32]byte
	V               uint8
}

// standardFirstPartyPriceCases is an adapted version of testutil.StandardPriceCase for FirstPartyStork.
func standardFirstPartyPriceCases() []FirstPartyPriceCase {
	return []FirstPartyPriceCase{
		validPositiveFirstPartyCase(),
		validNegativeFirstPartyCase(),
		validZeroFirstPartyCase(),
		invalidVSignatureFirstPartyCase(),
		invalidQuantizedPriceFirstPartyCase(),
	}
}

func defaultFirstPartyPriceCase() FirstPartyPriceCase {
	pubKey := shared.PublisherKey("0x0810E094a8b0e750c7ACB66F469AfBBd595FF69b")
	assetID := shared.AssetID("ETHUSD")
	quantizedPrice := shared.QuantizedPrice("1000000000000000000")
	timestampNs := uint64(1680210934000000000)

	return FirstPartyPriceCase{
		Name: "valid positive price",
		Update: publisher_agent.SignedPriceUpdate[*shared.EvmSignature]{
			OracleID: "test-oracle",
			AssetID:  assetID,
			Trigger:  publisher_agent.ClockTriggerType,
			SignedPrice: publisher_agent.SignedPrice[*shared.EvmSignature]{
				PublisherKey:    pubKey,
				ExternalAssetID: string(assetID),
				SignatureType:   shared.EvmSignatureType,
				QuantizedPrice:  quantizedPrice,
				TimestampedSignature: shared.TimestampedSignature[*shared.EvmSignature]{
					TimestampNano: timestampNs,
					Signature: &shared.EvmSignature{
						R: "0xd80926f0433827d55e17bc77953b44788fb40057c55b2578da4f59361d758555",
						S: "0x69703bad148facb6ba7e5d61676240d6e50162d97e0e7e31d7c7ccd35db6df5f",
						V: "0x1b",
					},
				},
			},
		},
		AssetEntry: types.AssetEntry{
			AssetID:    assetID,
			PublicKey:  pubKey,
			Historical: false,
		},
		ExpectedData: &ExpectedUpdateData{
			PubKey:         common.HexToAddress(string(pubKey)),
			AssetPairID:    string(assetID),
			TimestampNs:    timestampNs,
			QuantizedValue: big.NewInt(1000000000000000000),
			R: [32]byte{
				0xd8, 0x09, 0x26, 0xf0, 0x43, 0x38, 0x27, 0xd5,
				0x5e, 0x17, 0xbc, 0x77, 0x95, 0x3b, 0x44, 0x78,
				0x8f, 0xb4, 0x00, 0x57, 0xc5, 0x5b, 0x25, 0x78,
				0xda, 0x4f, 0x59, 0x36, 0x1d, 0x75, 0x85, 0x55,
			},
			S: [32]byte{
				0x69, 0x70, 0x3b, 0xad, 0x14, 0x8f, 0xac, 0xb6,
				0xba, 0x7e, 0x5d, 0x61, 0x67, 0x62, 0x40, 0xd6,
				0xe5, 0x01, 0x62, 0xd9, 0x7e, 0x0e, 0x7e, 0x31,
				0xd7, 0xc7, 0xcc, 0xd3, 0x5d, 0xb6, 0xdf, 0x5f,
			},
			V: 27,
		},
		WantError: false,
	}
}

func validPositiveFirstPartyCase() FirstPartyPriceCase {
	return defaultFirstPartyPriceCase()
}

func validNegativeFirstPartyCase() FirstPartyPriceCase {
	defaultFirstPartyPriceCase := defaultFirstPartyPriceCase()

	defaultFirstPartyPriceCase.Update.SignedPrice.QuantizedPrice = shared.QuantizedPrice("-1000000000000000000")
	defaultFirstPartyPriceCase.ExpectedData.QuantizedValue = big.NewInt(-1000000000000000000)

	return defaultFirstPartyPriceCase
}

func validZeroFirstPartyCase() FirstPartyPriceCase {
	defaultFirstPartyPriceCase := defaultFirstPartyPriceCase()

	defaultFirstPartyPriceCase.Update.SignedPrice.QuantizedPrice = shared.QuantizedPrice("0")
	defaultFirstPartyPriceCase.ExpectedData.QuantizedValue = big.NewInt(0)

	return defaultFirstPartyPriceCase
}

func invalidVSignatureFirstPartyCase() FirstPartyPriceCase {
	defaultFirstPartyPriceCase := defaultFirstPartyPriceCase()

	defaultFirstPartyPriceCase.Update.SignedPrice.TimestampedSignature.Signature.V = "invalid"
	defaultFirstPartyPriceCase.WantError = true

	return defaultFirstPartyPriceCase
}

func invalidQuantizedPriceFirstPartyCase() FirstPartyPriceCase {
	defaultFirstPartyPriceCase := defaultFirstPartyPriceCase()

	defaultFirstPartyPriceCase.Update.SignedPrice.QuantizedPrice = shared.QuantizedPrice("not_a_number")
	defaultFirstPartyPriceCase.WantError = true

	return defaultFirstPartyPriceCase
}

func TestGetUpdatePayload(t *testing.T) {
	t.Parallel()

	testCases := standardFirstPartyPriceCases()

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			updatesByEntry := map[types.AssetEntry]publisher_agent.SignedPriceUpdate[*shared.EvmSignature]{
				tc.AssetEntry: tc.Update,
			}

			logger := zerolog.Nop()
			ci := &ContractInteractor{
				logger: logger,
			}

			payload, err := ci.getUpdatePayload(updatesByEntry)

			if tc.WantError {
				assert.Error(t, err)

				return
			}

			require.NoError(t, err)
			require.Len(t, payload, 1)

			assert.Equal(t, tc.ExpectedData.PubKey, payload[0].PubKey)
			assert.Equal(t, tc.ExpectedData.AssetPairID, payload[0].AssetPairId)
			assert.Equal(t, tc.ExpectedData.StoreHistorical, payload[0].StoreHistorical)
			assert.Equal(t, tc.ExpectedData.TimestampNs, payload[0].TemporalNumericValue.TimestampNs)
			assert.Equal(t, tc.ExpectedData.QuantizedValue, payload[0].TemporalNumericValue.QuantizedValue)
			assert.Equal(t, tc.ExpectedData.R, payload[0].R)
			assert.Equal(t, tc.ExpectedData.S, payload[0].S)
			assert.Equal(t, tc.ExpectedData.V, payload[0].V)
		})
	}

	t.Run("multiple updates with mixed historical flags", func(t *testing.T) {
		t.Parallel()

		case1 := validPositiveFirstPartyCase()
		case1.AssetEntry.Historical = true

		case2 := validZeroFirstPartyCase()
		case2.AssetEntry.Historical = false

		updatesByEntry := map[types.AssetEntry]publisher_agent.SignedPriceUpdate[*shared.EvmSignature]{
			case1.AssetEntry: case1.Update,
			case2.AssetEntry: case2.Update,
		}

		logger := zerolog.Nop()
		ci := &ContractInteractor{
			logger: logger,
		}

		payload, err := ci.getUpdatePayload(updatesByEntry)
		require.NoError(t, err)
		assert.Len(t, payload, 2)

		historicalCount := 0

		for _, p := range payload {
			if p.StoreHistorical {
				historicalCount++
			}
		}

		assert.Equal(t, 1, historicalCount, "should have exactly one historical update")
	})
}
