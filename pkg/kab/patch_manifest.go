/*
 * Copyright 2019 The original author or authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package kab

import (
	"bytes"
	"os"
	"strconv"

	"github.com/projectriff/cnab-k8s-installer-base/pkg/apis/kab/v1alpha1"
	log "github.com/sirupsen/logrus"
)

const (
	NODE_PORT_ENV_VAR              = "NODE_PORT"
	CNAB_INSTALLATION_NAME_ENV_VAR = "CNAB_INSTALLATION_NAME"
	LABEL_KEY_NAME                 = "cnab-k8s-installer-installation-name"
)

func (c *Client) PatchManifest(manifest *v1alpha1.Manifest) error {

	err := manifest.PatchResourceContent(c.applyLabels)
	if err != nil {
		return err
	}

	err = manifest.PatchResourceContent(c.patchForNodePort)
	if err != nil {
		return err
	}

	setName(manifest)

	return nil
}

func setName(manifest *v1alpha1.Manifest) {
	installName := GetInstallationName()
	if installName != "" {
		manifest.Name = installName
	}
}

func GetInstallationName() string {
	installName := os.Getenv(CNAB_INSTALLATION_NAME_ENV_VAR)
	return installName
}

func (c *Client) applyLabels(res *v1alpha1.KabResource) (content string, e error) {

	labels := addLabels(res.Labels)
	res.Labels = labels

	log.Tracef("Applying labels resource: %s Labels: %+v...", res.Name, res.Labels)

	byteContent, err := c.kustomizer.ApplyLabels(res.Content, res.Labels)
	if err != nil {
		return "", err
	}

	log.Traceln("done")

	return string(byteContent), nil
}

func addLabels(labels map[string]string) map[string]string {
	if labels == nil {
		labels = map[string]string{}
	}
	labels[LABEL_KEY_NAME] = GetInstallationName()
	return labels
}

func (c *Client) patchForNodePort(res *v1alpha1.KabResource) (string, error) {
	var err error
	nodePort, err := isNodePortSet()
	if err != nil {
		return "", err
	}

	if nodePort {
		byteContent := []byte(res.Content)
		byteContent = bytes.Replace(byteContent, []byte("type: LoadBalancer"), []byte("type: NodePort"), -1)
		return string(byteContent), nil
	}
	return res.Content, nil
}

func isNodePortSet() (bool, error) {
	nodePort := os.Getenv(NODE_PORT_ENV_VAR)
	if nodePort == "" {
		return false, nil
	}
	retVal, err := strconv.ParseBool(nodePort)
	if err != nil {
		return false, err
	}
	return retVal, nil
}
