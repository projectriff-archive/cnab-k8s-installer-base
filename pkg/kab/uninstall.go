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
	e "errors"
	"fmt"
	"strings"

	"github.com/projectriff/cnab-k8s-installer-base/pkg/apis/kab/v1alpha1"
	"github.com/projectriff/cnab-k8s-installer-base/pkg/scan"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Client) Uninstall(name string) error {
	var manifest *v1alpha1.Manifest
	var err error
	kindList := []string{}

	manifest, err = c.LookupManifest(name)
	if err != nil {
		return e.New(fmt.Sprintf("unable to lookup manifest: %v", err))
	}
	for _, resource := range manifest.Spec.Resources {
		kinds, err := scan.ListKindFromContent([]byte(resource.Content))
		if err != nil {
			return err
		}
		kindList = append(kindList, kinds...)
	}
	installationName := GetInstallationName()

	log.Infof("uninstalling %s...\n", installationName)

	label := LABEL_KEY_NAME + "=" + installationName

	log.Debugf("Issuing kubectl delete %s -l %s\n", strings.Join(kindList, ","), label)
	out, err := c.kubectl.Exec([]string{"delete", strings.Join(kindList, ","), "-l", label})
	log.Debugf(out)
	if err != nil {
		return e.New(fmt.Sprintf("error while uninstalling: %v, due to: %s", err, out))
	}

	log.Infoln("uninstalling bundle manifest from cluster")
	err = c.kabClient.ProjectriffV1alpha1().Manifests(manifest.Namespace).Delete(manifest.Name, &metav1.DeleteOptions{})
	if err != nil {
		return e.New(fmt.Sprintf("error while deleting the manifest: %v", err))
	}
	return nil
}

func (c *Client) LookupManifest(name string) (*v1alpha1.Manifest, error) {
	namespaceList, err := c.coreClient.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, ns := range namespaceList.Items {
		manifest, err := c.kabClient.ProjectriffV1alpha1().Manifests(ns.Name).Get(name, metav1.GetOptions{})
		if err != nil {
			if errors.IsNotFound(err) {
				continue
			}
			return nil, err
		}
		manifest.Namespace = ns.Name
		return manifest, nil
	}
	return nil, e.New(fmt.Sprintf("could not find manifest for installation name: %s", name))
}
