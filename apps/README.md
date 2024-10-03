# Stork Apps

A suite of tools to interact with Stork's services and on-chain contracts.

## Chain Pusher

Stork signed data feeds are delivered off-chain from publishers to subscribers via Stork's aggregation network. In order for this data to be usable on-chain, it must be written to the Stork contract on any compatible network. This tool is used to push signed data feeds to the Stork contract.

Stork does not write this data to the chain directly by default, but any subscriber can choose to write the data to the chain if they so choose. This tool can be used to facilitate that process.

See [Stork Pusher Docs](docs/chain_pusher.md).

## Publisher Agent

The Stork Network receives signed data feeds from publishers and aggregates them using various Composite Oracle Services. These aggregated data feeds are then delivered to subscribers.

The easiest way to become a Stork Publisher is to run the Stork Publisher Agent docker container on your infrastructure and send price updates to the Agent through a local websocket. The Stork Publisher Agent will sign your price updates with your private key and send them to the Stork Network.

See [Stork Publisher Agent Docs](docs/publisher_agent.md).
