# Stork Publisher Agent

The easiest way to become a Stork Publisher is to run the Stork Publisher Agent docker container on your infrastructure and send price updates to the Agent through a local websocket. The Stork Publisher Agent will sign your price updates with your private key and send them to the Stork Network.

## Getting Started

### Setup
To run the agent on your infrastructure, you'll need to first create a `config.json` file containing all non-secret configuration and a `secrets.json` file containing all secret configuration, and then run the Stork Publisher Agent docker container.

Here we open a websocket at `ws://localhost:5216/publish` which will both EVM sign every update.

`config.json`:
```json
{
  "SignatureTypes": ["evm"],
  "IncomingWsPort": 5216,
  "SignEveryUpdate": true
}
```

You'll need to generate your own EVM and/or Stark keys, choose a unique 5 character oracle id and get a StorkAuth key from Stork.

`secrets.json`:
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
To publish a price to the Stork Network, you can just connect to your Stork Publisher Agent's local port and send prices for each asset.

In practice you'll probably want to use a websocket client in your chosen language, but here's an example using bash for simplicity:
```bash
ubuntu@ip-10-0-30-216:~$ wscat -c "ws://localhost:5216/publish"
Connected (press CTRL+C to quit)
> {"type": "prices", "data": [{"t":1725930515326901000, "a": "BTCUSD", "v": 57565.21}, {"t":1725930515326901500, "a": "ETHUSD", "v": 2565.21}]}
```
The only information you need to pass is the asset's name, its price and the timestamp you observed that price in nanos. You can pass multiple price updates in one websocket message.

## Using Clock and Delta Triggers
If you have very fast-updating price feeds, signing every price update can be CPU-intensive. If your publisher agent is not able to keep up with the updates you send it, you may want to remove the `"SignEveryUpdate": true` configuration from your `config.json`.

Without this configuration the Stork Publisher Agent will sign the latest price every 500 ms, or if the price changes by more than 0.1%. This may significantly reduce the CPU needed for the Stork Publisher Agent.

## Using a Pull-Based Websocket
If you already have a websocket server which accepts subscriptions and outputs prices, you can leave out the `IncomingWsPort` configuration and instead set `PullBasedWsUrl` in your config.json, plus `PullBasedWsSubscriptionRequest` and `PullBasedAuth` if needed.

Note that this assumes your websocket server outputs price updates that are structured like this:
```json
{"type":"prices","data":[{"t":1725931226413064599,"a":"1000000BONKUSD","p":17.17585875,"r":"clock"},{"t":1725931226413065579,"a":"1000000BONKUSDMARK","p":17.167358324999995,"r":"clock"}}
```