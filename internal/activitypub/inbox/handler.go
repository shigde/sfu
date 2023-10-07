package inbox

import (
	"context"
	"fmt"

	"github.com/shigde/sfu/internal/activitypub/models"
	"github.com/shigde/sfu/internal/activitypub/remote"
	"github.com/shigde/sfu/internal/activitypub/services"
)

type handler struct {
	resolver      *remote.Resolver
	acceptInbox   *acceptInbox
	announceInbox *announceInbox
	updateInbox   *updateInbox
	deleteInbox   *deleteInbox
}

func newHandler(
	followRep *models.FollowRepository,
	videoService *services.VideoService,
	resolver *remote.Resolver,
) *handler {
	return &handler{
		resolver:      resolver,
		acceptInbox:   newAcceptInbox(followRep),
		announceInbox: newAnnounceInbox(videoService),
		updateInbox:   newUpdateInbox(videoService),
		deleteInbox:   newDeleteInbox(videoService),
	}
}

func (h *handler) resolve(ctx context.Context, request InboxRequest) error {
	if err := h.resolver.Resolve(ctx, request.Body,
		h.announceInbox.handleAnnounceRequest,
		h.updateInbox.handleUpdateRequest,
		h.deleteInbox.handleDeleteRequest,
		h.acceptInbox.handleAcceptRequest,
	); err != nil {
		return fmt.Errorf("handel resolve: %w", err)
	}
	return nil
}
