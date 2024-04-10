package sessions

import (
	"github.com/pion/webrtc/v3"
)

/**+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

Shig cannot access a central signaling server, so Shig uses WebRTC data channels to exchange metadata
and signaling. Basically, that's why every WebRTC connection is established with a data channel.
Depending on the situation, the application decides whether the channel is used for signaling.

Furthermore, it should be noted that Shig uses two connections for media exchange.
Always, an egress endpoint connects to an ingress endpoint. An ingress endpoint is always passive.
It will never renegotiate a connection since it is only a receiver. An egress endpoint is always active
and will renegotiate changes. For this, the egress endpoint needs a signal channel.

Additionally, in a user-server connection, the user is always static. The user will never change their
connection status. If a user changes their media, they reconnect. It's different in a server-server connection.
Here, both parties renegotiate their connection.

A) User - Server Scenario:
==========================

                   +------------------------------------------------------------+
                   | The Server uses this (UnidirectionalSignalChannel) channel |
                   | for signaling changes in its egress connection.            |
                   +------------+-----------------------------------------------+
                                |
+----------+                    |            +----------+
|          +---Egress-----------+--Ingress-->+          |
|  User    |                                 |  Server  |
|          +<--Ingress--+-----------Egress---+          |
+----------+            |                    +----------+
                        |
                      +--------------------------------------------------------------+
                      | No one use this (SilentSignalChannel) Channel for Signalling |
                      | The user doesn't need a signaling channel because it didn't  |
                      | change the type of traffic it's sending.                     |
                      +--------------------------------------------------------------+

B) Server - Server Scenario:
============================

                                +----------------------------------------------------------------------+
                                | Server A&B use a (BidirectionalSignalChannel) Channel for Signalling |
                                +----------------------------------------------------------------------+
                                |
+----------+                    |            +----------+
|          +---Egress-----------+--Ingress-->+          |
| Server A |                                 | Server B |
|          +<--Ingress--+-----------Egress---+          |
+----------+            |                    +----------+
                        |
                      +-----------------------------------------------------------------+
                      | No Server use this (SilentSignalChannel) Channel for Signalling |
                      +-----------------------------------------------------------------+
**********************************************************************/

type SignalChannelKind int

const (
	SilentSignalChannel                  SignalChannelKind = iota + 1
	UnidirectionalReceivingSignalChannel                   // This type not exist on sever side
	UnidirectionalSendingSignalChannel
	BidirectionalSignalChannel
)

func buildSignalDataChannelCbk(s *Session, kind SignalChannelKind) func(*webrtc.DataChannel) {
	if kind == SilentSignalChannel {
		return s.signal.OnSilentChannel
	}
	return s.signal.OnSenderChannel
}
