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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/projectriff/cnab-k8s-installer-base/pkg/apis/kab/v1alpha1"
	"github.com/projectriff/cnab-k8s-installer-base/pkg/client/clientset/versioned/fake"
	"github.com/projectriff/cnab-k8s-installer-base/pkg/kab"
	"github.com/projectriff/cnab-k8s-installer-base/pkg/kab/vendor_mocks"
	vendor_mocks_ext "github.com/projectriff/cnab-k8s-installer-base/pkg/kab/vendor_mocks/ext"
	mockkubectl "github.com/projectriff/cnab-k8s-installer-base/pkg/kubectl/mocks"
	mockkustomize "github.com/projectriff/cnab-k8s-installer-base/pkg/kustomize/mocks"
	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/testing"
)

var _ = Describe("test install", func() {

	var (
		client        *kab.Client
		kubeClient    *vendor_mocks.Interface
		fakeKabClient *fake.Clientset
		mockKustomize *mockkustomize.Kustomizer
		manifest      *v1alpha1.Manifest
		err           error
	)

	BeforeEach(func() {
		kubeClient = new(vendor_mocks.Interface)
		fakeKabClient = fake.NewSimpleClientset()
		mockKustomize = new(mockkustomize.Kustomizer)

		client = kab.NewKnbClient(kubeClient, nil, fakeKabClient, nil, mockKustomize, nil)
	})

	Describe("test CreateCrdObject()", func() {

		It("allows only one crd object to be created", func() {
			manifest = &v1alpha1.Manifest{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: v1alpha1.KabSpec{
					Resources: []v1alpha1.KabResource{
						{
							Name: "foo",
							Path: "http://some.resource.com",
						},
					},
				},
			}
			oldManifest := manifest.DeepCopy()
			fakeKabClient.PrependReactor("get", "*", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
				return true, oldManifest, nil
			})
			_, err = client.CreateCRDObject(manifest, wait.Backoff{Steps: 2})
			Expect(err).To(MatchError("bundle already installed"))
		})

		It("retries if the crd is not ready", func() {
			manifest = &v1alpha1.Manifest{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: v1alpha1.KabSpec{
					Resources: []v1alpha1.KabResource{
						{
							Name: "foo",
							Path: "http://some.resource.com",
						},
					},
				},
			}
			fakeKabClient.PrependReactor("get", "*", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
				return true, nil, errors.NewNotFound(schema.GroupResource{}, "test")
			})
			invocations := 1
			fakeKabClient.PrependReactor("create", "*", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
				invocations++
				return true, nil, errors.NewNotFound(schema.GroupResource{}, "*")
			})
			_, err = client.CreateCRDObject(manifest, wait.Backoff{Steps: 2})
			Expect(err).To(MatchError("timed out creating custom resource definition"))
			Expect(invocations).To(BeNumerically(">", 1))
		})
	})

	Describe("test install", func() {
		var (
			mockExtensionClientSet *vendor_mocks_ext.Interface
			mockExtensionInterface *vendor_mocks_ext.ApiextensionsV1beta1Interface
			mockCrdi               *vendor_mocks_ext.CustomResourceDefinitionInterface
			mockKubectl            *mockkubectl.KubeCtl
			err                    error
		)

		JustBeforeEach(func() {
			mockExtensionClientSet = new(vendor_mocks_ext.Interface)
			mockExtensionInterface = new(vendor_mocks_ext.ApiextensionsV1beta1Interface)
			mockCrdi = new(vendor_mocks_ext.CustomResourceDefinitionInterface)
			mockKubectl = new(mockkubectl.KubeCtl)
			fakeKabClient = fake.NewSimpleClientset()
			mockExtensionClientSet.On("ApiextensionsV1beta1").Return(mockExtensionInterface)
			mockExtensionInterface.On("CustomResourceDefinitions").Return(mockCrdi)
			mockCrdi.On("Create", mock.Anything).Return(nil, nil).Once()
			fakeKabClient.PrependReactor("get", "*", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
				return true, nil, nil
			})
			fakeKabClient.PrependReactor("create", "*", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
				return true, nil, nil
			})
			client = kab.NewKnbClient(kubeClient, mockExtensionClientSet, fakeKabClient, nil, nil, mockKubectl)
		})

		JustAfterEach(func() {
			mockCrdi.AssertExpectations(GinkgoT())
		})

		Context("when Install is called", func() {
			It("The crd and crd object are created", func() {
				manifest = &v1alpha1.Manifest{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test",
					},
					Spec: v1alpha1.KabSpec{
						Resources: []v1alpha1.KabResource{},
					},
				}
				err = client.Install(manifest)
				Expect(err).To(BeNil())
				Expect(len(fakeKabClient.Actions())).To(Equal(2))
				Expect(fakeKabClient.Actions()[0].GetVerb()).To(Equal("get"))
				Expect(fakeKabClient.Actions()[1].GetVerb()).To(Equal("create"))
			})
		})
		Context("when manifest has a deferred resource", func() {
			It("the resource is not installed", func() {
				manifest = &v1alpha1.Manifest{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test",
					},
					Spec: v1alpha1.KabSpec{
						Resources: []v1alpha1.KabResource{
							{
								Name:     "deferred",
								Deferred: true,
							},
						},
					},
				}
				err = client.Install(manifest)
				Expect(err).To(BeNil())
				Expect(len(mockKubectl.Calls)).To(Equal(0))
			})
		})
	})
})
