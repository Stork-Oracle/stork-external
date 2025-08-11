import { Command } from "commander";
import * as anchor from "@coral-xyz/anchor";
import { PublicKey } from "@solana/web3.js";
import { Example } from "../target/types/example";

const anchorProviderUrl = process.env.ANCHOR_PROVIDER_URL;
const anchorWallet = process.env.ANCHOR_WALLET;

if (!anchorProviderUrl || !anchorWallet) {
  throw new Error("ANCHOR_PROVIDER_URL and ANCHOR_WALLET must be set");
}

function hexStringToByteArray(hexString) {
  if (hexString.startsWith("0x")) {
    hexString = hexString.slice(2);
  }
  return Array.from(Buffer.from(hexString, "hex"));
}

const initializeCliProgram = (): {
  program: anchor.Program<Example>;
  payer: anchor.web3.Keypair;
} => {
  console.log("Initializing provider...");
  const provider = anchor.AnchorProvider.env();
  const fs = require("fs");

  const payerKeypair = anchor.web3.Keypair.fromSecretKey(
    Uint8Array.from(
      JSON.parse(fs.readFileSync(process.env.ANCHOR_WALLET!, "utf-8"))
    )
  );

  anchor.setProvider(provider);

  const idl = require("../target/idl/example.json");
  console.log("Program ID:", idl.address);

  return {
    program: new anchor.Program<Example>(idl, provider),
    payer: payerKeypair,
  };
};

const cliProgram = new Command();
cliProgram
  .name("example")
  .description("Solana Stork example client")
  .version("0.1.0");

cliProgram
  .command("read-price")
  .description("Read latest price from a feed")
  .argument("feed_id", "The feed ID (hex format)")
  .action(async (feedId) => {
    const { program } = initializeCliProgram();
    try {
      const feedIdBytes = hexStringToByteArray(feedId);
      const response = await program.methods
        .readPrice(feedIdBytes)
        .accounts({})
        .rpc();
      
      console.log("Price read successfully from feed:", feedId);
      console.log("TX:", response);
    } catch (error) {
      console.error("Error reading price:", error);
      if (error instanceof anchor.web3.SendTransactionError) {
        console.error("Transaction logs:", error.logs);
      }
    }
  });

cliProgram.parse();
