package v7_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	uuid "github.com/nu7hatch/gouuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
)

func TestV3(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "V7 Command Suite")
}

var _ = BeforeEach(func() {
	log.SetLevel(log.PanicLevel)
})

// RandomString provides a random string
func RandomString(prefix string) string {
	guid, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}

	return prefix + "-" + guid.String()
}

func setFlag(cmd interface{}, flag string, values ...interface{}) {
	var value reflect.Value
	switch len(values) {
	case 0:
		value = reflect.ValueOf(true)
	case 1:
		value = reflect.ValueOf(values[0])
	default:
		Fail(fmt.Sprintf("cannot take more than one value for flag '%s'", flag))
	}

	var key, trimmedFlag string
	switch {
	case strings.HasPrefix(flag, "--"):
		trimmedFlag = strings.TrimPrefix(flag, "--")
		key = "long"
	case strings.HasPrefix(flag, "-"):
		trimmedFlag = strings.TrimPrefix(flag, "-")
		key = "short"
	default:
		Fail("flag must start with prefix '--' or '-'")
	}

	ptr := reflect.ValueOf(cmd)
	if ptr.Kind() != reflect.Ptr {
		Fail("need to pass a pointer to the command struct")
	}

	val := ptr.Elem()
	if val.Kind() != reflect.Struct {
		Fail("need to pass a command struct")
	}

	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		if tagValue, ok := field.Tag.Lookup(key); ok {
			if tagValue == trimmedFlag {
				switch {
				case value.Type().ConvertibleTo(field.Type):
					val.Field(i).Set(value.Convert(field.Type))
					return
				case reflect.PtrTo(field.Type).Implements(reflect.TypeOf((*interface{ UnmarshalFlag(string) error })(nil)).Elem()):
					elem := reflect.New(field.Type)
					e := elem.MethodByName("UnmarshalFlag").Call([]reflect.Value{value})
					if !e[0].IsNil() {
						Fail(fmt.Sprintf(
							"Could not set field '%s' type '%s' to '%v' type '%s' for flag '%s' because %s",
							field.Name,
							field.Type,
							value.Interface(),
							value.Type(),
							flag,
							e[0].Interface(),
						))
					}
					val.Field(i).Set(elem.Elem())
					return
				default:
					Fail(fmt.Sprintf(
						"Could not set field '%s' type '%s' to '%v' type '%s' for flag '%s'",
						field.Name,
						field.Type,
						value.Interface(),
						value.Type(),
						flag,
					))
				}
			}
		}
	}

	Fail(fmt.Sprintf("could not find flag '%s' in command struct", flag))
}

func setPositionalFlags(cmd interface{}, values ...interface{}) {
	ptr := reflect.ValueOf(cmd)
	if ptr.Kind() != reflect.Ptr {
		Fail("need to pass a pointer to the command struct")
	}

	val := ptr.Elem()
	if val.Kind() != reflect.Struct {
		Fail("need to pass a command struct")
	}

	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		if tagValue, ok := field.Tag.Lookup("positional-args"); ok && tagValue == "yes" && field.Type.Kind() == reflect.Struct {
			if len(values) != field.Type.NumField() {
				Fail(fmt.Sprintf("%d values provided but positional args struct %s has %d fields", len(values), field.Name, field.Type.NumField()))
			}

			for j := 0; j < field.Type.NumField(); j++ {
				posField := field.Type.Field(j)
				value := reflect.ValueOf(values[j])
				if value.Type().ConvertibleTo(posField.Type) {
					val.Field(i).Field(j).Set(value.Convert(posField.Type))
				} else {
					Fail(fmt.Sprintf(
						"Could not set field '%s' type '%s' to '%v' type '%s'",
						posField.Name,
						posField.Type,
						value.Interface(),
						value.Type(),
					))
				}
			}

			return
		}
	}

	Fail(`Did not find a field with 'positional-args:"yes"' in the struct`)
}
