package media

import (
	"context"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/shigde/sfu/internal/auth"
	"github.com/shigde/sfu/internal/logging"
	"github.com/shigde/sfu/internal/rtp"
	"github.com/shigde/sfu/internal/stream"
	"github.com/shigde/sfu/internal/telemetry"
)

const tracerName = telemetry.TracerName

func NewRouter(
	securityConfig *auth.SecurityConfig,
	rtpConfig *rtp.RtpConfig,
	accountService *auth.AccountService,
	streamService *stream.LiveStreamService,
	liveLobbyService *stream.LiveLobbyService,
) *mux.Router {
	router := mux.NewRouter()
	cors := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
	)
	//	handlers.AllowedOrigins([]string{"http://localhost:3000/"}),
	//	handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"}),
	//	handlers.AllowedHeaders([]string{"X-Req-Token"}),
	//)
	router.Use(cors)
	// Auth
	router.Use(logging.LoggingMiddleware)

	router.HandleFunc("/authenticate", getAuthenticationHandler(accountService)).Methods("POST")
	// Space and LiveStream Resource Endpoints
	router.HandleFunc("/space/{space}/streams", auth.HttpMiddleware(securityConfig, getStreamList(streamService))).Methods("GET")
	router.HandleFunc("/space/{space}/stream/{id}", auth.HttpMiddleware(securityConfig, getStream(streamService))).Methods("GET")

	// Lobby User Endpoints
	router.HandleFunc("/space/setting", auth.Csrf(auth.HttpMiddleware(securityConfig, getSettings(rtpConfig)))).Methods("GET")
	router.HandleFunc("/space/{space}/stream/{id}/whip", auth.HttpMiddleware(securityConfig, whip(streamService, liveLobbyService))).Methods("POST")
	router.HandleFunc("/space/{space}/stream/{id}/whep", auth.TokenMiddleware(whep(streamService, liveLobbyService))).Methods("POST")
	router.HandleFunc("/space/{space}/stream/{id}/res", auth.TokenMiddleware(whipDelete(streamService, liveLobbyService))).Methods("DELETE")

	// RTMP Live Endpoints
	router.HandleFunc("/space/{space}/stream/{id}/live", auth.TokenMiddleware(publishLiveStream(streamService, liveLobbyService))).Methods("POST")
	router.HandleFunc("/space/{space}/stream/{id}/live", auth.TokenMiddleware(getStatusOfLiveStream(streamService))).Methods("GET")
	router.HandleFunc("/space/{space}/stream/{id}/live", auth.TokenMiddleware(stopLiveStream(streamService, liveLobbyService))).Methods("DELETE")

	// Federartion api endpoints
	router.HandleFunc("/fed/space/{space}/stream/{id}/whep", auth.HttpMiddleware(securityConfig, fedWhep(streamService, liveLobbyService))).Methods("POST")
	router.HandleFunc("/fed/space/{space}/stream/{id}/whip", auth.HttpMiddleware(securityConfig, fedWhip(streamService, liveLobbyService))).Methods("POST")
	router.HandleFunc("/fed/space/{space}/stream/{id}/res", auth.HttpMiddleware(securityConfig, fedResource(streamService, liveLobbyService))).Methods("DELETE")
	router.NotFoundHandler = indexHTMLWhenNotFound(http.Dir("./web")) // Fallthrough for HTML5 routing
	http.Handle("/", router)
	return router
}

type spaceGetCreator interface {
	GetSpace(ctx context.Context, id string) (*stream.Space, error)
	CreateSpace(ctx context.Context, id string) (*stream.Space, error)
}
