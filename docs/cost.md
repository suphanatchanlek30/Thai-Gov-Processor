# ค่าใช้จ่าย

ตัวเลขเป็นค่าประมาณสำหรับ region Singapore ตรวจราคาล่าสุดด้วย [AWS Pricing Calculator](https://calculator.aws/) ก่อนใช้งานจริง

| รายการ | รายละเอียด | หมายเหตุ |
| --- | --- | --- |
| EC2 `m7i-flex.large` | 2 vCPU / 8 GB, on-demand | คิดเป็นชั่วโมง ค่าใช้จ่ายหลักของโปรเจกต์ ไม่ใช่ `t3a.large` เพราะบัญชีนี้ติดข้อจำกัด Free Tier ชั่วคราว (ปรับที่ `instance_type`) |
| EBS gp3 35 GB | ดิสก์ของเครื่อง | ถูกลบพร้อมเครื่อง (`terraform destroy -target`) |
| Public IPv4 (Elastic IP) | $0.005/ชม. | คิดแม้ปิดเครื่อง ถ้ายังถือ IP ไว้ |
| S3 + ECR | ไฟล์ชั่วคราวและ image < 1 GB | เกือบ $0 |
| EKS control plane | ไม่ได้ใช้ (ใช้ K3s แทน) | ประหยัด ~$73/เดือน |
| AWS Budgets | 2 budget แรกต่อบัญชีฟรี | โปรเจกต์นี้ใช้ 1 |

> **ปิดโปรเจกต์แล้ว:** ลบทุกอย่างบน AWS แล้ว (ตรวจว่าไม่เหลือ EC2, Elastic IP, EBS, ECR, S3, IAM, state bucket) ค่าใช้จ่ายของโปรเจกต์นี้จึงเป็น $0 ตารางด้านล่างคือค่าประมาณตอนที่ยังรันอยู่

## วิธีคุมงบ

- **เปิดเครื่องเฉพาะตอนใช้:** `terraform destroy -target=aws_instance.app` หลังเดโมจะเหลือ Elastic IP (~$3.6/เดือน คิดต่อชั่วโมงแม้ไม่ได้ผูกกับเครื่อง), ECR และ S3 (เกือบ $0) รวมราว $4/เดือน สร้างเครื่องคืนด้วย `terraform apply` ได้ IP และ hostname เดิม ถ้าไม่ใช้อีกเลยให้ `terraform destroy` ทั้งชุด
- **AWS Budgets** (`iac/budget.tf`): งบ $10/เดือน เตือนทางอีเมลเมื่อค่าใช้จ่ายจริงเกิน 80% และเมื่อ forecast ว่าจะเกิน 100% อีเมลตั้งใน `terraform.tfvars` (`budget_alert_email`) ไม่อยู่ใน repo
- state bucket และ ECR ที่เหลือหลัง destroy มีค่าใช้จ่ายแทบเป็นศูนย์ (ECR เก็บแค่ 10 image ล่าสุดต่อ repo)
- Jenkins รันได้ทีละ build จึงไม่ต้องเพิ่มเครื่องเพื่อรองรับ build ซ้อน
