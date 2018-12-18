package image_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/pfs/pkg/image"
)

var _ = Describe("Id", func() {
	It("should sanitise filenames to suit Windows", func() {
	    i := image.NewId("sha:aa")
	    Expect(i.Filename()).To(Equal("sha-aa"))
	})
})
