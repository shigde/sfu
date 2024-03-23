package media

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/shigde/sfu/internal/media/mocks"
	"github.com/stretchr/testify/assert"
)

func TestWhepReq(t *testing.T) {
	t.Run("Request to start WHEP, but have no active web session", func(t *testing.T) {
		th, space, stream, _, bearer := testRouterSetup(t)

		req := newSDPContentRequest("POST", fmt.Sprintf("/space/%s/stream/%s/whep", space.Identifier, stream.UUID.String()), nil, bearer, 0)
		rr := httptest.NewRecorder()
		th.router.ServeHTTP(rr, req)

		// Then: status is 403 because no active web session
		assert.Equal(t, http.StatusForbidden, rr.Code)
	})

	t.Run("Request WHEP resource", func(t *testing.T) {
		th, space, stream, _, bearer := testRouterSetup(t)
		sessionCookie, reqToken := runWhipRequest(t, th.router, space.Identifier, stream.UUID.String(), bearer)
		offer := []byte(mocks.Offer)
		body := bytes.NewBuffer(offer)

		startReq := newSDPContentRequest("POST", fmt.Sprintf("/space/%s/stream/%s/whep", space.Identifier, stream.UUID.String()), body, bearer, len(offer))
		startReq.AddCookie(sessionCookie)
		startReq.Header.Set(mocks.ReqTokenHeaderName, reqToken)
		startRr := httptest.NewRecorder()
		th.router.ServeHTTP(startRr, startReq)

		assert.Equal(t, http.StatusCreated, startRr.Code)
		assert.Equal(t, "application/sdp", startRr.Header().Get("Content-Type"))
		assert.Equal(t, strconv.Itoa(len([]byte(mocks.Answer))), startRr.Header().Get("Content-Length"))
		assert.Equal(t, mocks.Answer, startRr.Body.String())
	})
}

//func TestWhepStaticOfferReq(t *testing.T) {
//	t.Run("Static WHEP Request without offer", func(t *testing.T) {
//		th, space, stream, _, bearer := testRouterSetup(t)
//
//		req := newSDPContentRequest("POST", fmt.Sprintf("/space/%s/stream/%s/static/whep", space.Identifier, stream.UUID.String()), nil, bearer, 0)
//		rr := httptest.NewRecorder()
//		th.router.ServeHTTP(rr, req)
//		// Then: status is 400 because payload empty
//		assert.Equal(t, http.StatusBadRequest, rr.Code)
//	})
//
//	t.Run("Static WHEP Request offer", func(t *testing.T) {
//		th, space, stream, _, bearer := testRouterSetup(t)
//
//		offer := []byte(testOffer)
//		body := bytes.NewBuffer(offer)
//		startReq := newSDPContentRequest("POST", fmt.Sprintf("/space/%s/stream/%s/static/whep", space.Identifier, stream.UUID.String()), body, bearer, len(offer))
//		startRr := httptest.NewRecorder()
//		th.router.ServeHTTP(startRr, startReq)
//
//		assert.Equal(t, http.StatusCreated, startRr.Code)
//		assert.Equal(t, "application/sdp", startRr.Header().Get("Content-Type"))
//		assert.Equal(t, strconv.Itoa(len([]byte(testAnswer))), startRr.Header().Get("Content-Length"))
//		assert.Equal(t, testAnswer, startRr.Body.String())
//	})
//}
