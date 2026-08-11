import hardhatToolboxViemPlugin from "@nomicfoundation/hardhat-toolbox-viem";
import { configVariable, defineConfig } from "hardhat/config";
import {
  verificationFeeInWei,
  signerAddresses,
  updateVerificationFeeInWei,
  addSignerAddress,
  removeSignerAddress,
  version,
  verifyPayload,
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
    verifyPayload,
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
  verify: {
    etherscan: {
      apiKey: configVariable("ETHERSCAN_API_KEY"),
    },
    sourcify: {
      enabled: false,
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
      url: 'https://eth-sepolia-testnet.api.pocket.network',
      accounts: [configVariable("DEPLOYER_PRIVATE_KEY")],
    },
    relay: {
      type: "http",
      chainType: "generic",
      chainId: 537713,
      url: "https://rpc.chain.relay.link",
      accounts: [configVariable("DEPLOYER_PRIVATE_KEY")],
    }
  },
  chainDescriptors: {
    537713: {
      name: "Relay",
      chainType: "generic",
      blockExplorers: {
        blockscout: {
          name: "Relay Explorer",
          url: "https://explorer.chain.relay.link",
          apiUrl: "https://explorer.chain.relay.link/api",
        },
      },
    },
  },
});
