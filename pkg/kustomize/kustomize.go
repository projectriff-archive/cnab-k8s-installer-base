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

package kustomize

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/url"

	"github.com/pivotal/go-ape/pkg/furl"
	"sigs.k8s.io/kustomize/k8sdeps"
	"sigs.k8s.io/kustomize/pkg/commands/build"
	"sigs.k8s.io/kustomize/pkg/fs"
	"sort"
	"strings"
	"time"
)

type Kustomizer interface {
	// Applies the provided labels to the provided remote or local resource definition
	// Returns the customized resource contents
	// Returns an error if
	// - the URL scheme is not supported (only file, http and https are)
	// - retrieving the content fails
	// - applying the customization fails
	// As of the current implementation, it is not safe to call this function concurrently
	ApplyLabels(resourceUri *url.URL, labels map[string]string) ([]byte, error)
}

type kustomizer struct {
	fakeDir     string
	fs          fs.FileSystem
	httpTimeout time.Duration
}

func MakeKustomizer(timeout time.Duration) Kustomizer {
	return &kustomizer{
		fs:          fs.MakeFakeFS(), // keep contents in-memory
		fakeDir:     "/",
		httpTimeout: timeout,
	}
}

func (kust *kustomizer) ApplyLabels(resourceUri *url.URL, labels map[string]string) ([]byte, error) {
	resourcePath, err := kust.writeResourceFile(resourceUri)
	if err != nil {
		return nil, err
	}
	err = kust.writeKustomizationFile(resourcePath, labels)
	if err != nil {
		return nil, err
	}
	return kust.runBuild()
}

func (kust *kustomizer) writeResourceFile(resourceUri *url.URL) (string, error) {
	resourceContents, err := furl.ReadUrl(resourceUri, kust.httpTimeout)
	if err != nil {
		return "", err
	}
	resourcePath := "resource.yaml"
	err = kust.fs.WriteFile(kust.fakeDir+resourcePath, []byte(resourceContents))
	if err != nil {
		return "", err
	}
	return resourcePath, nil
}

func (kust *kustomizer) writeKustomizationFile(resourcePath string, labels map[string]string) error {
	err := kust.fs.WriteFile(kust.fakeDir+"kustomization.yaml", []byte(fmt.Sprintf(`
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
commonLabels:%s
resources:
  - %s
`, formatLabels(labels), resourcePath)))
	if err != nil {
		return err
	}
	return nil
}

func (kust *kustomizer) runBuild() ([]byte, error) {
	var out bytes.Buffer
	kustomizeFactory := k8sdeps.NewFactory()
	kustomizeBuildCommand := build.NewCmdBuild(&out, kust.fs, kustomizeFactory.ResmapF, kustomizeFactory.TransformerF)
	kustomizeBuildCommand.SetArgs([]string{kust.fakeDir})
	kustomizeBuildCommand.SetOutput(ioutil.Discard)
	_, err := kustomizeBuildCommand.ExecuteC()
	if err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func formatLabels(labels map[string]string) string {
	builder := strings.Builder{}
	keys := keysOf(labels)
	sort.Strings(keys)
	for _, key := range keys {
		builder.WriteString(fmt.Sprintf("\n  %s: %s", key, labels[key]))
	}
	return builder.String()
}

func keysOf(dict map[string]string) []string {
	result := make([]string, len(dict))
	i := 0
	for k := range dict {
		result[i] = k
		i++
	}
	return result
}
