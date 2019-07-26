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
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/projectriff/cnab-k8s-installer-base/pkg/apis/kab/v1alpha1"
	"github.com/projectriff/cnab-k8s-installer-base/pkg/kab"
	"github.com/projectriff/cnab-k8s-installer-base/pkg/kab/vendor_mocks"
	mockkustomize "github.com/projectriff/cnab-k8s-installer-base/pkg/kustomize/mocks"
	"github.com/stretchr/testify/mock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("test patching manifest", func() {
	Describe("patch the manifest content", func() {

		var (
			client        *kab.Client
			kubeClient    *vendor_mocks.Interface
			mockKustomize *mockkustomize.Kustomizer
			manifest      *v1alpha1.Manifest
			content       string
			err           error
		)

		BeforeEach(func() {
			kubeClient = new(vendor_mocks.Interface)
			mockKustomize = new(mockkustomize.Kustomizer)
			content = "sometext: type: LoadBalancer"

			client = kab.NewKnbClient(kubeClient, nil, nil, mockKustomize, nil)
		})

		JustAfterEach(func() {
			kubeClient.AssertExpectations(GinkgoT())
			mockKustomize.AssertExpectations(GinkgoT())
		})

		Context("When the node-port env variable is set", func() {

			JustBeforeEach(func() {
				os.Setenv(kab.NODE_PORT_ENV_VAR, "true")
				content = "sometext: type: LoadBalancer"

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

			JustAfterEach(func() {
				os.Unsetenv(kab.NODE_PORT_ENV_VAR)
			})
		})

		Context("When the node is neither minikube nor docker-for-desktop", func() {

			JustBeforeEach(func() {
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
	})

	Describe("patch manifest name", func() {
		var (
			client        *kab.Client
			kubeClient    *vendor_mocks.Interface
			mockKustomize *mockkustomize.Kustomizer
			manifest      *v1alpha1.Manifest
			installName   string
			err           error
		)

		BeforeEach(func() {
			kubeClient = new(vendor_mocks.Interface)
			mockKustomize = new(mockkustomize.Kustomizer)
			installName = "myInstallation"
			mockKustomize.On("ApplyLabels", mock.Anything, mock.Anything).Return([]byte(""), nil)

			client = kab.NewKnbClient(kubeClient, nil, nil, mockKustomize, nil)
		})

		JustAfterEach(func() {
			kubeClient.AssertExpectations(GinkgoT())
			mockKustomize.AssertExpectations(GinkgoT())
		})

		Context("when the installation name is specified in env var", func() {
			JustBeforeEach(func() {
				os.Setenv("CNAB_INSTALLATION_NAME", installName)
			})
			JustAfterEach(func() {
				os.Unsetenv("CNAB_INSTALLATION_NAME")
			})
			It("the installation name is used for manifest", func() {
				manifest = &v1alpha1.Manifest{
					ObjectMeta: metav1.ObjectMeta{
						Name: "defaultName",
					},
					Spec: v1alpha1.KabSpec{
						Resources: []v1alpha1.KabResource{
							{
								Name: "foo",
							},
						},
					},
				}
				err = client.PatchManifest(manifest)
				Expect(err).To(BeNil())
				Expect(manifest.Name).To(Equal(installName))
			})
		})
		Context("when the installation name is not specified in env var", func() {
			It("the installation name remains unchanged", func() {
				manifest = &v1alpha1.Manifest{
					ObjectMeta: metav1.ObjectMeta{
						Name: "defaultName",
					},
					Spec: v1alpha1.KabSpec{
						Resources: []v1alpha1.KabResource{
							{
								Name: "foo",
							},
						},
					},
				}
				err = client.PatchManifest(manifest)
				Expect(err).To(BeNil())
				Expect(manifest.Name).To(Equal("defaultName"))
			})
		})
	})

	Describe("applying installation label", func() {
		var (
			client        *kab.Client
			kubeClient    *vendor_mocks.Interface
			mockKustomize *mockkustomize.Kustomizer
			manifest      *v1alpha1.Manifest
			err           error
		)

		BeforeEach(func() {
			kubeClient = new(vendor_mocks.Interface)
			mockKustomize = new(mockkustomize.Kustomizer)
			mockKustomize.On("ApplyLabels", mock.Anything, mock.Anything).Return([]byte(""), nil)

			client = kab.NewKnbClient(kubeClient, nil, nil, mockKustomize, nil)
		})
		JustBeforeEach(func() {
			os.Setenv(kab.CNAB_INSTALLATION_NAME_ENV_VAR, "myInstallation")
		})
		JustAfterEach(func() {
			os.Unsetenv(kab.CNAB_INSTALLATION_NAME_ENV_VAR)
			kubeClient.AssertExpectations(GinkgoT())
			mockKustomize.AssertExpectations(GinkgoT())
		})

		Context("when resource does not have labels", func() {
			It("the installation label is the only label applied", func() {
				manifest = &v1alpha1.Manifest{
					Spec: v1alpha1.KabSpec{
						Resources: []v1alpha1.KabResource{
							{
								Name: "foo",
							},
						},
					},
				}
				err = client.PatchManifest(manifest)
				Expect(err).To(BeNil())
				Expect(len(manifest.Spec.Resources[0].Labels)).To(Equal(1))
				Expect(manifest.Spec.Resources[0].Labels).Should(HaveKeyWithValue(kab.LABEL_KEY_NAME, "myInstallation"))
			})
		})
		Context("when resource has labels", func() {
			It("the installation label is appended to the existing labels", func() {
				manifest = &v1alpha1.Manifest{
					Spec: v1alpha1.KabSpec{
						Resources: []v1alpha1.KabResource{
							{
								Name:   "foo",
								Labels: map[string]string{"k1": "v1"},
							},
						},
					},
				}
				err = client.PatchManifest(manifest)
				Expect(err).To(BeNil())
				Expect(len(manifest.Spec.Resources[0].Labels)).To(Equal(2))
				Expect(manifest.Spec.Resources[0].Labels).Should(HaveKeyWithValue("k1", "v1"))
				Expect(manifest.Spec.Resources[0].Labels).Should(HaveKeyWithValue(kab.LABEL_KEY_NAME, "myInstallation"))
			})
		})
	})
})
