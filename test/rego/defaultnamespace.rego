package gatekeeper.library.kubernetes.admission.defaultnamespace

deny[msg] {
	input.request.kind.kind = "Pod"
	input.request.operation = "CREATE"
	input.request.object.metadata.namespace = "default"
	msg = "pods are not permitted to run in the default namespace"
}
