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
	"cnab-k8s-installer-base/pkg/apis/kab/v1alpha1"
	"fmt"
	log "github.com/sirupsen/logrus"
	extApi "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	extClientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const NAME = "manifests"

func CreateCRD(clientset extClientset.Interface) error {
	log.Traceln("Creating CRD")

	_, err := clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Create(
		&extApi.CustomResourceDefinition{
			ObjectMeta: meta_v1.ObjectMeta{
				Name: fmt.Sprintf("%s.%s", NAME, v1alpha1.GroupName),
			},
			TypeMeta: meta_v1.TypeMeta{
				APIVersion: "apiextensions.k8s.io/v1beta1",
				Kind: "CustomResourceDefinition",
			},
			Spec: extApi.CustomResourceDefinitionSpec{
				Group: v1alpha1.GroupName,
				Versions: []extApi.CustomResourceDefinitionVersion {
					{
						Name:    v1alpha1.VersionNumber,
						Served:  true,
						Storage: true,
					},
				},
				Scope: extApi.ClusterScoped,
				Names: extApi.CustomResourceDefinitionNames{
					Singular: "manifest",
					Plural: NAME,
					Kind: "Manifest",
				},
			},
		})

	if err != nil && apierrors.IsAlreadyExists(err) {
		log.Traceln("CRD already existed")
		return nil
	}
	return err
}
