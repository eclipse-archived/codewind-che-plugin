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

SCRIPTS_DIR=/scripts

## touch a deploy messages json file
DEPLOY_STATUS_FILE=$SCRIPTS_DIR/deploy-status.json
echo '{}' >$DEPLOY_STATUS_FILE
echo "Created an empty deployment status file."

function updateDeployStatus() {
    key=$1
    file=$2
    default_message="Failed to update $key in template file." 
    message=${3:-$default_message}
    
    echo "$message Error message appended to $file with key = $key"
    echo $(jq . $file | jq --arg "$key" "$message" ". + {\"$key\": \"$message\"}") >$file
}

# Need to determine the PVC for the workspace that mounts /projects
# If the claim-che-workspace PVC doesn't exist, we need to get the PVC tagged with the workspace id
kubectl get pvc claim-che-workspace > /dev/null 2>&1
if [[ $? != 0 ]]; then
    echo "Unable to find claim-che-workspace PVC, Che workspace running in shared namespace"
    WORKSPACE_PVC=$(kubectl get pvc --selector=che.workspace.volume_name=projects,che.workspace_id=$CHE_WORKSPACE_ID -o jsonpath="{.items[0].metadata.name}")
    echo "Found $WORKSPACE_PVC instead."
else
    echo "Found claim-che-workspace PVC, using it..."
    WORKSPACE_PVC=claim-che-workspace
fi
sed -i "s/PVC_NAME_PLACEHOLDER/$WORKSPACE_PVC/g" $SCRIPTS_DIR/kube/codewind_template.yaml
exit_code=$?
if [[ $exit_code != 0 ]]; then
    key=PVC_NAME_PLACEHOLDER
    updateDeployStatus $key $DEPLOY_STATUS_FILE
fi

# Need to determine the service account for the workspace
WORKSPACE_SERVICE_ACCOUNT=$(kubectl get po --selector=che.original_name=che-workspace-pod -o jsonpath="{.items[0].spec.serviceAccountName}")
if [[ -z $WORKSPACE_SERVICE_ACCOUNT ]]; then
    echo "Unable to find the service account name of the workspace namespace"
    echo "Defaulting to che-workspace"
    WORKSPACE_SERVICE_ACCOUNT=che-workspace
else
    echo "Found the workspace namespace service account, using it..."
fi
sed -i "s/SERVICE_ACCOUNT_PLACEHOLDER/$WORKSPACE_SERVICE_ACCOUNT/g" $SCRIPTS_DIR/kube/codewind_template.yaml
exit_code=$?
if [[ $exit_code != 0 ]]; then
    key=SERVICE_ACCOUNT_PLACEHOLDER
    updateDeployStatus $key $DEPLOY_STATUS_FILE
fi

# Set the subpath for the projects volume mount
echo "Setting the subpath for the projects volume mount"
sed -i "s/WORKSPACE_ID_PLACEHOLDER/$CHE_WORKSPACE_ID/g" $SCRIPTS_DIR/kube/codewind_template.yaml
exit_code=$?
if [[ $exit_code != 0 ]]; then
    key=WORKSPACE_ID_PLACEHOLDER
    updateDeployStatus $key $DEPLOY_STATUS_FILE
fi

# replace environment specific values
NAMESPACE=$( kubectl get po --selector=che.original_name=che-workspace-pod -o jsonpath='{.items[0].metadata.namespace}' )

# Set the Docker registry
DOCKER_REGISTRY=mycluster.icp:8500\\/$NAMESPACE
echo "Setting the docker registry"
sed -i "s/DOCKER_REGISTRY_PLACEHOLDER/$DOCKER_REGISTRY/g" $SCRIPTS_DIR/kube/codewind_template.yaml
exit_code=$?
if [[ $exit_code != 0 ]]; then
    key=DOCKER_REGISTRY_PLACEHOLDER
    updateDeployStatus $key $DEPLOY_STATUS_FILE
fi

# Set the Docker registry secret
kubectl get secret $CHE_WORKSPACE_ID-private-registries > /dev/null 2>&1
if [[ $? != 0 ]]; then
    message="Unable to find Che docker registry secret. PFE deployment may fail. Please create a secret via the Che dashboard."
    echo "$message"
    key=DOCKER_REGISTRY_SECRET_NOT_FOUND
    updateDeployStatus $key $DEPLOY_STATUS_FILE $message
else
    REGISTRY_SECRET=$CHE_WORKSPACE_ID-private-registries
    echo "Setting the registry secret"
    sed -i "s/REGISTRY_SECRET_PLACEHOLDER/$REGISTRY_SECRET/g" $SCRIPTS_DIR/kube/codewind_template.yaml
    exit_code=$?
    if [[ $exit_code != 0 ]]; then
        key=REGISTRY_SECRET_PLACEHOLDER
        updateDeployStatus $key $DEPLOY_STATUS_FILE
    fi
fi

# Patch the service account before deploying
echo "Patching the service account with the registry secret"
kubectl patch serviceaccount $WORKSPACE_SERVICE_ACCOUNT -p "{\"imagePullSecrets\": [{\"name\": \"$REGISTRY_SECRET\"}]}"

# Set the owner reference to the same owner of the che workspace pod to ensure Codewind resources are cleaned up when the workspace is deleted
OWNER_REFERENCE_NAME=$( kubectl get po --selector=che.original_name=che-workspace-pod,che.workspace_id=$CHE_WORKSPACE_ID -o jsonpath='{.items[0].metadata.ownerReferences[0].name}' )
sed -i "s/OWNER_REFERENCE_NAME_PLACEHOLDER/$OWNER_REFERENCE_NAME/g" /scripts/kube/codewind_template.yaml

OWNER_REFERENCE_UID=$( kubectl get po --selector=che.original_name=che-workspace-pod,che.workspace_id=$CHE_WORKSPACE_ID -o jsonpath='{.items[0].metadata.ownerReferences[0].uid}' )
sed -i "s/OWNER_REFERENCE_UID_PLACEHOLDER/$OWNER_REFERENCE_UID/g" /scripts/kube/codewind_template.yaml

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

echo "Creating the Codewind deployment and service"
sed "s/KUBE_NAMESPACE_PLACEHOLDER/$NAMESPACE/g" $SCRIPTS_DIR/kube/codewind_template.yaml | kubectl apply -f -
exit_code=$?
if [[ $exit_code != 0 ]]; then
    key=KUBE_NAMESPACE_PLACEHOLDER
    updateDeployStatus $key $DEPLOY_STATUS_FILE
fi
