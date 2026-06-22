package v7

// Internal test file (package v7, not v7_test) so we can reach the unexported
// validateSourceFlags and resolveSource helpers.  We can't import v7fakes here
// because it imports this package (v7) and that would be a circular dependency.
// Instead we build a minimal stub actor inline that embeds the Actor interface
// for compile-time satisfaction and overrides only the three methods that
// resolveSource calls.

import (
	"errors"
	"fmt"

	"code.cloudfoundry.org/cli/v9/actor/actionerror"
	"code.cloudfoundry.org/cli/v9/actor/v7action"
	"code.cloudfoundry.org/cli/v9/command/commandfakes"
	"code.cloudfoundry.org/cli/v9/command/translatableerror"
	"code.cloudfoundry.org/cli/v9/resources"
	"code.cloudfoundry.org/cli/v9/util/configv3"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// stubSourceActor implements Actor by embedding the interface (nil value).
// Any method not explicitly overridden will panic if called — which is fine
// because our tests only exercise the three methods resolveSource uses.
type stubSourceActor struct {
	Actor
	getOrgByName         func(string) (resources.Organization, v7action.Warnings, error)
	getSpaceByNameAndOrg func(string, string) (resources.Space, v7action.Warnings, error)
	getAppByNameAndSpace func(string, string) (resources.Application, v7action.Warnings, error)
}

func (s *stubSourceActor) GetOrganizationByName(name string) (resources.Organization, v7action.Warnings, error) {
	return s.getOrgByName(name)
}
func (s *stubSourceActor) GetSpaceByNameAndOrganization(spaceName, orgGUID string) (resources.Space, v7action.Warnings, error) {
	return s.getSpaceByNameAndOrg(spaceName, orgGUID)
}
func (s *stubSourceActor) GetApplicationByNameAndSpace(appName, spaceGUID string) (resources.Application, v7action.Warnings, error) {
	return s.getAppByNameAndSpace(appName, spaceGUID)
}

var _ = Describe("RoutePolicySourceFlags", func() {

	Describe("validateSourceFlags", func() {
		It("returns RequiredArgumentError when no source flag is given", func() {
			f := RoutePolicySourceFlags{}
			Expect(f.validateSourceFlags()).To(MatchError(translatableerror.RequiredArgumentError{
				ArgumentName: "one of: --source-app, --source-space, --source-org, --source-any, or --source",
			}))
		})

		It("accepts --source alone", func() {
			f := RoutePolicySourceFlags{Source: "cf:any"}
			Expect(f.validateSourceFlags()).To(Succeed())
		})

		It("accepts --source-any alone", func() {
			f := RoutePolicySourceFlags{SourceAny: true}
			Expect(f.validateSourceFlags()).To(Succeed())
		})

		It("accepts --source-app alone", func() {
			f := RoutePolicySourceFlags{SourceApp: "my-app"}
			Expect(f.validateSourceFlags()).To(Succeed())
		})

		It("accepts --source-space alone (standalone space policy)", func() {
			f := RoutePolicySourceFlags{SourceSpace: "my-space"}
			Expect(f.validateSourceFlags()).To(Succeed())
		})

		It("accepts --source-org alone (standalone org policy)", func() {
			f := RoutePolicySourceFlags{SourceOrg: "my-org"}
			Expect(f.validateSourceFlags()).To(Succeed())
		})

		It("accepts --source-app + --source-space (space qualifies the app lookup)", func() {
			f := RoutePolicySourceFlags{SourceApp: "my-app", SourceSpace: "other-space"}
			Expect(f.validateSourceFlags()).To(Succeed())
		})

		It("accepts --source-app + --source-space + --source-org", func() {
			f := RoutePolicySourceFlags{SourceApp: "my-app", SourceSpace: "other-space", SourceOrg: "other-org"}
			Expect(f.validateSourceFlags()).To(Succeed())
		})

		It("accepts --source-space + --source-org (org qualifies the space lookup)", func() {
			f := RoutePolicySourceFlags{SourceSpace: "my-space", SourceOrg: "other-org"}
			Expect(f.validateSourceFlags()).To(Succeed())
		})

		It("returns RequiredFlagsError when --source-org is combined with --source-app but --source-space is missing", func() {
			f := RoutePolicySourceFlags{SourceApp: "my-app", SourceOrg: "my-org"}
			Expect(f.validateSourceFlags()).To(MatchError(translatableerror.RequiredFlagsError{
				Arg1: "--source-org",
				Arg2: "--source-space",
			}))
		})

		It("returns ArgumentCombinationError when two primary flags are given (--source-app + --source-any)", func() {
			f := RoutePolicySourceFlags{SourceApp: "my-app", SourceAny: true}
			err := f.validateSourceFlags()
			Expect(err).To(BeAssignableToTypeOf(translatableerror.ArgumentCombinationError{}))
			combo := err.(translatableerror.ArgumentCombinationError)
			Expect(combo.Args).To(ConsistOf("--source-app", "--source-any"))
		})

		It("returns ArgumentCombinationError when two primary flags are given (--source + --source-any)", func() {
			f := RoutePolicySourceFlags{Source: "cf:any", SourceAny: true}
			err := f.validateSourceFlags()
			Expect(err).To(BeAssignableToTypeOf(translatableerror.ArgumentCombinationError{}))
			combo := err.(translatableerror.ArgumentCombinationError)
			Expect(combo.Args).To(ConsistOf("--source", "--source-any"))
		})
	})

	Describe("resolveSource", func() {
		var (
			fakeConfig *commandfakes.FakeConfig
			actor      *stubSourceActor
		)

		BeforeEach(func() {
			fakeConfig = new(commandfakes.FakeConfig)
			fakeConfig.TargetedSpaceReturns(configv3.Space{GUID: "targeted-space-guid", Name: "targeted-space"})
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{GUID: "targeted-org-guid", Name: "targeted-org"})

			actor = &stubSourceActor{}
		})

		Context("--source (raw passthrough)", func() {
			It("returns the value as-is without calling the actor", func() {
				f := RoutePolicySourceFlags{Source: "cf:app:some-guid"}
				src, scope, warnings, err := resolveSource(f, actor, fakeConfig)
				Expect(err).NotTo(HaveOccurred())
				Expect(src).To(Equal("cf:app:some-guid"))
				Expect(scope).To(Equal("source: cf:app:some-guid"))
				Expect(warnings).To(BeEmpty())
			})
		})

		Context("--source-any", func() {
			It("returns cf:any with the expected scope display", func() {
				f := RoutePolicySourceFlags{SourceAny: true}
				src, scope, warnings, err := resolveSource(f, actor, fakeConfig)
				Expect(err).NotTo(HaveOccurred())
				Expect(src).To(Equal("cf:any"))
				Expect(scope).To(Equal("scope: any, source: any authenticated app"))
				Expect(warnings).To(BeEmpty())
			})
		})

		Context("--source-app (app in currently targeted space)", func() {
			BeforeEach(func() {
				actor.getAppByNameAndSpace = func(appName, spaceGUID string) (resources.Application, v7action.Warnings, error) {
					Expect(appName).To(Equal("my-app"))
					Expect(spaceGUID).To(Equal("targeted-space-guid"))
					return resources.Application{GUID: "app-guid"}, v7action.Warnings{"app-warning"}, nil
				}
			})

			It("resolves the app GUID and returns the expected source and scope", func() {
				f := RoutePolicySourceFlags{SourceApp: "my-app"}
				src, scope, warnings, err := resolveSource(f, actor, fakeConfig)
				Expect(err).NotTo(HaveOccurred())
				Expect(src).To(Equal("cf:app:app-guid"))
				Expect(scope).To(Equal("scope: app, source: my-app"))
				Expect(warnings).To(ConsistOf("app-warning"))
			})
		})

		Context("--source-app when the app is not found (without --source-space)", func() {
			BeforeEach(func() {
				actor.getAppByNameAndSpace = func(appName, spaceGUID string) (resources.Application, v7action.Warnings, error) {
					return resources.Application{}, nil, actionerror.ApplicationNotFoundError{Name: appName}
				}
			})

			It("returns a friendly error with a TIP about --source-space / --source-org", func() {
				f := RoutePolicySourceFlags{SourceApp: "my-app"}
				_, _, _, err := resolveSource(f, actor, fakeConfig)
				Expect(err).To(MatchError(fmt.Sprintf(
					"App 'my-app' not found in space 'targeted-space' / org 'targeted-org'.\nTIP: If the app is in a different space or org, use --source-space and/or --source-org flags.",
				)))
			})
		})

		Context("--source-app with a non-NotFound actor error (without --source-space)", func() {
			BeforeEach(func() {
				actor.getAppByNameAndSpace = func(_, _ string) (resources.Application, v7action.Warnings, error) {
					return resources.Application{}, v7action.Warnings{"w1"}, errors.New("upstream error")
				}
			})

			It("passes the error through", func() {
				f := RoutePolicySourceFlags{SourceApp: "my-app"}
				_, _, warnings, err := resolveSource(f, actor, fakeConfig)
				Expect(err).To(MatchError("upstream error"))
				Expect(warnings).To(ConsistOf("w1"))
			})
		})

		Context("--source-app + --source-space (cross-space app lookup)", func() {
			BeforeEach(func() {
				actor.getSpaceByNameAndOrg = func(spaceName, orgGUID string) (resources.Space, v7action.Warnings, error) {
					Expect(spaceName).To(Equal("other-space"))
					Expect(orgGUID).To(Equal("targeted-org-guid"))
					return resources.Space{GUID: "other-space-guid", Name: "other-space"}, v7action.Warnings{"space-warning"}, nil
				}
				actor.getAppByNameAndSpace = func(appName, spaceGUID string) (resources.Application, v7action.Warnings, error) {
					Expect(appName).To(Equal("my-app"))
					Expect(spaceGUID).To(Equal("other-space-guid"))
					return resources.Application{GUID: "app-guid"}, v7action.Warnings{"app-warning"}, nil
				}
			})

			It("resolves space then app, includes space in scope display", func() {
				f := RoutePolicySourceFlags{SourceApp: "my-app", SourceSpace: "other-space"}
				src, scope, warnings, err := resolveSource(f, actor, fakeConfig)
				Expect(err).NotTo(HaveOccurred())
				Expect(src).To(Equal("cf:app:app-guid"))
				Expect(scope).To(Equal("scope: app, source: my-app (space: other-space)"))
				Expect(warnings).To(ConsistOf("space-warning", "app-warning"))
			})
		})

		Context("--source-app + --source-space + --source-org (cross-org/space app lookup)", func() {
			BeforeEach(func() {
				actor.getOrgByName = func(orgName string) (resources.Organization, v7action.Warnings, error) {
					Expect(orgName).To(Equal("other-org"))
					return resources.Organization{GUID: "other-org-guid"}, v7action.Warnings{"org-warning"}, nil
				}
				actor.getSpaceByNameAndOrg = func(spaceName, orgGUID string) (resources.Space, v7action.Warnings, error) {
					Expect(spaceName).To(Equal("other-space"))
					Expect(orgGUID).To(Equal("other-org-guid"))
					return resources.Space{GUID: "other-space-guid", Name: "other-space"}, v7action.Warnings{"space-warning"}, nil
				}
				actor.getAppByNameAndSpace = func(appName, spaceGUID string) (resources.Application, v7action.Warnings, error) {
					Expect(appName).To(Equal("my-app"))
					Expect(spaceGUID).To(Equal("other-space-guid"))
					return resources.Application{GUID: "app-guid"}, v7action.Warnings{"app-warning"}, nil
				}
			})

			It("resolves org, space, then app; includes both in scope display", func() {
				f := RoutePolicySourceFlags{SourceApp: "my-app", SourceSpace: "other-space", SourceOrg: "other-org"}
				src, scope, warnings, err := resolveSource(f, actor, fakeConfig)
				Expect(err).NotTo(HaveOccurred())
				Expect(src).To(Equal("cf:app:app-guid"))
				Expect(scope).To(Equal("scope: app, source: my-app (space: other-space, org: other-org)"))
				Expect(warnings).To(ConsistOf("org-warning", "space-warning", "app-warning"))
			})
		})

		Context("--source-space (standalone space policy in targeted org)", func() {
			BeforeEach(func() {
				actor.getSpaceByNameAndOrg = func(spaceName, orgGUID string) (resources.Space, v7action.Warnings, error) {
					Expect(spaceName).To(Equal("my-space"))
					Expect(orgGUID).To(Equal("targeted-org-guid"))
					return resources.Space{GUID: "space-guid"}, v7action.Warnings{"space-warning"}, nil
				}
			})

			It("returns cf:space:<guid> with space scope display", func() {
				f := RoutePolicySourceFlags{SourceSpace: "my-space"}
				src, scope, warnings, err := resolveSource(f, actor, fakeConfig)
				Expect(err).NotTo(HaveOccurred())
				Expect(src).To(Equal("cf:space:space-guid"))
				Expect(scope).To(Equal("scope: space, source: my-space"))
				Expect(warnings).To(ConsistOf("space-warning"))
			})
		})

		Context("--source-space + --source-org (space in a different org)", func() {
			BeforeEach(func() {
				actor.getOrgByName = func(orgName string) (resources.Organization, v7action.Warnings, error) {
					Expect(orgName).To(Equal("other-org"))
					return resources.Organization{GUID: "other-org-guid"}, v7action.Warnings{"org-warning"}, nil
				}
				actor.getSpaceByNameAndOrg = func(spaceName, orgGUID string) (resources.Space, v7action.Warnings, error) {
					Expect(spaceName).To(Equal("my-space"))
					Expect(orgGUID).To(Equal("other-org-guid"))
					return resources.Space{GUID: "space-guid"}, v7action.Warnings{"space-warning"}, nil
				}
			})

			It("resolves org then space; includes org in scope display", func() {
				f := RoutePolicySourceFlags{SourceSpace: "my-space", SourceOrg: "other-org"}
				src, scope, warnings, err := resolveSource(f, actor, fakeConfig)
				Expect(err).NotTo(HaveOccurred())
				Expect(src).To(Equal("cf:space:space-guid"))
				Expect(scope).To(Equal("scope: space, source: my-space (org: other-org)"))
				Expect(warnings).To(ConsistOf("org-warning", "space-warning"))
			})
		})

		Context("--source-org (standalone org policy)", func() {
			BeforeEach(func() {
				actor.getOrgByName = func(orgName string) (resources.Organization, v7action.Warnings, error) {
					Expect(orgName).To(Equal("my-org"))
					return resources.Organization{GUID: "org-guid"}, v7action.Warnings{"org-warning"}, nil
				}
			})

			It("returns cf:org:<guid> with org scope display", func() {
				f := RoutePolicySourceFlags{SourceOrg: "my-org"}
				src, scope, warnings, err := resolveSource(f, actor, fakeConfig)
				Expect(err).NotTo(HaveOccurred())
				Expect(src).To(Equal("cf:org:org-guid"))
				Expect(scope).To(Equal("scope: org, source: my-org"))
				Expect(warnings).To(ConsistOf("org-warning"))
			})
		})

		Context("error propagation", func() {
			It("returns the error and warnings when GetOrganizationByName fails (--source-org)", func() {
				actor.getOrgByName = func(string) (resources.Organization, v7action.Warnings, error) {
					return resources.Organization{}, v7action.Warnings{"warn"}, errors.New("org-lookup-error")
				}
				f := RoutePolicySourceFlags{SourceOrg: "bad-org"}
				_, _, warnings, err := resolveSource(f, actor, fakeConfig)
				Expect(err).To(MatchError("org-lookup-error"))
				Expect(warnings).To(ConsistOf("warn"))
			})

			It("returns the error and warnings when GetOrganizationByName fails (--source-space + --source-org)", func() {
				actor.getOrgByName = func(string) (resources.Organization, v7action.Warnings, error) {
					return resources.Organization{}, v7action.Warnings{"warn"}, errors.New("org-lookup-error")
				}
				f := RoutePolicySourceFlags{SourceSpace: "my-space", SourceOrg: "bad-org"}
				_, _, warnings, err := resolveSource(f, actor, fakeConfig)
				Expect(err).To(MatchError("org-lookup-error"))
				Expect(warnings).To(ConsistOf("warn"))
			})

			It("returns the error and warnings when GetSpaceByNameAndOrganization fails (--source-space)", func() {
				actor.getSpaceByNameAndOrg = func(string, string) (resources.Space, v7action.Warnings, error) {
					return resources.Space{}, v7action.Warnings{"warn"}, errors.New("space-lookup-error")
				}
				f := RoutePolicySourceFlags{SourceSpace: "bad-space"}
				_, _, warnings, err := resolveSource(f, actor, fakeConfig)
				Expect(err).To(MatchError("space-lookup-error"))
				Expect(warnings).To(ConsistOf("warn"))
			})

			It("returns the error when GetOrganizationByName fails (--source-app + --source-space + --source-org)", func() {
				actor.getOrgByName = func(string) (resources.Organization, v7action.Warnings, error) {
					return resources.Organization{}, v7action.Warnings{"warn"}, errors.New("org-lookup-error")
				}
				f := RoutePolicySourceFlags{SourceApp: "my-app", SourceSpace: "my-space", SourceOrg: "bad-org"}
				_, _, _, err := resolveSource(f, actor, fakeConfig)
				Expect(err).To(MatchError("org-lookup-error"))
			})

			It("returns the error when GetSpaceByNameAndOrganization fails (--source-app + --source-space)", func() {
				actor.getSpaceByNameAndOrg = func(string, string) (resources.Space, v7action.Warnings, error) {
					return resources.Space{}, v7action.Warnings{"warn"}, errors.New("space-lookup-error")
				}
				f := RoutePolicySourceFlags{SourceApp: "my-app", SourceSpace: "bad-space"}
				_, _, _, err := resolveSource(f, actor, fakeConfig)
				Expect(err).To(MatchError("space-lookup-error"))
			})
		})
	})

	Describe("resolveOrgGUID", func() {
		var (
			fakeConfig *commandfakes.FakeConfig
			actor      *stubSourceActor
		)

		BeforeEach(func() {
			fakeConfig = new(commandfakes.FakeConfig)
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{GUID: "targeted-org-guid", Name: "targeted-org"})
			actor = &stubSourceActor{}
		})

		It("returns the targeted org when --source-org is not set", func() {
			f := RoutePolicySourceFlags{}
			guid, name, warnings, err := resolveOrgGUID(f, actor, fakeConfig)
			Expect(err).NotTo(HaveOccurred())
			Expect(guid).To(Equal("targeted-org-guid"))
			Expect(name).To(Equal("targeted-org"))
			Expect(warnings).To(BeEmpty())
		})

		It("resolves and returns the named org when --source-org is set", func() {
			actor.getOrgByName = func(orgName string) (resources.Organization, v7action.Warnings, error) {
				Expect(orgName).To(Equal("other-org"))
				return resources.Organization{GUID: "other-org-guid"}, v7action.Warnings{"org-warning"}, nil
			}
			f := RoutePolicySourceFlags{SourceOrg: "other-org"}
			guid, name, warnings, err := resolveOrgGUID(f, actor, fakeConfig)
			Expect(err).NotTo(HaveOccurred())
			Expect(guid).To(Equal("other-org-guid"))
			Expect(name).To(Equal("other-org"))
			Expect(warnings).To(ConsistOf("org-warning"))
		})

		It("propagates the error and warnings when org lookup fails", func() {
			actor.getOrgByName = func(string) (resources.Organization, v7action.Warnings, error) {
				return resources.Organization{}, v7action.Warnings{"warn"}, errors.New("org-error")
			}
			f := RoutePolicySourceFlags{SourceOrg: "bad-org"}
			_, _, warnings, err := resolveOrgGUID(f, actor, fakeConfig)
			Expect(err).To(MatchError("org-error"))
			Expect(warnings).To(ConsistOf("warn"))
		})
	})

	Describe("resolveSpaceGUID", func() {
		var actor *stubSourceActor

		BeforeEach(func() {
			actor = &stubSourceActor{}
		})

		It("returns the space GUID and warnings on success", func() {
			actor.getSpaceByNameAndOrg = func(spaceName, orgGUID string) (resources.Space, v7action.Warnings, error) {
				Expect(spaceName).To(Equal("my-space"))
				Expect(orgGUID).To(Equal("my-org-guid"))
				return resources.Space{GUID: "my-space-guid"}, v7action.Warnings{"space-warning"}, nil
			}
			guid, warnings, err := resolveSpaceGUID("my-space", "my-org-guid", actor)
			Expect(err).NotTo(HaveOccurred())
			Expect(guid).To(Equal("my-space-guid"))
			Expect(warnings).To(ConsistOf("space-warning"))
		})

		It("propagates the error and warnings when space lookup fails", func() {
			actor.getSpaceByNameAndOrg = func(string, string) (resources.Space, v7action.Warnings, error) {
				return resources.Space{}, v7action.Warnings{"warn"}, errors.New("space-error")
			}
			_, warnings, err := resolveSpaceGUID("bad-space", "some-org-guid", actor)
			Expect(err).To(MatchError("space-error"))
			Expect(warnings).To(ConsistOf("warn"))
		})
	})
})
