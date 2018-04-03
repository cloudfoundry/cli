package manifest

import (
	"fmt"
)

type ReleaseRef struct {
	Name string
	URL  string
	SHA1 string
}

func (r ReleaseRef) GetURL() string  { return r.URL }
func (r ReleaseRef) GetSHA1() string { return r.SHA1 }

func (r ReleaseRef) Description() string {
	return fmt.Sprintf("release '%s'", r.Name)
}
