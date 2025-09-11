import { Command } from "commander";
import { SigningCosmWasmClient } from "@cosmjs/cosmwasm-stargate";
import { DirectSecp256k1HdWallet } from "@cosmjs/proto-signing";
import { Decimal } from "@cosmjs/math";
import { InstantiateMsg } from "./client/Example.types";
import { ExampleClient } from "./client/Example.client";
import keccak256 from 'keccak256';

const DEFAULT_RPC_URL = process.env.RPC_URL;
const EXAMPLE_CONTRACT_ADDRESS = process.env.EXAMPLE_CONTRACT_ADDRESS;
const MNEMONIC = process.env.MNEMONIC;

const DEFAULT_GAS_PRICE = process.env.GAS_PRICE;
const DEFAULT_DENOM = process.env.NATIVE_DENOM;
const PREFIX = process.env.CHAIN_PREFIX;


async function getSender() {
    if (!MNEMONIC) {
        throw new Error("MNEMONIC environment variable is not set");
    }
    const options = {
        prefix: PREFIX
    }
    const wallet = await DirectSecp256k1HdWallet.fromMnemonic(MNEMONIC, options);
    const [firstAccount] = await wallet.getAccounts();
    return firstAccount.address;
}

async function getSigningClient(): Promise<SigningCosmWasmClient> {
    if (!DEFAULT_RPC_URL) {
        throw new Error("RPC_URL environment variable is not set");
    }
    if (!DEFAULT_GAS_PRICE) {
        throw new Error("GAS_PRICE environment variable is not set");
    }
    if (!PREFIX) {
        throw new Error("CHAIN_PREFIX environment variable is not set");
    }
    if (!DEFAULT_DENOM) {
        throw new Error("NATIVE_DENOM environment variable is not set");
    }
    if (!MNEMONIC) {
        throw new Error("MNEMONIC environment variable is not set");
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
    .name("example")
    .description("CosmWasm Stork example client")
    .version("0.1.0");

cliProgram.command("instantiate")
    .description("Instantiate the example contract")
    .argument("<code-id>", "The code ID of the uploaded contract")
    .argument("<stork-contract-address>", "The address of the Stork contract")
    .action(async (codeId: string, storkContractAddress: string) => {
        const sender = await getSender();
        const client = await getSigningClient();

        const msg: InstantiateMsg = {
            stork_contract_address: storkContractAddress
        }

        const result = await client.instantiate(
            sender,
            parseInt(codeId),
            msg,
            "Example",
            "auto",
        )

        console.log("Contract instantiated:", {
            contractAddress: result.contractAddress,
            transactionHash: result.transactionHash
        });
    });

cliProgram.command("read-price")
    .description("Call the use_stork_price function")
    .argument("<asset>", "Plaintext Asset identifier (will be hashed)")
    .action(async (asset: string) => {
        if (!EXAMPLE_CONTRACT_ADDRESS) {
            throw new Error("EXAMPLE_CONTRACT_ADDRESS environment variable is not set");
        }
    const sender = await getSender();
    const client = await getSigningClient();

    const exampleClient = new ExampleClient(client, sender, EXAMPLE_CONTRACT_ADDRESS);
    
    const feedId = keccak256(asset) as Uint8Array;

    const response = await exampleClient.useStorkPrice({
        feedId: Array.from(feedId)
    });

    const event = response.events.find(e => e.type === "wasm-stork_price_used");
    const price = event?.attributes.find(a => a.key === "value");

    console.log("Price:", price?.value);

});

cliProgram.parse();
