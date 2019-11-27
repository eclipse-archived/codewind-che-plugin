## Che setup script

The script `./che-setup.sh` can be leveraged to deploy a che instance instance to your favorite OpenShift cluster. At the moment, the script only support openshift kube environment - future enhancement to the script is planned to support other kube environment.

**Script usage**
```
Usage: che-setup.sh: [-<option letter> <option value> | -h]
Options:
    --cluster-ip        Cluster ip - Required
    --cluster-user      Cluster username - Required
    --cluster-pass      Cluster password - Required
    --cluster-port      Cluster port - default: 8443
    --cluster-token     Cluster token - Optional (can be used instead of user/pass)
    --che-ns            Namespace to install Che - default: che
    --che-version       Che version to install - default: next
    --clean-deploy      Deploy a clean che - default: n
    --operator-yaml     Absolute Path to che operator yaml - default: github.com/eclipse/codewind-che-plugin/master/setup/install_che/che-operator/codewind-checluster.yaml
    --operator-image    The container image of the operator - default: uses the default operator container image
    --service-account   Service account name - default: che-user
    --podreadytimeout   Pod ready timeout - default: 600000
    --podwaittimeout    Pod wait timeout - default: 1200000
    --default-registry  Enable this flag to add the default docker registry - default: n
    --install-codewind  Enable this flag to install codewind from a devfile - default: https://raw.githubusercontent.com/eclipse/codewind-che-plugin/master/devfiles/latest/devfile.yaml
    -h | --help         Display the man page
```

**Script example**

1. Deploy a fresh che instance using my favorite operator file, set up a default docker registry and install codewind from my favorite devfile.
```
./che-setup.sh --cluster-ip=<cluster_ip> --cluster-user=<cluster_user> --cluster-pass=<cluster_pass> --che-ns=che --operator-yaml=<my_favorite_operator_yaml> --clean-deploy --default-registry --install-codewind=<my_favorite_devfile>
```

2. Deploy a fresh che instance using the default operator file, set up a default docker registry and install codewind from the master devfile.
```
./che-setup.sh --cluster-ip=<cluster_ip> --cluster-user=<cluster_user> --cluster-pass=<cluster_pass> --che-ns=che --operator-yaml --clean-deploy --default-registry --install-codewind
```

3. Deploy a fresh che instance using the stable che version (che's latest release), set up a default docker registry but don't install codewind.
```
./che-setup.sh --cluster-ip=<cluster_ip> --cluster-user=<cluster_user> --cluster-pass=<cluster_pass> --che-ns=che --che-version=next --clean-deploy --default-registry
```
