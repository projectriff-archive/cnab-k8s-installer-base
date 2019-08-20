#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
CODEGEN_PKG=${CODEGEN_PKG:-$(cd "${SCRIPT_ROOT}"; ls -d -1 ./vendor/k8s.io/code-generator 2>/dev/null || echo ../code-generator)}


"${CODEGEN_PKG}"/generate-groups.sh "deepcopy,client" \
  github.com/projectriff/cnab-k8s-installer-base/pkg/client github.com/projectriff/cnab-k8s-installer-base/pkg/apis \
  kab:v1alpha1 \
  --go-header-file "${SCRIPT_ROOT}"/hack/copyright.go.txt

