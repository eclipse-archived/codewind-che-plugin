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

// RedeployCodewind redeploys the Codewind PFE and Performance instances
func RedeployCodewind(clientset *kubernetes.Clientset, codewind Codewind) error {
	// Redeploy PFE
	log.Infoln("Re-Deploying Codewind...")
	err := deleteDeployment(clientset, codewind.PFEImage, codewind.Namespace, "app=codewind-pfe,codewindWorkspace="+codewind.WorkspaceID)
	if err != nil {
		return err
	}
	pfeDeploy := createPFEDeploy(codewind)
	_, err = clientset.AppsV1().Deployments(codewind.Namespace).Create(&pfeDeploy)
	if err != nil {
		log.Errorf("Unable to create Codewind deployment: %v\n", err)
		return err
	}

	// Update performance image
	err = deleteDeployment(clientset, codewind.PerformanceImage, codewind.Namespace, "app=codewind-performance,codewindWorkspace="+codewind.WorkspaceID)
	if err != nil {
		return err
	}
	performanceDeploy := createPerformanceDeploy(codewind)
	_, err = clientset.AppsV1().Deployments(codewind.Namespace).Create(&performanceDeploy)
	if err != nil {
		log.Errorf("Unable to create Codewind Performance Dashboard deployment: %v\n", err)
		return err
	}
	return nil
}

func deleteDeployment(clientset *kubernetes.Clientset, image string, namespace string, selector string) error {
	deploymentList, err := clientset.AppsV1().Deployments(namespace).List(metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil || deploymentList == nil || len(deploymentList.Items) != 1 {
		return err
	}
	dep := deploymentList.Items[0]

	// Delete the deployment
	gracePeriod := int64(0)
	deletePolicy := metav1.DeletePropagationForeground
	err = clientset.AppsV1().Deployments(namespace).Delete(dep.GetName(), &metav1.DeleteOptions{
		GracePeriodSeconds: &gracePeriod,
		PropagationPolicy:  &deletePolicy,
	})
	return err
}
