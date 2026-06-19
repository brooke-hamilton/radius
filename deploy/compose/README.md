# Radius on Docker Compose

Run the Radius control plane locally as containers with Docker Compose, as an alternative to installing the Helm chart on a Kubernetes cluster. This is the containerized equivalent of `make debug-start`: it uses the **published, unmodified Radius images** — no source changes are required.

## What this runs

| Service             | Purpose                                                          | Port |
|---------------------|------------------------------------------------------------------|------|
| `ucp`               | Universal Control Plane (the entry point the `rad` CLI talks to) | 9000 |
| `applications-rp`   | Applications resource provider                                   | 5443 |
| `dynamic-rp`        | Dynamic resource provider                                        | 8082 |
| `deployment-engine` | Bicep/ARM deployment engine                                      | 6443 |
| `controller`        | Kubernetes-native controller (optional, `full` profile)          | 8083 |
| `postgres`          | Control-plane state store                                        | 5432 |
| `k3s`               | In-Compose Kubernetes API (secrets + encryption key)             | 6443 |

State is stored in **PostgreSQL**, not the Kubernetes API server. The `k3s` container exists only because the control-plane components require a reachable Kubernetes API for `secretProvider: kubernetes` and for the dynamic-rp encryption key — it is not used as a workload deployment target by default.

## Prerequisites

- Docker Desktop (or any Docker Engine that supports Compose v2) with the ability to run a `privileged` container (required by k3s).
- The `rad` CLI on your host.

## Start the control plane

```bash
cd deploy/compose
docker compose up --detach
```

The bootstrap jobs (`bootstrap-db`, `bootstrap-k8s`) run once to create the databases, the `radius-system` namespace, the encryption key, and the CRDs, then the control-plane services start. Watch progress with:

```bash
docker compose ps
docker compose logs --follow ucp applications-rp dynamic-rp deployment-engine
```

To also run the optional Kubernetes-native controller:

```bash
docker compose --profile full up --detach
```

## Connect the rad CLI

The control plane is reachable at `http://localhost:9000`. Configure a `rad` workspace that points directly at it with a connection override, so the CLI talks to UCP over HTTP and never needs a kubeconfig.

> **`rad init` does not work with this deployment.** `rad init` is a Kubernetes/Helm operation: it verifies and installs Radius as a Helm release in a cluster using your kubeconfig's current context, ignoring the `overrides.ucp` connection. Because this deployment runs the control plane as containers (not a Helm release), `rad init` has nothing to discover and will fail against your kube context. Configure the workspace manually as shown below instead — that covers everything `rad init` would otherwise set up.

Add a workspace to `~/.rad/config.yaml`:

```yaml
workspaces:
  default: compose
  items:
    compose:
      connection:
        kind: kubernetes
        context: ""
        overrides:
          ucp: http://localhost:9000
      environment: /planes/radius/local/resourceGroups/default/providers/Applications.Core/environments/default
      scope: /planes/radius/local/resourceGroups/default
```

> The connection `kind` must be `kubernetes` (the only supported kind), but the `overrides.ucp` value makes the CLI use a direct HTTP connection — the `context` is ignored.

Then create the default resource group and environment:

```bash
rad group create default
rad env create default
rad env switch default
```

Verify connectivity:

```bash
rad group list
rad env list
```

## Configuration

Image registries, tags, and versions are controlled by [.env](.env):

| Variable            | Default                                           | Description                        |
|---------------------|---------------------------------------------------|------------------------------------|
| `REGISTRY`          | `ghcr.io/radius-project`                          | Registry/org for the Go components |
| `TAG`               | `latest`                                          | Image tag for the Go components    |
| `DE_IMAGE`          | `ghcr.io/radius-project/deployment-engine:latest` | Deployment Engine image            |
| `K3S_VERSION`       | `v1.33.4-k3s1`                                    | k3s image tag                      |
| `KUBECTL_VERSION`   | `1.33.4`                                          | kubectl image tag (bootstrap job)  |
| `POSTGRES_PASSWORD` | `radius_pass`                                     | PostgreSQL superuser password      |

To pin a specific Radius release, set `TAG` and `DE_IMAGE` to that version.

## Stop and clean up

```bash
docker compose down            # stop containers, keep state
docker compose down --volumes  # also delete PostgreSQL, k3s, and kubeconfig volumes
```

## How it works

- **No code changes.** The component images are run with their existing entrypoints; only `--config-file` and environment variables are supplied.
- **Kubeconfig delivery.** The Radius binaries read their kubeconfig from `$HOME/.kube/config` (they do not honor `KUBECONFIG`). Each component sets `HOME=/home/radius` and mounts the shared `kubeconfig` volume at `/home/radius/.kube`. `bootstrap-k8s` rewrites the k3s kubeconfig so the API server address resolves to the `k3s` Compose service.
- **Config files.** The files under [config/](config/) mirror the debug configs in `build/configs/`, changing only hostnames (Compose service DNS) and the PostgreSQL/manifest paths.
- **Provider routing.** UCP routes resource-provider requests using the provider manifests baked into the `ucpd` image, which point at the Kubernetes service names `applications-rp.radius-system:5443` and `bicep-de.radius-system:6443`. Rather than ship custom manifests (which would drift per image version), the `applications-rp` and `deployment-engine` services expose those exact names via Compose network `aliases` and listen on the matching ports.
- **In-cluster workloads.** The Deployment Engine runs with `--kubernetes=true` and the shared kubeconfig, so containers and recipe-based resource types are scheduled into the bundled `k3s` cluster. The [test/](test/) sample exercises this end to end.

## Deploy a sample app

Once the control plane is running and the `rad` CLI is connected, the [test/](test/README.md) folder contains a small to-do list app (a container plus a recipe-provisioned Redis cache) that deploys into the bundled k3s cluster. Follow [test/README.md](test/README.md) to register the Redis recipe and run `rad deploy`.

## Limitations

- Deploying cloud (Azure/AWS) resource types directly from Bicep is not configured in this setup (no cloud credentials are wired into the Deployment Engine). Kubernetes resource types and recipes work; cloud providers do not.
- The `k3s` container runs `privileged`, which Docker Desktop supports but some locked-down environments may not.
