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
const EXAMPLE_PACKAGE_ADDRESS = process.env.EXAMPLE_PACKAGE_ADDRESS;
const CHAIN_ID = process.env.CHAIN_ID || "initiation-2";

function getWallet(): Wallet {
    if (!MNEMONIC) {
        throw new Error("MNEMONIC environment variable is not set");
    }

    const rest = new RESTClient(RPC_URL, {
        chainId: CHAIN_ID,
        gasPrices: "0.15uinit",
        gasAdjustment: "1.5",
    });

    const mk = new MnemonicKey({ mnemonic: MNEMONIC, coinType: 60, eth: true });
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

        // Parse and display emitted events
        console.log("\nEmitted Events:");
        const events = (txInfo as any).events || [];

        for (const event of events) {
            if (event.type === 'move') {
                const attributes = event.attributes || [];
                const typeTag = attributes.find((attr: any) => attr.key === 'type_tag')?.value;
                const data = attributes.find((attr: any) => attr.key === 'data')?.value;

                // Look for our ExampleStorkPriceEvent
                if (typeTag && typeTag.includes('ExampleStorkPriceEvent')) {
                    console.log(`\n  Event Type: ${typeTag}`);
                    if (data) {
                        try {
                            const eventData = JSON.parse(data);
                            console.log(`  Event Data:`);
                            console.log(`    Timestamp (ns): ${eventData.timestamp}`);
                            console.log(`    Magnitude: ${eventData.magnitude}`);
                            console.log(`    Negative: ${eventData.negative}`);
                            console.log(`    Value: ${eventData.negative ? '-' : ''}${eventData.magnitude}`);
                        } catch (e) {
                            console.log(`  Raw Data: ${data}`);
                        }
                    }
                }
            }
        }
    }
}

const cliProgram = new Command();
cliProgram
    .name("stork-example")
    .description("Initia MiniMove Stork example client")
    .version("0.1.0");

cliProgram
    .command("read-price")
    .description("Read price from Stork feed and emit event")
    .argument("<asset>", "Asset identifier (e.g., INITUSD)")
    .action(async (asset: string) => {
        if (!EXAMPLE_PACKAGE_ADDRESS) {
            throw new Error("EXAMPLE_PACKAGE_ADDRESS is not set");
        }

        const wallet = getWallet();

        // Encode the asset ID using keccak256
        const { keccak256 } = await import("@initia/initia.js");
        const encodedAssetId = keccak256(Buffer.from(asset));
        const assetIdBytes = Array.from(encodedAssetId);

        const msg = new MsgExecute(
            wallet.key.accAddress,
            EXAMPLE_PACKAGE_ADDRESS,
            "example",
            "use_stork_price",
            [],
            [
                bcs.vector(bcs.u8()).serialize(assetIdBytes).toBase64(),
            ]
        );

        await executeTx(msg);
    });

cliProgram.parse();
