package codewind

import (
	log "github.com/sirupsen/logrus"

	"deploy-pfe/pkg/constants"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
)

// DeployCodewind takes in a `codewind` object and deploys Codewind and the performance dashboard into the specified namespace
func DeployCodewind(clientset *kubernetes.Clientset, codewind Codewind, namespace string) error {
	// Create a PVC for PFE
	// Determine if we're running on OpenShift on IKS (and thus ned to use the ibm-file-bronze storage class)
	storageClass := ""
	sc, err := clientset.StorageV1().StorageClasses().Get(constants.ROKSStorageClass, metav1.GetOptions{})
	if err == nil && sc != nil {
		storageClass = sc.Name
		log.Infof("Setting storage class to %s\n", storageClass)
	}

	pvc := generatePVC(codewind, constants.PFEVolumeSize, storageClass)
	_, err = clientset.CoreV1().PersistentVolumeClaims(namespace).Create(&pvc)
	if err != nil {
		log.Errorf("Unable to create Persistent Volume Claim for PFE: %v\n", err)
		return err
	}

	// Deploy Codewind PFE
	service := createPFEService(codewind)
	deploy := createPFEDeploy(codewind)

	log.Infoln("Deploying Codewind...")
	_, err = clientset.CoreV1().Services(namespace).Create(&service)
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
		"app":               "codewind-pfe",
		"codewindWorkspace": codewind.WorkspaceID,
	}

	volumes, volumeMounts := setPFEVolumes(codewind)
	envVars := setPFEEnvVars(codewind)

	return generateDeployment(codewind, constants.PFEPrefix, codewind.PFEImage, constants.PFEContainerPort, volumes, volumeMounts, envVars, labels)
}

// createPFEService creates a Kubernetes service for Codewind, exposing port 9191
func createPFEService(codewind Codewind) corev1.Service {
	labels := map[string]string{
		"app":               "codewind-pfe",
		"codewindWorkspace": codewind.WorkspaceID,
	}
	return generateService(codewind, constants.PFEPrefix, constants.PFEContainerPort, labels)
}

func createPerformanceDeploy(codewind Codewind) appsv1.Deployment {
	labels := map[string]string{
		"app":               constants.PerformancePrefix,
		"codewindWorkspace": codewind.WorkspaceID,
	}

	volumes := []corev1.Volume{}
	volumeMounts := []corev1.VolumeMount{}
	envVars := setPerformanceEnvVars(codewind)
	return generateDeployment(codewind, constants.PerformancePrefix, codewind.PerformanceImage, constants.PerformanceContainerPort, volumes, volumeMounts, envVars, labels)
}

func createPerformanceService(codewind Codewind) corev1.Service {
	labels := map[string]string{
		"app":               constants.PerformancePrefix,
		"codewindWorkspace": codewind.WorkspaceID,
	}
	return generateService(codewind, constants.PerformancePrefix, constants.PerformanceContainerPort, labels)

}
