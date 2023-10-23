package media

import (
	"context"

	"github.com/gorilla/mux"
	"github.com/shigde/sfu/internal/auth"
	"github.com/shigde/sfu/internal/rtp"
	"github.com/shigde/sfu/internal/stream"
)

const tracerName = "github.com/shigde/sfu/internal/media"

func NewRouter(
	securityConfig *auth.SecurityConfig,
	rtpConfig *rtp.RtpConfig,
	accountService *auth.AccountService,
	streamService *stream.LiveStreamService,
	liveLobbyService *stream.LiveLobbyService,
) *mux.Router {
	router := mux.NewRouter()
	//cors := handlers.CORS(
	//	handlers.AllowedOrigins([]string{"http://localhost:3000/"}),
	//	handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"}),
	//	handlers.AllowedHeaders([]string{"X-Req-Token"}),
	//)
	// router.Use(cors)
	// Auth
	// router.Use(func(next http.Handler) http.Handler { return handlers.LoggingHandler(os.Stdout, next) })
	router.HandleFunc("/authenticate", getAuthenticationHandler(accountService)).Methods("POST")
	// Space
	router.HandleFunc("/space/{space}/streams", auth.HttpMiddleware(securityConfig, getStreamList(streamService))).Methods("GET")
	router.HandleFunc("/space/{space}/stream", auth.HttpMiddleware(securityConfig, createStream(streamService))).Methods("POST")
	router.HandleFunc("/space/{space}/stream/{id}", auth.HttpMiddleware(securityConfig, getStream(streamService))).Methods("GET")
	router.HandleFunc("/space/{space}/stream", auth.HttpMiddleware(securityConfig, updateStream(streamService))).Methods("PUT")
	router.HandleFunc("/space/{space}/stream/{id}", auth.HttpMiddleware(securityConfig, deleteStream(streamService))).Methods("DELETE")
	// Lobby
	router.HandleFunc("/space/setting", auth.Csrf(auth.HttpMiddleware(securityConfig, getSettings(rtpConfig)))).Methods("GET")
	router.HandleFunc("/space/{space}/stream/{id}/whip", auth.HttpMiddleware(securityConfig, whip(streamService, liveLobbyService))).Methods("POST")
	router.HandleFunc("/space/{space}/stream/{id}/whip", auth.TokenMiddleware(whipDelete(streamService, liveLobbyService))).Methods("DELETE")
	router.HandleFunc("/space/{space}/stream/{id}/whep", auth.TokenMiddleware(whepOffer(streamService, liveLobbyService))).Methods("POST")
	router.HandleFunc("/space/{space}/stream/{id}/whep", auth.TokenMiddleware(whepAnswer(streamService, liveLobbyService))).Methods("PATCH")
	// Live
	router.HandleFunc("/space/{space}/stream/{id}/live", auth.TokenMiddleware(publishLiveStream(streamService, liveLobbyService))).Methods("POST")
	router.HandleFunc("/space/{space}/stream/{id}/live/{liveId}", auth.TokenMiddleware(getStatusOfLiveStream(streamService, liveLobbyService))).Methods("GET")
	router.HandleFunc("/space/{space}/stream/{id}/live/{liveId}", auth.TokenMiddleware(stopLiveStream(streamService, liveLobbyService))).Methods("DELETE")
	return router
}

type spaceGetCreator interface {
	GetSpace(ctx context.Context, id string) (*stream.Space, error)
	CreateSpace(ctx context.Context, id string) (*stream.Space, error)
}
