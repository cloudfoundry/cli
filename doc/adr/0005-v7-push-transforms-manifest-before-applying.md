# 5. V7 `cf push` Transforms Manifest Before Applying

Date: 2019-09-16

## Status

Accepted

## Context

> _The issue motivating this decision, and any context that influences or constrains the decision._

**NOTE:** This ADR builds on ideas introduced in [the last `push` refactor](/doc/adr/0004-v7-push-refactor.md). If you haven't read that ADR, please do so before reading this one.

As the V3 Acceleration Team was working on the V7 CLI version of `cf push`, we found that the implementation made certain features difficult (or impossible) to implement. We refactored the first part of `cf push` as described below to solve some of these problems.

### The dream of the server-side manifest

In December 2018, the CAPI, CLI, and V3 Acceleration teams began discussing the idea of "server-side manifests", in which CAPI would build endpoints that received a manifest and made the necessary changes on the API side. Previously, the `push` manifest file was a concept invented and implemented fully by the CLI.

The results of this work were the V3 [app manifest](https://v3-apidocs.cloudfoundry.org#app-manifest) and [space manifest](https://v3-apidocs.cloudfoundry.org#space-manifest) endpoints.

One part of the dream of server-side manifests was that the CLI would no longer need to have any knowledge of manifest contents. The `push` command would pass manifest contents directly to the API without ever parsing the YAML. The API would apply the manifest as given, and the CLI would then upload source code for staging.

The goal of this was to allow CAPI to add additional manifest features (i.e. properties) without requiring any implementation changes on the CLI side.

See here for more on this topic, including the original goals of server-side manifests:
* [V3 Server-Side Multi-App Manifests](https://docs.google.com/document/d/1z-0Ev-QCtuoT8nJCoNaVJmkMPNk4PYi_VGWzc8CwBIs/edit#)
* [Server Side Manifest Exploration](https://docs.google.com/document/d/1HKNz5Qaza9fx8QSQp4oWqzcg3Sv1tkyEhm_6qQFaQAM/edit#heading=h.qt2a8po833hn)

### Adding support for diffing

At this point in time, V7 `push` does not print diff output like it does in V6. This is a feature of V6 `push` that we got a lot of positive feedback about, and it is the last piece of `push` we need to implement for parity with V6.

The diff output is meant to show the difference between the current state of the pushed apps and the desired state (as represented by the manifest and flag overrides), like this:

```diff
$ cf push -i 3
Pushing from manifest to org org / space space as admin...
Using manifest file /home/pivotal/workspace/cf-acceptance-tests/assets/dora/manifest.yml
Getting app info...
Updating app with these attributes...
  name:                dora
  path:                /home/pivotal/go/src/github.com/cloudfoundry/cf-acceptance-tests/assets/dora
  command:             bundle exec rackup config.ru -p $PORT
  disk quota:          1G
  health check type:   port
- instances:           2
+ instances:           3
  memory:

  stack:               cflinuxfs3
  routes:
    dora.sheer-shark.lite.cli.fun
```

When we set out to explore adding the diff output to V7, we found there was **no straightforward way to do so**. In V6, we were already parsing the whole manifest and "actualizing" each property individually, but with V7, we intentionally didn't parse the whole manifest, so we did not have knowledge of the full desired state. Without this knowledge, we could not do the comparison necessary to output the diff.

### Unfixable bugs due to overrides being handled too late

There was a whole class of bugs that could not be fixed in V7 `push` before this refactor. These were cases where a manifest property value is invalid, but the corresponding flag override is acceptable. For example, you might have `memory_in_mb: 2000GB` in your manifest, which is over your quota, but you push with `-m 20MB`, which is fine.

Prior to this change, manifest properties and flag overrides were handled in separate API calls. Manifest properties were applied first exactly as they are in the manifest file. Then if the apply-manifest request was successful, the flag overrides would be handled (e.g. by sending a request to scale the memory to the number specified by the `-m` flag). But if the manifest was invalid (e.g. had a memory setting that would put an app over its quota), the apply-manifest step would fail before even getting to the override step.

We believe this was confusing and unexpected behavior for users: if a user passes `-m 20MB`, they would expect that to be the only number that mattered (not the memory value they're overriding in the manifest).

Here is an example of such a bug:
* [**CLI** user should not get a failed push when flag overrides are below quota limits but manifest values are above](https://www.pivotaltracker.com/story/show/167576661)

## Decision

> _The change that we're proposing or have agreed to implement._

We refactored the first half of V7 `push` (all of the steps that come before uploading source code and creating a new droplet). Read on for details.

### How did `push` work before this change?

Before this change, `push` preserved manifest properties exactly as-is and sent these to the apply-manifest endpoint.
The flag overrides would be handled after applying the manifest was complete.

### How does `push` work after this change?

Now, `push` parses the manifest into an in-memory representation of the manifest. It transforms its representation of the manifest based on the given flag overrides, and then generates new YAML to send to the apply-manifest endpoint.

The effect of this is that we eliminate the extra steps of "fixing" the app after the apply-manifest step runs. For example, we no longer need to send additional requests to scale an app's process when the `-i 4` flag is given. Instead, the manifest that we send to the API will have been modified to say `instances: 4`. See below for a detailed explanation of how this works.

In addition, the API is now fully responsible for handling the logic of validating and resolving conflicts between manifest properties.

### Case study: Lifecycle of a flag override

To see how this is implemented, we will follow the path that a flag override takes from the user, through the code, to the API.

Imagine a user runs `cf push -i 4`, with a manifest that looks like:

```yaml
applications:
- name: dora
  instances: 2
```

The `PushCommand` parses the manifest as-written and the given flag overrides. It passes them into `HandleFlagOverrides`, which returns a transformed manifest:

```go
func (cmd PushCommand) Execute(args []string) error {
    //...
    transformedManifest, err := cmd.Actor.HandleFlagOverrides(baseManifest, flagOverrides)
    if err != nil {
        return err
    }
    //...
}
```

The `HandleFlagOverrides` method is another example of the hexagonal pattern established in [this refactor](/doc/adr/0004-v7-push-refactor.md). It passes the manifest through a sequence of functions (`TransformManifestSequence`):

```go
func (actor Actor) HandleFlagOverrides(baseManifest pushmanifestparser.Manifest, flagOverrides FlagOverrides) (pushmanifestparser.Manifest, error) {
	newManifest := baseManifest

	for _, transformPlan := range actor.TransformManifestSequence { // <== sequence of transform functions
		var err error
		newManifest, err = transformPlan(newManifest, flagOverrides)
		if err != nil {
			return pushmanifestparser.Manifest{}, err
		}
	}

	return newManifest, nil
}
```

Here is the list of functions in the transform sequence:

```go
actor.TransformManifestSequence = []HandleFlagOverrideFunc{
    // app name override must come first, so it can trim the manifest
    // from multiple apps down to just one
    HandleAppNameOverride,

    HandleInstancesOverride, // <== instances transform function!
    HandleStartCommandOverride,
    HandleHealthCheckTypeOverride,
    HandleHealthCheckEndpointOverride,
    HandleHealthCheckTimeoutOverride,
    HandleMemoryOverride,
    HandleDiskOverride,
    HandleNoRouteOverride,
    HandleRandomRouteOverride,

    // this must come after all routing related transforms
    HandleDefaultRouteOverride,

    HandleDockerImageOverride,
    HandleDockerUsernameOverride,
    HandleStackOverride,
    HandleBuildpacksOverride,
    HandleStrategyOverride,
    HandleAppPathOverride,
    HandleDropletPathOverride,
}
```

This is the `HandleInstancesOverride` method, which is responsible for transforming the manifest based on the `-i/--instances` flag, if given:

```go
func HandleInstancesOverride(manifest pushmanifestparser.Manifest, overrides FlagOverrides) (pushmanifestparser.Manifest, error) {
    if overrides.Instances.IsSet {
        if manifest.ContainsMultipleApps() {
            return manifest, translatableerror.CommandLineArgsWithMultipleAppsError{}
        }

        webProcess := manifest.GetFirstAppWebProcess()
        if webProcess != nil {
            webProcess.Instances = &overrides.Instances.Value
        } else {
            app := manifest.GetFirstApp()
            app.Instances = &overrides.Instances.Value
        }
    }

    return manifest, nil
}
```

After the manifest is transformed, we generate the following new YAML that is sent to the apply-manifest endpoint:

```yaml
applications:
- name: dora
  instances: 4 # <== updated!
```

From the API's point of view, there are no "flag overrides". The override behavior is resolved fully on the CLI side before the manifest is ever applied.

## Consequences

> _What becomes easier or more difficult to do and any risks introduced by the change that will need to be mitigated._

### Cleaner separation of concerns between CLI and API

One key outcome is that we have achieved a clearer separation of concerns between the CLI and the API when it comes to the manifest. The idea of "flag overrides" has always been a CLI-only concept, but there were cases where the responsibilities were getting blurred.

For example, consider `no-route`, which is available as a flag (`--no-route`) or as a manifest property (`no-route: true`). Before this refactor, in order to apply the manifest exactly as written but still allow the overriding behavior that the CLI required, we introduced `no_route` as a query parameter on the apply-manifest endpoint. So, when `--no-route` was given, the resulting request would look like `POST /v3/spaces/:guid/apply_manifest?no_route=true`. Although this worked, it was an example of a specific, one-off solution to a broader problem. We did not want to add query parameters for every overridable manifest property, and it was an instance of making the API overly-tailored to the CLI's use case.

With this change, the `--no-route` override is handled fully on the CLI side, before applying the manifest, since flag overrides are the CLI's business. We were able to remove the `no_route` query parameter from the API endpoint and clean up some related code on the API side.

### Server-side manifests: closer to or further from the dream?

If the goal of server-side manifests was to push as much as possible of the work related to validating/applying/resolving configuration changes to the API side, then this refactor gets us much closer to that goal. The CLI now leans fully on the API to apply the manifest, rather than applying the manifest and then "correcting" the configuration with additional API calls.

However, if the goal of server-side manifests was for the CLI to know as little as possible about the contents of the manifest, then this refactor is a deliberate departure from that goal. In order to apply the flag overrides, the CLI must parse almost the entire manifest.

#### Risks

With server-side manifests, we wanted to allow CAPI to introduce new manifest properties that could be leveraged by users (in their manifests) without requiring code changes in the CLI. The **biggest risk** is that after this refactor, CAPI could make a change to the manifest specification that would break users' `push` experience until the CLI releases a new version.

#### Mitigations

We accept the risk described above for these reasons:
- We designed the CLI's new manifest parser to preserve any unrecognized YAML properties in the in-memory representation of the YAML, so they will be sent along to the API in the end. This means CAPI is free to make additive changes to the manifest specification, and users can leverage them without needing any CLI changes.
- If CAPI makes a _breaking_ change to the manifest specification, this will impact the CLI. However, we discussed this with the CAPI team, and they are unlikely to do so anytime soon, since it will also force any users to refactor their manifests (and require changes in any other clients dependent on the manifest spec). If they do want to make a breaking change, they will likely introduce the idea of _versioned manifests_ and continue to support multiple manifest specifications for some time.

### Fewer leftovers when `push` fails
Before the refactor, failed pushes could result in state changes. These changes would not get rolled back after the `push` exited with an error.

Setup:
Suppose the CLI user started with no existing app and a space memory quota of 500 MB.
If they supplied a manifest:
```yaml
applications:
- name: dora
  instances: 50
  memory_in_mb: 10
```

And they pushed with this command:
`cf push -m 20MB`


Before this refactor:
`push` would apply the manifest to create 50 instances, with 10 MB allotted per instance.  In this case, the allotted memory would be within the quota, so the manifest would be applied successfully.
Once apply manifest succeeded, a separate API call would apply the flag override to allot 20MB per instance instead.  However in this case, `push` would fail with an error indicating that the desired memory is over quota.
However the 50 instances created would stick around despite the failed `push`.


After this refactor:
`push` transforms the in-memory representation of the manifest to the following:
```yaml
applications:
- name: dora
  instances: 50
  memory_in_mb: 20
```
`push` would apply the manifest, and would fail with a validation error indicating that the desired memory is over quota.  Therefore we never even get to the step of scaling app instances.

### Fewer API requests per app

A nice side-effect of this refactor is that the number of API calls required to push an app decreases considerably. Previously, to apply the manifest properties, we had to make one call to apply the manifest and then `n` calls for each app that we're pushing (where `n` is roughly the number of flag overrides). Now, we only have to make one call to apply the manifest for all pushed apps.

This means that `push` is faster, more resilient to flaky network connections (less likely to fail in the middle due to a poor connection), and requires less code to implement.

### Maintainability

This refactor builds on the hexagonal architecture established in [this push refactor](/doc/adr/0004-v7-push-refactor.md).  This makes `push` more consistent and easier to reason about.

### Diff possibility

Now, we have a simpler path towards implementing the diff output feature in V7. We now have a representation of the full desired manifest, after flag overrides have been applied. We can make [a request to the API for the current manifest](https://v3-apidocs.cloudfoundry.org/version/release-candidate/#generate-an-app-manifest) and compare the two manifests.

### Better captures user intent of flag overrides superceding manifest properties

We believe that if a user runs `cf push -i 4` with a manifest that looks like this:
```yaml
applications:
- name: dora
  instances: 1
```
...then their intention is really to push with 4 instances. Before the refactor, we would create one instance first, and then scale it up to 4. With the refactor, we feel we are better capturing user intent from the outset (without the intermediate step where an "incorrect" manifest is applied).

