package v7pushaction_test

import (
	"errors"
	"io/ioutil"
	"path/filepath"
	"time"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	. "code.cloudfoundry.org/cli/actor/v7pushaction"
	"code.cloudfoundry.org/cli/actor/v7pushaction/v7pushactionfakes"
	"code.cloudfoundry.org/cli/util/manifestparser"
	"github.com/cloudfoundry/bosh-cli/director/template"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

func PrepareSpaceStreamsDrainedAndClosed(
	appNameStream <-chan []string,
	eventStream <-chan Event,
	warningsStream <-chan Warnings,
	errorStream <-chan error,
) bool {
	var configStreamClosed, eventStreamClosed, warningsStreamClosed, errorStreamClosed bool
	for {
		select {
		case _, ok := <-appNameStream:
			if !ok {
				configStreamClosed = true
			}
		case _, ok := <-eventStream:
			if !ok {
				eventStreamClosed = true
			}
		case _, ok := <-warningsStream:
			if !ok {
				warningsStreamClosed = true
			}
		case _, ok := <-errorStream:
			if !ok {
				errorStreamClosed = true
			}
		}
		if configStreamClosed && eventStreamClosed && warningsStreamClosed && errorStreamClosed {
			break
		}
	}
	return true
}

func getPrepareNextEvent(c <-chan []string, e <-chan Event, w <-chan Warnings) func() Event {
	timeOut := time.Tick(500 * time.Millisecond)

	return func() Event {
		for {
			select {
			case <-c:
			case event, ok := <-e:
				if ok {
					log.WithField("event", event).Debug("getNextEvent")
					return event
				}
				return ""
			case <-w:
			case <-timeOut:
				return ""
			}
		}
	}
}

var _ = Describe("PrepareSpace", func() {
	var (
		actor       *Actor
		fakeV7Actor *v7pushactionfakes.FakeV7Actor

		spaceGUID string
		appName   string
		parser    *manifestparser.Parser
		manifest  []byte

		appNameStream  <-chan []string
		eventStream    <-chan Event
		warningsStream <-chan Warnings
		errorStream    <-chan error
	)

	BeforeEach(func() {
		fakeV7Actor = new(v7pushactionfakes.FakeV7Actor)
		actor, _, fakeV7Actor, _ = getTestPushActor()

		parser = manifestparser.NewParser()
		spaceGUID = "some-space-guid"
		appName = "some-app-name"
	})

	AfterEach(func() {
		Eventually(PrepareSpaceStreamsDrainedAndClosed(appNameStream, eventStream, warningsStream, errorStream)).Should(BeTrue())
	})

	JustBeforeEach(func() {
		appNameStream, eventStream, warningsStream, errorStream = actor.PrepareSpace(spaceGUID, appName, parser)
	})

	var yamlUnmarshalMarshal = func(b []byte) []byte {
		var obj interface{}
		yaml.Unmarshal(b, &obj)
		postMarshal, err := yaml.Marshal(obj)
		Expect(err).ToNot(HaveOccurred())
		return postMarshal
	}
	When("A single app manifest is present", func() {
		BeforeEach(func() {
			tempDir, err := ioutil.TempDir("", "conceptualize-unit")

			manifest = []byte("---\napplications:\n- name: some-app-name")
			pathToYAMLFile := filepath.Join(tempDir, "manifest.yml")
			err = ioutil.WriteFile(pathToYAMLFile, manifest, 0644)
			Expect(err).ToNot(HaveOccurred())

			err = parser.InterpolateAndParse(pathToYAMLFile, []string{}, []template.VarKV{})
			Expect(err).ToNot(HaveOccurred())
			fakeV7Actor.SetSpaceManifestReturns(v7action.Warnings{"set-space-warning"}, nil)
		})

		It("applies the manifest", func() {
			Consistently(fakeV7Actor.CreateApplicationInSpaceCallCount).Should(Equal(0))
			Eventually(getPrepareNextEvent(appNameStream, eventStream, warningsStream)).Should(Equal(ApplyManifest))
			Eventually(fakeV7Actor.SetSpaceManifestCallCount).Should(Equal(1))
			actualSpaceGuid, actualManifestBytes := fakeV7Actor.SetSpaceManifestArgsForCall(0)
			Expect(actualSpaceGuid).To(Equal(spaceGUID))
			Expect(actualManifestBytes).To(Equal(yamlUnmarshalMarshal(manifest)))
			Eventually(warningsStream).Should(Receive(Equal(Warnings{"set-space-warning"})))
			Eventually(errorStream).Should(Receive(Succeed()))
			Eventually(appNameStream).Should(Receive(ConsistOf("some-app-name")))
			Eventually(getPrepareNextEvent(appNameStream, eventStream, warningsStream)).Should(Equal(ApplyManifestComplete))
		})
	})
	// When("there is also an appName provided", func() {
	// 	It("errors", func() {

	// 	})
	// })

	When("there are multiple applications in the manifest", func() {

		BeforeEach(func() {
			tempDir, err := ioutil.TempDir("", "conceptualize-unit")

			manifest = []byte(`---
applications:
- name: some-app-name
- name: orange
- name: mushroom
`)
			pathToYAMLFile := filepath.Join(tempDir, "manifest.yml")
			err = ioutil.WriteFile(pathToYAMLFile, manifest, 0644)
			Expect(err).ToNot(HaveOccurred())

			err = parser.InterpolateAndParse(pathToYAMLFile, []string{}, []template.VarKV{})
			Expect(err).ToNot(HaveOccurred())

			fakeV7Actor.SetSpaceManifestReturns(v7action.Warnings{"set-space-warning"}, nil)
		})

		It("applies the manifest", func() {
			Consistently(fakeV7Actor.CreateApplicationInSpaceCallCount).Should(Equal(0))
			Eventually(getPrepareNextEvent(appNameStream, eventStream, warningsStream)).Should(Equal(ApplyManifest))
			Eventually(fakeV7Actor.SetSpaceManifestCallCount).Should(Equal(1))
			actualSpaceGuid, actualManifestBytes := fakeV7Actor.SetSpaceManifestArgsForCall(0)
			Expect(actualSpaceGuid).To(Equal(spaceGUID))
			Expect(actualManifestBytes).To(Equal(yamlUnmarshalMarshal(manifest)))
			Eventually(warningsStream).Should(Receive(Equal(Warnings{"set-space-warning"})))
			Eventually(errorStream).Should(Receive(Succeed()))
			Eventually(appNameStream).Should(Receive(ConsistOf("some-app-name", "orange", "mushroom")))
			Eventually(getPrepareNextEvent(appNameStream, eventStream, warningsStream)).Should(Equal(ApplyManifestComplete))
		})
	})

	When("There is not a manifest and the app doesnt exist", func() {
		BeforeEach(func() {
			fakeV7Actor.CreateApplicationInSpaceReturns(
				v7action.Application{Name: "some-app-name"},
				v7action.Warnings{"create-app-warning"},
				nil,
			)
		})

		It("does not apply the manifest", func() {
			Consistently(fakeV7Actor.SetSpaceManifestCallCount).Should(Equal(0))
			Eventually(getPrepareNextEvent(appNameStream, eventStream, warningsStream)).Should(Equal(CreatingApplication))
			Eventually(fakeV7Actor.CreateApplicationInSpaceCallCount).Should(Equal(1))
			actualApp, actualSpaceGuid := fakeV7Actor.CreateApplicationInSpaceArgsForCall(0)
			Expect(actualApp.Name).To(Equal(appName))
			Expect(actualSpaceGuid).To(Equal(spaceGUID))
			Eventually(warningsStream).Should(Receive(Equal(Warnings{"create-app-warning"})))
			Eventually(errorStream).Should(Receive(Succeed()))
			Eventually(appNameStream).Should(Receive(ConsistOf("some-app-name")))
			Eventually(getPrepareNextEvent(appNameStream, eventStream, warningsStream)).Should(Equal(CreatedApplication))
		})
	})

	When("There is not a manifest and the app exists", func() {
		BeforeEach(func() {
			fakeV7Actor.CreateApplicationInSpaceReturns(
				v7action.Application{},
				v7action.Warnings{"create-app-warning"},
				actionerror.ApplicationAlreadyExistsError{Name: "some-app-name"},
			)
		})

		It("does not apply the manifest", func() {
			Consistently(fakeV7Actor.SetSpaceManifestCallCount).Should(Equal(0))
			Eventually(getPrepareNextEvent(appNameStream, eventStream, warningsStream)).Should(Equal(CreatingApplication))
			Eventually(fakeV7Actor.CreateApplicationInSpaceCallCount).Should(Equal(1))
			actualApp, actualSpaceGuid := fakeV7Actor.CreateApplicationInSpaceArgsForCall(0)
			Expect(actualApp.Name).To(Equal(appName))
			Expect(actualSpaceGuid).To(Equal(spaceGUID))
			Eventually(warningsStream).Should(Receive(Equal(Warnings{"create-app-warning"})))
			Eventually(errorStream).Should(Receive(Succeed()))
			Eventually(appNameStream).Should(Receive(ConsistOf("some-app-name")))
			Eventually(getPrepareNextEvent(appNameStream, eventStream, warningsStream)).Should(Equal(ApplicationAlreadyExists))
		})
	})

	When("Applying the manifest errors", func() {
		BeforeEach(func() {
			tempDir, err := ioutil.TempDir("", "conceptualize-unit")

			manifest = []byte("---\napplications:\n- name: some-app-name")
			pathToYAMLFile := filepath.Join(tempDir, "manifest.yml")
			err = ioutil.WriteFile(pathToYAMLFile, manifest, 0644)
			Expect(err).ToNot(HaveOccurred())

			err = parser.InterpolateAndParse(pathToYAMLFile, []string{}, []template.VarKV{})
			Expect(err).ToNot(HaveOccurred())
			fakeV7Actor.SetSpaceManifestReturns(v7action.Warnings{"set-space-warning"}, errors.New("some-error"))
		})

		It("returns the error and exits", func() {
			Consistently(fakeV7Actor.CreateApplicationInSpaceCallCount).Should(Equal(0))
			Eventually(getPrepareNextEvent(appNameStream, eventStream, warningsStream)).Should(Equal(ApplyManifest))
			Eventually(fakeV7Actor.SetSpaceManifestCallCount).Should(Equal(1))
			actualSpaceGuid, actualManifestBytes := fakeV7Actor.SetSpaceManifestArgsForCall(0)
			Expect(actualSpaceGuid).To(Equal(spaceGUID))
			Expect(actualManifestBytes).To(Equal(yamlUnmarshalMarshal(manifest)))
			Eventually(warningsStream).Should(Receive(Equal(Warnings{"set-space-warning"})))
			Eventually(errorStream).Should(Receive(Equal(errors.New("some-error"))))
			Consistently(appNameStream).ShouldNot(Receive(ConsistOf("some-app-name")))
			Consistently(getPrepareNextEvent(appNameStream, eventStream, warningsStream)).ShouldNot(Equal(ApplyManifestComplete))
		})
	})

	When("There is not a manifest and creating the app fails", func() {
		BeforeEach(func() {
			fakeV7Actor.CreateApplicationInSpaceReturns(
				v7action.Application{},
				v7action.Warnings{"create-app-warning"},
				errors.New("some-create-error"),
			)
		})

		It("does not apply the manifest", func() {
			Consistently(fakeV7Actor.SetSpaceManifestCallCount).Should(Equal(0))
			Eventually(getPrepareNextEvent(appNameStream, eventStream, warningsStream)).Should(Equal(CreatingApplication))
			Eventually(fakeV7Actor.CreateApplicationInSpaceCallCount).Should(Equal(1))
			actualApp, actualSpaceGuid := fakeV7Actor.CreateApplicationInSpaceArgsForCall(0)
			Expect(actualApp.Name).To(Equal(appName))
			Expect(actualSpaceGuid).To(Equal(spaceGUID))
			Eventually(warningsStream).Should(Receive(Equal(Warnings{"create-app-warning"})))
			Eventually(errorStream).Should(Receive(Equal(errors.New("some-create-error"))))
			Consistently(getPrepareNextEvent(appNameStream, eventStream, warningsStream)).ShouldNot(Equal(ApplicationAlreadyExists))
		})
	})
})
