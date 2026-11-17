# sample-app

> Part of the [GitOps Platform Blueprint](https://github.com/MIERAI/gitops-platform-blueprint)

The public reference application for the delivery flow: a minimal Go HTTP service.

| Endpoint | |
|---|---|
| `GET /healthz` | `ok`; header `X-App-Version` = image version (used to confirm which version is serving) |
| `GET /hello?name=x` | `{"message": "hello, x", "version": "…"}` |

## CI (`.github/workflows/ci.yml`)

- Pull request: `go vet` + `go test -race`
- main: test → **shared** `build-image` (GHCR, tag = commit, Trivy, SBOM, cosign signature) → **shared** `promote-dev`
  (updates `gitops-manifests/sample-app/overlays/dev`) → Argo CD rolls it out

```sh
go test ./... && docker build -t sample-app . && docker run -p 8080:8080 sample-app
```
