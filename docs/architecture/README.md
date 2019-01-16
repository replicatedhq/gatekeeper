# Gatekeeper Architecture

![Gatekeeper Runtime](https://github.com/replicatedhq/gatekeeper/blob/master/docs/architecturue/assets/arch-1.png)

- Kubebuilder Manager watches for `AdmissionPolicy` CRs
- On CR creation, manager creates:
    - Gatekeeper Deployment
    - CA and TLS certs for OPA service
    - OPA deployment
    - OPA Service
    - Kubernetes (mutating or validating) Webhook that queries OPA service on resource creation

Policies are POST'd to the OPA service by the manager. The TLS cert is passed into the OPA deployment, and the CA is registered in the Kubernetes Webhook.


### Gatekeeper Deployment

The Gatekeeper deployment is mostly aspirational. It is currently used as the server-side counterpart to the `gatekeeper` CLI. There are also some thoughts on using it to manage auditing and introspection by being the Kubernetes webhook target, and forwarding requests down to OPA, while caching and storing stats about the results.


