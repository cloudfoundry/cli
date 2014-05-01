/*
                       WARNING WARNING WARNING

                Attention all potential contributors

   This testfile is not in the best state. We've been slowly transitioning
   from the built in "testing" package to using Ginkgo. As you can see, we've
   changed the format, but a lot of the setup, test body, descriptions, etc
   are either hardcoded, completely lacking, or misleading.

   For example:

   Describe("Testing with ginkgo"...)      // This is not a great description
   It("TestDoesSoemthing"...)              // This is a horrible description

   Describe("create-user command"...       // Describe the actual object under test
   It("creates a user when provided ..."   // this is more descriptive

   For good examples of writing Ginkgo tests for the cli, refer to

   src/github.com/cloudfoundry/cli/cf/commands/application/delete_app_test.go
   src/github.com/cloudfoundry/cli/cf/terminal/ui_test.go
   src/github.com/cloudfoundry/loggregator_consumer/consumer_test.go
*/

package configuration_test

import (
	. "github.com/cloudfoundry/cli/cf/configuration"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing with ginkgo", func() {
	It("TestDecodeTokenInfoWithoutRestoringPadding", func() {

		accessToken := "bearer eyJhbGciOiJSUzI1NiJ9.eyJqdGkiOiJjNDE4OTllNS1kZTE1LTQ5NGQtYWFiNC04ZmNlYzUxN2UwMDUiLCJzdWIiOiI3NzJkZGEzZi02NjlmLTQyNzYtYjJiZC05MDQ4NmFiZTFmNmYiLCJzY29wZSI6WyJjbG91ZF9jb250cm9sbGVyLnJlYWQiLCJjbG91ZF9jb250cm9sbGVyLndyaXRlIiwib3BlbmlkIiwicGFzc3dvcmQud3JpdGUiXSwiY2xpZW50X2lkIjoiY2YiLCJjaWQiOiJjZiIsImdyYW50X3R5cGUiOiJwYXNzd29yZCIsInVzZXJfaWQiOiI3NzJkZGEzZi02NjlmLTQyNzYtYjJiZC05MDQ4NmFiZTFmNmYiLCJ1c2VyX25hbWUiOiJ1c2VyMUBleGFtcGxlLmNvbSIsImVtYWlsIjoidXNlcjFAZXhhbXBsZS5jb20iLCJpYXQiOjEzNzcwMjgzNTYsImV4cCI6MTM3NzAzNTU1NiwiaXNzIjoiaHR0cHM6Ly91YWEuYXJib3JnbGVuLmNmLWFwcC5jb20vb2F1dGgvdG9rZW4iLCJhdWQiOlsib3BlbmlkIiwiY2xvdWRfY29udHJvbGxlciIsInBhc3N3b3JkIl19.kjFJHi0Qir9kfqi2eyhHy6kdewhicAFu8hrPR1a5AxFvxGB45slKEjuP0_72cM_vEYICgZn3PcUUkHU9wghJO9wjZ6kiIKK1h5f2K9g-Iprv9BbTOWUODu1HoLIvg2TtGsINxcRYy_8LW1RtvQc1b4dBPoopaEH4no-BIzp0E5E"
		decodedInfo, err := DecodeAccessToken(accessToken)

		Expect(err).NotTo(HaveOccurred())
		Expect(string(decodedInfo)).To(ContainSubstring("user1@example.com"))
	})
	It("TestDecodeTokenInfoWhenRestoringPadding", func() {

		accessToken := "bearer eyJhbGciOiJSUzI1NiJ9.eyJqdGkiOiIwNTg2MjlkNC04NjEwLTQ3NTEtOTg3Ny0yOGMwNzE3YTE5ZTciLCJzdWIiOiIzNGFiMDhkOC04YmVmLTQ1MzQtOGYyOC0zODhhYWI1MjAwMmEiLCJzY29wZSI6WyJjbG91ZF9jb250cm9sbGVyLnJlYWQiLCJjbG91ZF9jb250cm9sbGVyLndyaXRlIiwib3BlbmlkIiwicGFzc3dvcmQud3JpdGUiXSwiY2xpZW50X2lkIjoiY2YiLCJjaWQiOiJjZiIsImdyYW50X3R5cGUiOiJwYXNzd29yZCIsInVzZXJfaWQiOiIzNGFiMDhkOC04YmVmLTQ1MzQtOGYyOC0zODhhYWI1MjAwMmEiLCJ1c2VyX25hbWUiOiJ0bGFuZ0Bnb3Bpdm90YWwuY29tIiwiZW1haWwiOiJ0bGFuZ0Bnb3Bpdm90YWwuY29tIiwiaWF0IjoxMzc3MDk1ODM5LCJleHAiOjEzNzcxMzkwMzksImlzcyI6Imh0dHBzOi8vdWFhLnJ1bi5waXZvdGFsLmlvL29hdXRoL3Rva2VuIiwiYXVkIjpbIm9wZW5pZCIsImNsb3VkX2NvbnRyb2xsZXIiLCJwYXNzd29yZCJdfQ.dcgrGjPvTjYvg8dTSZY5ecZZTNt59IYd442VaEXXvLNB_WQCAdbVOxiJ14ogzQkkzDDw60Q2lbw4z6HrqM1a-BNpYfRmvaIP_79GpIZC6OzQy_PgA1whL27pO7_ABkSJT1CEgJQJMTQlYOiZNHvFTWen3G4O6ey680cxIN5VvbFjmmQHCuwANE9_GqnYYvoI9tS1nERku8DX2H9KH5NAgDa52-p0NhLnZRqYjGss6EyPYkwYN5w2OizfYUmEYVWo8K1Q45_TGMoE-LgZe2mGWwv0euLYBoFTkYhtBMj91dQagLrL1aGcmDKPc6ivkXtfpN4Zv7FJ9OXJ2DPQyHKRpw"
		decodedInfo, err := DecodeAccessToken(accessToken)

		Expect(err).NotTo(HaveOccurred())
		Expect(string(decodedInfo)).To(ContainSubstring("tlang@gopivotal.com"))
	})
})
