# Phoenix

Current chart version is `0.3.0`.

## Deploy Phoenix

We do not provide a public Docker repository where you can download the `phoenix` image from. You need to create your own repository and push the image over there and set it here.

The default values will create all the services that are required to make Phoenix work out of the box. Make sure you are setting properly the `ENV` variables to connect to `S3`. To install the chart simply run the below command:

```bash
$: helm upgrade --install phoenix ./chart -f values.yaml --namespace phoenix --debug --recreate-pods
```

If you find something that doesn't work, please open up an Issue or a PR! We :heart: contibutions

## Chart Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| fullnameOverride | string | `""` |  |
| nameOverride | string | `""` |  |
| image.pullPolicy | string | `"Always"` | Image pull policy |
| image.repository | string | `"repository/phoenix"` | Image repository name |
| image.tag | string | `"latest"` | Image tag |
| ingress.annotations | object | `{}` | Ingress annotations (values are templated) |
| ingress.enabled | bool | `false` | Enables Ingress |
| ingress.hosts | array | `[]` | Ingress accepted hostnames |
| internal.affinity | object | `{}` | Affinity settings for pod assignment |
| internal.data | string | `{}` | ENV Varibles to be set for the internal service |
| internal.nodeSelector | object | `{}` | Node labels for pod assignment |
| internal.replicaCount | int | `1` | Number of nodes |
| internal.resources | object | `{}` | CPU/Memory resource requests/limits |
| internal.service.port | int | `8081` | Kubernetes port where service is exposed |
| internal.service.type | string | `"ClusterIP"` | Kubernetes service type |
| internal.tolerations | object | `{}` | Toleration labels for pod assignment  |
| public.affinity | object | `{}` | Affinity settings for pod assignment |
| public.data | string | `{}` | ENV Varibles to be set for the public service |
| public.nodeSelector | object | `{}` | Node labels for pod assignment |
| public.replicaCount | int | `1` | Number of nodes |
| public.resources | object | `{}` | CPU/Memory resource requests/limits |
| public.secrets | object | `{}` | Secrets to add to the public service |
| public.service.port | int | `8082` | Kubernetes port where service is exposed |
| public.service.type | string | `"ClusterIP"` | Kubernetes service type |
| public.tolerations | object | `{}` | Toleration labels for pod assignment |
| redis.enabled | bool | `true` | Enable the deployment of a local redis instance |
| redis.resources | object | `{}` | CPU/Memory resource requests/limits  |
| worker.affinity | object | `{}` | Affinity settings for pod assignment |
| worker.data | string | `{}` | ENV Varibles to be set for the worker service |
| worker.replicaCount | int | `1` | Number of nodes |
| worker.resources | object | `{}` | CPU/Memory resource requests/limits |
| worker.tolerations | object | `{}` | Toleration labels for pod assignment |