package main

import (
	"encoding/json"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes"
	rest "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"

	foov1 "github.com/cnadolny/sample-controller/pkg/apis/samplecontroller/v1alpha1"
	foov2 "github.com/cnadolny/sample-controller/pkg/apis/samplecontroller/v2"
)

const controllerAgentName = "sample-controller"

type EventType int

const (
	Created EventType = 0
	Deleted EventType = 1
	Updated EventType = 2
)

// Controller is the controller implementation for Foo resources
type Controller struct {
	kubeclientset kubernetes.Interface

	fooInformerV1 cache.SharedInformer
	fooInformerV2 cache.SharedInformer

	restClient   *rest.RESTClient
	eventChannel chan EventType
}

func NewController(
	config *rest.Config,
	kubeclientset kubernetes.Interface,
	eventCh chan EventType) *Controller {

	restClient, err := newRestClient(config)
	if err != nil {
		klog.Fatalf("Error: %s", err.Error())
	}

	fooListWatch := newFooListWatch(restClient)
	fooInformerv1, fooInformerv2, err := newFooInformer(restClient, eventCh, fooListWatch)
	if err != nil {
		klog.Error(err)
		return nil
	}

	controller := &Controller{
		restClient:    restClient,
		kubeclientset: kubeclientset,
		fooInformerV1: fooInformerv1,
		fooInformerV2: fooInformerv2,
		eventChannel:  eventCh,
	}

	return controller
}

func (c *Controller) Run(stopCh <-chan struct{}) error {
	go c.fooInformerV1.Run(stopCh)
	go c.fooInformerV2.Run(stopCh)
	var event EventType

	for {
		select {
		case <-stopCh:
			return nil
		case event = <-c.eventChannel:
			klog.Info("Received event: ", event)
		}

		var result map[string]interface{}
		data, err := c.restClient.Get().AbsPath("apis/samplecontroller.k8s.io/v1alpha1/foos").DoRaw()
		if err != nil {
			klog.Fatalf("Error: %s", err.Error())
		}

		json.Unmarshal(data, &result)
		fmt.Println("\nPrinting V1 JSON\n-----------------")
		for _, v := range result["items"].([]interface{}) {
			fmt.Println(v.(map[string]interface{})["apiVersion"])
			fmt.Println(v.(map[string]interface{})["metadata"].(map[string]interface{})["name"])
			fmt.Println(v.(map[string]interface{})["spec"])
		}
		fmt.Println("-----------------")

		var result2 map[string]interface{}
		datav2, err := c.restClient.Get().AbsPath("apis/samplecontroller.k8s.io/v2/foos").DoRaw()
		if err != nil {
			klog.Fatalf("Error: %s", err.Error())
		}

		fmt.Println("\nPrinting V2 JSON\n-----------------")
		json.Unmarshal(datav2, &result2)
		for _, v := range result2["items"].([]interface{}) {
			fmt.Println(v.(map[string]interface{})["apiVersion"])
			fmt.Println(v.(map[string]interface{})["metadata"].(map[string]interface{})["name"])
			fmt.Println(v.(map[string]interface{})["spec"])
		}
		fmt.Println("-----------------\n")

		// List foo objects from ListWatch V1
		list := c.fooInformerV1.GetStore().List()
		for _, foo := range list {
			o, ok := foo.(*foov1.Foo)
			if !ok {
				o2, ok2 := foo.(*foov2.Foo)
				if !ok2 {
					err := fmt.Errorf("could not cast %T to %s", foo, "foov2")
					fmt.Println(err)
				} else {
					fmt.Println("\n Listing foo V2 object, ", *o2)
				}
			} else {
				fmt.Println("\n Listing foo V1 object", o.Name, o.Spec)
			}
		}

		// List foo objects from ListWatch V2
		listV2 := c.fooInformerV2.GetStore().List()
		for _, fooV2 := range listV2 {
			o, ok := fooV2.(*foov2.Foo)
			if !ok {
				o2, ok2 := fooV2.(*foov1.Foo)
				if !ok2 {
					err := fmt.Errorf("FooInformerV2: could not cast %T to %s", fooV2, "foov1")
					fmt.Println(err)
				} else {
					fmt.Println("\n FooInformerV2: Listing foo V1 object, ", *o2)
				}
			} else {
				fmt.Println("\n FooInformerV2: Listing foo V2 object", o.Name, o.Spec)
			}
		}
	}

	return nil
}

func newRestClient(config *rest.Config) (r *rest.RESTClient, err error) {
	crdconfig := *config
	crdconfig.GroupVersion = &schema.GroupVersion{Group: "samplecontroller.k8s.io", Version: "v1alpha1"}
	crdconfig.APIPath = "/apis"
	crdconfig.ContentType = runtime.ContentTypeJSON
	s := runtime.NewScheme()
	s.AddKnownTypes(*crdconfig.GroupVersion,
		&foov1.Foo{},
		&foov1.FooList{})
	crdconfig.NegotiatedSerializer = serializer.WithoutConversionCodecFactory{
		CodecFactory: serializer.NewCodecFactory(s)}

	restClient, err := rest.RESTClientFor(&crdconfig)
	if err != nil {
		return nil, err
	}
	return restClient, nil
}

func newFooListWatch(r *rest.RESTClient) *cache.ListWatch {
	return cache.NewListWatchFromClient(r, "foos", corev1.NamespaceAll, fields.Everything())
}

func newFooInformer(r *rest.RESTClient, eventCh chan EventType, lw *cache.ListWatch) (cache.SharedInformer, cache.SharedInformer, error) {
	fooInformer := cache.NewSharedInformer(
		lw,
		&foov1.Foo{},
		time.Minute*10)
	if fooInformer == nil {
		return nil, nil, nil
	}
	fooInformer.AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				klog.Info("created")
				eventCh <- Created
			},
			DeleteFunc: func(obj interface{}) {
				klog.Info("deleted")
				eventCh <- Deleted
			},
			UpdateFunc: func(OldObj, newObj interface{}) {
				klog.Info("updated")
				eventCh <- Updated
			},
		},
	)

	fooInformerv2 := cache.NewSharedInformer(
		lw,
		&foov2.Foo{},
		time.Minute*10)
	if fooInformerv2 == nil {
		return nil, nil, nil
	}
	fooInformerv2.AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				klog.Info("created")
				eventCh <- Created
			},
			DeleteFunc: func(obj interface{}) {
				klog.Info("deleted")
				eventCh <- Deleted
			},
			UpdateFunc: func(OldObj, newObj interface{}) {
				klog.Info("updated")
				eventCh <- Updated
			},
		},
	)
	return fooInformer, fooInformerv2, nil
}
