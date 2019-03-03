package kab_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestKab(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Kab Suite")
}
