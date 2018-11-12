# Open Policy Agent Controller

The GateKeeper operator can install and operate an installation of Open Policy Agent in the cluster by applying the following YAML to the cluster:

```yaml
apiVersion: controllers.replicated.com/v1alpha1
kind: OpenPolicyAgent
metadata:
  labels:
    controller-tools.k8s.io: "1.0"
  name: my-opa-server
spec:
  name: my-opa-server
  enabledFailureModes:
    ignore: true
    fail: false
```

Explaining each field here:

`spec.name`: The name of this installation of OPA. GateKeeper supports multiple installations, but at the time, is limited to a single installation of OPA per namespace.

`spec.enabledFailureModes`: Each mode that's enabled will deploy a separate instance of Open Policy Agent in the namespace. This allows policies to target different failure modes, without having to decide on a single failure mode for all policies.

`spec.enabledFailureModes.ignore`: Enabled by default, this will deploy an installation of OPA and deploy a ValidatingWebhookConfiguration with `failureMode` set to `Ignore`. In this mode, the policies will be applied, but if the OPA server is unavailable or returns an error, Kubernetes will default to permissive mode and allow the request to be scheduled.

`spec.enabledFailureModes.fail`: Disabled by default. When enabled, this will deploy an installation of OPA and another ValidatingWebhookConfiguration with `failureMode` set to `Fail`. In this mode, if the OPA server is unavailable or returns an error, Kubernetes will not allow the request to be scheduled. This is disabled by default because all requests will be routed through this webhook and a misconfigured server here will prevent anything from executing in the cluster.

## Resources

The Open Policy Agent Controller generates several resources for each failure mode:

- [TLS Secret](https://github.com/replicatedhq/gatekeeper/tree/master/docs/opa-controller/opa-tls.md): A Secret to contain TLS keys and certs.
- [Service](https://github.com/replicatedhq/gatekeeper/tree/master/docs/opa-controller/opa-service.md): A Service to expose the OPA service on a Cluster IP.
- [Deployment](https://github.com/replicatedhq/gatekeeper/tree/master/docs/opa-controller/opa-deployment.md): A Deployment to run the Open Policy Agent pod.
- [ValidatingWebhookConfiguration](https://github.com/replicatedhq/gatekeeper/tree/master/docs/opa-controller/validatingwebhookconfiguration.md): A ValidatingWebhookConfiguration to enable to Admission Controller in Kubernetes.
