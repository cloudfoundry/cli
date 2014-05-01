/*
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package candiedyaml

import (
	"bytes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"math"
	"time"
)

var _ = Describe("Encode", func() {
	var buf *bytes.Buffer
	var enc *Encoder

	BeforeEach(func() {
		buf = &bytes.Buffer{}
		enc = NewEncoder(buf)
	})

	Context("Scalars", func() {
		It("handles strings", func() {
			enc.Encode("abc")
			Ω(buf.String()).Should(Equal(`"abc"
`))
		})

		Context("handles ints", func() {
			It("handles ints", func() {
				enc.Encode(13)
				Ω(buf.String()).Should(Equal("13\n"))
			})

			It("handles uints", func() {
				enc.Encode(uint64(1))
				Ω(buf.String()).Should(Equal("1\n"))
			})
		})

		Context("handles floats", func() {
			It("handles float32", func() {
				enc.Encode(float32(1.234))
				Ω(buf.String()).Should(Equal("1.234\n"))

			})

			It("handles float64", func() {
				enc.Encode(float64(1.2e23))
				Ω(buf.String()).Should(Equal("1.2e+23\n"))
			})

			It("handles NaN", func() {
				enc.Encode(math.NaN())
				Ω(buf.String()).Should(Equal(".nan\n"))
			})

			It("handles infinity", func() {
				enc.Encode(math.Inf(-1))
				Ω(buf.String()).Should(Equal("-.inf\n"))
			})
		})

		It("handles bools", func() {
			enc.Encode(true)
			Ω(buf.String()).Should(Equal("true\n"))
		})

		It("handles time.Time", func() {
			t := time.Now()
			enc.Encode(t)
			bytes, _ := t.MarshalText()
			Ω(buf.String()).Should(Equal(string(bytes) + "\n"))
		})

		Context("Null", func() {
			It("fails on nil", func() {
				Ω(enc.Encode(nil)).Should(HaveOccurred())
			})
		})

		It("handles []byte", func() {
			enc.Encode([]byte{'a', 'b', 'c'})
			Ω(buf.String()).Should(Equal("!!binary YWJj\n"))
		})

		Context("Ptrs", func() {
			It("handles ptr of a type", func() {
				p := new(int)
				*p = 10
				enc.Encode(p)
				Ω(buf.String()).Should(Equal("10\n"))
			})

			It("handles nil ptr", func() {
				var p *int
				enc.Encode(p)
				Ω(buf.String()).Should(Equal("null\n"))
			})
		})

		Context("Structs", func() {
			It("handles simple structs", func() {
				type batter struct {
					Name string
					HR   int64
					AVG  float64
				}

				batters := []batter{
					batter{Name: "Mark McGwire", HR: 65, AVG: 0.278},
					batter{Name: "Sammy Sosa", HR: 63, AVG: 0.288},
				}
				enc.Encode(batters)
				Ω(buf.String()).Should(Equal(`- "Name": "Mark McGwire"
  "HR": 65
  "AVG": 0.278
- "Name": "Sammy Sosa"
  "HR": 63
  "AVG": 0.288
`))
			})

			It("handles tagged structs", func() {
				type batter struct {
					Name string `yaml:"name"`
					HR   int64
					AVG  float64 `yaml:"avg"`
				}

				batters := []batter{
					batter{Name: "Mark McGwire", HR: 65, AVG: 0.278},
					batter{Name: "Sammy Sosa", HR: 63, AVG: 0.288},
				}
				enc.Encode(batters)
				Ω(buf.String()).Should(Equal(`- "name": "Mark McGwire"
  "HR": 65
  "avg": 0.278
- "name": "Sammy Sosa"
  "HR": 63
  "avg": 0.288
`))
			})

			It("handles nested structs", func() {
				type nestedConfig struct {
					AString string `yaml:"str"`
					Integer int    `yaml:"int"`
				}
				type config struct {
					TopString string
					Nested    nestedConfig
				}

				cfg := config{
					TopString: "def",
					Nested: nestedConfig{
						AString: "abc",
						Integer: 123,
					},
				}

				enc.Encode(cfg)

				Ω(buf.String()).Should(Equal(`"TopString": "def"
"Nested":
  "str": "abc"
  "int": 123
`))
			})

			It("handles inline structs", func() {
				type NestedConfig struct {
					AString string `yaml:"str"`
					Integer int    `yaml:"int"`
				}
				type config struct {
					TopString string
					NestedConfig
				}

				cfg := config{
					TopString: "def",
					NestedConfig: NestedConfig{
						AString: "abc",
						Integer: 123,
					},
				}

				enc.Encode(cfg)

				Ω(buf.String()).Should(Equal(`"TopString": "def"
"str": "abc"
"int": 123
`))
			})

			It("handles inline structs with conflicts", func() {
				type NestedConfig struct {
					AString string `yaml:"str"`
					Integer int    `yaml:"int"`
				}
				type config struct {
					AString string `yaml:"str"`
					NestedConfig
				}

				cfg := config{
					AString: "def",
					NestedConfig: NestedConfig{
						AString: "abc",
						Integer: 123,
					},
				}

				enc.Encode(cfg)

				Ω(buf.String()).Should(Equal(`"str": "def"
"int": 123
`))
			})

		})

	})

	Context("Sequence", func() {
		It("handles slices", func() {
			val := []string{"a", "b", "c"}
			enc.Encode(val)

			Ω(buf.String()).Should(Equal(`- "a"
- "b"
- "c"
`))
		})
	})

	Context("Maps", func() {
		It("Decodes simple maps", func() {
			enc.Encode(&map[string]string{
				"name": "Mark McGwire",
				"hr":   "65",
				"avg":  "0.278",
			})

			Ω(buf.String()).Should(Equal(`"avg": "0.278"
"hr": "65"
"name": "Mark McGwire"
`))
		})

		It("Decodes mix types", func() {
			enc.Encode(&map[string]interface{}{
				"name": "Mark McGwire",
				"hr":   65,
				"avg":  0.278,
			})

			Ω(buf.String()).Should(Equal(`"avg": 0.278
"hr": 65
"name": "Mark McGwire"
`))
		})
	})

	Context("Sequence of Maps", func() {
		It("decodes", func() {
			enc.Encode([]map[string]interface{}{
				{"name": "Mark McGwire",
					"hr":  65,
					"avg": 0.278,
				},
				{"name": "Sammy Sosa",
					"hr":  63,
					"avg": 0.288,
				},
			})

			Ω(buf.String()).Should(Equal(`- "avg": 0.278
  "hr": 65
  "name": "Mark McGwire"
- "avg": 0.288
  "hr": 63
  "name": "Sammy Sosa"
`))
		})
	})

	Context("Maps of Sequence", func() {
		It("decodes", func() {
			enc.Encode(map[string][]interface{}{
				"name": []interface{}{"Mark McGwire", "Sammy Sosa"},
				"hr":   []interface{}{65, 63},
				"avg":  []interface{}{0.278, 0.288},
			})

			Ω(buf.String()).Should(Equal(`"avg":
- 0.278
- 0.288
"hr":
- 65
- 63
"name":
- "Mark McGwire"
- "Sammy Sosa"
`))
		})
	})

	Context("Flow", func() {
		It("flows structs", func() {
			type i struct {
				A string
			}
			type o struct {
				I i `yaml:"i,flow"`
			}

			enc.Encode(o{
				I: i{A: "abc"},
			})
			Ω(buf.String()).Should(Equal(`"i": {"A": "abc"}
`))
		})

		It("flows sequences", func() {
			type i struct {
				A string
			}
			type o struct {
				I []i `yaml:"i,flow"`
			}

			enc.Encode(o{
				I: []i{{A: "abc"}},
			})
			Ω(buf.String()).Should(Equal(`"i": [{"A": "abc"}]
`))
		})
	})

	Context("Omit empty", func() {
		It("omits nil ptrs", func() {
			type i struct {
				A *string `yaml:"a,omitempty"`
			}
			type o struct {
				I []i `yaml:"i,flow"`
			}

			enc.Encode(o{
				I: []i{{A: nil}},
			})
			Ω(buf.String()).Should(Equal(`"i": [{}]
`))
		})

	})

	Context("Skip field", func() {
		It("does not include the field", func() {
			type a struct {
				B string `yaml:"-"`
				C string
			}

			enc.Encode(a{B: "b", C: "c"})
			Ω(buf.String()).Should(Equal(`"C": "c"
`))
		})
	})

	Context("Marshaler support", func() {
		Context("Receiver is a value", func() {
			It("uses the Marshaler interface when a value", func() {
				enc.Encode(hasMarshaler{Value: 123})
				Ω(buf.String()).Should(Equal("123\n"))
			})

			It("uses the Marshaler interface when a pointer", func() {
				enc.Encode(&hasMarshaler{Value: "abc"})
				Ω(buf.String()).Should(Equal(`"abc"
`))
			})
		})

		Context("Receiver is a pointer", func() {
			It("uses the Marshaler interface when a pointer", func() {
				enc.Encode(&hasPtrMarshaler{Value: map[string]string{"a": "b"}})
				Ω(buf.String()).Should(Equal(`"a": "b"
`))
			})
			It("skips the Marshaler when its a value", func() {
				enc.Encode(hasPtrMarshaler{Value: map[string]string{"a": "b"}})
				Ω(buf.String()).Should(Equal(`"Tag": ""
"Value":
  "a": "b"
`))
			})
		})
	})

	Context("Number type", func() {
		It("encodes as a number", func() {
			n := Number("12345")
			enc.Encode(n)
			Ω(buf.String()).Should(Equal("12345\n"))
		})
	})
})

type hasMarshaler struct {
	Value interface{}
}

func (m hasMarshaler) MarshalYAML() (tag string, value interface{}) {
	return "", m.Value
}

func (m hasMarshaler) UnmarshalYAML(tag string, value interface{}) error {
	m.Value = value
	return nil
}

type hasPtrMarshaler struct {
	Tag   string
	Value interface{}
}

func (m *hasPtrMarshaler) MarshalYAML() (tag string, value interface{}) {
	return "", m.Value
}

func (m *hasPtrMarshaler) UnmarshalYAML(tag string, value interface{}) error {
	m.Tag = tag
	m.Value = value
	return nil
}
