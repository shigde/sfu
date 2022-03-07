FROM golang:stretch

WORKDIR $GOPATH/src/github.com/shigde/sfu

COPY go.mod go.sum ./
RUN cd $GOPATH/src/github.com/shigde/sfu && go mod download

COPY pkg/ $GOPATH/src/github.com/shigde/sfu/pkg
COPY cmd/ $GOPATH/src/github.com/shigde/sfu/cmd

WORKDIR $GOPATH/src/github.com/shigde/sfu/cmd
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /sfu .

FROM alpine:3.12.3

RUN apk --no-cache add ca-certificates
COPY --from=0 /sfu /usr/local/bin/sfu

COPY config.toml /configs/sfu.toml

ENTRYPOINT ["/usr/local/bin/sfu"]
CMD ["-c", "/configs/sfu.toml"]
