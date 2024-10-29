import { Command } from "commander";

import * as anchor from "@coral-xyz/anchor";
import { PublicKey } from "@solana/web3.js";
import { BN } from "bn.js";
import { Stork } from "../target/types/stork";

const anchorProviderUrl = process.env.ANCHOR_PROVIDER_URL;
const anchorWallet = process.env.ANCHOR_WALLET;

// Constants
const STORK_FEED_SEED = Buffer.from("stork_feed");
const STORK_CONFIG_SEED = Buffer.from("stork_config");
const STORK_TREASURY_SEED = Buffer.from("stork_treasury");

const idl = require("../target/idl/stork.json");
console.log("Loading IDL...");

if (!anchorProviderUrl || !anchorWallet) {
  throw new Error("ANCHOR_PROVIDER_URL and ANCHOR_WALLET must be set");
}

function hexStringToByteArray(hexString) {
  if (hexString.startsWith("0x")) {
    hexString = hexString.slice(2);
  }
  return Array.from(Buffer.from(hexString, "hex"));
}

function hexArrayToString(hexArray: number[]) {
  return Buffer.from(hexArray).toString("hex");
}

const initializeCliProgram = (): {
  program: anchor.Program<Stork>;
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

  console.log("Program ID:", idl.address);
  return {
    program: new anchor.Program<Stork>(idl, provider),
    payer: payerKeypair,
  };
};

const cliProgram = new Command();
cliProgram
  .name("admin")
  .description("Solana Stork admin client")
  .version("0.1.0");

const DEFAULT_SINGLE_UPDATE_FEE_IN_LAMPORTS = 1;
const DEFAULT_STORK_SOLANA_ADDRESS =
  "stk8Pgk9NBcMR3qcfjdqewQsSuAQHqw9j1CKd6dB5GL";
const DEFAULT_STORK_EVM_ADDRESS = "0x0a803F9b1CCe32e2773e0d2e98b37E0775cA5d44";

cliProgram
  .command("initialize")
  .description("Initialize the Stork program")
  .argument(
    "[stork_solana_address]",
    "The Stork Solana address",
    (value) => value,
    DEFAULT_STORK_SOLANA_ADDRESS
  )
  .argument(
    "[stork_evm_address]",
    "The Stork EVM address",
    (value) => value,
    DEFAULT_STORK_EVM_ADDRESS
  )
  .argument(
    "[single_update_fee_in_lamports]",
    "The single update fee in lamports",
    (value) => parseInt(value),
    DEFAULT_SINGLE_UPDATE_FEE_IN_LAMPORTS
  )
  .action(
    async (
      storkSolanaAddress,
      storkEvmPublicKey,
      singleUpdateFeeInLamports
    ) => {
      const { program, payer } = initializeCliProgram();
      try {
        await program.methods
          .initialize(
            new PublicKey(storkSolanaAddress),
            hexStringToByteArray(storkEvmPublicKey),
            new BN(singleUpdateFeeInLamports)
          )
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
          .rpc();

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
  );

cliProgram
  .command("stork-evm-public-key")
  .description("Get the EVM address")
  .action(async () => {
    const { program } = initializeCliProgram();

    const [configPda] = await PublicKey.findProgramAddressSync(
      [STORK_CONFIG_SEED],
      program.programId
    );

    const config = await program.account.storkConfig.fetch(configPda);
    console.log(hexArrayToString(config.storkEvmPublicKey));
  });

cliProgram
  .command("stork-solana-public-key")
  .description("Get the Solana address")
  .action(async () => {
    const { program } = initializeCliProgram();

    const [configPda] = await PublicKey.findProgramAddressSync(
      [STORK_CONFIG_SEED],
      program.programId
    );

    const config = await program.account.storkConfig.fetch(configPda);
    console.log(config.storkSolPublicKey.toBase58());
  });

cliProgram
  .command("single-update-fee")
  .description("Get the single update fee")
  .action(async () => {
    const { program } = initializeCliProgram();

    const [configPda] = await PublicKey.findProgramAddressSync(
      [STORK_CONFIG_SEED],
      program.programId
    );
    const config = await program.account.storkConfig.fetch(configPda);
    console.log(config.singleUpdateFeeInLamports.toString());
  });

cliProgram
  .command("update-stork-evm-public-key")
  .description("Update the EVM public key")
  .argument("evm_public_key", "The EVM public key")
  .action(async (evmPublicKey) => {
    const { program } = initializeCliProgram();
    const configPda = await PublicKey.findProgramAddressSync(
      [STORK_CONFIG_SEED],
      program.programId
    );
    await program.methods
      .updateStorkEvmPublicKey(hexStringToByteArray(evmPublicKey))
      .accounts({
        config: configPda,
      })
      .rpc();
  });

cliProgram
  .command("update-stork-solana-public-key")
  .description("Update the Solana public key")
  .argument("solana_public_key", "The Solana public key")
  .action(async (solanaPubKey) => {
    const { program } = initializeCliProgram();
    const [configPda] = await PublicKey.findProgramAddressSync(
      [STORK_CONFIG_SEED],
      program.programId
    );
    await program.methods
      .updateStorkSolPublicKey(new PublicKey(solanaPubKey))
      .accounts({
        config: configPda,
      })
      .rpc();
    console.log("Solana public key updated successfully.");
  });

cliProgram
  .command("update-single-update-fee")
  .description("Update the single update fee")
  .argument("fee", "The new single update fee in lamports")
  .action(async (fee) => {
    const { program } = initializeCliProgram();
    const [configPda] = await PublicKey.findProgramAddressSync(
      [STORK_CONFIG_SEED],
      program.programId
    );
    await program.methods
      .updateSingleUpdateFeeInLamports(new BN(fee))
      .accounts({
        config: configPda,
      })
      .rpc();
    console.log("Single update fee updated successfully.");
  });

async function fetchTransactionDetails(connection, transactionSignature) {
  let transactionDetails = null;
  const maxRetries = 5;
  const delay = 2000; // 2 seconds

  for (let attempt = 0; attempt < maxRetries; attempt++) {
    transactionDetails = await connection.getParsedTransaction(
      transactionSignature,
      {
        commitment: "confirmed",
      }
    );

    if (transactionDetails) {
      break;
    }

    console.log(
      `Retrying to fetch transaction details... Attempt ${attempt + 1}`
    );
    await new Promise((resolve) => setTimeout(resolve, delay));
  }

  return transactionDetails;
}

const writeToFeedsAction = async (
  assetPairs,
  restUrl,
  authKey,
  { treasuryId, report = false }
) => {
  console.log("Writing to feeds...");
  const { program, payer } = initializeCliProgram();
  try {
    const result = await fetch(
      `${restUrl}/v1/prices/latest\?assets\=${assetPairs}`,
      {
        headers: {
          Authorization: `Basic ${authKey}`,
        },
      }
    );
    const rawJson = await result.text();
    const safeJsonText = rawJson.replace(
      /(?<!["\d])\b\d{16,}\b(?!["])/g, // Regex to find large integers not already in quotes
      (match) => `"${match}"` // Convert large numbers to strings
    );

    const responseData = JSON.parse(safeJsonText);

    if (treasuryId === "") {
      treasuryId = Math.floor(Math.random() * 256);
      console.log(`Generated random treasury ID: ${treasuryId}`);
    } else {
      console.log(`Using provided treasury ID: ${JSON.stringify(treasuryId)}`);
    }

    const [treasuryPDA] = await PublicKey.findProgramAddressSync(
      [STORK_TREASURY_SEED, new Uint8Array([treasuryId])],
      program.programId
    );

    const instructions = await Promise.all(Object.values(responseData.data).map((data: any) => {
      const updateData = {
        temporalNumericValue: {
          timestampNs: new anchor.BN(
            data.stork_signed_price.timestamped_signature.timestamp
          ),
          quantizedValue: new anchor.BN(data.stork_signed_price.price),
        },
        id: hexStringToByteArray(data.stork_signed_price.encoded_asset_id),
        publisherMerkleRoot: hexStringToByteArray(
          data.stork_signed_price.publisher_merkle_root
        ),
        valueComputeAlgHash: hexStringToByteArray(
          data.stork_signed_price.calculation_alg.checksum
        ),
        r: hexStringToByteArray(
          data.stork_signed_price.timestamped_signature.signature.r
        ),
        s: hexStringToByteArray(
          data.stork_signed_price.timestamped_signature.signature.s
        ),
        v: hexStringToByteArray(
          data.stork_signed_price.timestamped_signature.signature.v
        )[0],
        treasuryId,
      };

      return program.methods
        .updateTemporalNumericValueEvm(updateData)
        .accounts({
          treasury: treasuryPDA,
          payer: payer.publicKey,
        })
        .instruction();
    }));


    const transaction = new anchor.web3.Transaction();

    transaction.add(...instructions);

    const transactionSignature = await program.provider.sendAndConfirm(
      transaction
    );

    if (report) {
      const transactionDetails = await fetchTransactionDetails(
        program.provider.connection,
        transactionSignature
      );

      console.log(transactionDetails);

      if (transactionDetails && transactionDetails.meta) {
        console.log(
          "Compute Units Used:",
          transactionDetails.meta.computeUnitsConsumed
        );
      } else {
        console.log("Failed to retrieve transaction details.");
      }
    }

    console.log(`Feeds updated successfully!`);
  } catch (error) {
    console.error("Error writing to feed:", error);
  }
};

cliProgram
  .command("bootstrap-treasuries")
  .description("Bootstrap the treasuries")
  .argument("endpoint", "The REST endpoint")
  .argument("auth_key", "The auth key")
  .action(async (restUrl, authKey) => {
    const btcAssetPair = "BTCUSD";

    for (let i = 0; i < 256; i++) {
      await writeToFeedsAction(btcAssetPair, restUrl, authKey, {
        treasuryId: i,
      });
    }
  });

cliProgram
  .command("write-to-feeds")
  .description("Write to feeds")
  .argument("asset_pairs", "The asset pairs (comma separated)")
  .argument("endpoint", "The REST endpoint")
  .argument("auth_key", "The auth key")
  .option("-t, --treasury-id", "The treasury ID", "")
  .option("-r, --report", "Report the results", false)
  .action(writeToFeedsAction);

cliProgram
  .command("get-feed")
  .description("Get a feed")
  .argument("feed_id", "The feed ID")
  .action(async (feedId) => {
    const { program } = initializeCliProgram();
    const [feedPda] = await PublicKey.findProgramAddressSync(
      [STORK_FEED_SEED, Buffer.from(hexStringToByteArray(feedId))],
      program.programId
    );
    const feed = await program.account.temporalNumericValueFeed.fetch(feedPda);
    console.log(feed.latestValue.quantizedValue.toString());
    console.log(feed.latestValue.timestampNs.toString());
  });

cliProgram
  .command("get-treasury-balance")
  .argument("treasury_id", "The treasury ID")
  .action(async (treasuryId) => {
    const { program } = initializeCliProgram();

    const [treasuryPda] = await PublicKey.findProgramAddressSync(
      [STORK_TREASURY_SEED, new Uint8Array([treasuryId])],
      program.programId
    );
    const balance = await program.provider.connection.getBalance(treasuryPda);
    console.log(`Treasury balance ${treasuryId}`, balance);
  });

cliProgram.parse();
