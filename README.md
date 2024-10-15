# Deploy a “OP stack devnet” in given machine.

https://stack.optimism.io/docs/build/getting-started

# You need to mine Sepholia from faucet

## automate.go 

- Go File to automate the deployment of OP Stack Devnet on the given machine

The specific command "build/bin/geth init --datadir=datadir --state.scheme=hash genesis.json" will differ from the original in the website. Specfically modify "--state.scheme=hash" for it to work with "--gcmode=archive".

## kind_automate.go 

- Go File to automate the deployment of OP Stack in kind cluster on the given machine

This automation uses the op-chain-charts repository, available at https://geo-web-project.github.io/op-chain-charts/.