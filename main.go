package main

import (
	"cnab-k8s-installer-base/pkg/apis/kab/v1alpha1"
	"cnab-k8s-installer-base/pkg/client/clientset/versioned"
	"cnab-k8s-installer-base/pkg/docker"
	"cnab-k8s-installer-base/pkg/kab"
	"flag"
	"fmt"
	apiext "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
)

const BaseDir = "/cnab/app/kab"

func main()  {

	args := os.Args[1:]
	var path string
	if len(args) > 0 {
		path = args[0]
	} else {
		path = "/cnab/app/kab/template.yaml"
	}
	action := os.Getenv("CNAB_ACTION")
	switch action {
	case "install":
		install(path)
	case "uninstall":
	case "upgrade":
	}
}

func install(path string) {
	manifest, err := v1alpha1.NewManifest(path)
	if err != nil {
		_, err = fmt.Fprintf(os.Stderr, "error while reading from %s: %v", path, err)
		os.Exit(1)
	}
	knbClient := createKnbClient()
	err = knbClient.Relocate(manifest, os.Getenv("target_registry"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	err = knbClient.Install(manifest, BaseDir)
	if err != nil {
		_, err = fmt.Fprintf(os.Stderr, "error while installing from %s: %v", path, err)
		os.Exit(1)
	}
}

func createKnbClient() *kab.Client {
	config, err := getRestConfig()
	if err != nil {
		fmt.Println("Could not get kubernetes configuration", err)
	}
	coreClient, err := getCoreClient(config)
	if err != nil {
		fmt.Println("Could not create kubernetes core client", err)
	}
	extClient, err := getExtensionClient(config)
	if err != nil {
		fmt.Println("Could not create kubernetes extension client", err)
	}
	kabClient, err := getKabClient(config)
	if err != nil {
		fmt.Println("Could not create kubernetes kab client", err)
	}
	dClient, err := docker.NewDockerClient()
	if err != nil {
		fmt.Println("Could not create kubernetes kab client", err)
	}

	knbClient := kab.NewKnbClient(coreClient, extClient, kabClient, dClient)
	return knbClient
}

func getRestConfig() (*rest.Config, error) {
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func getCoreClient(config *rest.Config) (*kubernetes.Clientset, error) {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}

func getExtensionClient(config *rest.Config) (*apiext.Clientset, error) {
	extClient, err := apiext.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return extClient, nil
}

func getKabClient(config *rest.Config) (*versioned.Clientset, error) {
	kabClient, err := versioned.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return kabClient, nil
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}