package che

import (
	"fmt"
	"net/url"
	"os"

	"deploy-pfe/pkg/kube"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

// GetWorkspacePVC retrieves a PVC (Persistent Volume Claim) associated with the Che workspace we're deploying Codewind in
func GetWorkspacePVC(clientset *kubernetes.Clientset, namespace string, cheWorkspaceID string) *corev1.PersistentVolumeClaim {
	PVCs, err := clientset.CoreV1().PersistentVolumeClaims(namespace).List(metav1.ListOptions{
		LabelSelector: "che.workspace_id=" + cheWorkspaceID,
	})
	if err != nil || PVCs == nil {
		log.Errorf("Error, unable to retrieve PVCs: %v\n", err)
		os.Exit(1)
	} else if len(PVCs.Items) < 1 {
		// We couldn't find the workspace PVC, so need to find an alternative.
		PVC, err := clientset.CoreV1().PersistentVolumeClaims(namespace).Get("claim-che-workspace", metav1.GetOptions{})
		if err != nil || PVC == nil {
			log.Errorf("Error, unable to retrieve PVCs: %v\n", err)
			os.Exit(1)
		} else {
			return PVC
		}
	}
	return &PVCs.Items[0]
}

// GetWorkspaceServiceAccount retrieves the Service Account associated with the Che workspace we're deploying Codewind in
func GetWorkspaceServiceAccount(clientset *kubernetes.Clientset, namespace string, cheWorkspaceID string) string {
	var serviceAccountName string

	// Retrieve the workspace service account labeled with the Che Workspace ID
	workspacePod, err := clientset.CoreV1().Pods(namespace).List(metav1.ListOptions{
		LabelSelector: "che.original_name=che-workspace-pod,che.workspace_id=" + cheWorkspaceID,
	})
	if err != nil || workspacePod == nil {
		log.Errorf("Error retrieving the Che workspace pod %v\n", err)
		os.Exit(1)
	} else if len(workspacePod.Items) != 1 {
		// Default to che-workspace as the Service Account name if one couldn't be found
		serviceAccountName = "che-workspace"
	} else {
		serviceAccountName = workspacePod.Items[0].Spec.ServiceAccountName
	}

	return serviceAccountName

}

// GetWorkspaceRegistrySecret retrieves the Kubernetes ImagePullSecret associated with the Che workspace we're deploying Codewind in
func GetWorkspaceRegistrySecret(clientset *kubernetes.Clientset, namespace string, cheWorkspaceID string) string {
	var secretName string
	// Retrieve the secret tagged with the workspace ID label
	// If the secret is missing, fall back on a default value
	registrySecret, err := clientset.CoreV1().Secrets(namespace).List(metav1.ListOptions{
		LabelSelector: "che.workspace_id=" + cheWorkspaceID,
	})
	if err != nil {
		log.Errorf("Error retrieving the list of secrets: %v\n", err)
		os.Exit(1)
	} else if len(registrySecret.Items) != 1 {
		secretName = cheWorkspaceID + "-private-registries"
	} else {
		secretName = registrySecret.Items[0].GetName()
	}
	return secretName
}

// GetOwnerReferences retrieves the owner reference name and UID, allowing us to tie any Codewind resources to the Che workspace
// Enabling the Kubernetes garbage collector clean everything up when the workspace is deleted
func GetOwnerReferences(clientset *kubernetes.Clientset, namespace string, cheWorkspaceID string) (string, types.UID) {
	// Get the Workspace pod
	var ownerReferenceName string
	var ownerReferenceUID types.UID

	workspacePod, err := clientset.CoreV1().Pods(namespace).List(metav1.ListOptions{
		LabelSelector: "che.original_name=che-workspace-pod,che.workspace_id=" + cheWorkspaceID,
	})
	if err != nil {
		log.Errorf("Error: Unable to retrieve the workspace pod %v\n", err)
		os.Exit(1)
	}
	// Retrieve the owner reference name and UID from the workspace pod. This will allow Codewind to be garbage collected by Kube
	ownerReferenceName = workspacePod.Items[0].GetOwnerReferences()[0].Name
	ownerReferenceUID = workspacePod.Items[0].GetOwnerReferences()[0].UID

	return ownerReferenceName, ownerReferenceUID
}

// GetCheIngress parses the Che ingress domain from the Che API URL that was passed in
func GetCheIngress(cheAPI string) (string, error) {
	// Log an error and return if a blank string was passed in
	if cheAPI == "" {
		return "", fmt.Errorf("Che API URL was not set")
	}

	cheURL, err := url.Parse(cheAPI)
	if err != nil {
		return "", fmt.Errorf("unable to parse the Che API URL")
	}

	parsedURL := cheURL.Hostname()
	if parsedURL == "" {
		return "", fmt.Errorf("parsed Che API URL is empty")
	}

	// Return the hostname of the Che API URL. This will have the http/https and path stripped out
	return parsedURL, nil

}

// GetPFEService returns the service name for the specified workspace ID
func GetPFEService(clientset *kubernetes.Clientset, namespace string, workspaceID string) string {
	service, err := clientset.CoreV1().Services(kube.GetCurrentNamespace()).List(metav1.ListOptions{
		LabelSelector: "app=codewind-pfe,codewindWorkspace=" + workspaceID,
	})
	if err != nil || len(service.Items) < 1 {
		return ""
	}
	return service.Items[0].GetName()
}
