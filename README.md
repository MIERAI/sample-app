# sample-app

> Example application of [slo-platform](https://github.com/MIERAI/slo-platform) — its quickstart example, deployed by
> [examples/sample-app](https://github.com/MIERAI/slo-platform/tree/main/examples/sample-app); principles and shared
> CI: [GitOps Platform Blueprint](https://github.com/MIERAI/gitops-platform-blueprint)

The public reference application for the delivery flow: a minimal Go HTTP service. What it shows:

- CI on every push to main: tests, a signed image (cosign, with SBOM and Trivy scan), promotion to dev
- canary releases with Argo Rollouts (analysis: the canary's error ratio and p95 latency from Prometheus)
- SLOs (availability, latency, probe) declared in slo-platform's
  [examples/sample-app/app.yaml](https://github.com/MIERAI/slo-platform/blob/main/examples/sample-app/app.yaml)

Run it locally: `go test ./... && go run .`, then `curl 'localhost:8080/hello?name=x'` (`PORT` changes the port).

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
