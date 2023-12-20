## <div align="center"> [![PkgGoDev](https://pkg.go.dev/badge/github.com/movewp3/keycloakclient-controller)](https://pkg.go.dev/github.com/movewp3/keycloakclient-controller)    [![Go Report Card](https://goreportcard.com/badge/github.com/movewp3/keycloakclient-controller)](https://goreportcard.com/report/github.com/movewp3/keycloakclient-controller)   [![codecov](https://codecov.io/gh/movewp3/keycloakclient-controller/branch/main/graph/badge.svg?token=tNKcOjlxLo)](https://codecov.io/gh/movewp3/keycloakclient-controller)      [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
</div>

# keycloakclient-controller
The keycloakclient-controller **manages keycloak clients in independent keycloak installations**. 

To create a KeycloakClient in a Keycloak Installation, a **KeycloakClient-CustomResource** is created, and the keycloakclient-controller sees to creating, changing, deleting the KeycloakClient as specified with the CustomResource.


## Description

This Operator has its origin from the [Legacy Keycloak Operator](https://github.com/keycloak/keycloak-operator).
If you look for the official KeycloakOperator from RedHat, please look into the [KeycloakOperator](https://github.com/keycloak/keycloak/tree/main/operator).

The Operator is opinionated in a way that it expects that Keycloak and
the KeyclokRealm are already set up (i.e. with one of the available Helm Charts) and it only has
to handle the KeycloakClients for a Keycloak Installation and a specific realm.

This fits our need as we set up Keycloak and the realm with Helm, and we have a lot of microservices that require their own KeycloakClient.
The Microservices are deployed via Helm, so it is easy to simply deploy a KeycloakClient Resource together with the other artefacts of the Microservice and let
the Operator handle the creation of the KeycloakClient in Keycloak.

## Getting Started
You’ll need a Kubernetes cluster to run against. You can use [KIND](https://sigs.k8s.io/kind) to get a local cluster for testing, or run against a remote cluster.
**Note:** Your controller will automatically use the current context in your kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).

### Running on the cluster
1. Install Instances of Custom Resources:

    ```sh
    make install
    ```

2. Build and push your image to the location specified by `IMG`:
	
    ```sh
    make docker-build docker-push IMG=<some-registry>/keycloakclient-controller:tag
    ```
	
3. Deploy the controller to the cluster with the image specified by `IMG`:

    ```sh
    make deploy IMG=<some-registry>/keycloakclient-controller:tag
    ```

### Uninstall CRDs
To delete the CRDs from the cluster:

```sh
make uninstall
```

### Undeploy controller
UnDeploy the controller to the cluster:

```sh
make undeploy
```

## Contributing
// TODO(user): Add detailed information on how you would like others to contribute to this project

### How it works
This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/) 
which provides a reconcile function responsible for synchronizing resources until the desired state is reached on the cluster 

### Test It Out
1. Install the CRDs into the cluster:

    ```sh
    make install
    ```

2. Run your controller (this will run in the foreground, so switch to a new terminal if you want to leave it running):

    ```sh
    make run
    ```

**NOTE:** You can also run this in one step by running: `make install run`

### Modifying the API definitions
If you are editing the API definitions, generate the manifests such as CRs or CRDs using:

```sh
make manifests
```

**NOTE:** Run `make --help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

function myFunction() {
// Get the text field
var copyText = document.getElementById("myInput");

// Select the text field
copyText.select();
copyText.setSelectionRange(0, 99999); // For mobile devices

// Copy the text inside the text field
navigator.clipboard.writeText(copyText.value);

// Alert the copied text
alert("Copied the text: " + copyText.value);
} 
