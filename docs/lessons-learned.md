# หลักฐานการทำงานและบทเรียน

## หลักฐานจากระบบจริง

| หลักฐาน | ภาพ |
| --- | --- |
| แอป production และ staging | [01-app-prod](screenshots/01-app-prod.png) · [02-app-staging](screenshots/02-app-staging.png) |
| Jenkins: ทุก branch และ PR | [12-jenkins-branches](screenshots/12-jenkins-branches.png) |
| Jenkins: stage ของแต่ละ run (build เต็มเทียบกับ skip) | [10-jenkins-stages-dev](screenshots/10-jenkins-stages-dev.png) · [11-jenkins-stage-view-dev](screenshots/11-jenkins-stage-view-dev.png) |
| Git flow `dev → staging → main` | [20-github-network-graph](screenshots/20-github-network-graph.png) |
| PR ที่ promote (merge commit + tag ที่ build ครั้งเดียว) | [21-pr-dev-to-staging](screenshots/21-pr-dev-to-staging.png) · [22-pr-staging-to-production](screenshots/22-pr-staging-to-production.png) |
| ประวัติ `main`: merge ของคนสลับกับ `deploy … [skip ci]` ของบอท | [23-main-commit-history](screenshots/23-main-commit-history.png) |
| Argo CD: สอง Application, resource tree, history | [30](screenshots/30-argocd-applications.png) · [31](screenshots/31-argocd-tree-prod.png) · [32](screenshots/32-argocd-history.png) |
| Discord: deployed / pushed / promoted / drift | [40-discord-notifications](screenshots/40-discord-notifications.png) |

ผลทดสอบเป็นข้อๆ พร้อมคำสั่งที่ใช้ อยู่ที่ [build-checklist.md](build-checklist.md) ภาพที่ยังไม่มี (ถ้าจะเก็บเพิ่ม): AWS Console (EC2, ECR, S3, Budgets), `terraform plan` ผล *No changes*, PR ที่ Trivy บล็อก

## ปัญหาที่เจอและวิธีแก้

| # | อาการ | สาเหตุ | แก้อย่างไร | ตรวจพบได้อย่างไร |
| --- | --- | --- | --- | --- |
| 1 | pod `ImagePullBackOff` หลังผ่านไป 12 ชม. | token ของ ECR หมดอายุ | `ecr-credential-provider` ใช้ IAM role ของเครื่อง | ทดสอบ pod ใหม่หลัง node อายุเกิน 12 ชม. |
| 2 | Jenkins `golangci-lint` ล้มใน container ทั้งที่ผ่านบนเครื่อง | git "dubious ownership" ใน workspace | `GOFLAGS=-buildvcs=false` | build แรกของ pipeline |
| 3 | Trivy เจอ CRITICAL จริงในโค้ดเดิม | Next.js, Go stdlib, npm/tar, libgnutls ใน base image | อัปเดตเวอร์ชัน แล้ว gate ผ่าน | PR ทดสอบที่ใส่ `minimist 1.2.5` ถูกบล็อกจริง |
| 4 | commit ที่ Jenkins สร้างทำให้ build วน | commit tag trigger webhook ซ้ำ | `[skip ci]` + ข้าม build ที่ commit บนสุดเป็นของบอท | log: build ที่เกิดจาก `deploy …` ถูกข้าม |
| 5 | HPA กับ self-heal แย่งกันตั้ง `replicas` | Deployment ระบุ `replicas: 2` คู่กับ HPA | เอา `replicas` ออกให้ HPA คุม | วิเคราะห์ตอนเขียน Application (ถ้าไม่แก้ scale-out จะถูก reset) |
| 6 | `merge-pdf` ทำ backend ตาย (502/503) ทั้งที่ `/healthz` เขียว | pdfcpu เขียน config ลง `$HOME` แล้ว `os.Exit(1)` เมื่อ filesystem เป็น read-only | `api.DisableConfigDir()` + smoke test อัปโหลดไฟล์เข้า `merge-pdf` | ผู้ใช้เจอ 502 · `kubectl logs --previous` เห็น `pdfcpu: config problem: … read-only file system` |
| 7 | pod ของ build ถูก evict (`DiskPressure`) | build สองตัวพร้อมกันกินดิสก์ชั่วคราวเกือบ 20 GiB บนเครื่อง 33 GB | `agent.containerCap: 1` | `kubectl describe node` เห็น taint `disk-pressure` |
| 8 | prod ได้ URL ของ staging ติดมากับ image | `NEXT_PUBLIC_API_BASE_URL` ฝังตอน build (build once ใช้ไม่ได้) | frontend เรียก `/api` แบบ same-origin | อ่านโค้ดตอนออกแบบ flow สามชั้น · bundle ไม่มี `localhost:8080` |
| 9 | `git revert` แล้วเว็บไม่ย้อน | revert กลับเร็วกว่ารอบ poll ของ Argo (3 นาที) Argo ไม่ทันเห็น | รอให้ pod เปลี่ยนก่อน หรือ `refresh=hard` | history ของ Argo ไม่มี sync ของ commit revert |
| 10 | build รอบเกินหลัง merge หลาย PR ติดกัน | `scmSkip` อ่านข้อความจาก changelog (เห็น commit ของฟีเจอร์) ไม่ใช่ commit บนสุด | `Classify` อ่าน `git log -1` เอง | log: `NOT matched on message: <commit ของฟีเจอร์>` ทั้งที่ commit บนสุดคือ `deploy … [skip ci]` |
| 11 | PR ที่หัวคือ commit บอท `[skip ci]` ค้าง BLOCKED | ถ้า skip ที่ระดับ PR เช็คที่บังคับจะไม่ถูกรายงาน | ไม่ข้ามด้วย `[skip ci]` สำหรับ PR | คาดไว้ตอนออกแบบ แล้วพิสูจน์ด้วย PR dev→staging จริง |
| 12 | เช็ค "แก้แต่เอกสาร" ให้ผลผิด (โค้ดปนเอกสารผ่านเป็น docs-only) | `grep -q -v` ให้ผลต่างกันระหว่าง BSD (Mac) กับ GNU/BusyBox | นับบรรทัดไม่ตรงด้วย `grep -vc` แล้วเทสต์ 6 กรณีในโปรเจกต์ทดลอง | เทสต์ shell logic แยกก่อนใส่ Jenkinsfile |
| 13 | เช็ค docs-only บน branch ซ่อนโค้ดได้ | build ที่รอคิวเช็คเอาหัว branch ตอนเริ่ม แล้ว `HEAD^1 HEAD` เห็นแค่ merge สุดท้าย | ตัดสิน docs-only เฉพาะ PR (diff ทั้ง PR) | ต่อยอดจากเคส scmSkip ข้อ 10 (หลักการเดียวกัน: build ที่รอคิวเห็นหัว branch ตอนเริ่ม) |
| 14 | job `terraform-drift` ไม่โผล่ใน Jenkins หลัง `helm upgrade` | Job DSL เขียน `config.xml` ลงดิสก์แต่ Jenkins ยังไม่โหลดในบูตแรก | restart controller ครั้งที่สอง | `/job/terraform-drift/api/json` ตอบ 404 ทั้งที่ไฟล์อยู่บนดิสก์ |
| 15 | สั่ง build ผ่าน API ได้ 403 | crumb ผูกกับ session | ใช้ cookie jar เดียวกันตอนขอ crumb และ POST | curl ตอบ 403 · ลองใหม่ด้วย `-c/-b` ได้ 201 |
| 16 | `curl` ถาม Jenkins API แล้วได้คำตอบว่าง | `[ ]` ใน URL ถูกตีความเป็น pattern | `curl -g` | JSON parse error |
| 17 | terminal ใหม่ `helm`/`kubectl` ขึ้น `x509: certificate signed by unknown authority` | ไม่ได้ตั้ง `KUBECONFIG` เลยไปคุยกับ `localhost:8080` ซึ่งคือ port-forward ของ Argo CD | `export KUBECONFIG=~/.kube/thai-gov-k3s.yaml` ทุก terminal | error อ้าง `https://localhost:8080` |
| 18 | คำสั่งทดสอบ notification พิมพ์ URL ของ Discord webhook ลง log | `argocd admin notifications template notify` log request เต็มที่ระดับ debug | กรอง output ให้เหลือแค่รหัสสถานะ | เห็นใน output ของคำสั่ง |
| 19 | `terraform destroy` ล้มตอนปิดโปรเจกต์ | ลบ ECR repo ไม่ได้ถ้ามี image (ไม่ได้ตั้ง `force_delete`), S3 bucket ไม่ว่าง, IAM user มี access key ที่สร้างด้วยมือ และ bucket ของ state มี versioning | `scripts/teardown.sh` เคลียร์สามอย่างก่อน แล้วลบ bucket state ทุก version ตอนท้าย | `terraform destroy` ขึ้น `RepositoryNotEmptyException` |
| 20 | ลบ image ใน ECR ทีละ tag แล้วยัง `RepositoryNotEmpty` | ยังเหลือ image แบบไม่มี tag (untagged manifest) 2 ตัวต่อ repo ที่ลบตาม tag ไม่ถึง | `aws ecr delete-repository --force` ใน script | `aws ecr list-images` หลังลบ ยังเห็น `imageDigest` ที่ไม่มี `imageTag` |

## สิ่งที่จะทำต่างออกไป

- เพิ่ม `merge-pdf` เข้า smoke test **ตั้งแต่ Phase 3** เช็คของ Phase 3 ยิงแค่ `/presets` จึงไม่เคยเจอบั๊กข้อ 6
- ทดสอบทุกอย่างในสภาพ container จริง (`readOnlyRootFilesystem`, ไม่มี `$HOME` ที่เขียนได้) ไม่ใช่แค่ Docker Compose
- ตัดสินใจเรื่อง flow ของ branch (build once, tag จาก staging overlay, ด่านต้นทาง) ก่อนเขียน Jenkinsfile ไม่ใช่ปรับทีหลัง
- เก็บ screenshot ไปพร้อมกับแต่ละ phase ไม่ใช่ท้ายโปรเจกต์ (ภาพฝั่ง AWS Console เก็บไม่ทันก่อนลบ)
- ออกแบบ teardown ตั้งแต่แรก: `force_delete` ที่ ECR และ `force_destroy` ที่ S3 ของไฟล์ผลลัพธ์ จะทำให้ `terraform destroy` ผ่านในคำสั่งเดียว (ไม่ได้ใส่ไว้เพราะกลัวลบข้อมูลโดยไม่ตั้งใจ แต่ทำให้ปิดโปรเจกต์ยุ่งยาก)
