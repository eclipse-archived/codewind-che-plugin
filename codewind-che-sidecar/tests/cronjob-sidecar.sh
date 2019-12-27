#!/usr/bin/env bash

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

CODEWIND_CHE_PLUGIN_DIR=~/codewind-che-plugin
CODEWIND_CHE_PLUGIN_TEST_DIR=$CODEWIND_CHE_PLUGIN_DIR/codewind-che-sidecar/tests
CODEWIND_CHE_PLUGIN_REPO=git@github.com:eclipse/codewind-che-plugin.git
TEST_BRANCH="master"

DATE_NOW=$(date +"%d-%m-%Y")
TIME_NOW=$(date +"%H.%M.%S")
BUCKET_NAME=codewind-che-plugin-sidecar
TEST_OUTPUT_DIR=~/test_results/sidecar/$DATE_NOW/$TIME_NOW
TEST_OUTPUT_TAP=$TEST_OUTPUT_DIR/test_output.tap
TEST_OUTPUT_XML=$TEST_OUTPUT_DIR/test_output.xml

if [[ (-z $NAMESPACE) ]]; then
  echo -e "${RED}Mandatory argument NAMESPACE is not set up. ${RESET}\n"
  echo -e "${RED}Please export variable NAMESPACE to run the Codewind sidecar tests. ${RESET}\n"
  exit 1
fi
if [[ (-z $CLUSTER_IP) ]]; then
  echo -e "${RED}Mandatory argument CLUSTER_IP is not set up. ${RESET}\n"
  echo -e "${RED}Please export variable CLUSTER_IP to run the Codewind sidecar tests. ${RESET}\n"
  exit 1
fi

if [[ (-z $CLUSTER_USER) ]]; then
  echo -e "${RED}Mandatory argument CLUSTER_USER is not set up. ${RESET}\n"
  echo -e "${RED}Please export variable CLUSTER_USER to run the Codewind sidecar tests. ${RESET}\n"
  exit 1
fi

if [[ (-z $CLUSTER_PASSWORD) ]]; then
  echo -e "${RED}Mandatory argument CLUSTER_PASSWORD is not set up. ${RESET}\n"
  echo -e "${RED}Please export variable CLUSTER_PASSWORD to run the Codewind sidecar tests. ${RESET}\n"
  exit 1
fi
if [[ (-z $DASHBOARD_IP) ]]; then
  echo -e "${RED}Dashboard IP is required to upload test results. ${RESET}\n"
  exit 1
fi

oc login $CLUSTER_IP:8443 -u $CLUSTER_USER -p $CLUSTER_PASSWORD
oc project $NAMESPACE
if [[ $? -eq 0 ]]; then
  echo -e "${GREEN}Successfully logged into the OKD cluster ${RESET}\n"
else
  echo -e "${RED}Failed to log into the OKD cluster ${RESET}\n"
  exit 1
fi

export CHE_INGRESS_DOMAIN=$(kubectl get routes --selector=component=che -o jsonpath="{.items[0].spec.host}" 2>&1)
export CHE_NAMESPACE=$NAMESPACE
export CLUSTER_IP

rm -rf $CODEWIND_CHE_PLUGIN_DIR \
&& mkdir -p $TEST_OUTPUT_DIR \
&& git clone $CODEWIND_CHE_PLUGIN_REPO -b $TEST_BRANCH \
&& cd $CODEWIND_CHE_PLUGIN_TEST_DIR \
&& bats --tap sidecarfvt.bats 2>&1 | tee $TEST_OUTPUT_TAP \
&& cat $TEST_OUTPUT_TAP | tap-xunit > $TEST_OUTPUT_XML \
&& curl --header "Content-Type:text/xml" --data-binary @$TEST_OUTPUT_XML --insecure "https://$DASHBOARD_IP/postxmlresult/$BUCKET_NAME/test" > /dev/null \
&& rm -rf $CODEWIND_CHE_PLUGIN_DIR

if [[ ($? -ne 0) ]]; then
    echo -e "${RED}Cronjob has failed. ${RESET}\n"
    exit 1
fi
