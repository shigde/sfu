# Backpressure in federative live streaming
> Resistance or force opposing the desired flow of data through software.


## Situation

Client --> ShigA --> ShigB --> Streamer

## Congestion

### Receiver Estimated Max Bitrate
We have three Estimation in a pipe
 - Client -> ShigA
 - ShigA -> ShigB
 - ShigB -> Streamer

Questions:
- Would we calculate and accumulate all three to One?
- Is Simulcast enough? When ShigB has small bitrate we would call sende don't send max track size?

