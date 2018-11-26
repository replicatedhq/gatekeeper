# GateKeeper CLI

The GateKeeper CLI is designed to run on a workstation that has access to the cluster. The CLI can query stats and metrics from the GateKeeper instance running. The CLI is not designed to submit new policies; those will remain `kubectl apply` commands.

## Authenticating

The GateKeeper CLI will use the current kubectl context. If your kubectl is pointing to the correct cluster, GateKeeper CLI will work.

## Listing Policies

To list all policies installed in GateKeeper, run:

```shell
gatekeeper policies
```

The output will contain a list of all policies and some basic metrics:

```shell
$ gatekeeper policies

NAME
```
