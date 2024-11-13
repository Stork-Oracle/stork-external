module stork::stork {

    // === Imports ===

    use stork::admin::AdminCap;

    // === Errors ===

    // === Structs ===

    


    fun init(ctx: &mut TxContext) {
        let admin_capability = AdminCapability {
            id: object::new(ctx),
        };

        transfer::transfer(admin_capability, tx_context::sender(ctx));
    }
}
