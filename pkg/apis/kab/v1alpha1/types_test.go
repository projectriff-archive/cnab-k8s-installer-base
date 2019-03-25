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
	"cnab-k8s-installer-base/pkg/apis/kab/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"path/filepath"
	"runtime"
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
})
