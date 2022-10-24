# 10. Synchronizing V7 and V8 (Main) code

Date: 2020-06-04

## Status

Submitted

## Context

_The issue motivating this decision, and any context that influences or constrains the decision._

**Current**: Changes committed to the [V7
branch](https://github.com/cloudfoundry/cli/tree/v7) need to be cherry-picked
onto the [main (V8) branch](https://github.com/cloudfoundry/cli), a manual
process.

This depends on the pair to remember to cherry-pick. If forgotten, the commit
will not be merged, and we run the risk of bugfixes (and features) not
manifested in both branches.

## Decision

_The change that we're proposing or have agreed to implement._

Commit changes on the v7 branch then merge to main/v8, e.g.:

```bash
git co v7
git pull -r
 # make changes
make
git add -p
git ci
bin/push.sh
```

[`push.sh`](https://github.com/cloudfoundry/cli/blob/main/bin/push.sh) is a
scripts which pushes the v7 commit, rebases against main, checks out main,
pulls, merges v7, pushes, and then checks out v7. It automates some of the more
tedious aspects of the commit cycle.

This is the process that Ops Manager follows to maintain several branches.

It is not without drawbacks, though. Here are the notes from the Ops Mgr anchor:

> We're still using merge forward as the primary method of sharing features / bug fixes across all of our releases. It generally still works well for us, but there are drawbacks.
> * If you care about a "clean", fairly linear git history for each branch. The merge commits can definitely make things difficult to follow sometimes. This isn't much of an issue for us, but I can imagine it being more important for ya'll since it's an open source project
> * It assumes that everything in the earlier branch belongs in the later branch. This makes a lot of sense for common bug fixes, but perhaps there are feature differences that could make it trickier in your case.
> * It somewhat discourages large refactors or cleanups. In 2.7 we did a major refactor of our React code, so anytime we need to do a UI bug fix in 2.6 or below, it becomes a pain to merge forward and we basically have to re-do the work. (Though this probably would apply to a cherry-pick strategy as well!)
> I think for Ops Manager, since we're maintaining so many branches at once, and the fact that most of what we would do on earlier branches is bug fixes for ALL branches, it's slightly less overhead for us to do merge forward. If the CF CLI is diverging in any significant manner between versions, and you're only going to be dealing with 2 versions, cherry-picking from one branch to the other might be cleaner. Merge forward should still work, but maybe you'd be getting less benefit out of it.

## Consequences

_What becomes easier or more difficult to do and any risks introduced by the change that will need to be mitigated._

If a pair forgets to merge their commit to main, the subsequent merge to
main will pick up that change (this is a good thing).

Sometimes a merge conflict will occur. This happens with cherry-picks, too.

The first merge will entail duplicate commits (the ones that have been
cherry-picked).
