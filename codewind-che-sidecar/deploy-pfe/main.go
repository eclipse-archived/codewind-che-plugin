package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	// Get the Kube config and clientsets
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	// Get the current namespace
	namespace := getCurrentNamespace()
	fmt.Println("*** Namespace: " + namespace)

	// Get the Che workspace ID
	cheWorkspaceID := os.Getenv("CHE_WORKSPACE_ID")
	if cheWorkspaceID == "" {
		log.Fatal("Che Workspace ID not set and unable to deploy PFE, exiting...")
	}
	log.Printf("*** Che Workspace ID: %s\n", cheWorkspaceID)

	workspacePVC := getWorkspacePVC(clientset, namespace, cheWorkspaceID)
	log.Printf("*** PVC: %s\n", workspacePVC)

	// Get the Che workspace service account to use with Codewind
	serviceAccountName := getWorkspaceServiceAccount(clientset, namespace, cheWorkspaceID)
	log.Printf("*** Service Account: %s\n", serviceAccountName)

	// Get the name of the secret containing the workspace's registry secrets
	secretName := getWorkspaceRegistrySecret(clientset, namespace, cheWorkspaceID)
	log.Printf("*** Secret: %s\n", secretName)

	// Get the Owner reference name and uid
	ownerReferenceName, ownerReferenceUID := getOwnerReferences(clientset, namespace, cheWorkspaceID)

	// Create the Codewind deployment object
	codewind := CodewindDeployment{
		Name:               CodewindPrefix + cheWorkspaceID,
		Namespace:          namespace,
		WorkspaceID:        cheWorkspaceID,
		PVCName:            workspacePVC,
		ServiceAccountName: serviceAccountName,
		PullSecret:         secretName,
		OwnerReferenceName: ownerReferenceName,
		OwnerReferenceUID:  ownerReferenceUID,
		Privileged:         true,
	}

	// Patch the Che workspace service account
	err = patchServiceAccount(clientset, codewind)
	fmt.Println(err)

	// Deploy Codewind
	service := createPFEService(codewind)
	deploy := createPFEDeploy(codewind)

	log.Println("Creating Codewind...")
	_, err = clientset.CoreV1().Services(namespace).Create(&service)
	if err != nil {
		log.Fatalf("Error: Unable to create Codewind service: %v\n", err)
	}
	_, err = clientset.AppsV1().Deployments(namespace).Create(&deploy)
	if err != nil {
		log.Fatalf("Error: Unable to create Codewind deployment: %v\n", err)
	}
}

func getKubeClientConfig() clientcmd.ClientConfig {
	// Instantiate loader for kubeconfig file.
	kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)
	return kubeconfig
}

// GetCurrentNamespace gets the current namespace in the Kubernetes context
func getCurrentNamespace() string {
	// Instantiate loader for kubeconfig file.
	kubeconfig := getKubeClientConfig()
	namespace, _, err := kubeconfig.Namespace()
	if err != nil {
		panic(err)
	}
	return namespace
}

// PatchServiceAccount takes in a list of secret names, and patches it to the specified service account
func patchServiceAccount(clientset *kubernetes.Clientset, codewind CodewindDeployment) error {
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

func getWorkspacePVC(clientset *kubernetes.Clientset, namespace string, cheWorkspaceID string) string {
	var pvcName string

	PVCs, err := clientset.CoreV1().PersistentVolumeClaims(namespace).List(metav1.ListOptions{
		LabelSelector: "che.workspace.volume_name=projects,che.workspace_id=" + cheWorkspaceID,
	})
	if err != nil || PVCs == nil {
		log.Fatal(err)
	} else if len(PVCs.Items) < 1 {
		// We couldn't find the workspace PVC, so need to find an alternative.
		PVCs, err = clientset.CoreV1().PersistentVolumeClaims(namespace).List(metav1.ListOptions{
			LabelSelector: "che.workspace_id=" + cheWorkspaceID,
		})
		if err != nil || PVCs == nil {
			log.Fatal(err)
		} else if len(PVCs.Items) < 1 {
			pvcName = "claim-che-workspace"
		} else {
			pvcName = PVCs.Items[0].GetName()
		}
	} else {
		pvcName = PVCs.Items[0].GetName()
	}

	return pvcName
}

func getWorkspaceServiceAccount(clientset *kubernetes.Clientset, namespace string, cheWorkspaceID string) string {
	var serviceAccountName string

	workspacePod, err := clientset.CoreV1().Pods(namespace).List(metav1.ListOptions{
		LabelSelector: "che.original_name=che-workspace-pod,che.workspace_id=" + cheWorkspaceID,
	})
	if err != nil || workspacePod == nil {
		log.Fatalf("Error retrieving the Che workspace pod %v\n", err)
	} else if len(workspacePod.Items) < 1 {
		// Default to che-workspace as the Service Account name
		serviceAccountName = "che-workspace"
	} else {
		serviceAccountName = workspacePod.Items[0].Spec.ServiceAccountName
	}

	return serviceAccountName

}

func getWorkspaceRegistrySecret(clientset *kubernetes.Clientset, namespace string, cheWorkspaceID string) string {
	var secretName string
	// Retrieve the secret tagged with the workspace ID label
	// If the secret is missing, fall back on a default value
	registrySecret, err := clientset.CoreV1().Secrets(namespace).List(metav1.ListOptions{
		LabelSelector: "che.workspace_id=" + cheWorkspaceID,
	})
	if err != nil {
		log.Fatalf("Error retrieving the list of secrets: %v\n", err)
	} else if len(registrySecret.Items) < 1 {
		secretName = cheWorkspaceID + "-private-registries"
	} else {
		secretName = registrySecret.Items[0].GetName()
	}
	return secretName
}

func getOwnerReferences(clientset *kubernetes.Clientset, namespace string, cheWorkspaceID string) (string, types.UID) {
	// Get the Workspace pod
	var ownerReferenceName string
	var ownerReferenceUID types.UID

	workspacePod, err := clientset.CoreV1().Pods(namespace).List(metav1.ListOptions{
		LabelSelector: "che.original_name=che-workspace-pod,che.workspace_id=" + cheWorkspaceID,
	})
	if err != nil {
		log.Fatalf("Error: Unable to retrieve the workspace pod %v\n", err)
	}
	// Retrieve the owner reference name and UID from the workspace pod. This will allow Codewind to be garbage collected by Kube
	ownerReferenceName = workspacePod.Items[0].GetOwnerReferences()[0].Name
	ownerReferenceUID = workspacePod.Items[0].GetOwnerReferences()[0].UID

	return ownerReferenceName, ownerReferenceUID
}
