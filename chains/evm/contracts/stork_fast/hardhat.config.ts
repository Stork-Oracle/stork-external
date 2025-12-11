import hardhatToolboxViemPlugin from "@nomicfoundation/hardhat-toolbox-viem";
import { configVariable, defineConfig } from "hardhat/config";
import {
  verificationFeeInWei,
  storkFastAddress,
  updateVerificationFeeInWei,
  updateStorkFastAddress,
  version,
} from "./tasks/admin";

export default defineConfig({
  plugins: [hardhatToolboxViemPlugin],
  tasks: [
    verificationFeeInWei,
    storkFastAddress,
    updateVerificationFeeInWei,
    updateStorkFastAddress,
    version,
  ],
  solidity: {
    npmFilesToBuild: [
      "@openzeppelin/contracts/proxy/transparent/ProxyAdmin.sol",
      "@openzeppelin/contracts/proxy/transparent/TransparentUpgradeableProxy.sol",
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
