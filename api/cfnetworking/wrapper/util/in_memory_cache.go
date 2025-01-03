package util

type InMemoryCache struct {
	accessToken  string
	refreshToken string
}

func (c InMemoryCache) AccessToken() string {
	return c.accessToken
}

func (c InMemoryCache) RefreshToken() string {
	return c.refreshToken
}

func (c *InMemoryCache) SetAccessToken(token string) {
	c.accessToken = token
}

func (c *InMemoryCache) SetRefreshToken(token string) {
	c.refreshToken = token
}

func NewInMemoryTokenCache() *InMemoryCache {
	return new(InMemoryCache)
}
