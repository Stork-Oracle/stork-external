library;

use ::temporal_numeric_value::TemporalNumericValue;

pub enum StorkEvent {
    ValueUpdate: (b256, TemporalNumericValue),
    Initialized: (),
}
