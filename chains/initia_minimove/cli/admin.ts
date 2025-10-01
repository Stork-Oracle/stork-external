import { Command } from "commander";
import {
    RESTClient,
    MnemonicKey,
    Wallet,
    MsgExecute,
    bcs
} from "@initia/initia.js";

// Environment variables
const RPC_URL = process.env.RPC_URL || "https://rest.testnet.initia.xyz";
const MNEMONIC = process.env.MNEMONIC;
const CONTRACT_ADDRESS = process.env.STORK_CONTRACT_ADDRESS;
const DEFAULT_STORK_EVM_PUBLIC_KEY = "0x0a803F9b1CCe32e2773e0d2e98b37E0775cA5d44";
const DEFAULT_UPDATE_FEE_AMOUNT = "1";
const DEFAULT_UPDATE_FEE_DENOM = process.env.NATIVE_DENOM || "uinit";
const CHAIN_ID = process.env.CHAIN_ID || "initiation-2";

// Helper functions
function hexStringToByteArray(hexString: string): number[] {
    if (hexString.startsWith("0x")) {
        hexString = hexString.slice(2);
    }
    return Array.from(Buffer.from(hexString, "hex"));
}

function getWallet(): Wallet {
    if (!MNEMONIC) {
        throw new Error("MNEMONIC environment variable is not set");
    }

    const rest = new RESTClient(RPC_URL, {
        chainId: CHAIN_ID,
        gasPrices: "0.15uinit",
        gasAdjustment: "1.5",
    });

    const mk = new MnemonicKey({ mnemonic: MNEMONIC, coinType: 118 });
    console.log("Mnemonic:", MNEMONIC);
    console.log("Account Address:", mk.accAddress);
    return new Wallet(rest, mk);
}

async function executeTx(msg: MsgExecute) {
    const wallet = getWallet();

    const tx = await wallet.createAndSignTx({
        msgs: [msg],
    });

    const result = await wallet.rest.tx.broadcastSync(tx);
    console.log(`Transaction hash: ${result.txhash}`);

    // Wait for transaction to be included in a block
    await new Promise(resolve => setTimeout(resolve, 2000));

    const txInfo = await wallet.rest.tx.txInfo(result.txhash);
    if (txInfo.code !== 0) {
        console.error(`Transaction failed: ${txInfo.raw_log}`);
    } else {
        console.log(`Transaction succeeded`);
    }
}

// CLI setup
const cliProgram = new Command();
cliProgram
    .name("admin")
    .description("Initia MiniMove Stork admin client")
    .version("0.1.0");

cliProgram
    .command("initialize")
    .description("Initialize the Stork contract")
    .action(async () => {
        if (!CONTRACT_ADDRESS) {
            throw new Error("STORK_CONTRACT_ADDRESS is not set");
        }

        const wallet = getWallet();
        const evmKeyBytes = hexStringToByteArray(DEFAULT_STORK_EVM_PUBLIC_KEY);

        const msg = new MsgExecute(
            wallet.key.accAddress,
            CONTRACT_ADDRESS,
            "stork",
            "init_stork",
            [],
            [
                bcs.vector(bcs.u8()).serialize(evmKeyBytes).toBase64(),
                bcs.u64().serialize(parseInt(DEFAULT_UPDATE_FEE_AMOUNT)).toBase64(),
                bcs.string().serialize(DEFAULT_UPDATE_FEE_DENOM).toBase64(),
            ]
        );

        await executeTx(msg);
    });

cliProgram
    .command("get-state-info")
    .description("Get all StorkState info")
    .action(async () => {
        if (!CONTRACT_ADDRESS) {
            throw new Error("STORK_CONTRACT_ADDRESS is not set");
        }

        const wallet = getWallet();

        try {
            // Query owner
            const ownerResult = await wallet.rest.move.viewFunction(
                CONTRACT_ADDRESS,
                "state",
                "get_owner",
                [],
                []
            );
            console.log(`Owner: ${JSON.stringify(ownerResult)}`);

            // Query stork EVM public key
            const pubKeyResult = await wallet.rest.move.viewFunction(
                CONTRACT_ADDRESS,
                "state",
                "get_stork_evm_public_key",
                [],
                []
            );
            console.log(`Stork EVM public key: ${JSON.stringify(pubKeyResult)}`);

            // Query single update fee
            const feeResult = await wallet.rest.move.viewFunction(
                CONTRACT_ADDRESS,
                "state",
                "get_single_update_fee",
                [],
                []
            );
            console.log(`Single update fee: ${JSON.stringify(feeResult)}`);
        } catch (error) {
            console.error("Error querying state:", error);
        }
    });

cliProgram
    .command("set-owner")
    .description("Set the owner")
    .argument("<owner>", "The new owner address")
    .action(async (owner: string) => {
        if (!CONTRACT_ADDRESS) {
            throw new Error("STORK_CONTRACT_ADDRESS is not set");
        }

        const wallet = getWallet();

        const msg = new MsgExecute(
            wallet.key.accAddress,
            CONTRACT_ADDRESS,
            "state",
            "set_owner",
            [],
            [
                bcs.address().serialize(owner).toBase64(),
            ]
        );

        await executeTx(msg);
    });

cliProgram
    .command("set-update-fee")
    .description("Set the update fee")
    .argument("<amount>", "The new fee amount")
    .argument("<denom>", "The fee denomination")
    .action(async (amount: string, denom: string) => {
        if (!CONTRACT_ADDRESS) {
            throw new Error("STORK_CONTRACT_ADDRESS is not set");
        }

        const wallet = getWallet();

        const msg = new MsgExecute(
            wallet.key.accAddress,
            CONTRACT_ADDRESS,
            "state",
            "set_single_update_fee",
            [],
            [
                bcs.u64().serialize(parseInt(amount)).toBase64(),
                bcs.string().serialize(denom).toBase64(),
            ]
        );

        await executeTx(msg);
    });

cliProgram
    .command("set-stork-evm-public-key")
    .description("Set the stork EVM public key")
    .argument("<key>", "The new key (hex string)")
    .action(async (key: string) => {
        if (!CONTRACT_ADDRESS) {
            throw new Error("STORK_CONTRACT_ADDRESS is not set");
        }

        const wallet = getWallet();
        const keyBytes = hexStringToByteArray(key);

        const msg = new MsgExecute(
            wallet.key.accAddress,
            CONTRACT_ADDRESS,
            "state",
            "set_stork_evm_public_key",
            [],
            [
                bcs.vector(bcs.u8()).serialize(keyBytes).toBase64(),
            ]
        );

        await executeTx(msg);
    });

cliProgram
    .command("write-to-feeds")
    .description("Write to feeds")
    .argument("<asset_pairs>", "Comma separated list of asset pairs to write to")
    .argument("<endpoint>", "The stork REST endpoint")
    .argument("<auth_key>", "The stork auth key")
    .action(async (assetPairs: string, endpoint: string, authKey: string) => {
        if (!CONTRACT_ADDRESS) {
            throw new Error("STORK_CONTRACT_ADDRESS is not set");
        }

        const wallet = getWallet();

        const result = await fetch(`${endpoint}/v1/prices/latest?assets=${assetPairs}`, {
            headers: {
                "Authorization": `Basic ${authKey}`,
            },
        });

        const rawJson = await result.text();
        const safeJsonText = rawJson.replace(
            /(?<!["\d])\b\d{16,}\b(?!["])/g,
            (match) => `"${match}"`
        );

        const response = JSON.parse(safeJsonText);

        const ids: number[][] = [];
        const timestamps: number[] = [];
        const magnitudes: string[] = [];
        const negatives: boolean[] = [];
        const merkleRoots: number[][] = [];
        const algHashes: number[][] = [];
        const rs: number[][] = [];
        const ss: number[][] = [];
        const vs: number[] = [];

        Object.values(response.data).forEach((data: any) => {
            ids.push(hexStringToByteArray(data.stork_signed_price.encoded_asset_id));
            timestamps.push(parseInt(data.stork_signed_price.timestamped_signature.timestamp));
            magnitudes.push(data.stork_signed_price.price.toString());
            negatives.push(data.stork_signed_price.price < 0);
            merkleRoots.push(hexStringToByteArray(data.stork_signed_price.publisher_merkle_root));
            algHashes.push(hexStringToByteArray(data.stork_signed_price.calculation_alg.checksum));
            rs.push(hexStringToByteArray(data.stork_signed_price.timestamped_signature.signature.r));
            ss.push(hexStringToByteArray(data.stork_signed_price.timestamped_signature.signature.s));
            vs.push(hexStringToByteArray(data.stork_signed_price.timestamped_signature.signature.v)[0]);
        });

        const msg = new MsgExecute(
            wallet.key.accAddress,
            CONTRACT_ADDRESS,
            "stork",
            "update_multiple_temporal_numeric_values_evm",
            [],
            [
                bcs.vector(bcs.vector(bcs.u8())).serialize(ids).toBase64(),
                bcs.vector(bcs.u64()).serialize(timestamps).toBase64(),
                bcs.vector(bcs.u128()).serialize(magnitudes.map(m => BigInt(m))).toBase64(),
                bcs.vector(bcs.bool()).serialize(negatives).toBase64(),
                bcs.vector(bcs.vector(bcs.u8())).serialize(merkleRoots).toBase64(),
                bcs.vector(bcs.vector(bcs.u8())).serialize(algHashes).toBase64(),
                bcs.vector(bcs.vector(bcs.u8())).serialize(rs).toBase64(),
                bcs.vector(bcs.vector(bcs.u8())).serialize(ss).toBase64(),
                bcs.vector(bcs.u8()).serialize(vs).toBase64(),
            ]
        );

        await executeTx(msg);
    });

cliProgram
    .command("get-temporal-numeric-value")
    .description("Get temporal numeric value for an asset")
    .argument("<asset_id>", "The asset ID (hex string)")
    .action(async (assetId: string) => {
        if (!CONTRACT_ADDRESS) {
            throw new Error("STORK_CONTRACT_ADDRESS is not set");
        }

        const wallet = getWallet();
        const assetIdBytes = hexStringToByteArray(assetId);

        try {
            const result = await wallet.rest.move.viewFunction(
                CONTRACT_ADDRESS,
                "stork",
                "get_temporal_numeric_value_unchecked",
                [],
                [bcs.vector(bcs.u8()).serialize(assetIdBytes).toBase64()]
            );
            console.log(`Temporal numeric value: ${JSON.stringify(result)}`);
        } catch (error) {
            console.error("Error querying value:", error);
        }
    });

cliProgram.parse();
