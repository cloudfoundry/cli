package helpers

import (
	"encoding/json"
	"time"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

type User struct {
	GUID      string
	Username  string
	CreatedAt time.Time
}

// GetUsers returns all the users in the targeted environment
func GetUsers() []User {
	var userPagesResponse struct {
		NextURL   *string `json:"next_url"`
		Resources []struct {
			Metadata struct {
				GUID      string    `json:"guid"`
				CreatedAt time.Time `json:"created_at"`
			} `json:"metadata"`
			Entity struct {
				Username string `json:"username"`
			} `json:"entity"`
		} `json:"resources"`
	}

	var allUsers []User
	nextURL := "/v2/users?results-per-page=50"

	for {
		session := CF("curl", nextURL)
		Eventually(session).Should(Exit(0))

		err := json.Unmarshal(session.Out.Contents(), &userPagesResponse)
		Expect(err).NotTo(HaveOccurred())
		for _, resource := range userPagesResponse.Resources {
			allUsers = append(allUsers, User{
				GUID:      resource.Metadata.GUID,
				CreatedAt: resource.Metadata.CreatedAt,
				Username:  resource.Entity.Username,
			})
		}

		if userPagesResponse.NextURL == nil {
			break
		}
		nextURL = *userPagesResponse.NextURL
	}

	return allUsers
}
