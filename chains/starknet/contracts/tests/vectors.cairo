//! Real Stork-signed updates, lifted from
//! `apps/chain_pusher/internal/testutil/testdata` so the Cairo contract is verified against the
//! exact payloads the chain pusher forwards.

use starknet::EthAddress;
use stork_starknet_sdk::interface::TemporalNumericValueInput;
use stork_starknet_sdk::temporal_numeric_value::TemporalNumericValue;

/// The key every vector below is signed with.
pub fn stork_public_key() -> EthAddress {
    0xC4A02e7D370402F4afC36032076B05e74FF81786.try_into().unwrap()
}

/// A key that never signed any of these vectors.
pub fn other_public_key() -> EthAddress {
    0x0a803F9b1CCe32e2773e0d2e98b37E0775cA5d44.try_into().unwrap()
}

/// The checksum of the `median v1` algorithm, shared by every vector.
const MEDIAN_V1: u256 = 0x9be7e9f9ed459417d96112a7467bd0b27575a2c7847195c68f805b70ce1795ba;

pub fn positive_asset_1() -> TemporalNumericValueInput {
    TemporalNumericValueInput {
        id: 0x4de9a89eed25754cfff794b5d7e8a71234cf930ee6bfb71ea8b8aa0ce313699f,
        temporal_numeric_value: TemporalNumericValue {
            timestamp_ns: 1757543080509034593, quantized_value: 100002000000000000000000,
        },
        publisher_merkle_root: 0x7dbca74780f2c8da0f9b17f3bae1280ab787900930c557b215701de7f83c43c0,
        value_compute_alg_hash: MEDIAN_V1,
        r: 0x72604d29eb8cdb9009d51e1c2e482e4aa3929d3ac3bf7e9185ceec165edd4490,
        s: 0x1bcfeefd20bdc814c75c0745e1d546fd996cc59d6b3653c5779724c83dfe7e23,
        v: 0x1c,
    }
}

pub fn positive_asset_2() -> TemporalNumericValueInput {
    TemporalNumericValueInput {
        id: 0xc4667b76759b9712094e1ff4fb729ebc88b73dc297217375f5db651d45647036,
        temporal_numeric_value: TemporalNumericValue {
            timestamp_ns: 1757543111515515052, quantized_value: 20000002999999999999000000,
        },
        publisher_merkle_root: 0xa4e033a8ea8adc798f4aefef08076f2bfa4e6c27de56eab93fb9946abc174eab,
        value_compute_alg_hash: MEDIAN_V1,
        r: 0x63cb041b28313077fdc51f6eeae4bb780177845b305708568f177583212c38ba,
        s: 0x4a4aafdd4bd976df481086b4f5df33bdb94d33998dc6e9d2070cbda8dcb358b9,
        v: 0x1b,
    }
}

pub fn positive_asset_3() -> TemporalNumericValueInput {
    TemporalNumericValueInput {
        id: 0x30b57100028b795b2adff417ab3c20ecaf0a84569459542e42d9a98fdada153e,
        temporal_numeric_value: TemporalNumericValue {
            timestamp_ns: 1757543173718568928, quantized_value: 4000000000000000000,
        },
        publisher_merkle_root: 0xa1a7a9ba6c55b7ea1a2061af0c9d4a711fd53cf655c23e541ed6d32e98918cc6,
        value_compute_alg_hash: MEDIAN_V1,
        r: 0x4adca34ccd410a46c3182999a1bb602ed0d05502e7dda1f6b96f0400b8d01b86,
        s: 0x6116c9497d40893074a2674f2c0d42a2c46d635fb57ad5e95840ac3d82a47abe,
        v: 0x1c,
    }
}

pub fn negative_asset_1() -> TemporalNumericValueInput {
    TemporalNumericValueInput {
        id: 0x2368c2095acad21c871c2925da18e8e01e6ea6ed731dd1ffd57cd313935cd659,
        temporal_numeric_value: TemporalNumericValue {
            timestamp_ns: 1757543204021510845, quantized_value: -100000000000000000000,
        },
        publisher_merkle_root: 0x7e2f7ea09c7ee54e4bd4bc1b35c9218130a490245370900c11fe339f2d99654b,
        value_compute_alg_hash: MEDIAN_V1,
        r: 0x843a375411f084dea307c87f50ad4b9012c5041171206df630e06d620daca1fe,
        s: 0x5b130b256db75b20607636d731596d2a8008fff1cb6fc7fc66cbfa63dacc8b13,
        v: 0x1b,
    }
}

/// All four vectors, for batch tests.
pub fn all_vectors() -> Array<TemporalNumericValueInput> {
    array![positive_asset_1(), positive_asset_2(), positive_asset_3(), negative_asset_1()]
}
