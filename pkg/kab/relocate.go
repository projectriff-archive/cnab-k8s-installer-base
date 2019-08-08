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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/projectriff/cnab-k8s-installer-base/pkg/apis/kab/v1alpha1"
	log "github.com/sirupsen/logrus"
)

const standardRelocationMappingMountPoint = "/cnab/app/relocation-mapping.json"

var relocationMappingMountPoint = standardRelocationMappingMountPoint

// MaybeRelocate relocates the given manifest if a relocation mapping is present and otherwise does not modify the manifest.
func (c *Client) MaybeRelocate(manifest *v1alpha1.Manifest) error {
	relocationMap, err := getRelocationMapping()
	if err != nil {
		return err
	}
	// If there is no relocation mapping, return without modifying the manifest.
	if relocationMap == nil {
		return nil
	}

	replaceImagesInManifest(manifest, relocationMap)

	return nil
}

func getRelocationMapping() (map[string]string, error) {
	if _, err := os.Stat(relocationMappingMountPoint); os.IsNotExist(err) {
		return nil, nil
	}

	relMap := make(map[string]string)
	relMapBytes, err := ioutil.ReadFile(relocationMappingMountPoint)
	if err != nil {
		return nil, fmt.Errorf("failed to read relocation mapping from %s: %v", relocationMappingMountPoint, err)
	}

	if err = json.Unmarshal(relMapBytes, &relMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal relocation mapping: %v", err)
	}

	return relMap, nil
}

func replaceImagesInManifest(manifest *v1alpha1.Manifest, relocationMap map[string]string) {
	replacer := buildImageReplacer(relocationMap)
	for i := 0; i < len(manifest.Spec.Resources); i++ {
		resource := &manifest.Spec.Resources[i]
		resource.Content = replacer.Replace(resource.Content)
	}
	return
}

func buildImageReplacer(relocationMap map[string]string) *strings.Replacer {
	replacements := []string{}

	log.Traceln("building image replacements")
	for key, value := range relocationMap {
		replacements = append(replacements, key, value)
		log.Traceln(key, ":", value)
	}
	log.Traceln("done building image replacements")

	return strings.NewReplacer(replacements...)
}
