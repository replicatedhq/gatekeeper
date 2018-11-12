package gateekeeper.library.kubernetes.inputs.create

privileged = {
    "kind": "AdmissionReview",
    "apiVersion": "admission.k8s.io/v1beta1",
    "request": {
        "uid": "630656ce-dfc5-11e8-85bc-025000000001",
        "kind": {
            "group": "",
            "version": "v1",
            "kind": "Pod"
        },
        "resource": {
            "group": "",
            "version": "v1",
            "resource": "pods"
        },
        "namespace": "default",
        "operation": "CREATE",
        "userInfo": {
            "username": "docker-for-desktop",
            "groups": [
                "system:masters",
                "system:authenticated"
            ]
        },
        "object": {
            "metadata": {
                "name": "privileged",
                "namespace": "default",
                "uid": "6306355b-dfc5-11e8-85bc-025000000001",
                "creationTimestamp": "2018-11-04T00:05:49Z",
                "annotations": {
                    "kubectl.kubernetes.io/last-applied-configuration": "{\"apiVersion\":\"v1\",\"kind\":\"Pod\",\"metadata\":{\"annotations\":{},\"name\":\"privileged\",\"namespace\":\"default\"},\"spec\":{\"containers\":[{\"image\":\"k8s.gcr.io/pause\",\"name\":\"pause\",\"securityContext\":{\"privileged\":true}}]}}\n"
                }
            },
            "spec": {
                "volumes": [
                    {
                        "name": "default-token-swbl6",
                        "secret": {
                            "secretName": "default-token-swbl6"
                        }
                    }
                ],
                "containers": [
                    {
                        "name": "pause",
                        "image": "k8s.gcr.io/pause",
                        "resources": {},
                        "volumeMounts": [
                            {
                                "name": "default-token-swbl6",
                                "readOnly": true,
                                "mountPath": "/var/run/secrets/kubernetes.io/serviceaccount"
                            }
                        ],
                        "terminationMessagePath": "/dev/termination-log",
                        "terminationMessagePolicy": "File",
                        "imagePullPolicy": "Always",
                        "securityContext": {
                            "privileged": true
                        }
                    }
                ],
                "restartPolicy": "Always",
                "terminationGracePeriodSeconds": 30,
                "dnsPolicy": "ClusterFirst",
                "serviceAccountName": "default",
                "serviceAccount": "default",
                "securityContext": {},
                "schedulerName": "default-scheduler",
                "tolerations": [
                    {
                        "key": "node.kubernetes.io/not-ready",
                        "operator": "Exists",
                        "effect": "NoExecute",
                        "tolerationSeconds": 300
                    },
                    {
                        "key": "node.kubernetes.io/unreachable",
                        "operator": "Exists",
                        "effect": "NoExecute",
                        "tolerationSeconds": 300
                    }
                ]
            },
            "status": {
                "phase": "Pending",
                "qosClass": "BestEffort"
            }
        },
        "oldObject": null
    }
}

