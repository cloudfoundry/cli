package web

import (
	"encoding/json"
	"io"
	"net/http"
)

func NewListPluginsHandler(plugins PluginsJson, logger io.Writer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := json.Marshal(plugins)
		if err != nil {
			logger.Write([]byte("Error marshalling plugin models: " + err.Error()))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(b)
	})
}
