package actionerror

import "fmt"

type SecurityGroupJsonSyntaxError struct {
	Path string
}

func (e SecurityGroupJsonSyntaxError) Error() string {
	return fmt.Sprintf(`Incorrect json format: %s

Valid json file example:
[
  {
    "protocol": "tcp",
    "destination": "10.244.1.18",
    "ports": "3306"
  }
]
`, e.Path)
}
