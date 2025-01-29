use anchor_lang::prelude::*;
use anchor_lang::solana_program::system_instruction;

mod verify;
use verify::{verify_stork_evm_signature, EvmPubkey};

use stork_solana_sdk::{
    pda::{STORK_CONFIG_SEED, STORK_FEED_SEED, STORK_TREASURY_SEED},
    temporal_numeric_value::{TemporalNumericValue, TemporalNumericValueFeed},
};

// This needs to match the ID in the stork-sdk crate
declare_id!(stork_solana_sdk::PROGRAM_ID);

#[program]
pub mod stork {
    use super::*;

    pub fn initialize(
        ctx: Context<Initialize>,
        stork_sol_public_key: Pubkey,
        stork_evm_public_key: EvmPubkey,
        single_update_fee_in_lamports: u64,
    ) -> Result<()> {
        let config = &mut ctx.accounts.config;
        config.stork_sol_public_key = stork_sol_public_key;
        config.stork_evm_public_key = stork_evm_public_key;
        config.single_update_fee_in_lamports = single_update_fee_in_lamports;
        config.owner = ctx.accounts.owner.key();
        Ok(())
    }

    pub fn update_temporal_numeric_value_evm(
        ctx: Context<UpdateTemporalNumericValue>,
        update_data: TemporalNumericValueEvmInput,
    ) -> Result<()> {
        let config = &ctx.accounts.config;
        let treasury = &ctx.accounts.treasury;
        let feed = &mut ctx.accounts.feed;
        let payer = &ctx.accounts.payer;

        if feed.id == [0; 32] {
            feed.id = update_data.id;
        }

        // if the new value is older than the latest value, do nothing
        // this does not throw an error to account for multiple instructions per tx
        if feed.latest_value.timestamp_ns >= update_data.temporal_numeric_value.timestamp_ns {
            return Ok(());
        }

        if !verify_stork_evm_signature(
            &config.stork_evm_public_key,
            update_data.id,
            update_data.temporal_numeric_value.timestamp_ns,
            update_data.temporal_numeric_value.quantized_value,
            update_data.publisher_merkle_root,
            update_data.value_compute_alg_hash,
            update_data.r,
            update_data.s,
            update_data.v,
        ) {
            return err!(StorkError::InvalidSignature);
        }

        let amount_to_pay = config.single_update_fee_in_lamports;
        if payer.lamports()
            < Rent::get()?
                .minimum_balance(payer.data_len())
                .saturating_add(amount_to_pay)
        {
            return err!(StorkError::InsufficientFunds);
        };
        let transfer_instruction =
            system_instruction::transfer(payer.key, treasury.key, amount_to_pay);
        anchor_lang::solana_program::program::invoke(
            &transfer_instruction,
            &[payer.to_account_info(), treasury.to_account_info()],
        )?;

        feed.latest_value = update_data.temporal_numeric_value;

        Ok(())
    }

    pub fn update_single_update_fee_in_lamports(
        ctx: Context<AdminUpdate>,
        new_single_update_fee_in_lamports: u64,
    ) -> Result<()> {
        let config = &mut ctx.accounts.config;
        config.single_update_fee_in_lamports = new_single_update_fee_in_lamports;
        Ok(())
    }

    pub fn update_stork_sol_public_key(
        ctx: Context<AdminUpdate>,
        new_stork_sol_public_key: Pubkey,
    ) -> Result<()> {
        let config = &mut ctx.accounts.config;
        config.stork_sol_public_key = new_stork_sol_public_key;
        Ok(())
    }

    pub fn update_stork_evm_public_key(
        ctx: Context<AdminUpdate>,
        new_stork_evm_public_key: EvmPubkey,
    ) -> Result<()> {
        let config = &mut ctx.accounts.config;
        config.stork_evm_public_key = new_stork_evm_public_key;
        Ok(())
    }

}

#[derive(Accounts)]
#[instruction(stork_sol_public_key: Pubkey, stork_evm_public_key: EvmPubkey, single_update_fee_in_lamports: u64)]
pub struct Initialize<'info> {
    #[account(
        init,
        space = StorkConfig::LEN,
        payer = owner,
        seeds = [STORK_CONFIG_SEED],
        bump,
        rent_exempt = enforce
    )]
    pub config: Account<'info, StorkConfig>,
    #[account(mut)]
    pub owner: Signer<'info>,
    pub system_program: Program<'info, System>,
}

#[derive(Accounts)]
#[instruction(update_data: TemporalNumericValueEvmInput)]
pub struct UpdateTemporalNumericValue<'info> {
    #[account(
        seeds = [STORK_CONFIG_SEED],
        bump
    )]
    pub config: Account<'info, StorkConfig>,
    /// CHECK: this just holds the lamports paid by the payer. it is created if needed.
    #[account(
        init_if_needed,
        payer = payer,
        space = 8,
        seeds = [STORK_TREASURY_SEED.as_ref(), &[update_data.treasury_id]],
        bump,
        rent_exempt = enforce
    )]
    pub treasury: AccountInfo<'info>,
    #[account(
        init_if_needed,
        payer = payer,
        space = TemporalNumericValueFeed::LEN,
        seeds = [STORK_FEED_SEED.as_ref(), update_data.id.as_ref()],
        bump,
        rent_exempt = enforce
    )]
    pub feed: Account<'info, TemporalNumericValueFeed>,
    #[account(mut)]
    pub payer: Signer<'info>,
    pub system_program: Program<'info, System>,
}

#[derive(AnchorSerialize, AnchorDeserialize, Clone)]
pub struct TemporalNumericValueEvmInput {
    pub id: [u8; 32],
    pub temporal_numeric_value: TemporalNumericValue,
    pub publisher_merkle_root: [u8; 32],
    pub value_compute_alg_hash: [u8; 32],
    pub r: [u8; 32],
    pub s: [u8; 32],
    pub v: u8,
    pub treasury_id: u8,
}

#[derive(Accounts)]
pub struct AdminUpdate<'info> {
    #[account(
        mut,
        seeds = [STORK_CONFIG_SEED],
        has_one = owner @ StorkError::Unauthorized,
        bump
    )]
    pub config: Account<'info, StorkConfig>,
    pub owner: Signer<'info>,
}


#[account]
pub struct StorkConfig {
    pub stork_sol_public_key: Pubkey,
    pub stork_evm_public_key: EvmPubkey,
    pub single_update_fee_in_lamports: u64,
    pub owner: Pubkey,
}

impl StorkConfig {
    // these are the lengths of the fields in the struct
    // 32 is the length of the pubkey
    // 8 is the length of the u64
    // 8 is the length of the u64
    // 32 is the length of the pubkey
    // 32 + 8 + 8 + 32  = 80 
    // quadruple to leave space for future fields
    pub const LEN: usize = 320;
}

#[error_code]
pub enum StorkError {
    #[msg("Insufficient funds")]
    InsufficientFunds,
    #[msg("Invalid signature")]
    InvalidSignature,
    #[msg("Unauthorized")]
    Unauthorized,
}
