package uaa

// AccessToken returns the implicit grant access token
func (client *Client) AccessToken() string {
	return client.store.AccessToken()
}
