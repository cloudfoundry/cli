package actors_test

import (
	"errors"

	. "github.com/cloudfoundry/cli/cf/actors"

	"github.com/cloudfoundry/cli/cf/api/apifakes"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
)

var _ = Describe("Routes", func() {
	var (
		fakeUI              *terminal.FakeUI
		fakeRouteRepository *apifakes.FakeRouteRepository
		routeActor          RouteActor
	)

	BeforeEach(func() {
		fakeUI = &terminal.FakeUI{}
		fakeRouteRepository = new(apifakes.FakeRouteRepository)

		routeActor = NewRouteActor(fakeUI, fakeRouteRepository)
	})

	Describe("creating a random TCP route", func() {
		var (
			domain models.DomainFields
			route  models.Route
		)

		BeforeEach(func() {
			domain = models.DomainFields{
				Name: "dies-tcp.com",
			}

			route = models.Route{
				Guid: "some-guid",
			}

			fakeRouteRepository.CreateReturns(route, nil)
		})

		It("calls Create on the route repo", func() {
			routeActor.CreateRandomTCPRoute(domain)

			host, d, path, randomPort := fakeRouteRepository.CreateArgsForCall(0)
			Expect(host).To(BeEmpty())
			Expect(d).To(Equal(domain))
			Expect(path).To(BeEmpty())
			Expect(randomPort).To(BeTrue())
		})

		It("states which route it's creating", func() {
			routeActor.CreateRandomTCPRoute(domain)

			Expect(fakeUI.Outputs).To(ContainSubstrings(
				[]string{"Creating random route for dies-tcp.com..."},
			))
		})

		It("returns the route retrieved from the repository", func() {
			actualRoute := routeActor.CreateRandomTCPRoute(domain)

			Expect(actualRoute).To(Equal(route))
		})

		It("prints an error when creating the route fails", func() {
			fakeRouteRepository.CreateReturns(models.Route{}, errors.New("big bad error message"))

			var actualRoute models.Route
			Expect(func() {
				actualRoute = routeActor.CreateRandomTCPRoute(domain)
			}).To(Panic())

			Expect(fakeUI.Outputs).To(ContainSubstrings(
				[]string{"big bad error message"},
			))

			Expect(actualRoute).To(Equal(models.Route{}))
		})
	})
})
