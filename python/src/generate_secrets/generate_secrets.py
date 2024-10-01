#!/usr/bin/env python3

import argparse
import eth_account
import os
import random
import secrets
import string
import json
import base64

# local
import starknet


def main():
    parser = argparse.ArgumentParser(description='Generate a secrets.json file for the Stork Publisher Agent')

    parser.add_argument('--oracle-id',
                        required=True,
                        help='The 5 character name for the publisher, e.g. "nahsr"')
    parser.add_argument('--signature-types',
                        required=True,
                        nargs='+',
                        choices=["stark", "evm"],
                        help='The signature types you want to use for this publisher agent, space separated')
    parser.add_argument('--stork-user-name',
                        required=False,
                        help="The Stork user's username, e.g. myusername (must also set stork password)")
    parser.add_argument('--stork-password',
                        required=False,
                        help="The Stork user's password e.g. mypassword (must also set stork user name)")
    parser.add_argument('--stork-auth-token',
                        required=False,
                        help="The Stork user's base64 encoded username:password, e.g. bXl1c2VybmFtZTpteXBhc3N3b3Jk")
    parser.add_argument('--pull-based-auth',
                        required=False,
                        help='The auth token for your pull-based price source, if using')
    parser.add_argument('--output-path',
                        required=False,
                        default="/tmp/publisher-agent/",
                        help='The directory to write your secret to')
    args = parser.parse_args()

    if len(args.oracle_id) != 5:
        parser.error('oracle id  must be exactly 5 characters long')

    stork_auth_token = None
    if args.stork_user_name is not None and args.stork_password is not None:
        user_password = f"{args.stork_user_name}:{args.stork_password}"
        stork_auth_token = base64.b64encode(user_password.encode()).decode()
    else:
        if args.stork_auth_token is not None:
            stork_auth_token = args.stork_auth_token
        else:
            parser.error('must either set (stork user name and stork password) or (stork auth token)')

    secrets_dict = {
        "OracleId": args.oracle_id,
        "StorkAuth": stork_auth_token,
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

    os.makedirs(args.output_path, exist_ok=True)
    out_filepath = os.path.join(args.output_path, f'secrets.json')

    with open(out_filepath, 'w') as f:
        f.write(json.dumps(secrets_dict, indent=2))

    print(f'Wrote secrets file to {out_filepath}')


if __name__ == "__main__":
    main()
