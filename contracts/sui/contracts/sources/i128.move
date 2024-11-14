module stork::i128 {

    // === Errors ===

    const EMagnitudeTooLarge: u64 = 0;

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
        if (!negative) {
            assert!(magnitude <= MAX_POSITIVE_MAGNITUDE, EMagnitudeTooLarge);
        } else {
            assert!(magnitude <= MAX_NEGATIVE_MAGNITUDE, EMagnitudeTooLarge);
        }

        // Ensure consistent 0 representation corresponding to twos complements(positive sign)
        if (magnitude == 0) {
            negative = false;
        }
        I128 { 
            negative, 
            magnitude 
        }
    }

    public fun is_negative(i128: &I128): bool {
        i128.negative
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
        let negative = (value >> 127) == 1;
        if (!negative) {
            // if positive, keep the value as is
            new(value, false)
        } else {
            // if negative, take twos complement
            value = value ^ 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF + 1;
            new(value, true)
        }
    }

    #[test]
    fun test_max_positive_magnitude() {
        let max_positive_magnitude = new(0x7FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF, false);
        assert!(max_positive_magnitude.negative, false)
        assert!(max_positive_magnitude.magnitude == MAX_POSITIVE_MAGNITUDE, true);
        assert!(&new(1<<127 -1, false) == &from_u128(1<<127 -1), true);
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
        assert!(max_negative_magnitude.negative, true);
        assert!(max_negative_magnitude.magnitude == MAX_NEGATIVE_MAGNITUDE, true);
        assert!(&new(1<<127, true) == &from_u128(1<<127), true);
    }

    #[test]
    #[expected_failure(abort_code = 0)]
    fun test_magnitude_too_large_negative() {
        let magnitude_too_large_negative = 0x80000000000000000000000000000001;
        new(magnitude_too_large_negative, true);
    }

    #[test]
    fun test_is_negative() {
        assert!(is_negative(&new(1, false)), false);
        assert!(is_negative(&new(1, true)), true);
    }

    #[test]
    fun test_get_magnitude_if_negative() {
        assert!(get_magnitude_if_negative(&new(1, true)) == 1, true);
        assert!(get_magnitude_if_negative(&new(1, false)) == 0, true);
    }

    #[test]
    fun test_get_magnitude_if_positive() {
        assert!(get_magnitude_if_positive(&new(1, false)) == 1, true);
        assert!(get_magnitude_if_positive(&new(1, true)) == 0, true);
    }

    #[test]
    fun test_from_u128() {
        assert!(&new(1, false) == &from_u128(1), true);
        assert!(&new(1, true) == &from_u128(1<<127), true);
    }

    #[test]
    fun test_single_representation_of_zero() {
        assert!(&new(0, false) == &from_u128(0), true);
        assert!(&new(0, true) == &from_u128(0), true);
        let zero_positive = new(0, false);
        let zero_negative = new(0, true);
        assert!(&zero_positive == &zero_negative, true);
        assert!(is_negative(&zero_positive), false);
        assert!(is_negative(&zero_negative), false);
    }




}
