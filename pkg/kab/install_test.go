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
	"fmt"

	"cnab-k8s-installer-base/pkg/apis/kab/v1alpha1"
	"cnab-k8s-installer-base/pkg/client/clientset/versioned/fake"
	"cnab-k8s-installer-base/pkg/kab"
	"cnab-k8s-installer-base/pkg/kab/vendor_mocks"
	mockkustomize "cnab-k8s-installer-base/pkg/kustomize/mocks"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/projectriff/riff/pkg/env"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
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
		//mockCore = new(vendor_mocks.CoreV1Interface)
		fakeKabClient = fake.NewSimpleClientset()
		//mockNodes = new(vendor_mocks.NodeInterface)
		mockKustomize = new(mockkustomize.Kustomizer)

		client = kab.NewKnbClient(kubeClient, nil, fakeKabClient, nil, mockKustomize)
	})

	Describe("test CreateCrdObject()", func() {

		It("allows only one crd object to be created", func() {
			manifest = &v1alpha1.Manifest{
				ObjectMeta: v1.ObjectMeta{
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
			Expect(err).To(MatchError(fmt.Sprintf("%s already installed", env.Cli.Name)))
		})

		It("retries if the crd is not ready", func() {
			manifest = &v1alpha1.Manifest{
				ObjectMeta: v1.ObjectMeta{
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
			//oldManifest := manifest.DeepCopy()
			fakeKabClient.PrependReactor("get", "*", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
				return true, nil, errors.NewNotFound(schema.GroupResource{}, "test")
			})
			invocations := 1
			fakeKabClient.PrependReactor("create", "*", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
				invocations++
				return true, nil, errors.NewNotFound(schema.GroupResource{}, "*")
			})
			_, err = client.CreateCRDObject(manifest, wait.Backoff{Steps: 2})
			Expect(err).To(MatchError(fmt.Sprintf("timed out creating %s custom resource defiition", env.Cli.Name)))
			Expect(invocations).To(BeNumerically(">", 1))
		})
	})
})
