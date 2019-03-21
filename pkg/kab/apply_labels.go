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
	"cnab-k8s-installer-base/pkg/apis/kab/v1alpha1"
	"net/url"
)

func (c *Client) ApplyLabels(manifest *v1alpha1.Manifest) error {
	var resource *v1alpha1.KabResource
	var content []byte
	var path *url.URL
	var err error
	for i := 0; i < len(manifest.Spec.Resources); i++ {
		resource = &manifest.Spec.Resources[i]
		path, err = url.Parse(resource.Path)
		if err != nil {
			return err
		}
		content, err = c.kustomizer.ApplyLabels(path, resource.Labels)
		resource.Content = string(content)

	}
	return nil
}