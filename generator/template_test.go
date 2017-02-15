package generator

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Template", func() {
	It("Does not strip d from entities with a name like time_cardid", func() {
		Ω(cleanname("time_cardid")).Should(BeEquivalentTo("TimeCardID"))
	})
})
