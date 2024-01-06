#!/bin/bash

# generate ssh key and store it in keys/ssh_host_rsa_key
mkdir -p keys
rm -f ./keys/ssh_host_rsa_key
rm -f ./keys/ssh_host_rsa_key.pub
ssh-keygen -q -N "" -t rsa -b 4096 -f ./keys/ssh_host_rsa_key

# replace the last word in the public key with a fake test@test-client
sed -i 's/[^ ]*$/test@test-client/' ./keys/ssh_host_rsa_key.pub

SSH_KEY=$(cat ./keys/ssh_host_rsa_key.pub)
SSH_KEY_NAME=test

cp multi/config.env.example multi/config.env
sed -i \
    -e "s#CGIT_SSH_KEY=.*#CGIT_SSH_KEY=${SSH_KEY}#g" \
    -e "s#CGIT_SSH_KEY_NAME=.*#CGIT_SSH_KEY_NAME=${SSH_KEY_NAME}#g" \
    multi/config.env
