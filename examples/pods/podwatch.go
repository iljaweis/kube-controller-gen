package main

import (
	"flag"
	"fmt"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"

	corev1 "k8s.io/api/core/v1"
)

func main() {

	var kubeconfig string

	if e := os.Getenv("KUBECONFIG"); e != "" {
		kubeconfig = e
	}

	flag.StringVar(&kubeconfig, "kubeconfig", os.Getenv("HOME")+"/.kube/config", "location of your kubeconfig")
	flag.Parse()

	clientConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		panic(err.Error())
	}

	c := &Controller{Kubernetes: clientset}

	c.Initialize()
	c.Start()
}

func (c *Controller) PodCreatedOrUpdated(pod *corev1.Pod) error {
	fmt.Printf("Pod %s/%s created or updated\n", pod.Namespace, pod.Name)
	return nil
}

func (c *Controller) PodDeleted(pod *corev1.Pod) error {
	fmt.Printf("Pod %s/%s deleted\n", pod.Namespace, pod.Name)
	return nil
}
