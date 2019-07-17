package main

import "k8s.io/apimachinery/pkg/types"

// CodewindDeployment represents a Codewind deployment: name, namespace, volume, serviceaccount, and pull secrets
type CodewindDeployment struct {
	Name               string
	Namespace          string
	WorkspaceID        string
	PVCName            string
	ServiceAccountName string
	PullSecret         string
	OwnerReferenceName string
	OwnerReferenceUID  types.UID
	Privileged         bool
}

// ServiceAccountPatch contains an array of imagePullSecrets that will be patched into a Kubernetes service account
type ServiceAccountPatch struct {
	ImagePullSecrets *[]ImagePullSecret `json:"imagePullSecrets"`
}

// ImagePullSecret represents a Kubernetes imagePullSecret
type ImagePullSecret struct {
	Name string `json:"name"`
}
