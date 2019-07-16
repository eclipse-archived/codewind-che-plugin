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

# First use this layer to build the go version of filewatcherd
FROM golang:1.12 as builder

# Pull down any dependencies and build the filewatcher daemon go app
WORKDIR /go/src/github.com/eclipse/codewind-filewatchers/Filewatcherd-Go/
COPY ./codewind-filewatchers/Filewatcherd-Go/src/codewind/ .
RUN GOPATH= CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o filewatcherd .

# Build base image for the nginx based sidecar container
FROM nginx:stable-alpine

RUN apk --no-cache add curl openssl jq

# Copy the filewatcherd daemon over from the previous build stage
COPY --from=builder /go/src/github.com/eclipse/codewind-filewatchers/Filewatcherd-Go/filewatcherd /usr/local/bin

COPY ./nginx.conf /etc/nginx/nginx.conf
COPY ./site.conf /etc/nginx/conf.d-default/default.conf

# ensure non-root user 'www-data' exists
# set it to group 0 (root) so that arbitrary userIDs in that group used by kube platforms can also access relevant files/folders
RUN set -x ; \
  adduser -u 82 -D -S -G root www-data && exit 0 ; exit 1

RUN touch /var/run/nginx.pid

# ownership and permissions set up on relvant files/folders for user www-data (uid=82) and group root (gid=0)
RUN chown -R www-data:root /var/run/nginx.pid && \
  chown -R www-data:root /var/cache/nginx && \ 
  chown www-data:root /etc/nginx/conf.d /etc/nginx/nginx.conf /etc/nginx/conf.d-default/default.conf
RUN chmod g+rwx /var/run/nginx.pid && \
  chmod -R g+rwx /var/cache/nginx && \
  chmod -R g+rwx /etc/nginx/conf.d && \
  chmod -R g+rwx /etc/nginx/conf.d-default

RUN curl -LO https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl && \
    chmod +x ./kubectl && \
    mv ./kubectl /usr/local/bin/kubectl

COPY scripts/ /scripts

RUN chmod -R g+rwx /scripts && chown -R www-data:root /scripts

WORKDIR /scripts

USER www-data

EXPOSE 9090

ENTRYPOINT ["/scripts/entrypoint.sh"]