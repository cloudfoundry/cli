package healthcheck

import (
	"net/http"

	"github.com/tedsuo/rata"

	"code.cloudfoundry.org/lager"
)

type HealthCheckHandler struct {
	logger lager.Logger
}

func NewHandler(logger lager.Logger) http.Handler {
	routes := rata.Routes{
		{Name: "HealthCheck", Method: "GET", Path: "/"},
	}

	logger = logger.Session("healthcheck")

	actions := map[string]http.Handler{
		"HealthCheck": &HealthCheckHandler{logger: logger},
	}

	handler, err := rata.NewRouter(routes, actions)
	if err != nil {
		panic(err)
	}

	return handler
}

func (h *HealthCheckHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	h.logger.Debug("started")
	defer h.logger.Debug("finished")
	writer.WriteHeader(http.StatusOK)
}
