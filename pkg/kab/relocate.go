package kab

import (
	"cnab-k8s-installer-base/pkg/apis/kab/v1alpha1"
	"cnab-k8s-installer-base/pkg/fileutils"
	"cnab-k8s-installer-base/pkg/scan"
	"errors"
	"strings"
)

func (c *Client) Relocate(manifest *v1alpha1.Manifest, targetRegistry string) error {
	if targetRegistry == "" {
		return nil
	}

	err := RelocateManifest(manifest, targetRegistry)
	if err != nil {
		return err
	}

	manifest, err = c.kabClient.ProjectriffV1alpha1().Manifests(manifest.Namespace).Update(manifest)
	if err != nil {
		return err
	}

	return nil
}

func RelocateManifest(manifest *v1alpha1.Manifest, targetRegistry string) error {
	var err error

	err = embedResourceContent(manifest)
	if err != nil {
		return err
	}

	images, err := getAllImages(manifest)
	if err != nil {
		return err
	}
	relocatedImages, err := getRelocatedImages(targetRegistry, images)
	if err != nil {
		return err
	}
	// TODO add a images section to the manifest

	err = replaceImagesInManifest(manifest, images, relocatedImages)

	return nil
}

func replaceImagesInManifest(manifest *v1alpha1.Manifest, images []string, relocatedImages []string) error {
	replacer, err := buildImageReplacer(images, relocatedImages)
	if err != nil {
		return err
	}
	for i := 0; i < len(manifest.Spec.Resources); i++ {
		resource := &manifest.Spec.Resources[i]
		resource.Content = replacer.Replace(resource.Content)
	}
	return nil
}

func buildImageReplacer(images, relocatedImages []string) (*strings.Replacer, error) {
	replacements := []string{}

	if len(images) != len(relocatedImages) {
		return nil, errors.New("length of images and relocated images should be same")
	}
	for i := range images {
		replacements = append(replacements, images[i], relocatedImages[i])
	}
	return strings.NewReplacer(replacements...), nil
}

func embedResourceContent(manifest *v1alpha1.Manifest) error {

	for i := 0; i < len(manifest.Spec.Resources); i++ {
		resource := manifest.Spec.Resources[i]
		content, err := fileutils.Read(resource.Path, "")
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
	relocatedImages := []string{}
	if !strings.HasSuffix(targetRegistry, "/") {
		targetRegistry = targetRegistry + "/"
	}
	for _, img := range images {
		_, repoPath, err := splitHostAndRepo(img)
		if err != nil {
			return nil, err
		}
		repoPath = strings.ReplaceAll(repoPath, "/", "-")
		relocatedImg := targetRegistry + repoPath
		relocatedImages = append(relocatedImages, relocatedImg)
	}
	return relocatedImages, nil
}

func splitHostAndRepo(image string) (host string, repoPath string, err error) {
	s := strings.SplitN(image, "/", 2)
	if len(s) == 1 {
		return "", s[0], nil
	}
	return s[0], s[1], nil
}

func getAllImages(manifest *v1alpha1.Manifest) ([]string, error) {
	images := []string{}

	err := manifest.VisitResources(func(res v1alpha1.KabResource) error {
		tmpImgs, err := scan.ListImages(res.Name, res.Content, "")
		if err != nil {
			return err
		}
		images = append(images, tmpImgs...)
		return nil
	})

	return images, err
}
