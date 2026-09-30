#!/usr/bin/env bash
# Installs the cluster platform pieces on top of a fresh K3s node.
# Run from the admin machine with KUBECONFIG pointing at the cluster.
set -euo pipefail

CERT_MANAGER_VERSION="v1.21.2"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"

install_cert_manager() {
  kubectl apply -f "https://github.com/cert-manager/cert-manager/releases/download/${CERT_MANAGER_VERSION}/cert-manager.yaml"

  # ClusterIssuer is validated by cert-manager's webhook, which is not
  # ready the instant the manifest applies - creating the issuer too early
  # fails with "connection refused" from the webhook.
  kubectl -n cert-manager rollout status deploy/cert-manager-webhook --timeout=180s

  kubectl apply -f "${ROOT}/k8s/platform/cluster-issuer.yaml"
}

install_cert_manager
