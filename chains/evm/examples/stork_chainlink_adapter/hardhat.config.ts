import { HardhatUserConfig } from "hardhat/config";
import "@nomicfoundation/hardhat-toolbox";
import "@nomicfoundation/hardhat-ignition-ethers";

import './tasks/get_latest_round_data';

const config: HardhatUserConfig = {
  solidity: "0.8.24",
  networks: {
    hardhat: {},
    inMemoryNode: {
      url: "http://127.0.0.1:8545",
      chainId: 31337,
    }
  },
};

export default config;
