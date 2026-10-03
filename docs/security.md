# Security และ PDPA

## ใครมีสิทธิ์ทำอะไร

| ใคร | เข้าถึงอะไร | ยืนยันตัวตนด้วย | secret อยู่ที่ไหน |
| --- | --- | --- | --- |
| GitHub → Jenkins | ส่ง webhook | HMAC signature | Jenkins Credentials |
| Jenkins → ECR | push image | IAM role ของ EC2 | ไม่มี key |
| Jenkins → GitHub | commit tag ใหม่ | fine-grained token เฉพาะ repo | Jenkins Credentials |
| Jenkins → AWS (plan/drift) | อ่านอย่างเดียว | IAM user `tf-readonly` | Jenkins Credentials |
| Argo CD → GitHub | อ่าน repo | ไม่ต้องใช้ (repo เป็น public อ่านผ่าน https) | ไม่มี |
| Argo CD → K3s | เขียน ns `thai-gov` และ `thai-gov-staging` | ServiceAccount + RBAC | ภายใน cluster |
| Argo CD / Jenkins → Discord | ส่งแจ้งเตือน | URL ของ webhook | k8s Secret (`argocd-notifications-secret`, `jenkins-ci-secrets`) · ห้ามพิมพ์ลง log |
| บอท Jenkins → GitHub | push commit `deploy … [skip ci]` ผ่านกฎของ branch | token ของ admin (fine-grained) + bypass ใน ruleset | Jenkins Credentials |
| K3s → ECR | pull image | ecr-credential-provider + IAM role | ไม่มี key |
| Backend → S3 | Get/Put เฉพาะ bucket เดียว (ไม่มี Delete ตั้งใจ ให้ lifecycle ลบ) | IAM role ของเครื่อง | ไม่มี key |
| ผู้ดูแล → EC2 | shell | SSM Session Manager | ไม่ต้องมี SSH key |

## Security controls ตามชั้น

| ชั้น | สิ่งที่ทำ |
| --- | --- |
| Git | rulesets 3 branch: ต้องผ่าน PR + required check, merge-commit อย่างเดียว, ห้ามลบ/force push · ด่านต้นทาง branch (เข้า `staging` ต้องมาจาก `dev`, เข้า `main` ต้องมาจาก `staging`) |
| โค้ด | lint, unit test, Trivy secret scan (กัน AWS key หลุดเข้า Git) |
| Supply chain | Trivy gate สองรอบ, push ไฟล์ตัวเดียวกับที่สแกน, ECR IMMUTABLE + scan on push, pin เวอร์ชัน image ใน CI |
| Container | non-root, read-only root filesystem, drop capabilities ทั้งหมด, resource limits |
| Cluster | CI ไม่มีสิทธิ์ใน production, มีแค่ Argo CD ที่ deploy ได้, self-heal กันแก้ด้วยมือ |
| Network | เปิดแค่ 80/443, K3s API เฉพาะ IP ผู้ดูแล, ไม่เปิด SSH · Jenkins เปิดสาธารณะเฉพาะ `/github-webhook/` (ตรวจ HMAC) ส่วน UI ของ Jenkins และ Argo CD เข้าผ่าน `kubectl port-forward` |
| Host | IMDSv2 บังคับ, EBS เข้ารหัส, Ubuntu LTS |
| Data | S3 block public access, SSE, presigned URL อายุ 1 ชม., ลบ EXIF |

## PDPA: ข้อมูลอยู่นานแค่ไหน

| ข้อมูล | อยู่ที่ไหน | อยู่นานเท่าไร |
| --- | --- | --- |
| ไฟล์ต้นฉบับ | หน่วยความจำของ backend ระหว่าง request (ไม่ถูกเก็บลง S3) | หายเมื่อ request จบ |
| ไฟล์ที่แปลงแล้ว | `processed/` ใน S3 | ลิงก์ดาวน์โหลดใช้ได้ 1 ชม., ไฟล์ถูกลบโดย lifecycle (1 วัน) |
| ไฟล์ชั่วคราว | `/tmp` ที่เป็น `emptyDir` ของ pod (ใช้เมื่อ multipart ใหญ่เกินที่เก็บในหน่วยความจำ) | หายเมื่อ pod ถูกลบ |
| Log | stdout ของ pod | ไม่ log ชื่อไฟล์เดิมหรือเนื้อหาไฟล์ |

S3 นับวันหมดอายุโดยปัดไปเที่ยงคืน UTC และลบแบบ asynchronous ไฟล์ผลลัพธ์จึงอาจค้างอยู่ประมาณ 1–2 วัน ไม่ใช่ 24 ชั่วโมงเป๊ะ ส่วนไฟล์ต้นฉบับไม่เคยถูกเก็บลง S3 จึงไม่มีความเสี่ยงส่วนนี้ (ข้อมูลที่ค้างได้นานที่สุดคือผลลัพธ์ที่ผู้ใช้เพิ่งขอ)
