package codewind

import (
	"fmt"
	"testing"
)

func setup_codewind() Codewind {
	cheWorkspaceID := "workspace1erok6723m74axkg"

	return Codewind{
		PFEName:            PFEPrefix + cheWorkspaceID,
		PerformanceName:    PerformancePrefix + cheWorkspaceID,
		Namespace:          "default",
		WorkspaceID:        cheWorkspaceID,
		PVCName:            "claim-che-workspace",
		ServiceAccountName: "che-workspace",
		PullSecret:         "workspace1erok6723m74axkg-registry-secrets",
		OwnerReferenceName: "test",
		OwnerReferenceUID:  "test",
		Privileged:         true,
		Ingress:            PFEPrefix + "-" + cheWorkspaceID + "-" + "che.1.2.3.4.nip.io",
	}
}
func TestCreatePFEDeployment(t *testing.T) {
	// Create a test codewind instance
	codewindInstance := setup_codewind()
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
			if deploy.GetName() != PFEPrefix+"-"+tt.codewind.WorkspaceID {
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
			if labels["app"] != "codewind-pfe" || labels["pfeWorkspace"] != tt.codewind.WorkspaceID {
				t.Error("PFE deploymeny labels improperly set")
			}

			// Verify only one image is used
			deployContainers := pod.Spec.Containers
			if len(deployContainers) != 1 {
				t.Errorf("PFE deployment had %v containers, expected %v", len(deployContainers), 1)
			}
			pfeContainer := deployContainers[0]
			pfeImage := PFEImage + ":" + PFEImageTag

			// Verify proper PFE image is being used
			if pfeContainer.Image != pfeImage {
				t.Errorf("PFE container using invalid image, had %v, expected %v", pfeContainer.Image, pfeImage)
			}
		})
	}
}

func TestCreatePFEService(t *testing.T) {
	// Create a test codewind instance
	codewindInstance := setup_codewind()
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
			if service.GetName() != PFEPrefix+"-"+tt.codewind.WorkspaceID {
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
			if serviceSelector["app"] != pfeLabels["app"] || serviceSelector["pfeWorkspace"] != pfeLabels["pfeWorkspace"] {
				t.Errorf("PFE service selector labels and pod labels don't match.")
			}
			// Verify that the proper PFE port is exposed
			servicePorts := service.Spec.Ports
			if len(servicePorts) != 1 {
				t.Errorf("PFE service exposing wrong number of ports. Had %v, expected %v", len(servicePorts), 1)
			}
			if servicePorts[0].Port != PFEContainerPort {
				t.Errorf("PFE service port not properly exposed. Port %v exposed instead of %v", servicePorts[0].Port, PFEContainerPort)
			}
		})
	}
}

func TestCreatePerformanceDeployment(t *testing.T) {
	// Create a test codewind instance
	codewindInstance := setup_codewind()
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
			if deploy.GetName() != PerformancePrefix+"-"+tt.codewind.WorkspaceID {
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
			if labels["app"] != "codewind-performance" || labels["performanceWorkspace"] != tt.codewind.WorkspaceID {
				t.Error("Performance deployment labels improperly set")
			}

			// Verify only one image is used
			deployContainers := pod.Spec.Containers
			if len(deployContainers) != 1 {
				t.Errorf("Performance deployment had %v containers, expected %v", len(deployContainers), 1)
			}
			perfContainer := deployContainers[0]
			pfeImage := PerformanceImage + ":" + PerformanceTag

			// Verify proper PFE image is being used
			if perfContainer.Image != pfeImage {
				t.Errorf("Performance container using invalid image, had %v, expected %v", perfContainer.Image, pfeImage)
			}
		})
	}
}

func TestCreatePerformanceService(t *testing.T) {
	// Create a test codewind instance
	codewindInstance := setup_codewind()
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
			if service.GetName() != PerformancePrefix+"-"+tt.codewind.WorkspaceID {
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
			if serviceSelector["app"] != pfeLabels["app"] || serviceSelector["performanceWorkspace"] != pfeLabels["performanceWorkspace"] {
				t.Errorf("PFE service selector labels and pod labels don't match.")
			}
			// Verify that the proper PFE port is exposed
			servicePorts := service.Spec.Ports
			if len(servicePorts) != 1 {
				t.Errorf("PFE service exposing wrong number of ports. Had %v, expected %v", len(servicePorts), 1)
			}
			if servicePorts[0].Port != int32(PerformanceContainerPort) {
				t.Errorf("PFE service port not properly exposed. Port %v exposed instead of %v", servicePorts[0].Port, PerformanceContainerPort)
			}
		})
	}
}
