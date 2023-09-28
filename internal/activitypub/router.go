package activitypub

import (
	//"github.com/owncast/owncast/activitypub/controllers"
	//"github.com/owncast/owncast/router/middleware"

	"github.com/gorilla/mux"
	"github.com/shigde/sfu/internal/activitypub/handler"
)

// StartRouter will start the federation specific http router.
func ExtendRouter(router *mux.Router, config *FederationConfig) error {
	router.HandleFunc("/.well-known/webfinger", handler.GetWebfinger(config))

	//// Host Metadata
	//http.HandleFunc("/.well-known/host-meta", controllers.HostMetaController)
	//
	//// Nodeinfo v1
	//http.HandleFunc("/.well-known/nodeinfo", controllers.NodeInfoController)
	//
	//// x-nodeinfo v2
	//http.HandleFunc("/.well-known/x-nodeinfo2", controllers.XNodeInfo2Controller)
	//
	//// Nodeinfo v2
	//http.HandleFunc("/nodeinfo/2.0", controllers.NodeInfoV2Controller)
	//
	//// Instance details
	//http.HandleFunc("/api/v1/instance", controllers.InstanceV1Controller)
	//
	//// Single ActivityPub Actor
	//http.HandleFunc("/federation/user/", middleware.RequireActivityPubOrRedirect(controllers.ActorHandler))
	//
	//// Single AP object
	//http.HandleFunc("/federation/", middleware.RequireActivityPubOrRedirect(controllers.ObjectHandler))
	return nil
}
