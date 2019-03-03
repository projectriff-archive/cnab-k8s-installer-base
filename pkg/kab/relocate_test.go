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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("RelocateManifest", func() {

	var (
		manifest *v1alpha1.Manifest
		destRepo string
		err      error
	)

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

	Context("when the manifest has Spec.Resource.Content", func() {
		It("repository relocation is successful", func() {
			err = kab.RelocateManifest(manifest, destRepo)
			Expect(err).To(BeNil())

			for i := range manifest.Spec.Resources {
				Expect(manifest.Spec.Resources[i].Content).NotTo(ContainSubstring("docker.io"))
				Expect(manifest.Spec.Resources[i].Content).NotTo(ContainSubstring("gcr.io"))
				Expect(manifest.Spec.Resources[i].Content).To(ContainSubstring(destRepo))
			}
		})
	})

})
