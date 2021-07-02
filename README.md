# Argonaut

> Argonaut is a Kubernetes Operator leveraging Cloudflare services to "ingress" traffic to your workloads.

Turn's an opinionated CRD into automatically configured Cloudflare services and creates a managed Deployment of
cloudflared container image configured to tunnel traffic to your workload.

Example CRD
```yaml
apiVersion: argonaut.metalabs.no/v1beta1
kind: Argonaut
metadata:
  name: example
  namespace: example
spec:
  argoTunnelName: "example"
  cfAuthSecret:
    name: example
    namespace: example
  ingress:
    - hostname: test.example.com
      serviceSelector:
        matchLabels:
          app: nginx
```

Prerequisites for using this is a Cloudflare account, an existing DNS Zone and a API Token with appropriate 
permissions. This information should put inside a Kubernetes Secret with this structure, which is referenced in cfAuthSecret above:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: example
  namespace: example
data:
  accountid: <redacted>
  email: <redacted>
  token: <redacted>
kind: Secret
type: Opaque
```

## Status

DO NOT USE THIS FOR ANYTHING IMPORTANT, THIS IS VERY MUCH A WORK IN PROGRESS

I'm just figuring out what and how to build it here. Learning kubebuilder and controller-runtime as we go.
