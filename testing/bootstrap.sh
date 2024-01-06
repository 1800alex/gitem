#!/bin/bash

# generate ssh key and store it in keys/ssh_host_rsa_key
mkdir -p ./multi/keys
rm -f ./multi/keys/ssh_host_rsa_key
rm -f ./multi/keys/ssh_host_rsa_key.pub
ssh-keygen -q -N "" -t rsa -b 4096 -f ./multi/keys/ssh_host_rsa_key

# replace the last word in the public key with a fake test@test-client
sed -i 's/[^ ]*$/test@test-client/' ./multi/keys/ssh_host_rsa_key.pub

SSH_PRIVATEKEY=$(cat ./multi/keys/ssh_host_rsa_key)
SSH_KEY=$(cat ./multi/keys/ssh_host_rsa_key.pub)
SSH_KEY_NAME=test

cp multi/config.env.example multi/config.env
sed -i \
    -e "s#CGIT_SSH_KEY=.*#CGIT_SSH_KEY=${SSH_KEY}#g" \
    -e "s#CGIT_SSH_KEY_NAME=.*#CGIT_SSH_KEY_NAME=${SSH_KEY_NAME}#g" \
    multi/config.env

mkdir -p multi/gitea/conf
cp multi/app.ini multi/gitea/conf/app.ini