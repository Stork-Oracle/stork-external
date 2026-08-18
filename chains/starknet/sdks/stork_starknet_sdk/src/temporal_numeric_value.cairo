//! Provides the [`TemporalNumericValue`] struct and its storage packing implementation.

use starknet::storage_access::StorePacking;

/// The type for encoded asset ids. Stork asset ids are 32 bytes, which does not fit in a
/// `felt252` (252 bits), so they are represented as a `u256` throughout the contract.
pub type EncodedAssetId = u256;

/// 2^128, used to split/join the packed representation.
pub const TWO_POW_128: felt252 = 0x100000000000000000000000000000000;

/// 2^127, the smallest `u128` whose two's complement interpretation is negative.
const TWO_POW_127: u128 = 0x80000000000000000000000000000000;

/// A timestamped value reported by Stork.
#[derive(Copy, Drop, Serde, PartialEq, Debug)]
pub struct TemporalNumericValue {
    /// The unix timestamp of the value in nanoseconds.
    pub timestamp_ns: u64,
    /// The quantized value.
    pub quantized_value: i128,
}

/// Packs a [`TemporalNumericValue`] into a single `felt252`.
///
/// The timestamp occupies the high 64 bits and the two's complement representation of the
/// quantized value occupies the low 128 bits, for 192 bits total. This keeps a feed update to a
/// single storage write, which matters because updates are the hot path of this contract.
pub impl TemporalNumericValueStorePacking of StorePacking<TemporalNumericValue, felt252> {
    fn pack(value: TemporalNumericValue) -> felt252 {
        value.timestamp_ns.into() * TWO_POW_128 + twos_complement(value.quantized_value)
    }

    fn unpack(value: felt252) -> TemporalNumericValue {
        let packed: u256 = value.into();
        TemporalNumericValue {
            timestamp_ns: packed.high.try_into().unwrap(),
            quantized_value: from_twos_complement(packed.low),
        }
    }
}

/// Returns the 128-bit two's complement representation of `value` as a `felt252`.
///
/// For a negative `value`, `value.into()` is the field element `P - |value|`; adding 2^128 yields
/// `2^128 - |value|`, which is exactly the two's complement encoding.
pub fn twos_complement(value: i128) -> felt252 {
    if value < 0 {
        value.into() + TWO_POW_128
    } else {
        value.into()
    }
}

/// Inverse of [`twos_complement`].
pub fn from_twos_complement(value: u128) -> i128 {
    let as_felt: felt252 = value.into();
    if value >= TWO_POW_127 {
        (as_felt - TWO_POW_128).try_into().unwrap()
    } else {
        as_felt.try_into().unwrap()
    }
}
