package gateekeeper.library.kubernetes.admission.defaultns_test

import data.gateekeeper.library.kubernetes.admission.defaultns
import data.gateekeeper.library.kubernetes.inputs.create

test_admit {
	r = create.tiller
	defaultns.admit = true with input as r
}

test_nonadmit {
	r = create.redis
	defaultns.admit = false with input as r
}

