module stork::fee {

    // === Imports ===

    use initia_std::string::String;
    
    // === Structs ===

    /// The Fee struct
    struct Fee has copy, drop, store {
        amount: u64,
        denom: String,
    }

    // === Functions ===

    /// Creates a new fee
    public fun new(amount: u64, denom: String): Fee {
        Fee { amount, denom }
    }


    /// Returns the amount of the fee
    public fun get_amount(self: &Fee): u64 {
        self.amount
    }

    /// Returns the denom of the fee
    public fun get_denom(self: &Fee): String {
        self.denom
    }

    // === Test Imports ===

    #[test_only]
    use initia_std::string;

    // === Tests ===

    #[test]
    fun test_new() {
        let amount = 1;
        let denom = string::utf8(b"unit");
        let fee = new(amount, denom);
        assert!(fee.amount == 1, 0);
        assert!(fee.denom == string::utf8(b"unit"), 1);
    }

    #[test]
    fun test_get_amount() {
        let amount = 1;
        let denom = string::utf8(b"unit");
        let fee = new(amount, denom);
        assert!(get_amount(&fee) == 1, 0);
    }

    #[test]
    fun test_get_denom() {
        let amount = 1;
        let denom = string::utf8(b"unit");
        let fee = new(amount, denom);
        assert!(get_denom(&fee) == string::utf8(b"unit"), 0);
    }
}
