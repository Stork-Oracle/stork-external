// Package evm provides pure-Go EVM signing and verification for Stork publisher
// prices and auth headers. It has no CGo dependencies, so consumers that only
// need EVM can import it without linking the Rust signer_ffi library.
package evm

type (
	PublisherKey string
	PrivateKey   string
)
