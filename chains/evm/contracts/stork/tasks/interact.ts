import { scope } from 'hardhat/config';
import { loadContractDeploymentAddress } from './utils/helpers';
import { HardhatRuntimeEnvironment } from 'hardhat/types';
import { utils } from 'zksync-ethers';

const initializeContract = async (hre: HardhatRuntimeEnvironment) => {
    const contractAddress = await loadContractDeploymentAddress();
    if (!contractAddress) {
        throw new Error('Contract address not found. Please deploy the contract first.');
    }

    // @ts-expect-error ethers is loaded in hardhat/config
    const [deployer] = await ethers.getSigners();

    // @ts-expect-error artifacts is loaded in hardhat/config
    const contractArtifact = await artifacts.readArtifact('UpgradeableStork');

    // @ts-expect-error ethers is loaded in hardhat/config
    const contract = new ethers.Contract(contractAddress, contractArtifact.abi, deployer);

    console.log(`Network: ${hre.network.name} - ${hre.network.config.chainId}`);
    console.log(`Contract: ${contractAddress}`);

    return contract;
};

const interactScope = scope('interact', 'interact with the contract');

interactScope.task('version', 'Get the contract version').setAction(async (_: any, hre: HardhatRuntimeEnvironment) => {
    const contract = await initializeContract(hre);
    const version = await contract.version();
    console.log(`Contract version: ${version}`);
});

interactScope
    .task('verifyStorkSignatureV1', 'Verify a stork signature')
    .addPositionalParam<string>('assetId', 'The asset id to verify')
    .addPositionalParam<string>('endpoint', 'The endpoint to get the latest update data from')
    .addPositionalParam<string>('authKey', 'The auth key to use to get the latest update data')
    .setAction(async (args: any, hre: HardhatRuntimeEnvironment) => {
        const contract = await initializeContract(hre);
        const storkPubKey = await contract.storkPublicKey();
        const [verify] = await getLatestUpdateData(args.endpoint, args.authKey, args.assetId);
        const returnVal = await contract.verifyStorkSignatureV1(
            storkPubKey,
            verify.id,
            verify.temporalNumericValue.timestampNs,
            verify.temporalNumericValue.quantizedValue,
            verify.publisherMerkleRoot,
            verify.valueComputeAlgHash,
            verify.r,
            verify.s,
            verify.v,
        );
        console.log(returnVal);
    });

const getCustomData = (paymasterAddress: string) => {
    if (!paymasterAddress) {
        return {};
    }

    const params = utils.getPaymasterParams(paymasterAddress, {
        type: 'General',
        innerInput: new Uint8Array(),
    });

    return {
        gasPerPubdata: utils.DEFAULT_GAS_PER_PUBDATA_LIMIT,
        paymasterParams: params,
    };
};

interactScope
    .task('updateTemporalNumericValuesV1', 'Update the temporal numeric values')
    .addPositionalParam<string>('assetIds', 'The asset ids to update')
    .addPositionalParam<string>('endpoint', 'The endpoint to get the latest update data from')
    .addPositionalParam<string>('authKey', 'The auth key to use to get the latest update data')
    .addOptionalParam<string>('paymasterAddress', 'The paymaster to use to update the values')
    .setAction(async (args: any, hre: HardhatRuntimeEnvironment) => {
        const contract = await initializeContract(hre);
        const updates = await getLatestUpdateData(args.endpoint, args.authKey, args.assetIds);

        const customData = getCustomData(args.paymasterAddress);

        await contract.updateTemporalNumericValuesV1.send(updates, {
            value: Object.keys(customData).length === 0 ? updates.length : 0,
            customData,
            gasLimit: 10000000,
        });
    });

interactScope
    .task('getTemporalNumericValueV1', 'Get the temporal numeric value')
    .addPositionalParam<string>('assetId', 'The asset id to get the value for')
    .setAction(async ({ assetId }: { assetId: string }, hre: HardhatRuntimeEnvironment) => {
        const contract = await initializeContract(hre);
        // @ts-expect-error ethers is loaded in hardhat/config
        const encoded = ethers.keccak256(ethers.toUtf8Bytes(assetId));
        const returnVal = await contract.getTemporalNumericValueV1(encoded);
        console.log(returnVal);
    });

interactScope
    .task('getTemporalNumericValueUnsafeV1', 'Get the temporal numeric value (unsafe)')
    .addPositionalParam<string>('assetId', 'The asset id to get the value for')
    .setAction(async ({ assetId }: { assetId: string }, hre: HardhatRuntimeEnvironment) => {
        const contract = await initializeContract(hre);
        // @ts-expect-error ethers is loaded in hardhat/config
        const encoded = ethers.keccak256(ethers.toUtf8Bytes(assetId));
        const returnVal = await contract.getTemporalNumericValueUnsafeV1(encoded);
        console.log(returnVal);
    });

interactScope
    .task('getTemporalNumericValuesUnsafeV1', 'Get the temporal numeric values (unsafe)')
    .addPositionalParam<string>('assetIds', 'The asset ids to get the values for')
    .setAction(async ({ assetIds }: { assetIds: string[] }, hre: HardhatRuntimeEnvironment) => {
        const contract = await initializeContract(hre);
        // @ts-expect-error ethers is loaded in hardhat/config
        const encoded = assetIds.split(',').map((id: string) => ethers.keccak256(ethers.toUtf8Bytes(id.trim())));
        const returnVal = await contract.getTemporalNumericValuesUnsafeV1(encoded);
        console.log(returnVal);
    });

interactScope
    .task('updateValidTimePeriodSeconds', 'Update the valid time period seconds')
    .addPositionalParam<string>('seconds', 'The number of seconds to update the valid time period to')
    .setAction(async ({ seconds }: { seconds: number }, hre: HardhatRuntimeEnvironment) => {
        const contract = await initializeContract(hre);
        await contract.updateValidTimePeriodSeconds(seconds);
    });

interactScope
    .task('updateSingleUpdateFeeInWei', 'Update the single update fee in wei')
    .addPositionalParam<number>('fee', 'The fee to update the single update fee to')
    .setAction(async ({ fee }: { fee: number }, hre: HardhatRuntimeEnvironment) => {
        const contract = await initializeContract(hre);
        await contract.updateSingleUpdateFeeInWei(fee);
    });

interactScope
    .task('updateStorkPublicKey', 'Update the stork public key')
    .addPositionalParam<string>('key', 'The new stork public key')
    .setAction(async ({ key }: { key: string }, hre: HardhatRuntimeEnvironment) => {
        const contract = await initializeContract(hre);
        await contract.updateStorkPublicKey(key);
    });

interactScope
    .task('validTimePeriodSeconds', 'Get the valid time period seconds')
    .setAction(async (_: any, hre: HardhatRuntimeEnvironment) => {
        const contract = await initializeContract(hre);
        const returnVal = await contract.validTimePeriodSeconds();
        console.log(returnVal);
    });

interactScope
    .task('singleUpdateFeeInWei', 'Get the single update fee in wei')
    .setAction(async (_: any, hre: HardhatRuntimeEnvironment) => {
        const contract = await initializeContract(hre);
        const returnVal = await contract.singleUpdateFeeInWei();
        console.log(returnVal);
    });

interactScope
    .task('storkPublicKey', 'Get the stork public key')
    .setAction(async (_: any, hre: HardhatRuntimeEnvironment) => {
        const contract = await initializeContract(hre);
        const returnVal = await contract.storkPublicKey();
        console.log(returnVal);
    });

const getLatestUpdateData = async (endpoint: string, authKey: string, assetIds: string) => {
    const response = await fetch(`${endpoint}/v1/prices/latest?assets=${assetIds}`, {
        headers: {
            Authorization: `Basic ${authKey}`,
        },
    });

    const rawJson = await response.text();
    const safeJsonText = rawJson.replace(
        /(?<!["\d])\b\d{16,}\b(?!["])/g, // Regex to find large integers not already in quotes
        (match: any) => `"${match}"`, // Convert large numbers to strings
    );

    const responseData = JSON.parse(safeJsonText);

    return Object.keys(responseData.data).map((key: any) => {
        const data = responseData.data[key];

        return {
            temporalNumericValue: {
                timestampNs: data.stork_signed_price.timestamped_signature.timestamp,
                quantizedValue: data.stork_signed_price.price,
            },
            id: data.stork_signed_price.encoded_asset_id,
            publisherMerkleRoot: data.stork_signed_price.publisher_merkle_root,
            valueComputeAlgHash: '0x' + data.stork_signed_price.calculation_alg.checksum,
            r: data.stork_signed_price.timestamped_signature.signature.r,
            s: data.stork_signed_price.timestamped_signature.signature.s,
            v: data.stork_signed_price.timestamped_signature.signature.v,
        };
    });
};
