package director

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

/*
{
  "version": "1.3167.0 (00000000)",
  "uuid": "a57b8733-163e-443a-a3d7-9ac58a886d7b",
  "name": "my-bosh",

  "cpi": "vsphere_cpi",
  "user": null,

  "features": {
    "snapshots": {
      "status": false
    },
    "compiled_package_cache": {
      "extras": { "provider": null },
      "status": false
    },
    "dns": {
      "extras": { "domain_name": "bosh" },
      "status": false
    }
  },

  "user_authentication": {
    "options": { "url": "https://10.244.3.2:8443" },
    "type": "uaa"
  }
}
*/

type Info struct {
	Name    string
	UUID    string
	Version string

	User string
	Auth UserAuthentication

	Features map[string]bool

	CPI string
}

type UserAuthentication struct {
	Type    string
	Options map[string]interface{}
}

type InfoResp struct {
	Name    string // e.g. "Bosh Lite Director"
	UUID    string // e.g. "71d36859-4f21-446f-8a02-f18d7f1263c6"
	Version string // e.g. "1.2922.0 (00000000)"

	User string
	Auth UserAuthenticationResp `json:"user_authentication"`

	Features map[string]InfoFeatureResp

	CPI string
}

type InfoFeatureResp struct {
	Status bool
	// ignore extras
}

type UserAuthenticationResp struct {
	Type    string
	Options map[string]interface{}
}

func (d DirectorImpl) IsAuthenticated() (bool, error) {
	r, err := d.client.Info()
	if err != nil {
		return false, err
	}

	authed := len(r.User) > 0

	return authed, nil
}

func (d DirectorImpl) Info() (Info, error) {
	r, err := d.client.Info()
	if err != nil {
		return Info{}, err
	}

	info := Info{
		Name:    r.Name,
		UUID:    r.UUID,
		Version: r.Version,

		User: r.User,
		Auth: UserAuthentication{
			Type:    r.Auth.Type,
			Options: r.Auth.Options,
		},

		Features: map[string]bool{},

		CPI: r.CPI,
	}

	for k, featResp := range r.Features {
		info.Features[k] = featResp.Status
	}

	return info, nil
}

func (c Client) Info() (InfoResp, error) {
	var info InfoResp

	err := c.clientRequest.Get("/info", &info)
	if err != nil {
		return info, bosherr.WrapErrorf(err, "Fetching info")
	}

	return info, nil
}
