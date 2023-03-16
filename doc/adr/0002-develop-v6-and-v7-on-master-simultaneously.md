# 2. Develop v6 and v7 on main simultaneously

Date: 2018-09-25

## Status

Accepted

## Context

_The issue motivating this decision, and any context that influences or constrains the decision._

We have two teams developing the CLI. 
One (CLI) is building out functionality (such as a better UAA and login experience) for the current, stable v6 releases. 
One (VAT) is dropping compatibility with CAPI v2 to more quickly integrate with and build out CAPI v3. 
VAT is doing so in v7 releases of the CLI.

For the last three weeks, 
VAT has been developing on a long-running _vat_ branch based on v6.39. 
Part of the goals for VAT include 
incorporating all 7-10 new epics that will be built on _main_,
along with sharing any bugfixes back and forth between the two branches.
VAT has been tracking which commits should be cherry-picked back and forth each day, 
resulting in significant overhead and bookkeeping.
Rather than continuously merging two long-running branches into each other, 
the VAT team proposes to develop on _main_ simultaneously with CLI.
This would allow the CAPI v3 acceleration work to proceed faster.

The CLI team is constrained by needing to continue shipping stable versions of v6. 
There would be increased risk of shipping unstable versions, 
given that the VAT team's work involves modifying existing code
for which there are no integration tests.
The VAT team would bear little of this risk
and would reap most of the benefit.

## Decision

_The change that we're proposing or have agreed to implement._

VAT is proposing to have both teams develop in the _main_ branch simultaneously.
 
One code base would be building two different versions of the binary.
The v6 binary would not include any v7 actors or commands.
Integration tests will need to split to represent v6 behaviour vs v7 behaviour.
We will need two separate command lists.
The main.go file can either be copied as two separate packages 
or as a single file that relies on go build flags to determine the version.
Any v6 commands that are already backed by v3 endpoints will need to be carefully dealt with
in order to ensure backwards compatibility when working with the v6 CLI.
Several packages are differentiated by a v2/v3 prefix,
but should be v6/v7 going forward to avoid confusion.
Removal of v3 prefix commands from the v6 CLI to reduce overlapping surface area.

## Consequences

_What becomes easier or more difficult to do and any risks introduced by the change that will need to be mitigated._

The v7 CLI will automatically gain all new features, bug fixes, etc., from the v6 CLI.
When VAT makes fixes to the CLI as a whole (e.g. solving a race condition in a test),
the CLI team automatically gains that benefit.
There is less risk of dropping commits that should be shared.
It will be easier for each team to restructure the code base (or even rename files) and continue to share code.

There is an increased risk of blocking each others' pipelines.
There is an increased risk of inadvertently shipping breaking changes in the v6 CLI.
There is going to be confusion as to which team owns which section of the code.
The integration test structure could become difficult to work with, especially as the two products drift apart.

The teams can still decide to branch/fork in the future if this proposal becomes untenable,
however, branching/forking now will make it nigh-impossible to merge in the future.
