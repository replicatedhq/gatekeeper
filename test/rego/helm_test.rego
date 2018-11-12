package gateekeeper.library.kubernetes.admission.helm_test

import data.gateekeeper.library.kubernetes.admission.helm
import data.gateekeeper.library.kubernetes.inputs.create

test_admit {
	r = create.redis
	helm.admit = true with input as r
}

test_nonadmit {
	r = create.tiller
	helm.admit = false with input as r
}

