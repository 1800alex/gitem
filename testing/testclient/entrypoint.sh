#!/usr/bin/env bash

TESTCLIENT_SSH_PRIVATE_KEY=/keys/ssh_host_rsa_key
TESTCLIENT_SSH_PUBLIC_KEY=/keys/ssh_host_rsa_key.pub

function url_encode() {
	echo "$@" \
	| sed \
		-e 's/%/%25/g' \
		-e 's/ /%20/g' \
		-e 's/!/%21/g' \
		-e 's/"/%22/g' \
		-e "s/'/%27/g" \
		-e 's/#/%23/g' \
		-e 's/(/%28/g' \
		-e 's/)/%29/g' \
		-e 's/+/%2b/g' \
		-e 's/,/%2c/g' \
		-e 's/-/%2d/g' \
		-e 's/:/%3a/g' \
		-e 's/;/%3b/g' \
		-e 's/?/%3f/g' \
		-e 's/@/%40/g' \
		-e 's/\$/%24/g' \
		-e 's/\&/%26/g' \
		-e 's/\*/%2a/g' \
		-e 's/\./%2e/g' \
		-e 's/\//%2f/g' \
		-e 's/\[/%5b/g' \
		-e 's/\\/%5c/g' \
		-e 's/\]/%5d/g' \
		-e 's/\^/%5e/g' \
		-e 's/_/%5f/g' \
		-e 's/`/%60/g' \
		-e 's/{/%7b/g' \
		-e 's/|/%7c/g' \
		-e 's/}/%7d/g' \
		-e 's/~/%7e/g'
}

function gitbucket_ready() {
	while ! curl "http://gitbucket/" -sS -o /dev/null; do
		echo "Waiting for gitbucket to be ready..."
		sleep 1
	done
}

function gitbucket_login() {
	local sessid=$(curl "http://gitbucket/" -sS -D - -o /dev/null | grep -m1 JSESSIONID= | cut -d= -f2 | cut -d ';' -f1)
	curl "http://gitbucket/signin" -X POST -H "Content-Type: application/x-www-form-urlencoded" -H "Cookie: JSESSIONID=${sessid}" -H "Upgrade-Insecure-Requests: 1" --data-raw "userName=$1&password=$2&hash=" &>/dev/null
	echo "$sessid"
}

function gitbucket_post() {
	local sessid="$1"
	local url="$2"
	local data="$3"
	curl "$url" -X POST -H "Content-Type: application/x-www-form-urlencoded" -H "Cookie: JSESSIONID=${sessid}" --data-raw "$data" 2>/dev/null
}

function gitbucket_get() {
	local sessid="$1"
	local url="$2"
	curl "$url" -H "Content-Type: application/x-www-form-urlencoded" -H "Cookie: JSESSIONID=${sessid}" 2>/dev/null
}

function gitbucket_new_repo() {
	local url="http://gitbucket/new"
	local sessid="$1"
	local owner="$2"
	local name="$3"
	local description="$(url_encode "$4")"
	local isPrivate="${5:-false}"
	curl "$url" -X POST -H "Content-Type: application/x-www-form-urlencoded" -H "Cookie: JSESSIONID=${sessid}" --data-raw "owner=${owner}&name=${name}&description=${description}&isPrivate=${isPrivate}&initOption=EMPTY" 2>/dev/null
}

function softserve_new_repo() {
	local name="$1"
	local description="$2"
	ssh softserve repo create ${name} '-d '"${description}"''
}

function softserve_ready() {
	while ! ssh softserve info > /dev/null; do
		echo "Waiting for softserve to be ready..."
		sleep 1
	done
}

if [ ! -z "$TESTCLIENT_SSH_PRIVATE_KEY" ]; then
	mkdir -p ${HOME}/.ssh
	chmod 700 ${HOME}/.ssh
	cat "$TESTCLIENT_SSH_PRIVATE_KEY" > ${HOME}/.ssh/id_rsa
	cat "$TESTCLIENT_SSH_PUBLIC_KEY" > ${HOME}/.ssh/id_rsa.pub
	chmod 600 ${HOME}/.ssh/id_rsa
	chmod 600 ${HOME}/.ssh/id_rsa.pub

	echo "" > ${HOME}/.ssh/config
	echo "Host cgit" >> ${HOME}/.ssh/config
	echo "  User git" >> ${HOME}/.ssh/config
	echo "  IdentityFile ${HOME}/.ssh/id_rsa" >> ${HOME}/.ssh/config
	echo "  StrictHostKeyChecking no" >> ${HOME}/.ssh/config
	echo "" >> ${HOME}/.ssh/config

	echo "Host gitbucket" >> ${HOME}/.ssh/config
	echo "  User git" >> ${HOME}/.ssh/config
	echo "  IdentityFile ${HOME}/.ssh/id_rsa" >> ${HOME}/.ssh/config
	echo "  StrictHostKeyChecking no" >> ${HOME}/.ssh/config
	echo "" >> ${HOME}/.ssh/config

	echo "Host softserve" >> ${HOME}/.ssh/config
	echo "  User git" >> ${HOME}/.ssh/config
	echo "  IdentityFile ${HOME}/.ssh/id_rsa" >> ${HOME}/.ssh/config
	echo "  StrictHostKeyChecking no" >> ${HOME}/.ssh/config
	echo "" >> ${HOME}/.ssh/config
fi

git config --global user.email "test@example.com"
git config --global user.name "Test User"
git config --global init.defaultBranch master

# Wait for gitbucket to be ready
gitbucket_ready

# attempt a login and store the cookie
root_sessid=$(gitbucket_login root root)

# enable ssh
gitbucket_post "$root_sessid" "http://gitbucket/admin/system" "baseUrl=http%3A%2F%2Fgitbucket&information=&defaultBranch=main&skinName=skin-blue&userDefinedCss=&basicBehavior.allowAccountRegistration=false&basicBehavior.allowResetPassword=false&basicBehavior.repositoryOperation.create=true&basicBehavior.repositoryOperation.delete=true&basicBehavior.repositoryOperation.rename=true&basicBehavior.repositoryOperation.transfer=true&basicBehavior.repositoryOperation.fork=true&basicBehavior.isCreateRepoOptionPublic=true&basicBehavior.allowAnonymousAccess=true&showMailAddress=false&upload.maxFileSize=3145728&upload.timeout=30000&upload.largeMaxFileSize=3145728&upload.largeTimeout=30000&basicBehavior.limitVisibleRepositories=false&repositoryViewer.maxFiles=0&ssh.enabled=on&ssh.bindAddress.host=0.0.0.0&ssh.bindAddress.port=22&ssh.publicAddress.host=22&ssh.publicAddress.port=22"
# curl "http://gitbucket/admin/system" -X POST -H "Content-Type: application/x-www-form-urlencoded" -H "Cookie: JSESSIONID=${root_sessid}" --data-raw "baseUrl=http%3A%2F%2Fgitbucket&information=&defaultBranch=main&skinName=skin-blue&userDefinedCss=&basicBehavior.allowAccountRegistration=false&basicBehavior.allowResetPassword=false&basicBehavior.repositoryOperation.create=true&basicBehavior.repositoryOperation.delete=true&basicBehavior.repositoryOperation.rename=true&basicBehavior.repositoryOperation.transfer=true&basicBehavior.repositoryOperation.fork=true&basicBehavior.isCreateRepoOptionPublic=true&basicBehavior.allowAnonymousAccess=true&showMailAddress=false&upload.maxFileSize=3145728&upload.timeout=30000&upload.largeMaxFileSize=3145728&upload.largeTimeout=30000&basicBehavior.limitVisibleRepositories=false&repositoryViewer.maxFiles=0&ssh.enabled=on&ssh.bindAddress.host=0.0.0.0&ssh.bindAddress.port=22&ssh.publicAddress.host=22&ssh.publicAddress.port=22"

# add test user
gitbucket_post "$root_sessid" "http://gitbucket/admin/users/_newuser" "userName=test&password=password&fullName=test+user&mailAddress=test%40example.com&extraMailAddresses%5B0%5D=&isAdmin=true&url=&description=&fileId="
# curl "http://gitbucket/admin/users/_newuser" -X POST -H "Content-Type: application/x-www-form-urlencoded" -H "Cookie: JSESSIONID=${root_sessid}" --data-raw "userName=test&password=password&fullName=test+user&mailAddress=test%40example.com&extraMailAddresses%5B0%5D=&isAdmin=true&url=&description=&fileId="

# login as test user
test_sessid=$(gitbucket_login test password)

# add ssh key to test user
# read ssh public key from file and url encode it
sshkey=$(cat "$TESTCLIENT_SSH_PUBLIC_KEY")
formdata="title=test&publicKey=$(url_encode "$sshkey")"
gitbucket_post "$test_sessid" "http://gitbucket/test/_ssh" "$formdata"
# curl "http://gitbucket/test/_ssh" -X POST -H "Content-Type: application/x-www-form-urlencoded" -H "Cookie: JSESSIONID=${test_sessid}"  --data-raw "$formdata"


gitbucket_post "$test_sessid" "http://gitbucket/test/_personalToken" "note=test"
# curl "http://gitbucket/test/_personalToken" -X POST -H "Content-Type: application/x-www-form-urlencoded" -H "Cookie: JSESSIONID=${test_sessid}" --data-raw "note=test"
token=$(gitbucket_get "$test_sessid" "http://gitbucket/test/_application" | grep -m1 "data-clipboard-text" | cut -d'"' -f2)
# token=$(curl "http://gitbucket/test/_application" -H "Content-Type: application/x-www-form-urlencoded" -H "Cookie: JSESSIONID=${test_sessid}" | grep -m1 "data-clipboard-text" | cut -d'"' -f2)
echo "Token: $token"

## Setup gitbucket
# Initialize 5 git repositories, each with 10 commits and 10 tags
for i in $(seq 1 5); do
	mkdir -p repositories/repo$i
	(
		gitbucket_new_repo "$test_sessid" "test" "repo$i" "My gitbucket test repo $i"
		cd repositories/repo$i
		git init
		for j in $(seq 1 10); do
			printf "# repo$i\n\nCommit: $j" > README.md
			git add README.md
			git commit -m "repo$i commit $j"
			git tag "v0.0.$j"
		done

		git remote add origin git@gitbucket:/test/repo$i.git

		while ! git push -u origin master; do
			echo "git push failed, retrying..."
			sleep 1
		done

		git push --tags
	)

	# cleanup
	rm -rf repositories/repo$i
done

# Initialize 5 git repositories, each with an LFS binary file
for i in $(seq 6 10); do
	mkdir -p repositories/repo$i
	(
		gitbucket_new_repo "$test_sessid" "test" "repo$i" "My gitbucket test repo $i"
		cd repositories/repo$i
		git init
		printf "# repo$i\n\nCommit: $j" > README.md
		git add README.md
		git commit -m "repo$i commit 1"
		git remote add origin git@gitbucket:/test/repo$i.git
		while ! git push -u origin master; do
			echo "git push failed, retrying..."
			sleep 1
		done

		git config -f .lfsconfig lfs.url http://test:password@gitbucket/test/repo$i.git/info/lfs
		git add .lfsconfig
		git lfs install
		git lfs track "*.bin"
		git add .gitattributes
		dd if=/dev/urandom of=file.bin bs=512 count=1
		md5sum file.bin > file.bin.md5
		git add file.bin file.bin.md5
		git commit -m "repo$i lfs commit 2"
		git push
		git tag "v0.0.1"
		git push --tags
	)

	# cleanup
	rm -rf repositories/repo$i
done

## Setup softserve

# Wait for softserve to be ready
softserve_ready

# Initialize 5 git repositories, each with 10 commits and 10 tags
for i in $(seq 1 5); do
	mkdir -p repositories/icecream$i
	(
		softserve_new_repo "icecream$i" "My softserve icecream $i"
		cd repositories/icecream$i
		git init
		for j in $(seq 1 10); do
			printf "# icecream$i\n\nCommit: $j" > README.md
			git add README.md
			git commit -m "icecream$i commit $j"
			git tag "v0.0.$j"
		done

		git remote add origin git@softserve:/icecream$i.git

		while ! git push -u origin master; do
			echo "git push failed, retrying..."
			sleep 1
		done

		git push --tags
	)

	# cleanup
	rm -rf repositories/icecream$i
done

# Initialize 5 git repositories, each with an LFS binary file
for i in $(seq 6 10); do
	mkdir -p repositories/icecream$i
	(
		softserve_new_repo "icecream$i" "My softserve icecream $i"
		cd repositories/icecream$i
		git init
		printf "# icecream$i\n\nCommit: $j" > README.md
		git add README.md
		git commit -m "icecream$i commit 1"
		git remote add origin git@softserve:/icecream$i.git
		while ! git push -u origin master; do
			echo "git push failed, retrying..."
			sleep 1
		done

		# git config -f .lfsconfig lfs.url http://test:password@gitbucket/test/repo$i.git/info/lfs
		# git add .lfsconfig
		git lfs install
		git lfs track "*.bin"
		git add .gitattributes
		dd if=/dev/urandom of=file.bin bs=512 count=1
		md5sum file.bin > file.bin.md5
		git add file.bin file.bin.md5
		git commit -m "icecream$i lfs commit 2"
		git push
		git tag "v0.0.1"
		git push --tags
	)

	# cleanup
	rm -rf repositories/icecream$i
done

#############################################

use_cgit=0
use_cgit_lfs=0 # I don't understand how to make this work

if [ $use_cgit -eq 1 ]; then
	# Setup gitolite and cgit
	# Wait for cgit to be ready (check that ssh can connect)
	while ! ssh -o StrictHostKeyChecking=no git@cgit; do
		echo "Waiting for cgit to be ready..."
		sleep 1
	done

	mkdir -p repositories
	(
		cd repositories
		git clone git@cgit:/gitolite-admin.git
		cd gitolite-admin

		# see https://gitolite.com/gitolite/cookbook.html
		cat <<EOF > conf/gitolite.conf
	@devs = test

	repo gitolite-admin
		RW+ = test
EOF
		git add conf/gitolite.conf
		git commit -m "Add gitolite-admin"
		git push
	)

	# wait for lfs server to respond to http://cgit:8080/mgmt/users with basic auth user:pass
	while ! curl -u admin:admin http://cgit:8080/mgmt/users; do
		echo "Waiting for lfs server to be ready..."
		sleep 1
	done

	# add lfs user using POST http://cgit:8080/mgmt/add with basic auth user:pass and data of name=lfs&password=pass
	curl -X POST -u admin:admin -d "name=lfs&password=pass" http://cgit:8080/mgmt/add

	# Initialize 5 git repositories, each with 10 commits and 10 tags
	for i in $(seq 1 5); do
		# add repo$i to gitolite
		(
			cd repositories/gitolite-admin
			cat <<EOF >> conf/gitolite.conf
	repo repo$i
		RW+ = @devs
EOF
			git add conf/gitolite.conf
			git commit -m "Add repo$i"
			git push
		)

		mkdir -p repositories/repo$i
		(
			cd repositories/repo$i
			git init
			for j in $(seq 1 10); do
				printf "# repo$i\n\nCommit: $j" > README.md
				git add README.md
				git commit -m "repo$i commit $j"
				git tag "v0.0.$j"
			done

			git remote add origin git@cgit:/repo$i.git
			git push -u origin master
			git push --tags
		)

		# cleanup
		rm -rf repositories/repo$i
	done

	if [ $use_cgit_lfs -eq 1 ]; then
		# Initialize 5 git repositories, each with an LFS binary file
		for i in $(seq 6 10); do
			# add repo$i to gitolite
			(
				cd repositories/gitolite-admin
				cat <<EOF >> conf/gitolite.conf
		repo repo$i
			RW+ = @devs
EOF
				git add conf/gitolite.conf
				git commit -m "Add repo$i"
				git push
			)

			mkdir -p repositories/repo$i
			(
				cd repositories/repo$i
				git init
				git config -f .lfsconfig lfs.url http://lfs:pass@cgit:8080/test/repo$i.git

				printf "# repo$i\n\nLFS file" > README.md
				git add .lfsconfig README.md
				git commit -m "repo$i commit 1"
				git tag "v0.0.1"

				git remote add origin git@cgit:/repo$i.git
				git lfs install
				git lfs track "*.bin"
				git add .gitattributes
				dd if=/dev/urandom of=file.bin bs=512 count=1
				md5sum file.bin > file.bin.md5
				git add file.bin file.bin.md5
				git commit -m "repo$i lfs commit 2"
				git push -u origin master
				git push --tags
			)

			# cleanup
			rm -rf repositories/repo$i
		done
	fi
fi

# TODO
sleep infinity

# 		# prepare bare repo for cgit (only needed if not using gitolite)
# 		cd ..
# 		git clone --bare repo$i repo$i.git
# 		echo "My git/cgit test repo $1" > repo$i.git/description
# 		cat <<EOF >> repo$i.git/cgitrc
# section=Developer test
# owner=Test User
# description=My git/cgit test repo $i
# readme=README.md
# EOF
# 		# copy bare repo to cgit
# 		scp -r repo$i.git cgit:/

# 		# cleanup
# 		rm -rf repo$i.git