package kab

import (
	"cnab-k8s-installer-base/pkg/apis/kab/v1alpha1"
	"errors"
	"fmt"
	"github.com/projectriff/riff/pkg/env"
	"github.com/projectriff/riff/pkg/fileutils"
	"github.com/projectriff/riff/pkg/kubectl"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"strings"
	"time"
)

func (c *Client) Install(manifest *v1alpha1.Manifest, basedir string) error {
	err := CreateCRD(c.extClient)
	if err != nil {
		return errors.New(fmt.Sprintf("Could not create kab CRD: %s ", err))
	}
	manifest, err = c.createCRDObject(manifest, backOffSettings())
	if err != nil {
		return errors.New(fmt.Sprintf("Could not install riff: %s ", err))
	}
	fmt.Println("Installing", env.Cli.Name, "components")
	fmt.Println()
	err = c.installAndCheckResources(manifest, basedir)
	if err != nil {
		return errors.New(fmt.Sprintf("Could not install riff: %s ", err))
	}
	fmt.Print("Kubernetes Application Bundle installed\n\n")
	return nil
}

func backOffSettings() wait.Backoff {
	return wait.Backoff{
		Duration: minRetryInterval,
		Factor:   exponentialBackoffBase,
		Steps:    maxRetries,
	}
}

func (c *Client) createCRDObject(manifest *v1alpha1.Manifest, backOffSettings wait.Backoff) (*v1alpha1.Manifest, error) {

	err := wait.ExponentialBackoff(backOffSettings, func() (bool, error) {
		old, err := c.kabClient.ProjectriffV1alpha1().Manifests(manifest.Namespace).Get(manifest.Name, metav1.GetOptions{})
		if err != nil && !strings.Contains(err.Error(), "not found") {
			return false, nil
		}
		if !isEmpty(old) {
			return true, errors.New(fmt.Sprintf("%s already installed", env.Cli.Name))
		}
		_, err = c.kabClient.ProjectriffV1alpha1().Manifests(manifest.Namespace).Create(manifest)
		if err != nil {
			return false, nil
		}
		return true, nil
	})
	if err == wait.ErrWaitTimeout {
		return nil, errors.New(fmt.Sprintf("timed out creating %s custom resource defiition", env.Cli.Name))
	}
	return manifest, err
}

func isEmpty(manifest *v1alpha1.Manifest) bool {
	if len(manifest.Spec.Resources) == 0 {
		return true
	}
	return false
}

func (c *Client) installAndCheckResources(manifest *v1alpha1.Manifest, basedir string) error {
	for _, resource := range manifest.Spec.Resources {
		if resource.Deferred {
			fmt.Printf("Skipping install of %s\n", resource.Name)
			continue
		}
		err := c.installResource(resource, basedir)
		if err != nil {
			return err
		}
		err = c.checkResource(resource)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) installResource(res v1alpha1.KabResource, basedir string) error {
	if res.Path == "" {
		return errors.New("cannot install anything other than a url yet")
	}
	fmt.Printf("installing %s from %s...", res.Name, res.Path)
	yaml, err := fileutils.Read(res.Path, basedir)
	if err != nil {
		return err
	}
	c.coreClient.RESTClient().Post()
	kubectl := kubectl.RealKubeCtl()
	istioLog, err := kubectl.ExecStdin([]string{"apply", "-f", "-"}, &yaml)
	if err != nil {
		fmt.Printf("%s\n", istioLog)
		if strings.Contains(istioLog, "forbidden") {
			fmt.Print(`It looks like you don't have cluster-admin permissions.

To fix this you need to:
 1. Delete the current failed installation using:
      ` + env.Cli.Name + ` system uninstall --istio --force
 2. Give the user account used for installation cluster-admin permissions, you can use the following command:
      kubectl create clusterrolebinding cluster-admin-binding \
        --clusterrole=cluster-admin \
        --user=<install-user>
 3. Re-install ` + env.Cli.Name + `

`)
		}
		return err
	}
	return nil
}

// TODO this only supports checking Pods for phases, add more resources
func (c *Client) checkResource(resource v1alpha1.KabResource) error {
	cnt := 1
	for _, check := range resource.Checks {
		var ready bool
		var err error
		for i := 0; i < 360; i++ {
			if strings.EqualFold(check.Kind, "Pod") {
				ready, err = c.isPodReady(check, resource.Namespace)
				if err != nil {
					return err
				}
				if ready {
					break
				}
			} else {
				return errors.New("only Kind:Pod supported for resource checks")
			}
			time.Sleep(1 * time.Second)
			cnt++
			if cnt%5 == 0 {
				fmt.Print(".")
			}
		}
		if !ready {
			return errors.New(fmt.Sprintf("The resource %s did not initialize", resource.Name))
		}
	}
	fmt.Println("done")
	return nil
}

func (c *Client) isPodReady(check v1alpha1.ResourceChecks, namespace string) (bool, error) {
	pods := c.coreClient.CoreV1().Pods(namespace)
	podList, err := pods.List(metav1.ListOptions{
		LabelSelector: convertMapToString(check.Selector.MatchLabels),
	})
	if err != nil {
		return false, err
	}
	for _, pod := range podList.Items {
		if strings.EqualFold(string(pod.Status.Phase), check.Pattern) {
			return true, nil
		}
	}
	return false, nil
}

func convertMapToString(m map[string]string) string {
	var s string
	for k, v := range m {
		s += k + "=" + v + ","
	}
	if last := len(s) - 1; last >= 0 && s[last] == ',' {
		s = s[:last]
	}
	return s
}
