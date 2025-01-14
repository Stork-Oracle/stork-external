import { HardhatUserConfig } from "hardhat/config";
import "@nomicfoundation/hardhat-toolbox";
import "@nomicfoundation/hardhat-verify";
import "@openzeppelin/hardhat-upgrades";

import "@matterlabs/hardhat-zksync";
import "@matterlabs/hardhat-zksync-upgradable";

import './tasks/deploy';
import './tasks/upgrade';
import './tasks/interact';
import './tasks/print-abi';
import './tasks/deploy-zk';
import './tasks/upgrade-zk';

const config: HardhatUserConfig = {
  solidity: "0.8.24",
  zksolc: {
    version: "latest",
    settings: {},
  },
  defaultNetwork: "inMemoryNode",
  networks: {
    hardhat: {},
    inMemoryNode: {
      url: "http://127.0.0.1:8545",
      chainId: 31337,
      loggingEnabled: true,
    }
  },
};

export default config;
