# Tech Stack

## หน้าที่ของแต่ละเครื่องมือในระบบ

| เครื่องมือ | หน้าที่ในโปรเจกต์ |
| --- | --- |
| Next.js | Frontend สำหรับ Upload, เลือก Preset, Crop และ Preview |
| Go | Backend API สำหรับ resize, compress และสร้างไฟล์ผลลัพธ์ |
| Gin | HTTP web framework ของ Backend — routing, middleware, request binding |
| S3 | เก็บไฟล์ผู้ใช้และไฟล์ผลลัพธ์แบบชั่วคราว |
| Terraform | สร้าง AWS Infrastructure จาก Code |
| EC2 | เครื่องหลักที่ใช้รัน K3s |
| K3s | Kubernetes runtime สำหรับ Frontend, Backend, Jenkins และ Argo CD |
| Traefik | รับ request จาก Internet แล้ว route ไป Frontend / Backend / Jenkins |
| Jenkins | Continuous Integration: lint, test, build, scan และ push image |
| BuildKit | Build container image แบบ rootless |
| Trivy | Security Scan และบล็อก pipeline เมื่อพบช่องโหว่ระดับ CRITICAL |
| ECR | Container Registry สำหรับเก็บ Frontend / Backend images |
| Argo CD | Continuous Delivery แบบ GitOps (2 Applications: prod และ staging) พร้อม Notifications ส่ง Discord |
| Kustomize | จัดการ Kubernetes manifest: `base` + overlay `prod` / `staging` และ image tag |
| JCasC + Job DSL | ตั้งค่า Jenkins ทั้งหมดเป็นโค้ด (credentials, multibranch job, job drift) สร้างใหม่ได้เมื่อ rebuild cluster |
| Pipeline Graph View / Stage View | ดูแต่ละ run เป็น stage พร้อมเวลา |
| GitHub rulesets | บังคับ PR, required check, merge-commit อย่างเดียว ให้ `dev` / `staging` / `main` |
| cert-manager | จัดการ TLS certificate จาก Let's Encrypt |
| SSM Session Manager | ใช้เข้าถึง EC2 โดยไม่ต้องเปิด SSH port 22 |
| Discord | รับแจ้งเตือน build / deploy / drift |

## ทำไมเลือกใช้ตัวนี้ ไม่ใช้ตัวอื่น

| ส่วนประกอบ | เลือกใช้ | เหตุผล | ทางเลือกที่ไม่เลือก และเหตุผล |
| --- | --- | --- | --- |
| Backend | Go 1.24 | binary เล็ก, ใช้ memory น้อย, concurrency ดี | Node.js: ใช้ memory มากกว่าในงานประมวลผลภาพ |
| Web framework | Gin | สร้างบน `net/http` มาตรฐาน เลยใช้ middleware ของ community (Prometheus, OpenTelemetry) และ `context.Context` เรียก AWS SDK v2 ตรงๆ ได้ | Fiber: เร็วกว่าตรงที่สร้างบน `fasthttp` แต่ไม่ compatible กับ `net/http`, middleware ที่ใช้ได้น้อยกว่า, และ `fasthttp.RequestCtx` ถูก pool ใช้ซ้ำ ต้อง copy เองก่อนส่งเข้า goroutine ไม่งั้นข้อมูลเพี้ยน — คอขวดจริงของแอปนี้คือ libvips ไม่ใช่ HTTP layer จึงไม่ได้ประโยชน์จากความเร็วส่วนนี้ |
| Image processing | bimg (libvips) | เร็วและใช้ memory น้อยกว่า ImageMagick มาก | ImageMagick: หนักกว่า |
| PDF | pdfcpu | pure Go, merge และ optimize ได้ | Ghostscript: ต้องเรียก binary ภายนอก |
| Frontend | Next.js 15 + Tailwind | crop และ preview ฝั่ง client, standalone build เล็ก | – |
| Cluster | K3s | Kubernetes จริงแต่เบา, มี Traefik และ metrics-server ในตัว | EKS: control plane ~$73/เดือน เกินงบโปรเจกต์ฝึกหัด |
| CI | Jenkins (Helm) + Kubernetes plugin | ใช้แพร่หลายในองค์กรไทย, agent เป็น pod ชั่วคราว | GitHub Actions: ง่ายกว่า แต่ตั้งใจฝึก Jenkins |
| CD | Argo CD | GitOps, pull-based, Jenkins ไม่ต้องมีสิทธิ์ใน cluster | `kubectl` จาก Jenkins: ต้องให้สิทธิ์ cluster-admin กับ CI |
| Image build | BuildKit rootless | ไม่ต้องใช้ Docker socket หรือ privileged | Kaniko: repo ต้นฉบับถูก archive แล้ว / DinD: ต้อง privileged |
| Branch flow | `feature → dev → staging → main` build once | image ตัวเดียวผ่าน staging มาก่อนขึ้น prod, ด่านอนุมัติอยู่ที่ PR เข้า `main` | trunk-based + deploy ตรงจาก `main`: ไม่มีที่ลองของก่อนขึ้นจริง · build ซ้ำทุก branch: ของที่ขึ้น prod ไม่ใช่ตัวที่ทดสอบ |
| Image push | crane | push ไฟล์ `.tar` ตัวเดียวกับที่สแกน (build once) | build ซ้ำตอน push: เสี่ยงได้ image คนละตัวกับที่สแกน |
| Security scan | Trivy | สแกน vuln, secret และ misconfig ได้ในตัวเดียว | – |
| Manifest | Kustomize | base/overlay, แก้ tag ด้วย `sed` บรรทัดเดียว | Helm chart ของแอปเอง: เกินความจำเป็นสำหรับ 2 service |
| Storage | S3 + lifecycle | เก็บชั่วคราว ลบเองอัตโนมัติ | EBS/PVC: ไม่หมดอายุเอง |
| Registry | ECR | อยู่ region เดียวกัน, IAM auth, scan on push | Docker Hub: rate limit, private repo จำกัด |
| IaC | Terraform | declarative, สร้างและลบได้ทั้งระบบ | ClickOps: ทำซ้ำไม่ได้ |
| TLS | cert-manager + Let's Encrypt | ออก cert และต่ออายุอัตโนมัติ | – |
| เข้าเครื่อง | SSM Session Manager | ไม่เปิด port 22, มี audit log | SSH: ต้องจัดการ key และเปิด port |
