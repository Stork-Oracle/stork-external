import { Command } from "commander";
import { Provider, Wallet, bn, getAllDecodedLogs, keccak256} from "fuels";
import { Example } from "./types/contracts/Example";
import fs from "fs";

const contractIds = JSON.parse(fs.readFileSync("types/contract-ids.json", "utf8"));

const PRIVATE_KEY: string | undefined = process.env.PRIVATE_KEY;
const PROVIDER_URL: string | undefined = process.env.PROVIDER_URL;
let EXAMPLE_CONTRACT_ADDRESS: string | undefined = process.env.EXAMPLE_CONTRACT_ADDRESS;

if (!EXAMPLE_CONTRACT_ADDRESS) {
    console.log("EXAMPLE_CONTRACT_ADDRESS not set, attempting default from types/contract-ids.json");
    EXAMPLE_CONTRACT_ADDRESS = contractIds.example;
}

const program = new Command();
program
    .name("stork-example")
    .description("Fuel Stork example client")
    .version("0.1.0");

program
    .command("read-price")
    .description("Read price from Stork feed")
    .argument("<asset>", "Asset identifier (will be hashed)")
    .argument("<stork_contract_address>", "Stork contract address")
    .action(async (asset: string, storkContractAddress: string) => {
        if (!PRIVATE_KEY || !PROVIDER_URL) {
            throw new Error("PRIVATE_KEY and PROVIDER_URL must be set");
        }

        let encAssetId = `0x${Buffer.from(keccak256(Buffer.from(asset))).toString('hex')}`;

        const provider = new Provider(PROVIDER_URL!);
        const wallet = Wallet.fromPrivateKey(PRIVATE_KEY!, provider);
        const exampleContract = new Example(EXAMPLE_CONTRACT_ADDRESS!, wallet);
        try {
            const result = await exampleContract.functions.use_stork_price(encAssetId, storkContractAddress).get();
            let receipts = result.callResult.receipts;
            let logs = getAllDecodedLogs({receipts, mainAbi: exampleContract.interface.jsonAbi});
            let log = logs.logs[0] as any;
            let timestamp = log.timestamp;
            let quantizedValue = log.quantized_value;
            let upper = quantizedValue.underlying.upper;
            let lower = quantizedValue.underlying.lower;
            let mask64Bits = bn("18446744073709551615"); // 2^64 - 1
            let reconstructed = upper.shln(64).add(lower).sub(bn(1).shln(127));
            console.log("Temporal numeric value:");
            console.log("  Timestamp:", timestamp.toString());
            console.log("  Value:", reconstructed.toString());

        } catch (e: any) {
            console.error("Detailed error:", JSON.stringify(e.metadata || e, null, 2));
            throw e;
        }
    });

program.parse();
