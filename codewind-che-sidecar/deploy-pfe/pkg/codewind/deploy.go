package codewind

import (
	log "github.com/sirupsen/logrus"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"k8s.io/client-go/kubernetes"
)

// DeployCodewind takes in a `codewind` object and deploys Codewind and the performance dashboard into the specified namespace
func DeployCodewind(clientset *kubernetes.Clientset, codewind Codewind, namespace string) error {
	// Deploy Codewind PFE
	service := createPFEService(codewind)
	deploy := createPFEDeploy(codewind)

	log.Infoln("Deploying Codewind...")
	_, err := clientset.CoreV1().Services(namespace).Create(&service)
	if err != nil {
		log.Errorf("Unable to create Codewind service: %v\n", err)
		return err
	}
	_, err = clientset.AppsV1().Deployments(namespace).Create(&deploy)
	if err != nil {
		log.Errorf("Unable to create Codewind deployment: %v\n", err)
		return err
	}

	// Deploy the Performance dashboard
	performanceService := createPerformanceService(codewind)
	performanceDeploy := createPerformanceDeploy(codewind)

	log.Infoln("Deploying Codewind Performance Dashboard...")
	_, err = clientset.CoreV1().Services(namespace).Create(&performanceService)
	if err != nil {
		log.Errorf("Error: Unable to create Codewind Performance service: %v\n", err)
		return err
	}
	_, err = clientset.AppsV1().Deployments(namespace).Create(&performanceDeploy)
	if err != nil {
		log.Errorf("Error: Unable to create Codewind Performance deployment: %v\n", err)
		return err
	}
	return nil
}

// createPFEDeploy creates a Kubernetes deploy for Codewind, marking the Che workspace as its owner
func createPFEDeploy(codewind Codewind) appsv1.Deployment {
	labels := map[string]string{
		"app":          "codewind-pfe",
		"pfeWorkspace": codewind.WorkspaceID,
	}

	volumes, volumeMounts := setPFEVolumes(codewind)
	envVars := setPFEEnvVars(codewind)

	return generateDeployment(codewind, PFEPrefix, PFEImage+":"+PFEImageTag, PFEContainerPort, volumes, volumeMounts, envVars, labels)
}

// createPFEService creates a Kubernetes service for Codewind, exposing port 9191
func createPFEService(codewind Codewind) corev1.Service {
	labels := map[string]string{
		"app":          "codewind-pfe",
		"pfeWorkspace": codewind.WorkspaceID,
	}
	return generateService(codewind, PFEPrefix, PFEContainerPort, labels)
}

func createPerformanceDeploy(codewind Codewind) appsv1.Deployment {
	labels := map[string]string{
		"app":                  PerformancePrefix,
		"performanceWorkspace": codewind.WorkspaceID,
	}

	volumes := []corev1.Volume{}
	volumeMounts := []corev1.VolumeMount{}
	envVars := setPerformanceEnvVars(codewind)
	return generateDeployment(codewind, PerformancePrefix, PerformanceImage+":"+PerformanceTag, PerformanceContainerPort, volumes, volumeMounts, envVars, labels)
}

func createPerformanceService(codewind Codewind) corev1.Service {
	labels := map[string]string{
		"app":                  PerformancePrefix,
		"performanceWorkspace": codewind.WorkspaceID,
	}
	return generateService(codewind, PerformancePrefix, PerformanceContainerPort, labels)

}
