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

git clone https://github.com/eclipse/codewind-filewatchers.git

# the command below will echo the head commit if the branch exists, else it just exits
if [[ -n \$(git ls-remote --heads \$INSTALLER_REPO ${env.BRANCH_NAME}) ]]; then
    echo "Will use matching ${env.BRANCH_NAME} branch on \$INSTALLER_REPO"
    export CW_CLI_BRANCH=${env.BRANCH_NAME}
else
    echo "No matching branch on \$INSTALLER_REPO - using \$CW_CLI_BRANCH branch"
fi


docker build --build-arg CW_CLI_BRANCH="$CW_CLI_BRANCH" -t codewind-che-sidecar .
