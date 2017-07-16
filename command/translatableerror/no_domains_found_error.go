package translatableerror

import "fmt"

type NoDomainsFoundError struct {
}

func (NoDomainsFoundError) Error() string {
	return fmt.Sprintf("No private or shared domains found in this organization")
}

func (e NoDomainsFoundError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error())
}
