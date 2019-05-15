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

package kab_test

import (
	"cnab-k8s-installer-base/pkg/apis/kab/v1alpha1"
	"cnab-k8s-installer-base/pkg/kab"
	"cnab-k8s-installer-base/pkg/kab/vendor_mocks"
	"cnab-k8s-installer-base/pkg/registry/mocks"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal/image-relocation/pkg/image"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("RelocateManifest", func() {

	var (
		client             *kab.Client
		kubeClient         *vendor_mocks.Interface
		mockRegistryClient *mockregistry.Client
		manifest           *v1alpha1.Manifest
		destRepo           string
		err                error
	)

	BeforeEach(func() {
		kubeClient = new(vendor_mocks.Interface)
		mockRegistryClient = new(mockregistry.Client)

		client = kab.NewKnbClient(kubeClient, nil, nil, mockRegistryClient, nil)
	})

	BeforeEach(func() {
		manifest = &v1alpha1.Manifest{
			Spec: v1alpha1.KabSpec{
				Resources: []v1alpha1.KabResource{
					{
						Content: "spec:\n" +
							"  template:\n" +
							"    spec:\n" +
							"      containers:\n" +
							"      - -builder\n" +
							"      - cluster\n" +
							"      - INFO\n" +
							"      image: gcr.io/knative-releases/x/y\n" +
							"      name: build-webhook\n",
					},
					{
						Content: "spec:\n" +
							"  containers:\n" +
							"  - image: mysql:5.6\n" +
							"  name: mysql",
					},
				},
			},
		}
		destRepo = "my.private.repo"
	})

	Context("when destination registry is empty", func() {
		It("the call is a no-op", func() {
			oldManifest := manifest.DeepCopy()
			err = client.Relocate(manifest, "")
			Expect(err).To(BeNil())
			Expect(manifest).To(Equal(oldManifest))
		})
	})
	Context("when the manifest has Spec.Resource.Content", func() {
		It("repository relocation is successful", func() {
			imgRef := destRepo + "/istio-proxyv2-f93a2cacc6cafa0474a2d6990a4dd1a0:1.0.7@sha256:9c6663cddbc984e88c27530d8acac7dca83070c4ad6d2570604cc4fff6c36a7a"
			img, _ := image.NewName(imgRef)
			mockRegistryClient.On("Relocate", mock.Anything, mock.Anything).Return(img, nil)

			err = client.Relocate(manifest, destRepo)
			Expect(err).To(BeNil())

			for i := range manifest.Spec.Resources {
				Expect(manifest.Spec.Resources[i].Content).NotTo(ContainSubstring("docker.io"))
				Expect(manifest.Spec.Resources[i].Content).NotTo(ContainSubstring("gcr.io"))
				Expect(manifest.Spec.Resources[i].Content).To(ContainSubstring(destRepo))
			}
		})
	})

})
