# 4. V7 `cf push` Refactor

Date: 2019-07-02

## Status

Accepted

## Context

_The issue motivating this decision, and any context that influences or constrains the decision._

As the V3 Acceleration Team was developing the V7 version of `cf push`, the team saw opportunities to structure the code to make it easier to understand and maintain. In general, the `push` command logic is some of the most complex in the codebase, and this command was still considered experimental, so it seemed like the right time and place to invest our energy in refactoring.

In the V6 CLI, the code for `push` was split roughly into two methods: `Conceptualize` and `Actualize`. `Conceptualize` was responsible for taking user input (flags, manifest properties, etc) and generating a "push plan" for each app that is being pushed (a struct containing various pieces of information needed to create/push/update the app). `Actualize` was responsible for taking these push plans and, based on the plans, taking the necessary actions to complete the push process. The refactor preserves the spirit of this division, but prevents these two methods from growing too large and unmaintainable.

Some of the goals of the refactor were:

- Make it easier to add features to `push` (e.g. new flags, new options for flags, new manifest properties)
- Make it easier to unit test the components of the `push` workflow. This means meant splitting the command into several smaller functions that can be tested individually but composed into sequences based on the given user input.

## Decision

_The change that we're proposing or have agreed to implement._

Both the `Conceptualize` and `Actualize` parts of the push process have been split up and refactored in a manner inspired by [hexagonal architecture](https://medium.com/@nicolopigna/demystifying-the-hexagon-5e58cb57bbda).

### Hexagonal architecture

For our purposes, "hexagonal architecture" means the following:

- There is one place in the code responsible for calling a given sequence of functions in order. The output of one function is passed in as input to the next function.
- All of these functions have the same signature.
- The central place where the functions are called is agnostic to what the functions are doing.
- Any new features/branches/logic should be added to one or more of these functions (or, a new function should be added to encapsulate it). We generally shouldn't need to touch the place where the functions are called.

### Splitting up `Conceptualize`

What was previously called `Conceptualize` now looks roughly like this:

_Note: We are still iterating on this code, so it may evolve over time and no longer look exactly like this. Still, the code below illustrates the idea we were going for._

```go
var pushPlans []PushPlan

for _, manifestApplication := range getEligibleApplications(parser, appNameArg) {
    plan := PushPlan{
        OrgGUID:   orgGUID,
        SpaceGUID: spaceGUID,
    }

    for _, updatePlan := range actor.PreparePushPlanSequence {
        var err error
        plan, err = updatePlan(plan, overrides, manifestApplication)
        if err != nil {
            return nil, err
        }
    }

    pushPlans = append(pushPlans, plan)
}

return pushPlans, nil
```

This is the central place where all of the "prepare push plan" functions are called. We loop through `actor.PreparePushPlanSequence`, an array of functions, and call each one with a push plan. They each have the opportunity to **return a modified push plan**, which then gets **passed into the next function**.

In this case, each function is also called with `overrides` and `manifestApplication`, which represent user input (in the form of flags and manifest properties, respectively). This lets each function inspect the user input and modify the push plan accordingly.

When this loop completes, the original push plan has flowed through each function in the sequence and has been modified to include information based on the given flags/manifest.

Let's now look at how `actor.PreparePushPlanSequence` is defined:

```go
actor.PreparePushPlanSequence = []UpdatePushPlanFunc{
    SetupApplicationForPushPlan,
    SetupDockerImageCredentialsForPushPlan,
    SetupBitsPathForPushPlan,
    SetupDropletPathForPushPlan,
    actor.SetupAllResourcesForPushPlan,
    SetupDeploymentStrategyForPushPlan,
    SetupNoStartForPushPlan,
    SetupNoWaitForPushPlan,
    SetupSkipRouteCreationForPushPlan,
    SetupScaleWebProcessForPushPlan,
    SetupUpdateWebProcessForPushPlan,
}
```

This is just a simple array with a bunch of functions that all conform to the correct interface (they all have type `UpdatePushPlanFunc`).

Here is an example of one of them:

```go
func SetupScaleWebProcessForPushPlan(pushPlan PushPlan, overrides FlagOverrides, manifestApp manifestparser.Application) (PushPlan, error) {
	if overrides.Memory.IsSet || overrides.Disk.IsSet || overrides.Instances.IsSet {
		pushPlan.ScaleWebProcessNeedsUpdate = true

		pushPlan.ScaleWebProcess = v7action.Process{
			Type:       constant.ProcessTypeWeb,
			DiskInMB:   overrides.Disk,
			Instances:  overrides.Instances,
			MemoryInMB: overrides.Memory,
		}
	}
	return pushPlan, nil
}
```

This simple function populates fields on the push plan based on flag overrides and returns the enhanced push plan, making it easy to test on its own.

When all of the functions have run and updated the push plan, the "actualize" step doesn't need to know about flags and manifest properties anymore; it only needs to receive a push plan because all user input has been resolved and combined into the push plan object.

### Splitting up `Actualize`

There is still (at the time of this writing) a function called `Actualize`, but it now looks like this:

```go
for _, changeAppFunc := range actor.ChangeApplicationSequence(plan) {
    plan, warnings, err = changeAppFunc(plan, eventStream, progressBar)
    warningsStream <- warnings
    if err != nil {
        errorStream <- err
        return
    }
    planStream <- plan
}
```

This is quite similar to the loop through `actor.PreparePushPlanSequence` above. We loop through `actor.ChangeApplicationSequence(plan)`, which returns an array of functions, and we call each one with the push plan. Each one returns a push plan that then gets passed into the next function.

_Note: The rest of that code uses streams to report progress, errors, and warnings, but this is not the focus of this ADR (and we may end up changing this as well)._

The **biggest difference** from `Conceptualize` is that instead of being a static list of functions (like `actor.PreparePushPlanSequence`), `actor.ChangeApplicationSequence` is a function that takes in a push plan and returns an array of `ChangeApplicationFunc`s. This allows us to dynamically build up the sequence of actions we run based on the push plan, rather than run the same sequence every time.

Let's look at how that works:

```go
actor.ChangeApplicationSequence = func(plan PushPlan) []ChangeApplicationFunc {
    var sequence []ChangeApplicationFunc
    sequence = append(sequence, actor.GetUpdateSequence(plan)...)
    sequence = append(sequence, actor.GetPrepareApplicationSourceSequence(plan)...)
    sequence = append(sequence, actor.GetRuntimeSequence(plan)...)
    return sequence
}
```

This function is responsible for building up the sequence based on the given plan. It delegates to three helpers, each of which builds up a subsequence of actions. Here's one of them:

```go
func ShouldCreateBitsPackage(plan PushPlan) bool {
	return plan.DropletPath == "" && !plan.DockerImageCredentialsNeedsUpdate
}

// ...

func (actor Actor) GetPrepareApplicationSourceSequence(plan PushPlan) []ChangeApplicationFunc {
	var prepareSourceSequence []ChangeApplicationFunc
	switch {
	case ShouldCreateBitsPackage(plan):
		prepareSourceSequence = append(prepareSourceSequence, actor.CreateBitsPackageForApplication)
	case ShouldCreateDockerPackage(plan):
		prepareSourceSequence = append(prepareSourceSequence, actor.CreateDockerPackageForApplication)
	case ShouldCreateDroplet(plan):
		prepareSourceSequence = append(prepareSourceSequence, actor.CreateDropletForApplication)
	}
	return prepareSourceSequence
}
```

In this case, we only want to include one of these three functions in the final sequence, which is determined based on properties of the push plan.

Since all of these functions are small and straightforward on their own, they are easy to unit test. They can be **composed together in different sequences to build up different push workflows** (based on different flags/manifests), and this is where the refactor really starts to pay off.

## Consequences

_What becomes easier or more difficult to do and any risks introduced by the change that will need to be mitigated._

### What becomes easier?

Figuring out where to write the code to add a new flag becomes easier. Consider [this recent commit](https://github.com/cloudfoundry/cli/commit/9dbddd165ae77dbc33e7dc34ae896e6f880ce3ff), which added the `--no-wait` flag to the V7 `push` command. These were the bulk of the changes needed to add this new branch to the workflow:

- [New, three-line method](https://github.com/cloudfoundry/cli/commit/9dbddd165ae77dbc33e7dc34ae896e6f880ce3ff#diff-8efa044476f78f1ca7dbe0f90addeca4) implementing the `UpdatePushPlanFunc` interface, plus [unit tests for it](https://github.com/cloudfoundry/cli/commit/9dbddd165ae77dbc33e7dc34ae896e6f880ce3ff#diff-03f194b5c186fb49de70b64082df2191)
- [One-line change](https://github.com/cloudfoundry/cli/commit/9dbddd165ae77dbc33e7dc34ae896e6f880ce3ff#diff-d84d0497f89a03505f051cbe0a418739) to the actor to add this new method to the `ChangeApplicationSequence`
- [Simple change](https://github.com/cloudfoundry/cli/commit/9dbddd165ae77dbc33e7dc34ae896e6f880ce3ff#diff-59e4c63c96b86434900c067fc7d0e49f) to pass the new push plan property as into the methods that need it

There are more changes in that commit, but that highlights the parts relevant to this ADR.

### What becomes harder?

It is slightly harder to grasp how all these pieces fit together at first glance. For the arrays of functions detailed above, it is not immediately clear where or how they are called (since that is abstracted away into a different part of the codebase). We believe that after spending time understanding the new structure of things, developers will appreciate how straightforward it is to make changes.
