package che

import (
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// GetKubeClientConfig retrieves the Kubernetes client config from the cluster
func GetKubeClientConfig() clientcmd.ClientConfig {
	// Instantiate loader for kubeconfig file.
	kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)
	return kubeconfig
}

// GetCurrentNamespace gets the current namespace in the Kubernetes context
func GetCurrentNamespace() string {
	// Instantiate loader for kubeconfig file.
	kubeconfig := GetKubeClientConfig()
	namespace, _, err := kubeconfig.Namespace()
	if err != nil {
		panic(err)
	}
	return namespace
}

// GetWorkspacePVC retrieves the PVC (Persistent Volume Claim) associated with the Che workspace we're deploying Codewind in
func GetWorkspacePVC(clientset *kubernetes.Clientset, namespace string, cheWorkspaceID string) string {
	var pvcName string

	PVCs, err := clientset.CoreV1().PersistentVolumeClaims(namespace).List(metav1.ListOptions{
		LabelSelector: "che.workspace.volume_name=projects,che.workspace_id=" + cheWorkspaceID,
	})
	if err != nil || PVCs == nil {
		log.Errorf("Error, unable to retrieve PVCs: %v\n", err)
		os.Exit(1)
	} else if len(PVCs.Items) < 1 {
		// We couldn't find the workspace PVC, so need to find an alternative.
		PVCs, err = clientset.CoreV1().PersistentVolumeClaims(namespace).List(metav1.ListOptions{
			LabelSelector: "che.workspace_id=" + cheWorkspaceID,
		})
		if err != nil || PVCs == nil {
			log.Errorf("Error, unable to retrieve PVCs: %v\n", err)
			os.Exit(1)
		} else if len(PVCs.Items) != 1 {
			pvcName = "claim-che-workspace"
		} else {
			pvcName = PVCs.Items[0].GetName()
		}
	} else {
		pvcName = PVCs.Items[0].GetName()
	}

	return pvcName
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

// GetCheIngress parses the Che ingress domain from the `CHE_API` environment variable
func GetCheIngress() string {
	cheAPI := os.Getenv("CHE_API")
	if cheAPI == "" {
		log.Errorf("Che Workspace ID not set and unable to deploy PFE, exiting...")
		os.Exit(1)
	}
	cheIngress := strings.TrimLeft(strings.TrimRight(cheAPI, "/api"), "http://")
	return cheIngress

}

// DetectOpenShift3 determines if we're running on an OpenShift 3.x cluster
// From https://github.com/eclipse/che-operator/blob/2f639261d8b5416b2934591e12925ee0935814dd/pkg/util/util.go#L63
func DetectOpenShift3(config *rest.Config) bool {
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		log.Errorf("Unable to detect if running on OpenShift: %v\n", err)
		os.Exit(1)
	}
	apiList, err := discoveryClient.ServerGroups()
	if err != nil {
		log.Errorf("Error attempting to retrieve list of API Groups: %v\n", err)
		os.Exit(1)
	}
	apiGroups := apiList.Groups
	for _, group := range apiGroups {
		if group.Name == "route.openshift.io" {
			return true
		}
	}
	return false
}

// GetPFEService returns the service name for the specified workspace ID
func GetPFEService(clientset *kubernetes.Clientset, namespace string, workspaceID string) string {
	service, err := clientset.CoreV1().Services(GetCurrentNamespace()).List(metav1.ListOptions{
		LabelSelector: "pfeWorkspace=" + workspaceID,
	})
	if err != nil || len(service.Items) < 1 {
		return ""
	}
	return service.Items[0].GetName()
}
