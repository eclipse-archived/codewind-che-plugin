package main

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// createPFEDeploy creates a Kubernetes deploy for Codewind, marking the Che workspace as its owner
func createPFEDeploy(codewind CodewindDeployment) appsv1.Deployment {
	labels := map[string]string{
		"app":          "codewind-pfe",
		"pfeWorkspace": codewind.WorkspaceID,
	}

	blockOwnerDeletion := true
	controller := true
	replicas := int32(1)
	secretMode := int32(511)
	deployment := appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      CodewindPrefix + codewind.WorkspaceID,
			Namespace: codewind.Namespace,
			Labels:    labels,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         "apps/v1",
					BlockOwnerDeletion: &blockOwnerDeletion,
					Controller:         &controller,
					Kind:               "ReplicaSet",
					Name:               codewind.OwnerReferenceName,
					UID:                codewind.OwnerReferenceUID,
				},
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: codewind.ServiceAccountName,
					Volumes: []corev1.Volume{
						{
							Name: "shared-workspace",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: codewind.PVCName,
								},
							},
						},
						{
							Name: "buildah-volume",
						},
						{
							Name: "registry-secret",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									DefaultMode: &secretMode,
									SecretName:  codewind.PullSecret,
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:            PFEContainerName,
							Image:           PFEImage + ":" + PFEImageTag,
							ImagePullPolicy: PFEImagePullPolicy,
							SecurityContext: &corev1.SecurityContext{
								Privileged: &codewind.Privileged,
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "shared-workspace",
									MountPath: "/codewind-workspace",
									SubPath:   codewind.WorkspaceID + "/projects",
								},
								{
									Name:      "buildah-volume",
									MountPath: "/var/lib/containers",
								},
								{
									Name:      "registry-secret",
									MountPath: "/tmp/secret",
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "IN_K8",
									Value: "true",
								},
								{
									Name:  "PORTAL_HTTPS",
									Value: "true",
								},
								{
									Name:  "KUBE_NAMESPACE",
									Value: codewind.Namespace,
								},
								{
									Name:  "TILLER_NAMESPACE",
									Value: codewind.Namespace,
								},
								{
									Name:  "CHE_WORKSPACE_ID",
									Value: codewind.WorkspaceID,
								},
								{
									Name:  "PVC_NAME",
									Value: codewind.PVCName,
								},
								{
									Name:  "SERVICE_NAME",
									Value: "codewind-" + codewind.WorkspaceID,
								},
								{
									Name:  "SERVICE_NAME",
									Value: "codewind-" + codewind.WorkspaceID,
								},
								{
									Name:  "SERVICE_ACCOUNT_NAME",
									Value: codewind.ServiceAccountName,
								},
								{
									Name:  "MICROCLIMATE_RELEASE_NAME",
									Value: "RELEASE-NAME",
								},
								{
									Name:  "HOST_WORKSPACE_DIRECTORY",
									Value: "/projects",
								},
								{
									Name:  "CONTAINER_WORKSPACE_DIRECTORY",
									Value: "/codewind-workspace",
								},
								{
									Name:  "OWNER_REF_NAME",
									Value: codewind.OwnerReferenceName,
								},
								{
									Name:  "OWNER_REF_UID",
									Value: string(codewind.OwnerReferenceUID),
								},
							},
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: int32(PFEContainerPort),
								},
							},
						},
					},
				},
			},
		},
	}
	return deployment
}

// createPFEService creates a Kubernetes service for Codewind, exposing port 9191
func createPFEService(codewind CodewindDeployment) corev1.Service {
	labels := map[string]string{
		"app":          "codewind-pfe",
		"pfeWorkspace": codewind.WorkspaceID,
	}
	service := corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      codewind.Name,
			Namespace: codewind.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Port: int32(PFEContainerPort),
					Name: "portal-http",
				},
			},
			Selector: labels,
		},
	}
	return service
}
