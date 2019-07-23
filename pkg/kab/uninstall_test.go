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
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/projectriff/cnab-k8s-installer-base/pkg/apis/kab/v1alpha1"
	"github.com/projectriff/cnab-k8s-installer-base/pkg/client/clientset/versioned/fake"
	"github.com/projectriff/cnab-k8s-installer-base/pkg/kab"
	"github.com/projectriff/cnab-k8s-installer-base/pkg/kab/vendor_mocks"
	mockkubectl "github.com/projectriff/cnab-k8s-installer-base/pkg/kubectl/mocks"
	"github.com/stretchr/testify/mock"
	v12 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/testing"
)

var _ = Describe("LookupManifest Tests", func() {

	var (
		client         *kab.Client
		mockCore       *vendor_mocks.CoreV1Interface
		mockKubeClient *vendor_mocks.Interface
		mockNamespace  *vendor_mocks.NamespaceInterface
		fakeKabClient  *fake.Clientset
		err            error
	)

	BeforeEach(func() {
		mockKubeClient = new(vendor_mocks.Interface)
		mockCore = new(vendor_mocks.CoreV1Interface)
		mockNamespace = new(vendor_mocks.NamespaceInterface)
		fakeKabClient = fake.NewSimpleClientset()
		mockKubeClient.On("CoreV1").Return(mockCore)
		mockCore.On("Namespaces").Return(mockNamespace)

		client = kab.NewKnbClient(mockKubeClient, nil, fakeKabClient, nil, nil)
	})

	Context("When there is error listing manifests", func() {
		It("the error is returned to the caller", func() {
			mockNamespace.On("List", mock.Anything).Return(nil, errors.NewUnauthorized("test error"))
			_, err = client.LookupManifest("myInstallation")
			Expect(err).To(MatchError("test error"))
		})
	})
	Context("When the manifest does not exist in the cluster", func() {
		It("throws an exception", func() {

			mockNamespace.On("List", mock.Anything).Return(&v12.NamespaceList{
				Items: []v12.Namespace{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "my-ns",
						},
					},
				},
			}, nil)
			fakeKabClient.PrependReactor("get", "*", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
				return true, nil, errors.NewNotFound(schema.GroupResource{}, "none")
			})
			_, err = client.LookupManifest("myInstallation")
			Expect(err).To(MatchError("could not find manifest for installation name: myInstallation"))
		})
	})
	Context("When there is more than one namespaces", func() {
		It("all namespaces are looked up to find a manifest", func() {
			manifest := &v1alpha1.Manifest{}
			mockNamespace.On("List", mock.Anything).Return(&v12.NamespaceList{
				Items: []v12.Namespace{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "my-ns1",
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "my-ns2",
						},
					},
				},
			}, nil)
			invocations := 0
			fakeKabClient.PrependReactor("get", "*", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
				invocations++
				if invocations == 1 {
					return true, nil, errors.NewNotFound(schema.GroupResource{}, "none")
				}
				return true, manifest, nil
			})
			m, err := client.LookupManifest("myInstallation")
			Expect(err).To(BeNil())
			Expect(m).Should(Equal(manifest))
			Expect(invocations).Should(Equal(2))
		})
	})
})

var _ = Describe("LookupManifest Tests", func() {

	var (
		client           *kab.Client
		mockCore         *vendor_mocks.CoreV1Interface
		mockKubeClient   *vendor_mocks.Interface
		mockNamespace    *vendor_mocks.NamespaceInterface
		fakeKabClient    *fake.Clientset
		mockKubectl      *mockkubectl.KubeCtl
		installationName string
		err              error
	)

	BeforeEach(func() {
		mockKubeClient = new(vendor_mocks.Interface)
		mockCore = new(vendor_mocks.CoreV1Interface)
		mockNamespace = new(vendor_mocks.NamespaceInterface)
		fakeKabClient = fake.NewSimpleClientset()
		mockKubectl = new(mockkubectl.KubeCtl)

		installationName = "myInstall"
		mockKubeClient.On("CoreV1").Return(mockCore)
		mockCore.On("Namespaces").Return(mockNamespace)

		client = kab.NewKnbClient(mockKubeClient, nil, fakeKabClient, nil, mockKubectl)
	})

	JustBeforeEach(func() {
		os.Setenv(kab.CNAB_INSTALLATION_NAME_ENV_VAR, installationName)
	})
	JustAfterEach(func() {
		os.Unsetenv(kab.CNAB_INSTALLATION_NAME_ENV_VAR)
	})

	Context("When manifest does not exist", func() {
		It("uninstall fails", func() {
			mockNamespace.On("List", mock.Anything).Return(&v12.NamespaceList{
				Items: []v12.Namespace{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "my-ns",
						},
					},
				},
			}, nil)
			fakeKabClient.PrependReactor("get", "*", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
				return true, nil, errors.NewNotFound(schema.GroupResource{}, "none")
			})
			err = client.Uninstall("myInstallation")
			Expect(err).To(MatchError("unable to lookup manifest: could not find manifest for installation name: myInstallation"))
		})
	})
	Context("When a valid manifest exists", func() {
		JustBeforeEach(func() {
			manifest := &v1alpha1.Manifest{
				Spec: v1alpha1.KabSpec{
					Resources: []v1alpha1.KabResource{
						{
							Name: "res1",
							Content: `---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: build-controller
  namespace: knative-build
---
apiVersion: v1
kind: Namespace
metadata:
  name: knative-build
---
`,
						},
					},
				},
			}
			mockNamespace.On("List", mock.Anything).Return(&v12.NamespaceList{
				Items: []v12.Namespace{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "my-ns1",
						},
					},
				},
			}, nil)
			fakeKabClient.PrependReactor("get", "*", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
				return true, manifest, nil
			})
		})
		Context("when there is an error while deleting using kubectl", func() {
			It("the error is returned to the caller", func() {
				fakeKabClient.PrependReactor("delete", "*", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
					return true, nil, nil
				})
				mockKubectl.On("Exec", []string{"delete", "Namespace,ServiceAccount", "-l",
					kab.LABEL_KEY_NAME + "=" + installationName}).Return("error", errors.NewUnauthorized("test error"))
				err = client.Uninstall(installationName)
				Expect(err.Error()).To(HavePrefix("error while uninstalling: test error"))
			})
		})
		Context("when there is an error deleting the manifest", func() {
			It("the error is returned to the caller", func() {
				fakeKabClient.PrependReactor("delete", "*", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
					return true, nil, errors.NewUnauthorized("test error")
				})
				mockKubectl.On("Exec", []string{"delete", "Namespace,ServiceAccount", "-l",
					kab.LABEL_KEY_NAME + "=" + installationName}).Return("success", nil)

				err = client.Uninstall(installationName)
				Expect(err.Error()).To(HavePrefix("error while deleting the manifest: test error"))
			})
		})
		Context("when there are no errors", func() {
			It("uninstall succeeds", func() {
				fakeKabClient.PrependReactor("delete", "*", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
					return true, nil, nil
				})
				mockKubectl.On("Exec", []string{"delete", "Namespace,ServiceAccount", "-l",
					kab.LABEL_KEY_NAME + "=" + installationName}).Return("success", nil)

				err = client.Uninstall(installationName)
				Expect(err).To(BeNil())
			})
		})
	})
})
