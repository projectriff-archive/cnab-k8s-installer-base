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
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/projectriff/cnab-k8s-installer-base/pkg/apis/kab/v1alpha1"
	"github.com/projectriff/cnab-k8s-installer-base/pkg/kab/vendor_mocks"
)

var _ = Describe("RelocateManifest", func() {
	var (
		client     *Client
		kubeClient *vendor_mocks.Interface
		manifest   *v1alpha1.Manifest
		destRepo   string
		err        error
	)

	Describe("When Resource Content is inlined", func() {
		BeforeEach(func() {
			relocationMappingMountPoint = "fixtures/relocation-mapping.json"
			kubeClient = new(vendor_mocks.Interface)

			client = NewKnbClient(kubeClient, nil, nil, nil, nil)
			manifest = &v1alpha1.Manifest{
				Spec: v1alpha1.KabSpec{
					Resources: []v1alpha1.KabResource{
						{
							Content: `spec:
  template:
    spec:
      containers:
      - -builder
      - cluster
      - INFO
      image: gcr.io/knative-releases/x/y
      name: build-webhook`,
						},
						{
							Content: `spec:
  containers:
  - image: mysql:5.6
  name: mysql`,
						},
					},
				},
			}
			destRepo = "my.private.repo"
		})
		Context("when there is no relocation mapping", func() {
			BeforeEach(func() {
				relocationMappingMountPoint = "/no/such/file"
			})
			It("the call is a no-op", func() {
				oldManifest := manifest.DeepCopy()
				err = client.MaybeRelocate(manifest)
				Expect(err).To(BeNil())
				Expect(manifest).To(Equal(oldManifest))
			})
		})
		Context("when the relocation mapping file cannot be accessed", func() {
			BeforeEach(func() {
				var err error
				relocationMappingMountPoint, err = ioutil.TempDir("", "relocate-test")
				Expect(err).NotTo(HaveOccurred())
			})
			AfterEach(func() {
				os.RemoveAll(relocationMappingMountPoint)
			})
			It("should return a suitable error", func() {
				err = client.MaybeRelocate(manifest)
				Expect(err).To(MatchError(HavePrefix("failed to read relocation mapping from ")))
			})
		})
		Context("when the relocation mapping file content is malformed", func() {
			BeforeEach(func() {
				relocationMappingMountPoint = "relocate_test.go"
			})
			It("should return a suitable error", func() {
				err = client.MaybeRelocate(manifest)
				Expect(err).To(MatchError(HavePrefix("failed to unmarshal relocation mapping: ")))
			})
		})
		Context("when the manifest has Spec.Resource.Content", func() {
			It("repository relocation is successful", func() {
				err = client.MaybeRelocate(manifest)
				Expect(err).To(BeNil())

				for i := range manifest.Spec.Resources {
					Expect(manifest.Spec.Resources[i].Content).NotTo(ContainSubstring("mysql:5.6"))
					Expect(manifest.Spec.Resources[i].Content).NotTo(ContainSubstring("gcr.io/knative-releases"))
					Expect(manifest.Spec.Resources[i].Content).To(ContainSubstring(destRepo))
				}
			})
		})
	})

	Describe("When resource is a path", func() {
		BeforeEach(func() {
			kubeClient = new(vendor_mocks.Interface)

			client = NewKnbClient(kubeClient, nil, nil, nil, nil)
			content, err := ioutil.ReadFile("fixtures/test-resource.yaml")
			Expect(err).To(BeNil())

			manifest = &v1alpha1.Manifest{
				Spec: v1alpha1.KabSpec{
					Resources: []v1alpha1.KabResource{
						{
							Content: string(content),
						},
					},
				},
			}
			destRepo = "my.private.repo"
		})
		Context("When the resource file path is valid", func() {
			It("repository relocation is successful", func() {
				err = client.MaybeRelocate(manifest)
				Expect(err).To(BeNil())

				for i := range manifest.Spec.Resources {
					Expect(manifest.Spec.Resources[i].Content).NotTo(ContainSubstring("mysql:5.6"))
					Expect(manifest.Spec.Resources[i].Content).NotTo(ContainSubstring("gcr.io/knative-releases"))
					Expect(manifest.Spec.Resources[i].Content).To(ContainSubstring(destRepo))
				}
			})
		})
	})
})
