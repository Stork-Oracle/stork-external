# Stork Publisher Agent

The easiest way to become a Stork Publisher is to run the Stork Publisher Agent docker container on your infrastructure and send price updates to the Agent through a local websocket. The Stork Publisher Agent will sign your price updates with your private key and send them to the Stork Network.

## Getting Started

### Setup
To run the agent on your infrastructure, you'll need to first create a `config.json` file containing all non-secret configuration and a `secrets.json` file containing all secret configuration, and then run the Stork Publisher Agent docker container.

Here we open a websocket at `ws://localhost:5216/publish` to receive price updates. The agent will sign and send the latest price every 500 ms (clock updates) or when the price has changed by more than 0.1% (delta updates). 

`config.json` follows the structure of the [ConfigFile struct](config.go):
```json
{
  "SignatureTypes": ["evm"],
  "IncomingWsPort": 5216
}
```

You'll need to generate your own EVM and/or Stark keys, choose a 5 character oracle id and get a StorkAuth key from Stork.

`secrets.json` follows the structure of the [KeysFile struct](config.go):
```json
{
  "EvmPrivateKey": "0x8b558d5fc31eb64bb51d44b4b28658180e96764d5d5ac68e6d124f86f576d9de",
  "EvmPublicKey": "0x99e295e85cb07c16b7bb62a44df532a7f2620237",
  "OracleId": "oracl",
  "StorkAuth": "fake_auth"
}
```

You can run the docker container like this:
```bash
docker run --platform linux/arm64 --pull always --restart always --name publisher-agent -p 5216:5216 -v /home/ubuntu/config.json:/etc/config.json -v /home/ubuntu/secrets.json:/etc/secrets.json -d --log-opt max-size=1g storknetwork/stork-cli:v1.0.0 publisher-agent -c /etc/config.json -k /etc/secrets.json
```
The command will pull the docker image from our registry and run it in detached mode. If the container crashes it will automatically restart. This example assumes your config files are located in `/home/ubuntu` and that you're using port 5216 for the incoming websocket.

Note that you may need to change the `--platform` argument if you're using an amd64 architecture.

### Publishing Prices
To publish a price to the Stork Network, you can connect to your Stork Publisher Agent's local port and send prices for each asset.

In practice you'll probably want to use a websocket client in your chosen language, but here's an example using bash for simplicity:
```bash
ubuntu@ip-10-0-30-216:~$ wscat -c "ws://localhost:5216/publish"
Connected (press CTRL+C to quit)
> {"type": "prices", "data": [{"t":1725930515326901000, "a": "BTCUSD", "v": 57565.21}, {"t":1725930515326901500, "a": "ETHUSD", "v": 2565.21}]}
```
The only information you need to pass is the asset's name, its price and the timestamp you observed that price in nanos. You can pass multiple price updates in one websocket message.

## Give Stork Your Public Key
Let Stork know your public key(s) so we can route your data appropriately.

## Signing Every Update
To have the agent sign and send every update it receives (rather than using clock and delta update logic), add the configuration `"SignEveryUpdate": true` to your `config.json`.

If you have very fast-updating price feeds or many assets, signing every price update can be CPU-intensive.

## Using a Pull-Based Websocket
If you already have a websocket server which accepts subscriptions and outputs prices, you can leave out the `IncomingWsPort` configuration and instead set `PullBasedWsUrl` in your config.json, plus `PullBasedWsSubscriptionRequest` and `PullBasedAuth` if needed.

Note that this assumes your websocket server outputs price updates that are structured like this:
```json
{"type":"prices","data":[{"t":1725931226413064599,"a":"1000000BONKUSD","p":17.17585875},{"t":1725931226413065579,"a":"1000000BONKUSDMARK","p":17.167358324999995}}
```