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
	"cnab-k8s-installer-base/pkg/kubectl"
	"errors"
	"fmt"
	"github.com/pivotal/go-ape/pkg/furl"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"strings"
)

type rm struct {
	kubectl    kubectl.KubeCtl
	coreClient kubernetes.Interface
}

type ResourceManager interface {
	Install(resource v1alpha1.KabResource, backOffSettings wait.Backoff) error
	Check(resource v1alpha1.KabResource, backOffSettings wait.Backoff) error
}

func NewResourceManager(kubectl kubectl.KubeCtl, coreClient kubernetes.Interface) *rm {
	return &rm{kubectl: kubectl, coreClient: coreClient}
}

func (rm *rm) Install(res v1alpha1.KabResource, backOffSettings wait.Backoff) error {
	var installContent []byte
	var err error

	log.Infof("installing %s...", res.Name)
	err = wait.ExponentialBackoff(backOffSettings, func() (bool, error) {
		if res.Content != "" {
			installContent = []byte(res.Content)
		} else {
			if res.Path == "" {
				return false, errors.New(fmt.Sprintf("resource %s does not specify Content OR Path to yaml for install", res.Name))
			}
			installContent, err = furl.Read(res.Path, "")
			if err != nil {
				log.Debugln("error reading", err)
				return false, err
			}

		}

		resLog, err := rm.kubectl.ExecStdin([]string{"apply", "-f", "-"}, &installContent)
		if err != nil {
			if strings.Contains(resLog, "forbidden") {
				log.Warningf(`It looks like you don't have cluster-admin permissions.

To fix this you need to:
 1. Delete the current failed installation.
 2. Give the user account used for installation cluster-admin permissions, you can use the following command:
      kubectl create clusterrolebinding cluster-admin-binding \
        --clusterrole=cluster-admin \
        --user=<install-user>
 3. Re-install the bundle

`)
				return false, err
			}
			log.Debugf("retrying installing resource: %s due to error %+v\n", res.Name, err)
			return false, nil
		}
		return true, nil
	})
	if err == wait.ErrWaitTimeout {
		return errors.New(fmt.Sprintf("could not create resource: %s", res.Name))
	}
	return err
}

func (rm *rm) Check(res v1alpha1.KabResource, backOffSettings wait.Backoff) error {
	for _, check := range res.Checks {
		err := wait.ExponentialBackoff(backOffSettings, func() (bool, error) {
			var ready bool
			var innerErr error
			ready, innerErr = rm.IsResourceReady(check, res.Namespace)
			if innerErr != nil {
				return false, innerErr
			}
			if !ready {
				return false, nil
			}
			return true, nil
		})
		if err == wait.ErrWaitTimeout {
			return errors.New(fmt.Sprintf("resource %s did not initialize", res.Name))
		}
		if err != nil {
			return err
		}
	}
	log.Infof("done installing %s", res.Name)
	return nil
}
