import { Command } from "commander";
import { Account, Aptos, AptosConfig, Network, Ed25519PrivateKey, PrivateKey, PrivateKeyVariants} from "@aptos-labs/ts-sdk";
import keccak256 from 'keccak256';
const EXAMPLE_PACKAGE_ADDRESS = process.env.EXAMPLE_PACKAGE_ADDRESS;
const PRIVATE_KEY = process.env.PRIVATE_KEY;

const APTOS_CONFIG = new AptosConfig({
    network: process.env.RPC_ALIAS as Network,
});

const aptos = new Aptos(APTOS_CONFIG);

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
    .name("stork-example")
    .description("Aptos Stork example client")
    .version("0.1.0");

cliProgram
    .command("read-price")
    .description("Read price from Stork feed")
    .argument("<asset>", "Asset identifier (will be hashed)")
    .action(async (asset: string) => {
        const account = getAccount();
        const contractAddress = EXAMPLE_PACKAGE_ADDRESS;
        const tx = await aptos.transaction.build.simple({
            sender: account.accountAddress,
            data: {
                function: `${contractAddress}::example::use_stork_price`,
                functionArguments: [keccak256(asset)],
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
        }
    });
cliProgram.parse();
