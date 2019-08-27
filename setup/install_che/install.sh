#*******************************************************************************
# Copyright (c) 2019 IBM Corporation and others.
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Eclipse Public License v2.0
# which accompanies this distribution, and is available at
# http://www.eclipse.org/legal/epl-v20.html
#
# Contributors:
#     IBM Corporation - initial API and implementation
#*******************************************************************************

#!/bin/bash

BASE_DIR=$(cd "$(dirname "$0")"; pwd)
INSTALL_MODE=helm
CHE_NAMESPACE=che

function usage {
  	me=$(basename "$0")
  	cat <<EOF
Usage: ${me} [-<option letter> <option value>] [-h]
Options:
  -m  # Eclipse Che installation mode. Can be one of: helm, os, operator (defaults to ${INSTALL_MODE})
  -i  # Ingress domain, of the form <public-ip>.nip.io. Optional if installing on OKD or OpenShift.
  -n  # Namespace to install Che into (defaults to ${CHE_NAMESPACE})
EOF
	exit 0
}

while getopts "m:i:n:h" OPTION; do
    case "$OPTION" in
        m) INSTALL_MODE=$OPTARG ;;
        i) INGRESS_DOMAIN=$OPTARG ;;
        n) CHE_NAMESPACE=$OPTARG ;;
        *) usage ;;
    esac
done

# Verify there is an active Kubernetes context and can list pods in the current namespace
kubectl get pods > /dev/null 2>&1
if [[ $? != 0 ]]; then
    echo -e "Error: Unable to list pods in the current namespace. Please ensure there is an active Kubernetes context."
fi

# Create the cluster role needed for Codewind
kubectl apply -f ${BASE_DIR}/codewind-clusterrole.yaml
kubectl apply -f ${BASE_DIR}/codewind-tektonrole.yaml

if [[ "$INSTALL_MODE" == "operator" ]]; then
    # Install Che via the operator
    # Verify the operator is installed
    kubectl get deploy che-operator > /dev/null 2>&1
    if [[ $? != 0 ]]; then
        echo -e "Error: Please install the Che operator before proceeding."
        exit 1
    fi

    # Check if the ingress domain was set.
    if grep -q "ingressDomain: \'\'" ${BASE_DIR}/codewind-checluster.yaml; then
        echo -e "Error: The ingress domain needs to be set, using the form <ip>.nip.io"
        exit 1
    fi

    # Deploy the Codewind CheCluster
    kubectl create -f ${BASE_DIR}/operator/codewind-checluster.yaml
elif [[ "$INSTALL_MODE" == "os" ]]; then
    # Allow containers in the Che namespace to run as privileged and root
    echo "Setting privileged and anyuid SCCs for eclipse-che namespace"
    oc adm policy add-scc-to-group privileged system:serviceaccounts:eclipse-che
    oc adm policy add-scc-to-group anyuid system:serviceaccounts:eclipse-che

    # Install Che using the openshift deployment scripts
    git clone -b 7.0.x https://github.com/eclipse/che.git
    cd che/deploy/openshift
    
    # Deploy on OpenShift
    ./deploy_che.sh --image-che=eclipse/che-server:7.0.0

    # Create the role binding
    kubectl apply -f ${BASE_DIR}/codewind-rolebinding.yaml -n eclipse-che
    kubectl apply -f ${BASE_DIR}/codewind-tektonbinding.yaml -n eclipse-che
else
    # Deploy using the Helm chart
    if [[ -z "$INGRESS_DOMAIN" ]]; then
        echo -e "Error: The ingress domain must be set when using Helm to install Che. "
        usage
    fi

    # Clone the Che repositor, as that's where the Che helm chart resides
    git clone -b 7.0.x https://github.com/eclipse/che.git
    cd che/deploy/kubernetes/helm/che

    # Install Helm dependencies
    helm dependency update

    # Install Che helm chart
    helm upgrade --install che --namespace $CHE_NAMESPACE \
        --set cheImage=eclipse/che-server:7.0.0 \
        --set global.ingressDomain=$INGRESS_DOMAIN \
        --set global.cheWorkspacesNamespace=$CHE_NAMESPACE \
        --set global.cheWorkspaceClusterRole=eclipse-codewind \
        --set che.workspace.devfileRegistryUrl="https://che-devfile-registry.openshift.io/" \
        --set che.workspace.pluginRegistryUrl="https://che-plugin-registry.openshift.io/v3" ./
    
    kubectl apply -f ${BASE_DIR}/codewind-rolebinding.yaml -n $CHE_NAMESPACE
    kubectl apply -f ${BASE_DIR}/codewind-tektonbinding.yaml -n $CHE_NAMESPACE
fi
