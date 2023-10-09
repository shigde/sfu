# Shig SFU
Selected Forward Unit for distributed environments

## Documentaion

the documentation you will find [here in the docs folder](docs/README.md)

## Quick Start

(Currently you will need monitoring to start. I will change this coming soon!)
### Start SFU

```shell
make run
```

## Monitoring

### Requirement

please install grafana loki docker plugin

```shell
docker plugin install grafana/loki-docker-driver:latest --alias loki --grant-all-permissions
```

### Start Develop Monitoring 
```shell
make monitor
```

### Run Media Runner

With the Media Runner it is possible to stream static video files in a lobby.

##### Create static media files

Create IVF named output.ivf that contains a VP8/VP9/AV1 track and/or output.ogg that contains a Opus track

```
ffmpeg -i $INPUT_FILE -g 30 -b:v 2M output.ivf
ffmpeg -i $INPUT_FILE -c:a libopus -page_duration 20000 -vn output.ogg
```

Note: In the ffmpeg command which produces the .ivf file, the argument -b:v 2M specifies the video bitrate to be 2 megabits per second. 
We provide this default value to produce decent video quality, but if you experience problems with this configuration (such as dropped frames etc.), you can decrease this. 
See the [ffmpeg documentation](https://ffmpeg.org/ffmpeg.html#Options) for more information on the format of the value.

##### Run the Media Runner
```shell
go build -o ./bin/media_runner ./cmd/media_runner
./bin/media_runner -c config.toml -v output.ivf -a output.ogg
```
##### Run the Media Streamer
```shell
go build -o ./bin/media_streamer ./cmd/media_streamer
./bin/media_streamer -c config.toml
```


```shell
ffmpeg -protocol_whitelist file,udp,rtp -i rtp-forwarder.sdp -c:v libx264 -preset veryfast -b:v 3000k -maxrate 3000k -bufsize 6000k -pix_fmt yuv420p -g 50 -c:a aac -b:a 160k -ac 2 -ar 44100 -f flv rtmp://127.0.0.1:1935/live/15d2f10a-ba68-46c7-8755-52e9320cbd47 
```
