#!/usr/bin/env python3

import argparse
import json
import os
import secrets

import eth_account

# local
import starknet


def main():
    parser = argparse.ArgumentParser(
        description="Generate a keys.json file for the Stork Publisher Agent"
    )

    parser.add_argument(
        "--oracle-id",
        required=True,
        help='The 5 character name for the publisher, e.g. "nahsr"',
    )
    parser.add_argument(
        "--signature-types",
        required=True,
        nargs="+",
        choices=["stark", "evm"],
        help="The signature types you want to use for this publisher agent, space separated",
    )
    parser.add_argument(
        "--pull-based-auth",
        required=False,
        help="The auth token for your pull-based price source, if using",
    )
    parser.add_argument(
        "--output-path",
        required=False,
        default="/tmp/publisher-agent/",
        help="The directory to write your key to",
    )
    args = parser.parse_args()

    if len(args.oracle_id) != 5:
        parser.error("oracle id  must be exactly 5 characters long")

    keys_dict = {
        "OracleId": args.oracle_id,
    }
    if args.pull_based_auth:
        keys_dict["PullBasedAuth"] = args.pull_based_auth

    if "evm" in args.signature_types:
        evm_private_token_hex = secrets.token_hex(32)
        evm_private_key = "0x" + evm_private_token_hex
        evm_account = eth_account.Account.from_key(evm_private_key)
        evm_public_key = evm_account.address
        keys_dict["EvmPrivateKey"] = evm_private_key
        keys_dict["EvmPublicKey"] = evm_public_key

    if "stark" in args.signature_types:
        stark_private_key = starknet.get_random_private_key()
        stark_public_key = starknet.private_to_stark_key(stark_private_key)
        stark_private_key_hex = hex(stark_private_key)
        stark_public_key_hex = hex(stark_public_key)
        keys_dict["StarkPrivateKey"] = stark_private_key_hex
        keys_dict["StarkPublicKey"] = stark_public_key_hex

    os.makedirs(args.output_path, exist_ok=True)
    out_filepath = os.path.join(args.output_path, "keys.json")

    with open(out_filepath, "w") as f:
        f.write(json.dumps(keys_dict, indent=2))

    print(f"Wrote keys file to {out_filepath}")


if __name__ == "__main__":
    main()
