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

package v1alpha1_test

import (
	"errors"
	"path/filepath"
	"runtime"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/projectriff/cnab-k8s-installer-base/pkg/apis/kab/v1alpha1"
)

var _ = Describe("Manifest", func() {
	Describe("NewManifest", func() {

		var (
			manifestPath string
			manifest     *v1alpha1.Manifest
			err          error
		)

		JustBeforeEach(func() {
			manifest, err = v1alpha1.NewManifest(manifestPath)
		})

		Context("when an invalid path is provided", func() {
			BeforeEach(func() {
				manifestPath = ""
			})

			It("should return a suitable error", func() {
				Expect(err).To(MatchError(HavePrefix("error reading manifest file: ")))
			})
		})

		Context("when the file contains invalid YAML", func() {
			BeforeEach(func() {
				manifestPath = "./fixtures/invalid.yaml"
			})

			It("should return a suitable error", func() {
				Expect(err).To(MatchError(HavePrefix("error parsing manifest file: ")))
			})
		})

		Context("when the manifest has the wrong version", func() {
			BeforeEach(func() {
				manifestPath = "./fixtures/wrongversion.yaml"
			})

			It("should return a suitable error", func() {
				Expect(err).To(MatchError(ContainSubstring("Unsupported version")))
			})
		})

		Context("when the manifest contains a resource with an absolute path", func() {
			BeforeEach(func() {
				manifestPath = filepath.Join("fixtures")
				if runtime.GOOS == "windows" {
					manifestPath = filepath.Join(manifestPath, "absolutepath.windows.yaml")
				} else {
					manifestPath = filepath.Join(manifestPath, "absolutepath.yaml")
				}
			})

			It("should return a suitable error", func() {
				Expect(err).To(MatchError(ContainSubstring("resources must use a http or https URL or a relative path: absolute path not supported: ")))
			})
		})

		Context("when the manifest contains a resource with unsupported URL scheme", func() {
			BeforeEach(func() {
				manifestPath = "./fixtures/invalidscheme.yaml"
			})

			It("should return a suitable error", func() {
				Expect(err).To(MatchError(HavePrefix("resources must use a http or https URL or a relative path: scheme file not supported:")))
			})
		})

		Context("when the manifest is valid", func() {
			BeforeEach(func() {
				manifestPath = "./fixtures/valid.yaml"
			})

			It("should return with no error", func() {
				Expect(err).NotTo(HaveOccurred())
			})

			It("should parse the istio array", func() {
				Expect(manifest.Spec.Resources[0].Name).To(Equal("istio"))
			})

			It("should parse the Knative array", func() {
				releases := []string{}
				for _, res := range manifest.Spec.Resources {
					releases = append(releases, res.Name)
				}
				Expect(releases).To(ConsistOf("istio", "build", "eventing", "serving", "eventing-in-memory-channel", "riff-build-template", "riff-build-cache"))
			})

			It("should set the labels", func() {
				var res v1alpha1.KabResource
				for _, res = range manifest.Spec.Resources {
					if res.Name == "riff-build-template" {
						break
					}
				}
				Expect(res).ToNot(BeNil())
				Expect(res.Labels).To(HaveLen(2))
				Expect(res.Labels).Should(HaveKeyWithValue("key1", "value1"))
				Expect(res.Labels).Should(HaveKeyWithValue("key2", "value2"))

			})
		})
	})

	Describe("PatchResources", func() {

		var (
			manifestPath string
			manifest     *v1alpha1.Manifest
			err          error
		)
		JustBeforeEach(func() {
			manifestPath = "./fixtures/valid.yaml"
			manifest, err = v1alpha1.NewManifest(manifestPath)
		})

		Context("When the function returns modified content", func() {
			It("content of resource is updated", func() {
				err = manifest.PatchResourceContent(func(res *v1alpha1.KabResource) (string, error) {
					if res.Name == "istio" {
						return "my new content", nil
					}
					return res.Content, nil
				})
				Expect(err).ToNot(HaveOccurred())
				var istioRes v1alpha1.KabResource
				for _, res := range manifest.Spec.Resources {
					if res.Name == "istio" {
						istioRes = res
					}
				}
				Expect(istioRes.Content).To(Equal("my new content"))
			})
		})

		Context("When the function modifies attributes other than content", func() {
			It("those modifications are preserved", func() {
				lables := map[string]string{"k1": "v1", "k2": "v2"}
				err = manifest.PatchResourceContent(func(res *v1alpha1.KabResource) (string, error) {
					if res.Name == "istio" {
						res.Labels = lables
					}
					return res.Content, nil
				})
				Expect(err).ToNot(HaveOccurred())
				var istioRes v1alpha1.KabResource
				for _, res := range manifest.Spec.Resources {
					if res.Name == "istio" {
						istioRes = res
					}
				}
				Expect(istioRes.Labels).To(HaveLen(2))
				Expect(istioRes.Labels).To(HaveKeyWithValue("k1", "v1"))
				Expect(istioRes.Labels).To(HaveKeyWithValue("k2", "v2"))
			})
		})

		Context("When the function throws an error", func() {
			It("an error is returned", func() {
				err = manifest.PatchResourceContent(func(res *v1alpha1.KabResource) (string, error) {
					if res.Name == "istio" {
						return "", errors.New("my error")
					}
					return res.Content, nil
				})
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("my error"))
			})
		})
	})

	Describe("InlineContent", func() {

		var (
			manifestPath string
			manifest     *v1alpha1.Manifest
			err          error
		)

		Context("when the resource content is empty", func() {

			Context("When there is a valid path specified", func() {
				It("content of resource is updated", func() {
					manifestPath = "./fixtures/inline-mfst.yaml"
					manifest, err = v1alpha1.NewManifest(manifestPath)
					Expect(err).ToNot(HaveOccurred())
					err = manifest.InlineContent()
					Expect(err).ToNot(HaveOccurred())
					Expect(manifest.Spec.Resources[0].Content).To(ContainSubstring("test-ns"))
				})
			})

			Context("When there is an invalid path specified", func() {
				It("throws an exception", func() {
					manifestPath = "./fixtures/inline-invalid-mfst.yaml"
					manifest, err = v1alpha1.NewManifest(manifestPath)
					Expect(err).ToNot(HaveOccurred())
					err = manifest.InlineContent()
					Expect(err).To(HaveOccurred())
				})
			})
		})
		Context("when the resource content is not empty", func() {
			Context("when there is a valid path specified", func() {
				It("does not overwrite the resource content", func() {
					manifestPath = "./fixtures/inline-mfst-with-content.yaml"
					manifest, err = v1alpha1.NewManifest(manifestPath)
					Expect(err).ToNot(HaveOccurred())
					err = manifest.InlineContent()
					Expect(err).ToNot(HaveOccurred())
					Expect(manifest.Spec.Resources[0].Content).To(ContainSubstring("my content"))
				})
			})
		})
	})
})
