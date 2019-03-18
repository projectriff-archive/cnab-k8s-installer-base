package kab

import (
	"cnab-k8s-installer-base/pkg/apis/kab/v1alpha1"
	"errors"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

// TODO this only supports checking Pods for phases, add more resources
func (c *Client) IsResourceReady(check v1alpha1.ResourceChecks, namespace string) (bool, error) {
	n := strings.ToUpper(check.Kind)
	switch n {
	case "POD":
		return c.isPodReady(check, namespace)
	}
	return false, errors.New(fmt.Sprintf("unknown resource kind: %s", check.Kind))
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