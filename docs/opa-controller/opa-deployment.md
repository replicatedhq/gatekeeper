# Open Policy Agent Deployment

For each failureMode that's enabled, GateKeeper will create a deployment in the namespace to run the Open Policy Agent service (REST API). The deployment will be named `<CONTROLLER NAME>-<FAILURE MODE>` (e.g. `admission-ignore` when deployed with name: `admission` and failurePolicy `Ignore`).

The deployment spec will look similar to this:

```yaml
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: admission-ignore
  namespace: default
  labels:
    app: admission-ignore
    controller: openPolicyAgent
    failurePolicy: Ignore
    gatekeeper: admission
spec:
  replicas: 1
  selector:
    matchLabels:
      app: admission-ignore
      controller: openPolicyAgent
      failurePolicy: Ignore
      gatekeeper: admission
  template:
    metadata:
      name: admission-ignore
      labels:
        app: admission-ignore
        controller: openPolicyAgent
        failurePolicy: Ignore
        gatekeeper: admission
    spec:
      containers:
        - name: opa
          args:
            - run
            - --server
            - --tls-cert-file=/certs/tls.crt
            - --tls-private-key-file=/certs/tls.key
            - --addr=0.0.0.0:443
            - --addr=http://127.0.0.1:8181
          image: openpolicyagent/opa:0.10.1
          imagePullPolicy: IfNotPresent
          ports:
            - containerPort: 443
              name: https
              protocol: TCP
          volumeMounts:
            - mountPath: /certs
              name: tls
              readOnly: true
      volumes:
        - name: tls
          secret:
            defaultMode: 420
            secretName: gatekeeper-ignore
```
