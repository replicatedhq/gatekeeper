package gateekeeper.library.kubernetes.admission.defaultns

default admit = false

admit {
  not defaultns
}

defaultns {
  input.request.kind.kind = "Pod"
  input.request.operation = "CREATE"
  input.request.object.metadata.namespace = "default"
}
