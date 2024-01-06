FROM alpine:3.18

# Install needed packages
RUN set -xe \
	&& apk add --no-cache --purge -uU \
	# Install gitolite
	gitolite git git-lfs openssh \
	# install git-daemon
	git-daemon \
	# add nginx and cgit
	cgit nginx fcgiwrap spawn-fcgi \
	# python 3 pkg for highlight, about page
	py3-markdown py3-pygments \
	# go
	go \
	# Create nginx folder
	&& mkdir -p /run/nginx \
	# Clean up
	&& rm -rf /var/cache/apk/*

RUN go install github.com/github/lfs-test-server@latest

# Volume to store all ssh host key
VOLUME [ "/etc/ssh" ]

# Volume to store Gitolite data, used for Gitolite setup
VOLUME [ "/var/lib/git" ]

EXPOSE 22 80 9418
WORKDIR /var/lib/git

ENV CGIT_LFS_LISTEN="tcp://:8080" \
	CGIT_LFS_HOST="127.0.0.1:8080" \
	CGIT_LFS_CONTENTPATH="/var/lib/git/lfs" \
	CGIT_LFS_ADMINUSER="admin" \
	CGIT_LFS_ADMINPASS="admin"

COPY entrypoint.sh /
RUN chmod u+x /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
