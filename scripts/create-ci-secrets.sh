#!/usr/bin/env bash
# Creates the Secret that Jenkins' JCasC config reads its credentials from
# (k8s/platform/jenkins-values.yaml). Run it before the Jenkins chart is
# installed: the controller pod mounts the Secret when it starts. Values come
# from the environment so nothing secret lands in the repo or shell history
# (put a leading space before the command, or use `read -s`).
#
#   GITHUB_TOKEN          required  fine-grained PAT for the repo (contents RW, PRs/metadata R)
#   DISCORD_WEBHOOK_URL   required
#   WEBHOOK_SECRET        optional  generated and printed when unset; must match the GitHub webhook
#   TF_READONLY_KEY_ID / TF_READONLY_SECRET
#                         optional  access key of the tf-readonly IAM user
#
# Rotating a value means re-running this and restarting the controller pod.
set -euo pipefail

: "${GITHUB_TOKEN:?set GITHUB_TOKEN}"
: "${DISCORD_WEBHOOK_URL:?set DISCORD_WEBHOOK_URL}"

ROOT="$(cd "$(dirname "$0")/.." && pwd)"

if [[ -z "${WEBHOOK_SECRET:-}" ]]; then
  WEBHOOK_SECRET="$(openssl rand -hex 24)"
  echo "Generated webhook secret (paste it into the GitHub webhook): ${WEBHOOK_SECRET}"
fi

ADMIN_CIDR="$(sed -n 's/^admin_cidr *= *"\(.*\)"/\1/p' "${ROOT}/iac/terraform.tfvars")"
: "${ADMIN_CIDR:?admin_cidr not found in iac/terraform.tfvars}"

# --from-env-file keeps the values out of kubectl's argv.
env_file="$(mktemp)"
trap 'rm -f "${env_file}"' EXIT
chmod 600 "${env_file}"
cat > "${env_file}" <<ENV
github-token=${GITHUB_TOKEN}
github-webhook-secret=${WEBHOOK_SECRET}
discord-webhook=${DISCORD_WEBHOOK_URL}
aws-tf-readonly-key-id=${TF_READONLY_KEY_ID:-unset}
aws-tf-readonly-secret=${TF_READONLY_SECRET:-unset}
tf-admin-cidr=${ADMIN_CIDR}
ENV

kubectl create namespace jenkins --dry-run=client -o yaml | kubectl apply -f -
kubectl -n jenkins create secret generic jenkins-ci-secrets \
  --from-env-file="${env_file}" --dry-run=client -o yaml | kubectl apply -f -
