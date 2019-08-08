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

package kustomize_test

import (
	"io/ioutil"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/projectriff/cnab-k8s-installer-base/pkg/kustomize"
	"github.com/projectriff/cnab-k8s-installer-base/pkg/test_support"
)

var _ = Describe("Kustomize wrapper", func() {

	var (
		initLabels              map[string]string
		initialResourceContent  string
		expectedResourceContent string
		kustomizer              kustomize.Kustomizer
		timeout                 time.Duration
		workDir                 string
	)

	BeforeEach(func() {
		initLabels = map[string]string{"created-by": "riff", "because": "we-can"}
		initialResourceContent = `kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: riff-cnb-cache
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 8Gi`
		expectedResourceContent = `apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  labels:
    because: we-can
    created-by: riff
  name: riff-cnb-cache
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 8Gi
`
		timeout = 500 * time.Millisecond
		kustomizer = kustomize.MakeKustomizer(timeout)
		workDir = test_support.CreateTempDir()
	})

	AfterEach(func() {
		test_support.CleanupDirs(GinkgoT(), workDir)
	})

	It("customizes local resources with provided labels", func() {
		file := test_support.CreateFile(workDir, "pvc.yaml", initialResourceContent)
		content, err := ioutil.ReadFile(file)
		Expect(err).NotTo(HaveOccurred())

		result, err := kustomizer.ApplyLabels(string(content), initLabels)

		Expect(err).NotTo(HaveOccurred())
		Expect(string(result)).To(Equal(expectedResourceContent))
	})
})
