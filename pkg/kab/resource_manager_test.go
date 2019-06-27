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
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/projectriff/cnab-k8s-installer-base/pkg/apis/kab/v1alpha1"
	"github.com/projectriff/cnab-k8s-installer-base/pkg/kab"
	"github.com/projectriff/cnab-k8s-installer-base/pkg/kab/vendor_mocks"
	mockkubectl "github.com/projectriff/cnab-k8s-installer-base/pkg/kubectl/mocks"
	"github.com/stretchr/testify/mock"
	v12 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

var _ = Describe("ResourceManager Tests", func() {
	Describe("Install Tests", func() {

		var (
			mockKubeCtl     *mockkubectl.KubeCtl
			backoffSettings wait.Backoff
			err             error
		)

		BeforeEach(func() {
			backoffSettings = wait.Backoff{Steps: 2}
		})

		AfterEach(func() {
			mockKubeCtl.AssertExpectations(GinkgoT())
		})

		Context("When resource has a string in content field", func() {
			It("resource content is passed to kubectl", func() {
				mockKubeCtl = new(mockkubectl.KubeCtl)

				resMan := kab.NewResourceManager(mockKubeCtl, nil)
				contentString := `spec:
  template:
    spec:
      containers:
        - -builder
        - cluster
        - INFO
      image: gcr.io/knative-releases/x/y
      name: build-webhook`

				resource := v1alpha1.KabResource{
					Content: contentString,
				}

				byteContent := []byte(contentString)
				mockKubeCtl.On("ExecStdin", []string{"apply", "-f", "-"},
					&byteContent).Return("success", nil)

				err = resMan.Install(resource, backoffSettings)
				Expect(err).To(BeNil())
			})
		})

		Context("When resource has an invalid path", func() {
			It("an exception is returned", func() {
				mockKubeCtl = new(mockkubectl.KubeCtl)

				resMan := kab.NewResourceManager(mockKubeCtl, nil)

				resource := v1alpha1.KabResource{
					Path: "fixtures/invalid.yaml",
				}

				err = resMan.Install(resource, backoffSettings)
				Expect(err).ToNot(BeNil())
			})
		})

		Context("When resource has a valid path", func() {
			It("resource content is passed to kubectl", func() {
				mockKubeCtl = new(mockkubectl.KubeCtl)

				resMan := kab.NewResourceManager(mockKubeCtl, nil)
				contentString := `---
spec:
  template:
    spec:
      containers:
        - -builder
        - cluster
        - INFO
      image: gcr.io/knative-releases/x/y
      name: build-webhook
---
spec:
  containers:
    - image: mysql:5.6
  name: mysql
`
				resource := v1alpha1.KabResource{
					Path: "fixtures/test-resource.yaml",
				}

				byteContent := []byte(contentString)
				mockKubeCtl.On("ExecStdin", []string{"apply", "-f", "-"},
					&byteContent).Return("success", nil)

				err = resMan.Install(resource, backoffSettings)
				Expect(err).To(BeNil())
			})
		})

		Context("there is an permission error while doing kubectl apply", func() {
			It("the error is returned", func() {
				mockKubeCtl = new(mockkubectl.KubeCtl)

				resMan := kab.NewResourceManager(mockKubeCtl, nil)
				contentString := "foo"

				resource := v1alpha1.KabResource{
					Content: contentString,
				}

				byteContent := []byte(contentString)
				mockKubeCtl.On("ExecStdin", []string{"apply", "-f", "-"},
					&byteContent).Return("forbidden", errors.New("some error"))

				err = resMan.Install(resource, backoffSettings)
				Expect(err).ToNot(BeNil())
				Expect(err).To(MatchError("some error"))
			})
		})

		Context("there is an error while doing kubectl apply", func() {
			It("the operation is retried", func() {
				mockKubeCtl = new(mockkubectl.KubeCtl)

				resMan := kab.NewResourceManager(mockKubeCtl, nil)
				contentString := "foo"

				resource := v1alpha1.KabResource{
					Content: contentString,
				}

				byteContent := []byte(contentString)
				mockKubeCtl.On("ExecStdin", []string{"apply", "-f", "-"},
					&byteContent).Return("error", errors.New("error")).Twice()

				err = resMan.Install(resource, backoffSettings)
				Expect(err).ToNot(BeNil())
			})
		})

		Context("resource has neither content nor path", func() {
			It("an error is thrown", func() {
				mockKubeCtl = new(mockkubectl.KubeCtl)

				resMan := kab.NewResourceManager(mockKubeCtl, nil)

				resource := v1alpha1.KabResource{
					Name: "e1",
				}

				err = resMan.Install(resource, backoffSettings)
				Expect(err).ToNot(BeNil())
				Expect(err).To(MatchError("resource e1 does not specify Content OR Path to yaml for install"))
			})
		})
	})

	Describe("Check Tests", func() {
		var (
			backoffSettings wait.Backoff
			mockKubeClient  *vendor_mocks.Interface
			mockCore        *vendor_mocks.CoreV1Interface
			mockPods        *vendor_mocks.PodInterface
			err             error
		)

		BeforeEach(func() {
			backoffSettings = wait.Backoff{Steps: 2}
			mockKubeClient = new(vendor_mocks.Interface)
			mockCore = new(vendor_mocks.CoreV1Interface)
			mockPods = new(vendor_mocks.PodInterface)
		})

		AfterEach(func() {
			mockCore.AssertExpectations(GinkgoT())
			mockPods.AssertExpectations(GinkgoT())
		})

		Context("unknown resource type is passed", func() {
			It("an error for unsupported type is thrown", func() {
				resMan := kab.NewResourceManager(nil, nil)

				resource := v1alpha1.KabResource{
					Name: "r1",
					Checks: []v1alpha1.ResourceChecks{
						{
							Kind: "ingress",
						},
					},
				}

				err = resMan.Check(resource, backoffSettings)
				Expect(err).ToNot(BeNil())
				Expect(err).To(MatchError("unknown resource kind: ingress"))
			})
		})
		Context("when a pod resource type is used", func() {
			Context("When the pod is not found", func() {
				It("the operation is retried", func() {
					mockKubeClient.On("CoreV1").Return(mockCore)
					mockCore.On("Pods", mock.Anything).Return(mockPods)
					mockPods.On("List", mock.Anything).Return(&v12.PodList{}, nil).Twice()

					resMan := kab.NewResourceManager(nil, mockKubeClient)
					labelSelector := v1.LabelSelector{
						MatchLabels: map[string]string{"istio": "sidecar-injector"},
					}
					resource := v1alpha1.KabResource{
						Name: "r1",
						Checks: []v1alpha1.ResourceChecks{
							{
								Kind:     "Pod",
								JsonPath: ".status.phase",
								Selector: labelSelector,
							},
						},
					}
					err = resMan.Check(resource, backoffSettings)
					Expect(err).ToNot(BeNil())
					Expect(err).To(MatchError("resource r1 did not initialize"))
				})
			})
			Context("When the pod is found but the status is not the desired status", func() {
				It("the check fails", func() {
					mockKubeClient.On("CoreV1").Return(mockCore)
					mockCore.On("Pods", mock.Anything).Return(mockPods)
					mockPods.On("List", mock.Anything).Return(&v12.PodList{
						Items: []v12.Pod{
							{
								ObjectMeta: v1.ObjectMeta{
									Labels: map[string]string{"istio": "sidecar-injector"},
								},
								Status: v12.PodStatus{Phase: "ImagePullBackoff"},
							},
						},
					}, nil).Twice()

					resMan := kab.NewResourceManager(nil, mockKubeClient)
					labelSelector := v1.LabelSelector{
						MatchLabels: map[string]string{"istio": "sidecar-injector"},
					}
					resource := v1alpha1.KabResource{
						Name: "r1",
						Checks: []v1alpha1.ResourceChecks{
							{
								Kind:     "Pod",
								JsonPath: ".status.phase",
								Selector: labelSelector,
								Pattern:  "Running",
							},
						},
					}
					err = resMan.Check(resource, backoffSettings)
					Expect(err).ToNot(BeNil())
					Expect(err).To(MatchError("resource r1 did not initialize"))
				})
			})

			Context("When multiple pods match the label but one of them fails", func() {
				It("the check fails", func() {
					mockKubeClient.On("CoreV1").Return(mockCore)
					mockCore.On("Pods", mock.Anything).Return(mockPods)
					mockPods.On("List", mock.Anything).Return(&v12.PodList{
						Items: []v12.Pod{
							{
								ObjectMeta: v1.ObjectMeta{
									Labels: map[string]string{"istio": "sidecar-injector"},
								},
								Status: v12.PodStatus{Phase: "Running"},
							},
							{
								ObjectMeta: v1.ObjectMeta{
									Labels: map[string]string{"foo": "bar"},
								},
								Status: v12.PodStatus{Phase: "ImagePullBackoff"},
							},
						},
					}, nil)

					resMan := kab.NewResourceManager(nil, mockKubeClient)
					labelSelectorIstio := v1.LabelSelector{
						MatchLabels: map[string]string{"istio": "sidecar-injector"},
					}

					resource := v1alpha1.KabResource{
						Name: "r1",
						Checks: []v1alpha1.ResourceChecks{
							{
								Kind:     "Pod",
								JsonPath: ".status.phase",
								Selector: labelSelectorIstio,
								Pattern:  "Running",
							},
						},
					}
					err = resMan.Check(resource, backoffSettings)
					Expect(err).ToNot(BeNil())
					Expect(err).To(MatchError("resource r1 did not initialize"))
				})
			})

			Context("With multiple resource checks, when one resource check succeeds but the other fails", func() {
				It("the check fails", func() {
					labelSelectorIstio := v1.LabelSelector{
						MatchLabels: map[string]string{"istio": "sidecar-injector"},
					}
					labelSelectorfoo := v1.LabelSelector{
						MatchLabels: map[string]string{"foo": "bar"},
					}

					mockKubeClient.On("CoreV1").Return(mockCore)
					mockCore.On("Pods", mock.Anything).Return(mockPods)

					mockPods.On("List", v1.ListOptions{LabelSelector: "istio=sidecar-injector"}).Return(&v12.PodList{
						Items: []v12.Pod{
							{
								ObjectMeta: v1.ObjectMeta{
									Labels: map[string]string{"istio": "sidecar-injector"},
								},
								Status: v12.PodStatus{Phase: "Running"},
							},
						},
					}, nil).Once()

					mockPods.On("List", v1.ListOptions{LabelSelector: "foo=bar"}).Return(&v12.PodList{
						Items: []v12.Pod{
							{
								ObjectMeta: v1.ObjectMeta{
									Labels: map[string]string{"foo": "bar"},
								},
								Status: v12.PodStatus{Phase: "ImagePullBackoff"},
							},
						},
					}, nil).Twice()

					resMan := kab.NewResourceManager(nil, mockKubeClient)

					resource := v1alpha1.KabResource{
						Name: "r1",
						Checks: []v1alpha1.ResourceChecks{
							{
								Kind:     "Pod",
								JsonPath: ".status.phase",
								Selector: labelSelectorIstio,
								Pattern:  "Running",
							},
							{
								Kind:     "Pod",
								JsonPath: ".status.phase",
								Selector: labelSelectorfoo,
								Pattern:  "Running",
							},
						},
					}
					err = resMan.Check(resource, backoffSettings)
					Expect(err).ToNot(BeNil())
					Expect(err).To(MatchError("resource r1 did not initialize"))
				})
			})

			Context("When the pod is found and the status is the desired status", func() {
				It("the check succeeds", func() {
					mockKubeClient.On("CoreV1").Return(mockCore)
					mockCore.On("Pods", mock.Anything).Return(mockPods)
					mockPods.On("List", mock.Anything).Return(&v12.PodList{
						Items: []v12.Pod{
							{
								ObjectMeta: v1.ObjectMeta{
									Labels: map[string]string{"istio": "sidecar-injector"},
								},
								Status: v12.PodStatus{Phase: "Running"},
							},
						},
					}, nil)

					resMan := kab.NewResourceManager(nil, mockKubeClient)
					labelSelector := v1.LabelSelector{
						MatchLabels: map[string]string{"istio": "sidecar-injector"},
					}
					resource := v1alpha1.KabResource{
						Name: "r1",
						Checks: []v1alpha1.ResourceChecks{
							{
								Kind:     "Pod",
								JsonPath: ".status.phase",
								Selector: labelSelector,
								Pattern:  "Running",
							},
						},
					}
					err = resMan.Check(resource, backoffSettings)
					Expect(err).To(BeNil())
				})
			})
		})
	})
})
