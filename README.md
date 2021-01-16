# delta-operator (WIP)
Kubernetes operator for [Delta](https://github.com/AndrewNeudegg/delta).

## Table of contents

- [delta-operator (WIP)](#delta-operator-wip)
  - [Table of contents](#table-of-contents)
  - [Setup](#setup)
  - [Example](#example)

## Setup

Clone the project and apply various directories:

```sh
git clone https://github.com/AndrewNeudegg/delta-operator.git
# Apply the CRD.
kubectl apply -f ./delta-operator/config/crd/bases/routing.andrewneudegg.com_delta.yaml
# Apply the RBAC
kubectl apply -f ./delta-operator/config/rbac/
# TODO: Deployment.
# kubectl apply -f .... deployment ....
```

## Example

```yaml
apiVersion: routing.andrewneudegg.com/v1
kind: Delta
metadata:
  name: delta-sample
spec:
  image: "andrewneudegg/delta:0.0.2"
  podAnnotations:
    andrewneudegg.com/annotation: "something,something-else"
  config: |
    applicationSettings: {}
    pipeline:
      # This first pipeline generates and emits http events.
      - id: pipelines/fipfo
        config:
          input:
            - id: utilities/generators/v1
              config:
                interval: 10s
                numberEvents: 10000
                numberCollections: 1
          output:
            - id: http/v1
              config:
                targetAddress: http://localhost:8080

      # This second pipeline consumes those events and writes to stdout.
      - id: pipelines/fipfo
        config:
          input:
            - id: http/v1
              config:
                listenAddress: :8080
                maxBodySize: 1000000 # 1mb
          output:
            - id: utilities/performance/v1
              config:
                sampleWindow: 60s
              nodes:
                - id: utilities/console/v1
```