package codewind

import corev1 "k8s.io/api/core/v1"

const (
	PFEPrefix         = "codewind"
	PerformancePrefix = PFEPrefix + "-performance"

	PFEImage         = "eclipse/codewind-pfe-amd64"
	PerformanceImage = "eclipse/codewind-performance-amd64"

	PFEImageTag    = "latest"
	PerformanceTag = "latest"

	ImagePullPolicy = corev1.PullAlways

	PFEContainerPort         = 9191
	PerformanceContainerPort = 9095
)
