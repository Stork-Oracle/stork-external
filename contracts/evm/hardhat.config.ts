import { HardhatUserConfig } from "hardhat/config";
import "@nomicfoundation/hardhat-toolbox";

require("@openzeppelin/hardhat-upgrades");

import './tasks/deploy';
import './tasks/upgrade';
import './tasks/interact';
import './tasks/print-abi';

import { vars } from "hardhat/config";

const PRIVATE_KEY = vars.get("PRIVATE_KEY");
const ARBISCAN_API_KEY = vars.get("ARBISCAN_API_KEY");
const POLYGON_API_KEY = vars.get("POLYGON_API_KEY");
const ETHERSCAN_API_KEY = vars.get("ETHERSCAN_API_KEY");

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
    berachainTestnet: {
      url: "https://bartio.rpc.berachain.com/",
      accounts: [PRIVATE_KEY],
      chainId: 80084,
    },
    holesky: {
      url: "https://rpc.holesky.ethpandaops.io/",
      accounts: [PRIVATE_KEY],
      chainId: 17000
    },
    omniOmega: {
      url: "https://omega.omni.network/",
      accounts: [PRIVATE_KEY],
      chainId: 164
    },
    polygonAmoy: {
      url: "https://rpc-amoy.polygon.technology/",
      accounts: [PRIVATE_KEY],
      chainId: 80002,
    },
    volmexTestnet: {
      url: "https://volmex-testnet-custom-gas-0.rpc.caldera.xyz/http",
      accounts: [PRIVATE_KEY],
      chainId: 5633311,
    },
  },
  etherscan: {
    apiKey: {
      arbitrumSepolia: ARBISCAN_API_KEY,
      berachainTestnet: 'fake',
      holesky: ETHERSCAN_API_KEY,
      omniOmega: 'fake',
      polygonAmoy: POLYGON_API_KEY,
      volmexTestnet: 'fake'
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
      {
        network: "berachainTestnet",
        chainId: 80084,
        urls: {
          apiURL: "https://api.routescan.io/v2/network/testnet/evm/80084/etherscan",
          browserURL: "https://bartio.beratrail.io"
        }
      },
      {
        network: "omniOmega",
        chainId: 164,
        urls: {
          apiURL: "https://api.routescan.io/v2/network/testnet/evm/164_4/etherscan",
          browserURL: "https://omega.omniscan.network"
        }
      },
      {
        network: "volmexTestnet",
        chainId: 5633311,
        urls: {
          apiURL: "https://volmex-testnet-custom-gas-0.explorer.caldera.xyz/api",
          browserURL: "https://volmex-testnet-custom-gas-0.explorer.caldera.xyz"
        }
      }
    ],
  },
};

export default config;
