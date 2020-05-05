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

	It("unmarshals lists", func() {
		var s struct {
			A []string
			B *[]string
			C []int
		}

		data := `{"a": ["q", "w", "e"], "b": ["r", "t", "y"], "c": [1, 2, 3]}`
		err := jsonry.Unmarshal([]byte(data), &s)
		Expect(err).NotTo(HaveOccurred())
		Expect(s.A).To(Equal([]string{"q", "w", "e"}))
		Expect(s.B).To(PointTo(Equal([]string{"r", "t", "y"})))
		Expect(s.C).To(Equal([]int{1, 2, 3}))
	})

	It("unmarshals lists of struct", func() {
		type T struct {
			D string `json:"d"`
			E int    `json:"e"`
		}

		var s struct {
			P string
			A []T
		}

		data := `{"p":"baz", "a": [ {"d": "foo", "e": 123}, {"d": "bar", "e": 1} ]}`

		err := jsonry.Unmarshal([]byte(data), &s)

		Expect(err).NotTo(HaveOccurred())
		Expect(s.P).To(Equal("baz"))
		Expect(s.A[0].D).To(Equal("foo"))
		Expect(s.A[0].E).To(Equal(123))
		Expect(s.A[1].D).To(Equal("bar"))
		Expect(s.A[1].E).To(Equal(1))
	})

	It("unmarshals reference to nested JSON", func() {
		var s struct {
			A string `jsonry:"aa.bb.cccc.d.e7.foo"`
		}
		data := `{"aa":{"bb":{"cccc":{"d":{"e7":{"foo":"deep JSONry works!"}}}}}}`

		err := jsonry.Unmarshal([]byte(data), &s)
		Expect(err).NotTo(HaveOccurred())
		Expect(s.A).To(Equal("deep JSONry works!"))
	})

	It("unmarshals reference to nested JSON lists", func() {
		var s struct {
			G []int `jsonry:"aa.g"`
		}
		data := `{"aa": [{"g": 4}, {"g": 8}, {"g": 7}]}`

		err := jsonry.Unmarshal([]byte(data), &s)
		Expect(err).NotTo(HaveOccurred())
		Expect(s.G).To(Equal([]int{4, 8, 7}))
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

	It("unmarshals into aliased types", func() {
		type rope string

		var s struct {
			A rope
			B *rope
		}

		err := jsonry.Unmarshal([]byte(`{"a": "foo", "b": "bar"}`), &s)
		Expect(err).NotTo(HaveOccurred())
		Expect(s.A).To(Equal(rope("foo")))
		Expect(s.B).To(PointTo(Equal(rope("bar"))))
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

		When("in a list of structs", func() {
			It("fails when its not a struct", func() {
				var s struct{ S []struct{ A string } }
				err := jsonry.Unmarshal([]byte(`{"s": ["123"]}`), &s)
				Expect(err).To(MatchError(`could not convert value '123' type 'string' to 'struct { A string }' for field 'S' index 0`))
			})

			It("succeeds when the struct doesn't match", func() {
				var s struct {
					S []struct {
						A bool
					}
				}
				err := jsonry.Unmarshal([]byte(`{"s": [{"b": "123"}]}`), &s)
				Expect(err).ToNot(HaveOccurred())
				Expect(s.S[0].A).To(BeFalse())
			})
		})
	})

	Context("numbers", func() {
		It("can unmarshal an int", func() {
			var s struct{ I int }
			err := jsonry.Unmarshal([]byte(`{"i": 42}`), &s)
			Expect(err).NotTo(HaveOccurred())
			Expect(s.I).To(Equal(42))
		})

		It("can unmarshal a float", func() {
			var s struct{ F float64 }

			err := jsonry.Unmarshal([]byte(`{"f": 42.02}`), &s)

			Expect(err).NotTo(HaveOccurred())
			Expect(s.F).To(Equal(42.02))
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
