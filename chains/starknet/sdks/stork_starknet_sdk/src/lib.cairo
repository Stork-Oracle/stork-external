//! Stork Starknet SDK.
//!
//! Depend on this to read Stork price feeds from a Cairo contract. See
//! [`interface::IStorkDispatcher`] for the calls a consumer makes, and
//! [`temporal_numeric_value::TemporalNumericValue`] for what they return.

pub mod errors;
pub mod events;
pub mod interface;
pub mod temporal_numeric_value;
