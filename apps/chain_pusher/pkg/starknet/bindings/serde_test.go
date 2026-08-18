package bindings

import (
	"math/big"
	"testing"

	"github.com/NethermindEth/juno/core/felt"
	"github.com/NethermindEth/starknet.go/rpc"
	"github.com/NethermindEth/starknet.go/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func hexToFelt(t *testing.T, hex string) *felt.Felt {
	t.Helper()

	value, err := utils.HexToFelt(hex)
	require.NoError(t, err)

	return value
}

func bytes32(t *testing.T, hex string) [32]byte {
	t.Helper()

	value, ok := new(big.Int).SetString(hex, 16)
	require.True(t, ok, "invalid hex: %s", hex)

	var out [32]byte
	value.FillBytes(out[:])

	return out
}

func TestEncodeU256SplitsIntoLimbs(t *testing.T) {
	t.Parallel()

	// The low limb is the bottom 16 bytes, the high limb the top 16.
	id := bytes32(t, "0011223344556677889900aabbccddeeff112233445566778899aabbccddeeff")

	limbs := encodeU256(id[:])

	require.Len(t, limbs, 2)
	assert.Equal(t, hexToFelt(t, "0xff112233445566778899aabbccddeeff"), limbs[0], "low limb")
	assert.Equal(t, hexToFelt(t, "0x0011223344556677889900aabbccddee"), limbs[1], "high limb")
}

func TestU256RoundTrip(t *testing.T) {
	t.Parallel()

	for _, hex := range []string{
		"0000000000000000000000000000000000000000000000000000000000000000",
		"0000000000000000000000000000000000000000000000000000000000000001",
		"ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
		"4de9a89eed25754cfff794b5d7e8a71234cf930ee6bfb71ea8b8aa0ce313699f",
	} {
		value := bytes32(t, hex)
		limbs := encodeU256(value[:])

		var got [32]byte

		decodeU256(limbs[0], limbs[1]).FillBytes(got[:])

		assert.Equal(t, value, got, "round trip failed for %s", hex)
	}
}

// Cairo's Serde<i128> embeds a negative value into the field as P - |v|, which is not the same as
// the 128 bit two's complement encoding used inside the signed message. Getting these two mixed up
// would produce updates that fail signature verification on chain.
func TestEncodeI128UsesFieldEmbedding(t *testing.T) {
	t.Parallel()

	prime := fieldPrime()

	tests := []struct {
		name  string
		value *big.Int
		want  *big.Int
	}{
		{name: "zero", value: big.NewInt(0), want: big.NewInt(0)},
		{name: "positive", value: big.NewInt(1234), want: big.NewInt(1234)},
		{
			name:  "large positive price",
			value: mustBigInt(t, "100002000000000000000000"),
			want:  mustBigInt(t, "100002000000000000000000"),
		},
		{
			name:  "negative one",
			value: big.NewInt(-1),
			want:  new(big.Int).Sub(prime, big.NewInt(1)),
		},
		{
			name:  "negative price",
			value: mustBigInt(t, "-100000000000000000000"),
			want:  new(big.Int).Sub(prime, mustBigInt(t, "100000000000000000000")),
		},
		{
			name:  "i128 min",
			value: new(big.Int).Neg(new(big.Int).Lsh(big.NewInt(1), 127)),
			want:  new(big.Int).Sub(prime, new(big.Int).Lsh(big.NewInt(1), 127)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := utils.FeltToBigInt(encodeI128(tt.value))
			assert.Equal(t, 0, tt.want.Cmp(got), "want %s, got %s", tt.want, got)
		})
	}
}

func TestI128RoundTrip(t *testing.T) {
	t.Parallel()

	values := []*big.Int{
		big.NewInt(0),
		big.NewInt(1),
		big.NewInt(-1),
		mustBigInt(t, "100002000000000000000000"),
		mustBigInt(t, "-100000000000000000000"),
		mustBigInt(t, "20000002999999999999000000"),
		new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 127), big.NewInt(1)),
		new(big.Int).Neg(new(big.Int).Lsh(big.NewInt(1), 127)),
	}

	for _, value := range values {
		got := decodeI128(encodeI128(value))
		assert.Equal(t, 0, value.Cmp(got), "round trip failed: want %s, got %s", value, got)
	}
}

func TestEncodeUpdateDataLayout(t *testing.T) {
	t.Parallel()

	update := UpdateData{
		ID: bytes32(t, "4de9a89eed25754cfff794b5d7e8a71234cf930ee6bfb71ea8b8aa0ce313699f"),
		TemporalNumericValue: TemporalNumericValue{
			TimestampNs:    1757543080509034593,
			QuantizedValue: mustBigInt(t, "100002000000000000000000"),
		},
		PublisherMerkleRoot: bytes32(t, "7dbca74780f2c8da0f9b17f3bae1280ab787900930c557b215701de7f83c43c0"),
		ValueComputeAlgHash: bytes32(t, "9be7e9f9ed459417d96112a7467bd0b27575a2c7847195c68f805b70ce1795ba"),
		R:                   bytes32(t, "72604d29eb8cdb9009d51e1c2e482e4aa3929d3ac3bf7e9185ceec165edd4490"),
		S:                   bytes32(t, "1bcfeefd20bdc814c75c0745e1d546fd996cc59d6b3653c5779724c83dfe7e23"),
		V:                   28,
	}

	calldata, err := encodeUpdateData([]UpdateData{update})
	require.NoError(t, err)

	// Array length followed by exactly one serialized struct.
	require.Len(t, calldata, 1+feltsPerUpdate)

	assert.Equal(t, hexToFelt(t, "0x1"), calldata[0], "array length")
	assert.Equal(t, hexToFelt(t, "0x34cf930ee6bfb71ea8b8aa0ce313699f"), calldata[1], "id low")
	assert.Equal(t, hexToFelt(t, "0x4de9a89eed25754cfff794b5d7e8a712"), calldata[2], "id high")
	assert.Equal(t, hexToFelt(t, "0x18640c1eaf2fd061"), calldata[3], "timestamp_ns")
	assert.Equal(t, encodeI128(mustBigInt(t, "100002000000000000000000")), calldata[4], "quantized_value")
	assert.Equal(t, hexToFelt(t, "0xb787900930c557b215701de7f83c43c0"), calldata[5], "merkle root low")
	assert.Equal(t, hexToFelt(t, "0x7dbca74780f2c8da0f9b17f3bae1280a"), calldata[6], "merkle root high")
	assert.Equal(t, hexToFelt(t, "0x7575a2c7847195c68f805b70ce1795ba"), calldata[7], "alg hash low")
	assert.Equal(t, hexToFelt(t, "0x9be7e9f9ed459417d96112a7467bd0b2"), calldata[8], "alg hash high")
	assert.Equal(t, hexToFelt(t, "0xa3929d3ac3bf7e9185ceec165edd4490"), calldata[9], "r low")
	assert.Equal(t, hexToFelt(t, "0x72604d29eb8cdb9009d51e1c2e482e4a"), calldata[10], "r high")
	assert.Equal(t, hexToFelt(t, "0x996cc59d6b3653c5779724c83dfe7e23"), calldata[11], "s low")
	assert.Equal(t, hexToFelt(t, "0x1bcfeefd20bdc814c75c0745e1d546fd"), calldata[12], "s high")
	assert.Equal(t, hexToFelt(t, "0x1c"), calldata[13], "v")
}

func TestEncodeUpdateDataBatchLength(t *testing.T) {
	t.Parallel()

	update := UpdateData{
		TemporalNumericValue: TemporalNumericValue{TimestampNs: 1, QuantizedValue: big.NewInt(1)},
		V:                    27,
	}

	calldata, err := encodeUpdateData([]UpdateData{update, update, update})
	require.NoError(t, err)

	require.Len(t, calldata, 1+3*feltsPerUpdate)
	assert.Equal(t, hexToFelt(t, "0x3"), calldata[0])
}

func TestEncodeUpdateDataRejectsNonCanonicalV(t *testing.T) {
	t.Parallel()

	// The contract only accepts 27 and 28, so catch it before paying for a reverting transaction.
	for _, v := range []uint8{0, 1, 26, 29, 255} {
		_, err := encodeUpdateData([]UpdateData{{
			TemporalNumericValue: TemporalNumericValue{TimestampNs: 1, QuantizedValue: big.NewInt(1)},
			V:                    v,
		}})

		require.ErrorIs(t, err, ErrInvalidSignatureV, "v = %d should be rejected", v)
	}
}

func TestDecodeValueArray(t *testing.T) {
	t.Parallel()

	result := []*felt.Felt{
		hexToFelt(t, "0x2"),
		hexToFelt(t, "0x18640c1eaf2fd061"),
		encodeI128(mustBigInt(t, "100002000000000000000000")),
		hexToFelt(t, "0x18640c3b711aa2bd"),
		encodeI128(mustBigInt(t, "-100000000000000000000")),
	}

	values, err := decodeValueArray(result, 2)
	require.NoError(t, err)

	require.Len(t, values, 2)
	assert.Equal(t, uint64(1757543080509034593), values[0].TimestampNs)
	assert.Equal(t, "100002000000000000000000", values[0].QuantizedValue.String())
	assert.Equal(t, uint64(1757543204021510845), values[1].TimestampNs)
	assert.Equal(t, "-100000000000000000000", values[1].QuantizedValue.String())
}

func TestDecodeValueArrayRejectsMalformedResults(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		result []*felt.Felt
		expect int
	}{
		{name: "empty", result: nil, expect: 1},
		{
			name:   "length header disagrees with request",
			result: []*felt.Felt{hexToFelt(t, "0x1"), hexToFelt(t, "0x1"), hexToFelt(t, "0x1")},
			expect: 2,
		},
		{
			name:   "truncated payload",
			result: []*felt.Felt{hexToFelt(t, "0x2"), hexToFelt(t, "0x1"), hexToFelt(t, "0x1")},
			expect: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := decodeValueArray(tt.result, tt.expect)
			require.ErrorIs(t, err, ErrUnexpectedResultLength)
		})
	}
}

func TestDecodeValueArrayEmptyRequest(t *testing.T) {
	t.Parallel()

	values, err := decodeValueArray([]*felt.Felt{hexToFelt(t, "0x0")}, 0)

	require.NoError(t, err)
	assert.Empty(t, values)
}

// The Cairo event marks `id` as a key, so it lands in keys[1:3] after the event selector rather
// than in the data section.
func TestDecodeValueUpdateEvent(t *testing.T) {
	t.Parallel()

	id := bytes32(t, "4de9a89eed25754cfff794b5d7e8a71234cf930ee6bfb71ea8b8aa0ce313699f")
	limbs := encodeU256(id[:])

	event := &rpc.EmittedEventWithFinalityStatus{
		EmittedEvent: rpc.EmittedEvent{
			Event: rpc.Event{
				EventContent: rpc.EventContent{
					Keys: []*felt.Felt{
						utils.GetSelectorFromNameFelt("ValueUpdate"),
						limbs[0],
						limbs[1],
					},
					Data: []*felt.Felt{
						hexToFelt(t, "0x18640c1eaf2fd061"),
						encodeI128(mustBigInt(t, "-100000000000000000000")),
					},
				},
			},
		},
	}

	decoded, err := decodeValueUpdateEvent(event)
	require.NoError(t, err)

	assert.Equal(t, id, [32]byte(decoded.ID))
	assert.Equal(t, uint64(1757543080509034593), decoded.Value.TimestampNs)
	assert.Equal(t, "-100000000000000000000", decoded.Value.QuantizedValue.String())
}

func TestDecodeValueUpdateEventRejectsMalformedEvents(t *testing.T) {
	t.Parallel()

	missingKeys := &rpc.EmittedEventWithFinalityStatus{
		EmittedEvent: rpc.EmittedEvent{Event: rpc.Event{EventContent: rpc.EventContent{
			Keys: []*felt.Felt{utils.GetSelectorFromNameFelt("ValueUpdate")},
			Data: []*felt.Felt{hexToFelt(t, "0x1"), hexToFelt(t, "0x1")},
		}}},
	}

	_, err := decodeValueUpdateEvent(missingKeys)
	require.ErrorIs(t, err, ErrUnexpectedResultLength)

	missingData := &rpc.EmittedEventWithFinalityStatus{
		EmittedEvent: rpc.EmittedEvent{Event: rpc.Event{EventContent: rpc.EventContent{
			Keys: []*felt.Felt{
				utils.GetSelectorFromNameFelt("ValueUpdate"),
				hexToFelt(t, "0x1"),
				hexToFelt(t, "0x0"),
			},
			Data: []*felt.Felt{hexToFelt(t, "0x1")},
		}}},
	}

	_, err = decodeValueUpdateEvent(missingData)
	require.ErrorIs(t, err, ErrUnexpectedResultLength)
}

// Entry point and event selectors are the contract's real ABI surface: if a Cairo function or
// event is renamed, these change and calls silently start missing. The values are the ones
// observed on a deployed contract.
func TestSelectorsMatchDeployedContract(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"update_temporal_numeric_values_v1":              "0x3708378f8c0a945341750fe9c0e4ba9541c1413c14bea51b320111751c13157",
		"get_multiple_temporal_numeric_values_unchecked": "0x22db8d1f5e6eaa43a868210d968943b9f8e697342079ad9ff1c55007736d0f0",
		"ValueUpdate": "0x1909579031f521c23e848c1323651712c83c10cc33556e8d4eba96f46272277",
	}

	for name, want := range tests {
		assert.Equal(t, want, utils.GetSelectorFromNameFelt(name).String(), "selector for %s", name)
	}
}

func TestPublicKeyFromPrivateKeyIsDeterministic(t *testing.T) {
	t.Parallel()

	key := mustBigInt(t, "1234567890987654321")

	first := publicKeyFromPrivateKey(key)
	second := publicKeyFromPrivateKey(key)

	assert.Equal(t, first, second, "same key should derive the same public key")
	assert.NotEqual(t, first, publicKeyFromPrivateKey(big.NewInt(2)))
}

func mustBigInt(t *testing.T, value string) *big.Int {
	t.Helper()

	result, ok := new(big.Int).SetString(value, 10)
	require.True(t, ok, "invalid decimal: %s", value)

	return result
}
