FROM alpine:3.18

RUN apk add --no-cache \
	openjdk11-jre-headless \
	bash \
	git \
	su-exec \
	tini \
	&& rm -rf /var/cache/apk/*

RUN addgroup -g 1000 gitbucket \
	&& adduser -u 1000 -G gitbucket -s /bin/bash -D gitbucket

ENV GITBUCKET_VERSION=4.40.0 \
	GITBUCKET_HOME=/opt/gitbucket \
	GITBUCKET_PORT=8080 \
	GITBUCKET_LOG=/var/log/gitbucket \
	GITBUCKET_UID=1000 \
	GITBUCKET_GID=1000 \
	GITBUCKET_USER=gitbucket \
	GITBUCKET_GROUP=gitbucket \
	GITBUCKET_JAVA_OPTS="-Xmx1g -Xms1g -XX:+UseG1GC -XX:+UseStringDeduplication -XX:+UseCompressedOops -XX:+HeapDumpOnOutOfMemoryError -XX:HeapDumpPath=/opt/gitbucket/heapdump.hprof"

RUN mkdir -p ${GITBUCKET_HOME} \
	&& mkdir -p ${GITBUCKET_LOG} \
	&& chown -R ${GITBUCKET_UID}:${GITBUCKET_GID} ${GITBUCKET_HOME} \
	&& chown -R ${GITBUCKET_UID}:${GITBUCKET_GID} ${GITBUCKET_LOG}

RUN wget -q -O ${GITBUCKET_HOME}/gitbucket.war https://github.com/gitbucket/gitbucket/releases/download/${GITBUCKET_VERSION}/gitbucket.war

COPY entrypoint.sh /usr/local/bin/docker-entrypoint.sh
RUN chmod +x /usr/local/bin/docker-entrypoint.sh

ENTRYPOINT [ "/sbin/tini", "--", "/usr/local/bin/docker-entrypoint.sh" ]