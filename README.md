# Is it Observable
<p align="center"><img src="/image/logo.png" width="40%" alt="Is It observable Logo" /></p>

## Episode : What is KuberHealthy
This repository contains the files utilized during the tutorial presented in the dedicated IsItObservable episode related to Kuberhealthy.
<p align="center"><img src="/image/kuberhealthy.png" width="40%" alt="kuberhealthy Logo" /></p>

What you will learn
* How to use the [KuberHealthy](https://github.com/kuberhealthy/kuberhealthy)

This repository showcase the usage of Kuberhealthy  with :
* The Otel-demo
* The OpenTelemetry Operator
* Nginx ingress controller
* Dynatrace
* Locust
* TraceTest

We will send the Telemetry data produced by the Otel-demo application Dynatrace.

## Prerequisite
The following tools need to be install on your machine :
- jq
- kubectl
- git
- gcloud ( if you are using GKE)
- Helm


## Deployment Steps in GCP

You will first need a Kubernetes cluster with 2 Nodes.
You can either deploy on Minikube or K3s or follow the instructions to create GKE cluster:
### 1.Create a Google Cloud Platform Project
```shell
PROJECT_ID="<your-project-id>"
gcloud services enable container.googleapis.com --project ${PROJECT_ID}
gcloud services enable monitoring.googleapis.com \
    cloudtrace.googleapis.com \
    clouddebugger.googleapis.com \
    cloudprofiler.googleapis.com \
    --project ${PROJECT_ID}
```
### 2.Create a GKE cluster
```shell
ZONE=europe-west3-a
NAME=isitobservable-bindplane
gcloud container clusters create "${NAME}" \
 --zone ${ZONE} --machine-type=e2-standard-2 --num-nodes=3
```


## Getting started
### Dynatrace Tenant
#### 1. Dynatrace Tenant - start a trial
If you don't have any Dyntrace tenant , then i suggest to create a trial using the following link : [Dynatrace Trial](https://bit.ly/3KxWDvY)
Once you have your Tenant save the Dynatrace (including https) tenant URL in the variable `DT_TENANT_URL` (for example : https://dedededfrf.live.dynatrace.com)
```
DT_TENANT_URL=<YOUR TENANT URL>
```


#### 2. Create the Dynatrace API Tokens
Create a Dynatrace token with the following scope ( left menu Acces Token):
* ingest metrics
* ingest OpenTelemetry traces
<p align="center"><img src="/image/data_ingest.png" width="40%" alt="data token" /></p>
Save the value of the token . We will use it later to store in a k8S secret

```
DATA_INGEST_TOKEN=<YOUR TOKEN VALUE>
```
### 3.Clone the Github Repository
```shell
https://github.com/isItObservable/kuberhealthy
cd kuberhealthy
```
### 4.Deploy most of the components
The application will deploy the otel demo v1.2.1
```shell
chmod 777 deployment.sh
./deployment.sh  --clustername "${NAME}" --dturl "${DT_TENANT_URL}" --dttoken "${DATA_INGEST_TOKEN}"
```
### 5. Look at the predefined Khcheck
The helm chart deploying kuberhealthy also deployed a list of predefined checks :
* daemonset
* deploymnent
* dns-status-internal
Let's collect those checks :
```shell
kubectl get khcheck -n kuberhealthy
```
you should see the following output :
```shell
NAME                  AGE
daemonset             26d
deployment            26d
dns-status-internal   26d
```

Now let's collect the various status of those checks :
```shell
kubectl get khstate -n kuberhealthy
```
it will report the last status of the check and run duration of the check.

### 5. Look at the prometheus metrics automatically produced

The kuberhealthy is automatically prometheus metrics from our running checks on the `/metrics` endoint.
Let's create a port-forward on the service to look a the produced metrics :
```shell
kubectl port-forward svc/kuberhealthy 8085:80  -n kuberhealthy
```
open your browser on the following url: http://localhost:8085/metrics

### 6. Update our collector pipeline to collect the metrics from kuberhealthy
The updated collector has already been created.
This collector pipeline use the prometheus receiver to scrape a static target : kuberhealty .
```yaml
prometheus:
  config:
    scrape_configs:
    - job_name: 'kubernetes-kuberhealty'
      scrape_interval: 5s
      static_configs:
      - targets: ['kuberhealthy.kuberhealthy.svc:80']
```
apply the modified collector :
```shell
kubectl apply -f kubernetes-manifests/openTelemetry-manifest_prometheus.yaml
```
### 7. Create custom Kuberhealthy Checks

#### Locust
[Locust](https://docs.locust.io/en/stable/index.html) is an openSource loadtesting solution. The idea is to take advantage of an existing testing scenario to use if for our synthethic tests.
The following repository has a modified version of the locust script used to generate to traffic against the OpenTelemtry Demo.
This script is not producing traces , but report failes or sucess to Kuberhealty.

Let's have look at the modified script
```shell
cat loadgenerator/locustfile.py
```
i have prebuild a docker image with this script.
Now we need to create the khcheck using this new image :
let's have look at the khcheck :
```yaml 
apiVersion: comcast.github.io/v1
kind: KuberhealthyCheck
metadata:
  name: kh-loadgenerator-check
spec:
  runInterval: 30s # The interval that Kuberhealthy will run your check on
  timeout: 2m # After this much time, Kuberhealthy will kill your check and consider it "failed"
  extraLabels:
    app.kubernetes.io/name: example
    app.kubernetes.io/instance: example
    app.kubernetes.io/component: loadgeneratorcheck
  podSpec: # The exact pod spec that will run.  All normal pod spec is valid here.
    containers:
      - env: # Environment variables are optional but a recommended way to configure check behavior
          - name: FRONTEND_ADDR
            value: 'example-frontend:8080'
          - name: LOCUST_WEB_PORT
            value: "8089"
          - name: LOCUST_USERS
            value: "1"
          - name: LOCUST_SPAWN_RATE
            value: "1"
          - name: LOCUST_RUN_TIME
            value: 5m
          - name: LOCUST_HOST
            value: http://$(FRONTEND_ADDR)
          - name: LOCUST_HEADLESS
            value: "false"
          - name: LOCUST_AUTOSTART
            value: "true"
          - name: PROTOCOL_BUFFERS_PYTHON_IMPLEMENTATION
            value: python
          - name: LOADGENERATOR_PORT
            value: "8089"
        image: hrexed/oteldemo:v1.2.1-loadgenerator-kuberhealthy # The image of the check you just pushed
        imagePullPolicy: Always # During check development, it helps to set this to 'Always' to prevent on-node image caching.
        name: loadgeneratorcheck
```

This check is using the default environment variable used to run the script .

Let's deploy this new check :
```shell
kubectl apply -f kuberhealthy/loadgenerator_check.yaml -n otel-demo
```
#### Tracetest

[Tracetest](https://tracetest.io/) allow us to generate an end-to-end tests automatically from your openTelemtry Traces.
If you want to learn about Tracetest , check the [episode related to Tracetest](https://youtu.be/xj7tS2owRvk)

To use tracetest with kuberhealty, i briefly create go script that interact with the TraceTest CLI to trigger test.
The tracetest cli allow you to :
- configure the cli to connect to a specific tracetest instance
- load a datastore ( how tracetest can retrieve the openTelemtry Traces)
- run a test using a test definition file.

To run the tracetest check i need to attach to the "checker" pod :
- the datastore defintion file
- the test definition file

That is the reason why we need to apply first the configmap folding both settings :
```yaml 
apiVersion: v1
kind: ConfigMap
metadata:
  name: tracetest-config
data:
  test.yaml: |-
    type: Test
    spec:
      id: aOvr7Ro4R
      name: oteldemo_listproducts
      description: "OpenTelemetry Demo - List Products"
      trigger:
        type: http
        httpRequest:
          url: http://example-frontend.otel-demo.svc:8080/api/products
          method: GET
          headers:
          - key: Content-Type
            value: application/json
      specs:
        - selector: span[tracetest.span.type="rpc" name="hipstershop.ProductCatalogService/ListProducts"]
          assertions:
            - attr:rpc.grpc.status_code = 0
        - selector: span[tracetest.span.type="rpc" name="hipstershop.ProductCatalogService/ListProducts"
            rpc.system="grpc" rpc.method="ListProducts" rpc.service="hipstershop.ProductCatalogService"]
          assertions:
            - attr:rpc.grpc.status_code = 0
        - selector: span[tracetest.span.type="general" name="Tracetest trigger"]
          assertions:
            - attr:tracetest.response.status  =  200
            - attr:tracetest.span.duration < 50ms
  datastore.yaml: |-
      type: DataStore
      spec:
        name: Opentelemetry Collector pipeline
        type: otlp
        isDefault: true
```
let's first apply this configmap in the tracetest namespace:
```shell
kubectl apply -f kuberhealthy/tracetestcheck_config.yaml -n tracetestcheck_confi
```

let's have a quick look a how the go script is been structure:
```shell
cat traceTest/main.go
```

Now let's create the khcheck:
```yaml
apiVersion: comcast.github.io/v1
kind: KuberhealthyCheck
metadata:
  name: kh-tracetest-check
spec:
  runInterval: 30s # The interval that Kuberhealthy will run your check on
  timeout: 2m # After this much time, Kuberhealthy will kill your check and consider it "failed"
  podSpec: # The exact pod spec that will run.  All normal pod spec is valid here.
    containers:
      - env: # Environment variables are optional but a recommended way to configure check behavior
          - name: TRACETEST_URL
            value: "http://tracetest.tracetest.svc:11633"
        volumeMounts:
          - name: test-volume
            mountPath: /test
        image: hrexed/tracetest-kuberhealthycheck:0.1 # The image of the check you just pushed
        imagePullPolicy: Always # During check development, it helps to set this to 'Always' to prevent on-node image caching.
        name: tracetestcheck
    volumes:
      - name: test-volume
        configMap:
          # Provide the name of the ConfigMap containing the files you want
          # to add to the container
          name: tracetest-config
```

let's apply this khcheck:
```shell
kubectl apply -f kuberhealthy/tracetest_check.yaml -n tracetest
```
### 8. Look at the metrics ingested in Dynatrace







