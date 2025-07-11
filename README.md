# growi-manager

[![Go Report Card](https://goreportcard.com/badge/github.com/maeshinshin/growi-manager)](https://goreportcard.com/report/github.com/maeshinshin/growi-manager)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](

## Description

deploys and manages GROWI app on Kubernetes clusters.

- Growi app
- MongoDB
- Elasticsearch

## Getting Started

### Prerequisites
- go version v1.23.0+
- docker version 17.03+.
- kubectl version v1.11.3+.
- Access to a Kubernetes v1.11.3+ cluster.

### To Deploy on the cluster

#### With kubectl

```
kubectl apply -f https://raw.githubusercontent.com/maeshinshin/growi-manager/main/dist/install.yaml
```

#### With Helm

1. Install the Helm chart:

```
helm repo add growi-manager https://maeshinshin.github.io/growi-manager
helm repo update
```

2. Install the chart:

```sh
helm install growi-manager growi-manager/growi-manager --namespace growi-manager --create-namespace
```

### To Uninstall

#### With kubectl

```sh
kubectl delete -f https://raw.githubusercontent.com/maeshinshin/growi-manager/main/dist/install.yaml
```

#### With Helm

1. Uninstall the chart:

```sh
helm uninstall growi-manager --namespace growi-manager
```
2. Delete the namespace:

```sh
kubectl delete namespace growi-manager
```

## License

Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

