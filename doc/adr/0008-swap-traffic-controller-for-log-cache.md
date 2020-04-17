# 8. Swap Traffic Controller for Log Cache

Date: 2020-04-09

## Status

Accepted

## Context

_The issue motivating this decision, and any context that influences or constrains the decision._

The Loggregator team decided to move to a ["Shared
Nothing" architecture](https://docs.google.com/document/d/1_Ve4wAkeCC0fIJ1TiAWSfxzNp5zB_Ndoq5UmfUNJtgs/edit?usp=sharing)
that would improve scalability of log egress from the platform. This architecture deprecated the V1 Firehose Traffic Controller,
the component from which the cf CLI retrieved application logs. As a result, the Loggregator team proposed modifying the cf CLI to retrieve
logs from Log Cache instead.

## Decision

_The change that we're proposing or have agreed to implement._

After December 2019, the CLI's minimum supported version of cf-deployment contains Log Cache. Therefore, in 2020:

* cf CLI v7 will be modified to retrieve application logs from Log Cache
* cf CLI v6 will be modified to retrieve application logs from Log Cache _with the exception of experimental v3-prefixed commands_
* cf CLI v6/v7 legacy command implementations accessible through the Plugin API
  will be modified to retrieve application logs from Log Cache

There will be no user-facing or breaking changes.

## Consequences

_What becomes easier or more difficult to do and any risks introduced by the change that will need to be mitigated._

* Cloud Foundry log egress is expected to become more scalable and resource
  efficient
* We will be limited in our ability to modify the Log Cache API by our minimum
  supported version [policy](https://github.com/cloudfoundry/cli/wiki/Versioning-Policy).
* The Traffic Controller V1 implementation was based on a long-lived Websocket
  connection. Log Cache requires the client to make GET requests to an HTTP API
  to retrieve application logs.
* The log envelopes for an application are grouped together into
  source-id "buckets". This poses the risk that logs may be shown for a different
  operation than that being performed by the user.
* Authorization for log retrieval would previously need to be done once when
  establishing a long-lived connection for streaming logs. It must now be
  performed for each request made when "walking" Log Cache.
* Log Cache may perform caching of authorized source-ids. Users that push a
  number of applications in quick succession may be unable to retrieve logs.
