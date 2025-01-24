use crate::{
    error::StorkError,
    responses::{
        GetOwnerResponse, GetSingleUpdateFeeResponse, GetStorkEvmPublicKeyResponse,
        GetTemporalNumericValueResponse,
    },
    temporal_numeric_value::{EncodedAssetId, TemporalNumericValue},
    verify::{verify_stork_evm_signature, EvmPubkey},
    events::{new_stork_init_event, new_temporal_numeric_value_update_event},
};
use cw_storage_plus::{Item, Map};
#[cfg(not(feature = "library"))]
use sylvia::cw_std::Empty;
use sylvia::{
    contract,
    ctx::{ExecCtx, InstantiateCtx, MigrateCtx, QueryCtx},
    cw_schema::cw_serde,
    cw_std::{coin, has_coins, Addr, Coin, Event, Response, StdResult},
    types::{CustomMsg, CustomQuery},
};

pub struct StorkContract<E, Q> {
    pub stork_evm_public_key: Item<EvmPubkey>,
    pub single_update_fee: Item<Coin>,
    pub owner: Item<Addr>,
    pub temporal_numeric_value_feed_registry: Map<EncodedAssetId, TemporalNumericValue>,
    _phantom: std::marker::PhantomData<(E, Q)>,
}

#[cfg_attr(not(feature = "library"), sylvia::entry_points(generics<Empty, Empty>))]
#[contract]
#[sv::error(StorkError)]
#[sv::custom(msg = E, query = Q)]
impl<E, Q> StorkContract<E, Q>
where
    E: CustomMsg + 'static,
    Q: CustomQuery + 'static,
{
    pub const fn new() -> Self {
        Self {
            stork_evm_public_key: Item::new("stork_evm_public_key"),
            single_update_fee: Item::new("single_update_fee"),
            owner: Item::new("owner"),
            temporal_numeric_value_feed_registry: Map::new("temporal_numeric_value_feed_registry"),
            _phantom: std::marker::PhantomData,
        }
    }

    #[sv::msg(instantiate)]
    fn instantiate(
        &self,
        ctx: InstantiateCtx<Q>,
        stork_evm_public_key: EvmPubkey,
        single_update_fee: Coin,
    ) -> StdResult<Response<E>> {
        self.stork_evm_public_key
            .save(ctx.deps.storage, &stork_evm_public_key)?;
        self.single_update_fee
            .save(ctx.deps.storage, &single_update_fee)?;
        self.owner.save(ctx.deps.storage, &ctx.info.sender)?;
        Ok(Response::new().add_event(new_stork_init_event(
            stork_evm_public_key,
            single_update_fee,
            ctx.info.sender,
        )))
    }

    #[sv::msg(exec)]
    fn update_temporal_numeric_values_evm(
        &self,
        ctx: ExecCtx<Q>,
        update_data: Vec<UpdateData>,
    ) -> Result<Response<E>, StorkError> {
        let fee = self.single_update_fee.load(ctx.deps.storage)?;
        let stork_evm_public_key = self.stork_evm_public_key.load(ctx.deps.storage)?;
        let mut num_updates: u128 = 0;
        let api = ctx.deps.api;
        let mut events: Vec<Event> = Vec::new();
        for update in update_data {
            // recency
            if let Some(feed) = self
                .temporal_numeric_value_feed_registry
                .may_load(ctx.deps.storage, update.id)?
            {
                if feed.timestamp_ns >= update.temporal_numeric_value.timestamp_ns {
                    continue;
                }
            }

            // signature
            match verify_stork_evm_signature(
                api,
                &stork_evm_public_key,
                update.id,
                update.temporal_numeric_value.timestamp_ns.u64(),
                update.temporal_numeric_value.quantized_value.i128(),
                update.publisher_merkle_root,
                update.value_compute_alg_hash,
                update.r,
                update.s,
                update.v,
            ) {
                Ok(true) => {
                    // update feed
                    self.temporal_numeric_value_feed_registry.save(
                        ctx.deps.storage,
                        update.id,
                        &update.temporal_numeric_value,
                    )?;
                    num_updates += 1;
                    events.push(new_temporal_numeric_value_update_event(
                        update.id,
                        update.temporal_numeric_value,
                    ));
                }
                Err(e) => {
                    return Err(StorkError::InvalidSignature(e.to_string()));
                }
                Ok(false) => {
                    return Err(StorkError::InvalidSignature("Invalid signature".to_string()));
                }
            }
        }
        //ensure sender has enough funds
        let fee_amount: u128 = fee.amount.u128() * num_updates;
        let fee_denom = fee.denom;
        let total_fee = coin(fee_amount, fee_denom);
        let funds = ctx.info.funds;
        if !has_coins(funds.as_ref(), &total_fee) {
            return Err(StorkError::InsufficientFunds);
        }
        Ok(Response::new().add_events(events))
    }

    #[sv::msg(query)]
    fn get_latest_canonical_temporal_numeric_value_unchecked(
        &self,
        ctx: QueryCtx<Q>,
        id: EncodedAssetId,
    ) -> Result<GetTemporalNumericValueResponse, StorkError> {
        let temporal_numeric_value = self
            .temporal_numeric_value_feed_registry
            .may_load(ctx.deps.storage, id)?
            .ok_or(StorkError::FeedNotFound)?;
        Ok(GetTemporalNumericValueResponse {
            temporal_numeric_value,
        })
    }

    #[sv::msg(query)]
    fn get_single_update_fee(
        &self,
        ctx: QueryCtx<Q>,
    ) -> Result<GetSingleUpdateFeeResponse, StorkError> {
        let fee = self.single_update_fee.load(ctx.deps.storage)?;
        Ok(GetSingleUpdateFeeResponse { fee })
    }

    #[sv::msg(query)]
    fn get_stork_evm_public_key(
        &self,
        ctx: QueryCtx<Q>,
    ) -> Result<GetStorkEvmPublicKeyResponse, StorkError> {
        let stork_evm_public_key = self.stork_evm_public_key.load(ctx.deps.storage)?;
        Ok(GetStorkEvmPublicKeyResponse {
            stork_evm_public_key,
        })
    }

    #[sv::msg(query)]
    fn get_owner(&self, ctx: QueryCtx<Q>) -> Result<GetOwnerResponse, StorkError> {
        let owner = self.owner.load(ctx.deps.storage)?;
        Ok(GetOwnerResponse { owner })
    }

    // Admin functions
    #[sv::msg(exec)]
    fn set_single_update_fee(&self, ctx: ExecCtx<Q>, fee: Coin) -> Result<Response<E>, StorkError> {
        let owner = self.owner.load(ctx.deps.storage)?;
        if ctx.info.sender != owner {
            return Err(StorkError::NotAuthorized);
        }
        self.single_update_fee.save(ctx.deps.storage, &fee)?;
        Ok(Response::new())
    }

    #[sv::msg(exec)]
    fn set_stork_evm_public_key(
        &self,
        ctx: ExecCtx<Q>,
        stork_evm_public_key: EvmPubkey,
    ) -> Result<Response<E>, StorkError> {
        let owner = self.owner.load(ctx.deps.storage)?;
        if ctx.info.sender != owner {
            return Err(StorkError::NotAuthorized);
        }
        self.stork_evm_public_key
            .save(ctx.deps.storage, &stork_evm_public_key)?;
        Ok(Response::new())
    }

    #[sv::msg(exec)]
    fn set_owner(&self, ctx: ExecCtx<Q>, owner: Addr) -> Result<Response<E>, StorkError> {
        let current_owner = self.owner.load(ctx.deps.storage)?;
        if ctx.info.sender != current_owner {
            return Err(StorkError::NotAuthorized);
        }
        self.owner.save(ctx.deps.storage, &owner)?;
        Ok(Response::new())
    }

    #[sv::msg(migrate)]
    fn migrate(&self, _ctx: MigrateCtx<Q>) -> Result<Response<E>, StorkError> {
        Ok(Response::new())
    }
}

#[cw_serde(crate = "sylvia")]
pub struct UpdateData {
    pub id: EncodedAssetId,
    pub temporal_numeric_value: TemporalNumericValue,
    pub publisher_merkle_root: [u8; 32],
    pub value_compute_alg_hash: [u8; 32],
    pub r: [u8; 32],
    pub s: [u8; 32],
    pub v: u8,
}
