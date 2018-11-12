package gatekeeper.library.kubernetes.admission.latest_test

import data.gatekeeper.library.kubernetes.admission.latest
import data.gatekeeper.library.kubernetes.inputs.create

test_admit {
	r = create.redis
	latest.admit = true with input as r
}

test_nonadmit {
	r = create.redislatest
	latest.admit = false with input as r
}

