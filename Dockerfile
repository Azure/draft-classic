FROM alpine:3.5

COPY rootfs /

EXPOSE 44135

ENTRYPOINT ["/bin/prowd"]
CMD ["start"]
