package main

import (
	"fmt"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"deploy-pfe/pkg/che"
	"deploy-pfe/pkg/codewind"
	"deploy-pfe/pkg/constants"
	"deploy-pfe/pkg/kube"

	routev1 "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	// Get the Kube config and clientsets
	config, err := rest.InClusterConfig()
	if err != nil {
		// Couldn't find an InClusterConfig, may be running outside of Kube, so try to find a local kube config file
		kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			log.Errorf("Unable to retrieve Kubernetes InClusterConfig %v\n", err)
			os.Exit(1)
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Errorf("Unable to retrieve Kubernetes clientset %v\n", err)
		os.Exit(1)
	}

	// Get the current namespace
	namespace := kube.GetCurrentNamespace()

	// Get the Che workspace ID
	cheWorkspaceID := os.Getenv("CHE_WORKSPACE_ID")
	if cheWorkspaceID == "" {
		log.Errorln("Che Workspace ID not set and unable to deploy PFE, exiting...")
	}

	// If deploy-pfe was called with the `get-service` arg, retrieve the codewind service name if it exists, and exit
	if len(os.Args) > 1 {
		if os.Args[1] == "get-service" {
			fmt.Println(che.GetPFEService(clientset, namespace, cheWorkspaceID))
			return
		}
	}

	// Retrieve the PVC that's used for the workspace projects
	workspacePVC := che.GetWorkspacePVC(clientset, namespace, cheWorkspaceID)
	log.Infof("PVC: %s\n", workspacePVC)

	// Get the ingress domain used for Che (and Che workspaces)
	cheIngress := che.GetCheIngress()
	log.Infof("Ingress: %s\n", cheIngress)

	// Get the Che workspace service account to use with Codewind
	serviceAccountName := che.GetWorkspaceServiceAccount(clientset, namespace, cheWorkspaceID)
	log.Infof("Service Account: %s\n", serviceAccountName)

	// Get the name of the secret containing the workspace's registry secrets
	secretName := che.GetWorkspaceRegistrySecret(clientset, namespace, cheWorkspaceID)
	log.Infof("Registry Secret: %s\n", secretName)

	// Get the Owner reference name and uid
	ownerReferenceName, ownerReferenceUID := che.GetOwnerReferences(clientset, namespace, cheWorkspaceID)

	// Retrieve the images for PFE and Performance dashboard
	pfe, performance := codewind.GetImages()

	// Create the Codewind deployment object
	codewindInstance := codewind.Codewind{
		PFEName:            constants.PFEPrefix + cheWorkspaceID,
		PFEImage:           pfe,
		PerformanceName:    constants.PerformancePrefix + cheWorkspaceID,
		PerformanceImage:   performance,
		Namespace:          namespace,
		WorkspaceID:        cheWorkspaceID,
		PVCName:            workspacePVC,
		ServiceAccountName: serviceAccountName,
		PullSecret:         secretName,
		OwnerReferenceName: ownerReferenceName,
		OwnerReferenceUID:  ownerReferenceUID,
		Privileged:         true,
		Ingress:            constants.PFEPrefix + "-" + cheWorkspaceID + "-" + cheIngress,
	}

	// Patch the Che workspace service account
	err = codewind.PatchServiceAccount(clientset, codewindInstance)
	if err != nil {
		log.Errorf("Error: Unable to patch Che workspace service account: %v\n", err)
		os.Exit(1)
	}

	err = codewind.DeployCodewind(clientset, codewindInstance, namespace)
	if err != nil {
		log.Errorf("Codewind deployment failed, exiting...")
		os.Exit(1)
	}

	// Expose Codewind over an ingress or route
	isOpenShift := kube.DetectOpenShift3(config)
	if isOpenShift {
		// Deploy a route instead on OpenShift 3.x
		route := codewind.CreateRoute(codewindInstance)
		routev1client, err := routev1.NewForConfig(config)
		if err != nil {
			log.Errorf("Error retrieving route client for OpenShift: %v\n", err)
			os.Exit(1)
		}

		_, err = routev1client.Routes(namespace).Create(&route)
		if err != nil {
			log.Errorf("Error: Unable to create route for Codewind: %v\n", err)
			os.Exit(1)
		}

	} else {
		ingress := codewind.CreateIngress(codewindInstance)

		_, err = clientset.ExtensionsV1beta1().Ingresses(namespace).Create(&ingress)
		if err != nil {
			log.Errorf("Error: Unable to create ingress for Codewind: %v\n", err)
			os.Exit(1)
		}

	}

}
