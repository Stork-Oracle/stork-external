import { Command } from "commander";
import { Account, Aptos, AptosConfig, Network, Ed25519PrivateKey, PrivateKey, PrivateKeyVariants} from "@aptos-labs/ts-sdk";

const DEFAULT_CONTRACT_ADDRESS = process.env.STORK_CONTRACT_ADDRESS;
const DEFAULT_STORK_EVM_PUBLIC_KEY = "0x0a803F9b1CCe32e2773e0d2e98b37E0775cA5d44";
const DEFAULT_UPDATE_FEE_IN_OCTAS = 1;
const PRIVATE_KEY: string | undefined = process.env.PRIVATE_KEY;

const APTOS_CONFIG = new AptosConfig({
    network: process.env.RPC_ALIAS as Network,
});

const aptos = new Aptos(APTOS_CONFIG);

type UpdateData = {
    ids: number[][];
    temporal_numeric_value_timestamp_nss: bigint[];
    temporal_numeric_value_magnitudes: bigint[];
    temporal_numeric_value_negatives: boolean[];
    publisher_merkle_roots: number[][];
    value_compute_alg_hashes: number[][];
    rs: number[][];
    ss: number[][];
    vs: number[];
}

function getAccount() {
    if (!PRIVATE_KEY) {
        throw new Error("PRIVATE_KEY is not set");
    }
    
    const formattedKey = PrivateKey.formatPrivateKey(PRIVATE_KEY, PrivateKeyVariants.Ed25519);
    const privateKey = new Ed25519PrivateKey(formattedKey);
    const account = Account.fromPrivateKey({ privateKey, legacy: true });
    return account;
}

function hexStringToByteArray(hexString: string) {
    if (hexString.startsWith("0x")) {
      hexString = hexString.slice(2);
    }
    return Array.from(Buffer.from(hexString, "hex"));
}

const cliProgram = new Command();
cliProgram
    .name("admin")
    .description("Aptos Stork admin client")
    .version("0.1.0");

cliProgram
    .command("initialize")
    .description("Initialize the Stork contract")
    .action(async () => {
        const account = getAccount();
        const contractAddress = DEFAULT_CONTRACT_ADDRESS;
        const tx = await aptos.transaction.build.simple({
            sender: account.accountAddress,
            data: {
                function: `${contractAddress}::stork::init_stork`,
                functionArguments: [
                    hexStringToByteArray(DEFAULT_STORK_EVM_PUBLIC_KEY),
                    DEFAULT_UPDATE_FEE_IN_OCTAS,
                ],
            },
        });
        const senderAuthenticator = aptos.transaction.sign({
            signer: account,
            transaction: tx,
        });

        const committedTransaction = await aptos.transaction.submit.simple({
            transaction: tx,
            senderAuthenticator,
        });

        const executedTransaction = await aptos.waitForTransaction({ transactionHash: committedTransaction.hash });
        if (executedTransaction.success) {
            console.log(`Transaction succeeded: ${committedTransaction.hash}`);
        } else {
            console.error(`Transaction failed: ${committedTransaction.hash}`);
        }
    });

cliProgram
    .command("write-to-feeds")
    .description("Write to feeds")
    .argument("asset_pairs", "Comma separated list of asset pairs to write to")
    .argument("endpoint", "The stork REST endpoint")
    .argument("auth_key", "The stork auth key")
    .action(async (assetPairs: string, endpoint: string, authKey: string) => {
        const account = getAccount();
        const contractAddress = DEFAULT_CONTRACT_ADDRESS;

        const result = await fetch(`${endpoint}/v1/prices/latest\?assets=${assetPairs}`, 
            {
            headers: {
                "Authorization": `Basic ${authKey}`,
            },
        });

        
        const rawJson = await result.text();
        const safeJsonText = rawJson.replace(
            /(?<!["\d])\b\d{16,}\b(?!["])/g, // Regex to find large integers not already in quotes
            (match) => `"${match}"` // Convert large numbers to strings
            ); 
        
        const response = JSON.parse(safeJsonText);
        const updateData: UpdateData = {
            ids: [],
            temporal_numeric_value_timestamp_nss: [],
            temporal_numeric_value_magnitudes: [],
            temporal_numeric_value_negatives: [],
            publisher_merkle_roots: [],
            value_compute_alg_hashes: [],
            rs: [],
            ss: [],
            vs: [],
        }

        Object.values(response.data).forEach((data: any) => {
            updateData.ids.push(hexStringToByteArray(data.stork_signed_price.encoded_asset_id));
            updateData.temporal_numeric_value_timestamp_nss.push(BigInt(data.stork_signed_price.timestamped_signature.timestamp));
            updateData.temporal_numeric_value_magnitudes.push(BigInt(data.stork_signed_price.price));
            updateData.temporal_numeric_value_negatives.push(data.stork_signed_price.price < 0);
            updateData.publisher_merkle_roots.push(hexStringToByteArray(data.stork_signed_price.publisher_merkle_root));
            updateData.value_compute_alg_hashes.push(hexStringToByteArray(data.stork_signed_price.calculation_alg.checksum));
            updateData.rs.push(hexStringToByteArray(data.stork_signed_price.timestamped_signature.signature.r));
            updateData.ss.push(hexStringToByteArray(data.stork_signed_price.timestamped_signature.signature.s));
            updateData.vs.push(hexStringToByteArray(data.stork_signed_price.timestamped_signature.signature.v)[0]);
        });

        const tx = await aptos.transaction.build.simple({
            sender: account.accountAddress,
            data: {
                function: `${contractAddress}::stork::update_multiple_temporal_numeric_values_evm`,
                functionArguments: [
                    updateData.ids,
                    updateData.temporal_numeric_value_timestamp_nss,
                    updateData.temporal_numeric_value_magnitudes,
                    updateData.temporal_numeric_value_negatives,
                    updateData.publisher_merkle_roots,
                    updateData.value_compute_alg_hashes,
                    updateData.rs,
                    updateData.ss,
                    updateData.vs,
                ],
            },
        });

        const senderAuthenticator = aptos.transaction.sign({
            signer: account,
            transaction: tx,
        });

        const committedTransaction = await aptos.transaction.submit.simple({
            transaction: tx,
            senderAuthenticator,
        });

        const executedTransaction = await aptos.waitForTransaction({ transactionHash: committedTransaction.hash });
        if (executedTransaction.success) {
            console.log(`Transaction succeeded: ${committedTransaction.hash}`);
        } else {
            console.error(`Transaction failed: ${committedTransaction.hash}`);
        }
    });

cliProgram
    .command("get-state-info")
    .description("Get all StorkState info")
    .action(async () => { 
        const contractAddress = DEFAULT_CONTRACT_ADDRESS;
        
        const pubKeyResult = await aptos.view({
            payload: {
                function: `${contractAddress}::state::get_stork_evm_public_key`,
            },
        });
        
        const parsedResult = JSON.parse(JSON.stringify(pubKeyResult[0]));
        const storkEvmPublicKey = parsedResult.bytes;
        console.log(`Stork EVM public key: ${storkEvmPublicKey}`);

        const updateFeeResult = await aptos.view({
            payload: {
                function: `${contractAddress}::state::get_single_update_fee_in_octas`,
            },
        });
        let updateFee = updateFeeResult[0] as number;
        console.log(`Update fee: ${updateFee}`);

        const ownerResult = await aptos.view({
            payload: {
                function: `${contractAddress}::state::get_owner`,
            },
        });
        let owner = ownerResult[0] as string;
        console.log(`Owner: ${owner}`);
    });

cliProgram
    .command("set-owner")
    .description("Set the owner")
    .argument("owner", "The new owner")
    .action(async (owner: string) => {
        const account = getAccount();
        const contractAddress = DEFAULT_CONTRACT_ADDRESS;
        const tx = await aptos.transaction.build.simple({
            sender: account.accountAddress,
            data: {
                function: `${contractAddress}::state::set_owner`,
                functionArguments: [owner],
            },
        });
        const senderAuthenticator = aptos.transaction.sign({
            signer: account,
            transaction: tx,
        });

        const committedTransaction = await aptos.transaction.submit.simple({
            transaction: tx,
            senderAuthenticator,
        });

        const executedTransaction = await aptos.waitForTransaction({ transactionHash: committedTransaction.hash });
        if (executedTransaction.success) {
            console.log(`Transaction succeeded: ${committedTransaction.hash}`);
        } else {
            console.error(`Transaction failed: ${committedTransaction.hash}`);
        }
    });

cliProgram
    .command("set-update-fee")
    .description("Set the update fee")
    .argument("fee", "The new fee")
    .action(async (fee: number) => {
        const account = getAccount();
        const contractAddress = DEFAULT_CONTRACT_ADDRESS;
        const tx = await aptos.transaction.build.simple({
            sender: account.accountAddress,
            data: {
                function: `${contractAddress}::state::set_single_update_fee_in_octas`,
                functionArguments: [fee],
            },
        });
        const senderAuthenticator = aptos.transaction.sign({
            signer: account,
            transaction: tx,
        });

        const committedTransaction = await aptos.transaction.submit.simple({
            transaction: tx,
            senderAuthenticator,
        });

        const executedTransaction = await aptos.waitForTransaction({ transactionHash: committedTransaction.hash });
        if (executedTransaction.success) {
            console.log(`Transaction succeeded: ${committedTransaction.hash}`);
        } else {
            console.error(`Transaction failed: ${committedTransaction.hash}`);
        }

    });

cliProgram
    .command("set-stork-evm-public-key")
    .description("Set the stork EVM public key")
    .argument("key", "The new key")
    .action(async (key: string) => {
        const account = getAccount();
        const contractAddress = DEFAULT_CONTRACT_ADDRESS;
        const tx = await aptos.transaction.build.simple({
            sender: account.accountAddress,
            data: {
                function: `${contractAddress}::state::set_stork_evm_public_key`,
                functionArguments: [hexStringToByteArray(key)],
            },
        });
        const senderAuthenticator = aptos.transaction.sign({
            signer: account,
            transaction: tx,
        });
        const committedTransaction = await aptos.transaction.submit.simple({
            transaction: tx,
            senderAuthenticator,
        });
        const executedTransaction = await aptos.waitForTransaction({ transactionHash: committedTransaction.hash });
        if (executedTransaction.success) {
            console.log(`Transaction succeeded: ${committedTransaction.hash}`);
        } else {
            console.error(`Transaction failed: ${committedTransaction.hash}`);
        }
    });
cliProgram.parse();


