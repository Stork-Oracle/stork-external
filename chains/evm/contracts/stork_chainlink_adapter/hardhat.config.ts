import { HardhatUserConfig } from "hardhat/config";
import "@nomicfoundation/hardhat-toolbox";

import { vars } from "hardhat/config";

const PRIVATE_KEY = vars.get("PRIVATE_KEY");
const ETHERSCAN_API_KEY = vars.get("ETHERSCAN_API_KEY");

import './tasks/get_latest_round_data';

const config: HardhatUserConfig = {
  solidity: "0.8.28",
  networks: {
    inMemoryNode: {
      url: "http://127.0.0.1:8545",
      chainId: 31337,
      loggingEnabled: true,
    },
    baseSepolia: {
      url: "https://sepolia.base.org",
      accounts: [PRIVATE_KEY],
      chainId: 84532,
    },
    mainnet: {
      url: "https://ethereum-rpc.publicnode.com",
      accounts: [PRIVATE_KEY],
      chainId: 1,
    },
    plumeMainnet: {
      url: "https://phoenix-rpc.plumenetwork.xyz",
      accounts: [PRIVATE_KEY],
      chainId: 98866
    }
  },
  etherscan: {
    apiKey: {
      mainnet: ETHERSCAN_API_KEY,
      plumeMainnet: 'fake',
    },
    customChains: [
      {
        network: "plumeMainnet",
        chainId: 98866,
        urls: {
          apiURL: "https://explorer-plume-mainnet-1.t.conduit.xyz/api",
          browserURL: "https://explorer.plume.org"
        }
      }
    ]
  }
};

export default config;
