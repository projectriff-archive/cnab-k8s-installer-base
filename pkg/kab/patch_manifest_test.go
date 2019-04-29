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

package kab_test

import (
	"cnab-k8s-installer-base/pkg/apis/kab/v1alpha1"
	"cnab-k8s-installer-base/pkg/kab"
	"cnab-k8s-installer-base/pkg/kab/vendor_mocks"
	"cnab-k8s-installer-base/pkg/kustomize/mocks"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var _ = Describe("test patch for minikube", func() {

	var (
		client        *kab.Client
		kubeClient    *vendor_mocks.Interface
		mockCore      *vendor_mocks.CoreV1Interface
		mockNodes     *vendor_mocks.NodeInterface
		mockKustomize *mockkustomize.Kustomizer
		manifest      *v1alpha1.Manifest
		content       string
		err           error
	)

	BeforeEach(func() {
		kubeClient = new(vendor_mocks.Interface)
		mockCore = new(vendor_mocks.CoreV1Interface)
		mockNodes = new(vendor_mocks.NodeInterface)
		mockKustomize = new(mockkustomize.Kustomizer)
		content = "sometext: type: LoadBalancer"

		client = kab.NewKnbClient(kubeClient, nil, nil, nil, mockKustomize)
	})

	Context("When the node is minikube", func() {

		JustBeforeEach(func() {
			content = "sometext: type: LoadBalancer"

			kubeClient.On("CoreV1").Return(mockCore)
			mockCore.On("Nodes").Return(mockNodes)
			mockNodes.On("Get", "minikube", mock.Anything).Return(nil, nil)

			mockNodes.On("Get", "docker-for-desktop", mock.Anything).Return(nil,
				errors.NewNotFound(schema.GroupResource{}, "docker-for-desktop"))

			mockKustomize.On("ApplyLabels", mock.Anything, mock.Anything).Return([]byte(content), nil)
		})

		It("the content is patched", func() {
			manifest = &v1alpha1.Manifest{
				Spec: v1alpha1.KabSpec{
					Resources: []v1alpha1.KabResource{
						{
							Name:    "foo",
							Content: content,
						},
					},
				},
			}
			err = client.PatchManifest(manifest)
			Expect(err).To(BeNil())
			Expect(manifest.Spec.Resources[0].Content).ToNot(ContainSubstring("type: LoadBalancer"))
			Expect(manifest.Spec.Resources[0].Content).To(ContainSubstring("type: NodePort"))
		})
	})

	Context("When the node is docker-for-desktop", func() {

		JustBeforeEach(func() {
			content = "sometext: type: LoadBalancer"

			kubeClient.On("CoreV1").Return(mockCore)
			mockCore.On("Nodes").Return(mockNodes)
			mockNodes.On("Get", "docker-for-desktop", mock.Anything).Return(nil, nil)

			mockNodes.On("Get", "minikube", mock.Anything).Return(nil,
				errors.NewNotFound(schema.GroupResource{}, "minikube"))

			mockKustomize.On("ApplyLabels", mock.Anything, mock.Anything).Return([]byte(content), nil)
		})

		It("the content is patched", func() {
			manifest = &v1alpha1.Manifest{
				Spec: v1alpha1.KabSpec{
					Resources: []v1alpha1.KabResource{
						{
							Name:    "foo",
							Content: content,
						},
					},
				},
			}
			err = client.PatchManifest(manifest)
			Expect(err).To(BeNil())
			Expect(manifest.Spec.Resources[0].Content).ToNot(ContainSubstring("type: LoadBalancer"))
			Expect(manifest.Spec.Resources[0].Content).To(ContainSubstring("type: NodePort"))
		})
	})

	Context("When the node is neither minikube nor docker-for-desktop", func() {

		JustBeforeEach(func() {
			kubeClient.On("CoreV1").Return(mockCore)
			mockCore.On("Nodes").Return(mockNodes)

			mockNodes.On("Get", "minikube", mock.Anything).Return(nil,
				errors.NewNotFound(schema.GroupResource{}, "minikube"))

			mockNodes.On("Get", "docker-for-desktop", mock.Anything).Return(nil,
				errors.NewNotFound(schema.GroupResource{}, "docker-for-desktop"))

			mockKustomize.On("ApplyLabels", mock.Anything, mock.Anything).Return([]byte(content), nil)
		})

		It("the content is not patched", func() {
			manifest = &v1alpha1.Manifest{
				Spec: v1alpha1.KabSpec{
					Resources: []v1alpha1.KabResource{
						{
							Name:    "foo",
							Content: content,
						},
					},
				},
			}
			err = client.PatchManifest(manifest)
			Expect(err).To(BeNil())
			Expect(manifest.Spec.Resources[0].Content).To(ContainSubstring("type: LoadBalancer"))
			Expect(manifest.Spec.Resources[0].Content).ToNot(ContainSubstring("type: NodePort"))
		})
	})

	Context("When there is an error getting nodes", func() {

		JustBeforeEach(func() {
			kubeClient.On("CoreV1").Return(mockCore)
			mockCore.On("Nodes").Return(mockNodes)

			mockNodes.On("Get", "minikube", mock.Anything).Return(nil,
				errors.NewForbidden(schema.GroupResource{}, "", err))

			mockKustomize.On("ApplyLabels", mock.Anything, mock.Anything).Return([]byte(content), nil)
		})

		It("an error is returned", func() {
			manifest = &v1alpha1.Manifest{
				Spec: v1alpha1.KabSpec{
					Resources: []v1alpha1.KabResource{
						{
							Name:    "foo",
							Content: content,
						},
					},
				},
			}
			err = client.PatchManifest(manifest)
			Expect(err).ToNot(BeNil())
		})
	})
})
