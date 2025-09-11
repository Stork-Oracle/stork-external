//! Provides the [`TemporalNumericValue`] and [`TemporalNumericValueFeed`] account structs.
//! These structs are used to store the latest price update for an asset.

use {
    crate::error::GetTemporalNumericValueError,
    anchor_lang::prelude::{borsh::BorshSchema, *},
};

pub type FeedId = [u8; 32];

/// A struct representing a timestamped value.
#[derive(AnchorSerialize, AnchorDeserialize, Clone, Default, BorshSchema)]
pub struct TemporalNumericValue {
    /// The unix timestamp of the value in nanoseconds.
    pub timestamp_ns: u64,
    /// The quantized value.
    pub quantized_value: i128,
}

impl TemporalNumericValue {
    /// The length of a [`TemporalNumericValue`] struct.
    pub const LEN: usize = 8 + 16;
}

/// An account struct representing a Stork price feed.
#[account]
#[derive(BorshSchema)]
pub struct TemporalNumericValueFeed {
    /// The encoded ID of the asset associated with the feed.
    pub id: FeedId,
    /// The latest canonical temporal numeric value for the feed.
    pub latest_value: TemporalNumericValue,
}

impl TemporalNumericValueFeed {
    // 32 for the id
    // doubled to leave space for future fields
    /// The length of a [`TemporalNumericValueFeed`] struct.
    pub const LEN: usize = 2 * (32 + TemporalNumericValue::LEN);

    /// Gets the latest canonical temporal numeric value for a feed.
    pub fn get_latest_canonical_temporal_numeric_value_unchecked(
        &self,
        feed_id: &FeedId,
    ) -> Result<TemporalNumericValue> {
        require!(
            self.id == *feed_id,
            GetTemporalNumericValueError::InvalidFeedId
        );
        Ok(self.latest_value.clone())
    }
}

/// Implements the [`TryFrom`] trait for [`AccountInfo`] to convert a reference to an [`AccountInfo`] representing a [`TemporalNumericValueFeed`] into a [`TemporalNumericValueFeed`] struct.
impl<'info> TryFrom<&AccountInfo<'info>> for TemporalNumericValueFeed {
    type Error = anchor_lang::error::Error;
    fn try_from(info: &AccountInfo<'info>) -> Result<Self> {
        let data: &[u8] = &info.data.borrow();
        Self::try_deserialize(&mut data.as_ref()).map_err(|e| {
            msg!("Failed to deserialize TemporalNumericValueFeed: {}", e);
            GetTemporalNumericValueError::DeserializationError.into()
        })
    }
}

/// Implements the [`TryFrom`] trait for [`AccountInfo`] to convert an [`AccountInfo`] representing a [`TemporalNumericValueFeed`] into a [`TemporalNumericValueFeed`] struct.
impl<'info> TryFrom<AccountInfo<'info>> for TemporalNumericValueFeed {
    type Error = anchor_lang::error::Error;
    fn try_from(info: AccountInfo<'info>) -> Result<Self> {
        let data: &[u8] = &info.data.borrow();
        Self::try_deserialize(&mut data.as_ref()).map_err(|e| {
            msg!("Failed to deserialize TemporalNumericValueFeed: {}", e);
            GetTemporalNumericValueError::DeserializationError.into()
        })
    }
}
