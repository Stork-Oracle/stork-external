# Data Provider
The Stork Data Provider is a framework to pull arbitrary numeric data across many sources. It can be used on its own, or run alongside the Stork Publisher Agent to sign the data and send it to the Stork Network.

## Adding a New Data Source
If you want to report data from a data source which does not already have an [integration](../apps/data_provider/pkg/sources), you can add your own.

To add a new source:
1. Run `make install-stork-generate` to install the `stork-generate` CLI tool.
1. Run `stork-generate generate <pascal-case-name>` to add your new data provider.
   1. TODOs [data_source.go](./lib/sources/random/data_source.go): implement your DataSource object conforming to the [DataSource interface](../apps/data_provider/pkg/types/model.go). This includes setting needed parameters, as needed initialization code, and implementing the RunDataSource method.
   1. TODOs [data_source_test.go](../apps/data_provider/pkg/sources/random/data_source_test.go): write unit tests for your data source.
   1. TODOs [config.go](../apps/data_provider/pkg/sources/random/config.go): add necessary parameters to your configuration object.
   1. TODOs [JSON Schema](https://json-schema.org/) [config](../apps/data_provider/pkg/configs/resources/source_config_schemas/random.json): make sure this reflects the structure of the configuration object in [config.go](../apps/data_provider/pkg/sources/random/config.go)
   1. TODOs [source test](../apps/data_provider/pkg/configs/source_config_tests/random_test.go): write a test to ensure that your config can be deserialized correctly and that the DataSourceId can be extracted using `GetSourceSpecificConfig`.
1. Submit a Pull Request so other developers can use your new data source!

## Configuration
The Data Provider can report many feeds, each sourced from any of the data sources implemented in [sources](../apps/data_provider/pkg/sources).

You can configure the Data Provider by passing it a [config json file](../sample.data-provider.config.json) which can be deserialized into a [DataProviderConfig](../apps/data_provider/pkg/types/model.go) object. 

The `sources` tag is a list of configurations for different feeds, where each feed has a unique `id` and a `config` which can be deserialized into the appropriate [source config](../apps/data_provider/pkg/sources/random/config.go). 

## Running Local Code
You can test the Data Provider locally by running:
```
make start-data-provider ARGS="-c apps/data_provider/pkg/configs/example_data_source_config.json --verbose"
```
You will most likely want to replace the `apps/data_provider/pkg/configs/example_data_source_config.json` with a more useful config json. Also make sure any required environment variables like API keys are set in your local environment.

Running in `--verbose` mode with no output address set will just log every price update. If you want to actually send updates somewhere (like the websocket server of your local Publisher Agent), you can pass an output address flag:
```
make start-data-provider ARGS="-c apps/data_provider/pkg/configs/example_data_source_config.json -o ws://localhost:5216/"
```

## Running Published Docker Image
If all the data sources you want to use are already merged into Stork's repo, you can just pull the latest published Data Provider docker image and supply your own config: 
```
docker run --platform linux/arm64 --pull always --restart always --name data-provider -v ./sample.data-provider.config.json:/etc/config.json -d --log-opt max-size=1g storknetwork/data-provider:latest start -c /etc/config.json -o ws://localhost:5216/
```



