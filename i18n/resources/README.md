# DO NOT CHANGE ANYTHING IN THIS DIRECTORY
This directory will get populated by the pipeline during the `build-binaries` job. Do NOT modify otherwise.

## How this file was generated
```
$ cd code.cloudfoundry.org/cli/i18n
$ go-bindata -nometadata  -pkg resources -ignore ".go" -o resources/i18n_resources.go resources/*.all.json
```
