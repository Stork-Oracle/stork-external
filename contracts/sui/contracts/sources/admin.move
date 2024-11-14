module stork::admin {

    // === Structs ===

    public struct AdminCap has key, store{
        id: UID,
    }

    // === Init Function ===

    fun init(ctx: &mut TxContext) {
        transfer::transfer(
            AdminCap { id: object::new(ctx) },
            ctx.sender()
        );
    }
}