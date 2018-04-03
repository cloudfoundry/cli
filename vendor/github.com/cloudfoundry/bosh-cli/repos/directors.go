package repos

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	"github.com/cloudfoundry/bosh-cli/director"
)

type DirectorsRepo interface {
	FindAll() ([]director.Director, error)
	FindBySlug(string) (director.Director, error)
}

type directorsRepo struct {
	directors []director.Director

	logger boshlog.Logger
}

func NewDirectorsRepo(directors []director.Director, logger boshlog.Logger) DirectorsRepo {
	return directorsRepo{
		directors: directors,
		logger:    logger,
	}
}

func (r directorsRepo) FindAll() ([]director.Director, error) {
	return r.directors, nil
}

func (r directorsRepo) FindBySlug(slug string) (director.Director, error) {
	for _, d := range r.directors {
		if d.Source() == slug {
			return d, nil
		}
	}

	return director.Director{}, bosherr.Errorf("Director '%s' is not found", slug)
}
