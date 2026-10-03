# ข้อจำกัดและ Roadmap

## ข้อจำกัดที่รู้อยู่แล้ว

| ข้อจำกัด | ผลกระทบ | ถ้าจะแก้ |
| --- | --- | --- |
| Single node | เครื่องล่ม = เว็บล่ม, HPA ขยายได้แค่ในเครื่องเดียว | EKS หรือ K3s หลาย node + ALB |
| ทุก pod ใช้ IAM role ของเครื่องร่วมกัน | pod ใดก็ได้เรียกสิทธิ์ push ECR ได้ | EKS Pod Identity / IRSA แยก role ต่อ ServiceAccount |
| Jenkins อยู่เครื่องเดียวกับแอป | build หนักแย่ง CPU/RAM จากเว็บ | แยก node สำหรับ CI |
| Build ช้า ~9–10 นาที/รอบ | เครื่อง 2 vCPU และ agent สร้างใหม่ทุก build (apt, `go mod`, `npm ci`, image layer ทำใหม่หมด) | agent image สำเร็จรูป, BuildKit cache, เครื่อง build แยก |
| รัน build ได้ทีละตัว (`containerCap: 1`) | build ซ้อนกันเคยทำดิสก์เต็ม (`DiskPressure`) จน pod ถูก evict | เครื่อง build แยก |
| merge เอกสารเข้า `dev` ยังรัน build เต็ม | ข้ามได้เฉพาะระดับ PR | เทียบกับ revision ที่ build ล่าสุดแทน `HEAD^1` |
| ECR เก็บ 10 image ล่าสุด | prod ที่ตามหลัง staging เกิน 10 build หรือ revert ไปไกล อาจดึง image ไม่ได้ | เพิ่ม lifecycle count หรือกัน tag ที่ใช้อยู่ |
| Rollback ต้องมีคน `git revert` | ไม่อัตโนมัติ | Argo Rollouts + analysis |
| Jenkins เปิด `/github-webhook/` สู่ internet | เป็นเป้าโจมตี | จำกัดเฉพาะ IP ของ GitHub หรือใช้ relay |
| ไม่มี code review (โปรเจกต์คนเดียว) | ด่านอนุมัติ prod คือคนกด merge PR `staging → main` | เพิ่ม required review / CODEOWNERS เมื่อมีทีม |

## Roadmap

- [x] Terraform + remote state (S3 + lockfile)
- [x] Jenkins PR/main pipeline + Trivy gate
- [x] Argo CD GitOps + smoke test + Discord
- [x] Namespace `staging` + promote ไป `production` ด้วย PR (build once, ด่านต้นทาง branch)
- [x] Drift detect ทุกคืน + AWS Budgets
- [ ] Observability: Prometheus + Grafana (แบบเบา) และ dashboard latency/error rate
- [ ] Argo Rollouts (canary) + rollback อัตโนมัติ
- [ ] Sign image ด้วย cosign + verify ตอน deploy
- [ ] Policy as code (Kyverno): บังคับ non-root, ห้ามใช้ tag `latest`
- [ ] Load test ด้วย k6 แล้วดู HPA ขยาย pod จริง
- [ ] ปรับความเร็ว CI: agent image สำเร็จรูป + cache
