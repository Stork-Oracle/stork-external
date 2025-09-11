# Stork Cosmwasm Core No Dep Example

This is a very simple cosmwasm contract to show how you would use the Stork Cosmwasm Contract to consume Stork price updates in your cosmwasm contract without depending on the stork-cw crate. This is useful if your cosmwasm contract is using a different version of cosmwasm-std than the stork-cw crate. To demonstrate this, we use cosmwasm-std 1.5.7, where the stork-cw crate uses cosmwasm-std ^2.2.2 (via Sylvia).

See [README.md](../README.md) for more information.
