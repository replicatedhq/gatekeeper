# Open Policy Agent TLS

For each failureMode that's enabled, GateKeeper will create a secret in the namespace to contain generated TLS configuration. The secret will be named `gatekeeper-<FAILURE MODE>` (e.g. `gatekeeper-ignore` or `gatekeeper-fail`). GateKeeper uses the cfssl library to generate and manage TLS for each deployment. This secret contains 4 keys:

`ca.key`: The private key for the generate certificate authority.
`ca.crt`: The CA bundle that will be used to sign carts and can be installed in the client to trust the server.

`tls.key`: The private key for the web server (OPA server).
`tls.crt`: The TLS certificate, signed by the generated CA, that will be used to encrypt http requests in cluster.

