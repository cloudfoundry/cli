package jsonry_test

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/jsonry"
	"code.cloudfoundry.org/cli/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Marshal", func() {
	It("marshals a field", func() {
		s := struct{ Foo string }{Foo: "works"}

		json, err := jsonry.Marshal(s)
		Expect(err).NotTo(HaveOccurred())
		Expect(json).To(MatchJSON(`{"foo": "works"}`))
	})

	It("reads `json` and `jsonry` tags", func() {
		s := struct {
			A string `json:"foo"`
			B string `jsonry:"bar"`
			C string
		}{
			A: "JSON",
			B: "JSONry",
			C: "inference",
		}

		json, err := jsonry.Marshal(s)
		Expect(err).NotTo(HaveOccurred())
		Expect(json).To(MatchJSON(`{"foo": "JSON", "bar": "JSONry", "c": "inference"}`))
	})

	It("marshals from a struct pointer", func() {
		s := struct{ Foo string }{Foo: "struct pointer works"}

		json, err := jsonry.Marshal(&s)
		Expect(err).NotTo(HaveOccurred())
		Expect(json).To(MatchJSON(`{"foo": "struct pointer works"}`))
	})

	It("marshals from a pointer field", func() {
		r := "pointer works"
		s := struct{ A, B *string }{A: &r}

		json, err := jsonry.Marshal(s)
		Expect(err).NotTo(HaveOccurred())
		Expect(json).To(MatchJSON(`{"a": "pointer works", "b": null}`))
	})

	It("marshals lists", func() {
		s := struct {
			A []string
			B *[]string
			C []int
		}{
			A: []string{"q", "w", "e"},
			B: &[]string{"r", "t", "y"},
			C: []int{1, 2, 3},
		}

		json, err := jsonry.Marshal(s)
		Expect(err).NotTo(HaveOccurred())
		Expect(json).To(MatchJSON(`{"a": ["q", "w", "e"], "b": ["r", "t", "y"], "c": [1, 2, 3]}`))
	})

	It("marshals into nested JSON", func() {
		s := struct {
			A string `jsonry:"aa.bb.cccc.d.e7.foo"`
		}{
			A: "deep JSONry works!",
		}

		json, err := jsonry.Marshal(s)
		Expect(err).NotTo(HaveOccurred())
		Expect(json).To(MatchJSON(`{"aa":{"bb":{"cccc":{"d":{"e7":{"foo":"deep JSONry works!"}}}}}}`))
	})

	It("marshals into nested JSON lists using array hints", func() {
		s := struct {
			G []int    `jsonry:"aa[].g"`
			H []string `jsonry:"aa[].name"`
			I *[]bool  `jsonry:"aa[].ii"`
		}{
			G: []int{4, 8, 7},
			H: []string{"a", "b", "c", "d"},
			I: &[]bool{true},
		}

		e := `{
				"aa": [
					{"g": 4, "name": "a", "ii": true},
					{"g": 8, "name": "b"},
					{"g": 7, "name": "c"},
					{"name": "d"}
				]
			}`

		json, err := jsonry.Marshal(s)
		Expect(err).NotTo(HaveOccurred())
		Expect(json).To(MatchJSON(e))
	})

	It("marshals recursively", func() {
		type s struct {
			Foo string
		}

		t := struct {
			Bar s
			Baz *s
		}{
			Bar: s{Foo: "recursion works"},
			Baz: &s{Foo: "pointer recursion works"},
		}

		e := `{"bar": {"foo": "recursion works"}, "baz": {"foo": "pointer recursion works"}}`

		json, err := jsonry.Marshal(t)
		Expect(err).NotTo(HaveOccurred())
		Expect(json).To(MatchJSON(e))
	})

	It("can omit empty values", func() {
		type s struct {
			A string  `jsonry:",omitempty"`
			B string  `jsonry:"bee,omitempty"`
			C *string `jsonry:"c,omitempty"`
		}

		var t s
		json, err := jsonry.Marshal(t)
		Expect(err).NotTo(HaveOccurred())
		Expect(json).To(MatchJSON(`{}`))

		c := "foo"
		t.A = "AA"
		t.B = "BB"
		t.C = &c
		json, err = jsonry.Marshal(t)
		Expect(err).NotTo(HaveOccurred())
		Expect(json).To(MatchJSON(`{"a": "AA", "bee": "BB", "c": "foo"}`))
	})

	It("marshals a realistic object", func() {
		type metadata struct {
			Labels map[string]types.NullString `json:"labels,omitempty"`
		}

		st := struct {
			Name      string
			GUID      string    `jsonry:"guid"`
			URL       string    `json:"url"`
			Type      string    `jsonry:"authentication.type"`
			Username  string    `jsonry:"authentication.credentials.username"`
			Password  string    `jsonry:"authentication.credentials.password"`
			SpaceGUID string    `jsonry:"relationships.space.data.guid"`
			Metadata  *metadata `json:"metadata,omitempty"`
		}{
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
		}

		e := `
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

		json, err := jsonry.Marshal(st)
		Expect(err).NotTo(HaveOccurred())
		Expect(json).To(MatchJSON(e))
	})

	When("there is a type mismatch", func() {
		It("fails for simple types", func() {
			var s struct{ S int }
			err := jsonry.Unmarshal([]byte(`{"s": "hello"}`), &s)
			Expect(err).To(MatchError("could not convert value 'hello' type 'string' to 'int' for field 'S'"))
		})

		It("fails when the field expects a list", func() {
			var s struct{ S []string }
			err := jsonry.Unmarshal([]byte(`{"s": "hello"}`), &s)
			Expect(err).To(MatchError(`could not convert value 'hello' type 'string' to '[]string' for field 'S' because it is not a list type`))
		})

		It("fails for elements in a list", func() {
			var s struct{ S []int }
			err := jsonry.Unmarshal([]byte(`{"s": [4, "hello"]}`), &s)
			Expect(err).To(MatchError(`could not convert value 'hello' type 'string' to 'int' for field 'S' index 1`))
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

		It("leaves zero value in a list", func() {
			var s struct {
				T []string `jsonry:"t.a"`
			}

			data := `{"t": [{"a": "foo"}, {}, {"a": "bar"}]}`
			err := jsonry.Unmarshal([]byte(data), &s)
			Expect(err).NotTo(HaveOccurred())
			Expect(s.T).To(Equal([]string{"foo", "", "bar"}))
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

	When("marshalling a non-list into a list", func() {
		It("converts it into a list", func() {
			s := struct {
				G int `jsonry:"aa[].g"`
			}{
				G: 42,
			}

			json, err := jsonry.Marshal(s)
			Expect(err).NotTo(HaveOccurred())
			Expect(json).To(MatchJSON(`{"aa": [{"g": 42}]}`))
		})
	})

	When("the source object is not valid", func() {
		It("rejects a non-struct", func() {
			var s int
			_, err := jsonry.Marshal(s)
			Expect(err).To(MatchError("the source object must be a struct"))
		})

		It("rejects nil", func() {
			_, err := jsonry.Marshal(nil)
			Expect(err).To(MatchError("the source object must be a valid struct"))
		})
	})
})
