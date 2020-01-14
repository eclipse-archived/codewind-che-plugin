#!/usr/bin/env bats

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

# Load common utility functions
load testutil

# Setup code to run before each test
setup() {
    export time_before=$SECONDS

    if [ -z "$CHE_INGRESS_DOMAIN" ]; then
        echo "# Che ingress domain is not defined in " '$CHE_INGRESS_DOMAIN' >&3
        exit 1
    fi

    if [ -z "$CHE_NAMESPACE" ]; then
        echo "# Che namespace is not defined in " '$CHE_NAMESPACE' >&3
        exit 1
    fi

    if [ -z "$CLUSTER_IP" ]; then
        echo "# Cluster IP address is not defined in " '$CLUSTER_IP' >&3
        exit 1
    fi

    export CODEWIND_DEVFILE_URL=https://raw.githubusercontent.com/eclipse/codewind-che-plugin/master/devfiles/latest/devfile.yaml
    export KUBE_NAMESPACE_ARG=${CHE_NAMESPACE}

    # Discover workspace ID written into temporary file during workspace creation
    if [ -f che_workspace_id.txt ]; then
        export CHE_WORKSPACE_ID=$(cat che_workspace_id.txt)
    else
        export CHE_WORKSPACE_ID=workspace00000
    fi

    # Discover workspace pod and sidecar full names based on workspace ID
    export CHE_WORKSPACE_POD_FULLNAME=$(kubectl get pods -l che.original_name=che-workspace-pod --no-headers -o custom-columns=":metadata.name" -n $KUBE_NAMESPACE_ARG | grep $CHE_WORKSPACE_ID)
    export SIDECAR_CONTAINER_FULLNAME=$(kubectl get pods $CHE_WORKSPACE_POD_FULLNAME -o jsonpath='{.spec.containers[*].name}' -n $KUBE_NAMESPACE_ARG | sed 's/ /\n/g' | grep ^codewind-che-sidecar)

    # Set up Che access token for multi-user Che environment
    CHE_USER="admin"
    CHE_PASS="admin"
    KEYCLOAK_HOSTNAME=$(kubectl get routes --selector=component=keycloak -o jsonpath="{.items[0].spec.host}" 2>&1)
    TOKEN_ENDPOINT="http://${KEYCLOAK_HOSTNAME}/auth/realms/che/protocol/openid-connect/token" 
    export CHE_ACCESS_TOKEN=$(curl -sSL --data "grant_type=password&client_id=che-public&username=${CHE_USER}&password=${CHE_PASS}" ${TOKEN_ENDPOINT} | jq -r '.access_token')
}

# Teardown code after each test
teardown() {
    time_after=$SECONDS
    echo "# Test time taken: " $(($time_after-$time_before)) " seconds" >&3
}

@test "Codewind Sidecar Test 1: Create Che workspace from Codewind dev file" {
    deleteExistingCodewindCheWorkspaces
    createCodewindCheWorkspace
}

@test "Codewind Sidecar Test 2: Verify Codewind workspace pod is running" {
    # Check if pod has started, timeout after 10 minutes
    endtime=$(($SECONDS + 600))
    pod_running=false
    while (( $SECONDS < $endtime )); do
        run getCodewindPod
        if [[ $output = *"Running"* ]]; then
            pod_running=true
            break
        fi
        if [[ $output = *"Failure"* || $output = *"Unknown"* || $output = *"ImagePullBackOff"* || $output = *"CrashLoopBackOff"* || $output = *"PostStartHookError"* ]]; then
            echo "# Error: Codewind pod failed to start" >&3
            exit 1
        fi
    
        sleep 2
    done
    
    [ $pod_running = "true" ]
}

@test "Codewind Sidecar Test 3: Verify sidecar container is running and ready, and Codewind service successfully deployed" {
    # Check if sidecar main processes have started after codewind server deployment, timeout after 10 minutes
    endtime=$(($SECONDS + 600))
    nginx_process_running=false
    filewatcherd_process_running=false
    while (( $SECONDS < $endtime )); do
        run getPIDofProcessInContainer $CHE_WORKSPACE_POD_FULLNAME $SIDECAR_CONTAINER_FULLNAME nginx
        if [ "$status" -eq 0 ]; then
            nginx_process_running=true
        fi

        run getPIDofProcessInContainer $CHE_WORKSPACE_POD_FULLNAME $SIDECAR_CONTAINER_FULLNAME filewatcherd
        if [ "$status" -eq 0 ]; then
            filewatcherd_process_running=true
        fi

        if [[ $nginx_process_running = "true" && $filewatcherd_process_running = "true" ]]; then
            break
        fi

        sleep 2
    done

    [ $nginx_process_running = "true" ]
    [ $filewatcherd_process_running = "true" ]

    # Allow some more time for sidecar container to settle
    sleep 30

    checkSidecarContainerReady

    cw_service_name=$(kubectl get svc --selector=app=codewind-pfe,codewindWorkspace=$CHE_WORKSPACE_ID -o jsonpath="{.items[0].metadata.name}" -n $KUBE_NAMESPACE_ARG)
    [ ! -z "$cw_service_name" ]
}

@test "Codewind Sidecar Test 4: Verify filewatcher daemon is up & running" {
    # Check that the filewatcher daemon properly started, timeout after 2 minutes
    endtime=$(($SECONDS + 120))
    filewatcherd_ready=false
    while (( $SECONDS < $endtime )); do
        run checkFilewatcherDaemonRunning
        if [ "$status" -eq 0 ]; then 
            filewatcherd_ready=true
            break
        fi
    done

    [ $filewatcherd_ready = "true" ]
}

@test "Codewind Sidecar Test 5: Verify filewatcher daemon restarts after kill" {
    time_before_kill=$SECONDS

    # Kill filewatcherd process in the sidecar container
    fwd_pid=$(getPIDofProcessInContainer $CHE_WORKSPACE_POD_FULLNAME $SIDECAR_CONTAINER_FULLNAME filewatcherd)
    kubectl exec -t $CHE_WORKSPACE_POD_FULLNAME -n $KUBE_NAMESPACE_ARG --container $SIDECAR_CONTAINER_FULLNAME -- kill $fwd_pid

    # Check every 5 seconds if filewatcherd has restarted, timeout after 5 minutes
    endtime=$(($SECONDS + 300))
    fwd_restarted=false
    while (( $SECONDS < $endtime )); do
        sleep 5
        run getFileWatcherDaemonProcess
        if [ "$status" -eq 0 ]; then
            fwd_restarted=true
            break
        fi
    done
        
    [ $fwd_restarted = "true" ]

    # Allow some time for filewatcherd to settle
    sleep 10

    # Calculate approx time elapsed (in seconds) between filewatcher daemon kill and restart so as to only check the logs during that time
    time_elapsed="$(($SECONDS - $time_before_kill))"

    # Check if the filewatcher daemon started properly
    checkFilewatcherDaemonRunning "$time_elapsed"
}

@test "Codewind Sidecar Test 6: Stop and delete the Codewind Che workspace" {
    # Delete temporary file housing the workspace ID
    if [ -f che_workspace_id.txt ]; then
        rm che_workspace_id.txt
    fi

    stopCodewindCheWorkspace
    deleteCodewindCheWorkspace
}
