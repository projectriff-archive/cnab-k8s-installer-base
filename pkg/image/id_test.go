package image_test

import (
	"cnab-k8s-installer-base/pkg/image"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

)

var _ = Describe("Id", func() {
	It("should sanitise filenames to suit Windows", func() {
	    i := image.NewId("sha:aa")
	    Expect(i.Filename()).To(Equal("sha-aa"))
	})
})
