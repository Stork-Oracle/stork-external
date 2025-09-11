use cw_storage_plus::Item;
use sylvia::contract;
use sylvia::ctx::{ExecCtx, InstantiateCtx};
use sylvia::cw_std::Empty;
use sylvia::cw_std::{Response, StdResult};
use sylvia::types::{CustomMsg, CustomQuery};
use sylvia::types::Remote;
use stork_cw::contract::StorkContract;
use stork_cw::temporal_numeric_value::EncodedAssetId;
use stork_cw::contract::sv::Querier;
use sylvia::cw_std::Addr;
use sylvia::cw_std::Event;
pub struct ExampleContract<E, Q> {
    stork: Item<Remote<'static, StorkContract<E, Q>>>,
    _phantom: std::marker::PhantomData<(E, Q)>,
}

#[cfg_attr(not(feature = "library"), sylvia::entry_points(generics<Empty, Empty>))]
#[contract]
#[sv::custom(msg = E, query = Q)]
impl<E, Q> ExampleContract<E, Q>
where
    E: CustomMsg + 'static,
    Q: CustomQuery + 'static,
{
    pub const fn new() -> Self {
        Self {
            stork: Item::new("stork"),
            _phantom: std::marker::PhantomData,
        }
    }

    #[sv::msg(instantiate)]
    fn instantiate(&self, ctx: InstantiateCtx<Q>, stork_contract_address: Addr) -> StdResult<Response<E>> {
        let stork_contract = Remote::new(stork_contract_address);
        self.stork.save(ctx.deps.storage, &stork_contract)?;
        Ok(Response::new())
    }

    #[sv::msg(exec)]
    fn use_stork_price(&self, ctx: ExecCtx<Q>, feed_id: EncodedAssetId) -> StdResult<Response<E>> {
        
        // get the price from stork

        let temporal_numeric_value = self.stork
            .load(ctx.deps.storage)?
            .querier(&ctx.deps.querier)
            .get_latest_canonical_temporal_numeric_value_unchecked(feed_id)?
            .temporal_numeric_value;

        // Do something with the price
        let feed_id_hex = feed_id.iter().map(|b| format!("{:02x}", b)).collect::<Vec<String>>().join("");
        let event = Event::new("stork_price_used").add_attribute("feed_id", feed_id_hex).add_attribute("value", temporal_numeric_value.quantized_value.to_string());
        Ok(Response::new().add_events(vec![event]))
    }

}
