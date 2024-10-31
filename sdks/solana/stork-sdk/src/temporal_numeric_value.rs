use {
    crate::error::GetTemporalNumericValueError,
    anchor_lang::prelude::{
        borsh::BorshSchema,
        *,
    },
};

pub type FeedId = [u8; 32];

#[derive(AnchorSerialize, AnchorDeserialize, Clone, Default, BorshSchema)]
pub struct TemporalNumericValue {
    pub timestamp_ns: u64,
    pub quantized_value: i128,
}

impl TemporalNumericValue {
    pub const LEN: usize = 8 + 16;
}

#[account]
#[derive(BorshSchema)]
pub struct TemporalNumericValueFeed {
    pub id: FeedId,
    pub latest_value: TemporalNumericValue,
}

impl TemporalNumericValueFeed {
    // 32 for the id
    // doubled to leave space for future fields
    pub const LEN: usize = 2 * (32 + TemporalNumericValue::LEN);

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
