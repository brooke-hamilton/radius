#!/bin/bash

# Proves Repo Radius state can round-trip through GHCR while the workload
# remains on a separate Kubernetes cluster.

set -euo pipefail

readonly APP_NAME="repo-radius-state-e2e"
readonly RESOURCE_NAME="repo-radius-state-container"
readonly ENVIRONMENT_NAME="repo-radius-state-e2e"
readonly WORKLOAD_NAMESPACE="repo-radius-state-e2e"
readonly WORKSPACE_NAME="repo-radius-state-e2e"
readonly SELECTOR="radapp.io/application=${APP_NAME},radapp.io/resource=${RESOURCE_NAME}"
readonly REPOSITORY_ROOT="${GITHUB_WORKSPACE:-$(git rev-parse --show-toplevel)}"
readonly DIAGNOSTICS_DIR="${REPOSITORY_ROOT}/dist/repo-radius-state-e2e"
readonly SOURCE_APP_FILE="${REPOSITORY_ROOT}/test/functional-portable/statestore/noncloud/testdata/repo-radius-state-app.bicep"
readonly RUN_SUFFIX="${GITHUB_RUN_ID:-local}-${GITHUB_RUN_ATTEMPT:-1}"
readonly NETWORK_NAME="repo-radius-state-${RUN_SUFFIX}"
readonly REGISTRY_CONTAINER="repo-radius-registry-${RUN_SUFFIX}"
readonly REGISTRY_ALIAS="repo-radius-registry"
readonly CLUSTER_REGISTRY="${REGISTRY_ALIAS}:5000"
readonly WORKLOAD_CLUSTER="radius-workload-${RUN_SUFFIX}"
readonly CONTROL_PLANE_A="radius-cp-a-${RUN_SUFFIX}"
readonly CONTROL_PLANE_B="radius-cp-b-${RUN_SUFFIX}"
readonly WORK_DIR="${RUNNER_TEMP:-/tmp}/repo-radius-state-${RUN_SUFFIX}"
readonly HOST_WORKLOAD_KUBECONFIG="${WORK_DIR}/workload-host.kubeconfig"
readonly INTERNAL_WORKLOAD_KUBECONFIG="${WORK_DIR}/workload-internal.kubeconfig"
readonly REGISTRY_CONFIG="${WORK_DIR}/registries.yaml"
readonly APP_FILE="${WORK_DIR}/repo-radius-state-app.bicep"

: "${DOCKER_REGISTRY:?DOCKER_REGISTRY must be set}"
: "${DOCKER_TAG_VERSION:?DOCKER_TAG_VERSION must be set}"
: "${RADIUS_STATE_ARCHIVE:?RADIUS_STATE_ARCHIVE must be set}"
: "${RADIUS_STATE_BACKEND:?RADIUS_STATE_BACKEND must be set}"
: "${RADIUS_STATE_REGISTRY:?RADIUS_STATE_REGISTRY must be set}"

export RADIUS_PREVIEW=true
export RADIUS_STATE_REGISTRY="${RADIUS_STATE_REGISTRY,,}"
readonly STATE_REFERENCE="${RADIUS_STATE_REGISTRY}:${RADIUS_STATE_ARCHIVE}"

state_owned_by_run=false

collect_diagnostics() {
    mkdir -p "${DIAGNOSTICS_DIR}"

    rad app list --output json \
        >"${DIAGNOSTICS_DIR}/rad-app-list.json" 2>&1 || true
    oras manifest fetch --descriptor "${STATE_REFERENCE}" \
        >"${DIAGNOSTICS_DIR}/state-descriptor.json" 2>&1 || true

    local cluster
    for cluster in "${CONTROL_PLANE_A}" "${CONTROL_PLANE_B}"; do
        if ! k3d cluster list --no-headers \
            | awk '{print $1}' \
            | grep -Fxq "${cluster}"; then
            continue
        fi

        kubectl --context "k3d-${cluster}" get pods -A -o wide \
            >"${DIAGNOSTICS_DIR}/${cluster}-pods.txt" 2>&1 || true
        kubectl --context "k3d-${cluster}" get events -A \
            --sort-by=.lastTimestamp \
            >"${DIAGNOSTICS_DIR}/${cluster}-events.txt" 2>&1 || true

        local component
        for component in applications-rp dynamic-rp bicep-de controller ucp; do
            kubectl --context "k3d-${cluster}" logs \
                -n radius-system \
                -l "app.kubernetes.io/name=${component}" \
                --all-containers --tail=300 \
                >"${DIAGNOSTICS_DIR}/${cluster}-${component}.log" \
                2>&1 || true
        done
    done

    if [[ -f "${HOST_WORKLOAD_KUBECONFIG}" ]]; then
        kubectl --kubeconfig "${HOST_WORKLOAD_KUBECONFIG}" get all -A -o wide \
            >"${DIAGNOSTICS_DIR}/workload-resources.txt" 2>&1 || true
        kubectl --kubeconfig "${HOST_WORKLOAD_KUBECONFIG}" get deployment \
            -n "${WORKLOAD_NAMESPACE}" -l "${SELECTOR}" -o yaml \
            >"${DIAGNOSTICS_DIR}/workload-deployment.yaml" 2>&1 || true
    fi
}

delete_state_manifest() {
    local status

    if manifest_exists; then
        status=0
    else
        status=$?
    fi
    if ((status == 1)); then
        return 0
    fi
    if ((status != 0)); then
        return "${status}"
    fi

    oras manifest delete --force "${STATE_REFERENCE}"
    if manifest_exists; then
        echo "State manifest still exists after cleanup: ${STATE_REFERENCE}" >&2
        return 1
    else
        status=$?
    fi
    if ((status != 1)); then
        return "${status}"
    fi
}

cleanup() {
    local result=$?
    local cleanup_result=0
    set +e

    if ((result != 0)); then
        collect_diagnostics
    fi

    if [[ "${state_owned_by_run}" == "true" ]]; then
        delete_state_manifest || cleanup_result=1
    fi
    k3d cluster delete "${CONTROL_PLANE_A}" >/dev/null 2>&1
    k3d cluster delete "${CONTROL_PLANE_B}" >/dev/null 2>&1
    k3d cluster delete "${WORKLOAD_CLUSTER}" >/dev/null 2>&1
    docker rm --force "${REGISTRY_CONTAINER}" >/dev/null 2>&1
    docker network rm "${NETWORK_NAME}" >/dev/null 2>&1
    rm -rf "${WORK_DIR}"

    if ((result == 0 && cleanup_result != 0)); then
        result="${cleanup_result}"
    fi
    exit "${result}"
}

cluster_exists() {
    local cluster="$1"
    k3d cluster list --no-headers \
        | awk '{print $1}' \
        | grep -Fxq "${cluster}"
}

write_registry_config() {
    cat >"${REGISTRY_CONFIG}" <<EOF
mirrors:
  "${CLUSTER_REGISTRY}":
    endpoint:
      - "http://${CLUSTER_REGISTRY}"
EOF
}

start_registry() {
    docker network create "${NETWORK_NAME}" >/dev/null
    docker run --detach --rm \
        --name "${REGISTRY_CONTAINER}" \
        --network "${NETWORK_NAME}" \
        --network-alias "${REGISTRY_ALIAS}" \
        --publish 127.0.0.1:5000:5000 \
        registry:2 >/dev/null

    local _
    for _ in {1..30}; do
        if curl -fsS http://127.0.0.1:5000/v2/ >/dev/null; then
            return 0
        fi
        sleep 1
    done
    echo "Local OCI registry did not become ready." >&2
    return 1
}

publish_branch_artifacts() {
    local image
    for image in ucpd applications-rp dynamic-rp controller bicep; do
        docker push \
            "${DOCKER_REGISTRY}/${image}:${DOCKER_TAG_VERSION}"
    done

    cp "${SOURCE_APP_FILE}" "${APP_FILE}"
    cat >"${WORK_DIR}/bicepconfig.json" <<EOF
{
  "experimentalFeaturesEnabled": {
    "extensibility": true
  },
  "extensions": {
    "radius": "br:biceptypes.azurecr.io/radius:latest"
  }
}
EOF
    bicep build "${APP_FILE}" --stdout >/dev/null
}

create_workload_cluster() {
    k3d cluster create "${WORKLOAD_CLUSTER}" \
        --network "${NETWORK_NAME}" \
        --registry-config "${REGISTRY_CONFIG}" \
        --k3s-arg "--disable=traefik@server:*" \
        --wait

    k3d kubeconfig get "${WORKLOAD_CLUSTER}" \
        >"${HOST_WORKLOAD_KUBECONFIG}"
    cp "${HOST_WORKLOAD_KUBECONFIG}" \
        "${INTERNAL_WORKLOAD_KUBECONFIG}"

    local cluster_key
    local workload_ip
    cluster_key="$(kubectl \
        --kubeconfig "${INTERNAL_WORKLOAD_KUBECONFIG}" \
        config view --minify \
        -o jsonpath='{.contexts[0].context.cluster}')"
    workload_ip="$(docker inspect \
        --format \
        "{{(index .NetworkSettings.Networks \"${NETWORK_NAME}\").IPAddress}}" \
        "k3d-${WORKLOAD_CLUSTER}-server-0")"

    if [[ -z "${workload_ip}" ]]; then
        echo "Could not determine the workload cluster IP." >&2
        return 1
    fi

    kubectl --kubeconfig "${INTERNAL_WORKLOAD_KUBECONFIG}" \
        config set "clusters.${cluster_key}.server" \
        "https://${workload_ip}:6443" >/dev/null
    kubectl --kubeconfig "${INTERNAL_WORKLOAD_KUBECONFIG}" \
        config unset \
        "clusters.${cluster_key}.certificate-authority-data" >/dev/null
    kubectl --kubeconfig "${INTERNAL_WORKLOAD_KUBECONFIG}" \
        config set \
        "clusters.${cluster_key}.insecure-skip-tls-verify" \
        "true" >/dev/null

    kubectl --kubeconfig "${HOST_WORKLOAD_KUBECONFIG}" \
        create namespace "${WORKLOAD_NAMESPACE}"
}

configure_workspace() {
    local cluster="$1"

    kubectl config use-context "k3d-${cluster}" >/dev/null
    rad workspace create kubernetes "${WORKSPACE_NAME}" \
        --context "k3d-${cluster}" \
        --force
    rad workspace switch "${WORKSPACE_NAME}"
    rad group create default
    rad group switch default
}

install_control_plane() {
    local cluster="$1"

    k3d cluster create "${cluster}" \
        --network "${NETWORK_NAME}" \
        --registry-config "${REGISTRY_CONFIG}" \
        --k3s-arg "--disable=traefik@server:*" \
        --wait
    kubectl config use-context "k3d-${cluster}" >/dev/null

    kubectl create namespace radius-system
    kubectl create secret generic target-kubeconfig \
        --namespace radius-system \
        --from-file=kubeconfig="${INTERNAL_WORKLOAD_KUBECONFIG}"

    rad install kubernetes \
        --chart "${REPOSITORY_ROOT}/deploy/Chart" \
        --set database.enabled=true \
        --set global.targetCluster.enabled=true \
        --set rp.publicEndpointOverride=localhost \
        --set \
        "rp.image=${CLUSTER_REGISTRY}/applications-rp,rp.tag=${DOCKER_TAG_VERSION}" \
        --set \
        "dynamicrp.image=${CLUSTER_REGISTRY}/dynamic-rp,dynamicrp.tag=${DOCKER_TAG_VERSION}" \
        --set \
        "controller.image=${CLUSTER_REGISTRY}/controller,controller.tag=${DOCKER_TAG_VERSION}" \
        --set \
        "ucp.image=${CLUSTER_REGISTRY}/ucpd,ucp.tag=${DOCKER_TAG_VERSION}" \
        --set \
        "bicep.image=${CLUSTER_REGISTRY}/bicep,bicep.tag=${DOCKER_TAG_VERSION}"

    # Deployment Engine and dashboard are built in separate repositories. Their
    # chart defaults intentionally remain on the compatible edge channel.
    kubectl wait --for=condition=Available deployment --all \
        --namespace radius-system \
        --timeout=300s
    configure_workspace "${cluster}"
}

assert_application_listed() {
    local output_file="$1"

    rad app list --output json | tee "${output_file}"
    jq --exit-status --arg app "${APP_NAME}" \
        'type == "array" and
         length == 1 and
         ((.[0].name // .[0].Name // "") == $app)' \
        "${output_file}" >/dev/null
}

assert_workload_phase() {
    local phase="$1"
    local output_file="$2"
    local pod_output="${output_file%.json}-pods.json"

    kubectl --kubeconfig "${HOST_WORKLOAD_KUBECONFIG}" \
        rollout status deployment \
        --namespace "${WORKLOAD_NAMESPACE}" \
        --selector "${SELECTOR}" \
        --timeout=300s
    kubectl --kubeconfig "${HOST_WORKLOAD_KUBECONFIG}" \
        get deployment \
        --namespace "${WORKLOAD_NAMESPACE}" \
        --selector "${SELECTOR}" \
        --output json \
        | tee "${output_file}"
    kubectl --kubeconfig "${HOST_WORKLOAD_KUBECONFIG}" \
        wait --for=condition=Ready pod \
        --namespace "${WORKLOAD_NAMESPACE}" \
        --selector "${SELECTOR}" \
        --timeout=300s
    kubectl --kubeconfig "${HOST_WORKLOAD_KUBECONFIG}" \
        get pod \
        --namespace "${WORKLOAD_NAMESPACE}" \
        --selector "${SELECTOR}" \
        --output json \
        | tee "${pod_output}"
    jq --exit-status --arg phase "${phase}" \
        '(.items | length) == 1 and
         ([.items[].spec.template.spec.containers[].args[]]
          | any(contains($phase)))' \
        "${output_file}" >/dev/null
    jq --exit-status --arg phase "${phase}" \
        '(.items | length) >= 1 and
         ([.items[].spec.containers[].args[]]
          | any(contains($phase)))' \
        "${pod_output}" >/dev/null

    local pod
    pod="$(jq --raw-output '.items[0].metadata.name' "${pod_output}")"
    local _
    for _ in {1..30}; do
        if kubectl --kubeconfig "${HOST_WORKLOAD_KUBECONFIG}" \
            logs --namespace "${WORKLOAD_NAMESPACE}" "${pod}" \
            | grep -Fq "${phase}"; then
            return 0
        fi
        sleep 1
    done
    echo "The running workload did not log phase ${phase}." >&2
    return 1
}

assert_absent_from_control_plane() {
    local cluster="$1"
    local resources

    resources="$(kubectl --context "k3d-${cluster}" get all \
        --all-namespaces \
        --selector "${SELECTOR}" \
        --output name)"
    if [[ -n "${resources}" ]]; then
        echo "Workload resources unexpectedly exist on ${cluster}:" >&2
        echo "${resources}" >&2
        return 1
    fi
}

deploy_phase() {
    local phase="$1"
    local environment_id

    environment_id="$(rad env show "${ENVIRONMENT_NAME}" \
        --output json | jq --raw-output '.id // .Id')"
    if [[ -z "${environment_id}" || "${environment_id}" == "null" ]]; then
        echo "Could not resolve the Radius environment ID." >&2
        return 1
    fi

    rad deploy "${APP_FILE}" \
        --parameters "environment=${environment_id}" \
        --parameters "deploymentPhase=${phase}"
}

manifest_exists() {
    local output

    if output="$(oras manifest fetch --descriptor \
        "${STATE_REFERENCE}" 2>&1)"; then
        return 0
    fi
    if grep -Eqi '(^|[^0-9])(404|not found)([^0-9]|$)' <<<"${output}"; then
        return 1
    fi

    echo "${output}" >&2
    return 2
}

assert_state_absent() {
    local status

    if manifest_exists; then
        echo "Run-specific state already exists: ${STATE_REFERENCE}" >&2
        return 1
    else
        status=$?
    fi
    if ((status != 1)); then
        return "${status}"
    fi
}

state_digest() {
    oras manifest fetch --descriptor "${STATE_REFERENCE}" \
        | jq --exit-status --raw-output '.digest'
}

main() {
    trap cleanup EXIT
    mkdir -p "${WORK_DIR}" "${DIAGNOSTICS_DIR}"
    write_registry_config
    start_registry
    publish_branch_artifacts
    create_workload_cluster
    assert_state_absent

    install_control_plane "${CONTROL_PLANE_A}"
    # Repo Radius always invokes startup. Its first-run empty archive is a
    # deliberate no-op and proves the workflow needs no special first-run path.
    rad startup
    rad env create "${ENVIRONMENT_NAME}" \
        --kubernetes-namespace "${WORKLOAD_NAMESPACE}"
    deploy_phase "before-restore"
    assert_application_listed \
        "${DIAGNOSTICS_DIR}/apps-before-restore.json"
    assert_workload_phase "before-restore" \
        "${DIAGNOSTICS_DIR}/workload-before-restore.json"
    assert_absent_from_control_plane "${CONTROL_PLANE_A}"

    rad shutdown
    state_owned_by_run=true
    local saved_digest
    saved_digest="$(state_digest)"
    echo "Saved Radius state as ${saved_digest}."

    k3d cluster delete "${CONTROL_PLANE_A}"
    if cluster_exists "${CONTROL_PLANE_A}"; then
        echo "The first control plane was not deleted." >&2
        return 1
    fi

    install_control_plane "${CONTROL_PLANE_B}"
    rad startup
    assert_application_listed \
        "${DIAGNOSTICS_DIR}/apps-after-restore.json"

    local restored_digest
    restored_digest="$(state_digest)"
    if [[ "${restored_digest}" != "${saved_digest}" ]]; then
        echo "State digest changed during restore." >&2
        return 1
    fi

    deploy_phase "after-restore"
    assert_workload_phase "after-restore" \
        "${DIAGNOSTICS_DIR}/workload-after-restore.json"
    assert_absent_from_control_plane "${CONTROL_PLANE_B}"

    echo "Repo Radius GHCR state rehydration succeeded."
}

if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then
    main "$@"
fi
