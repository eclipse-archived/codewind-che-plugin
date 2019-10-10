package constants

import corev1 "k8s.io/api/core/v1"

const (
	// PFEPrefix is the prefix all PFE-related resources: deployment, service, and ingress/route
	PFEPrefix = "codewind"

	// PerformancePrefix is the prefix for all performance-dashboard related resources: deployment and service
	PerformancePrefix = PFEPrefix + "-performance"

	// PFEImage is the docker image that will be used in the Codewind-PFE pod
	PFEImage = "eclipse/codewind-pfe-amd64"

	// PerformanceImage is the docker image that will be used in the Performance dashboard pod
	PerformanceImage = "eclipse/codewind-performance-amd64"

	// PFEImageTag is the image tag associated with the docker image that's used for Codewind-PFE
	PFEImageTag = "0.5"

	// PerformanceTag is the image tag associated with the docker image that's used for the Performance dashboard
	PerformanceTag = "0.5"

	// ImagePullPolicy is the pull policy used for all containers in Codewind, defaults to Always
	ImagePullPolicy = corev1.PullAlways

	// PFEContainerPort is the port at which Codewind-PFE is exposed
	PFEContainerPort = 9191

	// PerformanceContainerPort is the port at which the Performance dashboard is exposed
	PerformanceContainerPort = 9095
)
