FROM alpine:3.3

RUN apk add -U \
	bash \
	nginx \
	&& rm -rf /var/cache/apk*

# forward request and error logs to docker log collector
RUN ln -sf /dev/stdout /var/log/nginx/access.log
RUN ln -sf /dev/stderr /var/log/nginx/error.log

ENV POWERED_BY Prow

COPY rootfs /

CMD ["/bin/boot"]
EXPOSE 80
