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
kubectl get po
kubectl logs <pod name>

kubectl create -f artifacts/crd/example-v1.yaml
kubectl logs <pod name>

log output
```
 result V1: %v
map[apiVersion:samplecontroller.k8s.io/v1alpha1 items:[map[apiVersion:samplecontroller.k8s.io/v1alpha1 kind:Foo metadata:map[creationTimestamp:2020-01-13T22:46:17Z generation:1 name:example-foo namespace:default resourceVersion:1723743 selfLink:/apis/samplecontroller.k8s.io/v1alpha1/namespaces/default/foos/example-foo uid:825996b4-3656-11ea-b081-e291c83199d7] spec:map[convertSpec:0 deploymentName:example-foo replicas:1]]] kind:FooList metadata:map[continue: resourceVersion:1723743 selfLink:/apis/samplecontroller.k8s.io/v1alpha1/foos]]
I0113 22:46:17.041225       1 controller.go:114]
 result V2: %v
map[apiVersion:samplecontroller.k8s.io/v2 items:[map[apiVersion:samplecontroller.k8s.io/v2 kind:Foo metadata:map[creationTimestamp:2020-01-13T22:46:17Z generation:1 name:example-foo namespace:default resourceVersion:1723743 selfLink:/apis/samplecontroller.k8s.io/v2/namespaces/default/foos/example-foo uid:825996b4-3656-11ea-b081-e291c83199d7] spec:map[convertSpec:0 deploymentName:example-foo replicas:1]]] kind:FooList metadata:map[continue: resourceVersion:1723743 selfLink:/apis/samplecontroller.k8s.io/v2/foos]]
I0113 22:46:17.041274       1 controller.go:129]
 Listing foo V1 object, %v
{{ } {example-foo  default /apis/samplecontroller.k8s.io/v1alpha1/namespaces/default/foos/example-foo 825996b4-3656-11ea-b081-e291c83199d7 1723743 1 2020-01-13 22:46:17 +0000 UTC <nil> <nil> map[] map[] [] []  []} {example-foo 0xc000045020 0} {0}}
```

kubectl create -f artifcats/crd/example-v2.yaml
```
W0113 22:47:06.402394       1 reflector.go:340] pkg/mod/k8s.io/client-go@v0.0.0-20200111153838-ea0a6e11838c/tools/cache/reflector.go:108: watch of *v2.Foo ended with: an error on the server ("unable to decode an event from the watch stream: unable to decode watch event: v1alpha1.Foo.Spec: v1alpha1.FooSpec.ConvertSpec: readUint64: unexpected character: \xff, error found in #10 byte of ...|ertSpec\":\"int value\"|..., bigger context ...|6-11ea-b081-e291c83199d7\"},\"spec\":{\"convertSpec\":\"int value\",\"deploymentName\":\"example-foo-v2\",\"repl|...") has prevented the request from succeeding
...
 result V1: %v
map[apiVersion:samplecontroller.k8s.io/v1alpha1 items:[map[apiVersion:samplecontroller.k8s.io/v1alpha1 kind:Foo metadata:map[creationTimestamp:2020-01-13T22:46:17Z generation:1 name:example-foo namespace:default resourceVersion:1723743 selfLink:/apis/samplecontroller.k8s.io/v1alpha1/namespaces/default/foos/example-foo uid:825996b4-3656-11ea-b081-e291c83199d7] spec:map[convertSpec:0 deploymentName:example-foo replicas:1]] map[apiVersion:samplecontroller.k8s.io/v1alpha1 kind:Foo metadata:map[creationTimestamp:2020-01-13T22:47:06Z generation:1 name:example-foo-v2 namespace:default resourceVersion:1723872 selfLink:/apis/samplecontroller.k8s.io/v1alpha1/namespaces/default/foos/example-foo-v2 uid:9fc39e76-3656-11ea-b081-e291c83199d7] spec:map[convertSpec:int value deploymentName:example-foo-v2 replicas:1]]] kind:FooList metadata:map[continue: resourceVersion:1723874 selfLink:/apis/samplecontroller.k8s.io/v1alpha1/foos]]
I0113 22:47:07.415989       1 controller.go:114]
 result V2: %v
map[apiVersion:samplecontroller.k8s.io/v2 items:[map[apiVersion:samplecontroller.k8s.io/v2 kind:Foo metadata:map[creationTimestamp:2020-01-13T22:46:17Z generation:1 name:example-foo namespace:default resourceVersion:1723743 selfLink:/apis/samplecontroller.k8s.io/v2/namespaces/default/foos/example-foo uid:825996b4-3656-11ea-b081-e291c83199d7] spec:map[convertSpec:0 deploymentName:example-foo replicas:1]] map[apiVersion:samplecontroller.k8s.io/v2 kind:Foo metadata:map[creationTimestamp:2020-01-13T22:47:06Z generation:1 name:example-foo-v2 namespace:default resourceVersion:1723872 selfLink:/apis/samplecontroller.k8s.io/v2/namespaces/default/foos/example-foo-v2 uid:9fc39e76-3656-11ea-b081-e291c83199d7] spec:map[convertSpec:int value deploymentName:example-foo-v2 replicas:1]]] kind:FooList metadata:map[continue: resourceVersion:1723874 selfLink:/apis/samplecontroller.k8s.io/v2/foos]]
```

# view JSON objects stored in kubernetes
kubectl proxy --port=8080 &
curl http://localhost:8080/apis/samplecontroller.k8s.io/v1alpha1/foos
curl http://localhost:8080/apis/samplecontroller.k8s.io/v2/foos

expected output, notice the differences in spec.convertSpec
```
{"apiVersion":"samplecontroller.k8s.io/v1alpha1","items":[{"apiVersion":"samplecontroller.k8s.io/v1alpha1","kind":"Foo","metadata":{"creationTimestamp":"2020-01-13T22:46:17Z","generation":1,"name":"example-foo","namespace":"default","resourceVersion":"1723743","selfLink":"/apis/samplecontroller.k8s.io/v1alpha1/namespaces/default/foos/example-foo","uid":"825996b4-3656-11ea-b081-e291c83199d7"},"spec":{"convertSpec":0,"deploymentName":"example-foo","replicas":1}},{"apiVersion":"samplecontroller.k8s.io/v1alpha1","kind":"Foo","metadata":{"creationTimestamp":"2020-01-13T22:47:06Z","generation":1,"name":"example-foo-v2","namespace":"default","resourceVersion":"1723872","selfLink":"/apis/samplecontroller.k8s.io/v1alpha1/namespaces/default/foos/example-foo-v2","uid":"9fc39e76-3656-11ea-b081-e291c83199d7"},"spec":{"convertSpec":"int value","deploymentName":"example-foo-v2","replicas":1}}],"kind":"FooList","metadata":{"continue":"","resourceVersion":"1724085","selfLink":"/apis/samplecontroller.k8s.io/v1alpha1/foos"}}
```