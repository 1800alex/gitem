FROM alpine:3.18

# Install needed packages
RUN set -xe \
	&& apk add --no-cache --purge -uU \
	bash git git-lfs openssh \
	# go
	go \
	# curl
	curl \
	# yq
	yq \
	# taskfile
	go-task-task \
	# make
	make \
	# Clean up
	&& rm -rf /var/cache/apk/*

# Add test user with UID 1000 GID 1000
RUN adduser -D -u 1000 -g 1000 test \
	&& mkdir -p /home/test \
	&& chown -R test:test /home/test

COPY entrypoint.sh /
RUN chmod u+x /entrypoint.sh

WORKDIR /home/test
USER test

ENTRYPOINT ["/entrypoint.sh"]
