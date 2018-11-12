package gateekeeper.library.kubernetes.admission.privilegedpods

default admit = false

admit {
  not privileged_pod
}

privileged_pod {
  input.request.kind.kind = "Pod"
  input.request.operation = "CREATE"
  input.request.object.spec.containers[i].securityContext.privileged = true
}