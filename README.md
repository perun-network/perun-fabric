# Perun's Hyperledger Fabric Chaincode and Backend

<p>
  <a href="https://www.apache.org/licenses/LICENSE-2.0.txt"><img src="https://img.shields.io/badge/license-Apache%202-blue" alt="License: Apache 2.0"></a>
  <a href="https://github.com/perun-network/perun-fabric/actions/workflows/ci.yml"><img src="https://github.com/perun-network/perun-fabric/actions/workflows/ci.yml/badge.svg?branch=main" alt="CI status"></a>
</p>

This repository provides the chaincode and [go-perun](https://github.com/hyperledger-labs/go-perun/) backend implementation
for [Hyperledger Fabric](https://github.com/hyperledger/fabric).

## Project structure
* `adjudicator/`: On-chain logic. Memory implementations for off-chain testing.
* `chaincode/`: Chaincode endpoint, ledger and asset implementation.
* `channel/`: Off-chain logic. Channel interface implementations.
    * `binding/` Chaincode bindings.
* `client/`: Helper functions for setting up a *go-perun* client.
    * `test/` End-2-end tests.
* `pkg/`: 3rd-party helpers.
* `scripts/`: Test environment setup.
* `wallet/`: Wallet interface implementations.

## Development Setup
If you want to locally develop with this project:

1. Setup the [fabric-samples](https://github.com/hyperledger/fabric-samples) test environment and add its directory as `FABRIC_SAMPLES_DIR` to your path.
```sh
curl -sSL https://raw.githubusercontent.com/hyperledger/fabric/v2.4.6/scripts/bootstrap.sh | bash -s
FABRIC_SAMPLES_DIR=$(pwd)/fabric-samples/
```

2. Clone this repo.
```sh
git clone https://github.com/perun-network/perun-fabric.git
cd perun-fabric
```

 3. Start the test chain and deploy the chaincode. This step needs a working [Docker](https://www.docker.com) instance running in the background.
```sh
./scripts/deployCC.sh
```

4. Run the tests. This step needs a working [Go distribution](https://golang.org), see [go.mod](go.mod) for the required version. Always ensure that `FABRIC_SAMPLES_DIR` is set.
```sh
go test ./... -p 1
```


Further, you can shut down the test environment.
```sh
./scripts/network.sh down
```

Or start the test environment without chaincode deployment.
```sh
./scripts/network.sh up
```

## Security Disclaimer

This software is still under development.
The authors take no responsibility for any loss of digital assets or other damage caused by the use of it.

## Copyright

Copyright 2022 - See [NOTICE file](NOTICE) for copyright holders. Use of the source code is governed by the Apache 2.0
license that can be found in the [LICENSE file](LICENSE).

Contact us at [info@perun.network](mailto:info@perun.network).
