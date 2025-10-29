# HIP-3 Pusher

The HIP-3 Pusher is a plug and play tool to facilitate using Stork's data with [Hyperliquid's HIP-3 builder deployed perpetuals](https://hyperliquid.gitbook.io/hyperliquid-docs/hyperliquid-improvement-proposals-hips/hip-3-builder-deployed-perpetuals) product.

## Getting Started

1. Install dependencies
```bash
pip install -r requirements.txt
```

2. Copy the Hip3 Pusher example config
```bash
cp examples/test_config.yaml config.yaml
```

3. Edit the config.yaml file to your liking

4. Create a private key file for the Hip3 Pusher
```bash
echo "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef" > private_key.secret
```

4. Run the Hip3 Pusher
```bash
python -m hip3_pusher push config.yaml -a <stork-ws-auth> -k private_key.secret -v
```
