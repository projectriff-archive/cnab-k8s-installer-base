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

	"github.com/projectriff/cnab-k8s-installer-base/pkg/apis/kab/v1alpha1"
	log "github.com/sirupsen/logrus"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

func (c *Client) Install(manifest *v1alpha1.Manifest) error {
	err := CreateCRD(c.extClient)
	if err != nil {
		return errors.New(fmt.Sprintf("Could not create kab CRD: %s ", err))
	}
	manifest, err = c.CreateCRDObject(manifest, backOffSettings())
	if err != nil {
		return errors.New(fmt.Sprintf("Could not install riff: %s ", err))
	}
	log.Infoln("Installing bundle components")
	log.Infoln()
	err = c.installAndCheckResources(manifest)
	if err != nil {
		return errors.New(fmt.Sprintf("Could not install riff: %s ", err))
	}
	log.Infof("Kubernetes Application Bundle installed\n\n")
	return nil
}

func backOffSettings() wait.Backoff {
	return wait.Backoff{
		Duration: minRetryInterval,
		Factor:   exponentialBackoffBase,
		Steps:    maxRetries,
	}
}

func (c *Client) CreateCRDObject(manifest *v1alpha1.Manifest, backOffSettings wait.Backoff) (*v1alpha1.Manifest, error) {

	log.Debugln("creating object", manifest.Name)
	err := wait.ExponentialBackoff(backOffSettings, func() (bool, error) {
		old, err := c.kabClient.ProjectriffV1alpha1().Manifests().Get(manifest.Name, metav1.GetOptions{})
		if err != nil && !k8serr.IsNotFound(err) {
			log.Debugln("error looking up object", err)
			return false, nil
		}
		if !isEmpty(old) {
			return true, errors.New("bundle already installed")
		}
		_, err = c.kabClient.ProjectriffV1alpha1().Manifests().Create(manifest)
		if err != nil {
			log.Debugln("error creating object", err)
			return false, nil
		}
		return true, nil
	})
	if err == wait.ErrWaitTimeout {
		return nil, errors.New("timed out creating custom resource definition")
	}
	return manifest, err
}

func isEmpty(manifest *v1alpha1.Manifest) bool {
	if manifest == nil || len(manifest.Spec.Resources) == 0 {
		return true
	}
	return false
}

func (c *Client) installAndCheckResources(manifest *v1alpha1.Manifest) error {
	rm := NewResourceManager(c.kubectl, c.coreClient)
	for _, resource := range manifest.Spec.Resources {
		if resource.Deferred {
			log.Debugf("Skipping install of %s\n", resource.Name)
			continue
		}
		err := rm.Install(resource, backOffSettings())
		if err != nil {
			return err
		}
		err = rm.Check(resource, backOffSettings())
		if err != nil {
			return err
		}
	}
	return nil
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
