package gateekeeper.library.kubernetes.admission.privilegedpods_test

import data.gateekeeper.library.kubernetes.admission.privilegedpods
import data.gateekeeper.library.kubernetes.inputs.create

test_admit {
	r = create.redis
	privilegedpods.admit = true with input as r
}

test_nonadmit {
	r = create.privileged
	privilegedpods.admit = false with input as r
}

