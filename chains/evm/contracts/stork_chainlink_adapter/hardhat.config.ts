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
    base: {
      url: "https://mainnet.base.org",
      accounts: [PRIVATE_KEY],
      chainId: 8453,
    },
    baseSepolia: {
      url: "https://sepolia.base.org",
      accounts: [PRIVATE_KEY],
      chainId: 84532,
    },
    berachainMainnet: {
      url: "https://rpc.berachain.com/",
      accounts: [PRIVATE_KEY],
      chainId: 80094,
    },
    hyperEvmMainnet: {
      url: "https://rpc.hyperliquid.xyz/evm",
      accounts: [PRIVATE_KEY],
      chainId: 999,
    },
    mainnet: {
      url: "https://ethereum-rpc.publicnode.com",
      accounts: [PRIVATE_KEY],
      chainId: 1,
    },
    plumeMainnet: {
      url: "https://rpc.plume.org",
      accounts: [PRIVATE_KEY],
      chainId: 98866
    },
    tacMainnet: {
      url: "https://rpc.tac.build",
      accounts: [PRIVATE_KEY],
      chainId: 239
    }
  },
  etherscan: {
    apiKey: ETHERSCAN_API_KEY,
    customChains: [
      {
        network: "berachainMainnet",
        chainId: 80094,
        urls: {
          apiURL: "https://api.etherscan.io/v2/api?chain_id=80094",
          browserURL: "https://explorer.berachain.com"
        }
      },
      {
        network: "plumeMainnet",
        chainId: 98866,
        urls: {
          apiURL: "https://explorer-plume-mainnet-1.t.conduit.xyz/api",
          browserURL: "https://explorer.plume.org"
        }
      },
      {
        network: "hyperEvmMainnet",
        chainId: 999,
        urls: {
          apiURL: "https://api.etherscan.io/v2/api",
          browserURL: "https://hyperevmscan.io"
        }
      }
    ]
  }
};

export default config;
