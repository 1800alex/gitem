FROM alpine:3.18

# Install needed packages
RUN set -xe \
	&& apk add --no-cache --purge -uU \
	bash git git-lfs openssh \
	# go
	go \
	# curl
	curl \
	# Clean up
	&& rm -rf /var/cache/apk/*

RUN adduser -D -h /home/test -s /bin/sh test \
	&& mkdir -p /home/test/.ssh \
	&& chown -R test:test /home/test

COPY entrypoint.sh /
RUN chmod u+x /entrypoint.sh

WORKDIR /root

ENTRYPOINT ["/entrypoint.sh"]
