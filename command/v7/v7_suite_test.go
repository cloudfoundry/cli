package v7_test

import (
	"fmt"
	"reflect"
	"strings"

	uuid "github.com/nu7hatch/gouuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"

	"testing"
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
				if value.Type().ConvertibleTo(field.Type) {
					val.Field(i).Set(value.Convert(field.Type))
					return
				}
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

	Fail(fmt.Sprintf("could not find flag '%s' in command struct", flag))
}
