
#!/bin/sh
set -e

pwd=$(pwd)
hash=$(git describe --always --tags)
#
## Install Kind
curl -Lo $pwd/kind https://kind.sigs.k8s.io/dl/v0.8.1/kind-linux-amd64
chmod a+x $pwd/kind

# ## Create Kind Cluster
$pwd/kind create cluster
kind load docker-image ghcr.io/kyverno/kyverno:$hash
kind load docker-image ghcr.io/kyverno/kyvernopre:$hash

cd $pwd/definitions
echo "Installing kustomize"
<<<<<<< HEAD
# curl -s "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"  | bash
echo "Kustomize image edit"
kustomize edit set image ghcr.io/kyverno/kyverno:$hash
kustomize edit set image ghcr.io/kyverno/kyvernopre:$hash
kustomize build $pwd/definitions/ -o $pwd/definitions/install.yaml
=======
curl -s "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"  | bash
kustomize edit set image ghcr.io/kyverno/kyverno:$hash
kustomize edit set image ghcr.io/kyverno/kyvernopre:$hash
kustomize build $pwd/definitions/ > $pwd/definitions/install.yaml
>>>>>>> 2ffe9b02... Added kustomize install script (#1392)
