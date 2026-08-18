package bindings

import (
	"fmt"
	"math/big"

	"github.com/NethermindEth/juno/core/felt"
	"github.com/NethermindEth/starknet.go/curve"
	"github.com/NethermindEth/starknet.go/rpc"
	"github.com/NethermindEth/starknet.go/utils"
)

const (
	// eventBufferSize bounds the websocket event channel.
	eventBufferSize = 128

	// feltsPerUpdate is the serialized width of one `TemporalNumericValueInput`:
	// id (2) + value (2) + merkle root (2) + alg hash (2) + r (2) + s (2) + v (1).
	feltsPerUpdate = 13

	// u256Limb is 2^128, the boundary between the low and high limbs of a Cairo u256.
	u256LimbBits = 128

	// The Starknet field prime is 2^251 + 17*2^192 + 1.
	fieldPrimeHighBit   = 251
	fieldPrimeMidFactor = 17
	fieldPrimeMidBit    = 192
)

// twoPow128 is the modulus of a single u256 limb.
//
//nolint:gochecknoglobals // Effectively a constant; big.Int cannot be one.
var twoPow128 = new(big.Int).Lsh(big.NewInt(1), u256LimbBits)

// encodeU256 splits a 32 byte big-endian value into Cairo's (low, high) limb pair.
func encodeU256(value []byte) []*felt.Felt {
	full := new(big.Int).SetBytes(value)

	high, low := new(big.Int).DivMod(full, twoPow128, new(big.Int))

	return []*felt.Felt{utils.BigIntToFelt(low), utils.BigIntToFelt(high)}
}

// decodeU256 rejoins a Cairo (low, high) limb pair.
func decodeU256(low, high *felt.Felt) *big.Int {
	value := new(big.Int).Lsh(utils.FeltToBigInt(high), u256LimbBits)

	return value.Add(value, utils.FeltToBigInt(low))
}

// encodeI128 embeds a signed value into the field the way Cairo's `Serde<i128>` does: a negative
// `v` is serialized as the field element `P - |v|`, not as a two's complement bit pattern.
//
// The reduction has to be done here rather than left to utils.BigIntToFelt, which builds the felt
// from big.Int.Bytes() and so silently discards the sign.
func encodeI128(value *big.Int) *felt.Felt {
	if value.Sign() < 0 {
		return utils.BigIntToFelt(new(big.Int).Add(fieldPrime(), value))
	}

	return utils.BigIntToFelt(value)
}

// decodeI128 is the inverse of encodeI128. Field elements above P/2 represent negative values,
// but only the top 2^127 of the range is reachable for an i128, so anything at or above
// `P - 2^127` is mapped back to a negative number.
func decodeI128(value *felt.Felt) *big.Int {
	asBigInt := utils.FeltToBigInt(value)

	// The felt prime; values in [P - 2^127, P) are negative i128s.
	prime := fieldPrime()
	negativeBoundary := new(big.Int).Sub(prime, new(big.Int).Rsh(twoPow128, 1))

	if asBigInt.Cmp(negativeBoundary) >= 0 {
		return new(big.Int).Sub(asBigInt, prime)
	}

	return asBigInt
}

// fieldPrime returns the Starknet field prime, 2^251 + 17*2^192 + 1.
func fieldPrime() *big.Int {
	prime := new(big.Int).Lsh(big.NewInt(1), fieldPrimeHighBit)
	prime.Add(prime, new(big.Int).Lsh(big.NewInt(fieldPrimeMidFactor), fieldPrimeMidBit))

	return prime.Add(prime, big.NewInt(1))
}

// encodeUpdateData serializes `Array<TemporalNumericValueInput>` as Cairo calldata.
func encodeUpdateData(updates []UpdateData) ([]*felt.Felt, error) {
	calldata := make([]*felt.Felt, 0, 1+len(updates)*feltsPerUpdate)
	calldata = append(calldata, utils.Uint64ToFelt(uint64(len(updates))))

	for _, update := range updates {
		// The contract rejects any other recovery id, so fail before paying for a transaction.
		if update.V != sigV27 && update.V != sigV28 {
			return nil, fmt.Errorf("%w: got %d", ErrInvalidSignatureV, update.V)
		}

		calldata = append(calldata, encodeU256(update.ID[:])...)
		calldata = append(calldata, utils.Uint64ToFelt(update.TemporalNumericValue.TimestampNs))
		calldata = append(calldata, encodeI128(update.TemporalNumericValue.QuantizedValue))
		calldata = append(calldata, encodeU256(update.PublisherMerkleRoot[:])...)
		calldata = append(calldata, encodeU256(update.ValueComputeAlgHash[:])...)
		calldata = append(calldata, encodeU256(update.R[:])...)
		calldata = append(calldata, encodeU256(update.S[:])...)
		calldata = append(calldata, utils.Uint64ToFelt(uint64(update.V)))
	}

	return calldata, nil
}

// decodeValueArray decodes an `Array<TemporalNumericValue>` return value.
func decodeValueArray(result []*felt.Felt, expected int) ([]TemporalNumericValue, error) {
	if len(result) == 0 {
		return nil, fmt.Errorf("%w: empty result", ErrUnexpectedResultLength)
	}

	length := utils.FeltToBigInt(result[0]).Int64()
	if int(length) != expected {
		return nil, fmt.Errorf("%w: contract returned %d values, want %d", ErrUnexpectedResultLength, length, expected)
	}

	if len(result) != 1+expected*feltsPerValue {
		return nil, fmt.Errorf(
			"%w: got %d felts, want %d", ErrUnexpectedResultLength, len(result), 1+expected*feltsPerValue,
		)
	}

	values := make([]TemporalNumericValue, 0, expected)
	for i := range expected {
		offset := 1 + i*feltsPerValue
		values = append(values, TemporalNumericValue{
			TimestampNs:    utils.FeltToBigInt(result[offset]).Uint64(),
			QuantizedValue: decodeI128(result[offset+1]),
		})
	}

	return values, nil
}

// decodeValueUpdateEvent decodes a `ValueUpdate` event.
//
// The Cairo event declares `id` as a key, so the u256 limbs land in keys[1] and keys[2] after the
// selector, while the timestamp and value are in data.
func decodeValueUpdateEvent(event *rpc.EmittedEventWithFinalityStatus) (ValueUpdateEvent, error) {
	//nolint:mnd // selector + the two limbs of the keyed u256 id.
	if len(event.Keys) < 3 {
		return ValueUpdateEvent{}, fmt.Errorf(
			"%w: got %d event keys, want 3", ErrUnexpectedResultLength, len(event.Keys),
		)
	}

	if len(event.Data) < feltsPerValue {
		return ValueUpdateEvent{}, fmt.Errorf(
			"%w: got %d event data felts, want %d", ErrUnexpectedResultLength, len(event.Data), feltsPerValue,
		)
	}

	var id EncodedAssetID

	decodeU256(event.Keys[1], event.Keys[2]).FillBytes(id[:])

	return ValueUpdateEvent{
		ID: id,
		Value: TemporalNumericValue{
			TimestampNs:    utils.FeltToBigInt(event.Data[0]).Uint64(),
			QuantizedValue: decodeI128(event.Data[1]),
		},
	}, nil
}

// publicKeyFromPrivateKey derives the Stark public key used to look the signer up in the keystore.
// Only the x coordinate identifies the key on Starknet.
func publicKeyFromPrivateKey(privateKey *big.Int) string {
	x, _ := curve.PrivateKeyToPoint(privateKey)

	return utils.BigIntToFelt(x).String()
}
