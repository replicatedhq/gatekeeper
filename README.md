# GateKeeper

[![CircleCI](https://circleci.com/gh/replicatedhq/gatekeeper/tree/master.svg?style=svg)](https://circleci.com/gh/replicatedhq/gatekeeper/tree/master)

Gatekeeper is currently an implementation of a Kubernetes Operator for installing, configuring and managing [Open Policy Agent](https://www.openpolicyagent.org/) to provide dynamic admission controllers in a cluster.

## Getting Started

The recommended way to configure Gatekeeper is to use [Replicated Ship](https://github.com/replicatedhq/ship):

```shell
brew install ship
ship init https://github.com/replicatedhq/gatekeeper/tree/master/docs/gatekeeper-k8s
```

Ship will download and give you an opportunity to review the Kubernetes manifests included to run Gatekeeper. You can create patches and overlays to make any changes necessary for your environment. Once finished, follow the instructions in Ship and `kubectl apply -f rendered.yaml`.

You can then use `ship watch && ship update` to watch and configure updates as they are shipped here.

For more information on the components, and other methods to install Gatekeeper, [read the docs](https://github.com/replicatedhq/gatekeeper/tree/master/docs/).

## Deploying Policies

After installing Gatekeeper to a cluster, a policy can be deployed using `kubectl apply -f ./config/samples/policies_v1alpha2_admissionpolicy.yaml`. (This is a sample policy that prevents any pod from using images tagged `:latest`). When the policy is applied, if OPA is running in the same namespace, the controller will deploy the policy from the YAML to the OPA instance. If OPA is not found, the controller will provision a new OPA instance, and deploy the policy to that new instance, whne it's ready.

Gatekeeper provisions Open Policy Agent with all of the necessary TLS configuration, webhook configuration, and underlying Kubernetes resources that are required to create a dynamic admission controller.

## Motivations

The Open Policy Agent (OPA) project is an ambitious project that does much more than just Kubernetes Admission Controllers.

#### Simplify the task of installing and configuring OPA in Kubernetes.
Installing OPA into a Kubernetes cluster is more complex than many applications. The recommended installation includes creating a new certificate authority (CA) and then creating a cert, signed by that CA. This TLS configuration should be deployed and referenced in the openpolicyagent deployment and also manually copied into the webhook configuration. Managing this through automation can be difficult and prone to errors. The Gatekeeper operator manages this in-cluster, so the keys never have to be transferred to the cluster, and the CA and certs are properly configured every time.

Dynamic admission controllers in Kubernetes are powerful, but can also be difficult to troubleshoot and configure. A goal of the Gatekeeper operator is to make it easier to roll out new admission policies, with as little risk as possible.

#### Provide a custom resource to manage policy files (`.rego`) instead of using ConfigMaps
This allows for easier listing and management of individual policies. Instead of using the existing `ConfigMap` and in-cluster sync, the Gatekeeper operator introduces a new type named `admissionpolicies.policies.replicated.com`. This makes it easy to just `kubectl get admissionpolicies.policies.replicated.com` and view all dynamic admission policies installed in the cluster.

#### Validation of policies before deployment
One future goal of Gatekeeper is to validate new policies and changes to existing policies before deploying. This includes compiling the policy and also backtesting it against previous requests received to ensure that the policy will have the expected effects.

## Architecture and Roadmap
Gatekeeper is still an early project that's evolving. To see our roadmap and better understand the current and planned architecture, [architecture doc](https://github.com/replicatedhq/gatekeeper/tree/master/docs/architecture) and the [roadmap doc](https://github.com/replicatedhq/gatekeeper/tree/master/docs/roadmap) in this repo.

## Contributing

Fork and clone this repo, and you can run it locally on a Kubernetes cluster:

```shell
make install  # this will install the CRDs to your cluster
skaffold dev  # this will start the manager and controllers in your cluster, and watch for file changes and redeploy
```

