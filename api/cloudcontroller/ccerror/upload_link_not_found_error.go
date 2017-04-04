package ccerror

import "fmt"

type UploadLinkNotFoundError struct {
	PackageGUID string
}

func (e UploadLinkNotFoundError) Error() string {
	return fmt.Sprintf("Upload link not found in for package with GUID %s", e.PackageGUID)
}
