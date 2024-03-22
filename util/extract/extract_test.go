package extract_test

import (
	"code.cloudfoundry.org/cli/util/extract"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("List", func() {
	It("extracts elements from a slice of structs", func() {
		input := []struct{ Name, Type string }{
			{Name: "flopsy", Type: "rabbit"},
			{Name: "mopsy", Type: "rabbit"},
			{Name: "cottontail", Type: "rabbit"},
			{Name: "peter", Type: "rabbit"},
			{Name: "mopsy", Type: "rabbit"},
			{Name: "jemima", Type: "duck"},
		}

		Expect(extract.List("Name", input)).To(Equal([]string{
			"flopsy",
			"mopsy",
			"cottontail",
			"peter",
			"mopsy",
			"jemima",
		}))

		Expect(extract.List("Type", input)).To(Equal([]string{
			"rabbit",
			"rabbit",
			"rabbit",
			"rabbit",
			"rabbit",
			"duck",
		}))
	})

	It("can recurse", func() {
		type a struct{ Name string }
		type b struct {
			Name     string
			Children []a
		}
		type c struct {
			Container b
		}
		type d struct {
			Name        string
			Descendents []c
		}
		input := []d{
			{
				Name: "alpha",
				Descendents: []c{
					{
						Container: b{
							Name: "foo",
							Children: []a{
								{Name: "flopsy"},
								{Name: "mopsy"},
							},
						},
					},
					{
						Container: b{
							Name:     "bar",
							Children: []a{},
						},
					},
				},
			},
			{
				Name: "beta",
			},
			{
				Name:        "gamma",
				Descendents: []c{},
			},
			{
				Name: "delta",
				Descendents: []c{
					{
						Container: b{
							Name: "baz",
							Children: []a{
								{Name: "cottontail"},
								{Name: "peter"},
							},
						},
					},
				},
			},
		}

		Expect(extract.List("Descendents.Container.Children.Name", input)).To(Equal([]string{
			"flopsy",
			"mopsy",
			"cottontail",
			"peter",
		}))
	})

	When("the expression does not match", func() {
		It("omits it from the result", func() {
			input := []interface{}{
				struct{ Name, Value string }{Name: "foo", Value: "bar"},
				struct{ Foo, Bar string }{Foo: "name", Bar: "value"},
			}

			Expect(extract.List("Name", input)).To(ConsistOf("foo"))
			Expect(extract.List("Foo", input)).To(ConsistOf("name"))
		})
	})

	When("the input is not a slice", func() {
		It("extracts a single value", func() {
			input := struct{ Foo, Bar string }{
				Foo: "foo",
				Bar: "bar",
			}

			Expect(extract.List("Foo", input)).To(ConsistOf("foo"))
		})
	})
})
