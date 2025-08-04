import { createConfig } from 'fuels';

export default createConfig({
  contracts: ['../'],
  output: './types',
  privateKey: process.env.PRIVATE_KEY,
  providerUrl: process.env.PROVIDER_URL,
  forcBuildFlags: ['--release'],
});

/**
 * Check the docs:
 * https://docs.fuel.network/docs/fuels-ts/fuels-cli/config-file/
 */
