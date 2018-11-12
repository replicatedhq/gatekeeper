package gateekeeper.library.kubernetes.admission.dockersock_test

import data.gateekeeper.library.kubernetes.admission.dockersock
import data.gateekeeper.library.kubernetes.inputs.create

test_admit {
	r = create.redis
	dockersock.admit = true with input as r
}

test_nonadmit {
	r = create.dockersock
	dockersock.admit = false with input as r
}

