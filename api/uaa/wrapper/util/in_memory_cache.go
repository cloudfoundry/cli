package util

import "time"

type InMemoryCache struct {
	accessToken  string
	refreshToken string
	expiryDate   time.Time
}

func (c InMemoryCache) AccessToken() string {
	return c.accessToken
}

func (c *InMemoryCache) AccessTokenExpiryDate() time.Time {
	return c.expiryDate
}

func (c InMemoryCache) RefreshToken() string {
	return c.refreshToken
}

func (c *InMemoryCache) SetAccessToken(token string) {
	c.accessToken = token
}

func (c *InMemoryCache) SetAccessTokenExpiryDate(t time.Time) {
	c.expiryDate = t
}

func (c *InMemoryCache) SetRefreshToken(token string) {
	c.refreshToken = token
}

func NewInMemoryTokenCache() *InMemoryCache {
	return new(InMemoryCache)
}
