# XLN
XLN is a service that runs on top of LND, providing an API that supports
multiple users, each with multiple wallets. Each user is given their own
API key, allowing the node operator to provide service to multiple,
independent parties.

XLN can be configured to serve both a gRPC and a REST API.
See [xln.proto](xlnrpc/xln.proto) for the gRPC service definition and
[xln.yaml](xlnrpc/xln.yaml) for the REST endpoints.

## Build
```bash
make build
```

## Run
```bash
./xln
```