// Taken from RakutenTech nozzle
// Thanks to them
package stats

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strconv"

	"github.com/fukata/golang-stats-api-handler"
)

const (
	// DefaultPort is default port to listen
	DefaultPort = "8080"

	// EnvPort is environmental variable to change port to listen
	EnvPort = "PORT"
)

// Server is used for various debugging.
// It opens runtime stats, pprof and appliclation stats.
type Server struct {
	Logger *log.Logger
	Stats  *Stats
}

// Start starts listening.
func (s *Server) Start() {

	http.HandleFunc("/", index)
	http.Handle("/stats/app", &statsHandler{
		stats:  s.Stats,
		logger: s.Logger,
	})
	http.HandleFunc("/stats/runtime", stats_api.Handler)

	port := DefaultPort
	if p := os.Getenv(EnvPort); p != "" {
		port = p
	}

	s.Logger.Printf("[INFO] Start server listening on :%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		s.Logger.Printf("[ERROR] Failed to start varz-server: %s", err)
	}
}

func index(w http.ResponseWriter, _ *http.Request) {
	body := `
<a href="https://github.com/cloudfoundry-community/firehose-to-syslog">firehose-to-syslog</a>
<ul>
  <li><a href="/stats/runtime">stats/runtime</a></li>
  <li><a href="/stats/app">stats/app</a></li>
  <li><a href="/debug/pprof/">pprof</a></li>
</ul>
`
	w.Header().Set("Content-type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(body))
}

type statsHandler struct {
	stats *Stats

	logger *log.Logger
}

func (h *statsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	body, err := h.stats.Json()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Internal Server Error: %s\n", err)
		return
	}
	h.logger.Printf("[DEBUG] Stats response body: %s", string(body))

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Contentt-Length", strconv.Itoa(len(body)))
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}
