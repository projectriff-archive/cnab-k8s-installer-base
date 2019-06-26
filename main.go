/*
 * Copyright 2019 The original author or authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/projectriff/cnab-k8s-installer-base/pkg/apis/kab/v1alpha1"
	"github.com/projectriff/cnab-k8s-installer-base/pkg/client/clientset/versioned"
	"github.com/projectriff/cnab-k8s-installer-base/pkg/kab"
	"github.com/projectriff/cnab-k8s-installer-base/pkg/kubectl"
	"github.com/projectriff/cnab-k8s-installer-base/pkg/kustomize"
	"github.com/projectriff/cnab-k8s-installer-base/pkg/registry"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	apiext "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	KUSTOMIZE_TIMEOUT       = 30 * time.Second
	CNAB_ACTION_ENV_VAR     = "CNAB_ACTION"
	MANIFEST_FILE_ENV_VAR   = "MANIFEST_FILE"
	TARGET_REGISTRY_ENV_VAR = "TARGET_REGISTRY"
	LOG_LEVEL_ENV_VAR       = "LOG_LEVEL"
)

func main() {

	setupLogging()

	path := getEnv(MANIFEST_FILE_ENV_VAR)
	if path == "" {
		// revert after duffle fixes the export parameter issue
		// https://github.com/deislabs/duffle/issues/753
		path = "/cnab/app/kab/manifest.yaml"
	}
	action := getEnv(CNAB_ACTION_ENV_VAR)
	action = strings.ToLower(action)
	log.Debugf("performing action: %s, manifest file: %s", action, path)
	switch action {
	case "install":
		install(path)
	case "uninstall":
		uninstall()
	case "upgrade":
	default:
		log.Fatalf("unknown action '%s'. please set CNAB_ACTION environment variable", action)
	}
}

func install(path string) {
	manifest, err := v1alpha1.NewManifest(path)
	if err != nil {
		_, err = fmt.Fprintf(os.Stderr, "error while reading from %s: %v", path, err)
		os.Exit(1)
	}
	knbClient, err := createKnbClient()
	if err != nil {
		log.Fatalln(err)
	}
	err = knbClient.PatchManifest(manifest)
	if err != nil {
		log.Fatalln(err)
	}
	err = knbClient.Relocate(manifest, getEnv(TARGET_REGISTRY_ENV_VAR))
	if err != nil {
		log.Fatalln(err)
	}
	err = knbClient.Install(manifest)
	if err != nil {
		log.Fatalf("error while installing from %s: %v\n", path, err)
	}
}

func uninstall()  {
	knbClient, err := createKnbClient()
	if err != nil {
		log.Fatalln(err)
	}
	err = knbClient.Uninstall(kab.GetInstallationName())
	if err != nil {
		log.Fatalln(err)
	}
}

func createKnbClient() (*kab.Client, error) {
	config, err := getRestConfig()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Could not get kubernetes configuration: %s", err))
	}
	coreClient, err := getCoreClient(config)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Could not create kubernetes core client: %s", err))
	}
	extClient, err := getExtensionClient(config)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Could not create kubernetes extension client: %s", err))
	}
	kabClient, err := getKabClient(config)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Could not create kubernetes kab client: %s", err))
	}
	dClient, err := registry.NewClient()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Could not create docker client: %s", err))
	}
	kustomizer := kustomize.MakeKustomizer(KUSTOMIZE_TIMEOUT)

	ctl := kubectl.RealKubeCtl()

	knbClient := kab.NewKnbClient(coreClient, extClient, kabClient, dClient, kustomizer, ctl)
	return knbClient, nil
}

func getRestConfig() (*rest.Config, error) {
	config, err := getOutOfClusterRestConfig()
	if err != nil {
		log.Debugln("error getting out of cluster rest config, trying in cluster")
		return getInClusterRestConfig()
	}
	return config, nil
}

func getOutOfClusterRestConfig() (*rest.Config, error) {
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

func getInClusterRestConfig() (*rest.Config, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return config, nil
}

func getCoreClient(config *rest.Config) (kubernetes.Interface, error) {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}

func getExtensionClient(config *rest.Config) (apiext.Interface, error) {
	extClient, err := apiext.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return extClient, nil
}

func getKabClient(config *rest.Config) (versioned.Interface, error) {
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

func setupLogging() {
	log.SetOutput(os.Stdout)
	log.SetLevel(getLogLevel())
}

func getLogLevel() log.Level {
	requestedLevel := getEnv(LOG_LEVEL_ENV_VAR)
	if requestedLevel == "" {
		return log.InfoLevel
	}
	level, err := log.ParseLevel(requestedLevel)
	if err != nil {
		log.Fatalf("Unknown log level %s", requestedLevel)
	}
	return level
}

// duffle sets the env value to "<nil>", so restore normal behavior
func getEnv(env_var string) (string) {
	val := os.Getenv(env_var)
	if strings.Contains(val, "nil") {
		return ""
	}
	return val
}
