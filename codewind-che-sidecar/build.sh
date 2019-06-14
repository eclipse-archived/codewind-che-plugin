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


# Move to the directory of the script that included this include file ------
export SCRIPT_LOCT=$( cd $( dirname $0 ); pwd )
cd $SCRIPT_LOCT

# Extract the filewatcherd codebase
if [ -d "codewind-filewatchers" ]; then
  rm -rf codewind-filewatchers
fi

git clone https://github.com/eclipse/codewind-filewatchers.git

docker build -t codewind-che-sidecar .
