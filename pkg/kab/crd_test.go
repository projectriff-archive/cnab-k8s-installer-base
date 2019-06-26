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
	"github.com/projectriff/cnab-k8s-installer-base/pkg/kab"
	"github.com/projectriff/cnab-k8s-installer-base/pkg/kab/vendor_mocks/ext"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var _ = Describe("create CRD", func() {

	var (
		mockExtensionClientSet *vendor_mocks_ext.Interface
		mockExtensionInterface *vendor_mocks_ext.ApiextensionsV1beta1Interface
		mockCrdi               *vendor_mocks_ext.CustomResourceDefinitionInterface
		err                    error
	)

	JustBeforeEach(func() {
		mockExtensionClientSet = new(vendor_mocks_ext.Interface)
		mockExtensionInterface = new(vendor_mocks_ext.ApiextensionsV1beta1Interface)
		mockCrdi = new(vendor_mocks_ext.CustomResourceDefinitionInterface)
	})

	Context("when crd creation returns an error from api server", func() {
		It("the exception is now swallowed", func() {
			mockExtensionClientSet.On("ApiextensionsV1beta1").Return(mockExtensionInterface)
			mockExtensionInterface.On("CustomResourceDefinitions").Return(mockCrdi)
			mockCrdi.On("Create", mock.AnythingOfType("*v1beta1.CustomResourceDefinition")).Return(nil, errors.NewUnauthorized("unknown"))
			err = kab.CreateCRD(mockExtensionClientSet)
			Expect(err).ToNot(BeNil())
		})
	})

	Context("when crd already exists", func() {
		It("an exception is not returned", func() {
			mockExtensionClientSet.On("ApiextensionsV1beta1").Return(mockExtensionInterface)
			mockExtensionInterface.On("CustomResourceDefinitions").Return(mockCrdi)
			mockCrdi.On("Create", mock.AnythingOfType("*v1beta1.CustomResourceDefinition")).Return(nil, errors.NewAlreadyExists(schema.GroupResource{}, "manifests"))
			err = kab.CreateCRD(mockExtensionClientSet)
			Expect(err).To(BeNil())
		})
	})

})
