import { Command } from "commander";
import { SuiClient, getFullnodeUrl, PaginatedObjectsResponse, SuiObjectData, SuiEventFilter } from "@mysten/sui/client";
import { Transaction, TransactionResult } from '@mysten/sui/transactions';
import { Ed25519Keypair } from '@mysten/sui/keypairs/ed25519';
import * as fs from 'fs';
import { fromBase64 } from "@mysten/sui/utils";

const DEFAULT_RPC_URL = getFullnodeUrl(process.env.RPC_ALIAS as 'mainnet' | 'testnet' | 'devnet' | 'localnet');

const client = new SuiClient({ url: DEFAULT_RPC_URL });

const STORK_CONTRACT_ADDRESS = process.env.STORK_CONTRACT_ADDRESS;
const DEFAULT_KEYSTORE_PATH = `${process.env.HOME}/.sui/sui_config/sui.keystore`;
const DEFAULT_ALIASES_PATH = `${process.env.HOME}/.sui/sui_config/sui.aliases`;
const SUI_KEY_ALIAS = process.env.SUI_KEY_ALIAS || 'main';
const DEFAULT_STORK_EVM_PUBLIC_KEY = "0x0a803F9b1CCe32e2773e0d2e98b37E0775cA5d44";
const DEFAULT_UPDATE_FEE_IN_MIST = 1;

type UpdateTemporalNumericValueEvmInputRaw = {
    ids: number[][],
    temporal_numeric_value_timestamp_nss: bigint[],
    temporal_numeric_value_magnitudes: bigint[],
    temporal_numeric_value_negatives: boolean[],
    publisher_merkle_roots: number[][],
    value_compute_alg_hashes: number[][],
    rs: number[][],
    ss: number[][],
    vs: number[],
}

function loadKeypairFromKeystore(key_alias: string = SUI_KEY_ALIAS): Ed25519Keypair {
    const aliasesContent = fs.readFileSync(DEFAULT_ALIASES_PATH, 'utf-8');
    const aliases = JSON.parse(aliasesContent);
    
    const aliasIndex = aliases.findIndex((entry: any) => entry.alias === key_alias);
    if (aliasIndex === -1) {
        throw new Error(`Alias "${key_alias}" not found in aliases file`);
    }

    const keystoreContent = fs.readFileSync(DEFAULT_KEYSTORE_PATH, 'utf-8');
    const keystore = JSON.parse(keystoreContent);
    
    const privateKeyBase64 = keystore[aliasIndex];
    if (!privateKeyBase64) {
        throw new Error(`No private key found for alias "${key_alias}" at index ${aliasIndex}`);
    }

    const privateKeyBytes = fromBase64(privateKeyBase64);
    const actualPrivateKey = privateKeyBytes.slice(1);
    
    return Ed25519Keypair.fromSecretKey(actualPrivateKey);
}

async function getAdminCap(keypair: Ed25519Keypair): Promise<SuiObjectData> {
    const accountAddress = keypair.getPublicKey().toSuiAddress();
    const objects: PaginatedObjectsResponse = await client.getOwnedObjects({
        owner: accountAddress,
        options: {
            showContent: true,
        }
    });

    const adminCap = objects.data.find((obj: any) => {
        return obj.data?.content?.type === `${STORK_CONTRACT_ADDRESS}::admin::AdminCap`;
    });

    if (!adminCap) {
        throw new Error("Admin cap not found");
    }

    return adminCap.data;
}

async function getStorkStateId(): Promise<string> {
    const eventFilter: SuiEventFilter = {
        MoveModule: {
            package: STORK_CONTRACT_ADDRESS,
            module: "stork",
        }
    };

    let events: any = await client.queryEvents({
        query: eventFilter,
        limit: 1,
        order: 'ascending',
    });
    
    return events.data[0].parsedJson.stork_state_id;
}

function byteArrayToHexString(byteArray: number[]) {
    return "0x" + Buffer.from(byteArray).toString('hex');
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
  .description("Sui Stork admin client")
  .version("0.1.0");

const DEFAULT_SINGLE_UPDATE_FEE_IN_MIST = 1;

cliProgram
    .command("initialize")
    .description("Initialize the Stork program")
    .argument("[stork_contract_address]", "The Stork contract address", (value) => value, STORK_CONTRACT_ADDRESS)
    .action(async (stork_contract_address: string) => {
        const keypair = loadKeypairFromKeystore();
        const tx = new Transaction();
        const stork_initialize_function = `${stork_contract_address}::stork::init_stork`;

        const adminCap = await getAdminCap(keypair);

        console.log("Initializing Stork...");
        const txResult: TransactionResult = await tx.moveCall({
            target: stork_initialize_function,
            arguments: [
                tx.object(adminCap.objectId),
                tx.pure.address(STORK_CONTRACT_ADDRESS),
                tx.pure.vector('u8', hexStringToByteArray(DEFAULT_STORK_EVM_PUBLIC_KEY)),
                tx.pure.u64(DEFAULT_UPDATE_FEE_IN_MIST),
                tx.pure.u64(1),
            ]
        });

        let result = await client.signAndExecuteTransaction({
            signer: keypair,
            transaction: tx,
            options: {
                showInput: true,
                showEvents: true,
                showEffects: true,
                showObjectChanges: true,
                showBalanceChanges: true,
            }
        });
        console.log(`Stork initialized: ${JSON.stringify(result)}`);
    });

cliProgram
    .command("write-to-feeds")
    .description("Write to feeds")
    .argument("asset_pairs", "The asset pairs (comma separated)")
    .argument("endpoint", "The REST endpoint")
    .argument("auth_key", "The auth key")
    .option("-r, --report", "Report the results", false)
    .action(async (asset_pairs: string, endpoint: string, auth_key: string, options: { report: boolean }) => {
        console.log(`Writing to feeds: ${asset_pairs}`);
        const keypair = loadKeypairFromKeystore();
        // try {
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
            
            // console.log(`Response: ${safeJsonText}`);
            const responseData = JSON.parse(safeJsonText);
            // console.log(`Response data: ${JSON.stringify(responseData)}`);
            const updateData: UpdateTemporalNumericValueEvmInputRaw = {
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

            Object.values(responseData.data).forEach((data: any) => {
                updateData.ids.push(hexStringToByteArray(data.stork_signed_price.encoded_asset_id));
                updateData.temporal_numeric_value_timestamp_nss.push(BigInt(data.stork_signed_price.timestamped_signature.timestamp));
                updateData.temporal_numeric_value_magnitudes.push(BigInt(data.stork_signed_price.price));
                updateData.temporal_numeric_value_negatives.push(data.stork_signed_price.price < 0);
                updateData.publisher_merkle_roots.push(hexStringToByteArray(data.stork_signed_price.publisher_merkle_root));
                updateData.value_compute_alg_hashes.push(hexStringToByteArray(data.stork_signed_price.calculation_alg.checksum));
                updateData.rs.push(hexStringToByteArray(data.stork_signed_price.timestamped_signature.signature.r));
                updateData.ss.push(hexStringToByteArray(data.stork_signed_price.timestamped_signature.signature.s));
                updateData.vs.push(data.stork_signed_price.timestamped_signature.signature.v);
            });

            const tx = new Transaction();
            const numUpdates = updateData.ids.length;
            const fee = DEFAULT_UPDATE_FEE_IN_MIST*numUpdates;
            console.log(`Fee: ${fee}`);
            //get update_multiple_temporal_numeric_values_evm inputs
            const [coin] = tx.splitCoins(tx.gas, [fee]);
            const storkState = await getStorkStateId();
            console.log(`Stork state: ${JSON.stringify(storkState)}`);
            const [updateTemporalNumericValueEvmInputVec] = tx.moveCall({
                target: `${STORK_CONTRACT_ADDRESS}::update_temporal_numeric_value_evm_input_vec::new`,
                arguments: [
                    tx.pure.vector('vector<u8>', updateData.ids),
                    tx.pure.vector('u64', updateData.temporal_numeric_value_timestamp_nss),
                    tx.pure.vector('u128', updateData.temporal_numeric_value_magnitudes),
                    tx.pure.vector('bool', updateData.temporal_numeric_value_negatives),
                    tx.pure.vector('vector<u8>', updateData.publisher_merkle_roots),
                    tx.pure.vector('vector<u8>', updateData.value_compute_alg_hashes),
                    tx.pure.vector('vector<u8>', updateData.rs),
                    tx.pure.vector('vector<u8>', updateData.ss),
                    tx.pure.vector('u8', updateData.vs),
                ]
            });

            

            tx.moveCall({
                target: `${STORK_CONTRACT_ADDRESS}::stork::update_multiple_temporal_numeric_values_evm`,
                arguments: [
                    tx.object(storkState),
                    updateTemporalNumericValueEvmInputVec,
                    coin,
                ]
            });
            const txResult = await client.signAndExecuteTransaction({
                signer: keypair,
                transaction: tx,
            });
            console.log(`Transaction result: ${txResult.digest}`);

        // } catch (error) {
        //     console.error(`Error: ${error}`);
        // }
    });

cliProgram
    .command("get-state-info")
    .description("Get all StorkState information")
    .action(async () => {
        const storkState = await getStorkStateId();
        const result = await client.getObject({
            id: storkState,
            options: { showContent: true }
        });
        
        if (result.data?.content?.dataType !== 'moveObject') {
            throw new Error("Could not fetch StorkState data");
        }
        
        const fields = result.data.content.fields;
        const evmPubkey = byteArrayToHexString(fields['stork_evm_public_key'].fields.bytes);
        console.log({
            stork_sui_public_key: fields['stork_sui_public_key'],
            stork_evm_public_key: evmPubkey,
            single_update_fee_in_mist: fields['single_update_fee_in_mist'],
            version: fields['version']
        });
    });

cliProgram
    .command("update-fee")
    .description("Update the single update fee")
    .argument("<new_fee>", "New fee in MIST")
    .action(async (new_fee: string) => {
        const keypair = loadKeypairFromKeystore();
        const adminCap = await getAdminCap(keypair);
        const storkState = await getStorkStateId();
        
        const tx = new Transaction();
        tx.moveCall({
            target: `${STORK_CONTRACT_ADDRESS}::state::update_single_update_fee_in_mist`,
            arguments: [
                tx.object(adminCap.objectId),
                tx.object(storkState),
                tx.pure.u64(BigInt(new_fee)),
            ]
        });
        
        const result = await client.signAndExecuteTransaction({
            signer: keypair,
            transaction: tx,
        });
        console.log("Fee updated:", result);
    });

cliProgram
    .command("update-sui-key")
    .description("Update the Stork SUI public key")
    .argument("<new_key>", "New SUI public key address")
    .action(async (new_key: string) => {
        const keypair = loadKeypairFromKeystore();
        const adminCap = await getAdminCap(keypair);
        const storkState = await getStorkStateId();
        
        const tx = new Transaction();
        tx.moveCall({
            target: `${STORK_CONTRACT_ADDRESS}::state::update_stork_sui_public_key`,
            arguments: [
                tx.object(adminCap.objectId),
                tx.object(storkState),
                tx.pure.address(new_key),
            ]
        });
        
        const result = await client.signAndExecuteTransaction({
            signer: keypair,
            transaction: tx,
        });
        console.log("SUI public key updated:", result);
    });

cliProgram
    .command("update-evm-key")
    .description("Update the Stork EVM public key")
    .argument("<new_key>", "New EVM public key (hex format)")
    .action(async (new_key: string) => {
        const keypair = loadKeypairFromKeystore();
        const adminCap = await getAdminCap(keypair);
        const storkState = await getStorkStateId();
        
        const tx = new Transaction();
        tx.moveCall({
            target: `${STORK_CONTRACT_ADDRESS}::state::update_stork_evm_public_key`,
            arguments: [
                tx.object(adminCap.objectId),
                tx.object(storkState),
                tx.pure.vector('u8', hexStringToByteArray(new_key)),
            ]
        });
        
        const result = await client.signAndExecuteTransaction({
            signer: keypair,
            transaction: tx,
        });
        console.log("EVM public key updated:", result);
    });

cliProgram
    .command("withdraw-fees")
    .description("Withdraw accumulated fees")
    .action(async () => {
        const keypair = loadKeypairFromKeystore();
        const adminCap = await getAdminCap(keypair);
        const storkState = await getStorkStateId();
        
        const tx = new Transaction();
        const [coin] = tx.moveCall({
            target: `${STORK_CONTRACT_ADDRESS}::state::withdraw_fees`,
            arguments: [
                tx.object(adminCap.objectId),
                tx.object(storkState),
            ]
        });

        tx.transferObjects([coin], keypair.getPublicKey().toSuiAddress());
        
        const result = await client.signAndExecuteTransaction({
            signer: keypair,
            transaction: tx,
        });
        console.log("Fees withdrawn:", result);
    });

cliProgram.parse();













