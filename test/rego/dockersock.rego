package gatekeeper.library.kubernetes.admission.dockersock

deny[msg] {
	input.request.kind.kind = "Pod"
	input.request.operation = "CREATE"
	input.request.object.spec.volumes[i].hostPath.path = "/var/run/docker.sock"
	msg = "pods may not mount the docker socket"
}
