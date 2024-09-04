import { HardhatUserConfig } from "hardhat/config";
import "@nomicfoundation/hardhat-toolbox";

require("@openzeppelin/hardhat-upgrades");
import "@matterlabs/hardhat-zksync";
import "@matterlabs/hardhat-zksync-upgradable";

import './tasks/deploy';
import './tasks/upgrade';
import './tasks/interact';
import './tasks/print-abi';

import { vars } from "hardhat/config";

const PRIVATE_KEY = vars.get("PRIVATE_KEY");
const ARBISCAN_API_KEY = vars.get("ARBISCAN_API_KEY");
const POLYGON_API_KEY = vars.get("POLYGON_API_KEY");
const ETHERSCAN_API_KEY = vars.get("ETHERSCAN_API_KEY");
const CORE_TESTNET_API_KEY = vars.get("CORE_TESTNET_API_KEY");
const ROOTSTOCK_TESTNET_API_KEY = vars.get("ROOTSTOCK_TESTNET_API_KEY");

const config: HardhatUserConfig = {
  solidity: "0.8.24",
  zksolc: {
    version: "latest",
    settings: {
      // find all available options in the official documentation
      // https://era.zksync.io/docs/tools/hardhat/hardhat-zksync-solc.html#configuration
    },
  },
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
    bevmTestnet: {
      url: "https://testnet.bevm.io/rpc",
      accounts: [PRIVATE_KEY],
      chainId: 11503,
    },
    bitlayerTestnet: {
      url: "https://testnet-rpc.bitlayer.org/",
      accounts: [PRIVATE_KEY],
      chainId: 200810,
    },
    coreTestnet: {
      url: "https://rpc.test.btcs.network",
      accounts: [PRIVATE_KEY],
      chainId: 1115,
    },
    glueTestnet: {
      url: "https://testnet-ws-1.server-1.glue.net/",
      accounts: [PRIVATE_KEY],
      chainId: 1300,
    },
    holesky: {
      url: "https://rpc.holesky.ethpandaops.io/",
      accounts: [PRIVATE_KEY],
      chainId: 17000
    },
    lorenzoTestnet: {
      url: "https://rpc-testnet.lorenzo-protocol.xyz",
      accounts: [PRIVATE_KEY],
      chainId: 83291
    },
    merlinTestnet: {
      url: "https://testnet-rpc.merlinchain.io/",
      accounts: [PRIVATE_KEY],
      chainId: 686868
    },
    molten: {
      url: "https://molten.calderachain.xyz/http",
      accounts: [PRIVATE_KEY],
      chainId: 360
    },
    omniOmega: {
      url: "https://omega.omni.network/",
      accounts: [PRIVATE_KEY],
      chainId: 164
    },
    polygon: {
      url: "https://polygon.llamarpc.com",
      accounts: [PRIVATE_KEY],
      chainId: 137,
    },
    polygonAmoy: {
      url: "https://rpc-amoy.polygon.technology/",
      accounts: [PRIVATE_KEY],
      chainId: 80002,
    },
    rootstockTestnet: {
      url: `https://rpc.testnet.rootstock.io/${ROOTSTOCK_TESTNET_API_KEY}`,
      accounts: [PRIVATE_KEY],
      chainId: 31,
    },
    sophonTestnet: {
      url: "https://rpc.testnet.sophon.xyz",
      accounts: [PRIVATE_KEY],
      chainId: 531050104,
      zksync: true,
      ethNetwork: "sepolia"
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
      bitlayerTestnet: 'fake',
      bevmTestnet: 'fake',
      coreTestnet: CORE_TESTNET_API_KEY,
      glueTestnet: 'fake',
      holesky: ETHERSCAN_API_KEY,
      lorenzoTestnet: 'fake',
      merlinTestnet: 'fake',
      molten: 'fake',
      omniOmega: 'fake',
      polygon: POLYGON_API_KEY,
      polygonAmoy: POLYGON_API_KEY,
      rootstockTestnet: 'fake',
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
        network: "bevmTestnet",
        chainId: 11503,
        urls: {
          apiURL: "https://scan-testnet-api.bevm.io/api",
          browserURL: "https://bevm.io"
        }
      },
      {
        network: "bitlayerTestnet",
        chainId: 200810,
        urls: {
          apiURL: "https://api-testnet.bitlayer.org/scan/api",
          browserURL: "https://testnet.btrscan.com/"
        }
      },
      {
        network: "coreTestnet",
        chainId: 1115,
        urls: {
          apiURL: "https://api.test.btcs.network/api",
          browserURL: "https://scan.test.btcs.network/"
        }
      },
      {
        network: "glueTestnet",
        chainId: 1300,
        urls: {
          apiURL: "https://backend.explorer.testnet.dev.gke.glue.net/api",
          browserURL: "https://explorer.testnet.dev.gke.glue.net/"
        }
      },
      {
        network: "lorenzoTestnet",
        chainId: 83291,
        urls: {
          apiURL: "https://scan-testnet.lorenzo-protocol.xyz/api",
          browserURL: "https://scan-testnet.lorenzo-protocol.xyz"
        }
      },
      {
        network: "merlinTestnet",
        chainId: 686868,
        urls: {
          apiURL: "https://testnet-scan.merlinchain.io/api",
          browserURL: "https://testnet-scan.merlinchain.io"
        }
      },
      {
        network: "molten",
        chainId: 360,
        urls: {
          apiURL: "https://molten.calderaexplorer.xyz/api",
          browserURL: "https://molten.calderaexplorer.xyz/"
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
        network: "rootstockTestnet",
        chainId: 31,
        urls: {
          apiURL: "https://rootstock-testnet.blockscout.com/api/",
          browserURL: "https://rootstock-testnet.blockscout.com/"
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
