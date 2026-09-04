package bindings

import (
	"crypto/tls"
	"fmt"
	"strings"

	rpcv2 "github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/sui/rpc/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

const defaultGrpcPort = "443"

// Client wraps the Sui fullnode gRPC v2 service clients used by StorkContract.
type Client struct {
	conn      *grpc.ClientConn
	Ledger    rpcv2.LedgerServiceClient
	State     rpcv2.StateServiceClient
	Execution rpcv2.TransactionExecutionServiceClient
}

// DialGrpc connects to a Sui fullnode gRPC endpoint. Accepts "host", "host:port",
// "https://host[:port]" (TLS), or "http://host[:port]" (plaintext, for local nodes).
func DialGrpc(target string) (*Client, error) {
	addr, plaintext := normalizeGrpcTarget(target)

	transportCredentials := credentials.NewTLS(&tls.Config{MinVersion: tls.VersionTLS12})
	if plaintext {
		transportCredentials = insecure.NewCredentials()
	}

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(transportCredentials))
	if err != nil {
		return nil, fmt.Errorf("failed to create Sui gRPC client: %w", err)
	}

	return &Client{
		conn:      conn,
		Ledger:    rpcv2.NewLedgerServiceClient(conn),
		State:     rpcv2.NewStateServiceClient(conn),
		Execution: rpcv2.NewTransactionExecutionServiceClient(conn),
	}, nil
}

// Close tears down the underlying gRPC connection.
func (c *Client) Close() error {
	err := c.conn.Close()
	if err != nil {
		return fmt.Errorf("failed to close Sui gRPC connection: %w", err)
	}

	return nil
}

// normalizeGrpcTarget converts URL-style endpoints into gRPC dial targets,
// defaulting to TLS on port 443 unless an explicit http:// scheme is given.
func normalizeGrpcTarget(target string) (string, bool) {
	plaintext := false

	switch {
	case strings.HasPrefix(target, "http://"):
		target = strings.TrimPrefix(target, "http://")
		plaintext = true
	case strings.HasPrefix(target, "https://"):
		target = strings.TrimPrefix(target, "https://")
	}

	target = strings.TrimSuffix(strings.TrimSpace(target), "/")

	if !strings.Contains(target, ":") {
		target = target + ":" + defaultGrpcPort
	}

	return target, plaintext
}
