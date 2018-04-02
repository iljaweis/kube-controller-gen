package main

import (
	"flag"
	"fmt"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"

	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	appsv1beta2 "k8s.io/api/apps/v1beta2"
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

func (c *Controller) DeploymentCreatedOrUpdated(deploy *extensionsv1beta1.Deployment) error {
	fmt.Printf("Deployment %s/%s created or updated\n", deploy.Namespace, deploy.Name)
	return nil
}

func (c *Controller) DeploymentDeleted(deploy *extensionsv1beta1.Deployment) error {
	fmt.Printf("Deployment %s/%s deleted\n", deploy.Namespace, deploy.Name)
	return nil
}

func (c *Controller) ReplicaSetCreatedOrUpdated(rs *extensionsv1beta1.ReplicaSet) error {
	fmt.Printf("ReplicaSet %s/%s created or updated\n", rs.Namespace, rs.Name)
	return nil
}

func (c *Controller) ReplicaSetDeleted(rs *extensionsv1beta1.ReplicaSet) error {
	fmt.Printf("ReplicaSet %s/%s deleted\n", rs.Namespace, rs.Name)
	return nil
}

func (c *Controller) DaemonSetCreatedOrUpdated(ds *extensionsv1beta1.DaemonSet) error {
	fmt.Printf("DaemonSet %s/%s created or updated\n", ds.Namespace, ds.Name)
	return nil
}

func (c *Controller) DaemonSetDeleted(ds *extensionsv1beta1.DaemonSet) error {
	fmt.Printf("DaemonSet %s/%s deleted\n", ds.Namespace, ds.Name)
	return nil
}

func (c *Controller) StatefulSetCreatedOrUpdated(statefulset *appsv1beta2.StatefulSet) error {
	fmt.Printf("StatefulSet %s/%s created or updated\n", statefulset.Namespace, statefulset.Name)
	return nil
}

func (c *Controller) StatefulSetDeleted(statefulset *appsv1beta2.StatefulSet) error {
	fmt.Printf("StatefulSet %s/%s deleted\n", statefulset.Namespace, statefulset.Name)
	return nil
}


