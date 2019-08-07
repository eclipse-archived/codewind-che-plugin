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

touch /var/run/nginx.pid

cp /etc/nginx/conf.d-default/default.conf /etc/nginx/conf.d/default.conf

# Generate SSL certificate/key
openssl req -subj '/CN=localhost' -x509 -newkey rsa:4096 -nodes -keyout /etc/nginx/conf.d/key.pem -out /etc/nginx/conf.d/cert.pem -days 1825

# Discovery of codewind service in a multi-workspace per namespace scenario
deploy-pfe

CWServiceName=$(deploy-pfe get-service)
if [ -z $CWServiceName ]; then
    echo "ERROR: The Codewind service was not found. Aborting..."
    exit 1
fi

echo "Codewind is now ready."
echo "Setting proxy to Codewind service: $CWServiceName"
CWServiceNameEndpoint=https://$CWServiceName:9191

_____FROM_DOMAIN_NAME="${_____FROM_DOMAIN_NAME:-localhost}"
_____TO_DOMAIN_NAME="${_____TO_DOMAIN_NAME:-$CWServiceNameEndpoint}"
_____LISTEN_PORT="${_____LISTEN_PORT:-9090}"

# Place nginx config values into the .conf file via substitution of templates
echo "Substituting values into nginx's default.conf"
sed -i 's|_____SUBSTITUTED_FROM_DOMAIN_NAME|'$_____FROM_DOMAIN_NAME'|g' /etc/nginx/conf.d/default.conf
sed -i 's|_____SUBSTITUTED_TO_DOMAIN_NAME|'$_____TO_DOMAIN_NAME'|g' /etc/nginx/conf.d/default.conf
sed -i 's|_____SUBSTITUTED_LISTEN_PORT|'$_____LISTEN_PORT'|g' /etc/nginx/conf.d/default.conf

# Start nginx process
nginx -g "daemon on;"
status=$?
if [ $status -ne 0 ]; then
    echo "Failed to start nginx: $status"
    exit $status
fi
echo "Started nginx"

# Start filewatcherd process
filewatcherd $CWServiceNameEndpoint &
status=$?
if [ $status -ne 0 ]; then
    echo "Failed to start filewatcherd: $status"
    exit $status
fi
echo "Started filewatcherd"

# Monitor nginx and filewatcherd processes every 10 seconds, exit if any of the two fail
while sleep 10; do
    ps aux | grep nginx | grep -q -v grep
    NGINX_PROCESS_STATUS=$?
    ps aux | grep filewatcherd | grep -q -v grep
    FILEWATCHERD_PROCESS_STATUS=$?
    if [ $NGINX_PROCESS_STATUS -ne 0 ]; then
        echo "Nginx process failed"
        exit 1
    fi
    if [ $FILEWATCHERD_PROCESS_STATUS -ne 0 ]; then
        filewatcherd $CWServiceNameEndpoint &
    fi
done