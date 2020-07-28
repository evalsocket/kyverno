#!/bin/sh
set -e

pwd=$(pwd)
hash=a=$(git log -1 --pretty=format:"%H")

cd $pwd/definitions
echo "Installing kustomize"
curl -s "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"  | bash
chmod a+x $pwd/definitions/kustomize
echo "Kustomize image edit"
$pwd/definitions/kustomize edit set image nirmata/kyverno:$hash
$pwd/definitions/kustomize edit set image nirmata/kyvernopre:$hash
$pwd/definitions/kustomize build . > $pwd/definitions/install.yaml