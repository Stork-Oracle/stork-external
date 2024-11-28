module stork::i128 {

    // === Errors ===

    const EMagnitudeTooLarge: u64 = 0;
    const EInvalidSign: u64 = 1;

    // === Constants ===
    
    const MAX_POSITIVE_MAGNITUDE: u128 = (1 << 127) - 1;
    const MAX_NEGATIVE_MAGNITUDE: u128 = (1 << 127);

    // === Structs ===

    // the magnitude is the absolute value of the number
    // positive 1 is represented as (1, false)
    // negative 1 is represented as (1, true)
    public struct I128 has copy, drop, store {
        // sign of the i128, True if positive, false if negative
        negative: bool,
        // magnitude of the i128
        magnitude: u128,
    }

    // === Public Functions ===

    public fun new(magnitude: u128, negative: bool): I128 {
        let mut negative = negative;
        if (!negative) {
            assert!(magnitude <= MAX_POSITIVE_MAGNITUDE, EMagnitudeTooLarge);
        } else {
            assert!(magnitude <= MAX_NEGATIVE_MAGNITUDE, EMagnitudeTooLarge);
        };

        // Ensure consistent 0 representation corresponding to twos complements(positive sign)
        if (magnitude == 0) {
            negative = false;
        };
        I128 { 
            negative, 
            magnitude 
        }
    }

    public fun is_negative(i128: &I128): bool {
        i128.negative
    }

    public fun get_magnitude_if_negative(i128: &I128): u128 {
        assert!(is_negative(i128), EInvalidSign);
        i128.magnitude
    }

    public fun get_magnitude_if_positive(i128: &I128): u128 {
        assert!(!is_negative(i128), EInvalidSign);
        i128.magnitude
    }

    // from u128 to i128, assumes value is in twos complement representation
    public fun from_u128(value: u128): I128 {
        // Check the MSB for sign
        let negative = (value >> 127) == 1;
        if (!negative) {
            // if positive, keep the value as is
            new(value, false)
        } else {
            // if negative, convert from twos complement
            let neg_value = (value ^ 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF) + 1;
            new(neg_value, true)
        }
    }
    
    /// Converts the I128 to a big-endian byte representation compatible with Ethereum's int256
    public fun to_bytes(value: I128): vector<u8> {
        let mut bytes = vector::empty<u8>();
        let mut_value = if (value.negative) {
            // convert to twos complement
            (value.magnitude - 1) ^ 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF
        } else {
            value.magnitude
        };
        
        // Convert to big-endian bytes
        let mut i = 16; // Start from most significant byte (16 bytes total)
        while (i > 0) {
            i = i - 1;
            let byte = ((mut_value >> (i * 8)) & 0xFF as u8);
            vector::push_back(&mut bytes, byte);
        };

        bytes
    }

    #[test]
    fun test_max_positive_magnitude() {
        let max_positive_magnitude = new(0x7FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF, false);
        assert!(!max_positive_magnitude.negative, 1);
        assert!(max_positive_magnitude.magnitude == MAX_POSITIVE_MAGNITUDE, 1);
        assert!(&new(1<<127 -1, false) == &from_u128(1<<127 -1), 1);
    }

    #[test]
    #[expected_failure(abort_code = 0)]
    fun test_magnitude_too_large_positive() {
        let magnitude_too_large_positive = 0x80000000000000000000000000000000;
        new(magnitude_too_large_positive, false);
    }

    #[test]
    fun test_max_negative_magnitude() {
        let max_negative_magnitude = new(0x80000000000000000000000000000000, true);
        assert!(max_negative_magnitude.negative, 1);
        assert!(max_negative_magnitude.magnitude == MAX_NEGATIVE_MAGNITUDE, 1);
        assert!(&new(1<<127, true) == &from_u128(1<<127), 1);
    }

    #[test]
    #[expected_failure(abort_code = 0)]
    fun test_magnitude_too_large_negative() {
        let magnitude_too_large_negative = 0x80000000000000000000000000000001;
        new(magnitude_too_large_negative, true);
    }

    #[test]
    fun test_is_negative() {
        assert!(!is_negative(&new(1, false)), 1);
        assert!(is_negative(&new(1, true)), 1);
    }

    #[test]
    fun test_get_magnitude_if_negative_negative() {
        assert!(get_magnitude_if_negative(&new(1, true)) == 1, 1);
    }

    #[test]
    #[expected_failure(abort_code = 1)]
    fun test_get_magnitude_if_negative_positive() {
        get_magnitude_if_negative(&new(1, false));
    }

    #[test]
    fun test_get_magnitude_if_positive_positive() {
        assert!(get_magnitude_if_positive(&new(1, false)) == 1, 1);
    }

    #[test]
    #[expected_failure(abort_code = 1)]
    fun test_get_magnitude_if_positive_negative() {
        get_magnitude_if_positive(&new(1, true));
    }

    #[test]
    fun test_from_u128_positive() {
        assert!(&new(1, false) == &from_u128(1), 1);
    }

    #[test]
    fun test_from_u128_negative() {
        assert!(&new(1, true) == &from_u128(0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF), 1);
    }

    #[test]
    fun test_single_representation_of_zero() {
        assert!(&new(0, false) == &from_u128(0), 1);
        assert!(&new(0, true) == &from_u128(0), 1);
        let zero_positive = new(0, false);
        let zero_negative = new(0, true);
        assert!(&zero_positive == &zero_negative, 1);
        assert!(!is_negative(&zero_positive), 1);
        assert!(!is_negative(&zero_negative), 1);
    }

    #[test]
    fun test_to_bytes_positive() {
        let value = new(1, false); // Positive 1
        let bytes = to_bytes(value);
        assert!(bytes == x"00000000000000000000000000000001", 0);
        
        let value = new(0x7FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF, false); // Max positive
        let bytes = to_bytes(value);
        assert!(bytes == x"7FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF", 0);
    }

    #[test]
    fun test_to_bytes_negative() {
        let value = new(1, true); // Negative 1
        let bytes = to_bytes(value);
        assert!(bytes == x"FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF", 0);
        let value = new(0x80000000000000000000000000000000, true); // Max negative
        let bytes = to_bytes(value);
        assert!(bytes == x"80000000000000000000000000000000", 0);
    }

    #[test]
    fun test_to_bytes_zero() {
        let value = new(0, false); // Zero
        let bytes = to_bytes(value);
        assert!(bytes == x"00000000000000000000000000000000", 0);
        
        // Zero should be the same whether marked negative or positive
        let value = new(0, true);
        let bytes = to_bytes(value);
        assert!(bytes == x"00000000000000000000000000000000", 0);
    }

}
