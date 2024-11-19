module stork::admin {

    // === Structs ===

    public struct AdminCap has key{
        id: UID,
    }

    // === Init ===

    fun init(ctx: &mut TxContext) {
        transfer::transfer(
            AdminCap { id: object::new(ctx) },
            ctx.sender()
        );
    }

    // === Public Functions ===

    
}