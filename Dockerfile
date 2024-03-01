FROM --platform=linux/amd64 golang:1.21

RUN groupadd -g 30000 shigde && useradd -r -u 30000 -g shigde shigde
RUN mkdir -p /etc/shigde
RUN mkdir -p /var/lib/shigde
COPY ./config.docker.toml /etc/shigde/config.toml
RUN chmod -R 0755 /etc/shigde/

RUN touch /var/log/shigde.log
RUN chown shigde:shigde /var/log/shigde.log
RUN chown shigde:shigde /var/lib/shigde

USER shigde

COPY ./bin/shig.linux.amd64 /bin/shig.linux.amd64

CMD ["/bin/shig.linux.amd64", "-config=/etc/shigde/config.toml"]
