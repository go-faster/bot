#!/bin/sh

gpg --quiet --batch --yes --decrypt --passphrase="${KUBECONFIG_GPG_PASS}" \
--output /tmp/kubeconfig ./_deploy/kubeconfig.gpg
