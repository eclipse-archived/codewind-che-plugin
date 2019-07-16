#!/bin/sh

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

set -o pipefail
source /scripts/kube/deployUtils.sh

# Create a folder to render the Codewind template in
mkdir -p $DEPLOY_DIR
cp -rf $SCRIPTS_DIR/kube/codewind_template.yaml $DEPLOY_DIR/codewind.yaml

## touch a deploy messages json file
echo '{}' >$DEPLOY_STATUS_FILE
echo "Created an empty deployment status file."

# Set the PVC and service account to use in Codewind
setCodewindPVC
setCodewindServiceAccount

# Set the subpath for the projects volume mount. Che uses /projects/$CHE_WORKSPACE_ID as the format
setTemplateValue $DEPLOY_DIR/codewind.yaml WORKSPACE_ID_PLACEHOLDER $CHE_WORKSPACE_ID

# Set the namespace to deploy Codewind under. Needs to be the same namespace as the Che workspace it's assocaited with
NAMESPACE=$( kubectl get po --selector=che.original_name=che-workspace-pod -o jsonpath='{.items[0].metadata.namespace}' )
setTemplateValue $DEPLOY_DIR/codewind.yaml KUBE_NAMESPACE_PLACEHOLDER $NAMESPACE

# Set the owner the Codewind deployment and service
setCodewindOwner

# Check if we're on IBM Cloud Private and if so, apply the ibm-privileged-psp
kubectl get images.icp.ibm.com
if [[ $? == 0 ]]; then
    echo "Running on IBM Cloud Private, so applying the 'ibm-privileged-psp' PodSecurityPolicy"
    sed -i "s/SERVICE_ACCOUNT_PLACEHOLDER/$WORKSPACE_SERVICE_ACCOUNT/g" /scripts/kube/ibm-privileged-psp-rb.yaml
    sed -i "s/WORKSPACE_ID_PLACEHOLDER/$CHE_WORKSPACE_ID/g" /scripts/kube/ibm-privileged-psp-rb.yaml
    sed -i "s/OWNER_REFERENCE_NAME_PLACEHOLDER/$OWNER_REFERENCE_NAME/g" /scripts/kube/ibm-privileged-psp-rb.yaml
    sed -i "s/OWNER_REFERENCE_UID_PLACEHOLDER/$OWNER_REFERENCE_UID/g" /scripts/kube/ibm-privileged-psp-rb.yaml
    kubectl create -f /scripts/kube/ibm-privileged-psp-rb.yaml
fi

# Deploy Codewind
echo "Creating the Codewind deployment and service"
kubectl apply -f $DEPLOY_DIR/codewind.yaml