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
DEPLOY_DIR=/tmp/$CHE_WORKSPACE_ID
DEPLOY_STATUS_FILE=$DEPLOY_DIR/deploy-status.json

function updateDeployStatus() {
    key=$1
    file=$2
    default_message="Failed to update $key in template file." 
    message=${3:-$default_message}
    
    echo "$message Error message appended to $file with key = $key"
    echo $(jq . $file | jq --arg "$key" "$message" ". + {\"$key\": \"$message\"}") >$file
}

# setTemplateValue takes in a templated Kubernetes yaml file, and a corresponding key and value to set
# Uses sed right now. ToDo: Investigate Kustomize
function setTemplateValue() {
    yaml_file=$1
    key=$2
    value=$3

    sed -i "s/$key/$value/g" $yaml_file
    exit_code=$?
    if [[ $exit_code != 0 ]]; then
        updateDeployStatus $key $DEPLOY_STATUS_FILE
    fi
}

# setCodewindPVC finds the Kubernetes PVC associated with the Che workspace and sets it in the 
# Codewind deployment yaml
function setCodewindPVC() {
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

    setTemplateValue $DEPLOY_DIR/codewind.yaml PVC_NAME_PLACEHOLDER $WORKSPACE_PVC
}

# setCodewindRegistrySecret finds the Docker registry secret the user set through Che and sets it in the Codewind deployment yaml.
# It also patches the Che workspace service account to use that secret
function setCodewindRegistrySecret() {
    serviceAccount=$1
    
    # Set the Docker registry secret
    kubectl get secret $CHE_WORKSPACE_ID-private-registries > /dev/null 2>&1
    if [[ $? != 0 ]]; then
        message="Unable to find Che docker registry secret. PFE deployment may fail. Please create a secret via the Che dashboard."
        echo "$message"
        key=DOCKER_REGISTRY_SECRET_NOT_FOUND
        updateDeployStatus $key $DEPLOY_STATUS_FILE $message
    else
        registrySecret=$CHE_WORKSPACE_ID-private-registries
        echo "Setting the registry secret"
        setTemplateValue $DEPLOY_DIR/codewind.yaml REGISTRY_SECRET_PLACEHOLDER $registrySecret

        # Patch the service account before deploying
        echo "Patching the service account with the registry secret"
        kubectl patch serviceaccount $serviceAccount -p "{\"imagePullSecrets\": [{\"name\": \"$registrySecret\"}]}"
    fi
}

# setCodewindServiceAccount finds the Kuebernetes service account associated with the Che workspace and sets
# it in the Codewind deployment yaml. This allows Codewind to use the same imagePullSecrets that's available to the Che workspace
function setCodewindServiceAccount() {
    # Need to determine the service account for the workspace
    WORKSPACE_SERVICE_ACCOUNT=$(kubectl get po --selector=che.original_name=che-workspace-pod -o jsonpath="{.items[0].spec.serviceAccountName}")
    if [[ -z $WORKSPACE_SERVICE_ACCOUNT ]]; then
        echo "Unable to find the service account name of the workspace namespace"
        echo "Defaulting to che-workspace"
        WORKSPACE_SERVICE_ACCOUNT=che-workspace
    else
        echo "Found the workspace namespace service account, using it..."
    fi

    setTemplateValue $DEPLOY_DIR/codewind.yaml SERVICE_ACCOUNT_PLACEHOLDER $WORKSPACE_SERVICE_ACCOUNT

    # Set the registry secret for Codewind too
    setCodewindRegistrySecret $WORKSPACE_SERVICE_ACCOUNT
}

function setCodewindOwner() {
    # Set the owner reference to the same owner of the che workspace pod to ensure Codewind resources are cleaned up when the workspace is deleted
    OWNER_REFERENCE_NAME=$( kubectl get po --selector=che.original_name=che-workspace-pod,che.workspace_id=$CHE_WORKSPACE_ID -o jsonpath='{.items[0].metadata.ownerReferences[0].name}' )
    OWNER_REFERENCE_UID=$( kubectl get po --selector=che.original_name=che-workspace-pod,che.workspace_id=$CHE_WORKSPACE_ID -o jsonpath='{.items[0].metadata.ownerReferences[0].uid}' )
    setTemplateValue $DEPLOY_DIR/codewind.yaml OWNER_REFERENCE_NAME_PLACEHOLDER $OWNER_REFERENCE_NAME
    setTemplateValue $DEPLOY_DIR/codewind.yaml OWNER_REFERENCE_UID_PLACEHOLDER $OWNER_REFERENCE_UID
}