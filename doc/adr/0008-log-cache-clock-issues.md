# 8. Log Cache and Clock Issues

Date: 2020-04-08

## Status

Accepted

## Context

_The issue motivating this decision, and any context that influences or constrains the decision._

The cf CLI previously used the traffic controller to retrieve application logs.
This was implemented as a long-lived websocket connection where logs were
streamed to the client.

With the migration to Log Cache we needed to preserve the experience of
streaming logs, but built [atop an
API](https://github.com/cloudfoundry/log-cache-release/blob/f08a3081c029d133300b1d6cb5ea8ebbd2108874/src/README.md)
where we make numerous HTTP requests to index into a collection of timestamped log envelopes.

There are two main questions we needed to answer:

1. What is the oldest log that should be shown?
   An application may have logs from a previous push already present in log cache
   that do not relate to the current staging operation. We would want to only
   show the logs for the current staging operation.
2. What is the newest log that can be shown? Log Cache is eventually
   consistent, so very recent logs may not have 'settled' and if we show them to
   the user and read forward immediately we will effectively drop logs.

## Decision

_The change that we're proposing or have agreed to implement._

1. To determine the oldest log that will be shown: we 'peek' at the latest log within
   Log Cache for the application and read from that point (applying a small offset).
1. To determine the newest log that will be shown: we do not show envelopes where the
   timestamp is very new (less than two seconds old).

### Peeking at the latest log

The initial cli implementation of Log Cache started reading from Log Cache at an
offset based on the current client clock time. This was flawed because an
incorrectly configured client clock would result in some unexpected behaviour:

* If the client clock was ahead of the server then either the logs would not be
  shown at all, or there would be a lengthy delay before some of the logs would
  be shown
* If the client clock was behind the server then logs relating to a previous
  operation might be shown

In an attempt to decouple ourselves from the client clock time we instead 'peek'
at the timestamp of the latest log envelope for the application and add an
offset to that timestamp in order to start reading from that point.

XXX: Describe how we arrived at the offset

If there are no envelopes present for the application we will continue to retry
until envelopes become available.

XXX: Consider if we are retrying on an empty response from Log Cache without a delay.

### Delaying the output of new logs

The Log Cache client library has a concept of a `WalkDelay` where log envelopes
with timestamps that are very 'new' are ignored until they are more than a
certain duration old: \
https://github.com/cloudfoundry/log-cache-release/blob/f08a3081c029d133300b1d6cb5ea8ebbd2108874/src/pkg/client/walk.go#L174-L183

This defaults to 1 second. In our testing we found that this default was not
sufficient to avoid effective loss of logs when running against a foundation
with more than one log cache node.

This is because multiple Log Cache nodes may be ingesting events for the same
application, but a single Log Cache node hosts the cache for a given
application. It's possible for us to see a newer log envelope timestamp and move
our timestamp cursor forwards before we have received an earlier log envelope:
https://www.pivotaltracker.com/story/show/171759407/comments/212391238

We experimented with changing this duration and found that a `WalkDelay` of 2 seconds was
sufficient to ensure that there were no 'missing' log envelopes.

## Consequences

_What becomes easier or more difficult to do and any risks introduced by the change that will need to be mitigated._

### Peeking at the latest log

* Peeking at the latest envelope for an application removes a dependency on the
  client clock, making us more tolerant to a situation where the client clock is
  not closely synchronised to the server
* As each component that generates logs is responsible for assigning the
  timestamp, there is still potential for clock drift within the foundation to
  cause strange behaviour
* The code is slightly more involved, but it seems like a good trade-off

### Delaying the output of new logs

* We are momentarily delaying the output of new log envelopes to the user, with
  the advantage that by delaying we are able to output a more consistent view of
  the logs
* The `WalkDelay` is implemented within the Log Cache client as a comparison of
  the log envelope timestamps against the client clock. It is therefore
  vulnerable to any misconfiguration of the client clock. We've considered
  strategies for removing this dependency on the client clock but have decided
  for now to wait for feedback from our users before attempting to make ourselves
  reliable in this situation.
