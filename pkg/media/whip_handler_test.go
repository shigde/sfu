package media

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/shigde/sfu/pkg/auth"
	"github.com/shigde/sfu/pkg/lobby"
	"github.com/shigde/sfu/pkg/stream"
	"github.com/stretchr/testify/assert"
)

func testWhipReqSetup(t *testing.T) (string, *mux.Router, *stream.LiveStreamRepository) {
	t.Helper()
	jwt := &auth.JwtToken{Enabled: true, Key: "SecretValueReplaceThis", DefaultExpireTime: 604800}
	config := &auth.AuthConfig{JWT: jwt}
	// Setup engine  mocks
	lobbyManager := lobby.NewLobbyManager()
	manager := stream.NewSpaceManager(lobbyManager)
	space := manager.GetOrCreateSpace(spaceId)

	s := &stream.LiveStream{}

	streamId := space.LiveStreamRepo.Add(s)
	router := NewRouter(config, manager)

	return streamId, router, space.LiveStreamRepo
}

func TestWhipReq(t *testing.T) {
	streamId, router, _ := testWhipReqSetup(t)

	// When: GET /streams is called
	req := newRequest("GET", fmt.Sprintf("/space/%s/streams", spaceId), nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Then: status is 200
	assert.Equal(t, http.StatusOK, rr.Code)

	// And: Body contains 1 product
	wanted := fmt.Sprintf(`[{"Id":"%s"}]%s`, streamId, "\n")
	assert.Equal(t, wanted, rr.Body.String())
}
