import { task } from 'hardhat/config';
import { CONTRACT_DEPLOYMENT, createFileIfNotExists, getDeployedAddressesPath } from './utils/helpers';

// before calling this, make sure you compile the contracts using `npx hardhat --network sophonTestnet compile --force`
// sophon paymaster address - 0x98546B226dbbA8230cf620635a1e4ab01F6A99B2

const STORK_PUBLIC_KEY = '0x0a803F9b1CCe32e2773e0d2e98b37E0775cA5d44';
const VALID_TIMEOUT_SECONDS = 3600;
const UPDATE_FEE_IN_WEI = 1;

async function main({ storkPublicKey, paymasterAddress }: { storkPublicKey: string; paymasterAddress: string }) {
    // @ts-expect-error ethers is loaded in hardhat/config
    const [deployer] = await zksyncEthers.getSigners();

    // @ts-expect-error artifacts is loaded in hardhat/config
    const UpgradeableStork = await zksyncEthers.getContractFactory('UpgradeableStork');

    let deployParams = {};
    if (paymasterAddress) {
        // @ts-expect-error zksyncEthers is loaded in hardhat/config
        const params = zksyncEthers.utils.getPaymasterParams(
            paymasterAddress,
            {
                type: 'General',
                innerInput: new Uint8Array(),
            },
        );
        deployParams = {
            paymasterProxyParams: params,
            paymasterImplParams: params,
        };
    }

    // @ts-expect-error zkUpgrades is loaded in hardhat/config
    const upgradeableStork = await zkUpgrades.deployProxy(
        UpgradeableStork,
        [deployer.address, storkPublicKey ?? STORK_PUBLIC_KEY, VALID_TIMEOUT_SECONDS, UPDATE_FEE_IN_WEI],
        {
            initializer: 'initialize',
            kind: 'uups',
            ...deployParams,
        },
    );

    await upgradeableStork.waitForDeployment();

    //   // write file to store the address at deployments/chain-<chainid>/deployed_addresses.json
    const deployedAddressPath = await getDeployedAddressesPath();
    createFileIfNotExists(
        deployedAddressPath,
        JSON.stringify({ [CONTRACT_DEPLOYMENT]: upgradeableStork.target }, null, 2),
    );

    console.log('UpgradeableStork deployed to:', upgradeableStork.target);
}

task('deploy-zk', 'A task to deploy the contract on zk')
    .setDescription('Very similar to deploy, but for zk')
    .addOptionalPositionalParam('storkPublicKey', 'The public key of the Stork contract')
    .addOptionalParam('paymasterAddress', 'The address of the paymaster')
    .setAction(main);
