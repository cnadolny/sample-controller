# build image
```
docker build . -t <repo>/sample
docker push <repo>/sample
```

update image in artifacts/deploy/deploy.yaml

# run
```
kubectl create clusterrolebinding default-cluster-admin --clusterrole=cluster-admin --serviceaccount=default:default
kubectl create -f artifacts/definition
kubectl create -f artifacts/deploy
```

# logs

This CRD will attempt to demonstrate the issue of listing deployed CRDs with schema changes in multiple versions. 

Create the V1 CRD:
```
kubectl create -f artifacts/crd/example-v1.yaml
```

Run `kubectl logs <pod name>`, which will output a few fields from the JSON that is stored within Kubernetes, as well as results from listing the objects from the Informer.
```
Printing V1 JSON
-----------------
samplecontroller.k8s.io/v1alpha1
example-foo
map[convertSpec:0 deploymentName:example-foo replicas:1]
-----------------
Printing V2 JSON
-----------------
samplecontroller.k8s.io/v2
example-foo
map[convertSpec:0 deploymentName:example-foo replicas:1]
-----------------
I0114 00:39:42.362080       1 controller.go:139]
 Listing foo V1 objectexample-foo{example-foo 0xc000285d30 0}
```

Next, we'll apply the second YAML, which has the V2 version of the CRD. The V2 version changes the field convertSpec from an int to a string.

`kubectl create -f artifacts/crd/example-v2.yaml`
`kubectl logs <pod-name>`
```
W0114 00:41:37.853093       1 reflector.go:340] pkg/mod/k8s.io/client-go@v0.0.0-20200111153838-ea0a6e11838c/tools/cache/
reflector.go:108: watch of *v2.Foo ended with: an error on the server ("unable to decode an event from the watch stream: unable to 
decode watch event: v1alpha1.Foo.Spec: v1alpha1.FooSpec.ConvertSpec: readUint64: unexpected character: \xff, error found in #10 
byte of ...|ertSpec\":\"int value\"|..., bigger context ...|6-11ea-b081-e291c83199d7\"},\"spec\":{\"convertSpec\":\"int value\",
\"deploymentName\":\"example-foo-v2\",\"repl|...") has prevented the request from succeeding

Printing V1 JSON
-----------------
samplecontroller.k8s.io/v1alpha1
example-foo
map[convertSpec:0 deploymentName:example-foo replicas:1]
samplecontroller.k8s.io/v1alpha1
example-foo-v2
map[convertSpec:int value deploymentName:example-foo-v2 replicas:1]
-----------------
Printing V2 JSON
-----------------
samplecontroller.k8s.io/v2
example-foo
map[convertSpec:0 deploymentName:example-foo replicas:1]
samplecontroller.k8s.io/v2
example-foo-v2
map[convertSpec:int value deploymentName:example-foo-v2 replicas:1]
-----------------
I0114 00:41:37.858927       1 controller.go:139]
 Listing foo V1 objectexample-foo{example-foo 0xc0000a9340 0}
```

Here we can see that both objects are stored in both versions, and the ListWatcher failed to list the V2 object. We also only see errors from the V2 CRD object, because the restClient can only take only one type of CRD spec. In this example I gave it the v1 schema, thus the restClient is attempting to cast the JSON onto the V1 schema and showing the errors about mismatched type.

# view JSON objects stored in kubernetes
```
kubectl proxy --port=8080 &
curl http://localhost:8080/apis/samplecontroller.k8s.io/v1alpha1/foos
curl http://localhost:8080/apis/samplecontroller.k8s.io/v2/foos
```

expected output, notice the differences in spec.convertSpec
```json
{
    "apiVersion": "samplecontroller.k8s.io/v1alpha1",
    "items": [{
        "apiVersion": "samplecontroller.k8s.io/v1alpha1",
        "kind": "Foo",
        "metadata": {
            "creationTimestamp": "2020-01-13T22:46:17Z",
            "generation": 1,
            "name": "example-foo",
            "namespace": "default",
            "resourceVersion": "1723743",
            "selfLink": "/apis/samplecontroller.k8s.io/v1alpha1/namespaces/default/foos/example-foo",
            "uid": "825996b4-3656-11ea-b081-e291c83199d7"
        },
        "spec": {
            "convertSpec": 0,
            "deploymentName": "example-foo",
            "replicas": 1
        }
    }, {
        "apiVersion": "samplecontroller.k8s.io/v1alpha1",
        "kind": "Foo",
        "metadata": {
            "creationTimestamp": "2020-01-13T22:47:06Z",
            "generation": 1,
            "name": "example-foo-v2",
            "namespace": "default",
            "resourceVersion": "1723872",
            "selfLink": "/apis/samplecontroller.k8s.io/v1alpha1/namespaces/default/foos/example-foo-v2",
            "uid": "9fc39e76-3656-11ea-b081-e291c83199d7"
        },
        "spec": {
            "convertSpec": "int value",
            "deploymentName": "example-foo-v2",
            "replicas": 1
        }
    }],
    "kind": "FooList",
    "metadata": {
        "continue": "",
        "resourceVersion": "1724085",
        "selfLink": "/apis/samplecontroller.k8s.io/v1alpha1/foos"
    }
}
```