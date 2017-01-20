FROM alpine:3.3

COPY rootfs /

EXPOSE 44135

ENTRYPOINT ["/bin/prowd"]
