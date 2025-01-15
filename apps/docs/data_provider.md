# Data Provider
The Stork Data Provider is a framework to pull arbitrary numeric data across many sources. It can be used on its own, or run alongside the Stork Publisher Agent to sign the data and send it to the Stork Network.

## Adding a New Data Source
If you want to report data from a data source which does not already have an [integration](../lib/data_provider/sources), you can add your own.

To add a new source:
1. Add a [package](../lib/data_provider/sources/random) in the [sources directory](../lib/data_provider/sources) with your data source's name
1. Run `python3 ./apps/scripts/update_shared_data_provider_code/py` to generate some framework code so that the framework is aware of your new source.
1. Add a [data_source.go](../lib/data_provider/sources/random/data_source.go) and implement a DataSource object conforming to the [DataSource interface](../lib/data_provider/types/model.go). This object will contain most of your source-specific logic, but it can leverage tools like the [scheduler](../lib/data_provider/sources/scheduler.go) or [ethereum_utils](../lib/data_provider/sources/ethereum_utils.go) as needed.
1. Add a [data_source_test.go](../lib/data_provider/sources/random/data_source_test.go) to unit test your data source.
1. Add a [config.go](../lib/data_provider/sources/random/config.go) which defines a configuration object corresponding to a single data feed in your source
   1. This config object must include a `DataSource` field.
1. Add a [JSON Schema](https://json-schema.org/) [config](../lib/data_provider/configs/resources/source_config_schemas/random.json) in the configs package defining the structure of the configuration object in [config.go](../lib/data_provider/sources/random/config.go)
1. Add a [config test](../lib/data_provider/configs/source_config_tests/random_test.go) to the configs package which tests that a valid Data Provider config json using your source:
   1. Passes schema validations
   1. Can be deserialized into your configuration object correctly
   1. Can be used to extract your DataSourceId using `GetSourceSpecificConfig`
1. Add an [init.go](../lib/data_provider/sources/random/init.go) to your package. This file can be almost identical for every source. This file is responsible for:
   1. Defining the DataSourceId variable for this source (which must be the same as the package name)
   1. Defining and registering a DataSourceFactory (which will just call to your DataSource constructor)
   1. Asserting the source's DataSource and DataSourceFactory satisfy our interfaces
   1. Defining a function to deserialize the source's config object
1. Submit a Pull Request so other developers can use your new data source!

## Configuration
The Data Provider can report many feeds, each sourced from any of the data sources implemented in [sources](../lib/data_provider/sources).

You can configure the Data Provider by passing it a [config json file](../../sample.data-provider.config.json) which can be deserialized into a [DataProviderConfig](../lib/data_provider/types/model.go) object. 

The `sources` tag is a list of configurations for different feeds, where each feed has a unique `id` and a `config` which can be deserialized into the appropriate [source config](../lib/data_provider/sources/random/config.go). 

## Running Local Code
You can test the Data Provider locally by running:
```
go run apps/cmd/data_provider/main.go start -c ./sample.data-provider.config.json --verbose
```
You will most likely want to replace the `./sample.data-provider.config.json` with a more useful config json. Also make sure any required environment variables like API keys are set in your local environment.

Running in `--verbose` mode with no output address set will just log every price update. If you want to actually send updates somewhere (like the websocket server of your local Publisher Agent), you can pass an output address flag:
```
go run apps/cmd/data_provider/main.go start -c ./sample.data-provider.config.json -o ws://localhost:5216/
```

## Running Published Docker Image
If all the data sources you want to use are already merged into Stork's repo, you can just pull the latest published Data Provider docker image and supply your own config: 
```
docker run --platform linux/arm64 --pull always --restart always --name data-provider -v ./sample.data-provider.config.json:/etc/config.json -d --log-opt max-size=1g storknetwork/data-provider:v1.0.4 start -c /etc/config.json -o ws://localhost:5216/
```



