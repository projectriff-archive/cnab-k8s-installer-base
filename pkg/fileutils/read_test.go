/*
 * Copyright 2018 The original author or authors
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

package fileutils_test

import (
	"cnab-k8s-installer-base/pkg/fileutils"
	"cnab-k8s-installer-base/pkg/test_support"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

var _ = Describe("Read", func() {

	var (
		file    string
		base    string
		content []byte
		err     error
	)

	JustBeforeEach(func() {
		content, err = fileutils.Read(file, base)
	})

	Context("when file is empty", func() {
		BeforeEach(func() {
			file = ""
			base = ""
		})
		It("does not return an error", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(Equal(""))
		})
	})

	Context("when file is a URL", func() {
		BeforeEach(func() {
			file = getwdAsURL() + "/fixtures/file.txt"

			base = "" // irrelevant when file is absolute
		})

		It("should read the file content", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(Equal("contents"))
		})
	})

	Context("when file is an absolute path", func() {
		BeforeEach(func() {
			cwd, err := os.Getwd()
			Expect(err).NotTo(HaveOccurred())
			file = filepath.Join(cwd, "fixtures", "file.txt")

			base = "" // irrelevant when file is absolute
		})

		It("should read the file content", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(Equal("contents"))
		})
	})

	Context("when file is a relative path", func() {
		BeforeEach(func() {
			file = filepath.Join("fixtures", "file.txt")

			base = "" // irrelevant when file is absolute
		})

		Context("when base is empty", func() {
			BeforeEach(func() {
				base = ""
			})

			It("should read the file content", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(string(content)).To(Equal("contents"))
			})
		})

		Context("when base is a URL", func() {
			BeforeEach(func() {
				base = getwdAsURL()
			})

			It("should read the file content", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(string(content)).To(Equal("contents"))
			})
		})

		Context("when base is an absolute file path", func() {
			BeforeEach(func() {
				var err error
				base, err = os.Getwd()
				Expect(err).NotTo(HaveOccurred())
			})

			It("should read the file content", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(string(content)).To(Equal("contents"))
			})
		})

		Context("when base is a relative file path", func() {
			BeforeEach(func() {
				base = "fixtures"

				file = "file.txt"
			})

			It("should read the file content", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(string(content)).To(Equal("contents"))
			})
		})

		Context("when base is a relative file path expressed using dot", func() {
			BeforeEach(func() {
				base = "./fixtures"

				file = "file.txt"
			})

			It("should read the file content", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(string(content)).To(Equal("contents"))
			})
		})

		Context("when a file is a URL with an unsupported protocol", func() {
			BeforeEach(func() {
				base = "" // irrelevant when file is absolute

				file = "ftp://localhost/some-file.txt"
			})

			It("should read the file content", func() {
				Expect(err).To(MatchError("unsupported URL scheme ftp in ftp://localhost/some-file.txt"))
			})
		})
	})

})

var _ = Describe("ReadUrl", func() {

	const (
		timeout = 200 * time.Millisecond
	)

	It("reads file URLs", func() {
		resourceUrl, _ := url.Parse(test_support.FileURL(test_support.AbsolutePath("fixtures/file.txt")))

		result, err := fileutils.ReadUrl(resourceUrl, timeout)

		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal([]byte("contents")))
	})

	It("reads HTTP URLs", func() {
		listener, _ := net.Listen("tcp", "127.0.0.1:0")
		go func() {
			err := test_support.Serve(listener, test_support.HttpResponse{
				Headers: map[string]string{"Content-Type": "text/plain"},
				Content: []byte("contents"),
			})
			Expect(err).NotTo(HaveOccurred())
		}()
		resourceUrl, _ := url.Parse(fmt.Sprintf("https://%s/%s", listener.Addr().String(), ""))

		result, err := fileutils.ReadUrl(resourceUrl, timeout)

		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal([]byte("contents")))
	})

	It("fails if fetching the remote resource to serve takes too long", func() {
		resourceListener, _ := net.Listen("tcp", "127.0.0.1:0")
		go func() {
			err := test_support.ServeSlow(resourceListener, test_support.HttpResponse{}, 2*timeout)
			Expect(err).NotTo(HaveOccurred())
		}()
		resourceUrl, _ := url.Parse(fmt.Sprintf("https://%s/%s", resourceListener.Addr().String(), ""))

		_, err := fileutils.ReadUrl(resourceUrl, timeout)

		Expect(err).To(SatisfyAll(
			Not(BeNil()),
			BeAssignableToTypeOf(&url.Error{})))
	})
})

func getwdAsURL() string {
	cwd, err := os.Getwd()
	Expect(err).NotTo(HaveOccurred())
	return "file:///" + filepath.ToSlash(cwd)
}
