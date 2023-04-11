package media

import (
	"fmt"
	"net/http"

	"github.com/shigde/sfu/pkg/auth"
	"github.com/shigde/sfu/pkg/config"
)

type MediaServer struct {
	config *config.ServerConfig
}

func NewMediaServer(config *config.ServerConfig) *MediaServer {
	return &MediaServer{config}
}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the HomePage!")
	fmt.Println("Endpoint Hit: homePage")
}
func (s *MediaServer) Serve() error {
	http.Handle("/hello", auth.HttpMiddleware(s.config.AuthConfig, homePage))
	return http.ListenAndServe(":10000", nil)
}

func (s *MediaServer) Shutdown() {
}
