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

package scan_test

import (
	"os"
	"path/filepath"

	"cnab-k8s-installer-base/pkg/scan"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ListKind", func() {
	var (
		res     string
		baseDir string
		kinds   []string
		err     error
	)

	BeforeEach(func() {
		wd, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		baseDir = filepath.Join(wd, "fixtures")
	})

	JustBeforeEach(func() {
		kinds, err = scan.ListKind(res, baseDir)
	})

	Context("when the resource file does not contain 'kind' key", func() {
		BeforeEach(func() {
			res = "simple.yaml"
		})

		It("an empty list is returned", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(kinds).To(BeEmpty())
		})
	})

	Context("when the resource file contains block scalars", func() {
		BeforeEach(func() {
			res = "block.yaml"
		})

		It("should list the kinds", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(kinds).To(ConsistOf("ConfigMap"))
		})
	})

	Context("when using a realistic resource file", func() {
		BeforeEach(func() {
			res = "complex.yaml"
		})

		It("should list the kinds in the resource file", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(kinds).To(ConsistOf("Namespace", "ClusterRole", "ServiceAccount", "ClusterRoleBinding",
				"CustomResourceDefinition", "Service", "ConfigMap", "Deployment", "Gateway", "HorizontalPodAutoscaler",
			))
		})
	})

	Context("when using a simple parameterized resource file", func() {
		BeforeEach(func() {
			res = "parameterized.yaml"
		})

		It("should list the kinds in the resource file", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(kinds).To(ConsistOf("BuildTemplate"))
		})
	})

	Context("when using a more complex parameterized resource file", func() {
		BeforeEach(func() {
			res = "parameterized-2.yaml"
		})

		It("should list the kinds", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(kinds).To(ConsistOf("ClusterBuildTemplate"))
		})
	})

	Context("when the resource file is not found", func() {
		BeforeEach(func() {
			res = "nosuch.yaml"
		})

		It("should return a suitable error", func() {
			Expect(os.IsNotExist(err)).To(BeTrue())
		})
	})

	Context("when the resource file contains invalid YAML", func() {
		BeforeEach(func() {
			res = "invalid.yaml"
		})

		It("should return a suitable error", func() {
			Expect(err).To(MatchError(HavePrefix("error parsing content")))
		})
	})
})
