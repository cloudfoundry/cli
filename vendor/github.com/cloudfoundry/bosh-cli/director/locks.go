package director

import (
	"strconv"
	"time"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type Lock struct {
	Type      string   // e.g. "deployment"
	Resource  []string // e.g. ["some-deployment-name"]
	ExpiresAt time.Time
	TaskID    string // e.g. "123456"
}

type LockResp struct {
	Type     string   // e.g. "deployment"
	Resource []string // e.g. ["some-deployment-name"]
	Timeout  string   // e.g. "1443889622.9964118"
	TaskID   string   `json:"task_id"` // e.g. "123456"
}

type TimeoutTime time.Time

func (d DirectorImpl) Locks() ([]Lock, error) {
	var locks []Lock

	resps, err := d.client.Locks()
	if err != nil {
		return locks, err
	}

	for _, r := range resps {
		f, err := strconv.ParseFloat(r.Timeout, 64)
		if err != nil {
			return locks, bosherr.WrapErrorf(err, "Converting timeout '%s' to float", r.Timeout)
		}

		lock := Lock{
			Type:      r.Type,
			Resource:  r.Resource,
			ExpiresAt: time.Unix(int64(f), 0).UTC(),
			TaskID:    r.TaskID,
		}

		locks = append(locks, lock)
	}

	return locks, nil
}

func (c Client) Locks() ([]LockResp, error) {
	var locks []LockResp

	err := c.clientRequest.Get("/locks", &locks)
	if err != nil {
		return locks, bosherr.WrapErrorf(err, "Finding locks")
	}

	return locks, nil
}

func (l LockResp) IsForDeployment(name string) bool {
	return l.Type == "deployment" && len(l.Resource) == 1 && l.Resource[0] == name
}
