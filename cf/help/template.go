package help

import . "code.cloudfoundry.org/cli/cf/i18n"

func GetHelpTemplate() string {
	return `{{.Title "` + T("NAME:") + `"}}
   {{.Name}} - {{.Usage}}

{{.Title "` + T("USAGE:") + `"}}
   ` + `{{.Name}} ` + T("[global options] command [arguments...] [command options]") + `

{{.Title "` + T("VERSION:") + `"}}
   {{.Version}}
   {{range .Commands}}
{{.SubTitle .Name}}{{range .CommandSubGroups}}
{{range .}}   {{.Name}} {{.Description}}
{{end}}{{end}}{{end}}
{{.Title "` + T("ENVIRONMENT VARIABLES:") + `"}}
   CF_COLOR=false                     ` + T("Do not colorize output") + `
   CF_HOME=path/to/dir/               ` + T("Override path to default config directory") + `
   CF_DIAL_TIMEOUT=5                  ` + T("Max wait time to establish a connection, including name resolution, in seconds") + `
   CF_PLUGIN_HOME=path/to/dir/        ` + T("Override path to default plugin config directory") + `
   CF_STAGING_TIMEOUT=15              ` + T("Max wait time for buildpack staging, in minutes") + `
   CF_STARTUP_TIMEOUT=5               ` + T("Max wait time for app instance startup, in minutes") + `
   CF_TRACE=true                      ` + T("Print API request diagnostics to stdout") + `
   CF_TRACE=path/to/trace.log         ` + T("Append API request diagnostics to a log file") + `
   https_proxy=proxy.example.com:8080 ` + T("Enable HTTP proxying for API requests") + `

{{.Title "` + T("GLOBAL OPTIONS:") + `"}}
   --help, -h                         ` + T("Show help") + `
   -v                                 ` + T("Print API request diagnostics to stdout") + `
`
}
