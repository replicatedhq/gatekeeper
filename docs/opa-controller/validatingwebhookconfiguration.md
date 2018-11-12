# ValidatingWebhookConfiguration

One webhook is deployed per failurePolicy enabled. When for a failurePolicy of `Ignore` and a deployment named `admission` the following webhook would be deployed:

```yaml
apiVersion: admissionregistration.k8s.io/v1beta1
kind: ValidatingWebhookConfiguration
metadata:
  name: admission-ignore
  namespace: ""
webhooks:
  - clientConfig:
      caBundle: <GENERATED CA>
      service:
        name: gatekeeper-opa-ignore
        namespace: default
    failurePolicy: Ignore
    name: validating-webhook.openpolicyagent.org
    namespaceSelector: {}
    rules:
      - apiGroups:
        - '*'
        apiVersions:
        - '*'
        operations:
        - CREATE
        - UPDATE
        resources:
        - '*'
```
