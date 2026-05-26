package signer

import "github.com/Stork-Oracle/stork-external/shared/signer/evm"

type (
	// EVM types are aliased from the evm subpackage so consumers can keep
	// importing them from the top-level signer package without pulling in
	// the Rust signer_ffi CGo dependency.
	EvmPublisherKey = evm.PublisherKey
	EvmPrivateKey   = evm.PrivateKey

	StarkPublisherKey string
	StarkPrivateKey   string
)
