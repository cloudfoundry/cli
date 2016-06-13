package actors_test

import (
	"errors"

	. "github.com/cloudfoundry/cli/cf/actors"

	"github.com/cloudfoundry/cli/cf/api/apifakes"
	"github.com/cloudfoundry/cli/cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/cli/cf/terminal/terminalfakes"
)

var _ = Describe("Routes", func() {
	var (
		fakeUI              *terminalfakes.FakeUI
		fakeRouteRepository *apifakes.FakeRouteRepository
		routeActor          RouteActor
	)

	BeforeEach(func() {
		fakeUI = &terminalfakes.FakeUI{}
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
				GUID: "some-guid",
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

		It("states that a route is being created", func() {
			routeActor.CreateRandomTCPRoute(domain)

			Expect(fakeUI.SayCallCount()).To(Equal(1))
			Expect(fakeUI.SayArgsForCall(0)).To(ContainSubstring("Creating random route for"))
		})

		It("returns the route retrieved from the repository", func() {
			actualRoute, err := routeActor.CreateRandomTCPRoute(domain)
			Expect(err).NotTo(HaveOccurred())

			Expect(actualRoute).To(Equal(route))
		})

		It("prints an error when creating the route fails", func() {
			expectedError := errors.New("big bad error message")
			fakeRouteRepository.CreateReturns(models.Route{}, expectedError)

			actualRoute, err := routeActor.CreateRandomTCPRoute(domain)
			Expect(err).To(Equal(expectedError))
			Expect(actualRoute).To(Equal(models.Route{}))
		})
	})
})
