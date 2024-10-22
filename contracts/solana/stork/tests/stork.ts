import * as anchor from "@coral-xyz/anchor";
import { Program } from "@coral-xyz/anchor";
import { Stork } from "../target/types/stork";
import { LAMPORTS_PER_SOL } from "@solana/web3.js";
import * as assert from "assert";

function hexStringToByteArray(hexString) {
  return Array.from(Buffer.from(
    hexString,
    "hex"
  ))
}

describe("Stork", () => {
  const provider = anchor.AnchorProvider.env();
  anchor.setProvider(provider);

  const program = anchor.workspace.Stork as Program<Stork>;
  const storkSignerSolKeypair = anchor.web3.Keypair.generate();
  const storkSignerEvmPublicKey = hexStringToByteArray("0a803F9b1CCe32e2773e0d2e98b37E0775cA5d44");
  const deployerKeypair = anchor.web3.Keypair.generate();
  const nonAdminKeypair = anchor.web3.Keypair.generate();

  const SINGLE_UPDATE_FEE_IN_LAMPORTS = 1000000; // 0.001 SOL

  before(async () => {
    // Request airdrop for deployerKeypair
    const airdropSignature = await provider.connection.requestAirdrop(
      deployerKeypair.publicKey,
      2 * LAMPORTS_PER_SOL // Airdrop 2 SOL
    );
    await provider.connection.confirmTransaction({
      signature: airdropSignature,
      blockhash: (await provider.connection.getLatestBlockhash()).blockhash,
      lastValidBlockHeight: (
        await provider.connection.getLatestBlockhash()
      ).lastValidBlockHeight,
    });
  });

  function initializeStork(deployingKeypair: anchor.web3.Keypair) {
    return program.methods
      .initialize(
        storkSignerSolKeypair.publicKey,
        storkSignerEvmPublicKey,
        new anchor.BN(SINGLE_UPDATE_FEE_IN_LAMPORTS)
      )
      .accounts({
        owner: deployingKeypair.publicKey,
      })
      .remainingAccounts([
        {
          pubkey: anchor.web3.SystemProgram.programId,
          isWritable: false,
          isSigner: false,
        },
      ])
      .signers([deployingKeypair])
      .rpc();
  }

  describe("initialize", () => {
    it("fails admin tasks before initialization", async () => {
      try {
        await program.methods
          .updateSingleUpdateFeeInLamports(new anchor.BN(120))
          .accounts({
            owner: deployerKeypair.publicKey,
          })
          .signers([deployerKeypair])
          .rpc();
      } catch (error) {
        assert.ok(error instanceof anchor.AnchorError, "Expected AnchorError");
        assert.strictEqual(
          error.error.errorCode.code,
          "AccountNotInitialized",
          "Expected AccountNotInitialized error"
        );
      }
    });

    it("initializes successfully", async () => {
      await initializeStork(deployerKeypair);

      const [configPda, _] = anchor.web3.PublicKey.findProgramAddressSync(
        [Buffer.from("stork_config")],
        program.programId
      );

      const configAccount = await program.account.storkConfig.fetch(configPda);

      assert.strictEqual(
        configAccount.storkSolPublicKey.toBase58(),
        storkSignerSolKeypair.publicKey.toBase58(),
        "Stork public key mismatch"
      );
      assert.deepStrictEqual(
        configAccount.storkEvmPublicKey,
        storkSignerEvmPublicKey,
        "Stork EVM public key mismatch"
      );
      assert.strictEqual(
        configAccount.singleUpdateFeeInLamports.toNumber(),
        SINGLE_UPDATE_FEE_IN_LAMPORTS,
        "Single update fee in lamports mismatch"
      );
      assert.strictEqual(
        configAccount.owner.toBase58(),
        deployerKeypair.publicKey.toBase58(),
        "Owner public key mismatch"
      );
    });

    it("fails if already initialized", async () => {
      try {
        await initializeStork(nonAdminKeypair);

        assert.fail("Expected initialization to fail, but it succeeded");
      } catch (error) {
        assert.ok(error instanceof Error, "Expected an error to be thrown");
        assert.ok(
          error.message.includes("custom program error: 0x0"),
          "Expected error message to include 'custom program error: 0x0'"
        );
      }
    });
  });



  describe("update_temporal_numeric_value_evm", () => {
    it("Creates feed with initial value", async () => {
      const id = hexStringToByteArray(
        "7404e3d104ea7841c3d9e6fd20adfe99b4ad586bc08d8f3bd3afef894cf184de"
      );
      const treasuryId = 1;
      const [treasuryPda] = await anchor.web3.PublicKey.findProgramAddressSync(
        [Buffer.from("stork_treasury"), Buffer.from([treasuryId])],
        program.programId
      );
      const updateData = {
        temporalNumericValue: {
          timestampNs: new anchor.BN("1722632569208762117"),
          quantizedValue: new anchor.BN("62507457175499998000000"),
        },
        id: id,
        publisherMerkleRoot: hexStringToByteArray("e5ff773b0316059c04aa157898766731017610dcbeede7d7f169bfeaab7cc318"),
        valueComputeAlgHash: hexStringToByteArray("9be7e9f9ed459417d96112a7467bd0b27575a2c7847195c68f805b70ce1795ba"),
        r: hexStringToByteArray("b9b3c9f80a355bd0cd6f609fff4a4b15fa4e3b4632adabb74c020f5bcd240741"),
        s: hexStringToByteArray("16fab526529ac795108d201832cff8c2d2b1c710da6711fe9f7ab288a7149758"),
        v: 28,
        treasuryId,
      };

      const tx = await program.methods
        .updateTemporalNumericValueEvm(updateData)
        .accounts({
          payer: provider.wallet.publicKey,
          treasury: treasuryPda
        })
        .rpc();

      const [feedPda, _] = anchor.web3.PublicKey.findProgramAddressSync(
        [Buffer.from("stork_feed"), id],
        program.programId
      );
      const feedAccount = await program.account.temporalNumericValueFeed.fetch(
        feedPda
      );
      for (let i = 0; i < 32; i++) {
        assert.strictEqual(
          feedAccount.id[i],
          parseInt(id[i].toFixed(0)),
          `Feed ID byte at position ${i} does not match`
        );
      }

      assert.strictEqual(
        feedAccount.latestValue.timestampNs.toString(),
        updateData.temporalNumericValue.timestampNs.toString(),
        "Timestamp not updated correctly"
      );
      assert.strictEqual(
        feedAccount.latestValue.quantizedValue.toString(),
        updateData.temporalNumericValue.quantizedValue.toString(),
        "Quantized value not updated correctly"
      );
    });

    it("Fails if ids don't match", async () => {
      const oldId = hexStringToByteArray(
        "0x7404e3d104ea7841c3d9e6fd20adfe99b4ad586bc08d8f3bd3afef894cf184de"
      );
      const newId = hexStringToByteArray(
        "0x7404e3d104ea7841c3d9e6fd20adfe99b4ad586bc08d8f3bd3afef894cf18000"
      );
      const treasuryId = 1;
      const [treasuryPda] = await anchor.web3.PublicKey.findProgramAddressSync(
        [Buffer.from("stork_treasury"), Buffer.from([treasuryId])],
        program.programId
      );
      const [feedPda, _] = anchor.web3.PublicKey.findProgramAddressSync(
        [Buffer.from("stork_feed"), oldId],
        program.programId
      );
      const updateData = {
        temporalNumericValue: {
          timestampNs: new anchor.BN("1720722087644999936"),
          quantizedValue: new anchor.BN("60000000000000000000000"),
        },
        id: newId,
        publisherMerkleRoot: Buffer.from("example data"),
        valueComputeAlgHash: Buffer.from("example data"),
        treasuryId
      };
      try {
        await program.methods
          .updateTemporalNumericValueEvm(updateData)
          .accounts({
            payer: provider.wallet.publicKey,
            feed: feedPda,
            treasury: treasuryPda,
          })
          .rpc();

        assert.fail("Expected update to fail, but it succeeded");
      } catch (error) {
        assert.ok(error instanceof anchor.AnchorError, "Expected AnchorError");
        assert.strictEqual(
          error.error.errorCode.code,
          "ConstraintSeeds",
          "Expected ConstraintSeeds error"
        );
      }
    });
  });

  describe("update_single_update_fee_in_lamports", () => {
    it("Updates single update fee in lamports", async () => {
      const newFee = 2000000; // 0.002 SOL

      await program.methods
        .updateSingleUpdateFeeInLamports(new anchor.BN(newFee))
        .accounts({
          owner: deployerKeypair.publicKey,
        })
        .signers([deployerKeypair])
        .rpc();

      const [configPda, _] = anchor.web3.PublicKey.findProgramAddressSync(
        [Buffer.from("stork_config")],
        program.programId
      );
      const configAccount = await program.account.storkConfig.fetch(configPda);
      assert.strictEqual(
        configAccount.singleUpdateFeeInLamports.toNumber(),
        newFee,
        "Single update fee not updated correctly"
      );
    });

    it("Fails if not owner", async () => {
      try {
        await program.methods
          .updateSingleUpdateFeeInLamports(new anchor.BN(2000000))
          .accounts({
            owner: nonAdminKeypair.publicKey,
          })
          .signers([nonAdminKeypair])
          .rpc();

        assert.fail("Expected update to fail, but it succeeded");
      } catch (error) {
        assert.ok(error instanceof anchor.AnchorError, "Expected AnchorError");
        assert.strictEqual(
          error.error.errorCode.code,
          "Unauthorized",
          "Expected Unauthorized error"
        );
      }
    });
  });

  describe("update_stork_sol_public_key", () => {
    const newStorkSolPublicKey = anchor.web3.Keypair.generate().publicKey;
    it("Updates Stork public key", async () => {
      await program.methods
        .updateStorkSolPublicKey(newStorkSolPublicKey)
        .accounts({
          owner: deployerKeypair.publicKey,
        })
        .signers([deployerKeypair])
        .rpc();

      const [configPda, _] = anchor.web3.PublicKey.findProgramAddressSync(
        [Buffer.from("stork_config")],
        program.programId
      );
      const configAccount = await program.account.storkConfig.fetch(configPda);
      assert.strictEqual(
        configAccount.storkSolPublicKey.toBase58(),
        newStorkSolPublicKey.toBase58(),
        "Stork public key not updated correctly"
      );
    });

    it("Fails if not owner", async () => {
      try {
        await program.methods
          .updateStorkSolPublicKey(newStorkSolPublicKey)
          .accounts({
            owner: nonAdminKeypair.publicKey,
          })
          .signers([nonAdminKeypair])
          .rpc();

        assert.fail("Expected update to fail, but it succeeded");
      } catch (error) {
        assert.ok(error instanceof anchor.AnchorError, "Expected AnchorError");
      }
    });
  });

  describe("update_stork_evm_public_key", () => {
    const newStorkEvmPublicKey = hexStringToByteArray("0a803F9b1CCe32e2773e0d2e98b37E0775cA5d44");
    it("Updates Stork public key", async () => {
      await program.methods
        .updateStorkEvmPublicKey(newStorkEvmPublicKey)
        .accounts({
          owner: deployerKeypair.publicKey,
        })
        .signers([deployerKeypair])
        .rpc();

      const [configPda, _] = anchor.web3.PublicKey.findProgramAddressSync(
        [Buffer.from("stork_config")],
        program.programId
      );
      const configAccount = await program.account.storkConfig.fetch(configPda);
      assert.deepStrictEqual(
        configAccount.storkEvmPublicKey,
        newStorkEvmPublicKey,
        "Stork public key not updated correctly"
      );
    });

    it("Fails if not owner", async () => {
      try {
        await program.methods
          .updateStorkEvmPublicKey(newStorkEvmPublicKey)
          .accounts({
            owner: nonAdminKeypair.publicKey,
          })
          .signers([nonAdminKeypair])
          .rpc();

        assert.fail("Expected update to fail, but it succeeded");
      } catch (error) {
        assert.ok(error instanceof anchor.AnchorError, "Expected AnchorError");
        assert.strictEqual(
          error.error.errorCode.code,
          "Unauthorized",
          "Expected Unauthorized error"
        );
      }
    });
  });

  describe("transfer_ownership", () => {
    const newOwner = anchor.web3.Keypair.generate();

    it("Transfers ownership", async () => {
      await program.methods
        .transferOwnership(newOwner.publicKey)
        .accounts({
          owner: deployerKeypair.publicKey,
        })
        .signers([deployerKeypair])
        .rpc();

      const [configPda, _] = anchor.web3.PublicKey.findProgramAddressSync(
        [Buffer.from("stork_config")],
        program.programId
      );
      const configAccount = await program.account.storkConfig.fetch(configPda);
      assert.strictEqual(
        configAccount.owner.toBase58(),
        newOwner.publicKey.toBase58(),
        "Owner not updated correctly"
      );
    });

    it("Fails if not owner", async () => {
      try {
        await program.methods
          .transferOwnership(newOwner.publicKey)
          .accounts({
            owner: nonAdminKeypair.publicKey,
          })
          .signers([nonAdminKeypair])
          .rpc();

        assert.fail("Expected update to fail, but it succeeded");
      } catch (error) {
        assert.ok(error instanceof anchor.AnchorError, "Expected AnchorError");
        assert.strictEqual(
          error.error.errorCode.code,
          "Unauthorized",
          "Expected Unauthorized error"
        );
      }
    });
  });
});
