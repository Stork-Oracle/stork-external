# Running on ec2

The evm-pusher runs on a per chain basis. Right now all pushers are on the same machine. It is running as a systemd service. Logs are being picked up by amazon-cloudwatch-agent.

## Setup Machine

### Cloudwatch Agent

1. Install Cloudwatch Agent 
```
sudo yum install amazon-cloudwatch-agent -y
```
2. Set up `/opt/aws/amazon-cloudwatch-agent/bin/config.json`
```json
{
        "agent": {
                "run_as_user": "cwagent"
        },
        "logs": {
                "logs_collected": {
                        "files": {
                                "collect_list": [
                                        {
                                                "file_path": "/var/log/berachain-testnet.log",
                                                "log_group_class": "STANDARD",
                                                "log_group_name": "/aws/ec2/dev-apps-evm-pusher",
                                                "log_stream_name": "berachain-testnet-[{instance_id}]",
                                                "retention_in_days": -1
                                        }
                                ]
                        }
                }
        }
}
```

3. Run cloudwatch agent
```bash
sudo /opt/aws/amazon-cloudwatch-agent/bin/amazon-cloudwatch-agent-ctl \
-a fetch-config -m ec2 -c file:/opt/aws/amazon-cloudwatch-agent/bin/config.json -s
```

4. Enable agent to start on boot
```
sudo systemctl enable amazon-cloudwatch-agent
```

## Deploy

1. Build the cli
```bash
make
```

2. Copy executable to machine
```
scp build/stork-linux-arm64 stork-dev-evm-pusher:
```

3. SSH to machine
```
ssh stork-dev-evm-pusher
```

4. Place CLI in /usr/local/bin
```
sudo cp stork-linux-arm64 /usr/local/bin/stork
```

5. Verify stork installation
```
stork help
```

6. Setup `.asset-config.yaml` and `.secret` files in user home directory.

7. Create a `systemd` definition file. e.g. in `/etc/systemd/system/berachain-testnet.service`

```
[Unit]
Description=Stork Berachain Pusher Service
After=network.target

[Service]
ExecStart=stork evm-push -w wss://api.dev.jp.stork-oracle.network -a fake -c https://bartio.rpc.berachain.com -x 0xacC0a0cF13571d30B4b8637996F5D6D774d4fd62 -f /home/ec2-user/berachain.asset-config.yaml -m /home/ec2-user/berachain-testnet.secret -b 60 -v
Restart=always
StandardOutput=append:/var/log/berachain-testnet.log
StandardError=append:/var/log/berachain-testnet.log
User=ec2-user
Group=ec2-user

[Install]
WantedBy=multi-user.target
```

8. Restart systemctl

```
sudo systemctl daemon-reload
sudo systemctl enable berachain-testnet.service
sudo systemctl restart berachain-testnet.service
```

9. (If necessary) restart amazon-cloudwatch-agent

```
sudo systemctl daemon-reload
sudo systemctl enable amazon-cloudwatch-agent
sudo systemctl restart amazon-cloudwatch-agent
```
