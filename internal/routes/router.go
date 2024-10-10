package routes

import (
	"context"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/shigde/sfu/internal/auth"
	"github.com/shigde/sfu/internal/auth/account"
	"github.com/shigde/sfu/internal/auth/request"
	"github.com/shigde/sfu/internal/auth/session"
	"github.com/shigde/sfu/internal/logging"
	"github.com/shigde/sfu/internal/rtp"
	"github.com/shigde/sfu/internal/stream"
	"github.com/shigde/sfu/internal/telemetry"
)

const tracerName = telemetry.TracerName

func NewRouter(
	securityConfig *session.SecurityConfig,
	rtpConfig *rtp.RtpConfig,
	accountService *account.AccountService,
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

	// Auth Routes
	auth.UseRoutes(router.PathPrefix("/api").Subrouter(), accountService)

	// Space and LiveStream Resource Endpoints
	router.HandleFunc("/space/{space}/streams", session.HttpMiddleware(securityConfig, getStreamList(streamService))).Methods("GET")
	router.HandleFunc("/space/{space}/stream/{id}", session.HttpMiddleware(securityConfig, getStream(streamService))).Methods("GET")

	// Lobby User Endpoints
	router.HandleFunc("/space/setting", request.Csrf(session.HttpMiddleware(securityConfig, getSettings(rtpConfig)))).Methods("GET")
	router.HandleFunc("/space/{space}/stream/{id}/whip", session.HttpMiddleware(securityConfig, whip(streamService, liveLobbyService))).Methods("POST")
	router.HandleFunc("/space/{space}/stream/{id}/whep", request.TokenMiddleware(whep(streamService, liveLobbyService))).Methods("POST")
	router.HandleFunc("/space/{space}/stream/{id}/res", request.TokenMiddleware(whipDelete(streamService, liveLobbyService))).Methods("DELETE")

	// RTMP Live Endpoints
	router.HandleFunc("/space/{space}/stream/{id}/live", request.TokenMiddleware(publishLiveStream(streamService, liveLobbyService))).Methods("POST")
	router.HandleFunc("/space/{space}/stream/{id}/live", request.TokenMiddleware(getStatusOfLiveStream(streamService))).Methods("GET")
	router.HandleFunc("/space/{space}/stream/{id}/live", request.TokenMiddleware(stopLiveStream(streamService, liveLobbyService))).Methods("DELETE")

	// Federartion api endpoints
	router.HandleFunc("/fed/space/{space}/stream/{id}/whep", session.HttpMiddleware(securityConfig, fedWhep(streamService, liveLobbyService))).Methods("POST")
	router.HandleFunc("/fed/space/{space}/stream/{id}/whip", session.HttpMiddleware(securityConfig, fedWhip(streamService, liveLobbyService))).Methods("POST")
	router.HandleFunc("/fed/space/{space}/stream/{id}/res", session.HttpMiddleware(securityConfig, fedResource(streamService, liveLobbyService))).Methods("DELETE")
	router.NotFoundHandler = indexHTMLWhenNotFound(http.Dir("./web/dist/web-client/browser/")) // Fallthrough for HTML5 routing
	http.Handle("/", router)
	return router
}

type spaceGetCreator interface {
	GetSpace(ctx context.Context, id string) (*stream.Space, error)
	CreateSpace(ctx context.Context, id string) (*stream.Space, error)
}
