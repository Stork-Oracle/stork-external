import { task } from 'hardhat/config';

export async function getImplementationAddress(proxyAddress: string) {
    // @ts-expect-error upgrades is loaded in hardhat/config
    const implementationAddress = await upgrades.erc1967.getImplementationAddress(proxyAddress);
    return implementationAddress;
}

task('get-impl', 'Get the implementation address of a proxy')
    .addPositionalParam<string>('proxyAddress', 'The address of the proxy contract')
    .setAction(async (args: any) => {
        const implementationAddress = await getImplementationAddress(args.proxyAddress);
        console.log(`Implementation address: ${implementationAddress}`);
    });
    