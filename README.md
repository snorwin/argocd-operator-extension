# argocd-operator-extension
[![GitHub Action](https://img.shields.io/badge/GitHub-Action-blue)](https://github.com/features/actions)
[![Build](https://img.shields.io/github/workflow/status/snorwin/argocd-operator-extension/CI?label=build&logo=github)](https://github.com/snorwin/argocd-operator-extension/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/snorwin/argocd-operator-extension)](https://goreportcard.com/report/github.com/snorwin/argocd-operator-extension)
[![Coverage Status](https://coveralls.io/repos/github/snorwin/argocd-operator-extension/badge.svg?branch=main)](https://coveralls.io/github/snorwin/argocd-operator-extension?branch=main)
[![Releases](https://img.shields.io/github/v/release/snorwin/argocd-operator-extension)](https://github.com/snorwin/argocd-operator-extension/releases)
[![Image](https://img.shields.io/badge/image_repository-ghcr.io-blue)](https://github.com/users/snorwin/packages/container/package/argocd-operator-extension)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

The **argocd-operator-extension** is an operator extension for the [Argo CD Operator](https://argocd-operator.readthedocs.io/) in order to automate the handling of the Kubernetes RBAC (i.e. service accounts, roles, role bindings) for multiple Argo CD instances in a shared cluster.

## Use case
By default, Argo CD requires [cluster-wide read privileges](https://argoproj.github.io/argo-cd/operator-manual/security/). 
However, this approach is not recommended at all and it does not follow the least privilege principle.
The service accounts of Argo CD should not have more capabilities as the user using it in order to prevent bypassing the Kubernetes RBAC and other security measures protected by Kubernetes RBAC using Argo CD.
For the same reason it is also not recommended that multiple teams with different Kubernetes RBAC policies share the same Argo CD instances, because the Kubernetes RBAC needs to somehow be converted to the Argo CD RBAC which can lead to mistakes and the integrity of each team's application cannot be guaranteed.
The ArgoCD service account, and not Argo CD RBAC, defines the baseline of capabilities that must not exceed the capabilities granted to the user by the Kubernetes RBAC.

The **argocd-operator-extension** solves all the issue described in the previous section by facilitating that an Argo CD instance is used only for a defined subset of the namespaces without using cluster role bindings. The view and edit roles are only granted to the individual Argo CD service accounts for dedicated namespaces.

## How it works
The **argocd-operator-extension** reconciles the `ArgoCD` customer resource of the Argo CD Operator and installs a Helm chart which contains the internal service accounts and role bindings as well as the role bindings to the `argocd-edit` and `argocd-view` cluster role for all the namespaces with the label `argocd.snorwin.io/name` and `argocd.snorwin.io/namespace` set to the namespaced name of the reconciled object.
The ArgoCD RBAC blueprint is defined as a [Helm chart](helm/charts/argocd-operator-extension/resources) and mounted to the extension using a config map which allows you to use this operator with your existing roles and adapt it that it fits your requirements without re-building the image of the extension.

Upgrading many Argo CD instances in a cluster by hand is inefficient, therefore the extension is able to manage the images and versions of Argo CD, Dex and Redis automatically in the `ArgoCD` customer resource based on the update policy (`None`, `Always` or `IfNotPresent`) annotated to the resource itself. The images and versions can be set using environment variables. 

## Getting Started
### Installation
There are two ways how the **argocd-operator-extension** can be installed:
#### Helm
1. Clone this repository and if required adapt the Argo CD RBAC blueprint helm chart.
2. Install the [Helm chart](helm/) in order to install the bundle of the Argo CD Operator including the extension.

#### Docker
_Prerequisite: Argo CD Operator is already installed_
1. Clone this repository and if required adapt the Argo CD RBAC blueprint helm chart.
2. Create a Dockerfile:
    ```
    FROM ghcr.io/snorwin/argocd-operator-extension:latest
    
    ENV HELM_DIRECTORY=/data/helm
    
    ADD ./helm/charts/argocd-operator-extension/resources /data/helm/
    ```
3. Deploy the image.

### Deploy the first ArgoCD instance
1. Create an `ArgoCD` instance [See the ArgoCD Reference](https://argocd-operator.readthedocs.io/en/latest/reference/argocd/).
    ```
    apiVersion: argoproj.io/v1alpha1
    kind: ArgoCD
    metadata:
        name: example-argocd
        namespace: example
        annotations:
            argocd.snorwin.io/image-update-policy: Always
    spec: {}
    ```
2. Add labels to the application's target namespace:
    ```
    kubectl label namespace <app namespace> argocd.snorwin.io/name=example-argocd argocd.snorwin.io/namespace=example
    ```
 
 ## Configuration
 ### Environment Variables
 - `HELM_DIRECTORY` - directory of the Helm chart in the container
 - `HELM_DRIVER` - helm storage driver. It can be set to one of the values: `configmap`, `secret`, `memory` (default value: `secret`)
 - `HELM_MAX_HISTORY` - limit the maximum number of revisions saved per helm release (default: 10). Use 0 for no limit.
 - `CLUSTER_ARGOCD_NAMESPACEDNAMES` - comma separated list of NamespacedNames (`namespace/name`) of Argo CD instances which run in cluster mode
 - `ARGOCD_IMAGE` - ArgoCD image and version `[<image>][:<version>]` used for automated version updates
 - `DEX_IMAGE` - Dex image and version `[<image>][:<version>]` used for automated version updates
 - `REDIS_IMAGE` - Redis image and version `[<image>][:<version>]` used for automated version updates
 
 ## Compatibility
 The **argocd-operator-extension** is compatible with the version [v0.0.14](https://github.com/argoproj-labs/argocd-operator/releases/tag/v0.0.14) and [v0.0.15](https://github.com/argoproj-labs/argocd-operator/releases/tag/v0.0.15) of the Argo CD Operator.
