# Open Policy Agent Service

For each failureMode that's enabled, GateKeeper will create a service in the namespace to expose the webhook to the cluster. This service will be named `gatekeeper-opa-<FAILURE MODE>` (e.g. `gatekeeper-opa-ignore` or `gatekeeper-opa-fail`).

The service spec will look similar to this:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: gatekeeper-opa-ignore
  namespace: default
spec:
  type: ClusterIP
  selector:
    controller: openPolicyAgent
    failurePolicy: Ignore
    gatekeeper: gatekeeper-name
  ports:
  - name: https
    port: 443
    protocol: TCP
    targetPort: 443
```
