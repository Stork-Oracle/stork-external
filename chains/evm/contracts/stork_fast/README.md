# Sample Hardhat 3 Beta Project (`node:test` and `viem`)

This directory contains a [Hardhat](https://hardhat.org/docs) project for the Stork Fast EVM compatible contract.

This contract is used to verify and deserialize Stork Fast signed ECDSA payloads.

In order to verify a signed ECDSA payload, use the `verifySignedECDSAPayload` function.

To also deserialize the update values from the signed ECDSA payload, use the `verifyAndDeserializeSignedECDSAPayload` function.

### Getting started

```bash
npm i
```

### Running Tests

```bash
npx hardhat test
```

### Local Development

#### Run local node

```bash
npx hardhat node
```

#### Deploy

```bash
npx hardhat compile
npx hardhat ignition deploy ignition/modules/StorkFast.ts --network hardhatLocal --deployment-id chain-31337-0 --parameters ignition/parameters.json
```

Note: the `--parameters` flag is required — the module has no defaults for `signerAddresses` and `verificationFeeInWei`, and Ignition does not auto-load `ignition/parameters.json`. Parameters must be keyed under `StorkFastProxyModule` (the module that reads them), as in the checked-in file. `--verify` is omitted here since Etherscan verification does not work against a local node.

#### Deployment id convention

Always pass an explicit `--deployment-id` of the form `chain-<chainId>-<n>`: `-0` for the initial deploy, incrementing for each subsequent upgrade (`chain-31337-0`, `chain-31337-1`, ...). This keeps the `ignition/deployments/` directory uniformly named and chronologically ordered per chain. (Without the flag, Ignition defaults to plain `chain-<chainId>`, which wouldn't line up with the upgrade entries.)

#### Interact

See `tasks/admin.ts` for available methods.

```bash
npx hardhat --network hardhatLocal <method>
```

### Deploy on-chain

1. Configure your `hardhat.config.ts` with the desired network.
2. Store the required secrets in the Hardhat keystore: the network's deployer key (e.g. `npx hardhat keystore set DEPLOYER_PRIVATE_KEY`) and `npx hardhat keystore set ETHERSCAN_API_KEY` (needed for `--verify`).
3. Update the `ignition/parameters.json` file with the desired signer addresses and verification fee in wei (keyed under `StorkFastProxyModule`).
4. Run `npx hardhat ignition deploy ignition/modules/StorkFast.ts --network <network> --deployment-id chain-<chainId>-0 --parameters ignition/parameters.json --verify` to deploy the contract.
5. Deployment will be saved in the `ignition/deployments/chain-<chainId>-0` directory.

If Etherscan verification fails (or `--verify` was omitted), re-run it without redeploying:

```bash
npx hardhat ignition verify --network <network> chain-<chainId>-0
```

### Upgrade

The contract uses the UUPS proxy pattern (`ERC1967Proxy` + `UpgradeableStorkFast`): an upgrade deploys a new implementation contract and calls `upgradeToAndCall` on the proxy, which is authorized via `_authorizeUpgrade` (`onlyOwner`). The proxy address — the one consumers use — never changes.

1. Make your contract changes and bump the string returned by `version()` in `contracts/StorkFast.sol`, so the upgrade can be confirmed afterwards.
2. Create a parameters file with the proxy address to upgrade (found in `ignition/deployments/chain-<chainId>-0/deployed_addresses.json` under `StorkFastProxyModule#ERC1967Proxy`):

   ```json
   {
     "StorkFastUpgradeModule": {
       "proxyAddress": "0x..."
     }
   }
   ```

3. Run the upgrade module. The deploying account (account 0 of the network config) must be the contract owner. Use a fresh `--deployment-id` for every upgrade, following the `chain-<chainId>-<n>` convention (next unused number in `ignition/deployments/`) — Ignition journals executed steps per deployment id, so re-running under a previous one would silently skip the upgrade:

   ```bash
   npx hardhat compile
   npx hardhat ignition deploy ignition/modules/StorkFastUpgrade.ts --network <network> --deployment-id chain-<chainId>-1 --parameters <parameters-file> --verify
   ```

   The new implementation address is recorded under `ignition/deployments/<deployment-id>/`.

4. Confirm the proxy serves the new version:

   ```bash
   npx hardhat --network <network> version <proxyAddress>
   ```

Note: the implementation's `initialize` is only called on first deploy; the upgrade passes empty calldata to `upgradeToAndCall`, so state (owner, signers, fee) is preserved. If a future version needs a migration call, add a reinitializer to the contract and pass its encoded call instead of `"0x"` in `ignition/modules/StorkFastUpgrade.ts`.
