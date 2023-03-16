# ADR 7: Test Coverage: Integration vs. Command

## Status

Proposed

## Context

The integration tests' coverage (e.g.
`integration/v7/isolated/create_org_command_test.go`) overlaps with the command
tests' (`command/v7/create_org_command_test.go`), and this causes confusion
for developers—does a test belong in integration, command, or both? This ADR
gives guidelines in order to assist developers making that decision.

## Decision

We want to separate [black
box](https://en.wikipedia.org/wiki/Black-box_testing) and [white
box](https://en.wikipedia.org/wiki/White-box_testing) testing, where
black box testing (functionality of an application) is done at the
integration level, and white box testing (internal structures or
workings of an application) is done at the unit (command) level.

#### Integration
Integration tests should continue to test things from a user's perspective:

- Build contexts around use cases
- Test output based on story parameters and the style guides.

For example, the following behavior (from the `unbind-security-group`
story) should be tested at the integration level because it describes
output/functionality:

> When `cf unbind-security-group` attempts to unbind a running security group
> that is not bound to a space, it should return the following:
>
> ```
> Security group my-group not bound to space my-space for lifecycle phase 'running'.
> OK
>
> TIP: If Dynamic ASG's are enabled, changes will automatically apply for running and staging applications. Otherwise, changes will require an app restart (for running) or restage (for staging) to apply to existing applications.
> ```

To do so, we would have to make a `BeforeEach()` that creates the scenario
laid out in the story and write expectations on each line of output in
the order they are listed. For example, from
`integration/v7/global/unbind_security_group_command_test.go` (edited
for readability):

```golang
BeforeEach(func() {
	port := "8443"
	description := "some-description"
	someSecurityGroup := helpers.NewSecurityGroup(securityGroupName, "tcp", "127.0.0.1", &port, &description)
	helpers.CreateSecurityGroup(someSecurityGroup)
	helpers.CreateSpace(spaceName)
})

When("the space isn't bound to the security group in any lifecycle", func() {
	It("successfully runs the command", func() {
		session := helpers.CF("unbind-security-group", securityGroupName, orgName, spaceName)
		Eventually(session).Should(Say(`Unbinding security group %s from org %s / space %s as %s\.\.\.`, securityGroupName, orgName, spaceName, username))
		Eventually(session.Err).Should(Say(`Security group %s not bound to space %s for lifecycle phase 'running'\.`, securityGroupName, spaceName))
		Eventually(session).Should(Say("OK"))
		Eventually(session).Should(Say(`TIP: If Dynamic ASG's are enabled, changes will automatically apply for running and staging applications. Otherwise, changes will require an app restart \(for running\) or restage \(for staging\) to apply to existing applications\.`))
		Eventually(session).Should(Exit(0))
	})
})
```

#### Command (Unit)
In unit tests we want to break away from this perspective. We
want to organize our tests by the [line of
sight](https://engineering.pivotal.io/post/go-flow-tests-like-code/)
method, where we cover code paths as we see them.

<!-- Brian is really uncomfortable linking to something he wrote -->

In general, every `if err != nil {` should have a corresponding `When("actor returns an
error", func() {`, but in most cases we do not need to test different
possible errors that could be returned.

Here's an example from `bind-security-group` code and unit test:

```golang
securityGroup, warnings, err := cmd.Actor.GetSecurityGroup(cmd.RequiredArgs.SecurityGroupName)
cmd.UI.DisplayWarnings(warnings)
if err != nil {
	return err
}
```

```golang
It("Retrieves the security group information", func() {
	Expect(fakeActor.GetSecurityGroupCallCount).To(Equal(1))
	securityGroupName := fakeActor.GetSecurityGroupArgsForCall(0)
	Expect(securityGroupName).To(Equal(cmd.RequiredArgs.SecurityGroupName))
})

It("prints the warnings", func() {
	Expect(testUI.Err).To(Say(getSecurityGroupWarning[0]))
})

When("an error is encountered getting the provided security group", func() {
	var expectedErr error

	BeforeEach(func() {
		expectedErr = errors.New("get security group error")
		fakeActor.GetSecurityGroupReturns(
			resources.SecurityGroup{},
			v7action.Warnings{"get security group warning"},
			expectedErr)
	})

	It("returns the error and displays all warnings", func() {
		Expect(executeErr).To(MatchError(expectedErr))
	})
})
```

## Consequences

- Since this is not a very big change, we are not expecting to lose any
  coverage or increase run times. This is mostly for organization
  purposes.

- There are no plans to retrofit existing tests to adhere to this ADR.

- These recommendations should not have any noticeable impact on the
  time it takes our unit and integration tests to complete; our
  integration are much lighter weight than other codebases.

## References

<https://testing.googleblog.com/2015/04/just-say-no-to-more-end-to-end-tests.html>

<https://kentcdodds.com/blog/write-tests>
