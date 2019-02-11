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

package v1alpha1

import metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type ResourceChecks struct {
	Kind     string               `json:"kind,omitempty"`
	Selector metaV1.LabelSelector `json:"selector,omitempty"`
	JsonPath string               `json:"jsonpath,omitempty"`
	Pattern  string               `json:"pattern,omitempty"`
}

type KabResource struct {
	Path      string           `json:"path,omitempty"`
	Content   string           `json:"content,omitempty"`
	Name      string           `json:"name,omitempty"`
	Namespace string           `json:"namespace,omitempty"`
	Install   bool             `json:"install,omitempty"`
	Checks    []ResourceChecks `json:"checks,omitempty"`
}

type KabSpec struct {
	Resources []KabResource `json:"resources,omitempty"`
}

type KabStatus struct {
	Status string `json:"status,omitempty"`
}

type Manifest struct {
	metaV1.TypeMeta   `json:",inline"`
	metaV1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KabSpec   `json:"spec,omitempty"`
	Status KabStatus `json:"status,omitempty"`
}
