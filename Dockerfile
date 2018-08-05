FROM alpine:3.8
ARG TAG=v1.1.14
ADD https://github.com/accept-nano/accept-nano/releases/download/$TAG/accept-nano /usr/bin/accept-nano
RUN ["chmod", "+x", "/usr/bin/accept-nano"]
RUN ["touch", "/etc/accept-nano.toml"]
ENTRYPOINT ["/usr/bin/accept-nano", "-config", "/etc/accept-nano.toml"]
EXPOSE 5000
