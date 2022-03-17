# go-perun - Hyperledger Fabric Backend

Chaincode and [go-perun](https://github.com/hyperledger-labs/go-perun/) backend implementation for [Hyperledger Fabric](https://github.com/hyperledger/fabric).

## Testing

The Go unit tests can be run as usual with `go test ./...`. The Github CI also
runs those tests.

The end-2-end tests can be run using the script `scripts/end2end.sh`. They
require the [`fabric-samples`](https://github.com/hyperledger/fabric-samples)
repository to be checked out locally because the scripts in `test-network` are
used to setup a Fabric network. By default, the `end2end.sh` script assumes the
`fabric-samples` repository to be checked out in the same directory as the
`perun-fabric` repository. If it is somewhere else, the script must be told
about its location via the environment variable `FABRIC_SAMPLES_DIR` (whose
default is `../fabric-samples`). [Docker](https://www.docker.com/) also needs to
be installed and the current user must have the correct system privileges to
control it.

When all prerequisites are met, the end-2-end tests can be simply run like
```sh
scripts/end2end.sh adjudicator
```
where the first and only argument specifies the chaincode to deploy and test,
currently `assetholder` or `adjudicator`, or `down` to shut down the network.

End-2-end tests can run multiple times without restarting the network every
time. This is realized by deploying the chaincode to a new random instance name
every time. When you're done running end-2-end tests, shut down the network with
argument `down`.

## Copyright

Copyright 2022 - See [NOTICE file](NOTICE) for copyright holders.
Use of the source code is governed by the Apache 2.0 license that can be found in the [LICENSE file](LICENSE).

Contact us at [info@perun.network](mailto:info@perun.network).
