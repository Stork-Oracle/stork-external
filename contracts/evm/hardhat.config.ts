import { HardhatUserConfig } from "hardhat/config";
import "@nomicfoundation/hardhat-toolbox";
import "hardhat-gas-reporter"

require("@openzeppelin/hardhat-upgrades");

import './tasks/deploy';
import './tasks/upgrade';
import './tasks/interact';
import './tasks/print-abi';

import { vars } from "hardhat/config";

const PRIVATE_KEY = vars.get("PRIVATE_KEY");
const ARBISCAN_API_KEY = vars.get("ARBISCAN_API_KEY");

const config: HardhatUserConfig = {
  solidity: "0.8.24",
  networks: {
    inMemoryNode: {
      url: "http://127.0.0.1:8545",
      chainId: 31337,
      loggingEnabled: true,
    },
    arbitrumSepolia: {
      url: "https://sepolia-rollup.arbitrum.io/rpc",
      accounts: [PRIVATE_KEY],
      chainId: 421614,
    },
  },
  etherscan: {
    apiKey: {
      arbitrumSepolia: ARBISCAN_API_KEY,
    },
    customChains: [
      {
        network: "arbitrumSepolia",
        chainId: 421614,
        urls: {
          apiURL: "https://api-sepolia.arbiscan.io/api",
          browserURL: "https://sepolia.arbiscan.io/",
        },
      },
    ],
  },
};

export default config;
