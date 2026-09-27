# Security และ PDPA

## ใครมีสิทธิ์ทำอะไร

| ใคร | เข้าถึงอะไร | ยืนยันตัวตนด้วย | secret อยู่ที่ไหน |
| --- | --- | --- | --- |
| GitHub → Jenkins | ส่ง webhook | HMAC signature | Jenkins Credentials |
| Jenkins → ECR | push image | IAM role ของ EC2 | ไม่มี key |
| Jenkins → GitHub | commit tag ใหม่ | fine-grained token เฉพาะ repo | Jenkins Credentials |
| Jenkins → AWS (plan/drift) | อ่านอย่างเดียว | IAM user `tf-readonly` | Jenkins Credentials |
| Argo CD → GitHub | อ่าน repo | deploy key แบบ read-only | K8s Secret (ns argocd) |
| Argo CD → K3s | เขียน ns `production` | ServiceAccount + RBAC | ภายใน cluster |
| K3s → ECR | pull image | ecr-credential-provider + IAM role | ไม่มี key |
| Backend → S3 | Put/Get/Delete เฉพาะ bucket เดียว | IAM role | ไม่มี key |
| ผู้ดูแล → EC2 | shell | SSM Session Manager | ไม่ต้องมี SSH key |

## Security controls ตามชั้น

| ชั้น | สิ่งที่ทำ |
| --- | --- |
| โค้ด | lint, unit test, Trivy secret scan (กัน AWS key หลุดเข้า Git) |
| Supply chain | Trivy gate สองรอบ, push ไฟล์ตัวเดียวกับที่สแกน, ECR IMMUTABLE + scan on push, pin เวอร์ชัน image ใน CI |
| Container | non-root, read-only root filesystem, drop capabilities ทั้งหมด, resource limits |
| Cluster | CI ไม่มีสิทธิ์ใน production, มีแค่ Argo CD ที่ deploy ได้, self-heal กันแก้ด้วยมือ |
| Network | เปิดแค่ 80/443, K3s API เฉพาะ IP ผู้ดูแล, ไม่เปิด SSH |
| Host | IMDSv2 บังคับ, EBS เข้ารหัส, Ubuntu LTS |
| Data | S3 block public access, SSE, presigned URL อายุ 1 ชม., ลบ EXIF |

## PDPA: ข้อมูลอยู่นานแค่ไหน

| ข้อมูล | อยู่ที่ไหน | อยู่นานเท่าไร |
| --- | --- | --- |
| ไฟล์ต้นฉบับ | `uploads/` ใน S3 | backend ลบทันทีหลังประมวลผลเสร็จ, lifecycle เป็นตัวสำรอง |
| ไฟล์ที่แปลงแล้ว | `processed/` ใน S3 | ลิงก์ดาวน์โหลดใช้ได้ 1 ชม., ไฟล์ถูกลบโดย lifecycle |
| ไฟล์ระหว่างประมวลผล | `emptyDir` ใน pod | หายไปเมื่อ request จบ หรือเมื่อ pod ถูกลบ |
| Log | stdout ของ pod | ไม่ log ชื่อไฟล์เดิมหรือเนื้อหาไฟล์ |

S3 นับวันหมดอายุโดยปัดไปเที่ยงคืน UTC และลบแบบ asynchronous ไฟล์จึงอาจค้างอยู่ประมาณ 1–2 วัน ไม่ใช่ 24 ชั่วโมงเป๊ะ ด้วยเหตุนี้ backend จึงลบไฟล์ต้นฉบับเองทันที และใช้ lifecycle เป็นตาข่ายรองรับอีกชั้น
