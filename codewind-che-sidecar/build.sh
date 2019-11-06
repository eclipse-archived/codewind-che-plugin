#!/bin/bash

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

# README FIRST !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
# 
# NOTE: change of this file should be in sync with 'Jenkinsfile(stage: Build Docker images)'
# Ping kjoseph@ca.ibm.com for details
#
# !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!

# Move to the directory of the script that included this include file ------
export SCRIPT_LOCT=$( cd $( dirname $0 ); pwd )
cd $SCRIPT_LOCT

# Extract the filewatcherd codebase
if [ -d "codewind-filewatchers" ]; then
  rm -rf codewind-filewatchers
fi

INSTALLER_REPO="https://github.com/eclipse/codewind-installer.git"
BRANCH_NAME=`git rev-parse --abbrev-ref HEAD`

git clone https://github.com/eclipse/codewind-filewatchers.git

BLUE='\033[1;34m'
NC='\033[0m'

printf "\n${BLUE}Warning: Verify that the installer target branch below is the branch you expect to be targeting with your build${NC}\n"
sleep 1

# the command below will echo the head commit if the branch exists, else it just exits
if [[ -n $(git ls-remote --heads $INSTALLER_REPO ${BRANCH_NAME}) ]]; then
    echo "Will use matching ${BRANCH_NAME} branch on $INSTALLER_REPO"
    export CW_CLI_BRANCH=${BRANCH_NAME}
else
    export CW_CLI_BRANCH=master
    source scripts/installer-branch-override.env
    echo "No matching branch on $INSTALLER_REPO - using $CW_CLI_BRANCH branch. Override this in installer-branch-override.env"
fi


echo

docker build --build-arg CW_CLI_BRANCH="$CW_CLI_BRANCH" -t codewind-che-sidecar .
