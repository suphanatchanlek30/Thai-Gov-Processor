# หลักฐานการทำงานและบทเรียน

แทนที่ภาพในส่วนนี้ด้วย screenshot จริงจากระบบของคุณ

| หลักฐาน | ภาพ |
| --- | --- |
| Jenkins stage view (PR ผ่าน) | `docs/images/jenkins-pr-pass.png` |
| PR ที่ Trivy บล็อก | `docs/images/trivy-blocked.png` |
| Argo CD แสดง Synced / Healthy | `docs/images/argocd-healthy.png` |
| Rollback ด้วย `git revert` | `docs/images/rollback.gif` |
| แจ้งเตือนใน Discord | `docs/images/discord.png` |
| `terraform apply` / `destroy` | `docs/images/terraform.png` |

## ปัญหาที่เจอและวิธีแก้ (บันทึกระหว่างทำ)

| ปัญหา | สาเหตุ | วิธีแก้ |
| --- | --- | --- |
| _ตัวอย่าง:_ pod ขึ้น `ImagePullBackOff` หลังผ่านไป 12 ชม. | token ของ ECR หมดอายุ | ติดตั้ง ecr-credential-provider |
| _ตัวอย่าง:_ pipeline วนรันไม่จบ | Jenkins commit tag แล้ว trigger ตัวเองซ้ำ | ใส่ `[skip ci]` + SCM Skip plugin |
| _เพิ่มของคุณเอง_ | | |
