package maker

var fixtureMap = map[string]string{
	"merged services": `
---
services:
- global-service
applications:
- name: app-with-redis-backend
  services:
  - nested-service
- name: app2
  services:
  - app2-service
`,

	"local services": `
---
applications:
- name: app-with-redis-backend
  services:
  - work-queue
`,

	"global services": `
---
services:
- work-queue
applications:
- name: app-with-redis-backend
`,

	"many apps": `
---
env:
  PATH: /u/apps/something/bin
  SOMETHING: nothing
applications:
- name: app1
  env:
    SOMETHING: definitely-something
- name: app2
`,

	"nulls": `
---
applications:
- name: hacker-manifesto
  command: null
  buildpack: null
  disk_quota: null
  instances: null
  memory: null
  env: null
`,

	"single app": `
---
env:
  PATH: /u/apps/my-app/bin
  FOO: bar
applications:
- name: manifest-app-name
  memory: 128M
  instances: 1
  host: manifest-host
  domain: manifest-example.com
  stack: custom-stack
  timeout: 360
  buildpack: some-buildpack
  command: JAVA_HOME=$PWD/.openjdk JAVA_OPTS="-Xss995K" ./bin/start.sh run
  path: ../../fixtures/example-app
  env:
    FOO: baz
`,

	"app with absolute unix path": `
---
env:
  PATH: /u/apps/my-app/bin
  FOO: bar
applications:
- name: manifest-app-name
  path: /absolute/path/to/example-app
`,

	"app with absolute windows path": `
---
env:
  PATH: /u/apps/my-app/bin
  FOO: bar
applications:
- name: manifest-app-name
  path: C:\absolute\path\to\example-app
`,

	"invalid": `
---
env:
- PATH
- USER
services:
  old-service-format:
    plan: free
    provider: nobody
    type: deprecated
applications:
- name: bad-services
  services:
    old-service-format:
      plan: paid
      provider: somebody
      type: deprecated
- name: bad-env
  env:
  - FOO
  - BAR
`,
	"invalid env": `
---
applications:
- name: app-name
env:
  foo: bar
  bar:
`,
}

func ManifestWithName(name string) (fixture string) {
	return fixtureMap[name]
}
