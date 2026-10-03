# Getting Started

ลำดับ deploy ตั้งแต่ศูนย์ ค่าและคำสั่งที่ใช้จริงอยู่ในไฟล์ที่อ้างถึง ผลทดสอบของแต่ละขั้นอยู่ที่ [build-checklist.md](build-checklist.md)

## สิ่งที่ต้องมี

- AWS account + AWS CLI v2 ที่ login แล้ว (และ Session Manager plugin)
- Terraform ≥ 1.10 (ใช้ `use_lockfile` ของ S3 backend)
- `kubectl`, `helm`, `gh` (GitHub CLI), Docker
- GitHub repo ที่เป็นของตัวเอง (repo เป็น public: Argo CD อ่านได้โดยไม่ต้องมี credential)
- โดเมนไม่จำเป็น: ใช้ `<ip-คั่นด้วยขีด>.sslip.io` เช่น `app.52-74-96-78.sslip.io`
- บัญชี AWS ใหม่บางครั้งจำกัดแค่ Free Tier instance type ตอนนี้ค่าเริ่มต้นคือ `m7i-flex.large` (8 GB) ปรับได้ที่ `instance_type`

## 1. รันบนเครื่องตัวเองก่อน

`docker compose up --build` รัน frontend, backend และ MinIO (จำลอง S3 เฉพาะบนเครื่อง) เปิด `http://localhost:3000` แล้วแปลงรูป ก.พ. ได้โดยยังไม่แตะ AWS บน cluster backend ใช้ Amazon S3 จริงด้วย IAM role ของเครื่อง (ไม่ตั้ง `S3_ENDPOINT`)

## 2. state bucket (ทำด้วยมือครั้งเดียว)

สร้าง S3 bucket เก็บ Terraform state แล้วเปิด versioning เหตุผลที่สร้างด้วย Terraform ไม่ได้: [infrastructure.md](infrastructure.md#เหตุผลที่ต้องสร้าง-state-bucket-ด้วยมือ)

## 3. สร้าง infrastructure

```bash
cp iac/terraform.tfvars.example iac/terraform.tfvars   # ใส่ admin_cidr (IP ตัวเอง /32) และ budget_alert_email
cd iac && terraform init && terraform plan && terraform apply
terraform plan        # ต้องได้ "No changes"
```
`terraform.tfvars` ถูก gitignore (มี IP และอีเมล) เมื่อ IP ที่บ้านเปลี่ยน ต้องแก้ `admin_cidr` แล้ว `apply` ใหม่ ไม่งั้น `kubectl` เข้า port 6443 ไม่ได้

## 4. kubeconfig (ผ่าน SSM ไม่ใช้ SSH)

รอ ~3–5 นาทีให้ `user_data` ติดตั้ง K3s เสร็จ แล้วอ่าน `/etc/rancher/k3s/k3s.yaml` ด้วย `aws ssm send-command` (`AWS-RunShellScript`) แทน `127.0.0.1` ด้วย Elastic IP เก็บเป็น `~/.kube/thai-gov-k3s.yaml` ทุก terminal ใหม่ต้อง `export KUBECONFIG=~/.kube/thai-gov-k3s.yaml` ก่อนใช้ `kubectl`/`helm` (ลืมแล้วจะขึ้น `x509: certificate signed by unknown authority` เพราะไปคุยกับ `localhost:8080`) ไฟล์นี้เปลี่ยนทุกครั้งที่สร้าง cluster ใหม่

## 5. Secrets และ platform

```bash
export GITHUB_TOKEN=...  DISCORD_WEBHOOK_URL=...      # ใส่ช่องว่างนำหน้าคำสั่ง ไม่ให้เข้า history
export TF_READONLY_KEY_ID=...  TF_READONLY_SECRET=... # aws iam create-access-key --user-name tf-readonly
./scripts/create-ci-secrets.sh    # สร้าง k8s Secret jenkins-ci-secrets (และพิมพ์ webhook secret ใหม่)
./scripts/bootstrap-cluster.sh    # cert-manager, Argo CD (+ notifications, Application ของ prod และ staging), Jenkins
```

| สิ่งที่ `bootstrap` ติดตั้ง | หมายเหตุ |
| --- | --- |
| cert-manager + ClusterIssuer (Let's Encrypt staging/prod) | HTTP-01 ผ่าน Traefik |
| Argo CD | ใช้ `--server-side` เพราะ CRD ApplicationSet ใหญ่เกิน limit ของ client-side apply |
| Argo CD notifications | copy Discord webhook จาก `jenkins-ci-secrets` ไปเป็น `argocd-notifications-secret` |
| Application `thai-gov` (`main`) และ `thai-gov-staging` (`staging`) | staging จะ error จนกว่าจะสร้าง branch `staging` (ขั้น 6) |
| Jenkins (Helm chart ปักเวอร์ชัน) | ตั้งค่าทั้งหมดผ่าน JCasC ใน `k8s/platform/jenkins-values.yaml`: credentials, multibranch job, job `terraform-drift` |

Jenkins ใช้เวลา ~10–15 นาทีโหลด plugin ครั้งแรก (`--wait`)

## 6. GitHub

| ตั้งค่า | ค่า |
| --- | --- |
| Branch `dev` และ `staging` | `git push origin main:dev main:staging` |
| Rulesets (`protect-main`, `protect-dev`, `protect-staging`) | ต้อง PR · เช็ค `continuous-integration/jenkins/pr-merge` · merge-commit อย่างเดียว · ห้ามลบ/force push · bypass: Repository admin (ให้บอทที่ใช้ PAT ของ admin push tag ได้) |
| Webhook | `https://jenkins.<ip>.sslip.io/github-webhook/` · JSON · secret = ค่าใน `jenkins-ci-secrets` (key `github-webhook-secret`) ส่วนที่เหลือของ Jenkins ไม่เปิดสาธารณะ |
| Fine-grained token (PAT) | เฉพาะ repo นี้: Contents (read/write), Pull requests (read), Commit statuses (read/write), Metadata (read) |
| Repository settings | ปิด squash/rebase merge · ติ๊ก "Automatically delete head branches" |

หลังสร้าง cluster ใหม่ webhook secret ถูกสร้างใหม่ ต้อง PATCH webhook ใน GitHub ด้วยค่าใหม่

## 7. Deploy ครั้งแรก

เปิด PR เล็กๆ เข้า `dev` → รอเช็คเขียว → merge (Jenkins `build`: push image, bump tag ของ staging) → PR `dev → staging` → merge (Argo sync staging) → PR `staging → main` → merge (Jenkins `promote`, Argo sync prod) ดู [cicd-pipeline.md](cicd-pipeline.md)

## 8. ประหยัดค่าใช้จ่ายหลังเดโม

`terraform destroy -target=aws_instance.app` ทำลายเฉพาะเครื่อง เก็บ Elastic IP, ECR, S3 และ state ไว้ (สร้างเครื่องคืนจะได้ IP เดิม ไม่ต้องแก้ hostname) หรือ `terraform destroy` ทั้งชุดถ้าไม่ใช้อีก
