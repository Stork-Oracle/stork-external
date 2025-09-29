import { Command } from "commander";
import { createClient, Pact, createSignWithKeypair } from "@kadena/client";
import { restoreKeyPairFromSecretKey } from "@kadena/cryptography-utils";
import fs from "fs";

const DEFAULT_NETWORK_ID = process.env.NETWORK_ID || "development";
const DEFAULT_CHAIN_ID = process.env.CHAIN_ID || "1";
const DEFAULT_API_HOST = process.env.API_HOST || "http://localhost:8080";
const STORK_CONTRACT_ADDRESS = process.env.STORK_CONTRACT_ADDRESS || "stork";
const ADMIN_ACCOUNT = process.env.ADMIN_ACCOUNT;
const ADMIN_SECRET_KEY = process.env.ADMIN_SECRET_KEY;
const DEFAULT_STORK_EVM_PUBLIC_KEY = "0x0a803F9b1CCe32e2773e0d2e98b37E0775cA5d44";
const DEFAULT_UPDATE_FEE_IN_STU = 1;

type TemporalNumericValue = {
    timestampNs: number;
    quantizedValue: number;
};

type UpdateData = {
    encodedAssetId: string;
    temporalNumericValueTimestampNs: string;
    temporalNumericValueQuantizedValue: string;
    publisherMerkleRoot: string;
    valueComputeAlgHash: string;
    r: string;
    s: string;
    v: string;
};

function getClient() {
    return createClient(`${DEFAULT_API_HOST}/chainweb/0.0/${DEFAULT_NETWORK_ID}/chain/${DEFAULT_CHAIN_ID}/pact`);
}

function getAdminKeypair() {
    if (!ADMIN_SECRET_KEY) {
        throw new Error("ADMIN_SECRET_KEY environment variable is not set");
    }
    if (!ADMIN_ACCOUNT) {
        throw new Error("ADMIN_ACCOUNT environment variable is not set");
    }
    
    return restoreKeyPairFromSecretKey(ADMIN_SECRET_KEY);
}

function signTransaction(transaction: any) {
    const keypair = getAdminKeypair();
    const signWithKeypair = createSignWithKeypair(keypair);
    return signWithKeypair(transaction);
}

function hexStringToByteArray(hexString: string): number[] {
    if (hexString.startsWith("0x")) {
        hexString = hexString.slice(2);
    }
    return Array.from(Buffer.from(hexString, "hex"));
}

const cliProgram = new Command();
cliProgram
    .name("admin")
    .description("Kadena Stork admin client")
    .version("0.1.0");

cliProgram
    .command("initialize")
    .description("Initialize the Stork contract")
    .option("--stork-key <key>", "Stork EVM public key", DEFAULT_STORK_EVM_PUBLIC_KEY)
    .option("--fee <amount>", "Single update fee in STU", DEFAULT_UPDATE_FEE_IN_STU.toString())
    .action(async (options) => {
        try {
            const client = getClient();
            
            const keypair = getAdminKeypair();
            const transaction = Pact.builder
                .execution(
                    `(free.${STORK_CONTRACT_ADDRESS}.initialize "${options.storkKey}" ${parseInt(options.fee)})`
                )
                .addSigner(keypair.publicKey, (withCapability) => [
                    withCapability("coin.GAS"),
                    withCapability("stork.GOVERNANCE")
                ])
                .setMeta({
                    chainId: DEFAULT_CHAIN_ID,
                    senderAccount: ADMIN_ACCOUNT!,
                    gasLimit: 100000,
                    gasPrice: 0.00001,
                    ttl: 7200
                })
                .setNetworkId(DEFAULT_NETWORK_ID)
                .createTransaction();

            const signedTx = await signTransaction(transaction);
            const result = await client.submit(signedTx);
            
            console.log("Transaction submitted, waiting for result...");
            console.log("Request key:", result.requestKey);
            
            // Wait for transaction to be mined and check result
            const txResult = await client.listen({ 
                requestKey: result.requestKey,
                chainId: DEFAULT_CHAIN_ID,
                networkId: DEFAULT_NETWORK_ID
            });
            
            if (txResult.result.status === 'success') {
                console.log("Contract initialized successfully!");
                console.log("Result:", txResult.result.data);
            } else {
                console.error("Contract initialization failed:");
                console.error("Error:", txResult.result.error);
                process.exit(1);
            }
            
        } catch (error) {
            console.error("Error initializing contract:", error);
            process.exit(1);
        }
    });

cliProgram
    .command("get-state-info")
    .description("Get the state info of the Stork contract")
    .action(async () => {
        try {
            const client = getClient();
            
            // Get EVM public key
            const evmKeyQuery = Pact.builder
                .execution(
                    `(free.${STORK_CONTRACT_ADDRESS}.get-stork-evm-public-key)`
                )
                .setMeta({ chainId: DEFAULT_CHAIN_ID })
                .setNetworkId(DEFAULT_NETWORK_ID)
                .createTransaction();

            const evmKeyResult = await client.local(evmKeyQuery, {
                preflight: false,
                signatureVerification: false
            });

            // Get update fee
            const feeQuery = Pact.builder
                .execution(
                    `(free.${STORK_CONTRACT_ADDRESS}.get-single-update-fee-in-stu)`
                )
                .setMeta({ chainId: DEFAULT_CHAIN_ID })
                .setNetworkId(DEFAULT_NETWORK_ID)
                .createTransaction();

            const feeResult = await client.local(feeQuery, {
                preflight: false,
                signatureVerification: false
            });

            console.log("EVM Key Result:", JSON.stringify(evmKeyResult, null, 2));
            console.log("Fee Result:", JSON.stringify(feeResult, null, 2));
            
            console.log({
                stork_evm_public_key: evmKeyResult.result.status === 'success' ? evmKeyResult.result.data : evmKeyResult.result,
                single_update_fee_in_stu: feeResult.result.status === 'success' ? feeResult.result.data : feeResult.result
            });

        } catch (error) {
            console.error("Error getting state info:", error);
            process.exit(1);
        }
    });

cliProgram
    .command("update-stork-public-key")
    .description("Updates the stork public key")
    .argument("evm_public_key", "The EVM public key")
    .action(async (evmPublicKey: string) => {
        try {
            const client = getClient();
            
            if (!evmPublicKey.startsWith("0x")) {
                evmPublicKey = `0x${evmPublicKey}`;
            }

            const transaction = Pact.builder
                .execution(
                    `(free.${STORK_CONTRACT_ADDRESS}.update-stork-evm-public-key "${evmPublicKey}")`
                )
                .addSigner(getAdminKeypair().publicKey, (withCapability) => [
                    withCapability("coin.GAS"),
                    withCapability("stork.GOVERNANCE")
                ])
                .setMeta({
                    chainId: DEFAULT_CHAIN_ID,
                    senderAccount: ADMIN_ACCOUNT!,
                    gasLimit: 10000,
                    gasPrice: 0.00001,
                    ttl: 7200
                })
                .setNetworkId(DEFAULT_NETWORK_ID)
                .createTransaction();

            const signedTx = await signTransaction(transaction);
            const result = await client.submit(signedTx);
            
            console.log("EVM public key updated successfully");
            console.log("Request key:", result.requestKey);

        } catch (error) {
            console.error("Error updating EVM public key:", error);
            process.exit(1);
        }
    });

cliProgram
    .command("update-single-update-fee-in-stu")
    .description("Updates the single update fee in STU")
    .argument("single_update_fee_in_stu", "The single update fee in STU")
    .action(async (singleUpdateFeeInStu: string) => {
        try {
            const client = getClient();
            
            const transaction = Pact.builder
                .execution(
                    `(free.${STORK_CONTRACT_ADDRESS}.update-single-update-fee-in-stu ${parseInt(singleUpdateFeeInStu)})`
                )
                .addSigner(getAdminKeypair().publicKey, (withCapability) => [
                    withCapability("coin.GAS"),
                    withCapability("stork.GOVERNANCE")
                ])
                .setMeta({
                    chainId: DEFAULT_CHAIN_ID,
                    senderAccount: ADMIN_ACCOUNT!,
                    gasLimit: 10000,
                    gasPrice: 0.00001,
                    ttl: 7200
                })
                .setNetworkId(DEFAULT_NETWORK_ID)
                .createTransaction();

            const signedTx = await signTransaction(transaction);
            const result = await client.submit(signedTx);
            
            console.log("Update fee updated successfully");
            console.log("Request key:", result.requestKey);

        } catch (error) {
            console.error("Error updating fee:", error);
            process.exit(1);
        }
    });

cliProgram
    .command("write-to-feeds")
    .description("Write to feeds")
    .argument("asset_pairs", "The asset pairs (comma separated)")
    .argument("endpoint", "The REST endpoint")
    .argument("auth_key", "The auth key")
    .option("--payer <account>", "The payer account", ADMIN_ACCOUNT)
    .action(async (assetPairs: string, endpoint: string, authKey: string, options) => {
        try {
            console.log(`Writing to feeds: ${assetPairs}`);
            
            const response = await fetch(
                `${endpoint}/v1/prices/latest?assets=${assetPairs}`,
                {
                    headers: {
                        Authorization: `Basic ${authKey}`,
                    },
                }
            );

            const rawJson = await response.text();
            const safeJsonText = rawJson.replace(
                /(?<!["\d])\b\d{16,}\b(?!["])/g,
                (match) => `"${match}"`
            );
            
            const responseData = JSON.parse(safeJsonText);
            const client = getClient();

            for (const data of Object.values(responseData.data) as any[]) {
                const updateData: UpdateData = {
                    encodedAssetId: data.stork_signed_price.encoded_asset_id,
                    temporalNumericValueTimestampNs: data.stork_signed_price.timestamped_signature.timestamp,
                    temporalNumericValueQuantizedValue: data.stork_signed_price.price,
                    publisherMerkleRoot: data.stork_signed_price.publisher_merkle_root,
                    valueComputeAlgHash: `0x${data.stork_signed_price.calculation_alg.checksum}`,
                    r: data.stork_signed_price.timestamped_signature.signature.r,
                    s: data.stork_signed_price.timestamped_signature.signature.s,
                    v: data.stork_signed_price.timestamped_signature.signature.v,
                };

                const transaction = Pact.builder
                    .execution(
                        `(free.${STORK_CONTRACT_ADDRESS}.update-temporal-numeric-value-evm "${options.payer}" "${updateData.encodedAssetId}" ${parseInt(updateData.temporalNumericValueTimestampNs)} ${parseInt(updateData.temporalNumericValueQuantizedValue)} "${updateData.publisherMerkleRoot}" "${updateData.valueComputeAlgHash}" "${updateData.r}" "${updateData.s}" "${updateData.v}")`
                    )
                    .addSigner(getAdminKeypair().publicKey, (withCapability) => [
                        withCapability("coin.GAS"),
                        withCapability("coin.TRANSFER", options.payer, "stork-treasury", { decimal: "1.0" })
                    ])
                    .setMeta({
                        chainId: DEFAULT_CHAIN_ID,
                        senderAccount: options.payer,
                        gasLimit: 15000,
                        gasPrice: 0.00001,
                        ttl: 7200
                    })
                    .setNetworkId(DEFAULT_NETWORK_ID)
                    .createTransaction();

                const signedTx = await signTransaction(transaction);
                const result = await client.submit(signedTx);
                
                console.log(`Updated ${updateData.encodedAssetId}`);
                console.log("Request key:", result.requestKey);
            }

        } catch (error) {
            console.error("Error writing to feeds:", error);
            process.exit(1);
        }
    });

cliProgram
    .command("read-from-feed")
    .description("Read from feed")
    .argument("asset_id", "The plaintext asset id")
    .action(async (assetId: string) => {
        try {
            // For Kadena, we would need to encode the asset ID similar to other chains
            // This is a simplified version - you may need to implement the same hashing logic
            const encodedAssetId = `0x${Buffer.from(assetId).toString('hex')}`;
            
            console.log("Encoded asset id:", encodedAssetId);
            
            const client = getClient();
            
            const valueQuery = Pact.builder
                .execution(
                    `(free.${STORK_CONTRACT_ADDRESS}.get-latest-temporal-numeric-value-unchecked "${encodedAssetId}")`
                )
                .setMeta({ chainId: DEFAULT_CHAIN_ID })
                .setNetworkId(DEFAULT_NETWORK_ID)
                .createTransaction();

            const valueResult = await client.local(valueQuery, {
                preflight: false,
                signatureVerification: false
            });

            const value = valueResult.result.status === 'success' ? valueResult.result.data as TemporalNumericValue : null;
            
            if (value) {
                console.log("Temporal numeric value:");
                console.log("  Timestamp (ns):", value.timestampNs);
                console.log("  Quantized Value:", value.quantizedValue);
            } else {
                console.log("Error: Could not retrieve temporal numeric value");
            }

        } catch (error) {
            console.error("Error reading from feed:", error);
            process.exit(1);
        }
    });

cliProgram
    .command("register-namespace")
    .description("Register the stork namespace")
    .action(async () => {
        try {
            const client = getClient();
            const keypair = getAdminKeypair();
            
            const namespaceCode = `(define-namespace 'stork (read-keyset 'stork) (read-keyset 'stork))`;
            
            const transaction = Pact.builder
                .execution(namespaceCode)
                .addData("stork", {
                    keys: [keypair.publicKey],
                    pred: "keys-all"
                })
                .addSigner(keypair.publicKey, (withCapability) => [
                    withCapability("coin.GAS")
                ])
                .setMeta({
                    chainId: DEFAULT_CHAIN_ID,
                    senderAccount: ADMIN_ACCOUNT!,
                    gasLimit: 100000,
                    gasPrice: 0.00001,
                    ttl: 7200
                })
                .setNetworkId(DEFAULT_NETWORK_ID)
                .createTransaction();

            const signedTx = await signTransaction(transaction);
            const result = await client.submit(signedTx);
            
            console.log("Namespace registration submitted, waiting for result...");
            console.log("Request key:", result.requestKey);
            
            // Wait for transaction to be mined and check result
            const txResult = await client.listen({ 
                requestKey: result.requestKey,
                chainId: DEFAULT_CHAIN_ID,
                networkId: DEFAULT_NETWORK_ID
            });
            
            if (txResult.result.status === 'success') {
                console.log("Namespace registered successfully!");
                console.log("Result:", txResult.result.data);
            } else {
                console.error("Namespace registration failed:");
                console.error("Error:", txResult.result.error);
                process.exit(1);
            }

        } catch (error) {
            console.error("Error registering namespace:", error);
            process.exit(1);
        }
    });

cliProgram
    .command("deploy-minimal")
    .description("Deploy the minimal Stork contract for testing")
    .action(async () => {
        try {
            const client = getClient();
            const contractCode = fs.readFileSync("chains/kadena/contracts/src/stork-minimal.pact", "utf8");
            
            const keypair = getAdminKeypair();
            const transaction = Pact.builder
                .execution(contractCode)
                .addData("stork", {
                    keys: [keypair.publicKey],
                    pred: "keys-all"
                })
                .addSigner(keypair.publicKey, (withCapability) => [
                    withCapability("coin.GAS")
                ])
                .setMeta({
                    chainId: DEFAULT_CHAIN_ID,
                    senderAccount: ADMIN_ACCOUNT!,
                    gasLimit: 100000,
                    gasPrice: 0.00001,
                    ttl: 7200
                })
                .setNetworkId(DEFAULT_NETWORK_ID)
                .createTransaction();

            const signedTx = await signTransaction(transaction);
            const result = await client.submit(signedTx);
            
            console.log("Transaction submitted, waiting for result...");
            console.log("Request key:", result.requestKey);
            
            // Wait for transaction to be mined and check result
            const txResult = await client.listen({ 
                requestKey: result.requestKey,
                chainId: DEFAULT_CHAIN_ID,
                networkId: DEFAULT_NETWORK_ID
            });
            
            if (txResult.result.status === 'success') {
                console.log("Contract deployed successfully!");
                console.log("Result:", txResult.result.data);
            } else {
                console.error("Contract deployment failed:");
                console.error("Error:", txResult.result.error);
                process.exit(1);
            }

        } catch (error) {
            console.error("Error deploying contract:", error);
            process.exit(1);
        }
    });

cliProgram
    .command("deploy-simple")
    .description("Deploy the simplified Stork contract (without namespace)")
    .action(async () => {
        try {
            const client = getClient();
            const contractCode = fs.readFileSync("chains/kadena/contracts/src/stork-simple.pact", "utf8");
            
            const keypair = getAdminKeypair();
            const transaction = Pact.builder
                .execution(contractCode)
                .addData("stork", {
                    keys: [keypair.publicKey],
                    pred: "keys-all"
                })
                .addSigner(keypair.publicKey, (withCapability) => [
                    withCapability("coin.GAS")
                ])
                .setMeta({
                    chainId: DEFAULT_CHAIN_ID,
                    senderAccount: ADMIN_ACCOUNT!,
                    gasLimit: 100000,
                    gasPrice: 0.00001,
                    ttl: 7200
                })
                .setNetworkId(DEFAULT_NETWORK_ID)
                .createTransaction();

            const signedTx = await signTransaction(transaction);
            const result = await client.submit(signedTx);
            
            console.log("Transaction submitted, waiting for result...");
            console.log("Request key:", result.requestKey);
            
            // Wait for transaction to be mined and check result
            const txResult = await client.listen({ 
                requestKey: result.requestKey,
                chainId: DEFAULT_CHAIN_ID,
                networkId: DEFAULT_NETWORK_ID
            });
            
            if (txResult.result.status === 'success') {
                console.log("Contract deployed successfully!");
                console.log("Result:", txResult.result.data);
            } else {
                console.error("Contract deployment failed:");
                console.error("Error:", txResult.result.error);
                process.exit(1);
            }

        } catch (error) {
            console.error("Error deploying contract:", error);
            process.exit(1);
        }
    });

cliProgram
    .command("deploy")
    .description("Deploy the Stork contract")
    .action(async () => {
        try {
            const client = getClient();
            const contractCode = fs.readFileSync("chains/kadena/contracts/src/stork.pact", "utf8");
            
            const keypair = getAdminKeypair();
            const transaction = Pact.builder
                .execution(contractCode)
                .addData("stork", {
                    keys: [keypair.publicKey],
                    pred: "keys-all"
                })
                .addSigner(keypair.publicKey, (withCapability) => [
                    withCapability("coin.GAS")
                ])
                .setMeta({
                    chainId: DEFAULT_CHAIN_ID,
                    senderAccount: ADMIN_ACCOUNT!,
                    gasLimit: 100000,
                    gasPrice: 0.00001,
                    ttl: 7200
                })
                .setNetworkId(DEFAULT_NETWORK_ID)
                .createTransaction();

            const signedTx = await signTransaction(transaction);
            const result = await client.submit(signedTx);
            
            console.log("Transaction submitted, waiting for result...");
            console.log("Request key:", result.requestKey);
            
            // Wait for transaction to be mined and check result
            const txResult = await client.listen({ 
                requestKey: result.requestKey,
                chainId: DEFAULT_CHAIN_ID,
                networkId: DEFAULT_NETWORK_ID
            });
            
            if (txResult.result.status === 'success') {
                console.log("Contract deployed successfully!");
                console.log("Result:", txResult.result.data);
            } else {
                console.error("Contract deployment failed:");
                console.error("Error:", txResult.result.error);
                process.exit(1);
            }

        } catch (error) {
            console.error("Error deploying contract:", error);
            process.exit(1);
        }
    });

cliProgram.parse(process.argv);