#!/bin/sh

echo "${KUBECONFIG}" | base64 -d > /tmp/kubeconfig
