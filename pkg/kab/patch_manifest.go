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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"cnab-k8s-installer-base/pkg/apis/kab/v1alpha1"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/apis/meta/v1"

	"net/url"
)

const MINIKUBE_NODE_NAME = "minikube"

func (c *Client) PatchManifest(manifest *v1alpha1.Manifest) error {

	err := manifest.PatchResourceContent(c.applyLabels)
	if err != nil {
		return err
	}

	err = manifest.PatchResourceContent(c.patchForMinikube)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) applyLabels(res *v1alpha1.KabResource) (content string, e error) {
	var path *url.URL
	var err error

	log.Tracef("Applying labels resource: %s Labels: %+v...", res.Name, res.Labels)

	path, err = url.Parse(res.Path)
	if err != nil {
		return "", err
	}
	byteContent, err := c.kustomizer.ApplyLabels(path, res.Labels)
	log.Traceln("done")

	return string(byteContent), nil
}

func (c *Client) patchForMinikube(res *v1alpha1.KabResource) (string, error) {
	_, err := c.coreClient.CoreV1().Nodes().Get(MINIKUBE_NODE_NAME, v1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return res.Content, nil
		}
		return "", err
	}
	byteContent := []byte(res.Content)
	byteContent = bytes.Replace(byteContent, []byte("type: LoadBalancer"), []byte("type: NodePort"), -1)
	return string(byteContent), nil
}
