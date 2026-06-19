# Sample app: to-do list with Redis

This is a small end-to-end example that deploys a sample to-do list web app and a Redis cache into the k3s cluster that runs inside the Docker Compose stack. It exercises the full Radius path: a container, a connection, and a recipe-provisioned data store.

## What gets deployed

| Resource | Type                                  | Description                                                                                                |
|----------|---------------------------------------|------------------------------------------------------------------------------------------------------------|
| `todo`   | `Applications.Core/containers`        | The sample to-do web app (`ghcr.io/radius-project/samples/demo`) listening on port 3000.                   |
| `db`     | `Applications.Datastores/redisCaches` | A Redis cache created by the environment's recipe, which renders a Redis pod and service into the cluster. |

The `connections` block on the container wires the Redis connection (host, port, password) into the app as environment variables, so the app finds its database without any hard-coded values.

## Prerequisites

- The Compose control plane is up and healthy. See the [parent README](../README.md) for `docker compose up`.
- Your `rad` CLI is pointed at the Compose workspace (the `compose` workspace described in the [parent README](../README.md)).
- A resource group and a Kubernetes environment exist. The parent README walks through creating them; the commands below assume a group/environment both named `default`.

## Deploy

The Redis resource is fulfilled by a recipe, so register the local-dev Redis recipe once per environment before deploying:

```bash
rad recipe register default \
  --environment default \
  --template-kind bicep \
  --template-path ghcr.io/radius-project/recipes/local-dev/rediscaches:latest \
  --resource-type Applications.Datastores/redisCaches
```

Then deploy the app from this directory (the `bicepconfig.json` here lets `rad` resolve the `radius` Bicep extension):

```bash
cd deploy/compose/test
rad deploy app.bicep --application todo
```

A successful run prints both resources:

```text
Resources:
    todo            Applications.Core/containers
    db              Applications.Datastores/redisCaches
```

## Verify

Radius schedules the app into a namespace named `<environment-namespace>-<application>`. With the `default` environment (namespace `default`) and the `todo` application, that namespace is `default-todo`:

```bash
docker compose -f ../docker-compose.yml exec k3s kubectl get pods,svc -n default-todo
```

You should see a `todo` pod and a `redis-*` pod, both `Running`, plus their services:

```text
NAME                                       READY   STATUS    RESTARTS   AGE
pod/todo-...                               1/1     Running   0          18s
pod/redis-...                              2/2     Running   0          67s

NAME                          TYPE        CLUSTER-IP      PORT(S)
service/todo                  ClusterIP   10.43.x.x       3000/TCP
service/redis-...             ClusterIP   10.43.x.x       6379/TCP
```

You can also list the Radius resources:

```bash
rad resource list Applications.Core/containers --application todo
rad resource list Applications.Datastores/redisCaches --application todo
```

## Browse the app from your machine

The container port (3000) is not published to the host, so reach it with `kubectl port-forward`. Note that `rad resource expose` and `rad init` do not work against this stack because they use your host's current kube context (see the [parent README](../README.md)); the steps below point `kubectl` at the in-Compose cluster directly instead.

The `k3s` service publishes its API on `localhost:6443`, so your own `kubectl` works — you just need the cluster's kubeconfig, which k3s serves at a standard path (its server is already `https://127.0.0.1:6443`):

```bash
docker compose exec k3s cat /etc/rancher/k3s/k3s.yaml > k3s.kubeconfig
KUBECONFIG=$PWD/k3s.kubeconfig kubectl port-forward -n default-todo svc/todo 8080:3000
```

Leave that running and open <http://localhost:8080> in your browser (or `curl http://localhost:8080`). Press `Ctrl+C` to stop forwarding.

> The shared `kubeconfig` volume the Radius containers use points at `https://k3s:6443`, a hostname that only resolves inside the Compose network — that is why you need the `127.0.0.1:6443` kubeconfig above rather than that one.

If you would rather not install or configure `kubectl` on the host, run the port-forward inside a throwaway `kubectl` container attached to the Compose network; it reads the shared kubeconfig volume and publishes the forwarded port to your machine:

```bash
docker run --rm --name todo-pf \
  --network radius_default \
  -p 8080:8080 \
  -v radius_kubeconfig:/root/.kube:ro \
  alpine/kubectl:1.33.4 \
  port-forward --address 0.0.0.0 -n default-todo svc/todo 8080:3000
```

The `radius_default` network and `radius_kubeconfig` volume names assume the default Compose project name `radius`; adjust them if you started the stack with a different project name.

## Clean up

Delete the application and its resources:

```bash
rad app delete todo --yes
```

This removes the `todo` and `db` Radius resources and the workloads they created in the `default-todo` namespace.
