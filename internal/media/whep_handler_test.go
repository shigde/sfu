package media

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWhepOfferReq(t *testing.T) {
	t.Run("Request to start WHEP, but have no active web session", func(t *testing.T) {
		th, space, stream, _, bearer := testRouterSetup(t)

		req := newSDPContentRequest("POST", fmt.Sprintf("/space/%s/stream/%s/whep", space.Identifier, stream.UUID.String()), nil, bearer, 0)
		rr := httptest.NewRecorder()
		th.router.ServeHTTP(rr, req)

		// Then: status is 403 because no active web session
		assert.Equal(t, http.StatusForbidden, rr.Code)
	})

	t.Run("Request to start and listen WHEP", func(t *testing.T) {
		th, space, stream, _, bearer := testRouterSetup(t)
		sessionCookie, reqToken := runWhipRequest(t, th.router, space.Identifier, stream.UUID.String(), bearer)

		startReq := newSDPContentRequest("POST", fmt.Sprintf("/space/%s/stream/%s/whep", space.Identifier, stream.UUID.String()), nil, bearer, 0)
		startReq.AddCookie(sessionCookie)
		startReq.Header.Set(reqTokenHeaderName, reqToken)
		startRr := httptest.NewRecorder()
		th.router.ServeHTTP(startRr, startReq)

		assert.Equal(t, http.StatusCreated, startRr.Code)
		assert.Equal(t, "application/sdp", startRr.Header().Get("Content-Type"))
		assert.Equal(t, strconv.Itoa(len([]byte(testOffer))), startRr.Header().Get("Content-Length"))
		assert.Equal(t, testOffer, startRr.Body.String())

		answer := []byte(testAnswer)
		body := bytes.NewBuffer(answer)
		reqToken = startRr.Header().Get(reqTokenHeaderName)

		listenReq := newSDPContentRequest("PATCH", fmt.Sprintf("/space/%s/stream/%s/whep", space.Identifier, stream.UUID.String()), body, bearer, len(answer))
		listenReq.AddCookie(sessionCookie)
		listenReq.Header.Set(reqTokenHeaderName, reqToken)
		listenRr := httptest.NewRecorder()
		th.router.ServeHTTP(listenRr, listenReq)

		assert.Equal(t, http.StatusCreated, listenRr.Code)
		assert.Equal(t, "application/sdp", listenRr.Header().Get("Content-Type"))
		assert.Equal(t, "0", listenRr.Header().Get("Content-Length"))
	})
}
