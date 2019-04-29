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
	"cnab-k8s-installer-base/pkg/kab"
	"cnab-k8s-installer-base/pkg/kab/vendor_mocks/ext"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("CRD", func() {
	Describe("CRD definition", func() {

		var (
			mockExtensionClientSet *vendor_mocks_ext.Interface
			mockExtensionInterface *vendor_mocks_ext.ApiextensionsV1beta1Interface
			mockCrdi            *vendor_mocks_ext.CustomResourceDefinitionInterface
			err                 error
		)

		JustBeforeEach(func() {
			mockExtensionClientSet = new(vendor_mocks_ext.Interface)
			mockExtensionInterface = new(vendor_mocks_ext.ApiextensionsV1beta1Interface)
			mockCrdi = new(vendor_mocks_ext.CustomResourceDefinitionInterface)

			mockExtensionClientSet.On("ApiextensionsV1beta1").Return(mockExtensionInterface)
			mockExtensionInterface.On("CustomResourceDefinitions").Return(mockCrdi)
			mockCrdi.On("Create", mock.AnythingOfType("*v1beta1.CustomResourceDefinition")).Return(nil, errors.New("AlreadyExists"))

		})

		Context("when crd create has already been created", func() {
			It("does not throw an exception", func() {
				err = kab.CreateCRD(mockExtensionClientSet)
				Expect(err).To(Not(BeNil()))
			})
		})
	})
})
