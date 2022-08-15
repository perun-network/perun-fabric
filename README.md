# go-perun - Hyperledger Fabric Chaincode and Backend

Chaincode and [go-perun](https://github.com/hyperledger-labs/go-perun/) backend implementation
for [Hyperledger Fabric](https://github.com/hyperledger/fabric).

## Packages

### go-perun Backend

Packages `wallet` and `channel` implement the `go-perun` Fabric backend. To use the backend in Go, do

```go
import _ "github.com/perun-network/perun-fabric"
```

This sets the `wallet` and `channel` backends and randomizers in go-perun via
`init()` functions.

`Address`es and `Account`s are realized as `ecdsa.PublicKey`s and `PrivateKey`s, because these are the only keys
actually used in Fabric.

### `adjudicator`

Package `adjudicator` implements a backend-independent Adjudicator smart contract. It is a single-asset, two party
payment channel Adjudicator. It allows for an injectable ledger, on which state registrations and participant holdings
are stored. There is a `MemLedger` implementation that stores values in memory, which is used in tests.
Package `chaincode` includes a `StubLedger`, which uses a Fabric blockchain as storage.

An `Adjudicator` contains an `AssetHolder`, which defines the logic for funds management and works on a `HoldingLedger`
interface. Complex asset holding scenarios can be implemented through this `HoldingLedger` abstraction.

### `chaincode`

Package `chaincode` implements concrete Fabric Chaincode instances of the abstract `Adjudicator` and `AssetHolder` smart
contracts. In this sense, they are mere shims - all actual logic is implemented in the abstract smart contracts.

The `Deposit` and `Withdraw` functions use the transaction sender from the transaction context as the affected party,
realizing proper access control.

The `StubLedger` is an `adjudicator.Ledger` implementation that operates on a Fabric network.

### `tests`

Package `tests` contains end-2-end tests of the `Adjudicator` and `AssetHolder`
chaincodes. The files `tests/*/contract.go` show how to call/bind to chaincode. The tests itself then show how
client-side code can be written. File `setup.go` contains client setup code with respect to the Fabric
`test-network` from the `fabric-samples` repository, see below.

### `client`

Package `client` contains convenience functions for setting up
[`fabric-gateway`](https://github.com/hyperledger/fabric-gateway) client connections, in particular reading of PEM
certificates and private keys from files. It is used by the `tests` package.

## Testing

The Go unit tests can be run as usual with `go test ./...`. The Github CI also runs those tests.

The end-2-end tests can be run using the script `scripts/end2end.sh`. They require
the [`fabric-samples`](https://github.com/hyperledger/fabric-samples)
repository to be checked out locally because the scripts in `test-network` are used to setup a Fabric network. By
default, the `end2end.sh` script assumes the
`fabric-samples` repository to be checked out in the same directory as the
`perun-fabric` repository. If it is somewhere else, the script must be told about its location via the environment
variable `FABRIC_SAMPLES_DIR` (whose default is `../fabric-samples`). [Docker](https://www.docker.com/) also needs to be
installed and the current user must have the correct system privileges to control it.

When all prerequisites are met, the end-2-end tests can be simply run like

```sh
scripts/end2end.sh adjudicator
```

where the first and only argument specifies the chaincode to deploy and test, currently `assetholder` or `adjudicator`,
or `down` to shut down the network.

End-2-end tests can run multiple times without restarting the network every time. This is realized by deploying the
chaincode to a new random instance name every time. When you're done running end-2-end tests, shut down the network with
argument `down`.

## Copyright

Copyright 2022 - See [NOTICE file](NOTICE) for copyright holders. Use of the source code is governed by the Apache 2.0
license that can be found in the [LICENSE file](LICENSE).

Contact us at [info@perun.network](mailto:info@perun.network).
