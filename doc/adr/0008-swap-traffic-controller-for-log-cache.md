# 8. Swap Traffic Controller for Log Cache

Date: 2020-04-09

## Status

Accepted

## Context

_The issue motivating this decision, and any context that influences or constrains the decision._

The Loggregator Team proposed modifying the cf CLI to retrieve logs from Log
Cache rather than from Traffic Controller, as part of moving to a "Shared
Nothing" architecture: \
https://docs.google.com/document/d/1_Ve4wAkeCC0fIJ1TiAWSfxzNp5zB_Ndoq5UmfUNJtgs/edit?usp=sharing

This is intended to improve scalability of log egress from the platform.

The intention is to provide the same user experience but utilising Log Cache for
application logs.

We are limited in our ability to modify the Log Cache API by our versioning
policy which defines minimum versions of cf-deployment we will support: \
https://github.com/cloudfoundry/cli/wiki/Versioning-Policy

## Decision

_The change that we're proposing or have agreed to implement._

* cf CLI v7 will be modified to retrieve application logs from Log Cache
* cf CLI v6 will be largely modified to retrieve application logs from Log Cache
* cf CLI v6 v3-prefixed commands will not be modified to retrieve application
  logs from Log Cache. These commands are experimental and users should be
  encouraged to use the v7 CLI.
* cf CLI v6/v7 legacy command implementations accessible through the Plugin API
  will be modified to retrieve application logs from Log Cache

## Consequences

_What becomes easier or more difficult to do and any risks introduced by the change that will need to be mitigated._

* Cloud Foundry log egress is expected to become more scalable and resource
  efficient
* The Traffic Controller V1 implementation was based on a long-lived Websocket
  connection, where-as Log Cache requires the client to make the appropriate
  requests to retrieve the relevant contents of the cache.
* The log envelopes for an application are all grouped together in the same
  source-id "bucket". This poses the risk that logs may be shown for a different
  operation than that being performed by the user.
* Log envelopes are indexed by timestamp. Use of the client's clock when making
  decisions about which logs to retrieve or show is likely to be unreliable.
* Authorization for log retrieval would previously need to be done once when
  establishing a long-lived connection for streaming logs. It must now be
  performed for each request made when "walking" Log Cache.
* Log Cache may perform caching of authorized source-ids. Users that push a
  number of applications in quick succession may be unable to retrieve logs.
