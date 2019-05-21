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
	"strings"

	"cnab-k8s-installer-base/pkg/apis/kab/v1alpha1"
	"cnab-k8s-installer-base/pkg/scan"
	"github.com/pivotal/go-ape/pkg/furl"
	"github.com/pivotal/image-relocation/pkg/image"
	"github.com/pivotal/image-relocation/pkg/pathmapping"
	log "github.com/sirupsen/logrus"

)

func (c *Client) Relocate(manifest *v1alpha1.Manifest, targetRegistry string) error {
	if targetRegistry == "" {
		log.Traceln("skipping image relocation")
		return nil
	}

	var err error

	err = embedResourceContent(manifest)
	if err != nil {
		return err
	}

	relocationMap, err := buildRelocationImageMap(manifest, targetRegistry)

	err = c.updateRegistry(relocationMap)
	if err != nil {
		return err
	}
	// TODO add a images section to the manifest

	err = replaceImagesInManifest(manifest, relocationMap)

	return nil
}

// pull images, push them to the target registry and update the relocationMap with the
// newly pushed digested images
func (c *Client) updateRegistry(relocationMap map[string]string) error {
	log.Infoln("Relocating images...")

	for fromRef, toRef := range relocationMap {
		digestedRef, err := c.registryClient.Relocate(fromRef, toRef)
		if err != nil {
			return err
		}
		relocationMap[fromRef] = digestedRef.String()
	}
	log.Infoln("finished relocating images")

	return nil
}

func buildRelocationImageMap(manifest *v1alpha1.Manifest, targetRegistry string) (map[string]string, error) {
	relocationMap := map[string]string{}
	images, err := getAllImages(manifest)
	if err != nil {
		return nil, err
	}
	relocatedImages, err := getRelocatedImages(targetRegistry, images)
	if err != nil {
		return nil, err
	}
	if len(images) != len(relocatedImages) {
		return nil, errors.New("length of images and relocated images should be same")
	}
	log.Traceln("Relocation Image Map:")
	for i, fromRef := range images {
		relocationMap[fromRef] = relocatedImages[i]
		log.Traceln(fromRef, " : ", relocatedImages[i])
	}

	return relocationMap, nil
}

func replaceImagesInManifest(manifest *v1alpha1.Manifest, relocationMap map[string]string) error {
	replacer, err := buildImageReplacer(relocationMap)
	if err != nil {
		return err
	}
	for i := 0; i < len(manifest.Spec.Resources); i++ {
		resource := &manifest.Spec.Resources[i]
		resource.Content = replacer.Replace(resource.Content)
	}
	return nil
}

func buildImageReplacer(relocationMap map[string]string) (*strings.Replacer, error) {
	replacements := []string{}

	log.Traceln("building image replacements")
	for key, value := range relocationMap {
		replacements = append(replacements, key, value)
		log.Traceln(key, ":", value)
	}
	log.Traceln("done building image replacements")

	return strings.NewReplacer(replacements...), nil
}

func embedResourceContent(manifest *v1alpha1.Manifest) error {

	for i := 0; i < len(manifest.Spec.Resources); i++ {
		resource := &manifest.Spec.Resources[i]
		if resource.Path == "" {
			continue
		}
		content, err := furl.Read(resource.Path, "")
		if err != nil {
			return err
		}
		strContent := string(content)
		if strContent != "" {
			resource.Content = strContent
		}
	}
	return nil
}

func getRelocatedImages(targetRegistry string, images []string) ([]string, error) {
	mapping := getMapping(targetRegistry)
	relocatedImages := []string{}
	for _, img := range images {
		relocatedImg, err := mapping(img)
		if err != nil {
			return []string{}, err
		}
		relocatedImages = append(relocatedImages, relocatedImg)
	}
	return relocatedImages, nil
}

func getMapping(repoPrefix string) func(string) (string, error) {
	return func(originalImage string) (string, error) {
		n, err := image.NewName(originalImage)
		if err != nil {
			return "", err
		}
		return pathmapping.FlattenRepoPathPreserveTagDigest(repoPrefix, n).String(), nil
	}
}

func getAllImages(manifest *v1alpha1.Manifest) ([]string, error) {
	images := []string{}

	err := manifest.VisitResources(func(res v1alpha1.KabResource) error {
		tmpImgs, err := scan.ListImagesFromContent([]byte(res.Content))
		if err != nil {
			return err
		}
		images = append(images, tmpImgs...)
		return nil
	})

	return images, err
}
