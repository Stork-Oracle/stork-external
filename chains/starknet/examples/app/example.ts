// Reads a Stork price feed from a deployed contract.
//
//   STARKNET_RPC_URL=... STORK_CONTRACT_ADDRESS=0x... ENCODED_ASSET_ID=0x... npx tsx example.ts
//
// The encoded asset id is the keccak256 of the asset symbol; subscribe to the asset on the Stork
// network to get it.

import { readFileSync } from "fs";
import { join } from "path";
import { RpcProvider, Contract, cairo, json, num, type Abi, type BigNumberish } from "starknet";

const RPC_URL = process.env.STARKNET_RPC_URL;
const CONTRACT_ADDRESS = process.env.STORK_CONTRACT_ADDRESS;

// BTCUSD on the Stork network.
const ENCODED_ASSET_ID =
  process.env.ENCODED_ASSET_ID ??
  "0x7404e3d104ea7841c3d9e6fd20adfe99b4ad586bc08d8f3bd3afef894cf184de";

// The Starknet field prime. Cairo serialises i128 as a field element, negative values as P - |v|.
const FIELD_PRIME = 2n ** 251n + 17n * 2n ** 192n + 1n;

function decodeI128(value: bigint): bigint {
  return value >= FIELD_PRIME - 2n ** 127n ? value - FIELD_PRIME : value;
}

type Value = { timestamp_ns: bigint; quantized_value: BigNumberish };

async function read(stork: Contract, entrypoint: string): Promise<Value> {
  return (await stork.call(entrypoint, [cairo.uint256(ENCODED_ASSET_ID)])) as Value;
}

function report(label: string, value: Value): void {
  const quantized = decodeI128(BigInt(value.quantized_value));

  console.log(`${label}:`);
  console.log(`  timestamp_ns:    ${value.timestamp_ns}`);
  console.log(`  quantized_value: ${quantized}`);
  // Stork quantises prices to 18 decimals.
  console.log(`  price:           ${Number(quantized) / 1e18}`);
  console.log(`  as of:           ${new Date(Number(value.timestamp_ns / 1_000_000n)).toISOString()}\n`);
}

async function main(): Promise<void> {
  if (!RPC_URL) {
    throw new Error("STARKNET_RPC_URL is not set");
  }
  if (!CONTRACT_ADDRESS) {
    throw new Error("STORK_CONTRACT_ADDRESS is not set");
  }

  const abiPath = join(
    __dirname,
    "..",
    "..",
    "contracts",
    "target",
    "dev",
    "stork_Stork.contract_class.json",
  );
  const { abi } = json.parse(readFileSync(abiPath).toString("ascii"));

  const provider = new RpcProvider({ nodeUrl: RPC_URL });
  const stork = new Contract({ abi: abi as Abi, address: CONTRACT_ADDRESS, providerOrAccount: provider });

  console.log(`contract: ${num.toHex(CONTRACT_ADDRESS)}`);
  console.log(`feed:     ${ENCODED_ASSET_ID}\n`);

  // The checked read rejects anything older than the contract's valid time period. This is what
  // production code should use: it fails loudly rather than letting you act on a stale price.
  try {
    report("get_temporal_numeric_value_v1 (checked)", await read(stork, "get_temporal_numeric_value_v1"));
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    if (message.includes("stale value") || message.includes("0x53746f726b3a207374616c652076616c7565")) {
      console.log("get_temporal_numeric_value_v1 (checked): rejected, the feed is stale");
      console.log("  the contract refuses values older than its valid_time_period_seconds\n");
    } else {
      throw error;
    }
  }

  // The unchecked read returns whatever is stored. Only use it if you handle staleness yourself.
  report("get_temporal_numeric_value_unsafe_v1", await read(stork, "get_temporal_numeric_value_unsafe_v1"));
}

main().catch((error) => {
  console.error(`error: ${error instanceof Error ? error.message : error}`);
  process.exit(1);
});
