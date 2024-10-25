/*
* A minimal interface for the Stork contract including only
*/
interface IStorkTemporalNumericValueUnsafeGetter {
    function getTemporalNumericValueUnsafeV1(
        bytes32 id
    ) public view returns (StorkStructs.TemporalNumericValue memory value);
}