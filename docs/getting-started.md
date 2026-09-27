# Getting Started

เอกสารนี้บอกลำดับขั้นและเป้าหมายของแต่ละขั้นเท่านั้น ไม่ใส่โค้ดจริง ให้เขียนตาม [build-checklist.md](build-checklist.md) เป็นหลัก

## สิ่งที่ต้องมี

- AWS account + AWS CLI v2 ที่ login แล้ว
- Terraform ≥ 1.10 (ต้องการฟีเจอร์ `use_lockfile` ของ S3 backend)
- Session Manager plugin ของ AWS CLI
- Domain ที่ตั้งค่า DNS ได้ (หรือใช้ `sslip.io` ระหว่างทดสอบ)
- Docker สำหรับรันบนเครื่องตัวเอง

## 1. รันบนเครื่องตัวเองให้ได้ก่อน

ใช้ Docker Compose รัน frontend, backend และ MinIO (ใช้แทน S3) ในเครื่องเดียว เป้าหมายคือเปิด `localhost:3000` แล้วแปลงรูป ก.พ. ได้จริงโดยยังไม่แตะ AWS

## 2. เตรียม state bucket (ทำด้วยมือครั้งเดียว)

สร้าง S3 bucket สำหรับเก็บ Terraform state แล้วเปิด versioning — เหตุผลที่ทำด้วยมือ ดู [infrastructure.md](infrastructure.md#เหตุผลที่ต้องสร้าง-state-bucket-ด้วยมือ)

## 3. สร้าง infrastructure

ใส่ IP ของตัวเองในไฟล์ `terraform.tfvars` จากนั้นรัน `init` → `plan` (อ่านให้เข้าใจว่าจะสร้างอะไร) → `apply` เป้าหมายคือรัน `plan` ซ้ำหลัง apply แล้วต้องขึ้นว่า *No changes*

## 4. เข้าเครื่องและติดตั้ง platform

เข้าเครื่องผ่าน SSM Session Manager (ไม่ใช้ SSH) แล้วติดตั้งตามลำดับนี้:

1. **ecr-credential-provider** ให้ K3s ดึง image จาก ECR ได้ตลอด เพราะ token ของ ECR อายุแค่ 12 ชม.
2. **cert-manager** + ClusterIssuer ของ Let's Encrypt
3. **Argo CD** + Application ที่ชี้ไปที่ `k8s/overlays/prod`
4. **Jenkins** ผ่าน Helm chart ทางการ

## 5. ตั้งค่า GitHub

| ตั้งค่า | ค่า |
| --- | --- |
| Branch protection (`main`) | ต้องผ่าน PR, ต้องผ่าน status check ของ Jenkins, ห้าม force push |
| Webhook | ชี้ไปที่ `https://ci.<domain>/github-webhook/`, content type JSON, ตั้ง secret |
| Fine-grained token (ให้ Jenkins) | เฉพาะ repo นี้: Contents (read/write), Pull requests (read), Commit statuses (read/write) |
| Deploy key (ให้ Argo CD) | read-only |

## 6. ตั้งค่า Jenkins Credentials

| ID | ชนิด | ใช้ทำอะไร |
| --- | --- | --- |
| `github-app-token` | Username with password | สแกน repo และ push commit ที่แก้ tag |
| `github-webhook-secret` | Secret text | ตรวจลายเซ็น webhook |
| `discord-webhook` | Secret text | ส่งแจ้งเตือน |
| `aws-tf-readonly` | Username with password (access key / secret) | `terraform plan` และ drift detect (สิทธิ์ ReadOnlyAccess) |

จากนั้นสร้าง Multibranch Pipeline ชี้มาที่ repo นี้ และเปิดตัวเลือก "Discover pull requests from origin"

## 7. Deploy ครั้งแรก

เปิด PR เล็กๆ รอให้ขึ้น ✔ แล้ว merge จากนั้นดูต่อที่ Jenkins → Argo CD → เว็บจริง

## 8. ลบทิ้งหลังเดโม

`terraform destroy` ทันทีที่เดโมเสร็จ state bucket จะยังอยู่ ครั้งหน้า apply ใหม่ได้เลย
