import { HardhatUserConfig } from "hardhat/config";
import "@nomicfoundation/hardhat-toolbox";
import "@nomicfoundation/hardhat-verify";
require("@openzeppelin/hardhat-upgrades");

import "@matterlabs/hardhat-zksync";
import "@matterlabs/hardhat-zksync-upgradable";

import './tasks/deploy';
import './tasks/upgrade';
import './tasks/interact';
import './tasks/print-abi';
import './tasks/deploy-zk';
import './tasks/upgrade-zk';
import './tasks/get-impl';

import { vars } from "hardhat/config";

const PRIVATE_KEY = vars.get("PRIVATE_KEY");
const ARBISCAN_API_KEY = vars.get("ARBISCAN_API_KEY");
const POLYGON_API_KEY = vars.get("POLYGON_API_KEY");
const ETHERSCAN_API_KEY = vars.get("ETHERSCAN_API_KEY");
const CORE_TESTNET_API_KEY = vars.get("CORE_TESTNET_API_KEY");
const CORE_MAINNET_API_KEY = vars.get("CORE_MAINNET_API_KEY");
const ROOTSTOCK_TESTNET_API_KEY = vars.get("ROOTSTOCK_TESTNET_API_KEY");
const SCROLL_MAINNET_API_KEY = vars.get("SCROLL_MAINNET_API_KEY");
const SONEIUM_MAINNET_RPC_URL = vars.get("SONEIUM_MAINNET_RPC_URL");
const ETHERSCAN_SOPHON_API_KEY = vars.get("ETHERSCAN_SOPHON_API_KEY");
const SONEIUM_MAINNET_BLOCKSCOUT_URL = vars.get("SONEIUM_MAINNET_BLOCKSCOUT_URL");
const CRONOS_L2_API_KEY = vars.get("CRONOS_L2_API_KEY");

const config: HardhatUserConfig = {
  solidity: "0.8.24",
  zksolc: {
    version: "1.5.15",
    settings: {
      codegen: "yul",
      // find all available options in the official documentation
      // https://era.zksync.io/docs/tools/hardhat/hardhat-zksync-solc.html#configuration
    },
  },
  defaultNetwork: "inMemoryNode",
  networks: {
    hardhat: {},
    inMemoryNode: {
      url: "http://127.0.0.1:8545",
      chainId: 31337,
      loggingEnabled: true,
    },
    abstractTestnet: {
      url: "https://api.testnet.abs.xyz",
      ethNetwork: "sepolia",
      accounts: [PRIVATE_KEY],
      chainId: 11124,
      verifyURL: "https://api-explorer-verify.testnet.abs.xyz/contract_verification",
      zksync: true,
    },
    arbitrumSepolia: {
      url: "https://sepolia-rollup.arbitrum.io/rpc",
      accounts: [PRIVATE_KEY],
      chainId: 421614,
    },
    arbitrumMainnet: {
      url: "https://arbitrum-one.rpc.grove.city/v1/01fdb492",
      accounts: [PRIVATE_KEY],
      chainId: 42161,
    },
    avalanceCChain: {
      url: "https://avalanche-c-chain-rpc.publicnode.com",
      accounts: [PRIVATE_KEY],
      chainId: 43114,
    },
    baseMainnet: {
      url: "https://mainnet.base.org",
      accounts: [PRIVATE_KEY],
      chainId: 8453,
    },
    baseSepolia: {
      url: "https://sepolia.base.org",
      accounts: [PRIVATE_KEY],
      chainId: 84532,
    },
    // you might need to use a custom RPC URL for this chain, get one from quicknode
    berachainTestnet: {
      url: "https://bartio.rpc.berachain.com/",
      accounts: [PRIVATE_KEY],
      chainId: 80084,
    },
    berachainMainnet: {
      url: "https://rpc.berachain.com/",
      accounts: [PRIVATE_KEY],
      chainId: 80094,
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
    bitlayerMainnet: {
      url: "https://rpc.bitlayer.org/",
      accounts: [PRIVATE_KEY],
      chainId: 200901,
    },
    bnbSmartChainMainnet: {
      url: "https://bsc-rpc.publicnode.com",
      accounts: [PRIVATE_KEY],
      chainId: 56,
    },
    bobSepolia: {
      url: "https://bob-sepolia.rpc.gobob.xyz/",
      accounts: [PRIVATE_KEY],
      chainId: 808813,
    },
    citreaTestnet: {
      url: "https://rpc.testnet.citrea.xyz",
      accounts: [PRIVATE_KEY],
      chainId: 5115,
    },
    coreTestnet: {
      url: "https://rpc.test.btcs.network",
      accounts: [PRIVATE_KEY],
      chainId: 1115,
    },
    coreMainnet: {
      url: "https://rpc.coredao.org/",
      accounts: [PRIVATE_KEY],
      chainId: 1116,
    },
    cronosZkEvmMainnet: {
      url: "https://mainnet.zkevm.cronos.org/",
      accounts: [PRIVATE_KEY],
      chainId: 388,
      zksync: true,
      verifyURL: "https://explorer-api.zkevm.cronos.org/api/v1/contract/verify/hardhat?apikey=" + CRONOS_L2_API_KEY,
      ethNetwork: "mainnet",
    },
    cronosZkEvmTestnet: {
      url: "https://testnet.zkevm.cronos.org/",
      accounts: [PRIVATE_KEY],
      chainId: 240,
      zksync: true,
      verifyURL: "https://explorer-api.zkevm.cronos.org/testnet/api/v1/contract/verify/hardhat?apikey=" + CRONOS_L2_API_KEY,
      ethNetwork: "mainnet",
    },
    crossFiMainnet: {
      url: "https://rpc.mainnet.ms/",
      accounts: [PRIVATE_KEY],
      chainId: 4158,
    },
    ethereumMainnet: {
      url: "https://ethereum-rpc.publicnode.com",
      accounts: [PRIVATE_KEY],
      chainId: 1,
    },
    // not currently deployed, seems to have been reset
    expchainTestnet: {
      url: "https://rpc0-testnet.expchain.ai",
      accounts: [PRIVATE_KEY],
      chainId: 18880,
    },
    filecoinCalibration: {
      url: "https://rpc.ankr.com/filecoin_testnet",
      accounts: [PRIVATE_KEY],
      chainId: 314159,
    },
    goatTestnet: {
      url: "https://rpc.testnet3.goat.network/",
      accounts: [PRIVATE_KEY],
      chainId: 48816,
    },
    goatMainnet: {
      url: "https://rpc.goat.network",
      accounts: [PRIVATE_KEY],
      chainId: 2345,
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
    hyperEvmMainnet: {
      url: "https://rpc.hyperliquid.xyz/evm",
      accounts: [PRIVATE_KEY],
      chainId: 999,
    },
    lightlinkPegasusTestnet: {
      url: "https://replicator.pegasus.lightlink.io/rpc/v1",
      accounts: [PRIVATE_KEY],
      chainId: 1891
    },
    lightlinkPhoenixMainnet: {
      url: "https://replicator.phoenix.lightlink.io/rpc/v1",
      accounts: [PRIVATE_KEY],
      chainId: 1890
    },
    lorenzoTestnet: {
      url: "https://rpc-testnet.lorenzo-protocol.xyz",
      accounts: [PRIVATE_KEY],
      chainId: 83291
    },
    mantaSepolia: {
      url: "https://manta-sepolia.rpc.caldera.xyz/http",
      accounts: [PRIVATE_KEY],
      chainId: 3441006,
    },
    mantaPacificMainnet: {
      url: "https://pacific-rpc.manta.network/http",
      accounts: [PRIVATE_KEY],
      chainId: 169,
    },
    megaEthTestnetV2: {
      url: "https://timothy.megaeth.com/rpc",
      accounts: [PRIVATE_KEY],
      chainId: 6343,
    },
    merlinTestnet: {
      url: "https://testnet-rpc.merlinchain.io/",
      accounts: [PRIVATE_KEY],
      chainId: 686868
    },
    minionsTestnet: {
      url: "https://rpc.minions.t.raas.gelato.cloud",
      accounts: [PRIVATE_KEY],
      chainId: 123420001524
    },
    mitosisPrivateTestnet: {
      url: "https://rpc.testnet.mitosis.org",
      accounts: [PRIVATE_KEY],
      chainId: 124832
    },
    mitosisTestnet: {
      url: "https://rpc.badnet.mitosis.org",
      accounts: [PRIVATE_KEY],
      chainId: 124864
    },
    mitosisDognet: {
      url: "https://rpc.dognet.mitosis.org",
      accounts: [PRIVATE_KEY],
      chainId: 124859
    },
    molten: {
      url: "https://molten.calderachain.xyz/http",
      accounts: [PRIVATE_KEY],
      chainId: 360
    },
    monadDevnet: {
      url: "https://devnet1.monad.xyz/rpc/Hr83nzcKqOU2xOPXKme4bKm3BlDdxQPH99k9NAe1",
      accounts: [PRIVATE_KEY],
      chainId: 41454
    },
    monadTestnet: {
      url: "https://testnet-rpc.monad.xyz",
      accounts: [PRIVATE_KEY],
      chainId: 10143
    },
    movementTestnet: {
      url: "https://mevm.devnet.imola.movementlabs.xyz/",
      accounts: [PRIVATE_KEY],
      chainId: 30732
    },
    omniOmega: {
      url: "https://omega.omni.network/",
      accounts: [PRIVATE_KEY],
      chainId: 164
    },
    openCampusTestnet: {
      url: "https://rpc.open-campus-codex.gelato.digital",
      accounts: [PRIVATE_KEY],
      chainId: 656476
    },
    ozeanTestnet: {
      url: "https://ozean-testnet.rpc.caldera.xyz/http",
      accounts: [PRIVATE_KEY],
      chainId: 7849306
    },
    plume: {
      url: "https://phoenix-rpc.plumenetwork.xyz",
      accounts: [PRIVATE_KEY],
      chainId: 98865
    },
    plumeTestnet: {
      url: "https://testnet-rpc.plumenetwork.xyz",
      accounts: [PRIVATE_KEY],
      chainId: 98867
    },
    plumeMainnet: {
      url: "https://rpc-plume-mainnet-1.t.conduit.xyz",
      accounts: [PRIVATE_KEY],
      chainId: 98866
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
    riseTestnet: {
      url: "https://indexing.testnet.riselabs.xyz/",
      accounts: [PRIVATE_KEY],
      chainId: 11155931,
    },
    rootstockTestnet: {
      url: `https://rpc.testnet.rootstock.io/${ROOTSTOCK_TESTNET_API_KEY}`,
      accounts: [PRIVATE_KEY],
      chainId: 31,
    },
    scrollMainnet: {
      url: "https://rpc.scroll.io/",
      accounts: [PRIVATE_KEY],
      chainId: 534352
    },
    scrollSepolia: {
      url: "https://sepolia-rpc.scroll.io/",
      accounts: [PRIVATE_KEY],
      chainId: 534351
    },
    soneiumMinato: {
      url: "https://rpc.minato.soneium.org",
      accounts: [PRIVATE_KEY],
      chainId: 1946
    },
    soneiumMainnet: {
      url: SONEIUM_MAINNET_RPC_URL,
      accounts: [PRIVATE_KEY],
      chainId: 1868
    },
    sonicMainnet: {
      url: "https://rpc.soniclabs.com",
      accounts: [PRIVATE_KEY],
      chainId: 146,
    },
    sonicTestnet: {
      url: "https://rpc.testnet.soniclabs.com",
      accounts: [PRIVATE_KEY],
      chainId: 64165,
    },
    sonicBlazeTestnet: {
      url: "https://rpc.blaze.soniclabs.com",
      accounts: [PRIVATE_KEY],
      chainId: 57054,
    },
    sophonTestnet: {
      url: "https://rpc.testnet.sophon.xyz",
      ethNetwork: "sepolia",
      accounts: [PRIVATE_KEY],
      chainId: 531050104,
      verifyURL: "https://api-explorer-verify.testnet.sophon.xyz/contract_verification",
      zksync: true,
    },
    sophonMainnet: {
      url: "https://rpc.sophon.xyz",
      ethNetwork: "mainnet",
      accounts: [PRIVATE_KEY],
      chainId: 50104,
      verifyURL: "https://verification-explorer.sophon.xyz/contract_verification",
      zksync: true,
    },
    sophonZKsyncOSTestnet: {
      url: "https://zksync-os-testnet-sophon.zksync.dev",
      accounts: [PRIVATE_KEY],
      chainId: 531050204,
    },
    storyOdysseyTestnet: {
      url: "https://rpc.odyssey.storyrpc.io",
      accounts: [PRIVATE_KEY],
      chainId: 1516,
    },
    // testnet
    tacTurin: {
      url: "https://turin.rpc.tac.build",
      accounts: [PRIVATE_KEY],
      chainId: 2390
    },
    tacMainnet: {
      url: "https://rpc.tac.build",
      accounts: [PRIVATE_KEY],
      chainId: 239
    },
    taraxaMainnet: {
      url: "https://841.rpc.thirdweb.com/fake/anotherfake",
      accounts: [PRIVATE_KEY],
      chainId: 841,
      hardfork: 'berlin',
    },
    umiDevnet: {
      url: "https://devnet.uminetwork.com/evm",
      accounts: [PRIVATE_KEY],
      chainId: 42069,
    },
    unichainSepolia: {
      url: "https://rough-misty-frog.unichain-sepolia.quiknode.pro/ddb18152added7093553cbefdb422162f325ac7a",
      accounts: [PRIVATE_KEY],
      chainId: 1301,
    },
    volmexTestnet: {
      url: "https://volmex-testnet-custom-gas-0.rpc.caldera.xyz/http",
      accounts: [PRIVATE_KEY],
      chainId: 5633311,
    },
    xlayerTestnet: {
      url: "https://xlayertestrpc.okx.com",
      accounts: [PRIVATE_KEY],
      chainId: 195
    },
    zetachainTestnet: {
      url: "https://zetachain-athens-evm.blockpi.network/v1/rpc/public",
      accounts: [PRIVATE_KEY],
      chainId: 7001
    },
    zkSyncSepolia: {
      url: "https://sepolia.era.zksync.dev",
      ethNetwork: "sepolia",
      accounts: [PRIVATE_KEY],
      chainId: 300,
      verifyURL: "https://sepolia.explorer.zksync.io/contract_verification",
      zksync: true,
    },
    zkSyncMainnet: {
      url: "https://mainnet.era.zksync.io",
      ethNetwork: "mainnet",
      accounts: [PRIVATE_KEY],
      chainId: 324,
      verifyURL: "https://explorer.zksync.io/contract_verification",
      zksync: true,
    }
  },
  etherscan: {
    // enabled: false, // uncomment this for ZKSync verifications
    apiKey: {
      arbitrumSepolia: ARBISCAN_API_KEY,
      avalanceCChain: 'fake',
      berachainTestnet: 'fake',
      bevmTestnet: 'fake',
      bitlayerTestnet: 'fake',
      bitlayerMainnet: 'fake',
      bobSepolia: 'fake',
      citreaTestnet: 'fake',
      coreTestnet: CORE_TESTNET_API_KEY,
      coreMainnet: CORE_MAINNET_API_KEY,
      expchainTestnet: 'fake',
      glueTestnet: 'fake',
      goatTestnet: 'fake',
      holesky: ETHERSCAN_API_KEY,
      lightlinkPegasusTestnet: 'fake',
      lightlinkPhoenixMainnet: 'fake',
      lorenzoTestnet: 'fake',
      mantaSepolia: 'fake',
      mantaPacificMainnet: 'fake',
      megaEthTestnetV2: 'fake',
      merlinTestnet: 'fake',
      minionsTestnet: 'fake',
      mitosisTestnet: 'fake',
      mitosisPrivateTestnet: 'fake',
      molten: 'fake',
      monadDevnet: 'fake',
      omniOmega: 'fake',
      openCampusTestnet: 'fake',
      ozeanTestnet: 'fake',
      plume: 'fake',
      plumeDevnet: 'fake',
      polygon: POLYGON_API_KEY,
      polygonAmoy: POLYGON_API_KEY,
      rootstockTestnet: 'fake',
      scrollMainnet: SCROLL_MAINNET_API_KEY,
      soneiumMainnet: 'fake',
      soneiumMinato: 'fake',
      sophonMainnet: ETHERSCAN_SOPHON_API_KEY,
      storyOdysseyTestnet: 'fake',
      tacTurin: 'fake',
      taraxaMainnet: 'fake',
      unichainSepolia: 'fake',
      volmexTestnet: 'fake',
      xlayerTestnet: 'fake',
      zetachainTestnet: 'fake',
      cronosZkEvmMainnet: CRONOS_L2_API_KEY,
      cronosZkEvmTestnet: CRONOS_L2_API_KEY,
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
        network: "avalanceCChain",
        chainId: 43114,
        urls: {
          apiURL: "https://api.routescan.io/v2/network/mainnet/evm/43114/etherscan/api",
          browserURL: "https://snowtrace.io/"
        }
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
          browserURL: "https://scan-testnet.bevm.io/"
        }
      },
      {
        network: "bitlayerMainnet",
        chainId: 200901,
        urls: {
          apiURL: "https://api.btrscan.com/scan/api",
          browserURL: "https://www.btrscan.com"
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
        network: "bobSepolia",
        chainId: 808813,
        urls: {
          apiURL: "https://bob-sepolia.explorer.gobob.xyz/api",
          browserURL: "https://bob-sepolia.explorer.gobob.xyz/"
        }
      },
      {
        network: "citreaTestnet",
        chainId: 5115,
        urls: {
          apiURL: "https://explorer.testnet.citrea.xyz/api",
          browserURL: "https://explorer.testnet.citrea.xyz/"
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
        network: "coreMainnet",
        chainId: 1116,
        urls: {
          apiURL: "https://openapi.coredao.org/api",
          browserURL: "https://scan.coredao.org/"
        }
      },
      {
        network: "cronosZkEvmMainnet",
        chainId: 388,
        urls: {
          apiURL: "https://explorer-api.zkevm.cronos.org/api/v1/contract/verify/hardhat?apikey=" + CRONOS_L2_API_KEY,
          browserURL: "https://explorer.zkevm.cronos.org/"
        }
      },
      {
        network: "cronosZkEvmTestnet",
        chainId: 240,
        urls: {
          apiURL: "https://explorer-api.zkevm.cronos.org/testnet/api/v1/contract/verify/hardhat?apikey=" + CRONOS_L2_API_KEY,
          browserURL: "https://explorer.zkevm.cronos.org/testnet"
        }
      },
      {
        network: "expchainTestnet",
        chainId: 18880,
        urls: {
          apiURL: "https://blockscout-testnet.expchain.ai/api",
          browserURL: "https://blockscout-testnet.expchain.ai"
        }
      },
      {
        network: "goatTestnet",
        chainId: 48816,
        urls: {
          apiURL: "https://explorer.testnet3.goat.network/api",
          browserURL: "https://explorer.testnet3.goat.network/"
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
        network: "lightlinkPegasusTestnet",
        chainId: 1891,
        urls: {
          apiURL: "https://pegasus.lightlink.io/api",
          browserURL: "https://pegasus.lightlink.io/"
        }
      },
      {
        network: "lightlinkPhoenixMainnet",
        chainId: 1890,
        urls: {
          apiURL: "https://phoenix.lightlink.io/api",
          browserURL: "https://phoenix.lightlink.io/"
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
        network: "mantaSepolia",
        chainId: 3441006,
        urls: {
          apiURL: "https://manta-sepolia.explorer.caldera.xyz/api",
          browserURL: "https://manta-sepolia.explorer.caldera.xyz/"
        }
      },
      {
        network: "mantaPacificMainnet",
        chainId: 169,
        urls: {
          apiURL: "https://pacific-explorer.manta.network/api",
          browserURL: "https://pacific-explorer.manta.network/"
        }
      },
      {
        network: "megaEthTestnetV2",
        chainId: 6343,
        urls: {
          apiURL: "https://megaeth-testnet-v2.blockscout.com/api",
          browserURL: "https://megaeth-testnet-v2.blockscout.com/"
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
        network: "minionsTestnet",
        chainId: 123420001524,
        urls: {
          apiURL: "https://minions.cloud.blockscout.com/api",
          browserURL: "https://minions.cloud.blockscout.com/"
        }
      },
      {
        network: "mitosisTestnet",
        chainId: 124864,
        urls: {
          apiURL: "https://blockscout.badnet.mitosis.org/api",
          browserURL: "https://blockscout.badnet.mitosis.org/"
        }
      },
      {
        network: "mitosisPrivateTestnet",
        chainId: 124832,
        urls: {
          apiURL: "https://api.routescan.io/v2/network/testnet/evm/124832/etherscan",
          browserURL: "https://testnet.mitosiscan.xyz"
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
        network: "monadDevnet",
        chainId: 41454,
        urls: {
          apiURL: "https://brightstar-884.devnet1.monad.xyz/api",
          browserURL: "https://brightstar-884.devnet1.monad.xyz/"
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
        network: "openCampusTestnet",
        chainId: 656476,
        urls: {
          apiURL: "https://edu-chain-testnet.blockscout.com/api",
          browserURL: "https://edu-chain-testnet.blockscout.com/"
        }
      },
      {
        network: "ozeanTestnet",
        chainId: 7849306,
        urls: {
          apiURL: "https://ozean-testnet.explorer.caldera.xyz/api",
          browserURL: "https://ozean-testnet.explorer.caldera.xyz/"
        }
      },
      {
        network: "plume",
        chainId: 98865,
        urls: {
          apiURL: "https://phoenix-explorer.plumenetwork.xyz/api",
          browserURL: "https://phoenix-explorer.plumenetwork.xyz/"
        }
      },
      {
        network: "plumeDevnet",
        chainId: 98864,
        urls: {
          apiURL: "https://test-explorer.plumenetwork.xyz/api",
          browserURL: "https://test-explorer.plumenetwork.xyz/"
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
        network: "scrollMainnet",
        chainId: 534352,
        urls: {
          apiURL: "https://api.scrollscan.com/api",
          browserURL: "https://scrollscan.com/"
        }
      },
      {
        network: "soneiumMainnet",
        chainId: 1868,
        urls: {
          apiURL: `${SONEIUM_MAINNET_BLOCKSCOUT_URL}/api`,
          browserURL: SONEIUM_MAINNET_BLOCKSCOUT_URL
        }
      },
      {
        network: "soneiumMinato",
        chainId: 1946,
        urls: {
          apiURL: "https://explorer-testnet.soneium.org/api/",
          browserURL: "https://explorer-testnet.soneium.org/"
        }
      },
      {
        network: "sophonMainnet",
        chainId: 50104,
        urls: {
          apiURL: "https://api.sophscan.xyz/api",
          browserURL: "https://sophscan.xyz",
        },
      },
      {
        network: "storyOdysseyTestnet",
        chainId: 1516,
        urls: {
          apiURL: "https://odyssey.storyscan.xyz/api",
          browserURL: "https://odyssey.storyscan.xyz/"
        }
      },
      {
        network: "tacTurin",
        chainId: 2390,
        urls: {
          apiURL: "https://turin.explorer.tac.build/api",
          browserURL: "https://turin.explorer.tac.build/"
        }
      },
      {
        network: "taraxaMainnet",
        chainId: 841,
        urls: {
          apiURL: "https://tara.to/api",
          browserURL: "https://tara.to"
        }
      },
      {
        network: "unichainSepolia",
        chainId: 1301,
        urls: {
          apiURL: "https://api-sepolia.uniscan.xyz/api",
          browserURL: "https://sepolia.uniscan.xyz/"
        }
      },
      {
        network: "volmexTestnet",
        chainId: 5633311,
        urls: {
          apiURL: "https://volmex-testnet-custom-gas-0.explorer.caldera.xyz/api",
          browserURL: "https://volmex-testnet-custom-gas-0.explorer.caldera.xyz"
        }
      },
      {
        network: "xlayerTestnet",
        chainId: 195,
        urls: {
          apiURL: "https://www.oklink.com/api/v5/explorer/contract/verify-source-code-plugin/XLAYER_TESTNET",
          browserURL: "https://www.oklink.com/xlayer-test"
        }
      },
      {
        network: "zetachainTestnet",
        chainId: 7001,
        urls: {
          apiURL: "https://zetachain-testnet.blockscout.com/api",
          browserURL: "https://zetachain-testnet.blockscout.com/"
        }
      }
    ],
  },
};

export default config;
