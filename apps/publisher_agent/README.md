# Stork Publisher Agent

The easiest way to become a Stork Publisher is to run the Stork Publisher Agent docker container on your infrastructure and send price updates to the Agent through a local websocket. The Stork Publisher Agent will sign your price updates with your private key and send them to the Stork Network.

## Getting Started

### Setup
To run the agent on your infrastructure, you'll need to first create a `config.json` file containing all non-secret configuration and a `keys.json` file containing all keys including secret configuration, and then run the Stork Publisher Agent docker container.

Here we open a websocket at `ws://localhost:5216/publish` to receive price updates. The agent will sign and send the latest price every 500 ms (clock updates) or when the price has changed by more than 0.1% (delta updates). 

`config.json` follows the structure of the [Config struct](pkg/config.go):
```json
{
  "SignatureTypes": ["evm"],
  "IncomingWsPort": 5216
}
```

You can use the [generate_keys.py](../../utils/python/src/generate_keys/generate_keys.py) script to generate your EVM and/or Stork keys and to build the `keys.json` file. Make sure not to check this file into version control or share it in any way. 

You can also generate your own keys and build your own `keys.json` file - it follows the structure of the [Keys struct](pkg/config.go):
```json
{
  "EvmPrivateKey": "0x8b558d5fc31eb64bb51d44b4b28658180e96764d5d5ac68e6d124f86f576d9de",
  "EvmPublicKey": "0x99e295e85cb07c16b7bb62a44df532a7f2620237",
  "OracleId": "oracl"
}
```

The keys.json file can be substituted or combined with environment variables. If the same property has a value set in both the keys.json and environment variable, the environment variable takes priority.

The available environment variables are:
| Key | Description | Type |
|---------| --------- | --------- |
| STORK_EVM_PRIVATE_KEY | EVM Private Key | hex string |
| STORK_EVM_PUBLIC_KEY | EVM Public Key | hex string |
| STORK_STARK_PRIVATE_KEY | Stark Private Key | hex string |
| STORK_STARK_PUBLIC_KEY | Stark Public Key | hex string |
| STORK_ORACLE_ID | Oracle ID | string |
| STORK_PULL_BASED_AUTH | Pull-based websocket auth token | string |

You will need to send your Public keys to Stork before running your publisher agent so that we can whitelist them.

> [!CAUTION]
> NEVER SHARE YOUR PRIVATE KEYS WITH ANYONE, INCLUDING ANYONE CLAIMING TO BE A MEMBER OF STORK. WE WILL NEVER ASK YOU FOR IT.

> [!IMPORTANT]
> When running in production, it is important to use environment variables via your infrastructure's secrets manager rather than a `keys.json` file.


You can run the docker container like this using keys.json:
```bash
docker run --platform linux/arm64 --pull always --restart always --name publisher-agent -p 5216:5216 -v /home/ubuntu/config.json:/etc/config.json -v /home/ubuntu/keys.json:/etc/keys.json -d --log-opt max-size=1g storknetwork/publisher-agent:v1.0.3 start -c /etc/config.json -k /etc/keys.json
```

or using environment variables:
```bash
docker run --platform linux/arm64 --pull always --restart always --name publisher-agent -p 5216:5216 -v /home/ubuntu/config.json:/etc/config.json -e STORK_EVM_PRIVATE_KEY="0x8b558d5fc31eb64bb51d44b4b28658180e96764d5d5ac68e6d124f86f576d9de" -e STORK_EVM_PUBLIC_KEY="0x99e295e85cb07c16b7bb62a44df532a7f2620237" -e STORK_STARK_PRIVATE_KEY="0x66253bdeb3c1a235cf4376611e3a14474e2c00fd2fb225f9a388faae7fb095a" -e STORK_STARK_PUBLIC_KEY="0x418d3fd8219a2cf32a00d458f61802d17f01c5bcde5a4f82008ee4a7c8e9a06" -e STORK_ORACLE_ID="czowx" -e -d --log-opt max-size=1g storknetwork/publisher-agent:latest start -c /etc/config.
```

The command will pull the docker image from our registry and run it in detached mode. If the container crashes it will automatically restart. This example assumes your config files are located in `/home/ubuntu` and that you're using port 5216 for the incoming websocket.

Check `docker logs -f publisher-agent` for any error logs once you've launched the agent.

Note that you may need to change the `--platform` argument if you're using an amd64 architecture.

You should run the publisher agent from infrastructure in Tokyo (ideally AWS availability zone `ap-northeast-1a`) to ensure your updates reach the Stork Network quickly.
### Publishing Prices
To publish a price to the Stork Network, you can connect to your Stork Publisher Agent's local port and send prices for each asset.

In practice you'll probably want to use a websocket client in your chosen language, but here's an example using bash for simplicity:
```bash
ubuntu@ip-10-0-30-216:~$ wscat -c "ws://localhost:5216/publish"
Connected (press CTRL+C to quit)
> {"type": "prices", "data": [{"t":1725930515326901000, "a": "BTCUSD", "v": 57565.21}, {"t":1725930515326901500, "a": "ETHUSD", "v": 2565.21}]}
```
The only information you need to pass is the asset's name, its price and the timestamp you observed that price in nanos. You can pass multiple price updates in one websocket message.

## Signing Every Update
To have the agent sign and send every update it receives (rather than using clock and delta update logic), add the configuration `"SignEveryUpdate": true` to your `config.json`.

If you have very fast-updating price feeds or many assets, signing every price update can be CPU-intensive.

## Using a Pull-Based Websocket
If you already have a websocket server which accepts subscriptions and outputs prices, you can leave out the `IncomingWsPort` configuration and instead set `PullBasedWsUrl` in your config.json, plus `PullBasedWsSubscriptionRequest` and `PullBasedAuth` if needed.

Note that this assumes your websocket server outputs price updates that are structured like this:
```json
{"type":"prices","data":[{"t":1725931226413064599,"a":"1000000BONKUSD","p":17.17585875},{"t":1725931226413065579,"a":"1000000BONKUSDMARK","p":17.167358324999995}}
```
