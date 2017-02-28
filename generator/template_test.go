package generator

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Template", func() {
	Describe("cleanname()", func() {
		It("does not strip d from entities with a name like time_cardid", func() {
			Ω(cleanname("time_cardid")).Should(BeEquivalentTo("TimeCardID"))
		})

		It("transforms approvalprocess to ApprovalProcess", func() {
			Ω(cleanname("po_approvalprocess")).Should(BeEquivalentTo("PoApprovalProcess"))
		})

		It("transforms id to ID", func() {
			Ω(cleanname("id")).Should(BeEquivalentTo("ID"))
			Ω(cleanname("thing_id")).Should(BeEquivalentTo("ThingID"))
			Ω(cleanname("thingid")).Should(BeEquivalentTo("ThingID"))
			Ω(cleanname("thingId")).Should(BeEquivalentTo("ThingID"))
		})

		It("transforms single character segments to uppercase", func() {
			Ω(cleanname("the_a_thing")).Should(BeEquivalentTo("TheAThing"))
		})

		It("removes extraneous underscores", func() {
			Ω(cleanname("the___thing")).Should(BeEquivalentTo("TheThing"))
		})

		It("handles custom fields", func() {
			Ω(cleanname("the_custom_thing__c")).Should(BeEquivalentTo("TheCustomThingC"))
		})

		It("transforms api to API", func() {
			Ω(cleanname("theapi")).Should(BeEquivalentTo("TheAPI"))
			Ω(cleanname("theApi")).Should(BeEquivalentTo("TheAPI"))
			Ω(cleanname("api_thing")).Should(BeEquivalentTo("APIThing"))
			Ω(cleanname("Api_thing")).Should(BeEquivalentTo("APIThing"))
		})

		It("transforms url to URL", func() {
			Ω(cleanname("theurl")).Should(BeEquivalentTo("TheURL"))
			Ω(cleanname("theUrl")).Should(BeEquivalentTo("TheURL"))

			Ω(cleanname("url_thing")).Should(BeEquivalentTo("URLThing"))
			Ω(cleanname("Url_thing")).Should(BeEquivalentTo("URLThing"))
		})
	})

	Describe("tag()", func() {
		It("handles string types", func() {
			Ω(tag("fieldname", "string")).Should(BeEquivalentTo("`xml:\"fieldname,omitempty\" json:\"fieldname,omitempty\"`"))
			Ω(tag("fieldname", "")).Should(BeEquivalentTo("`xml:\"fieldname,omitempty\" json:\"fieldname,omitempty\"`"))
		})

		It("handles date types", func() {
			Ω(tag("fieldname", "Date")).Should(BeEquivalentTo("`xml:\"fieldname>Date,omitempty\" json:\"fieldname,omitempty\"`"))
			Ω(tag("fieldname", "date")).Should(BeEquivalentTo("`xml:\"fieldname>Date,omitempty\" json:\"fieldname,omitempty\"`"))
		})

		It("handles address types", func() {
			Ω(tag("fieldname", "Address")).Should(BeEquivalentTo("`xml:\"fieldname>Address,omitempty\" json:\"fieldname,omitempty\"`"))
			Ω(tag("fieldname", "address")).Should(BeEquivalentTo("`xml:\"fieldname>Address,omitempty\" json:\"fieldname,omitempty\"`"))
		})
	})

	Describe("xmltag()", func() {
		It("returns an xml struct tag", func() {
			Ω(xmltag("fieldname,attr")).Should(BeEquivalentTo("`xml:\"fieldname,attr\"`"))
		})
	})

	Describe("xmlrawtag()", func() {
		It("returns an xml struct tag", func() {
			Ω(xmlrawtag("fieldname")).Should(BeEquivalentTo("`xml:\"fieldname,omitempty\"`"))
		})
	})

	Describe("cleannamelower()", func() {
		It("returns a lowercase cleaned name", func() {
			Ω(cleannamelower("customer_thing")).Should(BeEquivalentTo("customerthing"))
			Ω(cleannamelower("customer_thing_id")).Should(BeEquivalentTo("customerthingid"))
		})
	})

	Describe("valueforkey()", func() {
		It("returns the value", func() {
			themap := map[string]string{
				"1": "one",
				"2": "two",
			}
			Ω(valueforkey("1", themap)).Should(BeEquivalentTo("one"))
		})
	})

	Describe("backtick()", func() {
		It("returns a backtick", func() {
			Ω(backtick()).Should(BeEquivalentTo("`"))
		})
	})
})
