# Required cluster role for Codewind 

This clusterrole.yaml file can be used to create a cluster role with the required permissions for Codewind to properly function.

Note: This needs to be applied <b>before</b> installing Che and the cluster role name must be configured via the Che helm chart.

## Usage:

1. Create the cluster role
`kubectl apply -f clusterrole.yaml`

2. If `global.cheWorkspaceNamespace` was set: Apply the cluster role binding
     - Change `<namespace>` to be the Che workspace namespace
     - Run `kubectl apply -f rolebinding.yaml`
2. Update the `cheWorkspaceClusterRole` in the Che helm chart's values.yaml file with the name of the cluster role
eg. `cheWorkspaceClusterRole: "eclipse-codewind"`

3. Install Che
