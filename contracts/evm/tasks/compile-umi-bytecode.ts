import { bcs } from '@mysten/bcs';
import { task } from 'hardhat/config';

const EVM_SERIALIZER = bcs.enum('ScriptOrDeployment', {
    Script: null,
    Module: null,
    EvmContract: bcs.byteVector(),
});

const serialize = (bytecode: string): string => {
    // Extract the byte array to serialize within the higher level enum
    const code = Uint8Array.from(Buffer.from(bytecode.replace('0x', ''), 'hex'));
    const evmContract = EVM_SERIALIZER.serialize({ EvmContract: code }).toBytes();
    return '0x' + Buffer.from(evmContract).toString('hex');
};

async function main() {
    // @ts-expect-error ethers is loaded in hardhat/config
    const Counter = await ethers.getContractFactory('UpgradeableStork');
    const code = serialize(Counter.bytecode);
    console.log('BYTECODE:', code);
    // Replace artifact bytecode with the printed bytecode here
    return;
}

task("compile-umi-bytecode", "A task to compile the bytecode for UMI")
  .setAction(main);
