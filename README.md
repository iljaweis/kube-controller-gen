# kube-controller-gen

Generate utility code for writing Kubernetes controllers in Go.

## How to use it

### Write a controller-gen.yaml

This file describes the clientsets, APIs and Resources to watch.

### Generate the controller code

```bash
kube-controller-gen
```

This creates a file `zz_generated_controller.go` that contains all the framework code.

### Write your controller

Write your controller. You will need an instance of a Controller:

```go

c := &Controller{Kubernetes: clientset}

c.Initialize()
c.Start()

```

For every resource you want to watch (and you have declared in `controller-gen.yaml`), write functions:

```go
func (c *Controller) XXXCreatedOrUpdated(x *XXX) error {
	// do something with your new or updated x
	return nil
}

func (c *Controller) XXXDeleted(x *XXX) error {
	// do something with your x as it is being deleted
	return nil
}
```

See `examples` for complete examples.

## controller-gen.yaml

Top level parameters:

- __package__ is the packaged name used for the generated code.
- __clientsets__ is a list of clientsets to use.

Clientsets:

- __name__ is the name of the clientset to use. Use `kubernetes` to generate the standard kubernetes clientset. Custom resources are supported here, use a symbolic name for your clientset then.
- __apis__ is a list of APIs used.
- __import__ (CRD only) is the name go package to import.

APIs:

- __name__ is the API name. For example, `core` for core Kubernetes resources like Pods or Services.
- __version__ is the API version
- __defaultresync__ controls how often to check your resources, in addition to reacting to events. In seconds. Set to 0 to disable resync.
- __resources__ are the resources in API to watch.
- __group__ is the API group name (empty for `core`)

Resources:

- __name__ is the name, like `Pod`.
- __plural__ is the plural of the name, like `Pods`.
- __scope__ is `Namespaced` or `Cluster`.
- __create__ / __update__ / __delete__ are booleans and control which of these event types you want to watch for your resource type.
