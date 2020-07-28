# 11. Rules for organizing `command`, `actor`, and `api` code

Date: 2020-07-28

## Status

Submitted

## Context

Several years ago, the CLI team began refactoring the legacy codebase in place, converting commands one at a time. When the V3 Acceleration Team (VAT) was created to focus on finishing V7 of the CLI, which would use primarily CAPI V3 endpoints, we picked up the patterns introduced as part of these refactors and began applying them to the new V7 commands that we ported from V6.

The architecture that we inherited looked roughly like this:

- The "command layer" lives under the `command` directory. The `command/v7` package contains one file/struct per V7 command (e.g. `v7.SpaceQuotaCommand`).
- The "actor layer" lives under the `action` directory. The `action/v7action` package contains roughly one file per API resource with methods related to that resource (e.g. `droplets.go`).
- The "API layer" lives under the `api` directory. The `api/cloudcontroller/ccv3` package contains one file per API resource with methods related to hitting CAPI V3 endpoints.

After porting a large number of commands to this architecture, VAT had accumulated a number of small frustrations. There were ambiguities with this architecture that left us asking each other the same questions over and over:

- When should logic live in the command layer vs. the actor layer?
- Why do we have a `ccv3.Space` struct and a `v7action.Space` struct, when they are very similar (often literally aliases of one another)?
- If we have actor methods that call other actor methods, how should we unit test these? We can't mock the actor methods at this level, so we have to mock the underlying API methods, which makes tests harder to reason about.

## Decision

We decided that we needed to come up with a **stricter set of rules** that we could overlay onto the current architecture to help us resolve these questions so we don't need to keep revisiting them.

These are the rules that we decided on. Note that most of the patterns described here were already used in various parts of the codebase, but they were not used consistently. You might think of this as picking-and-choosing the "greatest hits" from all of the pieces of CLI commands that we've written.

### API-layer rules & responsibilities

The API layer (also called the "client layer") is the simplest. This layer contains packages designed to hit specific APIs. The most commonly used one is `api/cloudcontroller/ccv3`, which contains methods that hit the CAPI V3 API.

Each method in this layer should do **one thing**. Usually, this means **make a single request** to an API endpoint. Inputs to these methods should correspond with inputs that the endpoint takes, and return values should correspond with the return values of the endpoint.

Most of the time, think of this layer as a 1:1 mapping from API endpoints to client methods.

However, it is permissible to add "helper" methods at this layer, as long as they are making a single request. A good example is `GetApplicationByNameAndSpaceGUID`.  This sort of helper is needed in lots of actor methods, but it doesn't map directly to a single endpoint; instead, it maps to `GET /v3/applications?names=<name>&space_guids=<guid>`. It is good to define this method at the API layer because it means we can easily mock it out in unit tests of the actor-layer methods that use it.

### Resource-layer rules & responsibilities

The resource layer is the second-simplest layer. It is a **new layer**, extracted from the API layer. It lives in the `resources` directory.

Previously, we had structs modeling each API resource defined in the API layer (e.g. `ccv3.SpaceQuota`). API methods would return them to the actor layer. The actor layer would then convert them into actor-layer struct types. Often, this was just done using type aliases:

```go
package v7action

// ...

type SpaceQuota ccv3.SpaceQuota

// ...

ccv3SpaceQuota := actor.CloudControllerClient.GetSpaceQuota(guid)

return SpaceQuota(ccv3SpaceQuota) // have to cast it
```

We felt this was an unnecessary amount of indirection and abstraction (_why do we need the struct defined in two places?_). Nobody could think of a strong reason for it. Plus, it made API layer unit tests fairly large and hard to read, since they were often testing both logic and JSON marshalling/unmarshalling in the same tests.

By extracting the structs to the resource layer, we have a **single shared struct definition for each API resource**. This is shared by all the other layers (API, command, actor).

Structs defined at this layer should **map directly to API resources**. No data should be present on these resources that did not come back from the primary API endpoint that returned them. (See [the summary pattern](#the-summary-pattern), in the actor-layer, for an example of how to do this correctly.)

This layer should also contain **all JSON marshalling/unmarshalling logic**.  This lets us unit-test this behavior individually, in resource-layer unit tests, without needing to clutter the API-layer tests.

To simplify field access, resource structs should be **as flat as possible**.  For example, on `resources.Application`, we have a `SpaceGUID` field. In the V3 API JSON response, this data lives at `app.relationships.space.data.guid`. We should handle all the nesting/unnesting of fields like this in the marshal/unmarshal logic, so that other layers don't need to know or care about the specific JSON representation that the API gives us. As a rule of thumb, the word "relationships" should never appear outside of the resource layer, since this is an API implementation detail.

For handling complex marshalling/unmarshalling logic, the [`jsonry` library](https://github.com/cloudfoundry/jsonry) will come in quite handy. We expect the resource layer to make heavy use of this package.

Finally, this layer **should not contain methods that make API requests**. Those should live at the API layer.

### Command-layer rules & responsibilities

The command layer is responsible for **taking user input** and **presenting output to the user**. The former involves things like defining and validating flags/positional arguments/environment variables. The latter involves formatting and printing messages to stdout and stderr. In between those, commands call out to the actor layer to do the "real" work.

Most commands should only make a **single call to an actor-layer method**. A command will usually give some indication of what it's doing (e.g. `Creating app my-app in space my-space / org my-org as some-user...`), and then give an indication that the action succeeded or failed (e.g. `OK`, a table of data, or an error message).

An exception to that rule is commands that perform multiple "steps." For example, `create-org` will first create an org, and then will assign the current user as an org manager in that org. These two discrete steps are both presented to the user, and they both display an `OK` on successful completion, so they should be two separate actor-layer calls.

See the [style guide](https://github.com/cloudfoundry/cli/wiki/CF-CLI-Style-Guide) for guidelines on how to format command output.

### Actor-layer rules & responsibilities

The actor layer has historically been the hardest layer to explain. It is easiest to start by describing all the things it _shouldn't_ be responsible for; see the preceding sections for that.

So what is left? This layer is primarily responsible for **taking input from the command layer**, **orchestrating multiple API-layer calls**, and (when necessary) **packaging multiple responses into resource-summary structs** to be returned to the command layer.

Actor-layer methods generally **should not wrap only a single API-layer call**.  If there is an actor method that is only making a single call, it is a sign that we are probably doing too much work at the command layer (where it is being called).  Consider putting some of this work in this actor method.

A good example is something like `ApplyOrgQuota(orgName string, quotaName string)`. This method takes **names** of resources, not guids, because that's typically what the user has provided to the command layer. This method will probably need to make a series of calls to achieve its single goal: fetch the org by name to retrieve the org guid, fetch the quota by name to retrieve the quota guid, then apply the quota in a request using these two guids. From the command layer's point-of-view, this method is doing only one thing (applying an org quota). But it needed to orchestrate three API calls to do it.

#### The summary pattern

Another common category of actor methods follow the **summary pattern**. These methods are used by "read" commands (e.g. `space`) and might be called something like `GetSpaceSummary`. Typically, these commands fetch a single primary resource (like a space), plus a handful of related resources (e.g. routes in that space, the space's parent org). The actor layer is then responsible for defining a struct type to contain all of this information. These actor methods will return a summary struct back to the command layer, which will handle displaying that information appropriately.

Here's what a summary struct might look like:

```go
package v7action

type SpaceSummary struct {
    Space    resources.Space
    Routes   []resources.Route
    Org      resources.Organization
}
```

Note that it is composed of several resources (defined in the `resources` package, at the resources layer). But since each of these resources are fetched from different API endpoints, no single `resources` struct can contain all of it; it is the job of the actor layer to assemble them into a summary struct.

## Consequences

Looked at one way, these rules classify existing code as _doing it wrong_ or _doing it right_. This is **not** the intended takeaway. We do not believe it is valuable to go back and refactor every place that is not adhering to these rules, because most often these things amount to minor cosmetic differences. Plus, like all rules, these are not perfect and will not apply 100% of the time. If there is a strong reason to deviate from them, developers should not feel guilty about it. These rules only aim to resolve confusion if it exists. We do not want to add additional confusion. If you never need to refer to these rules, that's fine! But they're here if you need them.

The primary goal here is that **going forward**, when questions arise, we have a framework for answering them. When a new teammate joins, we can point them to these rules as a starting point. When external contributors want to understand the codebase, they can refer to these rules.

### Pain points addressed

All that said, let's look at how these rules address some of the specific pain points we were feeling:

- The responsibilities of the actor layer are now much clearer. Several patterns that were sometimes seen at the actor layer have now been "moved" either to the command or to the API layer.
- By explicitly allowing "helper" methods to exist at the API layer, we can more often avoid having actor methods call other actor methods. This prevents the problem of being unable to mock these calls in unit tests. As much as possible, actors now only call API methods.
- By introducing the resource layer, we eliminate the need for separate resource structs that live at the actor and API layers. Now it's much simpler: every layer can refer to the shared resources. The new "summary struct" pattern handles cases where an actor needs to combine data from multiple resources.

## Links & resources

Here's a visual representation (a screenshot of the Miro board linked below). This may grow slightly out-of-date as time goes on, so this document and the linked Miro board should be considered more reliable sources of truth.

![diagram of proposed architecture](/doc/cli-architecture-proposal-adr-0011.jpg)

Some related links:

- The [Miro board](https://miro.com/app/board/o9J_kvBwLTE=/) where these ideas were first written down (and where this image is from).
