package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"strings"
	"text/template"

	"gopkg.in/yaml.v2"
)

type Resource struct {
	Name   string
	Plural string
	Scope  string
	Create bool
	Update bool
	Delete bool
}

type Api struct {
	Name      string
	Group     string
	Version   string
	Resources []Resource
}

type Clientset struct {
	Name          string
	Import        string
	Defaultresync int
	Apis          []Api
}

type Config struct {
	Package         string
	Clientsets      []Clientset
	Controllerextra string
	Imports         string
}

func main() {
	var configFile, outputDir string

	flag.StringVar(&configFile, "c", "controller-gen.yaml", "configuration file")
	flag.StringVar(&outputDir, "o", "./", "output directory")

	flag.Parse()

	configFileContents, err := ioutil.ReadFile(configFile)
	if err != nil {
		panic(err)
	}

	var c Config

	err = yaml.Unmarshal(configFileContents, &c)
	if err != nil {
		panic(err)
	}

	t1 := `package {{ .Package }}

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	kubernetesinformers "k8s.io/client-go/informers"

{{ range $clientset := .Clientsets }}
{{ range $api := .Apis }}
{{ if eq $clientset.Name "kubernetes" }}
{{ if eq $api.Name "core" }}
{{ if anyresourcewithdelete $api }}
	corev1 "k8s.io/api/core/v1"
{{ end }}
	corelisterv1 "k8s.io/client-go/listers/core/v1"
{{ else }}
{{ if anyresourcewithdelete $api }}
	{{ $api.Name }}{{ $api.Version }} "k8s.io/api/{{ $api.Name }}/{{ $api.Version }}"
{{ end }}
	{{ $api.Name }}lister{{ $api.Version }} "k8s.io/client-go/listers/{{ $api.Name }}/{{ $api.Version }}"
{{ end }}
{{ else }}
	{{ $api.Name }}clientset "{{ $clientset.Import }}/pkg/client/clientset/versioned"
{{ if anyresourcewithdelete $api }}
	{{ $api.Name }}{{ $api.Version }} "{{ $clientset.Import }}/pkg/apis/{{ $api.Group }}/{{ $api.Version }}"
{{ end -}}
	{{ $api.Name }}informers "{{ $clientset.Import }}/pkg/client/informers/externalversions"
	{{ $api.Name }}lister{{ $api.Version }} "{{ $clientset.Import }}/pkg/client/listers/{{ $api.Group }}/{{ $api.Version }}"
{{ end -}}
{{ end -}}
{{ end -}}

{{ .Imports }}
)

type Controller struct {
{{ range $clientset := .Clientsets -}}
{{ if eq $clientset.Name "kubernetes" }}
	Kubernetes kubernetes.Interface
	KubernetesFactory kubernetesinformers.SharedInformerFactory
{{- else }}
	{{ title $clientset.Name }}Client {{ $clientset.Name }}clientset.Interface
	{{ title $clientset.Name }}Factory {{ $clientset.Name }}informers.SharedInformerFactory
{{ end }}

{{ range $api := $clientset.Apis -}}

{{ range $res := $api.Resources }}
	{{ title $res.Name }}Queue workqueue.RateLimitingInterface
	{{ title $res.Name }}Lister {{ $api.Name }}lister{{ $api.Version }}.{{ title $res.Name }}Lister
	{{ title $res.Name }}Synced cache.InformerSynced
{{ end -}}{{/* resources*/}}
{{ end -}}{{/* api */}}
{{ end -}}{{/* clientset */}}

{{ .Controllerextra }}
}

// Expects the clientsets to be set.
func (c *Controller) Initialize() {
{{ range $clientset := .Clientsets }}
	if c.{{ clientsetname $clientset }} == nil {
		panic("c.{{ clientsetname $clientset }} is nil")
	}
	c.{{ title $clientset.Name }}Factory = {{ $clientset.Name }}informers.NewSharedInformerFactory(c.{{ clientsetname $clientset }}, time.Second*{{ $clientset.Defaultresync }})

{{ range $api := $clientset.Apis }}

{{ range $res := $api.Resources }}
	{{ $res.Name }}Informer := c.{{ title $clientset.Name }}Factory.{{ title $api.Name }}().{{ title $api.Version }}().{{ $res.Plural }}()
	{{ $res.Name }}Queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	c.{{ $res.Name }}Queue = {{ $res.Name }}Queue
	c.{{ $res.Name }}Lister = {{ $res.Name }}Informer.Lister()
	c.{{ $res.Name }}Synced = {{ $res.Name }}Informer.Informer().HasSynced

	{{ $res.Name }}Informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
{{ if $res.Create }}
		AddFunc: func(obj interface{}) {
			if key, err := cache.MetaNamespaceKeyFunc(obj); err == nil {
				{{ $res.Name }}Queue.Add(key)
			}
		},
{{ end }}
{{ if $res.Update }}
		UpdateFunc: func(old, new interface{}) {
			if key, err := cache.MetaNamespaceKeyFunc(new); err == nil {
				{{ $res.Name }}Queue.Add(key)
			}
		},
{{ end }}
{{ if $res.Delete }}
		DeleteFunc: func(obj interface{}) {
			o, ok := obj.(*{{ $api.Name }}{{ $api.Version }}.{{ $res.Name }})

			if !ok {
				tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					log.Errorf("couldn't get object from tombstone %+v", obj)
					return
				}
				o, ok = tombstone.Obj.(*{{ $api.Name }}{{ $api.Version }}.{{ $res.Name }})
				if !ok {
					log.Errorf("tombstone contained object that is not a {{ $res.Name }} %+v", obj)
					return
				}
			}

			err := c.{{ $res.Name }}Deleted(o)

			if err != nil {
				log.Errorf("failed to process deletion: %s", err.Error())
			}
		},
{{ end }}
	})


{{ end -}}{{/* res */}}
{{ end -}}{{/* api */}}
{{ end }}{{/* clientset */}}
	return
}

func (c *Controller) Start() {
	stopCh := make(chan struct{})
	defer close(stopCh)
{{ range $clientset := .Clientsets -}}
	go c.{{ title $clientset.Name }}Factory.Start(stopCh)
{{ end }}
	go c.Run(stopCh)

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGTERM)
	signal.Notify(sigterm, syscall.SIGINT)
	<-sigterm
}

func (c *Controller) Run(stopCh <-chan struct{}) {

	log.Infof("starting controller")

	defer runtime.HandleCrash()
{{ range $clientset := .Clientsets -}}
{{ range $api := $clientset.Apis -}}
{{ range $res := $api.Resources }}
	defer c.{{ $res.Name }}Queue.ShutDown()
{{- end -}}
{{ end -}}
{{ end }}

	if !cache.WaitForCacheSync(stopCh, {{ allsynced . }}) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	log.Debugf("starting workers")

{{ range $clientset := .Clientsets -}}
{{ range $api := $clientset.Apis -}}
{{ range $res := $api.Resources }}
	go wait.Until(c.run{{ $res.Name }}Worker, time.Second, stopCh)
{{ end -}}
{{ end -}}
{{ end }}

	log.Debugf("started workers")
	<-stopCh
	log.Debugf("shutting down workers")
}

{{ range $clientset := .Clientsets -}}
{{ range $api := $clientset.Apis -}}
{{ range $res := $api.Resources }}

func (c *Controller) run{{ $res.Name }}Worker() {
	for c.processNext{{ $res.Name }}() {
	}
}

func (c *Controller) processNext{{ $res.Name }}() bool {
	obj, shutdown := c.{{ $res.Name }}Queue.Get()
	if shutdown {
		return false
	}

	err := func(obj interface{}) error {
		defer c.{{ $res.Name }}Queue.Done(obj)
		var key string
		var ok bool

		if key, ok = obj.(string); !ok {
			c.{{ $res.Name }}Queue.Forget(obj)
			runtime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}

		if err := c.process{{ $res.Name }}(key); err != nil {
			return fmt.Errorf("error syncing '%s': %s", key, err.Error())
		}

		c.{{ $res.Name }}Queue.Forget(obj)
		return nil
	}(obj)

	if err != nil {
		runtime.HandleError(err)
		return true
	}

	return true
}

func (c *Controller) process{{ $res.Name }}(key string) error {
{{ if eq $res.Scope "Namespaced" }}
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return fmt.Errorf("could not parse name %s: %s", key, err.Error())
	}

	o, err := c.{{ $res.Name }}Lister.{{ $res.Plural }}(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			return fmt.Errorf("tried to get %s, but it was not found", key)
		} else {
			return fmt.Errorf("error getting %s from cache: %s", key, err.Error())
		}
	}

	return c.{{ $res.Name }}CreatedOrUpdated(o)
{{ else }}
	name := key

	o, err := c.{{ $res.Name }}Lister.Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			return fmt.Errorf("tried to get %s, but it was not found", key)
		} else {
			return fmt.Errorf("error getting %s from cache: %s", key, err.Error())
		}
	}

	return c.{{ $res.Name }}CreatedOrUpdated(o)
{{ end }}
}
{{- end -}}
{{ end -}}
{{ end -}}

`

	funcMap := template.FuncMap{
		"title": func(s string) string {
			return strings.Title(s)
		},
		"clientsetname": func(cs Clientset) string {
			if cs.Name == "kubernetes" {
				return "Kubernetes"
			} else {
				return strings.Title(cs.Name) + "Client"
			}
		},
		"allsynced": func(c Config) string {
			var syncs []string
			for _, cs := range c.Clientsets {
				for _, a := range cs.Apis {
					for _, r := range a.Resources {
						syncs = append(syncs, "c."+r.Name+"Synced")
					}
				}
			}
			return strings.Join(syncs, ", ")
		},
		"anyresourcewithdelete": func(a Api) bool {
			for _, r := range a.Resources {
				if r.Delete {
					return true
				}
			}
			return false
		},
	}

	t, err := template.New("controller").Funcs(funcMap).Parse(t1)

	var buffer bytes.Buffer

	err = t.Execute(&buffer, c)

	err = ioutil.WriteFile(outputDir+"zz_generated_controller.go", buffer.Bytes(), 0644)
	if err != nil {
		panic(err)
	}
}
