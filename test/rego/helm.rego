package gateekeeper.library.kubernetes.admission.helm

default admit = false

admit {
  not helm
}

helm {
  input.request.kind.kind = "Pod"
  input.request.operation = "CREATE"
  input.request.object.spec.containers[i].name = "tiller"
}
