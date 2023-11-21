# Command Line Tool (ShigCLT)

## Build the Command Line Tool

```shell
make build-clt
```

## Send Video and AUdio in Lobby with CLT

```shell
bin/shigClt -c .shigClt.toml send --video input.ivf --audio input.ogg --main --url https://stream.shig.de/space/root_channel@video.shig.de/stream/e86b352e-d276-4638-9d5f-6b3ff03b49fc
```
Note: If you set the ```--main``` flag, the stream will be declared as the main stream and streamed directly from the lobby to the publisher. 
Please also consider the settings in the ```.shigClt.toml``` file for this purpose.
Without the flag, the stream will only be streamed to the lobby.

## Config 
In the config, the stream endpoint streaming user id must be specified.

```toml
[shig]
user = "streamer@video.shig.de"
pass = "currently ignorered, because register token, but cumming soon"
registerToken = "this-token-must-be-changed-in-public"
```

Additionally, when live-streaming, the publisher's RTMP endpoint needs to be provided. 
```toml
[rtmp]
streamKey = "c4855252-00e1-43ef-baf8-230ace3249e9"
rtmpUrl = "rtmp://video.shig.de/live"
```

Optionally, STUN/TURN servers can be set.
```toml
[rtp]
iceServer = [{ urls = ["stun:stun.l.google.com:19302"] }]
```

##### Create static media files
Create IVF named output.ivf that contains a VP8/VP9/AV1 track and/or output.ogg that contains a Opus track

```
ffmpeg -i $INPUT_FILE -g 30 -b:v 2M output.ivf
ffmpeg -i $INPUT_FILE -c:a libopus -page_duration 20000 -vn output.ogg
```

Note: In the ffmpeg command which produces the .ivf file, the argument -b:v 2M specifies the video bitrate to be 2 megabits per second.
We provide this default value to produce decent video quality, but if you experience problems with this configuration (such as dropped frames etc.), you can decrease this.
See the [ffmpeg documentation](https://ffmpeg.org/ffmpeg.html#Options) for more information on the format of the value.
