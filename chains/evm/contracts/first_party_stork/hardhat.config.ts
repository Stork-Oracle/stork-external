import type { HardhatUserConfig } from "hardhat/config";
import hardhatViem from "@nomicfoundation/hardhat-viem";
import hardhatIgnition from "@nomicfoundation/hardhat-ignition";
import { configVariable } from "hardhat/config";

const PRIVATE_KEY = configVariable("PRIVATE_KEY");

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
    hardhatLocal: {
      type: "http",
      url: "http://localhost:8545",
      chainId: 31337,
    },
    plumeTestnet: {
      type: "http",
      url: "https://testnet-rpc.plume.org",
      chainId: 98867,
      accounts: [PRIVATE_KEY],
    },
    bakerlooTestnet: {
      type: "http",
      url: "https://autonity.rpc.web3cdn.network/testnet",
      chainId: 65010004,
      accounts: [PRIVATE_KEY],
    },
  },
};

export default config;
