package main

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestComplianceHubMain(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "compliance-hub main suite")
}
