import type { HardhatUserConfig } from "hardhat/config";
import hardhatViem from "@nomicfoundation/hardhat-viem";
import hardhatIgnition from "@nomicfoundation/hardhat-ignition";
import { configVariable } from "hardhat/config";

const config: HardhatUserConfig = {
  plugins: [hardhatViem, hardhatIgnition],
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
    localhost: {
      type: "http",
      chainType: "l1",
      url: "http://localhost:8545",
    },
    hardhatLocal: {
      type: "http",
      chainType: "l1",
      url: "http://localhost:8545",
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
};

export default config;
