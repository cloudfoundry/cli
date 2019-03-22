# 3. Resource Matching Implementation

Date: 2019-03-22

## Status

Accepted

## Context

_The issue motivating this decision, and any context that influences or constrains the decision._

When uploading applications to a particular foundation,
there may be duplicate files that are shared between apps.
In order to reduce the load from uploading duplicate files,
the Cloud Controller exposes an endpoint to match existing files prior to uploading application bits.
The client then uses this endpoint to determine which files require upload.

During the first stage of push,
PushPlans are created to determine what actions need to take place during the remaining push operation.
This introduced a unique decision,
perform resource match when creating push plans
or perform resource matching immediately prior to uploading application bits.

## Decision

_The change that we're proposing or have agreed to implement._

The decision was made to perform resource matching immediately prior to uploading application bits.

Performing resource matching as late as possible would result in the most optimized matching.
Since multiple applications can be pushed at a time,
if the client were to perform a resource match upfront,
none of the shared files between the apps would be matched because none of them have been uploaded yet.
By deferring the resource match to just before an app-specific upload,
files shared with any of the subsequent applications will then be matched before upload.

## Consequences

_What becomes easier or more difficult to do and any risks introduced by the change that will need to be mitigated._

The intention of the `CreatePushPlans` function was to have a single place which determines the actions needed during the push.
The decision to place resource matching later in the process diverges from this intention and may lead to confusion in the future.
The increase in performance when placing resource matching later in the process out-weighs the additional complexity due to the positioning.
