import { Command } from "commander";
import { Provider, Wallet, Address, EvmAddress, bn, Src14OwnedProxy, keccak256} from "fuels";
import fs from "fs";
import { I128Input, Stork, TemporalNumericValueInput, TemporalNumericValueInputInput } from "./types/contracts/Stork";

// contract address, defaults to whats stored in types/contract-ids.json
let STORK_CONTRACT_ADDRESS: string | undefined = process.env.STORK_CONTRACT_ADDRESS;

const contractIds = JSON.parse(fs.readFileSync("types/contract-ids.json", "utf8"));

// if not set, use the default addresses from types/contract-ids.json
if (!STORK_CONTRACT_ADDRESS) {
    console.log("STORK_CONTRACT_ADDRESS not set, using default from types/contract-ids.json");
    STORK_CONTRACT_ADDRESS = contractIds.stork;
}

const PRIVATE_KEY: string | undefined = process.env.PRIVATE_KEY;
const PROVIDER_URL: string | undefined = process.env.PROVIDER_URL;
const STORK_EVM_PUBLIC_KEY: string = process.env.STORK_EVM_PUBLIC_KEY || "0x0a803F9b1CCe32e2773e0d2e98b37E0775cA5d44";
const UPDATE_FEE_IN_WEI: number = parseInt(process.env.UPDATE_FEE_IN_WEI || "1");

function getEvmPubKey(evmPubKey: string): EvmAddress {
    // cut off the 0x prefix
    const evmPubKeyWithout0x = evmPubKey.slice(2);
    if (evmPubKeyWithout0x.length !== 40) {
        throw new Error("Invalid EVM public key");
    }
    return {
        bits: `0x000000000000000000000000${evmPubKeyWithout0x}`
    }
}

function stringToI128Input(value: string): I128Input {
    const valueBn = bn(value);
    const indent = bn(1).shln(127);
    const valueBnWithIndent = valueBn.add(indent);

    const mask64Bits = bn("18446744073709551615"); // 2^64 - 1
    const upper = valueBnWithIndent.shrn(64);
    const lower = valueBnWithIndent.and(mask64Bits);
    
    return {
        underlying: {
            upper: upper.toString(),
            lower: lower.toString()
        }
    };
}

if (!PROVIDER_URL) {
    throw new Error("PROVIDER_URL is not set");
}

const cliProgram = new Command();
cliProgram
    .name("admin")
    .description("Fuel Stork admin client")
    .version("0.1.0");

cliProgram
    .command("initialize")
    .description("Initialize the Stork contract")
    .action(async () => {
        const provider = new Provider(PROVIDER_URL!);
        const wallet = Wallet.fromPrivateKey(PRIVATE_KEY!, provider);
        const storkContract = new Stork(STORK_CONTRACT_ADDRESS!, wallet);

        // construct the call params
        const ownerAddress = new Address(wallet.publicKey);
        const ownerAddressInput = {bits: ownerAddress.toB256()}
        const ownerAddressIdentityInput = {Address: ownerAddressInput}

        const evmPubKey: EvmAddress = getEvmPubKey(STORK_EVM_PUBLIC_KEY);

        const singleUpdateFeeInWei = bn(UPDATE_FEE_IN_WEI);

        try {
            const tx = await storkContract.functions
                .initialize(ownerAddressIdentityInput, evmPubKey, singleUpdateFeeInWei)
                .addContracts([STORK_CONTRACT_ADDRESS!])
                .call();
            console.log("Transaction:", tx);
        } catch (e: any) {
            console.error("Detailed error:", JSON.stringify(e.metadata || e, null, 2));
            throw e;
        }
    });

cliProgram
    .command("write-to-feeds")
    .description("Write to feeds")
    .argument("asset_pairs", "The asset pairs (comma separated)")
    .argument("endpoint", "The REST endpoint")
    .argument("auth_key", "The auth key")
    .action(async (asset_pairs: string, endpoint: string, auth_key: string) => {
        console.log(`Writing to feeds: ${asset_pairs}`);
        const provider = new Provider(PROVIDER_URL!);
        const wallet = Wallet.fromPrivateKey(PRIVATE_KEY!, provider);
        const storkContract = new Stork(STORK_CONTRACT_ADDRESS!, wallet);

        try {
            const result = await fetch(
                `${endpoint}/v1/prices/latest\?assets\=${asset_pairs}`,
                {
                  headers: {
                    Authorization: `Basic ${auth_key}`,
                  },
                }
                );
            const rawJson = await result.text();
            const safeJsonText = rawJson.replace(
                /(?<!["\d])\b\d{16,}\b(?!["])/g, // Regex to find large integers not already in quotes
                (match) => `"${match}"` // Convert large numbers to strings
            );
            const responseData = JSON.parse(safeJsonText);
            // construct update data
            const updateData: TemporalNumericValueInputInput[] = [];
            Object.values(responseData.data).forEach((data: any) => {
                const id: string = data.stork_signed_price.encoded_asset_id;
                const recvTime: string = data.stork_signed_price.timestamped_signature.timestamp;
                const quantizedValue: string = data.stork_signed_price.price;
                const publisherMerkleRoot: string = data.stork_signed_price.publisher_merkle_root;
                const valueComputeAlgHash: string = `0x${data.stork_signed_price.calculation_alg.checksum}`;
                const r: string = data.stork_signed_price.timestamped_signature.signature.r;
                const s: string = data.stork_signed_price.timestamped_signature.signature.s;
                const v: string = data.stork_signed_price.timestamped_signature.signature.v;

                // construct quantized value
                const quantizedValueInput: I128Input = stringToI128Input(quantizedValue);
                // construct temporal numeric value
                const temporalNumericValue: TemporalNumericValueInput = {
                    timestamp_ns: bn(recvTime),
                    quantized_value: quantizedValueInput,
                }
                // convert v hex string to number
                const vNumber = parseInt(v.slice(2), 16);

                const temporalNumericValueInput: TemporalNumericValueInputInput = {
                    temporal_numeric_value: temporalNumericValue,
                    id,
                    publisher_merkle_root: publisherMerkleRoot,
                    value_compute_alg_hash: valueComputeAlgHash,
                    r,
                    s,
                    v: vNumber,
                }
                updateData.push(temporalNumericValueInput);
            });

            // get update fee
            const updateFee = await storkContract.functions.get_update_fee_v1(updateData).get();
            const feeValue = updateFee.value;
            console.log("Update fee:", feeValue.toString());

            // call update_temporal_numeric_values_v1
            const tx = await storkContract
                .functions
                .update_temporal_numeric_values_v1(updateData)
                .callParams({
                    forward: [feeValue, await provider.getBaseAssetId()],
                })
                .call();

            console.log("Transaction:", tx);
            process.exit(0);
            
        } catch (e: any) {
            console.error("Detailed error:", JSON.stringify(e.metadata || e, null, 2));
            throw e;
        }
    });

cliProgram
    .command("read-from-feed")
    .description("Read from feed")
    .argument("asset_id", "The plaintext asset id")
    .action(async (asset_id: string) => {
        let encAssetId = `0x${Buffer.from(keccak256(Buffer.from(asset_id))).toString('hex')}`;
        console.log("Encoded asset id:", encAssetId);
        const provider = new Provider(PROVIDER_URL!);
        const storkContract = new Stork(STORK_CONTRACT_ADDRESS!, provider);
        const temporalNumericValue = await storkContract.functions.get_temporal_numeric_value_unchecked_v1(encAssetId).get();
        
        const value = temporalNumericValue.value;
        const upper = bn(value.quantized_value.underlying.upper);
        const lower = bn(value.quantized_value.underlying.lower);
        const mask64Bits = bn("18446744073709551615"); // 2^64 - 1
        
        // Reconstruct the original number
        const reconstructed = upper.shln(64).add(lower).sub(bn(1).shln(127));
        
        console.log("Temporal numeric value:");
        console.log("  Timestamp:", value.timestamp_ns.toString());
        console.log("  Value:", reconstructed.toString());
    });

cliProgram
    .command("get-state-info")
    .description("Get the state info of the Stork contract")
    .action(async () => {
        const provider = new Provider(PROVIDER_URL!);
        const storkContract = new Stork(STORK_CONTRACT_ADDRESS!, provider);
        const evmPublicKey = await storkContract.functions.stork_public_key().get();
        console.log("EVM public key:", evmPublicKey.value.bits.toString());
        const singleUpdateFeeInWei = await storkContract.functions.single_update_fee_in_wei().get();
        console.log("Single update fee in wei:", singleUpdateFeeInWei.value.toString());
        const owner = await storkContract.functions.owner().get();
        console.log("Owner:", owner.value);
    });


cliProgram
    .command("get-proxy-target")
    .description("Gets the target of the proxy contract")
    .action(async () => {
        const provider = new Provider(PROVIDER_URL!);
        const storkContract = new Src14OwnedProxy(STORK_CONTRACT_ADDRESS!, provider);
        const proxyTarget = await storkContract.functions.proxy_target().get();
        console.log("Proxy target:", proxyTarget);
    });

cliProgram
    .command("update-stork-public-key")
    .description("Updates the stork public key")
    .argument("evm_public_key", "The EVM public key")
    .action(async (evm_public_key: string) => {
        // add 0x prefix if not present
        if (!evm_public_key.startsWith("0x")) {
            evm_public_key = `0x${evm_public_key}`;
        }
        const provider = new Provider(PROVIDER_URL!);
        const wallet = Wallet.fromPrivateKey(PRIVATE_KEY!, provider);
        const storkContract = new Stork(STORK_CONTRACT_ADDRESS!, wallet);
        const evmPubKey: EvmAddress = getEvmPubKey(evm_public_key);
        const tx = await storkContract.functions.update_stork_public_key(evmPubKey).call();
        console.log("Transaction:", tx);
    });

cliProgram
    .command("update-single-update-fee-in-wei")
    .description("Updates the single update fee in wei")
    .argument("single_update_fee_in_wei", "The single update fee in wei")
    .action(async (single_update_fee_in_wei: string) => {
        const provider = new Provider(PROVIDER_URL!);
        const wallet = Wallet.fromPrivateKey(PRIVATE_KEY!, provider);
        const storkContract = new Stork(STORK_CONTRACT_ADDRESS!, wallet);
        const tx = await storkContract.functions.update_single_update_fee_in_wei(bn(single_update_fee_in_wei)).call();
        console.log("Transaction:", tx);
    });

cliProgram.parse(process.argv);
