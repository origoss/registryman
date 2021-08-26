FROM alpine:3.14
WORKDIR /usr/lib

COPY skopeo .
RUN apk update && \
    apk upgrade && \
    apk add skopeo

ENTRYPOINT [ "/usr/bin/skopeo" ]
