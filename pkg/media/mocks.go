package media

import (
	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const bearer = "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwiaWF0IjoxNTE2MjM5MDIyLCJ1dWlkIjoiYTY0MzY1ZGItMTc0ZC00ZDExLThjYjEtZWIyYTM2MzlmZmU2In0._xbasA_1ljeszeWdqYqp96EWvJIbCnYOTOFxKgcd7vM"

const spaceId = "abc123"
const resourceID = "152cca71-7156-455b-8b30-a90a4bf8f772"
const rtpSessionId = "dac0039e-947a-4d22-8207-4e81a5cdcf19"

type testStore struct {
	db *gorm.DB
}

func newTestStore() *testStore {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	return &testStore{db}
}

func (s *testStore) GetDatabase() *gorm.DB {
	return s.db
}

type testLobbyManager struct {
}

func newTestLobbyManager() *testLobbyManager {
	return &testLobbyManager{}
}

func (l *testLobbyManager) AccessLobby(_ uuid.UUID, _ uuid.UUID, _ *webrtc.SessionDescription) (struct {
	Answer       *webrtc.SessionDescription
	Resource     uuid.UUID
	RtpSessionId uuid.UUID
}, error) {
	var data struct {
		Answer       *webrtc.SessionDescription
		Resource     uuid.UUID
		RtpSessionId uuid.UUID
	}
	data.Answer = &webrtc.SessionDescription{Type: webrtc.SDPTypeAnswer, SDP: testAnswer}
	data.Resource, _ = uuid.Parse(resourceID)
	data.RtpSessionId, _ = uuid.Parse(rtpSessionId)
	return data, nil
}

const testOffer = "v=0\no=- 5228595038118931041 2 IN IP4 127.0.0.1\ns=-\nt=0 0\na=group:BUNDLE 0 1\na=extmap-allow-mixed\na=msid-semantic: WMS\nm=audio 9 UDP/TLS/RTP/SAVPF 111\nc=IN IP4 0.0.0.0\na=rtcp:9 IN IP4 0.0.0.0\na=ice-ufrag:EsAw\na=ice-pwd:bP+XJMM09aR8AiX1jdukzR6Y\na=ice-options:trickle\na=fingerprint:sha-256 DA:7B:57:DC:28:CE:04:4F:31:79:85:C4:31:67:EB:27:58:29:ED:77:2A:0D:24:AE:ED:AD:30:BC:BD:F1:9C:02\na=setup:actpass\na=mid:0\na=bundle-only\na=extmap:4 urn:ietf:params:rtp-hdrext:sdes:mid\na=sendonly\na=msid:- d46fb922-d52a-4e9c-aa87-444eadc1521b\na=rtcp-mux\na=rtpmap:111 opus/48000/2\na=fmtp:111 minptime=10;useinbandfec=1\nm=video 9 UDP/TLS/RTP/SAVPF 96 97\nc=IN IP4 0.0.0.0\na=rtcp:9 IN IP4 0.0.0.0\na=ice-ufrag:EsAw\na=ice-pwd:bP+XJMM09aR8AiX1jdukzR6Y\na=ice-options:trickle\na=fingerprint:sha-256 DA:7B:57:DC:28:CE:04:4F:31:79:85:C4:31:67:EB:27:58:29:ED:77:2A:0D:24:AE:ED:AD:30:BC:BD:F1:9C:02\na=setup:actpass\na=mid:1\na=bundle-only\na=extmap:4 urn:ietf:params:rtp-hdrext:sdes:mid\na=extmap:10 urn:ietf:params:rtp-hdrext:sdes:rtp-stream-id\na=extmap:11 urn:ietf:params:rtp-hdrext:sdes:repaired-rtp-stream-id\na=sendonly\na=msid:- d46fb922-d52a-4e9c-aa87-444eadc1521b\na=rtcp-mux\na=rtcp-rsize\na=rtpmap:96 VP8/90000\na=rtcp-fb:96 ccm fir\na=rtcp-fb:96 nack\na=rtcp-fb:96 nack pli\na=rtpmap:97 rtx/90000\na=fmtp:97 apt=96"
const testAnswer = "v=0\no=- 1657793490019 1 IN IP4 127.0.0.1\ns=-\nt=0 0\na=group:BUNDLE 0 1\na=extmap-allow-mixed\na=ice-lite\na=msid-semantic: WMS *\nm=audio 9 UDP/TLS/RTP/SAVPF 111\nc=IN IP4 0.0.0.0\na=rtcp:9 IN IP4 0.0.0.0\na=ice-ufrag:38sdf4fdsf54\na=ice-pwd:2e13dde17c1cb009202f627fab90cbec358d766d049c9697\na=fingerprint:sha-256 F7:EB:F3:3E:AC:D2:EA:A7:C1:EC:79:D9:B3:8A:35:DA:70:86:4F:46:D9:2D:CC:D0:BC:81:9F:67:EF:34:2E:BD\na=candidate:1 1 UDP 2130706431 198.51.100.1 39132 typ host\na=setup:passive\na=mid:0\na=bundle-only\na=extmap:4 urn:ietf:params:rtp-hdrext:sdes:mid\na=recvonly\na=rtcp-mux\na=rtcp-rsize\na=rtpmap:111 opus/48000/2\na=fmtp:111 minptime=10;useinbandfec=1\nm=video 9 UDP/TLS/RTP/SAVPF 96 97\nc=IN IP4 0.0.0.0\na=rtcp:9 IN IP4 0.0.0.0\na=ice-ufrag:38sdf4fdsf54\na=ice-pwd:2e13dde17c1cb009202f627fab90cbec358d766d049c9697\na=fingerprint:sha-256 F7:EB:F3:3E:AC:D2:EA:A7:C1:EC:79:D9:B3:8A:35:DA:70:86:4F:46:D9:2D:CC:D0:BC:81:9F:67:EF:34:2E:BD\na=candidate:1 1 UDP 2130706431 198.51.100.1 39132 typ host\na=setup:passive\na=mid:1\na=bundle-only\na=extmap:4 urn:ietf:params:rtp-hdrext:sdes:mid\na=extmap:10 urn:ietf:params:rtp-hdrext:sdes:rtp-stream-id\na=extmap:11 urn:ietf:params:rtp-hdrext:sdes:repaired-rtp-stream-id\na=recvonly\na=rtcp-mux\na=rtcp-rsize\na=rtpmap:96 VP8/90000\na=rtcp-fb:96 ccm fir\na=rtcp-fb:96 nack\na=rtcp-fb:96 nack pli\na=rtpmap:97 rtx/90000\na=fmtp:97 apt=96"
const testAnswerETag = "38ee2e1fc076df403ff93ea9b18f97d8"
