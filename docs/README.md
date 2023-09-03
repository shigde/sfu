# Documentation

## Lobby Workflow

### Create a Lobby Session and send Video/Audio
To create a Lobby Session you have to send a Http (Post) request with an Authentication Token inside the request Header.
Within the Lobby Session you will create a Media Endpoint for publishing you own streaming audio and video content.
You have to send in the payload a Session Description Offer with Media Metadata you need to create a video or audio stream.

!["live-stream"](./uml/sequence/lobby-create-whip.png)

As Response the Server will respond with an SDP Answer.
The Server will send a session cookie which represent the Media Resource

### Delete a Lobby
You can not update your Lobby Session, because an Lobby Session is stateful unique. 
In case you want change you sending devices you have to delete your first session and create a new one.
That's a trade-off we accept to avoid complex and not needed signaling logic.

!["live-stream"](./uml/sequence/lobby-delete-whip.png)


### Listen a Lobby Session and receive Video/Audio 


## Live Stream vs. ActivePub Video

The space gives access to the lobby

!["live-stream"](./uml/class/lobby.class.png)

