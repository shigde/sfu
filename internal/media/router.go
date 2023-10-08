package media

import (
	"context"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/shigde/sfu/internal/auth"
	"github.com/shigde/sfu/internal/rtp"
	"github.com/shigde/sfu/internal/stream"
)

const tracerName = "github.com/shigde/sfu/internal/media"

func NewRouter(
	securityConfig *auth.SecurityConfig,
	rtpConfig *rtp.RtpConfig,
	accountServive *auth.AccountService,
	spaceManager spaceGetCreator,
) *mux.Router {
	router := mux.NewRouter()
	//cors := handlers.CORS(
	//	handlers.AllowedOrigins([]string{"http://localhost:3000/"}),
	//	handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"}),
	//	handlers.AllowedHeaders([]string{"X-Req-Token"}),
	//)
	// router.Use(cors)
	// Auth
	router.Use(func(next http.Handler) http.Handler { return handlers.LoggingHandler(os.Stdout, next) })
	router.HandleFunc("authtehnticate", getAuthenticationHandler(accountServive)).Methods("POST")

	// Space
	router.HandleFunc("/space/{space}/streams", auth.HttpMiddleware(securityConfig, getStreamList(spaceManager))).Methods("GET")
	router.HandleFunc("/space/{space}/stream", auth.HttpMiddleware(securityConfig, createStream(spaceManager))).Methods("POST")
	router.HandleFunc("/space/{space}/stream/{id}", auth.HttpMiddleware(securityConfig, getStream(spaceManager))).Methods("GET")
	router.HandleFunc("/space/{space}/stream", auth.HttpMiddleware(securityConfig, updateStream(spaceManager))).Methods("PUT")
	router.HandleFunc("/space/{space}/stream/{id}", auth.HttpMiddleware(securityConfig, deleteStream(spaceManager))).Methods("DELETE")
	// Lobby
	router.HandleFunc("/space/setting", auth.Csrf(auth.HttpMiddleware(securityConfig, getSettings(rtpConfig)))).Methods("GET")
	router.HandleFunc("/space/{space}/stream/{id}/whip", auth.HttpMiddleware(securityConfig, whip(spaceManager))).Methods("POST")
	router.HandleFunc("/space/{space}/stream/{id}/whip", auth.TokenMiddleware(whipDelete(spaceManager))).Methods("DELETE")
	router.HandleFunc("/space/{space}/stream/{id}/whep", auth.TokenMiddleware(whepOffer(spaceManager))).Methods("POST")
	router.HandleFunc("/space/{space}/stream/{id}/whep", auth.TokenMiddleware(whepAnswer(spaceManager))).Methods("PATCH")
	return router
}

type spaceGetCreator interface {
	GetSpace(ctx context.Context, id string) (*stream.Space, error)
	GetOrCreateSpace(ctx context.Context, id string) (*stream.Space, error)
}
