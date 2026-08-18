//! Tests for the packed storage representation of [`TemporalNumericValue`].

use stork_starknet_sdk::temporal_numeric_value::{
    TemporalNumericValue, TemporalNumericValueStorePacking, from_twos_complement, twos_complement,
};

fn roundtrip(timestamp_ns: u64, quantized_value: i128) {
    let value = TemporalNumericValue { timestamp_ns, quantized_value };
    let packed = TemporalNumericValueStorePacking::pack(value);
    let unpacked = TemporalNumericValueStorePacking::unpack(packed);

    assert!(unpacked == value, "roundtrip failed for {} / {}", timestamp_ns, quantized_value);
}

#[test]
fn test_pack_roundtrip_positive() {
    roundtrip(1757543080509034593, 100002000000000000000000);
}

#[test]
fn test_pack_roundtrip_negative() {
    roundtrip(1757543204021510845, -100000000000000000000);
}

#[test]
fn test_pack_roundtrip_zero() {
    roundtrip(0, 0);
}

#[test]
fn test_pack_roundtrip_extremes() {
    roundtrip(0xffffffffffffffff, 0x7fffffffffffffffffffffffffffffff);
    roundtrip(0xffffffffffffffff, -0x80000000000000000000000000000000);
    roundtrip(1, -1);
}

#[test]
fn test_unpack_zero_is_empty_feed() {
    // A never-written storage slot has to read back as the "not found" sentinel.
    let empty = TemporalNumericValueStorePacking::unpack(0);

    assert!(empty.timestamp_ns == 0);
    assert!(empty.quantized_value == 0);
}

#[test]
fn test_twos_complement() {
    assert!(twos_complement(0) == 0);
    assert!(twos_complement(1) == 1);
    assert!(twos_complement(-1) == 0xffffffffffffffffffffffffffffffff);
    assert!(twos_complement(-2) == 0xfffffffffffffffffffffffffffffffe);
    assert!(
        twos_complement(-0x80000000000000000000000000000000) == 0x80000000000000000000000000000000,
    );
}

#[test]
fn test_from_twos_complement() {
    assert!(from_twos_complement(0) == 0);
    assert!(from_twos_complement(1) == 1);
    assert!(from_twos_complement(0xffffffffffffffffffffffffffffffff) == -1);
    assert!(
        from_twos_complement(
            0x7fffffffffffffffffffffffffffffff,
        ) == 0x7fffffffffffffffffffffffffffffff,
    );
    assert!(
        from_twos_complement(
            0x80000000000000000000000000000000,
        ) == -0x80000000000000000000000000000000,
    );
}
