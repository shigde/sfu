package inbox

import (
	"context"
	"fmt"

	"github.com/shigde/sfu/internal/activitypub/models"
	"github.com/shigde/sfu/internal/activitypub/remote"
)

type handler struct {
	resolver    *remote.Resolver
	acceptInbox *acceptInbox
	createInbox *createInbox
	updateInbox *updateInbox
}

func newHandler(
	followStore *models.FollowRepository,
	resolver *remote.Resolver,
) *handler {
	return &handler{
		resolver:    resolver,
		acceptInbox: newAcceptInbox(followStore),
		createInbox: newCreateInbox(),
		updateInbox: newUpdateInbox(resolver),
	}
}

func (h *handler) resolve(ctx context.Context, request InboxRequest) error {
	if err := h.resolver.Resolve(ctx, request.Body,
		h.createInbox.handleCreateRequest,
		h.updateInbox.handleUpdateRequest,
		h.acceptInbox.handleAcceptRequest,
	); err != nil {
		return fmt.Errorf("handel resolve: %w", err)
	}
	return nil
}
