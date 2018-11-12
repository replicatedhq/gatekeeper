package gatekeeper.library.kubernetes.admission.privilegedpods

deny[msg] {
	input.request.kind.kind = "Pod"
	input.request.operation = "CREATE"
	input.request.object.spec.containers[i].securityContext.privileged = true
	msg = "privileged pods are not permitted"
}
