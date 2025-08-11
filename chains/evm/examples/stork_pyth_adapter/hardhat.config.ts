import { HardhatUserConfig } from "hardhat/config";
import "@nomicfoundation/hardhat-toolbox";

import './tasks/get_latest_price';

const config: HardhatUserConfig = {
  solidity: "0.8.27",
};

export default config;
