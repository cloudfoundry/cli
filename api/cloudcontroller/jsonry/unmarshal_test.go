package jsonry_test

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/jsonry"
	"code.cloudfoundry.org/cli/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("Unmarshal", func() {
	It("unmarshals a field", func() {
		var s struct{ Foo string }

		err := jsonry.Unmarshal([]byte(`{"foo": "works"}`), &s)
		Expect(err).NotTo(HaveOccurred())
		Expect(s.Foo).To(Equal("works"))
	})

	It("reads `json` and `jsonry` tags", func() {
		var s struct {
			A string `json:"foo"`
			B string `jsonry:"bar"`
			C string
		}

		data := `{"foo": "JSON", "bar": "JSONry", "c": "inference"}`

		err := jsonry.Unmarshal([]byte(data), &s)
		Expect(err).NotTo(HaveOccurred())
		Expect(s.A).To(Equal("JSON"))
		Expect(s.B).To(Equal("JSONry"))
		Expect(s.C).To(Equal("inference"))
	})

	It("unmarshals into a pointer", func() {
		var s struct{ A *string }

		err := jsonry.Unmarshal([]byte(`{"a": "pointer works"}`), &s)
		Expect(err).NotTo(HaveOccurred())
		Expect(s.A).To(PointTo(Equal("pointer works")))
	})

	It("unmarshals a deep JSONry reference", func() {
		var s struct {
			A string `jsonry:"aa.bb.cccc.d.e7.foo"`
		}
		data := `{"aa":{"bb":{"cccc":{"d":{"e7":{"foo":"deep JSONry works!"}}}}}}`

		err := jsonry.Unmarshal([]byte(data), &s)
		Expect(err).NotTo(HaveOccurred())
		Expect(s.A).To(Equal("deep JSONry works!"))
	})

	It("unmarshals recursively", func() {
		type s struct {
			Foo string
		}

		var t struct {
			Bar s
			Baz *s
		}

		data := `{"bar": {"foo": "recursion works"}, "baz": {"foo": "pointer recursion works"}}`

		err := jsonry.Unmarshal([]byte(data), &t)
		Expect(err).NotTo(HaveOccurred())
		Expect(t.Bar.Foo).To(Equal("recursion works"))
	})

	It("unmarshals a realistic object", func() {
		type metadata struct {
			Labels map[string]types.NullString `json:"labels,omitempty"`
		}

		type st struct {
			Name      string
			GUID      string    `jsonry:"guid"`
			URL       string    `json:"url"`
			Type      string    `jsonry:"authentication.type"`
			Username  string    `jsonry:"authentication.credentials.username"`
			Password  string    `jsonry:"authentication.credentials.password"`
			SpaceGUID string    `jsonry:"relationships.space.data.guid"`
			Metadata  *metadata `json:"metadata,omitempty"`
		}
		var s st

		data := `
        {
			"name": "foo",
			"guid": "long-guid",
			"url": "https://foo.com",
			"authentication": {
				"type": "basic",
				"credentials": {
					"username": "fake-user",
					"password": "fake secret"
				}
			},
			"metadata": {
				"annotations": {},
				"labels": {
					"l1": "foo",
					"l2": null
				}
			},
			"relationships": {
				"space": {
					"data": {
						"guid": "space-guid"
					}
				}
			}
		}`

		err := jsonry.Unmarshal([]byte(data), &s)
		Expect(err).NotTo(HaveOccurred())
		Expect(s).To(Equal(st{
			Name:      "foo",
			GUID:      "long-guid",
			URL:       "https://foo.com",
			Type:      "basic",
			Username:  "fake-user",
			Password:  "fake secret",
			SpaceGUID: "space-guid",
			Metadata: &metadata{
				Labels: map[string]types.NullString{
					"l1": types.NewNullString("foo"),
					"l2": types.NewNullString(),
				},
			},
		}))
	})

	Context("when there is a type mismatch", func() {
		It("fails", func() {
			var s struct{ S string }
			err := jsonry.Unmarshal([]byte(`{"s": 42}`), &s)
			Expect(err).To(MatchError("could not convert value '42' type 'json.Number' to 'string' for field 'S'"))
		})
	})

	Context("numbers", func() {
		It("can unmarshal an int", func() {
			var s struct{ I int }
			err := jsonry.Unmarshal([]byte(`{"i": 42}`), &s)
			Expect(err).NotTo(HaveOccurred())
			Expect(s.I).To(Equal(42))
		})

		It("fails when the JSON value is not an int", func() {
			var s struct{ I int }
			err := jsonry.Unmarshal([]byte(`{"i": "stuffs"}`), &s)
			Expect(err).To(MatchError("could not convert value 'stuffs' type 'string' to 'int' for field 'I'"))
		})
	})

	Context("the special map[string]types.NullString", func() {
		It("can unmarshal a map", func() {
			var s struct{ M map[string]types.NullString }
			err := jsonry.Unmarshal([]byte(`{"m": {"foo": "bar", "baz": null}}`), &s)
			Expect(err).NotTo(HaveOccurred())
			Expect(s.M).To(Equal(map[string]types.NullString{
				"foo": types.NewNullString("bar"),
				"baz": types.NewNullString(),
			}))
		})

		It("can unmarshal a pointer to a map", func() {
			var s struct{ M *map[string]types.NullString }
			err := jsonry.Unmarshal([]byte(`{"m": {"foo": "bar", "baz": null}}`), &s)
			Expect(err).NotTo(HaveOccurred())
			Expect(s.M).To(Equal(&map[string]types.NullString{
				"foo": types.NewNullString("bar"),
				"baz": types.NewNullString(),
			}))
		})
	})

	When("no data match is found", func() {
		It("leaves struct field at the default value", func() {
			var s struct {
				A string `jsonry:"foo"`
			}

			err := jsonry.Unmarshal([]byte(`{}`), &s)
			Expect(err).NotTo(HaveOccurred())
			Expect(s.A).To(Equal(""))
		})
	})

	When("the JSONry path navigates to a non-map", func() {
		It("leaves struct field at the default value", func() {
			var s struct {
				A string `jsonry:"foo.bar"`
			}

			err := jsonry.Unmarshal([]byte(`{"foo": 42}`), &s)
			Expect(err).NotTo(HaveOccurred())
			Expect(s.A).To(Equal(""))
		})
	})

	When("the data stream is invalid JSON", func() {
		It("fails", func() {
			var s struct {
				A string `jsonry:"foo"`
			}

			err := jsonry.Unmarshal([]byte(`{"invalid_json`), &s)
			Expect(err).To(MatchError("unexpected EOF"))
		})
	})

	When("the storage object is not valid", func() {
		It("rejects a non-pointer", func() {
			var s struct {
				A int `jsonry:"a"`
			}
			err := jsonry.Unmarshal([]byte{}, s)
			Expect(err).To(MatchError("the storage object must be a pointer"))
		})

		It("rejects a non-struct", func() {
			var s int
			err := jsonry.Unmarshal([]byte{}, &s)
			Expect(err).To(MatchError("the storage object pointer must point to a struct"))
		})
	})
})
