package manifestparser_test

import (
	. "code.cloudfoundry.org/cli/util/manifestparser"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Manifest", func() {
	var manifest Manifest

	BeforeEach(func() {
		manifest = Manifest{}
	})

	Describe("AppNames", func() {
		When("given a valid manifest file", func() {
			BeforeEach(func() {
				manifest.Applications = []Application{
					{Name: "app-1"},
					{Name: "app-2"},
				}
			})

			It("gets the app names", func() {
				appNames := manifest.AppNames()
				Expect(appNames).To(ConsistOf("app-1", "app-2"))
			})
		})
	})

	Describe("ContainsMultipleApps", func() {
		When("given a valid manifest file with multiple apps", func() {
			BeforeEach(func() {
				manifest.Applications = []Application{
					{Name: "app-1"},
					{Name: "app-2"}}
			})

			It("returns true", func() {
				Expect(manifest.ContainsMultipleApps()).To(BeTrue())
			})
		})

		When("given a valid manifest file with a single app", func() {
			BeforeEach(func() {
				manifest.Applications = []Application{{Name: "app-1"}}
			})

			It("returns false", func() {
				Expect(manifest.ContainsMultipleApps()).To(BeFalse())
			})
		})
	})

	Describe("ContainsPrivateDockerImages", func() {
		When("the manifest contains a docker image", func() {
			When("the image is public", func() {
				BeforeEach(func() {
					manifest.Applications = []Application{
						{Name: "app-1", Docker: &Docker{Image: "image-1"}},
						{Name: "app-2", Docker: &Docker{Image: "image-2"}}}
				})

				It("returns false", func() {
					Expect(manifest.ContainsPrivateDockerImages()).To(BeFalse())
				})
			})

			When("the image is private", func() {
				BeforeEach(func() {
					manifest.Applications = []Application{
						{Name: "app-1", Docker: &Docker{Image: "image-1"}},
						{Name: "app-2", Docker: &Docker{Image: "image-2", Username: "user"}},
					}
				})

				It("returns true", func() {
					Expect(manifest.ContainsPrivateDockerImages()).To(BeTrue())
				})
			})
		})

		When("the manifest does not contain a docker image", func() {
			BeforeEach(func() {
				manifest.Applications = []Application{
					{Name: "app-1"},
					{Name: "app-2"},
				}
			})

			It("returns false", func() {
				Expect(manifest.ContainsPrivateDockerImages()).To(BeFalse())
			})
		})
	})

	Describe("HasAppWithNoName", func() {
		It("returns true when there is an app with no name", func() {
			manifest.Applications = []Application{
				{Name: "some-app"},
				{},
			}

			Expect(manifest.HasAppWithNoName()).To(BeTrue())
		})

		It("returns false when all apps have names", func() {
			manifest.Applications = []Application{
				{Name: "some-app"},
				{Name: "some-other-app"},
			}

			Expect(manifest.HasAppWithNoName()).To(BeFalse())
		})
	})

	Describe("GetFirstAppWebProcess", func() {
		BeforeEach(func() {
			manifest.Applications = []Application{
				{
					Processes: []Process{
						{Type: "worker"},
						{Type: "web", Memory: "1G"},
					},
				},
				{
					Processes: []Process{
						{Type: "worker2"},
						{Type: "web", Memory: "2G"},
					},
				},
			}
		})

		It("returns the first app's web process", func() {
			Expect(manifest.GetFirstAppWebProcess()).To(Equal(&Process{
				Type: "web", Memory: "1G",
			}))
		})

		It("returns nil if there is no first-app-web-process", func() {
			manifest.Applications = []Application{
				{
					Name: "app1",
				},
			}

			Expect(manifest.GetFirstAppWebProcess()).To(BeNil())
		})
	})

	Describe("GetFirstApp", func() {
		BeforeEach(func() {
			manifest.Applications = []Application{
				{
					Name: "app1",
				},
				{
					Name: "app2",
				},
			}
		})

		It("returns the first app", func() {
			Expect(manifest.GetFirstApp()).To(Equal(&Application{
				Name: "app1",
			}))
		})
	})
})
