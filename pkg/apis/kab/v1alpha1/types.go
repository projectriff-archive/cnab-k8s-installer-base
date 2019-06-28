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

package v1alpha1

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/pivotal/go-ape/pkg/furl"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Manifest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec KabSpec `json:"spec,omitempty"`

	// +optional
	Status KabStatus `json:"status,omitempty"`
}

type ResourceChecks struct {
	Kind      string               `json:"kind,omitempty"`
	Namespace string               `json:"namespace,omitempty"`
	Selector  metav1.LabelSelector `json:"selector,omitempty"`
	JsonPath  string               `json:"jsonpath,omitempty"`
	Pattern   string               `json:"pattern,omitempty"`
}

type KabResource struct {
	Path     string            `json:"path,omitempty"`
	Content  string            `json:"content,omitempty"`
	Name     string            `json:"name,omitempty"`
	Labels   map[string]string `json:"labels,omitempty"`
	Deferred bool              `json:"deferred,omitempty"`
	Checks   []ResourceChecks  `json:"checks,omitempty"`
}

type KabSpec struct {
	Resources []KabResource `json:"resources,omitempty"`
}

type KabStatus struct {
	Status string `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ManifestList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `son:"metadata,omitempty"`

	Items []Manifest `json:"items"`
}

func NewManifest(path string) (manifest *Manifest, err error) {
	var m Manifest
	yamlFile, err := furl.Read(path, "")
	if err != nil {
		return nil, fmt.Errorf("error reading manifest file: %v", err)
	}

	err = yaml.Unmarshal(yamlFile, &m)
	if err != nil {
		if strings.Contains(err.Error(), "did not find expected key") {
			return nil, fmt.Errorf("error parsing manifest file: %v. Please ensure that manifest has supported version", err)
		}
		return nil, fmt.Errorf("error parsing manifest file: %v", err)
	}

	supportedVersion := fmt.Sprintf("%s/%s", GroupName, VersionNumber)
	if !strings.EqualFold(m.APIVersion, supportedVersion) {
		return nil, errors.New(fmt.Sprintf("Unsupported version %s. Supported version is %s", m.APIVersion, supportedVersion))
	}

	err = m.VisitResources(checkResourcePath)
	if err != nil {
		return nil, err
	}

	return &m, nil
}

func (m *Manifest) VisitResources(f func(res KabResource) error) error {

	for _, resource := range m.Spec.Resources {
		err := f(resource)
		if err != nil {
			return err
		}
	}
	return nil
}

// For each resource (res) in the manifest, replaces the Content (res.Content) with
// the returned value from the parameter function
func (m *Manifest) PatchResourceContent(f func(res *KabResource) (string, error)) error {
	var resource *KabResource

	for i := 0; i < len(m.Spec.Resources); i++ {
		resource = &m.Spec.Resources[i]

		newContent, err := f(resource)
		if err != nil {
			return err
		}
		resource.Content = newContent

	}
	return nil
}

func checkResourcePath(resource KabResource) error {
	if filepath.IsAbs(resource.Path) {
		return fmt.Errorf("resources must use a http or https URL or a relative path: absolute path not supported: %v", resource)
	}

	u, err := url.Parse(resource.Path)
	if err != nil {
		return err
	}
	if u.Scheme == "http" || u.Scheme == "https" || (u.Scheme == "" && !filepath.IsAbs(u.Path)) {
		return nil
	}

	if u.Scheme == "" {
		return fmt.Errorf("resources must use a http or https URL or a relative path: absolute path not supported: %v", resource)
	}

	return fmt.Errorf("resources must use a http or https URL or a relative path: scheme %s not supported: %v", u.Scheme, resource)
}
