package v7action_test

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"

	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/resources"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Security Group Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
		warnings                  Warnings
		executeErr                error
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v7actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil, nil, nil)

	})

	Describe("CreateSecurityGroup", func() {
		const securityGroupName = "security-group-name"
		var (
			filePath              string
			fileContents          []byte
			tempFile              *os.File
			returnedSecurityGroup resources.SecurityGroup
			secGrpPorts           string
			secGrpType            int
			secGrpCode            int
			secGrpDescription     string
			secGrpLog             bool
		)

		BeforeEach(func() {
			fileContents = []byte(`[
	{
		"protocol":"all",
		"destination":"some-destination",
		"ports":"some-ports",
		"type":1,
		"code":0,
		"description":"some-description",
		"log":false
	},
	{
      "protocol": "tcp",
      "destination": "10.10.10.0/24"
    }
]`)
			secGrpPorts = "some-ports"
			secGrpType = 1
			secGrpCode = 0
			secGrpDescription = "some-description"
			secGrpLog = false
			returnedSecurityGroup = resources.SecurityGroup{
				Name: securityGroupName,
				GUID: "some-sec-grp-guid",
				Rules: []resources.Rule{
					{
						Protocol:    "all",
						Destination: "some-destination",
						Ports:       &secGrpPorts,
						Type:        &secGrpType,
						Code:        &secGrpCode,
						Description: &secGrpDescription,
						Log:         &secGrpLog,
					},
					{
						Protocol:    "tcp",
						Destination: "10.10.10.0/24",
					},
				},
			}
			tempFile, executeErr = ioutil.TempFile("", "")
			Expect(executeErr).ToNot(HaveOccurred())
			filePath = tempFile.Name()

			fakeCloudControllerClient.CreateSecurityGroupReturns(returnedSecurityGroup, ccv3.Warnings{"create-sec-grp-warning"}, nil)
		})

		JustBeforeEach(func() {
			_, err := tempFile.Write(fileContents)
			Expect(err).ToNot(HaveOccurred())

			warnings, executeErr = actor.CreateSecurityGroup(securityGroupName, filePath)
		})

		AfterEach(func() {
			os.Remove(filePath)
		})

		When("the path does not exist", func() {
			BeforeEach(func() {
				filePath = "does-not-exist"
			})
			It("returns an error", func() {
				Expect(executeErr).To(HaveOccurred())
				_, ok := executeErr.(*os.PathError)
				Expect(ok).To(BeTrue())
				Expect(warnings).To(Equal(Warnings{}))
			})
		})

		When("Unmarshaling fails", func() {
			BeforeEach(func() {
				fileContents = []byte("not-valid-json")
			})
			It("returns an error", func() {
				Expect(executeErr).To(HaveOccurred())
				_, ok := executeErr.(*json.SyntaxError)
				Expect(ok).To(BeTrue())
				Expect(warnings).To(Equal(Warnings{}))
			})
		})

		It("calls the API with the generated security group resource and returns all warnings", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(warnings).To(Equal(Warnings{"create-sec-grp-warning"}))

			givenSecurityGroup := fakeCloudControllerClient.CreateSecurityGroupArgsForCall(0)

			returnedSecurityGroup.GUID = ""
			Expect(givenSecurityGroup).To(Equal(returnedSecurityGroup))

		})

		When("the security group can't be created", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CreateSecurityGroupReturns(resources.SecurityGroup{}, ccv3.Warnings{"a-warning"}, errors.New("create-sec-group-error"))
			})
			It("returns the error and warnings", func() {
				Expect(executeErr).To(HaveOccurred())
				Expect(executeErr).To(MatchError("create-sec-group-error"))
				Expect(warnings).To(Equal(Warnings{"a-warning"}))
			})

		})
	})
})
