package codewind

import (
	"deploy-pfe/pkg/constants"
	"fmt"
	"testing"
)

func setupCodewind() Codewind {
	cheWorkspaceID := "workspace1erok6723m74axkg"

	return Codewind{
		PFEName:            constants.PFEPrefix + cheWorkspaceID,
		PFEImage:           constants.PFEImage + ":" + constants.PFEImageTag,
		PerformanceName:    constants.PerformancePrefix + cheWorkspaceID,
		PerformanceImage:   constants.PerformanceImage + ":" + constants.PerformanceTag,
		Namespace:          "default",
		WorkspaceID:        cheWorkspaceID,
		ServiceAccountName: "che-workspace",
		PullSecret:         "workspace1erok6723m74axkg-registry-secrets",
		OwnerReferenceName: "codewind",
		OwnerReferenceUID:  "c22d4a29-ba20-11e9-ac2a-005056a04e5e",
		Privileged:         true,
		Ingress:            constants.PFEPrefix + "-" + cheWorkspaceID + "-" + "che.1.2.3.4.nip.io",
	}
}
func TestCreatePFEDeployment(t *testing.T) {
	// Create a test codewind instance
	codewindInstance := setupCodewind()
	tests := []struct {
		name     string
		codewind Codewind
	}{
		{
			name:     fmt.Sprintf("Verify Deployment object creation"),
			codewind: codewindInstance,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deploy := createPFEDeploy(tt.codewind)

			// Verify proper deployment name
			if deploy.GetName() != constants.PFEPrefix+"-"+tt.codewind.WorkspaceID {
				t.Error("PFE deployment name not properly set")
			}

			// Verify the owner is properly set
			if len(deploy.GetOwnerReferences()) != 1 {
				t.Errorf("PFE deployment does not have its owner references set")
			}
			ownerReference := deploy.GetOwnerReferences()[0]
			if ownerReference.Name != tt.codewind.OwnerReferenceName {
				t.Errorf("PFE deployment's owner reference name not properly set. Had %v, expected %v", ownerReference.Name, tt.codewind.OwnerReferenceName)
			}
			if ownerReference.UID != tt.codewind.OwnerReferenceUID {
				t.Errorf("PFE deployment's owner reference UID not properly set. Had %v, expected %v", ownerReference.UID, tt.codewind.OwnerReferenceUID)
			}

			// Verify pod labels
			pod := deploy.Spec.Template
			labels := pod.GetObjectMeta().GetLabels()
			if labels["app"] != "codewind-pfe" || labels["codewindWorkspace"] != tt.codewind.WorkspaceID {
				t.Error("PFE deploymeny labels improperly set")
			}

			// Verify only one image is used
			deployContainers := pod.Spec.Containers
			if len(deployContainers) != 1 {
				t.Errorf("PFE deployment had %v containers, expected %v", len(deployContainers), 1)
			}
			pfeContainer := deployContainers[0]
			pfeImage := tt.codewind.PFEImage

			// Verify proper PFE image is being used
			if pfeContainer.Image != pfeImage {
				t.Errorf("PFE container using invalid image, had %v, expected %v", pfeContainer.Image, pfeImage)
			}
		})
	}
}

func TestCreatePFEService(t *testing.T) {
	// Create a test codewind instance
	codewindInstance := setupCodewind()
	tests := []struct {
		name     string
		codewind Codewind
	}{
		{
			name:     fmt.Sprintf("Verify PFE service object creation"),
			codewind: codewindInstance,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := createPFEService(tt.codewind)
			deploy := createPFEDeploy(tt.codewind)

			// Verify proper deployment name
			if service.GetName() != constants.PFEPrefix+"-"+tt.codewind.WorkspaceID {
				t.Error("PFE service name not properly set")
			}

			// Verify the owner is properly set
			if len(service.GetOwnerReferences()) != 1 {
				t.Errorf("PFE service does not have its owner references set")
			}
			ownerReference := service.GetOwnerReferences()[0]
			if ownerReference.Name != tt.codewind.OwnerReferenceName {
				t.Errorf("PFE service's owner reference name not properly set. Had %v, expected %v", ownerReference.Name, tt.codewind.OwnerReferenceName)
			}
			if ownerReference.UID != tt.codewind.OwnerReferenceUID {
				t.Errorf("PFE service's owner reference UID not properly set. Had %v, expected %v", ownerReference.UID, tt.codewind.OwnerReferenceUID)
			}

			// Verify label selctor matches the labels on the PFE pod
			serviceSelector := service.Spec.Selector
			pfeLabels := deploy.Spec.Template.Labels
			if serviceSelector["app"] != pfeLabels["app"] || serviceSelector["codewindWorkspace"] != pfeLabels["codewindWorkspace"] {
				t.Errorf("PFE service selector labels and pod labels don't match.")
			}
			// Verify that the proper PFE port is exposed
			servicePorts := service.Spec.Ports
			if len(servicePorts) != 1 {
				t.Errorf("PFE service exposing wrong number of ports. Had %v, expected %v", len(servicePorts), 1)
			}
			if servicePorts[0].Port != constants.PFEContainerPort {
				t.Errorf("PFE service port not properly exposed. Port %v exposed instead of %v", servicePorts[0].Port, constants.PFEContainerPort)
			}
		})
	}
}

func TestCreatePerformanceDeployment(t *testing.T) {
	// Create a test codewind instance
	codewindInstance := setupCodewind()
	tests := []struct {
		name     string
		codewind Codewind
	}{
		{
			name:     fmt.Sprintf("Verify Performance Deployment object creation"),
			codewind: codewindInstance,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deploy := createPerformanceDeploy(tt.codewind)

			// Verify proper deployment name
			if deploy.GetName() != constants.PerformancePrefix+"-"+tt.codewind.WorkspaceID {
				t.Error("Performance deployment name not properly set")
			}

			// Verify the owner is properly set
			if len(deploy.GetOwnerReferences()) != 1 {
				t.Errorf("Performance deployment does not have its owner references set")
			}
			ownerReference := deploy.GetOwnerReferences()[0]
			if ownerReference.Name != tt.codewind.OwnerReferenceName {
				t.Errorf("Performance deployment's owner reference name not properly set. Had %v, expected %v", ownerReference.Name, tt.codewind.OwnerReferenceName)
			}
			if ownerReference.UID != tt.codewind.OwnerReferenceUID {
				t.Errorf("Performance deployment's owner reference UID not properly set. Had %v, expected %v", ownerReference.UID, tt.codewind.OwnerReferenceUID)
			}

			// Verify pod labels
			pod := deploy.Spec.Template
			labels := pod.GetObjectMeta().GetLabels()
			if labels["app"] != "codewind-performance" || labels["codewindWorkspace"] != tt.codewind.WorkspaceID {
				t.Error("Performance deployment labels improperly set")
			}

			// Verify only one image is used
			deployContainers := pod.Spec.Containers
			if len(deployContainers) != 1 {
				t.Errorf("Performance deployment had %v containers, expected %v", len(deployContainers), 1)
			}
			perfContainer := deployContainers[0]
			pfeImage := tt.codewind.PerformanceImage

			// Verify proper PFE image is being used
			if perfContainer.Image != pfeImage {
				t.Errorf("Performance container using invalid image, had %v, expected %v", perfContainer.Image, pfeImage)
			}
		})
	}
}

func TestCreatePerformanceService(t *testing.T) {
	// Create a test codewind instance
	codewindInstance := setupCodewind()
	tests := []struct {
		name     string
		codewind Codewind
	}{
		{
			name:     fmt.Sprintf("Verify PFE service object creation"),
			codewind: codewindInstance,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := createPerformanceService(tt.codewind)
			deploy := createPerformanceDeploy(tt.codewind)

			// Verify proper deployment name
			if service.GetName() != constants.PerformancePrefix+"-"+tt.codewind.WorkspaceID {
				t.Error("PFE service name not properly set")
			}

			// Verify the owner is properly set
			if len(service.GetOwnerReferences()) != 1 {
				t.Errorf("PFE service does not have its owner references set")
			}
			ownerReference := service.GetOwnerReferences()[0]
			if ownerReference.Name != tt.codewind.OwnerReferenceName {
				t.Errorf("PFE service's owner reference name not properly set. Had %v, expected %v", ownerReference.Name, tt.codewind.OwnerReferenceName)
			}
			if ownerReference.UID != tt.codewind.OwnerReferenceUID {
				t.Errorf("PFE service's owner reference UID not properly set. Had %v, expected %v", ownerReference.UID, tt.codewind.OwnerReferenceUID)
			}

			// Verify label selctor matches the labels on the PFE pod
			serviceSelector := service.Spec.Selector
			pfeLabels := deploy.Spec.Template.Labels
			if serviceSelector["app"] != pfeLabels["app"] || serviceSelector["codewindWorkspace"] != pfeLabels["codewindWorkspace"] {
				t.Errorf("PFE service selector labels and pod labels don't match.")
			}
			// Verify that the proper PFE port is exposed
			servicePorts := service.Spec.Ports
			if len(servicePorts) != 1 {
				t.Errorf("PFE service exposing wrong number of ports. Had %v, expected %v", len(servicePorts), 1)
			}
			if servicePorts[0].Port != int32(constants.PerformanceContainerPort) {
				t.Errorf("PFE service port not properly exposed. Port %v exposed instead of %v", servicePorts[0].Port, constants.PerformanceContainerPort)
			}
		})
	}
}

// TestVerifyPFEEnvVars verifies that the environment variables passed into Codewind-PFE have the proper values
func TestVerifyPFEEnvVars(t *testing.T) {
	codewindInstance := setupCodewind()
	tests := []struct {
		name     string
		codewind Codewind
	}{
		{
			name:     fmt.Sprintf("Verify PFE Environment Variables"),
			codewind: codewindInstance,
		},
	}
	for _, tt := range tests {
		pfeDeploy := createPFEDeploy(tt.codewind)
		pfeService := createPFEService(tt.codewind)

		performanceService := createPerformanceService(tt.codewind)

		pfeEnvVars := pfeDeploy.Spec.Template.Spec.Containers[0].Env

		for _, env := range pfeEnvVars {
			// Verify that KUBE_NAMESPACE and TILLER_NAMESPACE matches the namespace that codewind is deployed in
			if env.Name == "KUBE_NAMESPACE" {
				if env.Value != pfeDeploy.GetNamespace() {
					t.Errorf("KUBE_NAMESPACE doesn't match the namespace that Codewind is deployed in %v\n", pfeDeploy.GetNamespace())
				}
				continue
			}
			if env.Name == "TILLER_NAMESPACE" {
				if env.Value != pfeDeploy.GetNamespace() {
					t.Errorf("TILLER_NAMESPACE doesn't match Performance Dashboard service name: %v\n", pfeDeploy.GetNamespace())
				}
				continue
			}

			// Verify that SERVICE_ACCOUNT_NAME matches the service account that Codewind-PFE is running in
			if env.Name == "SERVICE_ACCOUNT_NAME" {
				if env.Value != pfeDeploy.Spec.Template.Spec.ServiceAccountName {
					t.Errorf("SERVICE_ACCOUNT_NAME doesn't match Codewind-PFE service account name: %v\n", pfeDeploy.Spec.Template.Spec.ServiceAccountName)
				}
				continue
			}

			// Verify that SERVICE_NAME matches the Codewind-PFE service name
			if env.Name == "SERVICE_NAME" {
				if env.Value != pfeService.GetName() {
					t.Errorf("SERVICE_NAME doesn't match PFE service name: %v\n", pfeService.GetName())
				}
				continue
			}

			// Verify that the Performance dashboard service name matches the value passed into the Codewind-PFE container
			if env.Name == "CODEWIND_PERFORMANCE_SERVICE" {
				if env.Value != performanceService.GetName() {
					t.Errorf("CODEWIND_PERFORMANCE_SERVICE doesn't match Performance Dashboard service name: %v\n", performanceService.GetName())
				}
				continue
			}

		}
	}
}
