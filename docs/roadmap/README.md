# Gatekeeper Roadmap

The Gatekeeper project has several goals that help control the roadmap. This is a good document to understand how the project will evolve over time.

## Goals
1. **Runtime Policy Enforcement**: All policies must be deployable to the cluster, and enforcable as validating or mutatating Admission Controllers. This is useful to enforce actions taken by `kubectl` and Kubernetes Operators that are using the Kubernetes API directly to create and manipulate resources.
2. **Offline Policy Enforcement**: Most policies should be able to run out of the cluster, without access to the cluster. This is useful to enforce policies in a CI process, or a GitOps deployment workflow. Offline policy enforcement must not require access to the cluster runtime to make policy decisions.
3. **Introspection**: All policies, whether enforced at runtime or offline, must support introspection to provide compliance and auditing capabilities. This includes determining if any specific policy is preventing, allowing or changing resources, either offline or at runtime. These metrics should be available on a CLI or aggregated to a monitoring and alerting daemon.

## Implementation
The current architecture of Gatekeeper has a few specific implementation decisions:

1. Open Policy Agent (`.rego`) language will provide policy definition for all (runtime and offline) policies.
2. The Open Policy Agent runtime will provide runtime enforcement.
3. Offline enforcement should be implemented by vendoring the `opa` project into Gatekeeper to provide exactly the same policy decisions.

