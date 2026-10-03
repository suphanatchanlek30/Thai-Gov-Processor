#!/usr/bin/env bash
# Installs the cluster platform pieces on top of a fresh K3s node.
# Run from the admin machine with KUBECONFIG pointing at the cluster.
set -euo pipefail

CERT_MANAGER_VERSION="v1.21.2"
ARGOCD_VERSION="v3.5.3"
JENKINS_CHART_VERSION="5.9.64"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"

install_cert_manager() {
  kubectl apply -f "https://github.com/cert-manager/cert-manager/releases/download/${CERT_MANAGER_VERSION}/cert-manager.yaml"

  # ClusterIssuer is validated by cert-manager's webhook, which is not
  # ready the instant the manifest applies - creating the issuer too early
  # fails with "connection refused" from the webhook.
  kubectl -n cert-manager rollout status deploy/cert-manager-webhook --timeout=180s

  kubectl apply -f "${ROOT}/k8s/platform/cluster-issuer.yaml"
}

install_argocd() {
  kubectl create namespace argocd --dry-run=client -o yaml | kubectl apply -f -

  # --server-side: the ApplicationSet CRD is larger than the 256 KB
  # annotation limit that client-side apply uses to track changes, so a
  # plain `kubectl apply` fails on it.
  kubectl apply -n argocd --server-side --force-conflicts \
    -f "https://raw.githubusercontent.com/argoproj/argo-cd/${ARGOCD_VERSION}/manifests/install.yaml"

  kubectl -n argocd rollout status deploy/argocd-server --timeout=300s
}

install_argocd_notifications() {
  if ! kubectl -n jenkins get secret jenkins-ci-secrets >/dev/null 2>&1; then
    echo "Secret jenkins-ci-secrets is missing: run scripts/create-ci-secrets.sh first" >&2
    exit 1
  fi

  # Reuse the Discord webhook already held by jenkins-ci-secrets. The value
  # is passed through still base64-encoded on stdin, never decoded or put
  # on a command line.
  local discord_b64
  discord_b64="$(kubectl -n jenkins get secret jenkins-ci-secrets -o jsonpath='{.data.discord-webhook}')"
  kubectl apply -f - <<SECRET
apiVersion: v1
kind: Secret
metadata:
  name: argocd-notifications-secret
  namespace: argocd
data:
  discord-webhook-url: ${discord_b64}
SECRET

  kubectl apply --server-side --force-conflicts -f "${ROOT}/k8s/platform/argocd-notifications.yaml"
}

install_argocd_app() {
  kubectl apply -f "${ROOT}/k8s/platform/argocd-application.yaml"
}

install_jenkins() {
  # The controller mounts this Secret on startup, so it must exist before
  # the chart is installed or the pod never becomes ready.
  if ! kubectl -n jenkins get secret jenkins-ci-secrets >/dev/null 2>&1; then
    echo "Secret jenkins-ci-secrets is missing: run scripts/create-ci-secrets.sh first" >&2
    exit 1
  fi

  helm repo add jenkins https://charts.jenkins.io --force-update >/dev/null
  # --wait blocks until the controller passes its readiness probe, which
  # only happens after every plugin has been downloaded and loaded.
  helm upgrade --install jenkins jenkins/jenkins \
    --version "${JENKINS_CHART_VERSION}" \
    --namespace jenkins --create-namespace \
    --values "${ROOT}/k8s/platform/jenkins-values.yaml" \
    --wait --timeout 15m
}

install_cert_manager
install_argocd
install_argocd_notifications
install_argocd_app
install_jenkins
