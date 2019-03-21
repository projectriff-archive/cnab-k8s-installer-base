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