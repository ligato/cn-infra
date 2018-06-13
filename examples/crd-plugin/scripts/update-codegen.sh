#!/bin/bash

# The only argument this script should ever be called with is '--verify-only'

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname ${BASH_SOURCE})/..
CODEGEN_PKG=${CODEGEN_PKG:-$(cd ${SCRIPT_ROOT}; ls -d -1 ./vendor/k8s.io/code-generator 2>/dev/null || echo ../code-generator)}

echo "Calling ${CODEGEN_PKG}/generate-groups.sh"
${CODEGEN_PKG}/generate-groups.sh all \
  github.com/ligato/cn-infra/examples/crd-plugin/pkg/client github.com/ligato/cn-infra/examples/crd-plugin/pkg/apis \
  crdexample.io:v1 \
  --output-base "${GOPATH}/src/" \
  --go-header-file ${SCRIPT_ROOT}/conf/boilerplate.txt

echo "Generating other deepcopy funcs"
${GOPATH}/bin/deepcopy-gen \
  --input-dirs ./crdexample \
  --go-header-file ${SCRIPT_ROOT}/conf/boilerplate.txt \
  --bounding-dirs ./api \
  -O zz_generated.deepcopy \
  -o "${GOPATH}/src"

