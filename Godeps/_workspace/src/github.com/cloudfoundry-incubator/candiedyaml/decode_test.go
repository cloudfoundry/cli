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
	"math"
	"os"
	"strconv"
	"strings"
	"time"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Decode", func() {
	It("Decodes a file", func() {
		f, _ := os.Open("fixtures/specification/example2_1.yaml")
		d := NewDecoder(f)
		var v interface{}
		err := d.Decode(&v)

		Ω(err).ShouldNot(HaveOccurred())
	})

	Context("strings", func() {
		It("Decodes an empty string", func() {
			d := NewDecoder(strings.NewReader(`""
`))
			var v string
			err := d.Decode(&v)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(v).Should(Equal(""))
		})

		It("Decodes an empty string to an interface", func() {
			d := NewDecoder(strings.NewReader(`""
`))
			var v interface{}
			err := d.Decode(&v)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(v).Should(Equal(""))
		})

		It("Decodes a map containing empty strings to an interface", func() {
			d := NewDecoder(strings.NewReader(`"" : ""
`))
			var v interface{}
			err := d.Decode(&v)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(v).Should(Equal(map[interface{}]interface{}{"": ""}))
		})
	})

	Context("Sequence", func() {
		It("Decodes to interface{}s", func() {
			f, _ := os.Open("fixtures/specification/example2_1.yaml")
			d := NewDecoder(f)
			var v interface{}
			err := d.Decode(&v)

			Ω(err).ShouldNot(HaveOccurred())
			Ω((v).([]interface{})).To(Equal([]interface{}{"Mark McGwire", "Sammy Sosa", "Ken Griffey"}))
		})

		It("Decodes to []string", func() {
			f, _ := os.Open("fixtures/specification/example2_1.yaml")
			d := NewDecoder(f)
			v := make([]string, 0, 3)

			err := d.Decode(&v)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(v).To(Equal([]string{"Mark McGwire", "Sammy Sosa", "Ken Griffey"}))
		})

		It("Decodes a sequence of maps", func() {
			f, _ := os.Open("fixtures/specification/example2_12.yaml")
			d := NewDecoder(f)
			v := make([]map[string]interface{}, 1)

			err := d.Decode(&v)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(v).To(Equal([]map[string]interface{}{
				{"item": "Super Hoop", "quantity": int64(1)},
				{"item": "Basketball", "quantity": int64(4)},
				{"item": "Big Shoes", "quantity": int64(1)},
			}))
		})

		Describe("As structs", func() {
			It("Simple struct", func() {
				f, _ := os.Open("fixtures/specification/example2_4.yaml")
				d := NewDecoder(f)

				type batter struct {
					Name string
					HR   int64
					AVG  float64
				}
				v := make([]batter, 0, 1)

				err := d.Decode(&v)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(v).To(Equal([]batter{
					batter{Name: "Mark McGwire", HR: 65, AVG: 0.278},
					batter{Name: "Sammy Sosa", HR: 63, AVG: 0.288},
				}))
			})

			It("Tagged struct", func() {
				f, _ := os.Open("fixtures/specification/example2_4.yaml")
				d := NewDecoder(f)

				type batter struct {
					N string  `yaml:"name"`
					H int64   `yaml:"hr"`
					A float64 `yaml:"avg"`
				}
				v := make([]batter, 0, 1)

				err := d.Decode(&v)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(v).To(Equal([]batter{
					batter{N: "Mark McGwire", H: 65, A: 0.278},
					batter{N: "Sammy Sosa", H: 63, A: 0.288},
				}))
			})

			It("ignores missing tags", func() {
				f, _ := os.Open("fixtures/specification/example2_4.yaml")
				d := NewDecoder(f)

				type batter struct {
					N  string `yaml:"name"`
					HR int64
					A  float64
				}
				v := make([]batter, 0, 1)

				err := d.Decode(&v)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(v).To(Equal([]batter{
					batter{N: "Mark McGwire", HR: 65},
					batter{N: "Sammy Sosa", HR: 63},
				}))
			})
		})

		It("Decodes a sequence of sequences", func() {
			f, _ := os.Open("fixtures/specification/example2_5.yaml")
			d := NewDecoder(f)
			v := make([][]interface{}, 1)

			err := d.Decode(&v)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(v).To(Equal([][]interface{}{
				{"name", "hr", "avg"},
				{"Mark McGwire", int64(65), float64(0.278)},
				{"Sammy Sosa", int64(63), float64(0.288)},
			}))
		})
	})

	Context("Maps", func() {
		It("Decodes to interface{}s", func() {
			f, _ := os.Open("fixtures/specification/example2_2.yaml")
			d := NewDecoder(f)
			var v interface{}

			err := d.Decode(&v)
			Ω(err).ShouldNot(HaveOccurred())
			Ω((v).(map[interface{}]interface{})).To(Equal(map[interface{}]interface{}{
				"hr":  int64(65),
				"avg": float64(0.278),
				"rbi": int64(147),
			}))
		})

		It("Decodes to a struct", func() {
			f, _ := os.Open("fixtures/specification/example2_2.yaml")
			d := NewDecoder(f)

			type batter struct {
				HR  int64
				AVG float64
				RBI int64
			}
			v := batter{}

			err := d.Decode(&v)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(v).To(Equal(batter{HR: 65, AVG: 0.278, RBI: 147}))
		})

		It("Decodes to a map of string arrays", func() {
			f, _ := os.Open("fixtures/specification/example2_9.yaml")
			d := NewDecoder(f)
			v := make(map[string][]string)

			err := d.Decode(&v)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(v).To(Equal(map[string][]string{"hr": []string{"Mark McGwire", "Sammy Sosa"}, "rbi": []string{"Sammy Sosa", "Ken Griffey"}}))
		})
	})

	Context("Sequence of Maps", func() {
		It("Decodes to interface{}s", func() {
			f, _ := os.Open("fixtures/specification/example2_4.yaml")
			d := NewDecoder(f)
			var v interface{}

			err := d.Decode(&v)
			Ω(err).ShouldNot(HaveOccurred())
			Ω((v).([]interface{})).To(Equal([]interface{}{
				map[interface{}]interface{}{"name": "Mark McGwire", "hr": int64(65), "avg": float64(0.278)},
				map[interface{}]interface{}{"name": "Sammy Sosa", "hr": int64(63), "avg": float64(0.288)},
			}))
		})
	})

	It("Decodes ascii art", func() {
		f, _ := os.Open("fixtures/specification/example2_13.yaml")
		d := NewDecoder(f)
		v := ""

		err := d.Decode(&v)
		Ω(err).ShouldNot(HaveOccurred())
		Ω(v).Should(Equal(`\//||\/||
// ||  ||__
`))
	})

	It("Decodes folded strings", func() {
		f, _ := os.Open("fixtures/specification/example2_15.yaml")
		d := NewDecoder(f)
		v := ""

		err := d.Decode(&v)
		Ω(err).ShouldNot(HaveOccurred())
		Ω(v).Should(Equal("Sammy Sosa completed another fine season with great stats.\n\n  63 Home Runs\n  0.288 Batting Average\n\nWhat a year!\n"))
	})

	It("Decodes literal and folded strings with indents", func() {
		f, _ := os.Open("fixtures/specification/example2_16.yaml")
		d := NewDecoder(f)
		v := make(map[string]string)

		err := d.Decode(&v)
		Ω(err).ShouldNot(HaveOccurred())
		Ω(v).Should(Equal(map[string]string{
			"name": "Mark McGwire",
			"accomplishment": `Mark set a major league home run record in 1998.
`,
			"stats": `65 Home Runs
0.278 Batting Average
`,
		}))
	})

	It("Decodes single quoted", func() {
		f, _ := os.Open("fixtures/specification/example2_17_quoted.yaml")
		d := NewDecoder(f)
		v := make(map[string]string)

		err := d.Decode(&v)
		Ω(err).ShouldNot(HaveOccurred())
		Ω(v).Should(Equal(map[string]string{
			"quoted": ` # not a 'comment'.`,
		}))
	})

	Context("ints", func() {
		It("Decodes into an interface{}", func() {
			f, _ := os.Open("fixtures/specification/example2_19.yaml")
			d := NewDecoder(f)
			v := make(map[string]interface{})

			err := d.Decode(&v)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(v).Should(Equal(map[string]interface{}{
				"canonical":   int64(12345),
				"decimal":     int64(12345),
				"sexagesimal": int64(12345),
				"octal":       int64(12),
				"hexadecimal": int64(12),
			}))
		})

		It("Decodes into int64", func() {
			f, _ := os.Open("fixtures/specification/example2_19.yaml")
			d := NewDecoder(f)
			v := make(map[string]int64)

			err := d.Decode(&v)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(v).Should(Equal(map[string]int64{
				"canonical":   int64(12345),
				"decimal":     int64(12345),
				"sexagesimal": int64(12345),
				"octal":       int64(12),
				"hexadecimal": int64(12),
			}))
		})

		Context("boundary values", func() {
			intoInt64 := func(val int64) {
				It("Decodes into an int64 value", func() {
					var v int64

					d := NewDecoder(strings.NewReader(strconv.FormatInt(val, 10)))
					err := d.Decode(&v)
					Ω(err).ShouldNot(HaveOccurred())
					Ω(v).Should(Equal(val))

				})
			}

			intoInt := func(val int) {
				It("Decodes into an int value", func() {
					var v int

					d := NewDecoder(strings.NewReader(strconv.FormatInt(int64(val), 10)))
					err := d.Decode(&v)
					Ω(err).ShouldNot(HaveOccurred())
					Ω(v).Should(Equal(val))

				})
			}

			intoInterface := func(val int64) {
				It("Decodes into an interface{}", func() {
					var v interface{}

					d := NewDecoder(strings.NewReader(strconv.FormatInt(val, 10)))
					err := d.Decode(&v)
					Ω(err).ShouldNot(HaveOccurred())
					Ω(v).Should(Equal(val))
				})
			}

			intoInt64(math.MaxInt64)
			intoInterface(math.MaxInt64)

			intoInt64(math.MinInt64)
			intoInterface(math.MinInt64)

			intoInt(math.MaxInt32)
			intoInt(math.MinInt32)
		})
	})

	It("Decodes a variety of floats", func() {
		f, _ := os.Open("fixtures/specification/example2_20.yaml")
		d := NewDecoder(f)
		v := make(map[string]float64)

		err := d.Decode(&v)
		Ω(err).ShouldNot(HaveOccurred())

		Ω(math.IsNaN(v["not a number"])).Should(BeTrue())
		delete(v, "not a number")

		Ω(v).Should(Equal(map[string]float64{
			"canonical":         float64(1230.15),
			"exponential":       float64(1230.15),
			"sexagesimal":       float64(1230.15),
			"fixed":             float64(1230.15),
			"negative infinity": math.Inf(-1),
		}))
	})

	It("Decodes booleans, nil and strings", func() {
		f, _ := os.Open("fixtures/specification/example2_21.yaml")
		d := NewDecoder(f)
		v := make(map[string]interface{})

		err := d.Decode(&v)
		Ω(err).ShouldNot(HaveOccurred())
		Ω(v).Should(Equal(map[string]interface{}{
			"":       interface{}(nil),
			"true":   true,
			"false":  false,
			"string": "12345",
		}))
	})

	It("Decodes dates/time", func() {
		f, _ := os.Open("fixtures/specification/example2_22.yaml")
		d := NewDecoder(f)
		v := make(map[string]time.Time)

		err := d.Decode(&v)
		Ω(err).ShouldNot(HaveOccurred())
		Ω(v).Should(Equal(map[string]time.Time{
			"canonical": time.Date(2001, time.December, 15, 2, 59, 43, int(1*time.Millisecond), time.UTC),
			"iso8601":   time.Date(2001, time.December, 14, 21, 59, 43, int(10*time.Millisecond), time.FixedZone("", -5*3600)),
			"spaced":    time.Date(2001, time.December, 14, 21, 59, 43, int(10*time.Millisecond), time.FixedZone("", -5*3600)),
			"date":      time.Date(2002, time.December, 14, 0, 0, 0, 0, time.UTC),
		}))
	})

	It("Respects tags", func() {
		f, _ := os.Open("fixtures/specification/example2_23_non_date.yaml")
		d := NewDecoder(f)
		v := make(map[string]string)

		err := d.Decode(&v)
		Ω(err).ShouldNot(HaveOccurred())
		Ω(v).Should(Equal(map[string]string{
			"not-date": "2002-04-28",
		}))
	})

	It("Decodes binary/base64", func() {
		f, _ := os.Open("fixtures/specification/example2_23_picture.yaml")
		d := NewDecoder(f)
		v := make(map[string][]byte)

		err := d.Decode(&v)
		Ω(err).ShouldNot(HaveOccurred())
		Ω(v).Should(Equal(map[string][]byte{
			"picture": []byte{0x47, 0x49, 0x46, 0x38, 0x39, 0x61, 0x0c, 0x00,
				0x0c, 0x00, 0x84, 0x00, 0x00, 0xff, 0xff, 0xf7, 0xf5, 0xf5, 0xee,
				0xe9, 0xe9, 0xe5, 0x66, 0x66, 0x66, 0x00, 0x00, 0x00, 0xe7, 0xe7,
				0xe7, 0x5e, 0x5e, 0x5e, 0xf3, 0xf3, 0xed, 0x8e, 0x8e, 0x8e, 0xe0,
				0xe0, 0xe0, 0x9f, 0x9f, 0x9f, 0x93, 0x93, 0x93, 0xa7, 0xa7, 0xa7,
				0x9e, 0x9e, 0x9e, 0x69, 0x5e, 0x10, 0x27, 0x20, 0x82, 0x0a, 0x01,
				0x00, 0x3b},
		}))
	})

	Context("Aliases", func() {
		Context("to known types", func() {
			It("aliases scalars", func() {
				f, _ := os.Open("fixtures/specification/example2_10.yaml")
				d := NewDecoder(f)
				v := make(map[string][]string)

				err := d.Decode(&v)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(v).Should(Equal(map[string][]string{
					"hr":  {"Mark McGwire", "Sammy Sosa"},
					"rbi": {"Sammy Sosa", "Ken Griffey"},
				}))
			})

			It("aliases sequences", func() {
				d := NewDecoder(strings.NewReader(`
---
hr: &ss
  - MG
  - SS
rbi: *ss
`))
				v := make(map[string][]string)
				err := d.Decode(&v)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(v).Should(Equal(map[string][]string{
					"hr":  {"MG", "SS"},
					"rbi": {"MG", "SS"},
				}))
			})

			It("aliases maps", func() {
				d := NewDecoder(strings.NewReader(`
---
hr: &ss
  MG : SS
rbi: *ss
`))
				v := make(map[string]map[string]string)
				err := d.Decode(&v)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(v).Should(Equal(map[string]map[string]string{
					"hr":  {"MG": "SS"},
					"rbi": {"MG": "SS"},
				}))
			})
		})

		Context("to Interface", func() {
			It("aliases scalars", func() {
				f, _ := os.Open("fixtures/specification/example2_10.yaml")
				d := NewDecoder(f)
				v := make(map[string]interface{})

				err := d.Decode(&v)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(v).Should(Equal(map[string]interface{}{
					"hr":  []interface{}{"Mark McGwire", "Sammy Sosa"},
					"rbi": []interface{}{"Sammy Sosa", "Ken Griffey"},
				}))
			})

			It("aliases sequences", func() {
				d := NewDecoder(strings.NewReader(`
---
hr: &ss
  - MG
  - SS
rbi: *ss
`))
				v := make(map[string]interface{})
				err := d.Decode(&v)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(v).Should(Equal(map[string]interface{}{
					"hr":  []interface{}{"MG", "SS"},
					"rbi": []interface{}{"MG", "SS"},
				}))
			})

			It("aliases maps", func() {
				d := NewDecoder(strings.NewReader(`
---
hr: &ss
  MG : SS
rbi: *ss
`))
				v := make(map[string]interface{})
				err := d.Decode(&v)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(v).Should(Equal(map[string]interface{}{
					"hr":  map[interface{}]interface{}{"MG": "SS"},
					"rbi": map[interface{}]interface{}{"MG": "SS"},
				}))
			})
		})
	})

	Context("When decoding fails", func() {
		It("returns an error", func() {
			f, _ := os.Open("fixtures/specification/example_empty.yaml")
			d := NewDecoder(f)
			var v interface{}

			err := d.Decode(&v)
			Ω(err).Should(HaveOccurred())
		})
	})

	Context("Unmarshaler support", func() {
		Context("Receiver is a value", func() {
			It("the Marshaler interface is not used", func() {
				d := NewDecoder(strings.NewReader("abc\n"))
				v := hasMarshaler{}

				err := d.Decode(&v)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(v.Value).Should(BeNil())
			})
		})

		Context("Receiver is a pointer", func() {
			It("uses the Marshaler interface when a pointer", func() {
				d := NewDecoder(strings.NewReader("abc\n"))
				v := hasPtrMarshaler{}

				err := d.Decode(&v)
				Ω(err).ShouldNot(HaveOccurred())
			})

			It("marshals a scalar", func() {
				d := NewDecoder(strings.NewReader("abc\n"))
				v := hasPtrMarshaler{}

				err := d.Decode(&v)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(v.Tag).Should(Equal("!!str"))
				Ω(v.Value).Should(Equal("abc"))
			})

			It("marshals a sequence", func() {
				d := NewDecoder(strings.NewReader("[abc, def]\n"))
				v := hasPtrMarshaler{}

				err := d.Decode(&v)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(v.Tag).Should(Equal("!!seq"))
				Ω(v.Value).Should(Equal([]interface{}{"abc", "def"}))
			})

			It("marshals a map", func() {
				d := NewDecoder(strings.NewReader("{ a: bc}\n"))
				v := hasPtrMarshaler{}

				err := d.Decode(&v)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(v.Tag).Should(Equal("!!map"))
				Ω(v.Value).Should(Equal(map[interface{}]interface{}{"a": "bc"}))
			})
		})
	})

	Context("Marshals into a Number", func() {
		It("when the number is an int", func() {
			d := NewDecoder(strings.NewReader("123\n"))
			d.UseNumber()
			var v Number

			err := d.Decode(&v)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(v.String()).Should(Equal("123"))
		})

		It("when the number is an float", func() {
			d := NewDecoder(strings.NewReader("1.23\n"))
			d.UseNumber()
			var v Number

			err := d.Decode(&v)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(v.String()).Should(Equal("1.23"))
		})

		It("it fails when its a non-Number", func() {
			d := NewDecoder(strings.NewReader("on\n"))
			d.UseNumber()
			var v Number

			err := d.Decode(&v)
			Ω(err).Should(HaveOccurred())
		})

		It("returns a Number", func() {
			d := NewDecoder(strings.NewReader("123\n"))
			d.UseNumber()
			var v interface{}

			err := d.Decode(&v)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(v).Should(BeAssignableToTypeOf(Number("")))

			n := v.(Number)
			Ω(n.String()).Should(Equal("123"))
		})
	})
})
