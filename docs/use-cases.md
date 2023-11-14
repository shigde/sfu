# Use Cases

## Embedded Streaming
Shig enables livestreaming directly from PeerTube without the need to configure OBS or transfer streaming secret keys to another software.

There is a PeerTube plugin for this purpose.
With the plugin, a Shig service can be registered, allowing Shig to call the Instance Actor and follow the instance.
Since both PeerTube and Shig have ActivityPub integration, Shig can retrieve information about the peerTube instance and  live videos.
This enables embedded streaming, meaning no other software is necessary.

## Multi Guest Streaming

With Shig you can invite other PeerTube users to join the stream.
I did a video to demonstrate this: https://video.shig.de/w/j8ZuLoCuJEkReuUWoy9g5d
 
## Scaling across federative instances

One of the fundamental ideas of the Fediverse is trust. That´s mean you select a PeerTube provider, and the content is delivered to you in a way that keeps your private data exclusively with that provider. 
With ActivityPub, it's possible to copy video files from one instance to another. 
However, there is currently no service that allows the copying of live streams (HLS) via ActivityPub from one instance to another.
Shig is attempting to do just that. Multi-guest streaming or direct streaming from PeerTube is only a byproduct on the way to solving this problem.


## Monetization with shared video pools in Live Streams

The network Fediverse of independent instances that exchange videos, texts, and messages among each other. 
Each individual instance is managed by an operator who has a fundamental interest in monetizing content. 
However, the decision-making and ownership structures of the involved groups in the Fediverse are somewhat more complex as in central systems:

- Instance operators determine which monetization service can deliver advertising content.
- Video creators determine how and how much advertising content is displayed.
- Video consumers decide which content they can access even when advertising is displayed.
- Advertisers decide where their ads should be delivered.
- The monetization service decides which instances to add to its delivery network.

It should be assumed that of the five different groups, only consumers have no interest in monetization. 
Therefore, it should be considered how much advertising disrupts the consumption of videos and live streams:

- The advertising video plays in a popup in the corner and is dismissible.
- The advertising video plays embedded in a corner of the main video and is not dismissible.
- The advertising displaces the video.
