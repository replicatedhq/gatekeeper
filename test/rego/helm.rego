package gatekeeper.library.kubernetes.admission.helm

deny[msg] {
	input.request.kind.kind = "Pod"
	input.request.operation = "CREATE"
	input.request.object.spec.containers[i].name = "tiller"
	msg = "helm and tiller are not permitted"
}
