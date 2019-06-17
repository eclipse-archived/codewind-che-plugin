## Installing Codewind on Kubernetes

### Prerequisites

1. Che operator installed. See https://github.com/eclipse/che-operator/
    **Note:** You may need the `oc` CLI installed to run deploy.sh
2. Kubernetes 1.11 or higher
3. Ingress

### Install

1. Determine your ingress domain. It should be of the form <IP>.nip.io.
    - If you're running on IBM Cloud Private, this will be the public IP address of your proxy node.
    - On other Kubernetes, use your master node IP address
2. Configure kubectl for your cluster
    - This will depend on your cluster
    - On OpenShift use `oc login`, on IBM Cloud Private, use `cloudctl login`.
3. Run `install.sh` to deploy Eclipse Che