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
	"errors"
	"fmt"
	"strings"

	"github.com/projectriff/cnab-k8s-installer-base/pkg/apis/kab/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TODO this only supports checking Pods for phases, add more resources
func (rm *rm) IsResourceReady(check v1alpha1.ResourceChecks) (bool, error) {
	n := strings.ToUpper(check.Kind)
	switch n {
	case "POD":
		return rm.isPodReady(check)
	}
	return false, errors.New(fmt.Sprintf("unknown resource kind: %s", check.Kind))
}

func (rm *rm) isPodReady(check v1alpha1.ResourceChecks) (bool, error) {
	pods := rm.coreClient.CoreV1().Pods(check.Namespace)
	podList, err := pods.List(metav1.ListOptions{
		LabelSelector: convertMapToString(check.Selector.MatchLabels),
	})
	if err != nil {
		return false, err
	}
	if len(podList.Items) == 0 {
		return false, nil
	}
	for _, pod := range podList.Items {
		if !strings.EqualFold(string(pod.Status.Phase), check.Pattern) {
			return false, nil
		}
	}
	return true, nil
}
