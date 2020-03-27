# ADR 6: Implementing start Command Client-side

## Status

Accepted

## Abstract

The V7 CLI's start command must function similarly to V6's `cf start` command in that it _both_ stages and starts an app after a `cf push --no-start` (see [story](https://www.pivotaltracker.com/story/show/162463497)) using only V3 capi endpoints (see [design board](https://miro.com/app/board/o9J_kykcZyc=/)). The decision whether to implement this feature in CAPI or CLI posed challenges, but ultimately the decision was to implement this in the CLI.

## Decision

We will implement staging unstaged packages during a `cf start` command on the CLI side using the following steps.

1. Check if the app is already started and exit 0 if so
1. Detect if there is a package ready to be staged. We assume the user has a package that is meant to be staged if it is the most recently uploaded package and it hasn't been successfully turned into a droplet already.
    * `cf curl -X GET v3/apps/$(cf app app-name --guid)/packages?order_by=-created_at&per_page=1`
    * Convert JSON response into a package and take the guid
    * `cf curl -X GET v3/packages/:package-guid/droplets?states=STAGED&per_page=1`
1. If the final curl to get droplets in the staged state that run the latest package came up with an empty list of resources, build the latest package
    - `cf curl -X POST v3/build -d "package": '{ "guid": "[package-guid]" }'`
1. Poll staging logs and wait for them to complete
1. Assign the resulting droplet to your app
    - `cf curl -X PATCH v3/apps/relationships/current_droplet -d '{ "data" : { "guid" : ":droplet-guid" } }'`
1. Start the App
    -  `cf curl -X POST v3/app/$(cf app app-name --guid)/actions/start`
1. Poll the get processes endpoint until all web processes are in the started state

We will share these steps and receive feedback from other clients that we are in close contact with, and provide it as an open source resource for other clients who are dependent on V2 start behavior.

Our implementation is not overly complex. It prefers not breaking the existing V6 user workflows, but leaves potential room for error around cases where users get into unpredictable states as it depends on imperative over declarative workflows. <!-- FIXME: example? -->

### Consequences Stemming from the Decision to Implement in the CLI

 **Positive Consequences**

 * CLI users not experience breaking changes
 * Maintains REST API purity
 * Granular endpoints give advanced API consumers lots of freedom to implement their own creative solutions.
 * Allows for future Cloud Foundry API clients who have no historic knowledge of V6 workflows to depend less on the V2 ways.
 * Allows for current clients to evolve and around V3 ways.

 **Negative Consequences**

 * Multiple clients will have to re-implement the same logic
 * Less coherent definition of `start` across clients
 * Causes more complex client-side logic
 * Will be difficult to predict how this specific implementation will interact with other new V3 workflows, such as rollbacks.


### Why We Chose Not to Implement on CAPI

Implementing this on CAPI would break fundamental [V3 principles](https://github.com/cloudfoundry/cloud_controller_ng/wiki/Notes-on-V3-Architecture) designed to avoid unpredictable orchestrator type API endpoints. While the `v3/apps/:guid/actions/start` endpoint still relies on a [model_hook](https://github.com/cloudfoundry/cloud_controller_ng/blob/77a125b56545e6ed003cbab83e540ce6f4006e20/app/models/runtime/process_model.rb#L530), the `v3/build` endpoint currently follows the [API style guide's](https://github.com/cloudfoundry/cc-api-v3-style-guide#asynchronicity) preference for having endpoints that need to asynchronously communicate with other components return pollable jobs. Combining these two into a single endpoint would definitely lead to some unwanted complexity, distribution of business logic across multiple components and resources, and general unpredictability. This was one reason why these two actions were originally separated in the migration from V2 to V3.

Furthermore, a server-side implementation would be an imperfect interpretation of RESTful design, which CAPI is meant to follow. One of the principles of RESTful architecture is a separation of client and server-side concerns. Doing this allows for user interfaces to evolve independently around a consistent API, and allows scalability on the server-side by simplifying components ([source](https://www.ics.uci.edu/~fielding/pubs/dissertation/rest_arch_style.htm)). Given that these concerns were brought to us by a specific client and CAPI is not meant to service a single client, implementing this on the server-side would cause unnecessary complexity.

Finally, a client-side implementation relies on using the API in an imperative way, where declarative workflows might be preferred. Implementing `cf start` thorough a series of imperative endpoints would mean the clients would be responsible for predicting what state the user would want to get into externally. This leaves more room for error if a user gets themselves into an unpredictable state - particularly around rollback cases.

### Feedback from CAPI Consumers

Although this concern (`cf start` will stage and start unstarted applications) was brought to us by CLI users, it is worth noting that other clients depend on this workflow as well. As the concept of `start` has changed greatly on the API side, it is not a stretch to imagine many other clients are dependent the old behavior and would have to re-implement any potential client-side interpretation. We opened a [dialogue](https://pivotal.slack.com/archives/C055JEH48/p1570657730016400) and held a meeting with the Apps Manager team, who have their own client-side implementation, to hear their concerns. They told us they were open to either solution, but would want to be notified of any API changes that might break their current start implementation that depended on V3 endpoints.

## History

After collecting [user feedback](https://docs.google.com/document/d/1OPJSUYXMQMtzZmVdnvwI4NiXE0xp4tuLxO3fhhXtGwI/edit), the V3 acceleration team PMs brought to attention that in order to maintain an essential V6 workflow, Given that VAT has access to both CLI and CAPI code bases it fell to the engineers to decide whether the changed behavior would be implemented in the CLI or in CAPI. The engineers chose to implement the behavior in the CLI.
