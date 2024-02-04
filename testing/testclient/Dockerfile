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
	# make
	make \
	# Clean up
	&& rm -rf /var/cache/apk/*

# Download and install taskfile version 3.34.1
RUN set -xe \
	&& curl -L https://github.com/go-task/task/releases/download/v3.34.1/task_linux_amd64.tar.gz -o /tmp/task.tar.gz \
	&& tar -xzf /tmp/task.tar.gz -C /tmp \
	&& mv /tmp/task /usr/local/bin/task \
	&& rm -rf /tmp/task.tar.gz /tmp/task && \
	chmod +x /usr/local/bin/task && \
	task --version

# Add test user with UID 1000 GID 1000
RUN adduser -D -u 1000 -g 1000 test \
	&& mkdir -p /home/test \
	&& chown -R test:test /home/test

COPY entrypoint.sh /
RUN chmod u+x /entrypoint.sh

WORKDIR /home/test
USER test

ENTRYPOINT ["/entrypoint.sh"]
