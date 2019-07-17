package main

import corev1 "k8s.io/api/core/v1"

const (
	CodewindPrefix     = "codewind-"
	PFEContainerName   = "codewind-pfe"
	PFEImage           = "sys-mcs-docker-local.artifactory.swg-devops.com/codewind-pfe-amd64"
	PFEImageTag        = "latest"
	PFEImagePullPolicy = corev1.PullAlways
	PFEContainerPort   = 9191
)
