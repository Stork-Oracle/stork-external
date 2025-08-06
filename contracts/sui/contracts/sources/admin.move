module stork::admin {

    // === Structs ===

    public struct AdminCap has key{
        id: UID,
    }

    // === Init ===

    fun init(ctx: &mut TxContext) {
        transfer::transfer(
            new(ctx),
            ctx.sender()
        );
    }

    fun new(ctx: &mut TxContext): AdminCap {
        AdminCap { id: object::new(ctx) }
    }

    // === Test Helpers ===

    #[test_only]
    public fun test_init(ctx: &mut TxContext) {
        init(ctx);
    }
}