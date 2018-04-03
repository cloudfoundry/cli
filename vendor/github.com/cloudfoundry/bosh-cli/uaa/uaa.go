package uaa

import (
	gourl "net/url"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type UAAImpl struct {
	client Client
}

type TokenResp struct {
	// JTI       string `json:"jti"` // e.g. "65545f5e-5355-4e0f-bcb5-01a4b9058452"
	// Scope     string // e.g. "bosh.admin"
	// ExpiresIn int    `json:"expires_in"` // e.g. 43199

	Type         string `json:"token_type"`    // e.g. "bearer"
	AccessToken  string `json:"access_token"`  // e.g. "eyJhbGciOiJSUzI1NiJ9.eyJq<snip>fQ.Mr<snip>RawG"
	RefreshToken string `json:"refresh_token"` // e.g. "eyJhbGciOiJSUzI1NiJ9.eyJq<snip>fQ.Mr<snip>RawG"
}

func (u UAAImpl) NewStaleAccessToken(refreshValue string) StaleAccessToken {
	if len(refreshValue) == 0 {
		panic("Expected non-empty refresh token value")
	}

	return &AccessTokenImpl{
		client:       u.client,
		type_:        "bearer",
		refreshValue: refreshValue,
	}
}

func (u UAAImpl) ClientCredentialsGrant() (Token, error) {
	resp, err := u.client.ClientCredentialsGrant()
	if err != nil {
		return nil, err
	}

	token := TokenImpl{
		type_: resp.Type,
		value: resp.AccessToken,
	}

	return token, nil
}

func (u UAAImpl) OwnerPasswordCredentialsGrant(answers []PromptAnswer) (AccessToken, error) {
	resp, err := u.client.OwnerPasswordCredentialsGrant(answers)
	if err != nil {
		return nil, err
	}

	token := AccessTokenImpl{
		client:       u.client,
		type_:        resp.Type,
		accessValue:  resp.AccessToken,
		refreshValue: resp.RefreshToken,
	}

	return token, nil
}

func (c Client) ClientCredentialsGrant() (TokenResp, error) {
	query := gourl.Values{}

	query.Add("grant_type", "client_credentials")

	var resp TokenResp

	err := c.clientRequest.Post("/oauth/token", []byte(query.Encode()), &resp)
	if err != nil {
		return resp, bosherr.WrapErrorf(err, "Requesting token via client credentials grant")
	}

	return resp, nil
}

func (c Client) OwnerPasswordCredentialsGrant(answers []PromptAnswer) (TokenResp, error) {
	query := gourl.Values{}

	query.Add("grant_type", "password")

	for _, answer := range answers {
		query.Add(answer.Key, answer.Value)
	}

	var resp TokenResp

	err := c.clientRequest.Post("/oauth/token", []byte(query.Encode()), &resp)
	if err != nil {
		return resp, bosherr.WrapErrorf(err, "Requesting token via client credentials grant")
	}

	return resp, nil
}

func (c Client) RefreshTokenGrant(refreshValue string) (TokenResp, error) {
	query := gourl.Values{}

	query.Add("grant_type", "refresh_token")
	query.Add("refresh_token", refreshValue)

	var resp TokenResp

	err := c.clientRequest.Post("/oauth/token", []byte(query.Encode()), &resp)
	if err != nil {
		return resp, bosherr.WrapErrorf(err, "Refreshing token")
	}

	return resp, nil
}
