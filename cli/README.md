# stork-cli

### Development

```
abigen --abi ../contracts/evm/stork.abi --pkg main --type StorkContract --out stork_contract.go
```

### Run locally

```
go run . evm-push -w ws://localhost:5211 -a fake -c ws://127.0.0.1:8545 -x 0xe7f1725e7734ce288f8367e1bb143e90bb3f0512 -f asset-config.yaml -m ./private-key.secret -v
```


