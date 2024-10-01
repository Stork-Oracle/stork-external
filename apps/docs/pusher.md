## Usage

1. Create an `asset-config.yaml` file. This file should be structured as follows:

```yaml
assets:
    BTCUSD:
        # The asset's symbol, used to subscribe to the asset on the Stork network
        asset_id: BTCUSD
        # The asset's encoded ID, used to write the asset's data to the Stork contract. This is the keccak256 hash of the asset's symbol
        # Subscribe to the asset on the Stork network to get this value
        encoded_asset_id: 0x7404e3d104ea7841c3d9e6fd20adfe99b4ad586bc08d8f3bd3afef894cf184de
        # If the data feed is not updated by any pusher within this period the asset should be added to the batched updates
        fallback_period_sec: 60
        # If the data feed changes by more than this percentage, the asset should be added to the batched updates
        percent_change_threshold: 1
```

See `sample.asset-config.yaml` for an example.

2. Create a `private-key.secret` file. This file should contain the private key of the user's wallet. This is needed to pay gas/transaction fees.

3. Run the pusher with your desired configurations

For full explanation of the flags, run:
```
go run . evm-push --help
```

```
go run ./cmd/pusher/main.go evm \
    -w wss://api.jp.stork-oracle.network \
    -a <stork-api-key> \
    -c <chain-rpc-url> \
    -x <contract-id> \
    -f <asset-config-file> \
    -m <private-key-file>
```

## Development

Generate the contract bindings
```
abigen --abi ../contracts/evm/stork.abi --pkg main --type StorkContract --out stork_contract.go
```

## Running on ec2

The pusher runs on a per chain basis. This example assumes that the log driver is AWS Cloudwatch.

## Deploy

1. Install docker

2. Setup `.asset-config.yaml` and `.secret` files in user home directory, e.g. `/home/ec2-user`

3. Run the docker command (polygon testnet example below, replace with the correct values)

```bash
docker run \
    -e AWS_REGION=ap-northeast-1 \
    --pull always \
    --name evm-polygon-testnet \
    -v /home/ec2-user/polygon.asset-config.yaml:/etc/asset-config.yaml \
    -v /home/ec2-user/polygon-testnet.secret:/etc/private-key.secret \
    -itd --restart=on-failure \
    --log-driver=awslogs \
    --log-opt awslogs-group=/aws/ec2/dev-apps-evm-pusher \
    --log-opt awslogs-stream=polygon-testnet \
    --log-opt mode=non-blocking \
    --log-opt max-buffer-size=4m \
    storknetwork/pusher:v1.0.0 evm \
    -w wss://api.jp.stork-oracle.network \
    -a <stork-api-key> \
    -c https://rpc-amoy.polygon.technology \
    -x 0xacc0a0cf13571d30b4b8637996f5d6d774d4fd62 \
    -f /etc/asset-config.yaml \
    -m /etc/private-key.secret \
    -b 60
```
