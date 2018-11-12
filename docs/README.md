# GateKeeper Documentation

GateKeeper is built using Kubebuilder as a Kubernetes Operator. This means that you can configure it, and all interactions with it, are achieved using Kubernetes manifests that are applied to the cluster.

The docs are divided into several sections:

## Policy Controller
The [Policy Controller](https://github.com/replicatedhq/gatekeeper/tree/master/docs/policy-controller/) docs are a reference to using the GateKeeper operator to manage admission controller policies. This part of the documentation explains how to create a new policy, apply it, manage it, and show all options available. Once you have GateKeeper isntalled, this is the reference docs on implementating policies.

## Open Policy Agent Controller
The [Open Policy Agent Controller](https://github.com//replicatedhq/gatekeeper/tree/master/docs/opa-controller/) docs are a reference to the GateKeeper implementation of the Open Policy Agent server. Because GateKeeper relies on OPA to run, GateKeeper installs and configured OPA automatically. The docs here explain how to confiigure that installation and how to troubleshoot it.
