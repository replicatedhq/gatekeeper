package gatekeeper.library.kubernetes.admission.latest

deny[msg] {
	input.request.kind.kind = "Pod"
	input.request.operation = "CREATE"
	endswith(input.request.object.spec.containers[_].image, ":latest")
	msg = "pod contains image using latest tag"
}
