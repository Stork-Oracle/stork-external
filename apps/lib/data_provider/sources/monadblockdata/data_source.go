// Code initially generated by gen.go.
// This file contains the implementation for pulling data from the data source and putting it on the updatesCh.

package monadblockdata

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/types"
	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/utils"
	"github.com/rs/zerolog"
)

type monadBlockDataDataSource struct {
	monadBlockDataConfig MonadBlockDataConfig
	valueId 	types.ValueId
	logger 		zerolog.Logger
	// TODO: set any necessary parameters
}

func newMonadBlockDataDataSource(sourceConfig types.DataProviderSourceConfig) *monadBlockDataDataSource {
	monadBlockDataConfig, err := GetSourceSpecificConfig(sourceConfig)
	if err != nil {
		panic("unable to decode config: " + err.Error())
	}

	// TODO: add any necessary initialization code
	return &monadBlockDataDataSource{
		monadBlockDataConfig: monadBlockDataConfig,
		valueId: 	sourceConfig.Id,
		logger: 	utils.DataSourceLogger(MonadBlockDataDataSourceId),
	}
}

func (r monadBlockDataDataSource) RunDataSource(ctx context.Context, updatesCh chan types.DataSourceUpdateMap) {
	// TODO: Write all logic to fetch data points and report them to updatesCh
	panic("implement me")
}

// Block structure for eth_getBlockByNumber response
type Block struct {
	Number       string `json:"number"`
	Timestamp    string `json:"timestamp"`
	Transactions string `json:"transactions,string"` // Force string type
	GasUsed      string `json:"gasUsed"`
	GasLimit     string `json:"gasLimit"`
	BaseFeePerGas string `json:"baseFeePerGas"`
}

// Transaction structure for eth_getBlockByNumber response
type Transaction struct {
	Hash             string `json:"hash"`
	TransactionIndex string `json:"transactionIndex"`
	From             string `json:"from"`
	To               string `json:"to"`
	Value            string `json:"value"`
	Gas              string `json:"gas"`
	GasPrice         string `json:"gasPrice"`
	Input            string `json:"input"`
}

// getBlockByNumber retrieves a block by number from the blockchain
func (m monadBlockDataDataSource) getBlockByNumber(blockNumber string) (*Block, error) {
	// Create a custom request to get the block without transactions
	request := jsonRPCRequest{
		JSONRPC: "2.0",
		Method:  "eth_getBlockByNumber",
		Params:  []interface{}{blockNumber, false}, // false = don't get transaction objects
		ID:      1,
	}

	// Send request to Monad RPC endpoint
	response, err := m.sendJSONRPCRequest(request)
	if err != nil {
		return nil, err
	}

	// First convert to a map to handle the transactions field
	var blockMap map[string]interface{}
	if err := json.Unmarshal(response.Result, &blockMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal block to map: %w", err)
	}

	// Convert transactions to a string representation
	txBytes, err := json.Marshal(blockMap["transactions"])
	if err != nil {
		return nil, fmt.Errorf("failed to marshal transactions: %w", err)
	}

	// Set up the Block struct
	block := &Block{
		Number:        fmt.Sprintf("%v", blockMap["number"]),
		Timestamp:     fmt.Sprintf("%v", blockMap["timestamp"]),
		Transactions:  string(txBytes),
		GasUsed:       fmt.Sprintf("%v", blockMap["gasUsed"]),
		GasLimit:      fmt.Sprintf("%v", blockMap["gasLimit"]),
		BaseFeePerGas: fmt.Sprintf("%v", blockMap["baseFeePerGas"]),
	}

	return block, nil
}
