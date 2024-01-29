package media

import (
	"context"

	"github.com/gorilla/mux"
	"github.com/shigde/sfu/internal/auth"
	"github.com/shigde/sfu/internal/logging"
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
	router.Use(logging.LoggingMiddleware)
	router.HandleFunc("/authenticate", getAuthenticationHandler(accountService)).Methods("POST")
	// Space and LiveStream Resource Endpoints
	router.HandleFunc("/space/{space}/streams", auth.HttpMiddleware(securityConfig, getStreamList(streamService))).Methods("GET")
	router.HandleFunc("/space/{space}/stream/{id}", auth.HttpMiddleware(securityConfig, getStream(streamService))).Methods("GET")

	// Lobby User Endpoints
	router.HandleFunc("/space/setting", auth.Csrf(auth.HttpMiddleware(securityConfig, getSettings(rtpConfig)))).Methods("GET")
	router.HandleFunc("/space/{space}/stream/{id}/whip", auth.HttpMiddleware(securityConfig, whip(streamService, liveLobbyService))).Methods("POST")
	router.HandleFunc("/space/{space}/stream/{id}/whip", auth.TokenMiddleware(whipDelete(streamService, liveLobbyService))).Methods("DELETE")
	router.HandleFunc("/space/{space}/stream/{id}/whep", auth.TokenMiddleware(whepOffer(streamService, liveLobbyService))).Methods("POST")
	router.HandleFunc("/space/{space}/stream/{id}/whep", auth.TokenMiddleware(whepAnswer(streamService, liveLobbyService))).Methods("PATCH")

	// Lobby User Live Endpoints
	router.HandleFunc("/space/{space}/stream/{id}/live", auth.TokenMiddleware(publishLiveStream(streamService, liveLobbyService))).Methods("POST")
	router.HandleFunc("/space/{space}/stream/{id}/live", auth.TokenMiddleware(getStatusOfLiveStream(streamService))).Methods("GET")
	router.HandleFunc("/space/{space}/stream/{id}/live", auth.TokenMiddleware(stopLiveStream(streamService, liveLobbyService))).Methods("DELETE")

	// Static Stream listeners
	router.HandleFunc("/space/{space}/stream/{id}/static/whep", auth.HttpMiddleware(securityConfig, whepStaticAnswer(streamService, liveLobbyService))).Methods("POST")

	// Server to Server endpoints
	router.HandleFunc("/space/{space}/stream/{id}/pipe", auth.HttpMiddleware(securityConfig, openPipe(streamService, liveLobbyService))).Methods("POST")
	router.HandleFunc("/space/{space}/stream/{id}/hostingress", auth.HttpMiddleware(securityConfig, openHostIngress(streamService, liveLobbyService))).Methods("POST")
	router.HandleFunc("/space/{space}/stream/{id}/pipe", auth.HttpMiddleware(securityConfig, closePipe(streamService, liveLobbyService))).Methods("DELETE")

	return router
}

type spaceGetCreator interface {
	GetSpace(ctx context.Context, id string) (*stream.Space, error)
	CreateSpace(ctx context.Context, id string) (*stream.Space, error)
}
