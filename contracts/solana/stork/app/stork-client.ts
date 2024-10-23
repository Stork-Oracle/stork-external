import * as anchor from "@coral-xyz/anchor";
import { SystemProgram, PublicKey } from "@solana/web3.js";
import { BN } from "bn.js";
import { Stork } from "../target/types/stork";

// Constants
const STORK_PROGRAM_ID = new PublicKey(
  "F8C2ffLfHGQTQxnix4LaroXuHh2Cdv8wJw7R9UDEfmX1"
);
const STORK_FEED_SEED = Buffer.from("stork_feed");
const STORK_CONFIG_SEED = Buffer.from("stork_config");
const STORK_TREASURY_SEED = Buffer.from("stork_treasury");

// Function to write a new value to the feed
async function writeToFeed(
  program: anchor.Program<Stork>,
  payer: anchor.web3.Keypair,
  updateData: any,
  treasuryId: number
) {
  // Derive the PDA for the feed and the treasury account
  const [treasuryPDA] = await PublicKey.findProgramAddressSync(
    [STORK_TREASURY_SEED, new Uint8Array([treasuryId])],
    program.programId
  );

  // Call the `update_temporal_numeric_value_evm` instruction

  await program.methods
    .updateTemporalNumericValueEvm(updateData)
    .accounts({
      treasury: treasuryPDA,
      payer: payer.publicKey,
    })
    .signers([payer])
    .rpc()
    .catch((err) => {
      console.error("Error:", err);
    });

  console.log("Feed updated successfully!");
}

// Function to retrieve the feed data
async function getFeed(program: anchor.Program<Stork>, feedId: any) {
  const [feedPDA] = await PublicKey.findProgramAddressSync(
    [STORK_FEED_SEED, Buffer.from(feedId)],
    program.programId
  );

  const feedAccount = await program.account.temporalNumericValueFeed.fetch(
    feedPDA
  );
  console.log("Feed Data:", feedAccount);
  return feedAccount;
}

// Function to get the config PDA
async function getConfigPda(program: anchor.Program<Stork>) {
  const [configPDA] = await PublicKey.findProgramAddressSync(
    [STORK_CONFIG_SEED],
    program.programId
  );
  return configPDA;
}

function hexStringToByteArray(hexString) {
  return Array.from(Buffer.from(hexString, "hex"));
}

// Add this new function to initialize the Stork program
async function initializeStorkProgram(
  program: anchor.Program<Stork>,
  payer: anchor.web3.Keypair
) {
  const [configPDA] = await PublicKey.findProgramAddressSync(
    [STORK_CONFIG_SEED],
    program.programId
  );

  // You'll need to replace these with actual values
  const storkEvmPublicKey = hexStringToByteArray("0a803F9b1CCe32e2773e0d2e98b37E0775cA5d44");
  const singleUpdateFeeInLamports = new BN(1);

  console.log("Config PDA:", configPDA.toBase58());
  console.log("Payer public key:", payer.publicKey.toBase58());
  console.log("Program ID:", program.programId.toBase58());

  try {
    const tx = await program.methods
      .initialize(payer.publicKey, storkEvmPublicKey, singleUpdateFeeInLamports)
      .accounts({
        owner: payer.publicKey,
      })
      .remainingAccounts([
        {
          pubkey: anchor.web3.SystemProgram.programId,
          isWritable: false,
          isSigner: false,
        },
      ])
      .signers([payer])
      .transaction();

    tx.feePayer = payer.publicKey;

    console.log("Transaction created, simulating...");
    const simulation = await program.provider.connection.simulateTransaction(tx);
    console.log("Simulation result:", simulation.value);

    if (simulation.value.err) {
      console.error("Simulation failed:", simulation.value.err);
      return;
    }

    console.log("Simulation successful, sending transaction...");
    const txid = await program.provider.sendAndConfirm(tx, [payer]);
    console.log("Transaction sent:", txid);

    console.log("Stork program initialized successfully!");
  } catch (error) {
    if (error.message.includes("already in use")) {
      console.log("Stork program already initialized.");
    } else {
      console.error("Error initializing Stork program:", error);
      if (error instanceof anchor.web3.SendTransactionError) {
        console.error("Transaction logs:", error.logs);
      }
      throw error;
    }
  }
}

// Main function to run the script
(async () => {
  console.log("Initializing provider...");
  const provider = anchor.AnchorProvider.env();
  const fs = require('fs');

  const payerKeypair = anchor.web3.Keypair.fromSecretKey(
    Uint8Array.from(JSON.parse(fs.readFileSync(process.env.ANCHOR_WALLET!, 'utf-8')))
  );

  anchor.setProvider(provider);

  console.log("Loading IDL...");
  const idl = require("../target/idl/stork.json");

  console.log("Program ID:", idl.address);

  console.log("Initializing program...");
  const program = new anchor.Program<Stork>(idl, provider);

  // Add this line to initialize the Stork program
  await initializeStorkProgram(program, payerKeypair);

  const treasuryId = 1;
  const feedId = hexStringToByteArray(
    "7404e3d104ea7841c3d9e6fd20adfe99b4ad586bc08d8f3bd3afef894cf184de"
  );
  const updateData = {
    temporalNumericValue: {
      timestampNs: new anchor.BN("1722632569208762117"),
      quantizedValue: new anchor.BN("62507457175499998000000"),
    },
    id: feedId,
    publisherMerkleRoot: hexStringToByteArray(
      "e5ff773b0316059c04aa157898766731017610dcbeede7d7f169bfeaab7cc318"
    ),
    valueComputeAlgHash: hexStringToByteArray(
      "9be7e9f9ed459417d96112a7467bd0b27575a2c7847195c68f805b70ce1795ba"
    ),
    r: hexStringToByteArray(
      "b9b3c9f80a355bd0cd6f609fff4a4b15fa4e3b4632adabb74c020f5bcd240741"
    ),
    s: hexStringToByteArray(
      "16fab526529ac795108d201832cff8c2d2b1c710da6711fe9f7ab288a7149758"
    ),
    v: 28,
    treasuryId,
  };
  // Write to the feed

  await writeToFeed(program, payerKeypair, updateData, treasuryId);

  // Retrieve the feed data
  const feedData = await getFeed(program, feedId);
  console.log("Retrieved feed data:", feedData);
})();
