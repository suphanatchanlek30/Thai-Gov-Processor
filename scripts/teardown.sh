#!/usr/bin/env bash
# Deletes the whole thai-gov-processor project from AWS, state bucket included.
# Irreversible. `terraform destroy` alone fails on a non-empty ECR repo or S3
# bucket and on an IAM user that has hand-made access keys, so those go first.
set -euo pipefail

REGION=ap-southeast-1
ACCOUNT="$(aws sts get-caller-identity --query Account --output text)"
STATE_BUCKET="thai-gov-processor-tfstate-${ACCOUNT}"
UPLOADS_BUCKET="thai-gov-processor-uploads-${ACCOUNT}"
IAC="$(cd "$(dirname "$0")/../iac" && pwd)"

cat <<MSG
This DELETES the thai-gov-processor project from AWS account ${ACCOUNT} (${REGION}):
  EC2 + EBS, Elastic IP, VPC, security group, IAM role/user, ECR repos and images,
  the uploads bucket and its files, and the Terraform state bucket.
Other projects in the account are not touched.
MSG
read -r -p "Type DELETE to continue: " answer
[ "${answer}" = "DELETE" ] || { echo "Cancelled."; exit 1; }

echo "1/6 Deleting the ECR repositories with their images (--force also removes untagged manifests)"
for repo in thai-gov-processor-backend thai-gov-processor-frontend; do
  aws ecr delete-repository --region "${REGION}" --repository-name "${repo}" --force >/dev/null 2>&1 || true
done

echo "2/6 Emptying the uploads bucket"
aws s3 rm "s3://${UPLOADS_BUCKET}" --recursive --region "${REGION}" || true

echo "3/6 Deleting the access keys of the tf-readonly user (created by hand, so terraform cannot delete the user)"
for key in $(aws iam list-access-keys --user-name tf-readonly --query 'AccessKeyMetadata[].AccessKeyId' --output text 2>/dev/null); do
  aws iam delete-access-key --user-name tf-readonly --access-key-id "${key}"
done

echo "4/6 terraform destroy (it will ask you to type yes)"
cd "${IAC}"
terraform destroy

echo "5/6 Deleting the Terraform state bucket (versioned, so every version goes first)"
python3 - "${STATE_BUCKET}" <<'PY'
import json, subprocess, sys
bucket = sys.argv[1]
out = json.loads(subprocess.check_output(
    ["aws", "s3api", "list-object-versions", "--bucket", bucket, "--output", "json"]))
objs = [{"Key": o["Key"], "VersionId": o["VersionId"]}
        for o in out.get("Versions", []) + out.get("DeleteMarkers", [])]
if objs:
    subprocess.check_call(["aws", "s3api", "delete-objects", "--bucket", bucket,
                           "--delete", json.dumps({"Objects": objs, "Quiet": True})],
                          stdout=subprocess.DEVNULL)
subprocess.check_call(["aws", "s3api", "delete-bucket", "--bucket", bucket])
print("state bucket deleted")
PY

echo "6/6 Checking that nothing is left (every line below should be empty)"
echo "EC2 instances : $(aws ec2 describe-instances --region "${REGION}" --filters Name=instance-state-name,Values=pending,running,stopping,stopped --query 'Reservations[].Instances[].InstanceId' --output text)"
echo "Elastic IPs   : $(aws ec2 describe-addresses --region "${REGION}" --query 'Addresses[].PublicIp' --output text)"
echo "EBS volumes   : $(aws ec2 describe-volumes --region "${REGION}" --query 'Volumes[].VolumeId' --output text)"
echo "ECR repos     : $(aws ecr describe-repositories --region "${REGION}" --query 'repositories[?contains(repositoryName,`thai-gov`)].repositoryName' --output text)"
echo "S3 buckets    : $(aws s3api list-buckets --query 'Buckets[?contains(Name,`thai-gov`)].Name' --output text)"
echo "IAM tf-readonly: $(aws iam list-users --query 'Users[?UserName==`tf-readonly`].UserName' --output text)"
echo "Done. Check the AWS Billing page tomorrow to confirm the charges stopped."
