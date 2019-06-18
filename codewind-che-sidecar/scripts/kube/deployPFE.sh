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
sed -i "s/PVC_NAME_PLACEHOLDER/$WORKSPACE_PVC/g" /scripts/kube/codewind_template.yaml

# Need to determine the service account for the workspace
WORKSPACE_SERVICE_ACCOUNT=$(kubectl get po --selector=che.original_name=che-workspace-pod -o jsonpath="{.items[0].spec.serviceAccountName}")
if [[ -z $WORKSPACE_SERVICE_ACCOUNT ]]; then
    echo "Unable to find the service account name of the workspace namespace"
    echo "Defaulting to che-workspace"
    WORKSPACE_SERVICE_ACCOUNT=che-workspace
else
    echo "Found the workspace namespace service account, using it..."
fi
sed -i "s/SERVICE_ACCOUNT_PLACEHOLDER/$WORKSPACE_SERVICE_ACCOUNT/g" /scripts/kube/codewind_template.yaml

# Set the subpath for the projects volume mount
echo "Setting the subpath for the projects volume mount"
sed -i "s/WORKSPACE_ID_PLACEHOLDER/$CHE_WORKSPACE_ID/g" /scripts/kube/codewind_template.yaml

# replace environment specific values
NAMESPACE=$( kubectl get po --selector=che.original_name=che-workspace-pod -o jsonpath='{.items[0].metadata.namespace}' )

# Set the Docker registry
DOCKER_REGISTRY=mycluster.icp:8500\\/$NAMESPACE
echo "Setting the docker registry"
sed -i "s/DOCKER_REGISTRY_PLACEHOLDER/$DOCKER_REGISTRY/g" /scripts/kube/codewind_template.yaml

# Set the Docker registry secret
kubectl get secret $CHE_WORKSPACE_ID-private-registries > /dev/null 2>&1
if [[ $? != 0 ]]; then
    echo "Unable to find Che docker registry secret. PFE deployment may fail."
    echo "Please create a secret via the Che dashboard."
else
    REGISTRY_SECRET=$CHE_WORKSPACE_ID-private-registries
    echo "Setting the registry secret"
    sed -i "s/REGISTRY_SECRET_PLACEHOLDER/$REGISTRY_SECRET/g" /scripts/kube/codewind_template.yaml
fi

# Patch the service account before deploying
echo "Patching the service account with the registry secret"
kubectl patch serviceaccount $WORKSPACE_SERVICE_ACCOUNT -p "{\"imagePullSecrets\": [{\"name\": \"$REGISTRY_SECRET\"}]}"

# Set the owner reference to ensure clean up
OWNER_REFERENCE_NAME=$( kubectl get po --selector=che.original_name=che-workspace-pod,che.workspace_id=$CHE_WORKSPACE_ID -o jsonpath='{.items[0].metadata.ownerReferences[0].name}' )
sed -i "s/OWNER_REFERENCE_NAME_PLACEHOLDER/$OWNER_REFERENCE_NAME/g" /scripts/kube/codewind_template.yaml

OWNER_REFERENCE_UID=$( kubectl get po --selector=che.original_name=che-workspace-pod,che.workspace_id=$CHE_WORKSPACE_ID -o jsonpath='{.items[0].metadata.ownerReferences[0].uid}' )
sed -i "s/OWNER_REFERENCE_UID_PLACEHOLDER/$OWNER_REFERENCE_UID/g" /scripts/kube/codewind_template.yaml

echo "Creating the Codewind deployment and service"
sed "s/KUBE_NAMESPACE_PLACEHOLDER/$NAMESPACE/g" /scripts/kube/codewind_template.yaml | kubectl apply -f -
