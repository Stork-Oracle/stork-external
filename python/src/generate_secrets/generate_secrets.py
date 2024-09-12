#!/usr/bin/env python3

import argparse
import eth_account
import os
import random
import secrets
import string
import json

# local
import starknet


def new_id(size=6, chars=string.ascii_lowercase):
    return ''.join(random.choice(chars) for _ in range(size))


def main():
    parser = argparse.ArgumentParser(description='Generate secrets.json file for the Stork Publisher Agent')

    parser.add_argument('--oracle-id',
                        required=True,
                        help='The 5 character name for the publisher, e.g. "nahsr"')
    parser.add_argument('--signature-types',
                        required=True,
                        nargs='+',
                        choices=["stark", "evm"],
                        help='The signature types you want to use for this publisher agent, space separated')
    parser.add_argument('--stork-auth',
                        required=True,
                        help='The stork auth header the publisher should use')
    parser.add_argument('--pull-based-auth',
                        required=False,
                        help='The auth token to use for your pull-based price source, if using')


    args = parser.parse_args()

    if len(args.oracle_id) != 5:
        parser.error('oracle name must be exactly 5 characters long')

    secrets_dict = {
        "OracleId": args.oracle_id,
        "StorkAuth": args.stork_auth,
    }
    if args.pull_based_auth:
        secrets_dict["PullBasedAuth"] = args.pull_based_auth

    if "evm" in args.signature_types:
        evm_private_token_hex = secrets.token_hex(32)
        evm_private_key = "0x" + evm_private_token_hex
        evm_account = eth_account.Account.from_key(evm_private_key)
        evm_public_key = evm_account.address
        secrets_dict["EvmPrivateKey"] = evm_private_key
        secrets_dict["EvmPublicKey"] = evm_public_key

    if "stark" in args.signature_types:
        stark_private_key = starknet.get_random_private_key()
        stark_public_key = starknet.private_to_stark_key(stark_private_key)
        stark_private_key_hex = hex(stark_private_key)
        stark_public_key_hex = hex(stark_public_key)
        secrets_dict["StarkPrivateKey"] = stark_private_key_hex
        secrets_dict["StarkPublicKey"] = stark_public_key_hex

    out_filepath = os.path.join(os.path.dirname(__file__), f'secrets.json')

    with open(out_filepath, 'w') as f:
        f.write(json.dumps(secrets_dict))

    print(f'Wrote secrets file to {out_filepath}')


if __name__ == "__main__":
    main()
