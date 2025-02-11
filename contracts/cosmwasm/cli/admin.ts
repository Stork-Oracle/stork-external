import { Command } from "commander";
import { SigningCosmWasmClient } from "@cosmjs/cosmwasm-stargate";
import { DirectSecp256k1HdWallet } from "@cosmjs/proto-signing";
import { StorkClient } from "./client/Stork.client";
import { Coin, UpdateData, TemporalNumericValue, InstantiateMsg } from "./client/Stork.types";
import { Decimal } from "@cosmjs/math";

const DEFAULT_RPC_URL = process.env.RPC_URL;
const STORK_CONTRACT_ADDRESS = process.env.STORK_CONTRACT_ADDRESS;
const MNEMONIC = process.env.MNEMONIC;

const DEFAULT_SINGLE_UPDATE_FEE = 1;
const DEFAULT_EVM_PUBLIC_KEY = "0x0a803F9b1CCe32e2773e0d2e98b37E0775cA5d44";

const DEFAULT_DENOM = process.env.NATIVE_DENOM
const DEFAULT_GAS_PRICE = process.env.GAS_PRICE
const PREFIX = process.env.CHAIN_PREFIX

async function getSender() {
    if (!MNEMONIC) {
        throw new Error("MNEMONIC environment variable is not set");
    }
    if (!PREFIX) {
        throw new Error("CHAIN_PREFIX environment variable is not set");
    }
    const options = {
        prefix: PREFIX
    }
    const wallet = await DirectSecp256k1HdWallet.fromMnemonic(MNEMONIC, options);
    const [firstAccount] = await wallet.getAccounts();
    return firstAccount.address;
}

async function getSigningClient(): Promise<SigningCosmWasmClient> {
    if (!MNEMONIC) {
        throw new Error("MNEMONIC environment variable is not set");
    }
    if (!DEFAULT_DENOM) {
        throw new Error("NATIVE_DENOM environment variable is not set");
    }
    if (!DEFAULT_GAS_PRICE) {
        throw new Error("GAS_PRICE environment variable is not set");
    }
    if (!DEFAULT_RPC_URL) {
        throw new Error("RPC_URL environment variable is not set");
    }
    if (!PREFIX) {
        throw new Error("CHAIN_PREFIX environment variable is not set");
    }
    const wallet_options = {
        prefix: PREFIX
    }
    const wallet = await DirectSecp256k1HdWallet.fromMnemonic(MNEMONIC, wallet_options);
    const options = {
        gasPrice: {
            amount: Decimal.fromUserInput(DEFAULT_GAS_PRICE, 6),
            denom: DEFAULT_DENOM
        }
    }
    return await SigningCosmWasmClient.connectWithSigner(DEFAULT_RPC_URL, wallet, options);
}

const cliProgram = new Command();
cliProgram
    .name("admin")
    .description("CosmWasm Stork admin client")
    .version("0.1.0");

cliProgram
    .command("get-info")
    .description("Get all Stork contract information")
    .action(async () => {
        if (!STORK_CONTRACT_ADDRESS) {
            throw new Error("STORK_CONTRACT_ADDRESS environment variable is not set");
        }

        const client = await getSigningClient();
        const storkClient = new StorkClient(client, await getSender(), STORK_CONTRACT_ADDRESS);

        const [fee, publicKey, owner] = await Promise.all([
            storkClient.getSingleUpdateFee(),
            storkClient.getStorkEvmPublicKey(),
            storkClient.getOwner()
        ]);

        console.log({
            single_update_fee: fee,
            stork_evm_public_key: publicKey,
            owner: owner
        });
    });

cliProgram
    .command("update-fee")
    .description("Update the single update fee")
    .argument("<amount>", "Fee amount")
    .argument("<denom>", "Fee denomination")
    .action(async (amount: string, denom: string) => {
        if (!STORK_CONTRACT_ADDRESS) {
            throw new Error("STORK_CONTRACT_ADDRESS environment variable is not set");
        }

        const signingClient = await getSigningClient();
        const storkClient = new StorkClient(signingClient, await getSender(), STORK_CONTRACT_ADDRESS);

        const fee: Coin = {
            amount: amount,
            denom: denom
        };

        const result = await storkClient.setSingleUpdateFee({ fee });
        console.log("Fee updated:", result);
    });

cliProgram
    .command("update-evm-key")
    .description("Update the Stork EVM public key")
    .argument("<public-key>", "New EVM public key (as array of 20 numbers)")
    .action(async (publicKey: string) => {
        if (!STORK_CONTRACT_ADDRESS) {
            throw new Error("STORK_CONTRACT_ADDRESS environment variable is not set");
        }

        const signingClient = await getSigningClient();
        const storkClient = new StorkClient(signingClient, await getSender(), STORK_CONTRACT_ADDRESS);

        // Convert string input to number array
        const keyArray = JSON.parse(publicKey);
        if (!Array.isArray(keyArray) || keyArray.length !== 20) {
            throw new Error("Public key must be an array of 20 numbers");
        }

        const result = await storkClient.setStorkEvmPublicKey({ storkEvmPublicKey: keyArray });
        console.log("EVM public key updated:", result);
    });

cliProgram
    .command("update-owner")
    .description("Update the contract owner")
    .argument("<new-owner>", "New owner address")
    .action(async (newOwner: string) => {
        if (!STORK_CONTRACT_ADDRESS) {
            throw new Error("STORK_CONTRACT_ADDRESS environment variable is not set");
        }

        const signingClient = await getSigningClient();
        const storkClient = new StorkClient(signingClient, await getSender(), STORK_CONTRACT_ADDRESS);

        const result = await storkClient.setOwner({ owner: newOwner });
        console.log("Owner updated:", result);
    });

cliProgram
    .command("update-values")
    .description("Update temporal numeric values")
    .argument("<update-data>", "Update data as JSON string")
    .action(async (updateDataStr: string) => {
        if (!STORK_CONTRACT_ADDRESS) {
            throw new Error("STORK_CONTRACT_ADDRESS environment variable is not set");
        }

        const signingClient = await getSigningClient();
        const storkClient = new StorkClient(signingClient, await getSender(), STORK_CONTRACT_ADDRESS);

        const updateData = JSON.parse(updateDataStr);
        const result = await storkClient.updateTemporalNumericValuesEvm({ updateData });
        console.log("Values updated:", result);
    });

cliProgram
    .command("instantiate")
    .description("Instantiate the Stork contract")
    .argument("<code-id>", "The code ID of the uploaded contract")
    .action(async (codeId: string) => {
        if (!MNEMONIC) {
            throw new Error("MNEMONIC environment variable is not set");
        }

        const signingClient = await getSigningClient();

        // Convert EVM public key from hex to byte array
        const evmKeyArray = hexStringToByteArray(DEFAULT_EVM_PUBLIC_KEY);
        if (evmKeyArray.length !== 20) {
            throw new Error("EVM public key must be 20 bytes");
        }

        const msg: InstantiateMsg = {
            single_update_fee: {
                amount: DEFAULT_SINGLE_UPDATE_FEE.toString(),
                denom: DEFAULT_DENOM as string
            },
            stork_evm_public_key: evmKeyArray as [number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number]
        };

        const options = {
            admin: await getSender(),
        }
        const result = await signingClient.instantiate(
            await getSender(),
            parseInt(codeId),
            msg,
            "Stork",
            "auto",
            options
        );

        console.log("Contract instantiated:", {
            contractAddress: result.contractAddress,
            transactionHash: result.transactionHash
        });
    });

cliProgram
    .command("write-to-feeds")
    .description("Write to feeds")
    .argument("asset_pairs", "The asset pairs (comma separated)")
    .argument("endpoint", "The REST endpoint")
    .argument("auth_key", "The auth key")
    .option("-r, --report", "Report the results", false)
    .action(async (asset_pairs: string, endpoint: string, auth_key: string, options: { report: boolean }) => {
        if (!STORK_CONTRACT_ADDRESS) {
            throw new Error("STORK_CONTRACT_ADDRESS environment variable is not set");
        }

        console.log(`Writing to feeds: ${asset_pairs}`);

        try {
            const response = await fetch(
                `${endpoint}/v1/prices/latest?assets=${asset_pairs}`,
                {
                    headers: {
                        Authorization: `Basic ${auth_key}`,
                    },
                }
            );
            const rawJson = await response.text();
            const safeJsonText = rawJson.replace(
                /(?<!["\d])\b\d{16,}\b(?!["])/g,
                (match) => `"${match}"`
            );
            const responseData = JSON.parse(safeJsonText);
            const updateData: UpdateData[] = [];

            Object.values(responseData.data).forEach((data: any) => {
                const temporalNumericValue: TemporalNumericValue = {
                    timestamp_ns: data.stork_signed_price.timestamped_signature.timestamp.toString(),
                    quantized_value: data.stork_signed_price.price.toString()
                };

                const id = hexStringToByteArray(data.stork_signed_price.encoded_asset_id).slice(0, 32) as [number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number];
                const publisher_merkle_root = hexStringToByteArray(data.stork_signed_price.publisher_merkle_root).slice(0, 32) as [number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number];
                const value_compute_alg_hash = hexStringToByteArray(data.stork_signed_price.calculation_alg.checksum).slice(0, 32) as [number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number];
                const r = hexStringToByteArray(data.stork_signed_price.timestamped_signature.signature.r).slice(0, 32) as [number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number];
                const s = hexStringToByteArray(data.stork_signed_price.timestamped_signature.signature.s).slice(0, 32) as [number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number, number];
                const v = hexStringToByteArray(data.stork_signed_price.timestamped_signature.signature.v).slice(0, 1)[0];

                
                updateData.push({
                    id,
                    temporal_numeric_value: temporalNumericValue,
                    publisher_merkle_root,
                    value_compute_alg_hash,
                    r,
                    s,
                    v
                });
            });

            const signingClient = await getSigningClient();
            const storkClient = new StorkClient(signingClient, await getSender(), STORK_CONTRACT_ADDRESS);

            const singleUpdateFeeResponse = await storkClient.getSingleUpdateFee();
            const amount = BigInt(singleUpdateFeeResponse.fee.amount) * BigInt(updateData.length);
            const denom = singleUpdateFeeResponse.fee.denom;
            const funds: Coin[] = [{ amount: amount.toString(), denom }];
            const result = await storkClient.updateTemporalNumericValuesEvm({ updateData }, 'auto', "", funds);
            console.log("Values updated:", result.transactionHash);

            if (options.report) {
                console.log("Update data:", JSON.stringify(updateData, null, 2));
            }

        } catch (error) {
            console.error("Error:", error);
            process.exit(1);
        }
    });

    function hexStringToByteArray(hexString: string) {
        if (hexString.startsWith("0x")) {
          hexString = hexString.slice(2);
        }
        return Array.from(Buffer.from(hexString, "hex"));
    }

cliProgram.parse();