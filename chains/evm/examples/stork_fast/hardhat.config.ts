import hardhatToolboxViemPlugin from "@nomicfoundation/hardhat-toolbox-viem";
import { configVariable, defineConfig } from "hardhat/config";
import { task } from "hardhat/config";

const useStorkFastTask = task(
  "use-stork-fast",
  "Call the useStorkFast function on the example contract."
)
  .addPositionalArgument({
    name: "exampleContractAddress",
    description: "The address of the deployed Example contract",
  })
  .addPositionalArgument({
    name: "payload",
    description: "The signed ECDSA payload (hex string)",
  })
  .setAction(() => import("./tasks/useStorkFastAction.js"))
  .build();

export default defineConfig({
  plugins: [hardhatToolboxViemPlugin],
  tasks: [useStorkFastTask],
  solidity: {
    profiles: {
      default: {
        version: "0.8.28",
      },
      production: {
        version: "0.8.28",
        settings: {
          optimizer: {
            enabled: true,
            runs: 200,
          },
        },
      },
    },
  },
  networks: {
    hardhatLocal: {
      type: "http",
      url: "http://localhost:8545",
      chainId: 31337,
    },
    hardhatMainnet: {
      type: "edr-simulated",
      chainType: "l1",
    },
    hardhatOp: {
      type: "edr-simulated",
      chainType: "op",
    },
    sepolia: {
      type: "http",
      chainType: "l1",
      url: configVariable("SEPOLIA_RPC_URL"),
      accounts: [configVariable("SEPOLIA_PRIVATE_KEY")],
    },
  },
});
