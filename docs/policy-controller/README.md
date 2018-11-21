# Policy Controller

GateKeeper installs a Policy controller to manage AdmissionPolicy documents in the cluster. When a Kubernetes manifest that matches the Kind: AdmissionPolicy is deployed, the controller will handle this and enable it in the appropriate Open Policy Agent deployment.

An example AdmissionPolicy is:

```yaml
apiVersion: policies.replicated.com/v1alpha2
kind: AdmissionPolicy
metadata:
  labels:
    controller-tools.k8s.io: "1.0"
  name: no-latest-tags
spec:
  name: latest
  policy: |
    package kubernetes.admission

    deny[msg] {
        input.request.kind.kind = "Pod"
        input.request.operation = "CREATE"
        endswith(input.request.object.spec.containers[_].image, ":latest")
        msg = "pod contains image using latest tag"
    }
```

The above policy will be deployed to the "Ignore" Open Policy Agent deployment.
