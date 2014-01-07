package fixtures

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
applications:
- name: hacker-manifesto
  command: null
  space_guid: null
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
  buildpack: some-buildpack
  command: JAVA_HOME=$PWD/.openjdk JAVA_OPTS="-Xss995K" ./bin/start.sh run
  path: ../../fixtures/example-app
  env:
    FOO: baz
`,
}

func FixtureWithName(name string) (fixture string) {
	return fixtureMap[name]
}
