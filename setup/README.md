# Required cluster role for Codewind 

This clusterrole.yaml file can be used to create a cluster role with the required permissions for Codewind to properly function.

Note: This needs to be applied <b>before</b> installing Che and the cluster role name must be configured via the Che helm chart.

## Usage:

1. Create the cluster role
`kubectl apply -f clusterrole.yaml`

1. Update the `cheWorkspaceClusterRole` in the Che helm chart's values.yaml file with the name of the cluster role
eg. `cheWorkspaceClusterRole: "eclipse-codewind"`

1. Install Che
