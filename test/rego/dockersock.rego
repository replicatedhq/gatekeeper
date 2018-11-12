package gateekeeper.library.kubernetes.admission.dockersock

default admit = false

admit {
  not mounting_dockersock
}

mounting_dockersock {
  input.request.kind.kind = "Pod"
  input.request.operation = "CREATE"
  input.request.object.spec.volumes[i].hostPath.path = "/var/run/docker.sock"
}