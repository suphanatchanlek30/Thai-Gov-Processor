# ข้อจำกัดและ Roadmap

## ข้อจำกัดที่รู้อยู่แล้ว

| ข้อจำกัด | ผลกระทบ | ถ้าจะแก้ |
| --- | --- | --- |
| Single node | เครื่องล่ม = เว็บล่ม, HPA ขยายได้แค่ในเครื่องเดียว | EKS หรือ K3s หลาย node + ALB |
| ทุก pod ใช้ IAM role ของเครื่องร่วมกัน | pod ใดก็ได้เรียกสิทธิ์ push ECR ได้ | EKS Pod Identity / IRSA แยก role ต่อ ServiceAccount |
| Jenkins อยู่เครื่องเดียวกับแอป | build หนักๆ อาจแย่ง CPU/RAM จากเว็บ | resource limit (ทำแล้ว) หรือแยก node สำหรับ CI |
| Jenkins เปิดสู่ internet เพื่อรับ webhook | เป็นเป้าโจมตี | จำกัด IP ให้เฉพาะช่วง IP webhook ของ GitHub หรือใช้ smee/relay |
| Rollback ต้องทำเอง | ต้องมีคน `git revert` | Argo Rollouts + analysis อัตโนมัติ |

## Roadmap

- [ ] Terraform + remote state
- [ ] Jenkins PR/main pipeline + Trivy gate
- [ ] Argo CD GitOps + smoke test
- [ ] Namespace `staging` + promote ไป `production` ด้วย PR
- [ ] Observability: Prometheus + Grafana (แบบเบา) และ dashboard latency/error rate
- [ ] Argo Rollouts (canary) + rollback อัตโนมัติ
- [ ] Sign image ด้วย cosign + verify ตอน deploy
- [ ] Policy as code (Kyverno): บังคับ non-root, ห้ามใช้ tag `latest`
- [ ] Load test ด้วย k6 แล้วดู HPA ขยาย pod จริง
