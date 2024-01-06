#!/bin/sh
# vim: nowrap ft=sh

gitbucket_home=${GITBUCKET_HOME:-/opt/gitbucket}
gitbucket_port=${GITBUCKET_PORT:-8080}

git config --global user.email "test@example.com"
git config --global user.name "Test User"

# Initialize 5 git repositories, each with 10 commits and 10 tags
for i in $(seq 1 5); do
	mkdir -p ${gitbucket_home}/repos/repo$i
	cd ${gitbucket_home}/repos/repo$i
	git init
	for j in $(seq 1 10); do
		printf "# repo$i\n\nCommit: $j" > README.md
		git add README.md
		git commit -m "repo$i commit $j"
	done
	for j in $(seq 1 10); do
		git tag "v0.0.$j"
	done
done


java -jar /opt/gitbucket/gitbucket.war --gitbucket.home=${gitbucket_home} --gitbucket.port=${gitbucket_port}
