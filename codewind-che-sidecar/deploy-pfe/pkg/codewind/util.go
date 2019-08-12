package codewind

import (
	"encoding/json"
	"os"

	"deploy-pfe/pkg/constants"

	"k8s.io/client-go/kubernetes"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	extensionsv1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"

	v1 "github.com/openshift/api/route/v1"
)

// PatchServiceAccount takes in a list of secret names, and patches it to the specified service account
func PatchServiceAccount(clientset *kubernetes.Clientset, codewind Codewind) error {
	patch := ServiceAccountPatch{
		ImagePullSecrets: &[]ImagePullSecret{
			{
				Name: codewind.PullSecret,
			},
		},
	}

	b, err := json.Marshal(patch)
	if err != nil {
		return err
	}

	_, err = clientset.CoreV1().ServiceAccounts(codewind.Namespace).Patch(codewind.ServiceAccountName, types.StrategicMergePatchType, b)
	if err != nil {
		return err
	}
	return nil
}

func setPFEEnvVars(codewind Codewind) []corev1.EnvVar {
	return []corev1.EnvVar{
		{
			Name:  "TEKTON_PIPELINE",
			Value: "tekton-pipelines",
		},
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
		{
			Name:  "CODEWIND_PERFORMANCE_SERVICE",
			Value: codewind.PerformanceName + "-" + codewind.WorkspaceID,
		},
		{
			Name:  "CHE_INGRESS_HOST",
			Value: codewind.Ingress,
		},
	}
}

func setPerformanceEnvVars(codewind Codewind) []corev1.EnvVar {
	return []corev1.EnvVar{
		{
			Name:  "IN_K8",
			Value: "true",
		},
		{
			Name:  "PORTAL_HTTPS",
			Value: "false",
		},
		{
			Name:  "CODEWIND_INGRESS",
			Value: codewind.Ingress,
		},
	}
}

// setPFEVolumes returns the 3 volumes & corresponding volume mounts required by the PFE container:
// project workspace, buildah volume, and the docker registry secret (the latter of which is optional)
func setPFEVolumes(codewind Codewind) ([]corev1.Volume, []corev1.VolumeMount) {
	secretMode := int32(511)
	isOptional := true

	volumes := []corev1.Volume{
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
					Optional:    &isOptional,
				},
			},
		},
	}

	volumeMounts := []corev1.VolumeMount{
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
	}

	return volumes, volumeMounts
}

// generateDeployment returns a Kubernetes deployment object with the given name for the given image.
// Additionally, volume/volumemounts and env vars can be specified.
func generateDeployment(codewind Codewind, name string, image string, port int, volumes []corev1.Volume, volumeMounts []corev1.VolumeMount, envVars []corev1.EnvVar, labels map[string]string) appsv1.Deployment {
	blockOwnerDeletion := true
	controller := true
	replicas := int32(1)
	deployment := appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name + "-" + codewind.WorkspaceID,
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
					Volumes:            volumes,
					Containers: []corev1.Container{
						{
							Name:            name,
							Image:           image,
							ImagePullPolicy: constants.ImagePullPolicy,
							SecurityContext: &corev1.SecurityContext{
								Privileged: &codewind.Privileged,
							},
							VolumeMounts: volumeMounts,
							Env:          envVars,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: int32(port),
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

// generateService returns a Kubernetes service object with the given name, exposed over the specified port
// for the container with the given labels.
func generateService(codewind Codewind, name string, port int, labels map[string]string) corev1.Service {
	blockOwnerDeletion := true
	controller := true

	service := corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name + "-" + codewind.WorkspaceID,
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
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Port: int32(port),
					Name: name + "-http",
				},
			},
			Selector: labels,
		},
	}
	return service
}

// CreateRoute returns an OpenShift route for the Codewind PFE service
func CreateRoute(codewind Codewind) v1.Route {
	labels := map[string]string{
		"app":               constants.PFEPrefix,
		"codewindWorkspace": codewind.WorkspaceID,
	}

	weight := int32(100)
	blockOwnerDeletion := true
	controller := true

	return v1.Route{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Route",
			APIVersion: "route.openshift.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   constants.PFEPrefix + "-" + codewind.WorkspaceID,
			Labels: labels,
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
		Spec: v1.RouteSpec{
			Host: codewind.Ingress,
			Port: &v1.RoutePort{
				TargetPort: intstr.FromInt(constants.PFEContainerPort),
			},
			TLS: &v1.TLSConfig{
				InsecureEdgeTerminationPolicy: v1.InsecureEdgeTerminationPolicyRedirect,
				Termination:                   v1.TLSTerminationPassthrough,
			},
			To: v1.RouteTargetReference{
				Kind:   "Service",
				Name:   constants.PFEPrefix + "-" + codewind.WorkspaceID,
				Weight: &weight,
			},
		},
	}
}

// CreateIngress returns a Kubernetes ingress for the Codewind PFE service
func CreateIngress(codewind Codewind) extensionsv1.Ingress {
	labels := map[string]string{
		"app":               constants.PFEPrefix,
		"codewindWorkspace": codewind.WorkspaceID,
	}

	annotations := map[string]string{
		"nginx.ingress.kubernetes.io/rewrite-target":   "/",
		"nginx.ingress.kubernetes.io/backend-protocol": "HTTPS",
	}
	blockOwnerDeletion := true
	controller := true

	return extensionsv1.Ingress{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "extensions/v1beta1",
			Kind:       "Ingress",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        constants.PFEPrefix + "-" + codewind.WorkspaceID,
			Annotations: annotations,
			Labels:      labels,
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
		Spec: extensionsv1.IngressSpec{
			Rules: []extensionsv1.IngressRule{
				{
					Host: codewind.Ingress,
					IngressRuleValue: extensionsv1.IngressRuleValue{
						HTTP: &extensionsv1.HTTPIngressRuleValue{
							Paths: []extensionsv1.HTTPIngressPath{
								{
									Path: "/",
									Backend: extensionsv1.IngressBackend{
										ServiceName: constants.PFEPrefix + "-" + codewind.WorkspaceID,
										ServicePort: intstr.FromInt(constants.PerformanceContainerPort),
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// GetImages returns the images that are to be used for PFE and the Performance dashboard in Codewind
// If environment vars are set (such as $PFE_IMAGE, $PFE_TAG, $PERFORMANCE_IMAGE, or $PERFORMANCE_TAG), it will use those,
// otherwise it defaults to the constants defined in constants/default.go
func GetImages() (string, string) {
	var pfeImage, performanceImage, pfeTag, performanceTag string

	if pfeImage = os.Getenv("PFE_IMAGE"); pfeImage == "" {
		pfeImage = constants.PFEImage
	}
	if performanceImage = os.Getenv("PERFORMANCE_IMAGE"); performanceImage == "" {
		performanceImage = constants.PerformanceImage
	}
	if pfeTag = os.Getenv("PFE_TAG"); pfeTag == "" {
		pfeTag = constants.PFEImageTag
	}
	if performanceTag = os.Getenv("PERFORMANCE_TAG"); performanceTag == "" {
		performanceTag = constants.PerformanceTag
	}

	return pfeImage + ":" + pfeTag, performanceImage + ":" + performanceTag
}
