import hardhatToolboxViemPlugin from "@nomicfoundation/hardhat-toolbox-viem";
import { configVariable, defineConfig } from "hardhat/config";
import {
  verificationFeeInWei,
  signerAddresses,
  updateVerificationFeeInWei,
  addSignerAddress,
  removeSignerAddress,
  version,
} from "./tasks/admin.js";

export default defineConfig({
  plugins: [hardhatToolboxViemPlugin],
  tasks: [
    verificationFeeInWei,
    signerAddresses,
    updateVerificationFeeInWei,
    addSignerAddress,
    removeSignerAddress,
    version,
  ],
  solidity: {
    npmFilesToBuild: [
      "@openzeppelin/contracts/proxy/ERC1967/ERC1967Proxy.sol",
    ],
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
