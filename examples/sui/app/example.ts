import { Command } from "commander";
import { SuiClient, getFullnodeUrl, SuiMoveNormalizedModules, SuiMoveNormalizedModule, SuiEventFilter } from "@mysten/sui/client";
import { Transaction } from '@mysten/sui/transactions';
import { Ed25519Keypair } from '@mysten/sui/keypairs/ed25519';
import { fromBase64 } from "@mysten/sui/utils";
import keccak256 from 'keccak256';
import * as fs from 'fs';

const DEFAULT_RPC_URL = getFullnodeUrl(process.env.RPC_ALIAS as 'mainnet' | 'testnet' | 'devnet' | 'localnet');
const client = new SuiClient({ url: DEFAULT_RPC_URL });
const DEFAULT_ALIASES_PATH = `${process.env.HOME}/.sui/sui_config/sui.aliases`;
const DEFAULT_KEYSTORE_PATH = `${process.env.HOME}/.sui/sui_config/sui.keystore`;
const KEY_ALIAS=process.env.SUI_KEY_ALIAS || 'main';
const EXAMPLE_PACKAGE_ADDRESS = process.env.EXAMPLE_PACKAGE_ADDRESS;

function loadKeypairFromKeystore(): Ed25519Keypair {
    const aliasesContent = fs.readFileSync(DEFAULT_ALIASES_PATH, 'utf-8');
    const aliases = JSON.parse(aliasesContent);
    
    const aliasIndex = aliases.findIndex((entry: any) => entry.alias === KEY_ALIAS);
    if (aliasIndex === -1) {
        throw new Error(`Alias "${KEY_ALIAS}" not found in aliases file`);
    }

    const keystoreContent = fs.readFileSync(DEFAULT_KEYSTORE_PATH, 'utf-8');
    const keystore = JSON.parse(keystoreContent);
    
    const privateKeyBase64 = keystore[aliasIndex];
    if (!privateKeyBase64) {
        throw new Error(`No private key found for alias "${KEY_ALIAS}" at index ${aliasIndex}`);
    }

    const privateKeyBytes = fromBase64(privateKeyBase64);
    const actualPrivateKey = privateKeyBytes.slice(1);
    
    return Ed25519Keypair.fromSecretKey(actualPrivateKey);
}

// This is necessary because contract addresses change on upgrade, while their types retain the package ID of the original deployment.
async function getOrigionalContractId(storkContractAddress: string): Promise<string> {
    const modules: SuiMoveNormalizedModules = await client.getNormalizedMoveModulesByPackage({
        package: storkContractAddress
    });

    const adminModule: SuiMoveNormalizedModule = modules[`admin`];
    return adminModule.address;
}

async function getStorkStateId(storkContractAddress: string): Promise<string> {
    const originalContractId = await getOrigionalContractId(storkContractAddress);
    const eventFilter: SuiEventFilter = {
        MoveModule: {
            package: originalContractId,
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

const program = new Command();
program
    .name("stork-example")
    .description("Sui Stork example client")
    .version("0.1.0");

program
    .command("read-price")
    .description("Read price from Stork feed")
    .argument("<asset>", "Asset identifier (will be hashed)")
    .argument("<stork_contract_address>", "Stork contract address")
    .action(async (asset: string, storkContractAddress: string) => {
        const storkStateId = await getStorkStateId(storkContractAddress);
        try {
            const keypair = loadKeypairFromKeystore();
            const tx = new Transaction();
            
            // Hash the asset string to get feed_id
            const feedId = keccak256(asset);
            
            tx.moveCall({
                target: `${EXAMPLE_PACKAGE_ADDRESS}::example::use_stork_price`,
                arguments: [
                    tx.pure.vector('u8', feedId),
                    tx.object(storkStateId),
                ]
            });

            const result = await client.signAndExecuteTransaction({
                signer: keypair,
                transaction: tx,
                options: {
                    showEvents: true
                }
            });

            // Display the event data
            const events = result.events;
            if (events && events.length > 0) {
                const priceEvent = events.find(e => 
                    e.type.includes('ExampleStorkPriceEvent')
                );
                if (priceEvent) {
                    console.log("Price Event:", priceEvent.parsedJson);
                }
            }
            
            console.log("Transaction digest:", result.digest);
        } catch (error) {
            console.error("Error:", error);
        }
    });

program.parse();
