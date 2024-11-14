module stork::i128 {

    // === Errors ===

    const EMagnitudeTooLarge: u64 = 0;

    // === Constants ===
    
    const MAX_POSITIVE_MAGNITUDE: u128 = (1 << 127) - 1;
    const MAX_NEGATIVE_MAGNITUDE: u128 = (1 << 127);

    // === Structs ===

    public struct I128 has copy, drop, store {
        // sign of the i128, True if positive, false if negative
        sign: bool,
        // magnitude of the i128
        magnitude: u128,
    }

    // === Public Functions ===

    public fun new(magnitude: u128, sign: bool): I128 {
        if (sign) {
            assert!(magnitude <= MAX_POSITIVE_MAGNITUDE, EMagnitudeTooLarge);
        } else {
            assert!(magnitude <= MAX_NEGATIVE_MAGNITUDE, EMagnitudeTooLarge);
        }

        // Ensure consistent 0 representation corresponding to twos complements(positive sign)
        if (magnitude == 0) {
            sign = true;
        }
        I128 { 
            sign, 
            magnitude 
        }
    }

    public fun is_negative(i128: &I128): bool {
        !i128.sign
    }

    public fun get_magnitude(i128: &I128): u128 {
        i128.magnitude
    }

    public fun get_magnitude_if_negative(i128: &I128): u128 {
        assert!(is_negative(i128), 0);
        i128.magnitude
    }

    public fun get_magnitude_if_positive(i128: &I128): u128 {
        assert!(!is_negative(i128), 0);
        i128.magnitude
    }

    // from u128 to i128, assumes value is in twos complement representation
    public fun from_u128(value: u128): I128 {
        // Check the MSB for sign
        let sign = (value >> 127) == 1;
        if (sign) {
            // if positive, keep the value as is
            new(value, sign)
        } else {
            // if negative, take twos complement
            value = value ^ 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF + 1;
            new(value, sign)
        }
    }

    #[test]
    fun test_max_positive_magnitude() {
        let max_positive_magnitude = new(0x7FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF, true);
        assert!(max_positive_magnitude.sign, true)
        assert!(max_positive_magnitude.magnitude == MAX_POSITIVE_MAGNITUDE, true);
    }

    #[test]
    #[expected_failure(abort_code = 0)]
    fun test_magnitude_too_large_positive() {
        let magnitude_too_large_positive = 0x80000000000000000000000000000000;
        new(magnitude_too_large_positive, true);
        
    }
        


}
