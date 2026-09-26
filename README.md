# 🇹🇭 Thai Gov Photo & Doc Processor

> เว็บแปลงรูปถ่ายและเอกสารให้ตรงสเปกระบบรับสมัครของหน่วยงานราชการไทยโดยอัตโนมัติ
> พร้อม **CI/CD pipeline แบบ GitOps** บน **K3s + AWS** ที่สร้างทั้งหมดด้วย **Terraform**

![Go](https://img.shields.io/badge/Go-1.22-00ADD8?logo=go&logoColor=white)
![Next.js](https://img.shields.io/badge/Next.js-14-000000?logo=nextdotjs)
![K3s](https://img.shields.io/badge/K3s-Kubernetes-FFC61C?logo=k3s&logoColor=black)
![Jenkins](https://img.shields.io/badge/CI-Jenkins-D24939?logo=jenkins&logoColor=white)
![Argo CD](https://img.shields.io/badge/CD-Argo%20CD-EF7B4D?logo=argo&logoColor=white)
![Terraform](https://img.shields.io/badge/IaC-Terraform-7B42BC?logo=terraform&logoColor=white)
![AWS](https://img.shields.io/badge/Cloud-AWS-FF9900?logo=amazonaws&logoColor=white)
![Trivy](https://img.shields.io/badge/Security-Trivy-1904DA?logo=aqua&logoColor=white)

โปรเจกต์นี้ทำขึ้นเพื่อ**เรียนรู้และสาธิตงาน DevOps แบบครบวงจร** ได้แก่ Infrastructure as Code, container, Kubernetes, CI/CD, security scanning, GitOps และการคุมค่าใช้จ่ายบน cloud โดยใช้โจทย์จริงที่คนไทยเจอเป็นตัวแอป

> **ขอบเขตที่ตั้งใจไว้:** ระบบนี้เป็น *production-like* บน EC2 เครื่องเดียว ไม่ได้ออกแบบมาเพื่อ high availability ข้อจำกัดและแนวทางต่อยอดอยู่ใน [หัวข้อ 15](#-15-ข้อจำกัดและ-roadmap)

---

## 📑 สารบัญ

1. [ปัญหาและที่มา](#-1-ปัญหาและที่มา)
2. [ฟีเจอร์](#-2-ฟีเจอร์)
3. [ภาพรวมสถาปัตยกรรม](#-3-ภาพรวมสถาปัตยกรรม)
4. [CI/CD pipeline แบบละเอียด](#-4-cicd-pipeline-แบบละเอียด)
5. [Runtime architecture บน AWS](#-5-runtime-architecture-บน-aws)
6. [Infrastructure as Code (Terraform)](#-6-infrastructure-as-code-terraform)
7. [Tech stack และเหตุผลที่เลือก](#-7-tech-stack-และเหตุผลที่เลือก)
8. [โครงสร้าง repository](#-8-โครงสร้าง-repository)
9. [เริ่มต้นใช้งาน (ทีละขั้น)](#-9-เริ่มต้นใช้งาน-ทีละขั้น)
10. [ไฟล์ config สำคัญ](#-10-ไฟล์-config-สำคัญ)
11. [REST API](#-11-rest-api)
12. [Backend: อัลกอริทึมบีบไฟล์](#-12-backend-อัลกอริทึมบีบไฟล์)
13. [Security และ PDPA](#-13-security-และ-pdpa)
14. [ค่าใช้จ่าย](#-14-ค่าใช้จ่าย)
15. [ข้อจำกัดและ roadmap](#-15-ข้อจำกัดและ-roadmap)
16. [หลักฐานการทำงานและบทเรียน](#-16-หลักฐานการทำงานและบทเรียน)

---

## 📌 1. ปัญหาและที่มา

ระบบรับสมัครงานราชการ รัฐวิสาหกิจ และงานทะเบียน มีข้อกำหนดเรื่องไฟล์ที่เข้มงวด:

| หน่วยงาน / วัตถุประสงค์ | ขนาดภาพ | ประเภท | ขนาดไฟล์สูงสุด | เงื่อนไขพิเศษ |
| :--- | :--- | :--- | :--- | :--- |
| **สำนักงาน ก.พ. (OCSC)** | 200 × 230 px (1 × 1.5 นิ้ว) | `.jpg` | **100 KB** (บางรอบ 50–100 KB) | หน้าตรง สัดส่วนต้องไม่บิดเบี้ยว |
| **หนังสือเดินทาง** | 500 × 500 px (2 × 2 นิ้ว) | `.jpg` | **200 KB** | ฉากหลังขาวล้วน |
| **ครูผู้ช่วย / ข้าราชการท้องถิ่น** | 150 × 200 ถึง 300 × 400 px | `.jpg` | **200 KB** | หน้าตรง เครื่องแต่งกายสุภาพ |
| **สำเนาเอกสาร (บัตร ปชช. / ทะเบียนบ้าน)** | A4 หลายหน้า | `.pdf` | รวม **500 KB** | อ่านเลขบัตรได้ชัด |

> ข้อกำหนดจริงเปลี่ยนตามประกาศแต่ละรอบ ค่าในตารางใช้เป็น preset เริ่มต้น และผู้ใช้เลือกแบบ `custom` ได้

**Pain points**

- ผู้ใช้ทั่วไปไม่มีโปรแกรมที่กำหนดขนาดไฟล์เป็น KB ได้แม่นยำ
- ย่อภาพผิดวิธีจนหน้าบิดเบี้ยว แล้วระบบของหน่วยงานปฏิเสธไฟล์
- เว็บแปลงไฟล์ฟรีหลายแห่งไม่มีนโยบายความเป็นส่วนตัวชัดเจน ซึ่งเสี่ยงมากกับสำเนาบัตรประชาชน (PDPA)

---

## ✨ 2. ฟีเจอร์

| ฝั่งผู้ใช้ | ฝั่ง DevOps |
| --- | --- |
| เลือก preset ตามหน่วยงาน ระบบตั้งขนาด น้ำหนักไฟล์ และสีฉากหลังให้ | Infrastructure ทั้งหมดสร้างด้วย Terraform (`apply` / `destroy` ได้ในคำสั่งเดียว) |
| Crop และหมุนภาพ โดยล็อกสัดส่วนตาม preset | Jenkins agent เป็น pod ชั่วคราว สร้างตอน build แล้วลบทิ้ง |
| บีบไฟล์ให้ต่ำกว่าเกณฑ์ KB โดยรักษาความคมชัดสูงสุด | Security gate ด้วย Trivy: เจอช่องโหว่ CRITICAL จะหยุด pipeline จริง |
| รวมหลายภาพเป็น PDF เดียวที่ไม่เกิน 500 KB | GitOps ด้วย Argo CD: Git คือแหล่งความจริงเดียว และ rollback ได้ด้วย `git revert` |
| เทียบขนาดก่อนและหลัง เช่น 3.4 MB → 89 KB | Zero-downtime rolling update พร้อม smoke test หลัง deploy |
| ไฟล์หมดอายุอัตโนมัติ ไม่เก็บถาวร | ไม่มี access key ใน Git ใช้ IAM role ทั้งหมด และเข้าเครื่องผ่าน SSM แทน SSH |

---

## 🗺️ 3. ภาพรวมสถาปัตยกรรม

ภาพนี้แสดงระบบทั้งหมดในภาพเดียว ส่วนรายละเอียดของแต่ละส่วนอยู่ในหัวข้อถัดไป

```mermaid
flowchart LR
    dev(["👩‍💻 Developer"])
    user(["👤 ผู้ใช้"])
    gh["🐙 GitHub<br/>monorepo"]
    jenkins["⚙️ Jenkins<br/>CI"]
    argo["🔄 Argo CD<br/>CD / GitOps"]
    ecr[("📦 Amazon ECR")]
    s3[("🪣 Amazon S3<br/>ไฟล์ชั่วคราว")]
    k3s["☸️ K3s บน EC2<br/>frontend + backend"]
    tf["🏗️ Terraform"]
    discord["💬 Discord"]

    dev -->|"1 · push / PR"| gh
    gh -->|"2 · webhook"| jenkins
    jenkins -->|"3 · push image"| ecr
    jenkins -->|"4 · commit tag ใหม่"| gh
    argo -->|"5 · เฝ้าดู repo"| gh
    argo -->|"6 · sync"| k3s
    ecr -->|"pull image"| k3s
    jenkins & argo -->|แจ้งเตือน| discord

    user -->|HTTPS| k3s
    k3s -->|เก็บ/อ่านไฟล์| s3

    tf -.->|สร้าง| k3s
    tf -.->|สร้าง| ecr
    tf -.->|สร้าง| s3

    classDef ci fill:#e8ebff,stroke:#3f4fc2,color:#1b2270
    classDef cd fill:#e3f4e8,stroke:#2b7f43,color:#14401f
    classDef aws fill:#fff1e0,stroke:#b8600a,color:#5a2f04
    classDef iac fill:#efe6fb,stroke:#6d3fba,color:#34185e
    class jenkins ci
    class argo cd
    class ecr,s3,k3s aws
    class tf iac
```

**สรุปใน 1 ประโยค:** Jenkins ทำหน้าที่ *ตรวจและแพ็ก* โค้ด (CI) แล้วเขียนเวอร์ชันใหม่ลง Git จากนั้น Argo CD *อ่าน Git แล้วเอาขึ้นเว็บจริง* (CD) ส่วน Terraform เป็นคน *สร้างเครื่องและบริการ* บน AWS ทั้งหมด

---

## 🔁 4. CI/CD pipeline แบบละเอียด

### 4.1 ภาพรวม 4 เลน

```mermaid
flowchart TB
    subgraph L1["① Developer & GitHub"]
        direction LR
        a1["สร้าง branch<br/><code>feat/*</code>"] --> a2["commit & push"] --> a3["เปิด Pull Request<br/>(branch protection)"] --> a4["GitHub webhook<br/>ลงนามด้วย HMAC secret"]
    end

    subgraph L2["② CI · PR pipeline (ตรวจอย่างเดียว ไม่ deploy)"]
        direction LR
        b1["Checkout"] --> b2["Lint<br/>go vet · golangci-lint<br/>eslint · tsc"] --> b3["Unit test<br/>go test -race -cover<br/>npm test"] --> b4["Build image<br/>BuildKit → .tar<br/>(ยังไม่ push)"] --> b5{{"🛡️ Trivy gate<br/>CRITICAL = fail"}} --> b6["Terraform<br/>fmt · validate · plan<br/>(ถ้าแก้ iac/)"] --> b7["ส่งสถานะ ✔/✘<br/>กลับไปที่ PR"]
    end

    subgraph L3["③ CI · main pipeline (build ของจริง)"]
        direction LR
        c1["Build image<br/>tag = git SHA"] --> c2{{"🛡️ Trivy gate<br/>สแกนซ้ำ"}} --> c3["Push ECR<br/>(IMMUTABLE)"] --> c4["แก้ tag ใน<br/>kustomization.yaml<br/>commit [skip ci]"] --> c5["แจ้ง Discord"]
    end

    subgraph L4["④ CD · Argo CD (GitOps)"]
        direction LR
        d1["ตรวจเจอ commit ใหม่<br/>(poll 3 นาที)"] --> d2["Sync เข้า<br/>ns: production"] --> d3["Rolling update<br/>maxUnavailable 0<br/>readinessProbe"] --> d4{{"🧪 Smoke test<br/>PostSync Job"}} --> d5["✅ แจ้งผล / ❌ git revert"]
    end

    a4 --> b1
    b7 --> rv["👀 Code review + Merge"]
    rv --> c1
    c5 --> d1

    classDef gate fill:#fde4e4,stroke:#b83030,color:#5c1414
    classDef pr fill:#e8ebff,stroke:#3f4fc2,color:#1b2270
    classDef main fill:#dff3f4,stroke:#0a7580,color:#053b40
    classDef cd fill:#e3f4e8,stroke:#2b7f43,color:#14401f
    class b1,b2,b3,b4,b6,b7 pr
    class c1,c3,c4,c5 main
    class d1,d2,d3,d5 cd
    class b5,c2,d4 gate
```

> กล่องหกเหลี่ยมสีแดงคือ **ด่านตรวจ** ถ้าไม่ผ่าน pipeline จะหยุดตรงนั้น

### 4.2 แต่ละขั้นทำอะไร

| เลน | ขั้น | ทำอะไร | Tool | ถ้าไม่ผ่าน |
| --- | --- | --- | --- | --- |
| ② PR | Checkout | ดึงโค้ดของ PR เข้า agent pod | Jenkins, git | – |
| ② PR | Lint | ตรวจรูปแบบโค้ดและ type error | `go vet`, golangci-lint, eslint, `tsc --noEmit` | หยุด ✘ |
| ② PR | Unit test | ทดสอบ logic เช่น ไฟล์ออกมาต้อง ≤ 100 KB และขนาด 200×230 จริง | `go test -race -cover`, `npm test` | หยุด ✘ |
| ② PR | Build image | ลอง build ให้แน่ใจว่า Dockerfile ใช้ได้ ผลลัพธ์เก็บเป็น `.tar` | BuildKit (rootless) | หยุด ✘ |
| ② PR | **Security gate** | สแกนช่องโหว่ของ dependency, OS package และ secret ที่หลุดเข้าโค้ด | `trivy fs`, `trivy image --input` | **บล็อก PR** |
| ② PR | Terraform | `fmt -check`, `validate`, `plan` ด้วย role แบบ read-only | Terraform | หยุด ✘ |
| ② PR | Status | ส่ง ✔/✘ เป็น required status check ของ GitHub | GitHub Branch Source | ปุ่ม Merge ถูกล็อก |
| ③ main | Build | build backend และ frontend ติด tag `a1b2c3d` (git SHA) | BuildKit | หยุด |
| ③ main | **Security gate** | สแกน image ตัวสุดท้ายซ้ำ เผื่อมี CVE ใหม่ประกาศหลัง PR ผ่าน | Trivy | **ไม่ push** |
| ③ main | Push | push **ไฟล์ .tar ตัวเดียวกับที่สแกนแล้ว** ขึ้น ECR | crane, IAM role | หยุด |
| ③ main | Update manifest | แก้ `newTag` ใน `k8s/overlays/prod/kustomization.yaml` แล้ว commit กลับ | yq, git | หยุด |
| ④ CD | Detect | Argo CD เห็นว่า Git ไม่ตรงกับ cluster | Argo CD | – |
| ④ CD | Sync | apply manifest, self-heal ถ้ามีคนแก้ใน cluster ด้วยมือ | Argo CD | สถานะ OutOfSync |
| ④ CD | Rolling update | สร้าง pod ใหม่ให้ ready ก่อน แล้วค่อยลบ pod เก่า | Deployment, readinessProbe | pod เก่ายังรับ traffic ต่อ |
| ④ CD | **Smoke test** | Job เรียก `/healthz` และ `/api/v1/selftest` (แปลงรูปตัวอย่างจริง) | PostSync hook | sync = Failed → แจ้ง ❌ |
| ④ CD | Rollback | `git revert` commit ที่แก้ tag แล้ว Argo sync กลับเวอร์ชันเดิม | git, Argo CD | – |

### 4.3 Deploy 1 ครั้งเกิดอะไรขึ้นบ้าง (sequence)

```mermaid
sequenceDiagram
    autonumber
    actor Dev as Developer
    participant GH as GitHub
    participant J as Jenkins
    participant ECR as Amazon ECR
    participant A as Argo CD
    participant K as K3s (production)
    participant D as Discord

    Dev->>GH: merge PR เข้า main
    GH->>J: webhook (push to main)
    J->>J: build image → Trivy scan
    J->>ECR: push thai-gov-backend:a1b2c3d
    J->>GH: commit "deploy a1b2c3d [skip ci]"
    GH-->>J: webhook อีกรอบ (ถูก scmSkip ข้าม)
    J->>D: ✅ build ผ่าน
    A->>GH: poll (ทุก 3 นาที)
    A->>K: apply manifest ใหม่
    K->>ECR: pull a1b2c3d
    K->>K: rolling update + readinessProbe
    A->>K: รัน smoke-test Job (PostSync)
    alt smoke test ผ่าน
        A->>D: ✅ a1b2c3d ขึ้น production แล้ว
    else smoke test ไม่ผ่าน
        A->>D: ❌ sync failed
        Dev->>GH: git revert (commit ที่แก้ tag)
        A->>K: sync กลับเวอร์ชันก่อนหน้า
    end
```

### 4.4 ประวัติ Git ที่เกิดขึ้นจริง

```mermaid
gitGraph
    commit id: "init"
    branch feat-ocsc-preset
    checkout feat-ocsc-preset
    commit id: "add OCSC preset"
    commit id: "fix unit test"
    checkout main
    merge feat-ocsc-preset id: "merge PR #12"
    commit id: "deploy a1b2c3d [skip ci]"
    commit id: "revert: rollback a1b2c3d" type: REVERSE
```

### 4.5 เหตุการณ์ไหนทำให้อะไรรัน

| เหตุการณ์ | สิ่งที่รัน | ขึ้นเว็บจริงไหม |
| --- | --- | --- |
| push เข้า feature branch ที่ยังไม่เปิด PR | ไม่รันอะไร | ไม่ |
| เปิด PR หรือ push เพิ่มเข้า PR | เลน ② (+ terraform plan ถ้าแก้ `iac/`) | ไม่ |
| merge เข้า `main` | เลน ③ ต่อด้วยเลน ④ | ใช่ ภายในไม่กี่นาที |
| commit ที่มี `[skip ci]` (Jenkins อัปเดต tag) | Jenkins ข้าม, Argo CD sync | ใช่ |
| `git revert` บน `main` | Argo CD sync กลับเวอร์ชันเดิม | ใช่ = rollback |
| ทุกคืน 02:00 | Terraform drift detect | ไม่ (แจ้งเตือนอย่างเดียว) |
| สั่ง `terraform apply` จากเครื่อง | เปลี่ยน infrastructure | เปลี่ยน infra |

---

## ☁️ 5. Runtime architecture บน AWS

```mermaid
flowchart TB
    user(["👤 ผู้ใช้"])
    gh["🐙 GitHub"]
    le["🔐 Let's Encrypt"]

    subgraph AWS["☁️ AWS · ap-southeast-1 (Singapore)"]
        subgraph VPC["VPC 10.0.0.0/16 → public subnet 10.0.1.0/24"]
            subgraph EC2["🖥️ EC2 t3a.large · 2 vCPU / 8 GB · Ubuntu 24.04 · IMDSv2 · Elastic IP<br/>SG: 80, 443 ทุกที่ · 6443 เฉพาะ IP ตัวเอง · ไม่เปิด 22 (ใช้ SSM)"]
                subgraph K3S["☸️ K3s single-node"]
                    subgraph KS["ns: kube-system"]
                        traefik["Traefik ingress<br/>:80 → :443"]
                        cm["cert-manager"]
                        ms["metrics-server"]
                    end
                    subgraph PROD["ns: production"]
                        fe["Frontend<br/>Next.js × 2"]
                        be["Backend<br/>Go + libvips<br/>HPA 2–4"]
                        smoke["smoke-test Job"]
                    end
                    subgraph JNS["ns: jenkins"]
                        jc["Jenkins controller<br/>PVC 10 GB"]
                        ja["Agent pods<br/>(ชั่วคราว)"]
                    end
                    subgraph ANS["ns: argocd"]
                        argo["Argo CD"]
                    end
                end
            end
        end
        ecr[("📦 ECR<br/>thai-gov-backend<br/>thai-gov-frontend")]
        s3[("🪣 S3 ไฟล์ผู้ใช้<br/>uploads/ · processed/<br/>หมดอายุ 1 วัน")]
        state[("🪣 S3 Terraform state")]
        iam["🔑 IAM role<br/>least privilege"]
    end

    user -->|HTTPS| traefik
    traefik -->|"app.example.com /"| fe
    traefik -->|"app.example.com /api"| be
    traefik -->|"ci.example.com"| jc
    le <-->|HTTP-01| cm
    cm -.->|TLS cert| traefik
    be -->|"PUT / GET / presigned URL"| s3
    ms -.->|CPU metrics| be
    gh -->|webhook| traefik
    jc --> ja
    ja -->|push| ecr
    ja -->|commit tag| gh
    argo -->|อ่าน repo| gh
    argo -->|sync| PROD
    ecr -->|pull| PROD
    iam -.->|instance profile| EC2

    classDef aws fill:#fff1e0,stroke:#b8600a,color:#5a2f04
    classDef app fill:#e3f4e8,stroke:#2b7f43,color:#14401f
    classDef ci fill:#e8ebff,stroke:#3f4fc2,color:#1b2270
    class ecr,s3,state,iam aws
    class fe,be,smoke,argo app
    class jc,ja ci
```

### 5.1 เส้นทางของผู้ใช้ 1 คน

```mermaid
sequenceDiagram
    autonumber
    actor U as ผู้ใช้
    participant T as Traefik
    participant FE as Frontend (Next.js)
    participant BE as Backend (Go)
    participant S3 as Amazon S3

    U->>T: เปิด https://app.example.com
    T->>FE: route "/"
    FE-->>U: หน้าเว็บ + preset
    U->>U: เลือก "ก.พ." แล้ว crop (ล็อก 4:5)
    U->>T: POST /api/v1/photos/preset (multipart)
    T->>BE: route "/api"
    BE->>BE: resize 200×230 → binary search quality (≤ 7 รอบ)
    BE->>S3: PutObject processed/ocsc_xxx.jpg
    BE->>S3: สร้าง presigned URL (1 ชม.)
    BE-->>U: JSON { size_kb: 84.6, download_url }
    U->>S3: ดาวน์โหลดไฟล์ตรงจาก S3
    Note over S3: lifecycle ลบไฟล์หลังหมดอายุ
```

---

## 🏗️ 6. Infrastructure as Code (Terraform)

```mermaid
flowchart LR
    pr["แก้ไฟล์ใน iac/<br/>ผ่าน PR"] --> plan["CI: fmt · validate · plan<br/>(role read-only)"]
    plan --> review["อ่าน plan ใน PR<br/>แล้ว merge"]
    review --> apply["terraform apply<br/>(จากเครื่องผู้ดูแล)"]
    apply --> state[("S3 state<br/>versioning + lockfile")]
    apply --> aws["AWS resources"]
    cron["⏰ ทุกคืน 02:00"] --> drift{"plan<br/>-detailed-exitcode"}
    drift -->|exit 0| ok["✅ ตรงกับโค้ด"]
    drift -->|exit 2| alert["⚠️ มีคนแก้ AWS ด้วยมือ<br/>แจ้ง Discord"]
    demo["จบเดโม"] --> destroy["terraform destroy<br/>ประหยัดค่าใช้จ่าย"]
```

**ทรัพยากรที่ Terraform สร้าง**

| ไฟล์ | สร้างอะไร |
| --- | --- |
| `versions.tf` | provider, remote state บน S3 (`use_lockfile`) |
| `network.tf` | VPC, public subnet, Internet Gateway, route table |
| `security_group.tf` | เปิด 80/443 ทุกที่, 6443 เฉพาะ `var.admin_cidr` |
| `iam.tf` | IAM role และ instance profile แบบ least privilege + SSM |
| `ec2.tf` | EC2 t3a.large, Elastic IP, IMDSv2, EBS เข้ารหัส, user_data ติดตั้ง K3s |
| `ecr.tf` | ECR 2 repo, IMMUTABLE, scan on push, lifecycle เก็บ 10 image ล่าสุด |
| `s3.tf` | bucket ไฟล์ผู้ใช้, block public access, SSE, lifecycle 1 วัน |
| `outputs.tf` | public IP, ชื่อ bucket, ECR URL |

ตัวอย่างโค้ดเต็มอยู่ใน [หัวข้อ 10.6](#106-terraform)

---

## 🛠️ 7. Tech stack และเหตุผลที่เลือก

| ส่วนประกอบ | เลือกใช้ | เหตุผล | ทางเลือกที่ไม่เลือก และเหตุผล |
| --- | --- | --- | --- |
| Backend | **Go 1.22** | binary เล็ก, ใช้ memory น้อย, concurrency ดี | Node.js: ใช้ memory มากกว่าในงานประมวลผลภาพ |
| Image processing | **bimg (libvips)** | เร็วและใช้ memory น้อยกว่า ImageMagick มาก | ImageMagick: หนักกว่า |
| PDF | **pdfcpu** | pure Go, merge และ optimize ได้ | Ghostscript: ต้องเรียก binary ภายนอก |
| Frontend | **Next.js 14 + Tailwind** | crop และ preview ฝั่ง client, standalone build เล็ก | – |
| Cluster | **K3s** | Kubernetes จริงแต่เบา, มี Traefik และ metrics-server ในตัว | EKS: control plane ~$73/เดือน เกินงบโปรเจกต์ฝึกหัด |
| CI | **Jenkins (Helm) + Kubernetes plugin** | ใช้แพร่หลายในองค์กรไทย, agent เป็น pod ชั่วคราว | GitHub Actions: ง่ายกว่า แต่ตั้งใจฝึก Jenkins |
| CD | **Argo CD** | GitOps, pull-based, Jenkins ไม่ต้องมีสิทธิ์ใน cluster | `kubectl` จาก Jenkins: ต้องให้สิทธิ์ cluster-admin กับ CI |
| Image build | **BuildKit rootless** | ไม่ต้องใช้ Docker socket หรือ privileged | Kaniko: repo ต้นฉบับถูก archive แล้ว / DinD: ต้อง privileged |
| Image push | **crane** | push ไฟล์ `.tar` ตัวเดียวกับที่สแกน (build once) | build ซ้ำตอน push: เสี่ยงได้ image คนละตัวกับที่สแกน |
| Security scan | **Trivy** | สแกน vuln, secret และ misconfig ได้ในตัวเดียว | – |
| Manifest | **Kustomize** | base/overlay, แก้ tag ด้วย yq ได้ง่าย | Helm chart ของแอปเอง: เกินความจำเป็นสำหรับ 2 service |
| Storage | **S3 + lifecycle** | เก็บชั่วคราว ลบเองอัตโนมัติ | EBS/PVC: ไม่หมดอายุเอง |
| Registry | **ECR** | อยู่ region เดียวกัน, IAM auth, scan on push | Docker Hub: rate limit, private repo จำกัด |
| IaC | **Terraform** | declarative, สร้างและลบได้ทั้งระบบ | ClickOps: ทำซ้ำไม่ได้ |
| TLS | **cert-manager + Let's Encrypt** | ออก cert และต่ออายุอัตโนมัติ | – |
| เข้าเครื่อง | **SSM Session Manager** | ไม่เปิด port 22, มี audit log | SSH: ต้องจัดการ key และเปิด port |

---

## 📂 8. โครงสร้าง repository

```text
thai-gov-processor/
├── backend/                         # Go API
│   ├── cmd/api/main.go              # entry point, router, graceful shutdown
│   ├── internal/
│   │   ├── handler/                 # HTTP handlers (preset, merge-pdf, health, selftest)
│   │   ├── processor/               # resize + binary search compress, pdfcpu
│   │   │   └── testdata/            # รูปตัวอย่างสำหรับ unit test และ selftest
│   │   ├── storage/                 # S3 client + presigned URL
│   │   └── preset/                  # ค่า preset ก.พ., passport, ครู
│   ├── Dockerfile
│   └── go.mod / go.sum
├── frontend/                        # Next.js 14
│   ├── src/app/                     # layout.tsx, page.tsx
│   ├── src/components/              # DropZone, PresetCard, ImageCropper, PreviewModal
│   ├── Dockerfile
│   └── package.json
├── iac/                             # Terraform
│   ├── versions.tf  variables.tf  network.tf  security_group.tf
│   ├── iam.tf  ec2.tf  ecr.tf  s3.tf  outputs.tf
│   ├── templates/user_data.sh.tftpl
│   └── terraform.tfvars.example
├── k8s/
│   ├── base/                        # deployment, service, hpa, ingress, smoke-test job
│   ├── overlays/prod/kustomization.yaml   # ← Jenkins แก้ image tag ที่ไฟล์นี้
│   ├── argocd/application.yaml
│   └── platform/
│       ├── jenkins-values.yaml
│       ├── cluster-issuer.yaml
│       └── credential-provider.yaml
├── ci/
│   ├── agent-pod.yaml               # spec ของ Jenkins agent pod
│   └── Jenkinsfile.drift            # job drift detect ทุกคืน
├── scripts/bootstrap-cluster.sh     # ติดตั้ง cert-manager, Argo CD, Jenkins
├── docker-compose.yml               # รันบนเครื่องตัวเอง
├── Jenkinsfile                      # pipeline หลัก (PR + main)
└── README.md
```

---

## 🚀 9. เริ่มต้นใช้งาน (ทีละขั้น)

### 9.0 สิ่งที่ต้องมี

- AWS account + AWS CLI v2 ที่ login แล้ว
- Terraform ≥ 1.10 (ใช้ `use_lockfile` ของ S3 backend)
- Session Manager plugin ของ AWS CLI
- Domain ที่ชี้ DNS ได้ (หรือใช้ `sslip.io` ระหว่างทดสอบ)
- Docker สำหรับรันบนเครื่องตัวเอง

### 9.1 รันบนเครื่องตัวเอง

```bash
git clone https://github.com/<OWNER>/thai-gov-processor.git
cd thai-gov-processor
docker compose up --build
# frontend: http://localhost:3000   backend: http://localhost:8080/healthz
```

`docker-compose.yml` ใช้ MinIO แทน S3 จึงไม่ต้องใช้ AWS ตอนพัฒนา

### 9.2 สร้าง state bucket (ทำครั้งเดียว)

```bash
aws s3api create-bucket --bucket <STATE_BUCKET> --region ap-southeast-1 \
  --create-bucket-configuration LocationConstraint=ap-southeast-1
aws s3api put-bucket-versioning --bucket <STATE_BUCKET> \
  --versioning-configuration Status=Enabled
```

### 9.3 สร้าง infrastructure

```bash
cd iac
cp terraform.tfvars.example terraform.tfvars   # ใส่ admin_cidr = "<IP ของคุณ>/32"
terraform init
terraform plan
terraform apply
terraform output
```

### 9.4 เข้าเครื่องและติดตั้ง platform

```bash
aws ssm start-session --target $(terraform output -raw instance_id)
# บนเครื่อง:
sudo bash /opt/bootstrap/bootstrap-cluster.sh
```

สคริปต์จะติดตั้งตามลำดับนี้

1. **ecr-credential-provider** ให้ K3s ดึง image จาก ECR ได้ (token ของ ECR อายุแค่ 12 ชม. provider จะขอใหม่ให้เอง)
2. **cert-manager** + `ClusterIssuer` ของ Let's Encrypt
3. **Argo CD** + `Application` ที่ชี้ไปที่ `k8s/overlays/prod`
4. **Jenkins** ผ่าน Helm ด้วย `k8s/platform/jenkins-values.yaml`

### 9.5 ตั้งค่า GitHub

| ตั้งค่า | ค่า |
| --- | --- |
| Branch protection (`main`) | ต้องผ่าน PR, ต้องผ่าน status check `continuous-integration/jenkins/pr-merge`, ห้าม force push |
| Webhook | `https://ci.example.com/github-webhook/`, content type JSON, ใส่ secret |
| Fine-grained token (สำหรับ Jenkins) | เฉพาะ repo นี้: Contents (read/write), Pull requests (read), Commit statuses (read/write) |
| Deploy key (สำหรับ Argo CD) | read-only |

### 9.6 ตั้งค่า Jenkins Credentials

| ID | ชนิด | ใช้ทำอะไร |
| --- | --- | --- |
| `github-app-token` | Username with password (`x-access-token` / token) | สแกน repo และ push commit tag |
| `github-webhook-secret` | Secret text | ตรวจลายเซ็น webhook |
| `discord-webhook` | Secret text | ส่งแจ้งเตือน |
| `aws-tf-readonly` | Username with password (access key / secret) | `terraform plan` และ drift detect (สิทธิ์ ReadOnlyAccess) |

จากนั้นสร้าง **Multibranch Pipeline** ชี้ไปที่ repo นี้ และเปิด "Discover pull requests from origin"

### 9.7 Deploy ครั้งแรก

เปิด PR เล็กๆ รอให้ขึ้น ✔ แล้ว merge ดูผลที่ Jenkins → Argo CD → `https://app.example.com`

### 9.8 ลบทิ้งหลังเดโม

```bash
cd iac && terraform destroy
```

---

## 🧩 10. ไฟล์ config สำคัญ

### 10.1 `Jenkinsfile`

```groovy
pipeline {
  agent {
    kubernetes {
      yamlFile 'ci/agent-pod.yaml'
      defaultContainer 'tools'
    }
  }

  options {
    timeout(time: 30, unit: 'MINUTES')
    disableConcurrentBuilds()
    buildDiscarder(logRotator(numToKeepStr: '30'))
  }

  environment {
    AWS_REGION    = 'ap-southeast-1'
    ECR_REGISTRY  = '<ACCOUNT_ID>.dkr.ecr.ap-southeast-1.amazonaws.com'
    BACKEND_REPO  = "${ECR_REGISTRY}/thai-gov-backend"
    FRONTEND_REPO = "${ECR_REGISTRY}/thai-gov-frontend"
    DOCKER_CONFIG = "${WORKSPACE}/.docker"
  }

  stages {
    stage('Prepare') {
      steps {
        scmSkip(deleteBuild: true, skipPattern: '.*\\[skip ci\\].*')
        sh '''
          apk add --no-cache git yq curl >/dev/null
          git config --global --add safe.directory '*'
        '''
        script {
          env.TAG = sh(script: 'git rev-parse --short=7 HEAD', returnStdout: true).trim()
        }
      }
    }

    stage('Lint & Test') {
      parallel {
        stage('Backend') {
          steps {
            container('golang') {
              dir('backend') {
                sh '''
                  apt-get update -qq && apt-get install -y -qq --no-install-recommends libvips-dev >/dev/null
                  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b /usr/local/bin v1.59.1
                  go vet ./...
                  golangci-lint run ./...
                  go test -race -coverprofile=coverage.out ./...
                  go tool cover -func=coverage.out | tail -1
                '''
              }
            }
          }
        }
        stage('Frontend') {
          steps {
            container('node') {
              dir('frontend') {
                sh '''
                  npm ci
                  npm run lint
                  npx tsc --noEmit
                  npm test --if-present
                '''
              }
            }
          }
        }
      }
    }

    stage('Build images') {
      steps {
        container('buildkit') {
          sh '''
            for svc in backend frontend; do
              repo=$([ "$svc" = backend ] && echo "$BACKEND_REPO" || echo "$FRONTEND_REPO")
              buildctl-daemonless.sh build \
                --frontend dockerfile.v0 \
                --local context=$svc --local dockerfile=$svc \
                --output type=docker,name=$repo:$TAG,dest=$svc.tar
            done
          '''
        }
      }
    }

    stage('Security gate') {
      steps {
        container('trivy') {
          sh '''
            trivy fs --scanners vuln,secret --severity CRITICAL --exit-code 1 --no-progress .
            trivy image --input backend.tar  --severity CRITICAL --ignore-unfixed --exit-code 1 --no-progress
            trivy image --input frontend.tar --severity CRITICAL --ignore-unfixed --exit-code 1 --no-progress
          '''
        }
      }
    }

    stage('Terraform plan') {
      when { allOf { changeRequest(); changeset 'iac/**' } }
      steps {
        container('terraform') {
          withCredentials([usernamePassword(credentialsId: 'aws-tf-readonly',
                           usernameVariable: 'AWS_ACCESS_KEY_ID',
                           passwordVariable: 'AWS_SECRET_ACCESS_KEY')]) {
            sh '''
              set -o pipefail
              terraform -chdir=iac init -input=false
              terraform -chdir=iac fmt -check -recursive
              terraform -chdir=iac validate
              terraform -chdir=iac plan -input=false -lock=false -no-color | tee tfplan.txt
            '''
          }
        }
      }
      post { always { archiveArtifacts artifacts: 'tfplan.txt', allowEmptyArchive: true } }
    }

    stage('Push to ECR') {
      when { branch 'main' }
      steps {
        container('aws') {
          sh '''
            mkdir -p "$DOCKER_CONFIG"
            TOKEN=$(aws ecr get-login-password --region "$AWS_REGION")
            AUTH=$(printf 'AWS:%s' "$TOKEN" | base64 | tr -d '\\n')
            printf '{"auths":{"%s":{"auth":"%s"}}}' "$ECR_REGISTRY" "$AUTH" > "$DOCKER_CONFIG/config.json"
          '''
        }
        container('crane') {
          sh '''
            crane push backend.tar  "$BACKEND_REPO:$TAG"
            crane push frontend.tar "$FRONTEND_REPO:$TAG"
          '''
        }
      }
    }

    stage('Update manifest') {
      when { branch 'main' }
      steps {
        withCredentials([usernamePassword(credentialsId: 'github-app-token',
                         usernameVariable: 'GH_USER', passwordVariable: 'GH_TOKEN')]) {
          sh '''
            cd k8s/overlays/prod
            yq -i '(.images[] | select(.name == "backend")).newTag  = strenv(TAG)' kustomization.yaml
            yq -i '(.images[] | select(.name == "frontend")).newTag = strenv(TAG)' kustomization.yaml
            cd -
            git config user.name  "jenkins-bot"
            git config user.email "jenkins-bot@users.noreply.github.com"
            git add k8s/overlays/prod/kustomization.yaml
            git commit -m "deploy: $TAG [skip ci]"
            git push "https://x-access-token:${GH_TOKEN}@github.com/<OWNER>/thai-gov-processor.git" HEAD:main
          '''
        }
      }
    }
  }

  post {
    success {
      withCredentials([string(credentialsId: 'discord-webhook', variable: 'DISCORD_URL')]) {
        sh 'curl -fsS -F "content=✅ ${JOB_NAME} #${BUILD_NUMBER} (${TAG}) ผ่าน" "$DISCORD_URL"'
      }
    }
    failure {
      withCredentials([string(credentialsId: 'discord-webhook', variable: 'DISCORD_URL')]) {
        sh 'curl -fsS -F "content=❌ ${JOB_NAME} #${BUILD_NUMBER} ล้มเหลว ${BUILD_URL}" "$DISCORD_URL"'
      }
    }
    always { cleanWs() }
  }
}
```

> **Plugin ที่ต้องมี:** Kubernetes, Pipeline, Git, GitHub Branch Source, Credentials Binding, SCM Skip, Workspace Cleanup, Pipeline Stage View

### 10.2 `ci/agent-pod.yaml`

```yaml
apiVersion: v1
kind: Pod
metadata:
  labels:
    role: jenkins-agent
spec:
  serviceAccountName: jenkins-agent        # ไม่มีสิทธิ์ใน namespace production
  containers:
    - name: tools
      image: alpine:3.20
      command: ["cat"]
      tty: true
    - name: golang
      image: golang:1.22-bookworm
      command: ["cat"]
      tty: true
      resources:
        requests: { cpu: "500m", memory: "512Mi" }
        limits:   { cpu: "1500m", memory: "1536Mi" }
    - name: node
      image: node:20-bookworm-slim
      command: ["cat"]
      tty: true
      resources:
        requests: { cpu: "250m", memory: "512Mi" }
        limits:   { cpu: "1000m", memory: "1536Mi" }
    - name: buildkit
      image: moby/buildkit:v0.15.1-rootless
      command: ["cat"]
      tty: true
      env:
        - name: BUILDKITD_FLAGS
          value: --oci-worker-no-process-sandbox
      securityContext:
        runAsUser: 1000
        runAsGroup: 1000
        seccompProfile:  { type: Unconfined }
        appArmorProfile: { type: Unconfined }
      resources:
        limits: { cpu: "2000m", memory: "3Gi" }
    - name: trivy
      image: aquasec/trivy:0.54.1
      command: ["cat"]
      tty: true
    - name: terraform
      image: hashicorp/terraform:1.9
      command: ["cat"]
      tty: true
    - name: aws
      image: amazon/aws-cli:2.17.0
      command: ["cat"]
      tty: true
    - name: crane
      image: gcr.io/go-containerregistry/crane:debug
      command: ["cat"]
      tty: true
```

> **ปรับเวอร์ชันของ image ทุกตัวให้เป็นเวอร์ชันล่าสุดที่คุณทดสอบแล้ว** ห้ามใช้ `latest` ใน CI เพราะผลลัพธ์จะเปลี่ยนไปเองโดยไม่รู้ตัว

### 10.3 Kubernetes manifests

**`k8s/base/backend-deployment.yaml`**

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: backend
  labels: { app: backend }
spec:
  replicas: 2
  revisionHistoryLimit: 5
  progressDeadlineSeconds: 180
  strategy:
    type: RollingUpdate
    rollingUpdate: { maxSurge: 1, maxUnavailable: 0 }
  selector:
    matchLabels: { app: backend }
  template:
    metadata:
      labels: { app: backend }
    spec:
      securityContext:
        runAsNonRoot: true
        runAsUser: 10001
      containers:
        - name: backend
          image: backend                    # kustomize แทนค่าเป็น ECR:tag
          ports: [{ containerPort: 8080 }]
          env:
            - { name: AWS_REGION, value: ap-southeast-1 }
            - { name: S3_BUCKET,  valueFrom: { configMapKeyRef: { name: app-config, key: s3_bucket } } }
            - { name: PRESIGN_TTL, value: "1h" }
          readinessProbe:
            httpGet: { path: /healthz, port: 8080 }
            periodSeconds: 5
          livenessProbe:
            httpGet: { path: /healthz, port: 8080 }
            initialDelaySeconds: 10
            periodSeconds: 10
          resources:
            requests: { cpu: "250m", memory: "256Mi" }
            limits:   { cpu: "1",    memory: "512Mi" }
          securityContext:
            allowPrivilegeEscalation: false
            readOnlyRootFilesystem: true
            capabilities: { drop: ["ALL"] }
          volumeMounts:
            - { name: tmp, mountPath: /tmp }
      volumes:
        - { name: tmp, emptyDir: { sizeLimit: 256Mi } }
```

**`k8s/base/backend-hpa.yaml`**

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: backend
spec:
  scaleTargetRef: { apiVersion: apps/v1, kind: Deployment, name: backend }
  minReplicas: 2
  maxReplicas: 4
  metrics:
    - type: Resource
      resource:
        name: cpu
        target: { type: Utilization, averageUtilization: 70 }
```

**`k8s/base/ingress.yaml`** (frontend และ backend อยู่ namespace เดียวกับ Ingress)

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: app
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
    traefik.ingress.kubernetes.io/router.entrypoints: websecure
spec:
  ingressClassName: traefik
  tls:
    - hosts: [app.example.com]
      secretName: app-tls
  rules:
    - host: app.example.com
      http:
        paths:
          - path: /api
            pathType: Prefix
            backend: { service: { name: backend, port: { number: 8080 } } }
          - path: /
            pathType: Prefix
            backend: { service: { name: frontend, port: { number: 3000 } } }
```

**`k8s/base/smoke-test.yaml`**

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  generateName: smoke-test-
  annotations:
    argocd.argoproj.io/hook: PostSync
    argocd.argoproj.io/hook-delete-policy: BeforeHookCreation
spec:
  backoffLimit: 2
  template:
    spec:
      restartPolicy: Never
      containers:
        - name: curl
          image: curlimages/curl:8.9.1
          command: ["sh", "-c"]
          args:
            - |
              set -e
              curl -fsS http://backend:8080/healthz
              curl -fsS http://backend:8080/api/v1/selftest
              curl -fsS http://frontend:3000/ >/dev/null
              echo "smoke test passed"
```

**`k8s/overlays/prod/kustomization.yaml`** (Jenkins แก้ `newTag` ที่ไฟล์นี้)

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: production
resources:
  - ../../base
images:
  - name: backend
    newName: <ACCOUNT_ID>.dkr.ecr.ap-southeast-1.amazonaws.com/thai-gov-backend
    newTag: a1b2c3d
  - name: frontend
    newName: <ACCOUNT_ID>.dkr.ecr.ap-southeast-1.amazonaws.com/thai-gov-frontend
    newTag: a1b2c3d
```

### 10.4 `k8s/argocd/application.yaml`

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: thai-gov-processor
  namespace: argocd
  annotations:
    notifications.argoproj.io/subscribe.on-deployed.discord: ""
    notifications.argoproj.io/subscribe.on-sync-failed.discord: ""
spec:
  project: default
  source:
    repoURL: git@github.com:<OWNER>/thai-gov-processor.git
    targetRevision: main
    path: k8s/overlays/prod
  destination:
    server: https://kubernetes.default.svc
    namespace: production
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
      - CreateNamespace=true
```

> **Argo CD ไม่ rollback ให้เองเมื่อ smoke test ไม่ผ่าน** มันจะแจ้ง sync failed แล้วเราทำ `git revert` ส่วนระหว่าง rolling update ถ้า pod ใหม่ไม่ ready ตัว Deployment จะไม่ลบ pod เก่า เว็บจึงไม่ล่ม ถ้าต้องการ rollback อัตโนมัติเต็มรูปแบบ ดู Argo Rollouts ใน roadmap

### 10.5 `k8s/platform/credential-provider.yaml`

วางไฟล์นี้ไว้ที่ `/var/lib/rancher/credentialprovider/config.yaml` และวาง binary `ecr-credential-provider` (จาก `kubernetes/cloud-provider-aws`) ไว้ที่ `/var/lib/rancher/credentialprovider/bin/` ซึ่งเป็น path เริ่มต้นที่ K3s ใช้หา credential provider

```yaml
apiVersion: kubelet.config.k8s.io/v1
kind: CredentialProviderConfig
providers:
  - name: ecr-credential-provider
    matchImages:
      - "*.dkr.ecr.*.amazonaws.com"
    defaultCacheDuration: "12h"
    apiVersion: credentialprovider.kubelet.k8s.io/v1
```

### 10.6 Terraform

**`iac/versions.tf`**

```hcl
terraform {
  required_version = ">= 1.10.0"
  required_providers {
    aws = { source = "hashicorp/aws", version = "~> 5.60" }
  }
  backend "s3" {
    bucket       = "<STATE_BUCKET>"
    key          = "thai-gov-processor/terraform.tfstate"
    region       = "ap-southeast-1"
    encrypt      = true
    use_lockfile = true
  }
}

provider "aws" {
  region = var.aws_region
  default_tags {
    tags = { Project = "thai-gov-processor", ManagedBy = "terraform" }
  }
}
```

**`iac/variables.tf`**

```hcl
variable "aws_region"    { default = "ap-southeast-1" }
variable "instance_type" { default = "t3a.large" }
variable "admin_cidr" {
  description = "IP ของผู้ดูแลที่เข้า K3s API ได้ เช่น 203.0.113.10/32"
  type        = string
}
```

**`iac/network.tf`**

```hcl
resource "aws_vpc" "main" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
}

resource "aws_internet_gateway" "main" { vpc_id = aws_vpc.main.id }

resource "aws_subnet" "public" {
  vpc_id                  = aws_vpc.main.id
  cidr_block              = "10.0.1.0/24"
  availability_zone       = "${var.aws_region}a"
  map_public_ip_on_launch = false
}

resource "aws_route_table" "public" {
  vpc_id = aws_vpc.main.id
  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.main.id
  }
}

resource "aws_route_table_association" "public" {
  subnet_id      = aws_subnet.public.id
  route_table_id = aws_route_table.public.id
}
```

**`iac/security_group.tf`**

```hcl
resource "aws_security_group" "k3s" {
  name   = "k3s-node"
  vpc_id = aws_vpc.main.id

  ingress {
    description = "HTTP (redirect to HTTPS + ACME challenge)"
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    description = "HTTPS"
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    description = "K3s API - admin only"
    from_port   = 6443
    to_port     = 6443
    protocol    = "tcp"
    cidr_blocks = [var.admin_cidr]
  }

  # ไม่เปิด port 22: เข้าเครื่องผ่าน SSM Session Manager

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}
```

**`iac/iam.tf`**

```hcl
data "aws_iam_policy_document" "assume_ec2" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["ec2.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "node" {
  name               = "k3s-node-role"
  assume_role_policy = data.aws_iam_policy_document.assume_ec2.json
}

data "aws_iam_policy_document" "node" {
  statement {
    sid       = "EcrAuth"
    actions   = ["ecr:GetAuthorizationToken"]
    resources = ["*"]
  }
  statement {
    sid = "EcrPushPull"
    actions = [
      "ecr:BatchCheckLayerAvailability", "ecr:BatchGetImage", "ecr:GetDownloadUrlForLayer",
      "ecr:InitiateLayerUpload", "ecr:UploadLayerPart", "ecr:CompleteLayerUpload", "ecr:PutImage",
    ]
    resources = [aws_ecr_repository.backend.arn, aws_ecr_repository.frontend.arn]
  }
  statement {
    sid       = "UserFiles"
    actions   = ["s3:PutObject", "s3:GetObject", "s3:DeleteObject"]
    resources = ["${aws_s3_bucket.files.arn}/*"]
  }
}

resource "aws_iam_role_policy" "node" {
  role   = aws_iam_role.node.id
  policy = data.aws_iam_policy_document.node.json
}

resource "aws_iam_role_policy_attachment" "ssm" {
  role       = aws_iam_role.node.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
}

resource "aws_iam_instance_profile" "node" {
  name = "k3s-node-profile"
  role = aws_iam_role.node.name
}
```

**`iac/ec2.tf`**

```hcl
data "aws_ami" "ubuntu" {
  most_recent = true
  owners      = ["099720109477"] # Canonical
  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd-gp3/ubuntu-noble-24.04-amd64-server-*"]
  }
}

resource "aws_eip" "k3s" { domain = "vpc" }

resource "aws_instance" "k3s" {
  ami                    = data.aws_ami.ubuntu.id
  instance_type          = var.instance_type
  subnet_id              = aws_subnet.public.id
  vpc_security_group_ids = [aws_security_group.k3s.id]
  iam_instance_profile   = aws_iam_instance_profile.node.name

  metadata_options {
    http_tokens                 = "required" # บังคับ IMDSv2
    http_put_response_hop_limit = 2          # ให้ pod ใช้ IAM role ของเครื่องได้
  }

  root_block_device {
    volume_size = 35
    volume_type = "gp3"
    encrypted   = true
  }

  user_data = templatefile("${path.module}/templates/user_data.sh.tftpl", {
    public_ip = aws_eip.k3s.public_ip
  })

  tags = { Name = "k3s-thai-gov-processor" }
}

resource "aws_eip_association" "k3s" {
  instance_id   = aws_instance.k3s.id
  allocation_id = aws_eip.k3s.id
}
```

**`iac/templates/user_data.sh.tftpl`**

```bash
#!/bin/bash
set -euo pipefail
apt-get update -y
apt-get install -y curl git

curl -sfL https://get.k3s.io | \
  INSTALL_K3S_EXEC="server --tls-san ${public_ip} --write-kubeconfig-mode 600" sh -

curl -fsSL https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
```

**`iac/ecr.tf`**

```hcl
locals { repos = { backend = "thai-gov-backend", frontend = "thai-gov-frontend" } }

resource "aws_ecr_repository" "backend" {
  name                 = local.repos.backend
  image_tag_mutability = "IMMUTABLE"
  force_delete         = true
  image_scanning_configuration { scan_on_push = true }
}

resource "aws_ecr_repository" "frontend" {
  name                 = local.repos.frontend
  image_tag_mutability = "IMMUTABLE"
  force_delete         = true
  image_scanning_configuration { scan_on_push = true }
}

resource "aws_ecr_lifecycle_policy" "keep_last_10" {
  for_each   = { backend = aws_ecr_repository.backend.name, frontend = aws_ecr_repository.frontend.name }
  repository = each.value
  policy = jsonencode({
    rules = [{
      rulePriority = 1
      description  = "keep last 10 images"
      selection    = { tagStatus = "any", countType = "imageCountMoreThan", countNumber = 10 }
      action       = { type = "expire" }
    }]
  })
}
```

**`iac/s3.tf`**

```hcl
resource "aws_s3_bucket" "files" {
  bucket_prefix = "thai-gov-files-"
  force_destroy = true
}

resource "aws_s3_bucket_public_access_block" "files" {
  bucket                  = aws_s3_bucket.files.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_server_side_encryption_configuration" "files" {
  bucket = aws_s3_bucket.files.id
  rule {
    apply_server_side_encryption_by_default { sse_algorithm = "AES256" }
  }
}

resource "aws_s3_bucket_lifecycle_configuration" "files" {
  bucket = aws_s3_bucket.files.id
  rule {
    id     = "expire-user-files"
    status = "Enabled"
    filter {}
    expiration { days = 1 }
    abort_incomplete_multipart_upload { days_after_initiation = 1 }
  }
}
```

**`iac/outputs.tf`**

```hcl
output "instance_id"  { value = aws_instance.k3s.id }
output "public_ip"    { value = aws_eip.k3s.public_ip }
output "files_bucket" { value = aws_s3_bucket.files.bucket }
output "ecr_backend"  { value = aws_ecr_repository.backend.repository_url }
output "ecr_frontend" { value = aws_ecr_repository.frontend.repository_url }
```

**`ci/Jenkinsfile.drift`** (job ตั้งเวลาทุกคืน)

```groovy
pipeline {
  agent { kubernetes { yamlFile 'ci/agent-pod.yaml'; defaultContainer 'terraform' } }
  triggers { cron('TZ=Asia/Bangkok\nH 2 * * *') }
  stages {
    stage('Drift detect') {
      steps {
        withCredentials([usernamePassword(credentialsId: 'aws-tf-readonly',
                         usernameVariable: 'AWS_ACCESS_KEY_ID', passwordVariable: 'AWS_SECRET_ACCESS_KEY')]) {
          script {
            sh 'terraform -chdir=iac init -input=false'
            def code = sh(script: 'terraform -chdir=iac plan -input=false -lock=false -detailed-exitcode',
                          returnStatus: true)
            if (code == 2) { unstable('พบ drift: มีการแก้ AWS นอก Terraform') }
            else if (code != 0) { error('terraform plan ล้มเหลว') }
          }
        }
      }
    }
  }
  post {
    unstable {
      withCredentials([string(credentialsId: 'discord-webhook', variable: 'DISCORD_URL')]) {
        container('tools') {
          sh 'apk add --no-cache curl >/dev/null && curl -fsS -F "content=⚠️ Terraform drift detected ${BUILD_URL}" "$DISCORD_URL"'
        }
      }
    }
  }
}
```

### 10.7 `k8s/platform/jenkins-values.yaml` (ย่อ)

```yaml
controller:
  resources:
    requests: { cpu: "500m", memory: "1Gi" }
    limits:   { cpu: "1",    memory: "2Gi" }
  javaOpts: "-Xmx1g"
  installPlugins:
    - kubernetes
    - workflow-aggregator
    - git
    - github-branch-source
    - credentials-binding
    - configuration-as-code
    - pipeline-stage-view
    - scm-skip
    - ws-cleanup
  ingress:
    enabled: true
    ingressClassName: traefik
    hostName: ci.example.com
    annotations:
      cert-manager.io/cluster-issuer: letsencrypt-prod
    tls:
      - hosts: [ci.example.com]
        secretName: jenkins-tls
persistence:
  size: 10Gi
serviceAccountAgent:
  create: true
  name: jenkins-agent
```

### 10.8 Dockerfile ของ backend

```dockerfile
FROM golang:1.22-bookworm AS build
RUN apt-get update && apt-get install -y --no-install-recommends libvips-dev \
 && rm -rf /var/lib/apt/lists/*
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends libvips42 ca-certificates \
 && rm -rf /var/lib/apt/lists/* \
 && useradd -r -u 10001 -s /usr/sbin/nologin app
COPY --from=build /out/api /usr/local/bin/api
USER 10001
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/api"]
```

---

## 📡 11. REST API

| Method | Path | ใช้ทำอะไร |
| --- | --- | --- |
| `GET` | `/healthz` | liveness/readiness probe |
| `GET` | `/api/v1/selftest` | แปลงรูปตัวอย่างที่ฝังไว้ แล้วเช็คว่าได้ขนาดตรง preset (ใช้ใน smoke test) |
| `GET` | `/api/v1/presets` | รายการ preset ทั้งหมด |
| `POST` | `/api/v1/photos/preset` | แปลงรูปตาม preset |
| `POST` | `/api/v1/documents/merge-pdf` | รวมหลายภาพเป็น PDF เดียว |

### `POST /api/v1/photos/preset`

`Content-Type: multipart/form-data`

| field | ค่า |
| --- | --- |
| `file` | JPG, PNG, HEIC, WEBP (สูงสุด 15 MB) |
| `preset` | `ocsc` \| `passport` \| `teacher` \| `custom` |
| `width`, `height`, `max_kb` | ใช้เฉพาะ `custom` |

```json
{
  "status": "success",
  "data": {
    "filename": "ocsc_photo_a8f3c1.jpg",
    "width": 200,
    "height": 230,
    "size_kb": 84.6,
    "quality": 82,
    "mime_type": "image/jpeg",
    "download_url": "https://<bucket>.s3.ap-southeast-1.amazonaws.com/processed/ocsc_photo_a8f3c1.jpg?X-Amz-Signature=...",
    "expires_in": 3600
  }
}
```

### `POST /api/v1/documents/merge-pdf`

| field | ค่า |
| --- | --- |
| `files[]` | หลายภาพ หรือ PDF ที่สแกนมา |
| `target_max_kb` | ค่าเริ่มต้น `500` |

```json
{
  "status": "success",
  "data": {
    "filename": "id_card_merged.pdf",
    "total_pages": 2,
    "size_kb": 412.3,
    "download_url": "https://<bucket>.s3.ap-southeast-1.amazonaws.com/processed/id_card_merged.pdf?X-Amz-Signature=...",
    "expires_in": 3600
  }
}
```

**Error ที่อาจเกิด:** `400` ไฟล์ไม่รองรับ · `413` ไฟล์ใหญ่เกิน · `422` บีบให้ต่ำกว่าเกณฑ์ไม่ได้โดยไม่เสียคุณภาพ

---

## ⚡ 12. Backend: อัลกอริทึมบีบไฟล์

ใช้ **binary search** หาค่า JPEG quality ที่สูงที่สุดซึ่งทำให้ไฟล์ยังไม่เกินเกณฑ์ ช่วง quality 30–95 ใช้ไม่เกิน 7 รอบ (log₂66 ≈ 6.04)

```go
package processor

import (
	"errors"
	"fmt"

	"github.com/h2non/bimg"
)

type ProcessOptions struct {
	TargetWidth  int
	TargetHeight int
	MaxSizeBytes int // เช่น 100 * 1024
}

var ErrCannotCompress = errors.New("cannot compress image below target size without extreme degradation")

// ProcessPhoto ย่อภาพตามขนาดที่กำหนด แล้วหา quality สูงสุดที่ไฟล์ยัง <= MaxSizeBytes
func ProcessPhoto(input []byte, opts ProcessOptions) ([]byte, int, error) {
	resized, err := bimg.NewImage(input).Process(bimg.Options{
		Width:   opts.TargetWidth,
		Height:  opts.TargetHeight,
		Crop:    true,
		Gravity: bimg.GravitySmart,
		Type:    bimg.JPEG,
		Quality: 100,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("resize: %w", err)
	}

	low, high := 30, 95
	var best []byte
	bestQ := 0

	for low <= high {
		mid := (low + high) / 2
		out, err := bimg.NewImage(resized).Process(bimg.Options{Quality: mid, Type: bimg.JPEG, StripMetadata: true})
		if err != nil {
			return nil, 0, fmt.Errorf("compress q=%d: %w", mid, err)
		}
		if len(out) <= opts.MaxSizeBytes {
			best, bestQ = out, mid // ผ่านเกณฑ์: ลองเพิ่ม quality
			low = mid + 1
		} else {
			high = mid - 1 // ใหญ่เกิน: ลด quality
		}
	}

	if best == nil {
		return nil, 0, ErrCannotCompress
	}
	return best, bestQ, nil
}
```

> `StripMetadata: true` ลบ EXIF ออก ทั้งช่วยลดขนาดไฟล์และลบข้อมูลส่วนตัว เช่น พิกัด GPS

---

## 🔐 13. Security และ PDPA

### 13.1 ใครมีสิทธิ์ทำอะไร

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

### 13.2 Security controls ตามชั้น

| ชั้น | สิ่งที่ทำ |
| --- | --- |
| โค้ด | lint, unit test, Trivy secret scan (กัน AWS key หลุดเข้า Git) |
| Supply chain | Trivy gate สองรอบ, push ไฟล์ตัวเดียวกับที่สแกน, ECR IMMUTABLE + scan on push, pin เวอร์ชัน image ใน CI |
| Container | non-root, read-only root filesystem, drop capabilities ทั้งหมด, resource limits |
| Cluster | CI ไม่มีสิทธิ์ใน production, มีแค่ Argo CD ที่ deploy ได้, self-heal กันแก้ด้วยมือ |
| Network | เปิดแค่ 80/443, K3s API เฉพาะ IP ผู้ดูแล, ไม่เปิด SSH |
| Host | IMDSv2 บังคับ, EBS เข้ารหัส, Ubuntu LTS |
| Data | S3 block public access, SSE, presigned URL อายุ 1 ชม., ลบ EXIF |

### 13.3 PDPA: ข้อมูลอยู่นานแค่ไหน

| ข้อมูล | อยู่ที่ไหน | อยู่นานเท่าไร |
| --- | --- | --- |
| ไฟล์ต้นฉบับ | `uploads/` ใน S3 | backend ลบทันทีหลังประมวลผลเสร็จ, lifecycle เป็นตัวสำรอง |
| ไฟล์ที่แปลงแล้ว | `processed/` ใน S3 | ลิงก์ดาวน์โหลดใช้ได้ 1 ชม., ไฟล์ถูกลบโดย lifecycle |
| ไฟล์ระหว่างประมวลผล | `emptyDir` ใน pod | หายไปเมื่อ request จบ หรือเมื่อ pod ถูกลบ |
| Log | stdout ของ pod | ไม่ log ชื่อไฟล์เดิมหรือเนื้อหาไฟล์ |

> **หมายเหตุเรื่อง lifecycle:** S3 นับวันหมดอายุโดยปัดไปเที่ยงคืน UTC และลบแบบ asynchronous ไฟล์จึงอาจค้างอยู่ประมาณ 1–2 วัน ไม่ใช่ 24 ชั่วโมงเป๊ะ ด้วยเหตุนี้ backend จึงลบไฟล์ต้นฉบับเองทันที และใช้ lifecycle เป็นตาข่ายรองรับอีกชั้น

---

## 💰 14. ค่าใช้จ่าย

ตัวเลขด้านล่างเป็นค่าประมาณสำหรับ region Singapore **ควรตรวจราคาล่าสุดด้วย [AWS Pricing Calculator](https://calculator.aws/)** ก่อนใช้งานจริง

| รายการ | รายละเอียด | เปิดทิ้งไว้ 1 วัน | เปิดทิ้งไว้ 1 เดือน |
| --- | --- | --- | --- |
| EC2 t3a.large | on-demand ราว $0.09/ชม. | ~$2.2 | ~$66 |
| EBS gp3 35 GB | | ~$0.1 | ~$3.4 |
| Public IPv4 (Elastic IP) | $0.005/ชม. | ~$0.12 | ~$3.6 |
| S3 + ECR | ไฟล์ชั่วคราวและ image < 1 GB | ~$0 | < $1 |
| EKS control plane | **ไม่ได้ใช้** (ใช้ K3s แทน) | $0 | ประหยัด ~$73 |
| **รวมโดยประมาณ** | | **~$2.5 (~85 บาท)** | **~$74 (~2,500 บาท)** |

**วิธีคุมงบ**

- `terraform apply` ก่อนเดโมหรือสัมภาษณ์ แล้ว `terraform destroy` ทันทีที่เสร็จ ค่าใช้จ่ายจะเหลือหลักสิบบาทต่อครั้ง
- ตั้ง **AWS Budgets** แจ้งเตือนเมื่อค่าใช้จ่ายเกิน $10/เดือน
- state bucket และ ECR ที่เหลือหลัง destroy มีค่าใช้จ่ายแทบเป็นศูนย์

---

## 🧭 15. ข้อจำกัดและ roadmap

### ข้อจำกัดที่รู้อยู่แล้ว

| ข้อจำกัด | ผลกระทบ | ถ้าจะแก้ |
| --- | --- | --- |
| Single node | เครื่องล่ม = เว็บล่ม, HPA ขยายได้แค่ในเครื่องเดียว | EKS หรือ K3s หลาย node + ALB |
| ทุก pod ใช้ IAM role ของเครื่องร่วมกัน | pod ใดก็ได้เรียกสิทธิ์ push ECR ได้ | EKS Pod Identity / IRSA แยก role ต่อ ServiceAccount |
| Jenkins อยู่เครื่องเดียวกับแอป | build หนักๆ อาจแย่ง CPU/RAM จากเว็บ | resource limit (ทำแล้ว) หรือแยก node สำหรับ CI |
| Jenkins เปิดสู่ internet เพื่อรับ webhook | เป็นเป้าโจมตี | จำกัด IP ให้เฉพาะช่วง IP webhook ของ GitHub หรือใช้ smee/relay |
| Rollback ต้องทำเอง | ต้องมีคน `git revert` | Argo Rollouts + analysis อัตโนมัติ |

### Roadmap

- [x] Terraform + remote state
- [x] Jenkins PR/main pipeline + Trivy gate
- [x] Argo CD GitOps + smoke test
- [ ] Namespace `staging` + promote ไป `production` ด้วย PR
- [ ] Observability: Prometheus + Grafana (แบบเบา) และ dashboard latency/error rate
- [ ] Argo Rollouts (canary) + rollback อัตโนมัติ
- [ ] Sign image ด้วย cosign + verify ตอน deploy
- [ ] Policy as code (Kyverno): บังคับ non-root, ห้ามใช้ tag `latest`
- [ ] Load test ด้วย k6 แล้วดู HPA ขยาย pod จริง

---

## 📸 16. หลักฐานการทำงานและบทเรียน

> แทนที่ภาพในส่วนนี้ด้วย screenshot จริงจากระบบของคุณ

| หลักฐาน | ภาพ |
| --- | --- |
| Jenkins stage view (PR ผ่าน) | `docs/images/jenkins-pr-pass.png` |
| PR ที่ Trivy บล็อก | `docs/images/trivy-blocked.png` |
| Argo CD แสดง Synced / Healthy | `docs/images/argocd-healthy.png` |
| Rollback ด้วย `git revert` | `docs/images/rollback.gif` |
| แจ้งเตือนใน Discord | `docs/images/discord.png` |
| `terraform apply` / `destroy` | `docs/images/terraform.png` |

### ปัญหาที่เจอและวิธีแก้ (บันทึกระหว่างทำ)

| ปัญหา | สาเหตุ | วิธีแก้ |
| --- | --- | --- |
| _ตัวอย่าง:_ pod ขึ้น `ImagePullBackOff` หลังผ่านไป 12 ชม. | token ของ ECR หมดอายุ | ติดตั้ง ecr-credential-provider |
| _ตัวอย่าง:_ pipeline วนรันไม่จบ | Jenkins commit tag แล้ว trigger ตัวเองซ้ำ | ใส่ `[skip ci]` + SCM Skip plugin |
| _เพิ่มของคุณเอง_ | | |

---

## 📄 License

MIT

%% ================= 1. CLIENT & UI FLOW =================
    subgraph CLIENT_TIER ["💻 Client & Frontend Tier"]
        User(["👤 Job Applicant / End User"]):::client
        
        subgraph NEXTJS_APP ["Next.js 14 Web Application (UI)"]
            UI_Home["🏠 Main Web Page: Drag & Drop File Upload"]
            UI_Presets["🎯 Presets Selector:<br/>• OCSC (200x230px, <100KB, JPG)<br/>• Passport (2x2 inch, White background)<br/>• PDF Copy Merger (<500KB)"]
            UI_Crop["✂️ Interactive Crop & Rotate Tool"]
            UI_Preview["👁️ Real-time Result Preview & Download Button"]
        end
    end

%% ================= 2. DEVELOPER & CI/CD PIPELINE =================
    subgraph DEVOPS_TIER ["⚙️ CI/CD Automation Tier (Jenkins in K3s)"]
        Dev["👨‍💻 DevOps / Backend Developer"]:::cicd
        GitHub["🐙 GitHub Repository<br/>(Monorepo: /frontend, /backend, /k8s)"]:::cicd

        subgraph JENKINS_PIPELINE ["Jenkins Controller Pod (Namespace: jenkins)"]
            Trigger["⚡ Webhook Trigger Receiver"]
            
            subgraph DYNAMIC_AGENT ["Dynamic K8s Agent Pod (Spawn on Demand)"]
                Stage_Lint["1. Lint & Unit Tests<br/>(golangci-lint / go test)"]
                Stage_Security["2. Code & CVE Scan<br/>(govulncheck / Trivy)"]
                Stage_Kaniko["3. Kaniko Container Builder<br/>(Build Docker Image without root)"]
                Stage_Deploy["4. Deployment Rollout<br/>(kubectl rollout restart)"]
            end
        end
    end

%% ================= 3. AWS CLOUD & RUNTIME INFRASTRUCTURE =================
    subgraph AWS_CLOUD ["☁️ AWS Cloud Platform (ap-southeast-1 Singapore)"]
        
        subgraph IAC_BOX ["🏗️ Infrastructure as Code"]
            Terraform["🟣 Terraform Engine<br/>• main.tf (S3, ECR)<br/>• ec2_k3s.tf (t3a.large)<br/>• security_group.tf"]:::security
        end

        ECR[("📦 Amazon ECR<br/>• Repo: thai-gov-processor<br/>• Tag: Git-SHA / Build-No")]:::storage

        subgraph EC2_INSTANCE ["🖥️ AWS EC2 Instance (t3a.large: 2 vCPU, 8GB RAM)"]
            SG["🛡️ Security Group Rules<br/>Inbound: 80, 443, 22 (SSH Restrict)"]:::security
            IAM["🔑 IAM Instance Profile<br/>• Read ECR Image<br/>• Put/Get S3 Bucket (No hardcoded keys)"]:::security

            subgraph K3S_CLUSTER ["☸️ K3s Single-Node Kubernetes Cluster"]
                Traefik["🌐 Traefik Ingress Controller (:80 / :443)<br/>SSL Termination via Let's Encrypt"]
                
                subgraph NS_FRONTEND ["Namespace: frontend"]
                    UIPod["📱 Next.js Pods (Replicas: 2)<br/>Node.js Runtime / SSR"]
                end

                subgraph NS_PROD ["Namespace: production"]
                    K8s_Service["🔀 ClusterIP Service (:8080)"]
                    
                    subgraph PODS_HPA ["Horizontal Pod Autoscaler (HPA: CPU > 70%)"]
                        Pod1["⚡ Go Worker Pod 1<br/>• bimg (Image Resize)<br/>• pdfcpu (PDF Engine)"]
                        Pod2["⚡ Go Worker Pod 2<br/>(Replica standby)"]
                    end
                end
            end
        end

        subgraph STORAGE_TIER ["🪣 Storage & Data Privacy Plane"]
            S3[("Amazon S3 Bucket<br/>• /uploads (Raw inputs)<br/>• /processed (Gov-ready outputs)")]:::aws
            S3_Lifecycle["⏳ S3 Lifecycle Policy<br/>Permanently delete all files after 24 Hours<br/>(Compliant with Thai PDPA)"]:::security
            S3 --- S3_Lifecycle
        end
    end

%% ================= PIPELINE CONNECTIONS =================
    Dev -->|1. git push| GitHub
    GitHub -->|2. Webhook Event| Trigger
    Trigger -->|3. Spawn Pod| DYNAMIC_AGENT
    Stage_Lint --> Stage_Security
    Stage_Security --> Stage_Kaniko
    Stage_Kaniko -->|4. Push Image via IAM| ECR
    Stage_Kaniko --> Stage_Deploy
    Stage_Deploy -->|5. kubectl rollout| NS_PROD
    ECR -.->|Pull Image| Pod1

%% ================= USER APPLICATION CONNECTIONS =================
    User -->|Access Web Site| UI_Home
    UI_Home --> UI_Presets
    UI_Presets --> UI_Crop
    UI_Crop -->|Upload Photo / PDF Request| Traefik
    
    Traefik -->|Route: domain.com| UIPod
    Traefik -->|Route: /api/v1/convert| K8s_Service
    Traefik -->|Route: ci.domain.com| Trigger
    
    K8s_Service --> Pod1
    K8s_Service --> Pod2
    
    Pod1 -->|Stream Temporary File| S3
    S3 -.->|Return Presigned Download URL| UI_Preview
    UI_Preview -->|Download Processed File| User

%% ================= IAC PROVISIONING =================
    Terraform -.->|Provision VPC / EC2 / S3 / ECR| EC2_INSTANCE
    Terraform -.->|Manage Bucket| S3
    Terraform -.->|Manage Registry| ECR
```

## 🛠️ 3. Tech Stack & Architectural Decisions

| ส่วนประกอบ | เทคโนโลยีที่เลือก | เหตุผลทางเทคนิค (Why this?) |
| --- | --- | --- |
| **Backend API** | **Go 1.22+** | รันเร็วระดับ Microseconds, Binary เล็ก, จัดการ Goroutines สูง, Memory Footprint ต่ำกว่า Node/Java 5-10 เท่า |
| **Image Processing** | `h2non/bimg` (libvips) | ประมวลผลภาพเร็วกว่า ImageMagick 4–8 เท่า และกิน Memory น้อยกว่ามาก |
| **PDF Processing** | `pdfcpu` | Pure Go PDF Processor จัดการ Split/Merge และ Optimize ขนาดไฟล์ได้อย่างรวดเร็ว |
| **Frontend UI** | **Next.js 14 (App Router)** | รองรับ Client-Side Image Crop/Canvas Preview และทำ Server-Side Rendering (SSR) ปลอดภัยต่อการทำ API Proxy |
| **Styling** | **Tailwind CSS + Lucide Icons** | ออกแบบ UI สะอาดตา สไตล์ราชการยุคใหม่ (GovTech) พร้อม Responsive Design 100% |
| **Container Engine** | **K3s (Kubernetes)** | น้ำหนักเบามาก (ใช้ RAM < 512MB สำหรับ K3s Base) มี Ingress (Traefik) ในตัว ไม่ต้องจ่ายค่า EKS Master Node ($73/เดือน) |
| **CI/CD Platform** | **Jenkins on K3s** | ติดตั้งผ่าน Official Helm Chart ใช้ Dynamic Kubernetes Pod Agent รันงานเฉพาะตอน Build ไม่เปลือง RAM เครื่อง |
| **Container Builder** | **Google Kaniko** | Build Docker image ภายใน Kubernetes Pod ได้โดยตรงโดยไม่ต้องใช้ Docker-in-Docker (DinD) หรือเปิดสิทธิ์ root socket (`/var/run/docker.sock`) |
| **Security Scanning** | **Trivy + govulncheck** | สแกนช่องโหว่ระดับ Dependencies และระดับ OS Packages ใน Container Image ก่อน Push |
| **Cloud Storage** | **Amazon S3** | เก็บไฟล์ชั่วคราว พร้อมตั้ง **S3 Lifecycle Rules ลบไฟล์ทิ้งอัตโนมัติภายใน 24 ชม.** เพื่อความปลอดภัยตามกฎหมาย PDPA |
| **Artifact Registry** | **Amazon ECR** | เก็บ Private Docker Images ภายใน Region เดียวกัน (ap-southeast-1) ดึง Image ฟรี ไม่เสียค่า Data Transfer |
| **Infrastructure as Code** | **Terraform** | จัดการ Lifecycle ของ Cloud ทั้งหมดแบบ Declarative สามารถสั่ง `apply` และ `destroy` เพื่อคุมงบได้ทันที |

---

## 📂 4. Project Directory Structure (Monorepo)

```text
thai-gov-processor/
├── .github/
│   └── workflows/              # GitHub Actions (ทางเลือกสำรอง หรือ Trigger Webhook)
├── backend/                    # Go API Microservice
│   ├── cmd/
│   │   └── api/
│   │       └── main.go         # Entry point, Router & Dependency Injection
│   ├── internal/
│   │   ├── handler/            # HTTP Handlers (Multipart, Presets, Health)
│   │   ├── processor/          # Core Logic: bimg resize, binary search compress, pdfcpu
│   │   ├── storage/            # S3 Client wrapper & Presigned URL generator
│   │   └── preset/             # Thai Gov Presets (ก.พ., Passport, etc.)
│   ├── Dockerfile              # Multi-stage Docker build with libvips-dev
│   ├── go.mod
│   └── go.sum
├── frontend/                   # Next.js 14 Application
│   ├── src/
│   │   ├── app/
│   │   │   ├── layout.tsx
│   │   │   ├── page.tsx        # Single-page Web App UI
│   │   │   └── api/convert/    # Next.js Proxy Route
│   │   ├── components/
│   │   │   ├── DropZone.tsx    # Drag & drop upload area
│   │   │   ├── PresetCard.tsx  # Preset selector component
│   │   │   ├── ImageCropper.tsx# Interactive canvas cropper (react-image-crop)
│   │   │   └── PreviewModal.tsx# Result visualizer & Download trigger
│   │   └── lib/
│   │       └── utils.ts
│   ├── Dockerfile              # Next.js Standalone build
│   ├── package.json
│   └── tailwind.config.ts
├── iac/                        # Terraform Infrastructure
│   ├── main.tf                 # Provider, S3 Bucket, Lifecycle Configuration
│   ├── ecr.tf                  # ECR Repositories (Backend & Frontend)
│   ├── security_group.tf       # EC2 Security Group (22, 80, 443, 6443)
│   ├── iam.tf                  # IAM Role & Instance Profile for EC2
│   ├── ec2_k3s.tf              # EC2 t3a.large Instance with UserData K3s Setup
│   ├── variables.tf
│   └── outputs.tf
├── k8s/                        # Kubernetes Manifests
│   ├── base/
│   │   ├── backend-deployment.yaml
│   │   ├── backend-service.yaml
│   │   ├── backend-hpa.yaml
│   │   ├── frontend-deployment.yaml
│   │   ├── frontend-service.yaml
│   │   └── ingress.yaml        # Traefik Ingress Routes
│   └── jenkins/
│       └── values.yaml         # Helm Values for Jenkins Controller on K3s
├── Jenkinsfile                 # Declarative Pipeline with Dynamic Kubernetes Pod Agent
└── README.md

```

---

## ⚡ 5. Backend Implementation (Go Core Processor)

Backend ใช้เทคนิค **Binary Search (ค้นหาทวิภาค)** เพื่อหาค่า Quality ของภาพ JPEG ที่ทำให้ไฟล์มีขนาดต่ำกว่าเกณฑ์สูงสุด (เช่น 100 KB) โดยยังคงความคมชัดสูงสุด และใช้เวลาประมวลผลน้อยที่สุด (ลูปไม่เกิน 7 ครั้ง):

```go
package processor

import (
	"errors"
	"fmt"
	"[github.com/h2non/bimg](https://github.com/h2non/bimg)"
)

type ProcessOptions struct {
	TargetWidth    int
	TargetHeight   int
	MaxSizeBytes   int    // เช่น 100 * 1024 (100KB)
	Format         string // "jpeg"
	EnforceWhiteBg bool
}

// ProcessPhoto Resize และบีบอัดรูปภาพด้วย Binary Search
func ProcessPhoto(inputBuffer []byte, opts ProcessOptions) ([]byte, error) {
	img := bimg.NewImage(inputBuffer)

	// 1. ตรวจสอบและแปลงสัดส่วนภาพแบบ Force Resize หรือ Smart Crop
	resized, err := img.Process(bimg.Options{
		Width:   opts.TargetWidth,
		Height:  opts.TargetHeight,
		Crop:    true,
		Gravity: bimg.GravityCentre,
		Type:    bimg.JPEG,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to resize: %w", err)
	}

	// 2. Binary Search หาค่า Quality ที่ดีที่สุดที่ขนาดไฟล์ <= opts.MaxSizeBytes
	lowQuality := 30
	highQuality := 95
	bestQuality := lowQuality
	var finalBuffer []byte

	for lowQuality <= highQuality {
		midQuality := (lowQuality + highQuality) / 2
		compressed, err := bimg.NewImage(resized).Process(bimg.Options{
			Quality: midQuality,
			Type:    bimg.JPEG,
		})
		if err != nil {
			return nil, err
		}

		if len(compressed) <= opts.MaxSizeBytes {
			// เก็บผลลัพธ์ที่ดีที่สุดไว้ แล้วลองขยับ Quality ให้ชัดขึ้นอีก
			bestQuality = midQuality
			finalBuffer = compressed
			lowQuality = midQuality + 1
		} else {
			// ขนาดไฟล์เกินเป้าหมาย ลด Quality ลง
			highQuality = midQuality - 1
		}
	}

	if finalBuffer == nil {
		return nil, errors.New("cannot compress image below target size without extreme degradation")
	}

	return finalBuffer, nil
}

```

---

## 📡 6. REST API Specification

### 1. `POST /api/v1/photos/preset`

แปลงรูปภาพตามพรีเซ็ตที่กำหนด

* **Headers:** `Content-Type: multipart/form-data`
* **Form-Data Parameters:**
* `file`: Binary Data (JPG, PNG, HEIC, WEBP)
* `preset`: `ocsc` | `passport` | `teacher` | `custom`
* `width`: (Optional สำหรับ `custom`)
* `height`: (Optional สำหรับ `custom`)
* `max_kb`: (Optional สำหรับ `custom`)


* **Response (JSON):**
```json
{
  "status": "success",
  "data": {
    "filename": "ocsc_photo_a8f3c1.jpg",
    "width": 200,
    "height": 230,
    "size_kb": 84.6,
    "mime_type": "image/jpeg",
    "download_url": "[https://thai-gov-photos.s3.ap-southeast-1.amazonaws.com/processed/ocsc_photo_a8f3c1.jpg?AWSAccessKeyId=](https://thai-gov-photos.s3.ap-southeast-1.amazonaws.com/processed/ocsc_photo_a8f3c1.jpg?AWSAccessKeyId=)...",
    "expires_in": 3600
  }
}

```



### 2. `POST /api/v1/documents/merge-pdf`

รวมไฟล์ภาพหลายรูปเป็น PDF เดียว พร้อมคุมขนาดไม่เกิน 500 KB

* **Headers:** `Content-Type: multipart/form-data`
* **Form-Data Parameters:**
* `files[]`: Multiple Images / Scanned PDF
* `target_max_kb`: `500`


* **Response (JSON):**
```json
{
  "status": "success",
  "data": {
    "filename": "id_card_merged.pdf",
    "total_pages": 2,
    "size_kb": 412.3,
    "download_url": "[https://thai-gov-photos.s3.ap-southeast-1.amazonaws.com/processed/id_card_merged.pdf](https://thai-gov-photos.s3.ap-southeast-1.amazonaws.com/processed/id_card_merged.pdf)?...",
    "expires_in": 3600
  }
}

```



---

## 🖥️ 7. Frontend UI / UX Features (Next.js 14)

1. **Preset-driven User Flow:**
* ผู้ใช้กดเลือกหน่วยงานที่ต้องการยื่นเอกสาร (ระบบจะตั้งค่ากว้าง x ยาว, น้ำหนักไฟล์ และเงื่อนไขสีฉากหลังให้อัตโนมัติ)


2. **Client-Side Image Manipulation (Interactive Cropper):**
* ใช้ `react-image-crop` เพื่อให้ผู้ใช้หมุน (Rotate 90°), พลิกภาพ (Flip), และลากกรอบครอบตัดภาพเฉพาะศีรษะถึงหน้าอก โดยระบบจะบังคับ Aspect Ratio (เช่น 4:5 สำหรับ ก.พ.) เพื่อป้องกันภาพผิดสัดส่วน


3. **Instant Validation Engine:**
* ตรวจสอบสกุลไฟล์และขนาดภาพตั้งแต่หน้าบ้าน (Client-side validation) แจ้งเตือนทันทีหากไฟล์เสียหาย


4. **Before / After Comparison Slider:**
* แสดงขนาดไฟล์เดิมเทียบกับขนาดไฟล์ใหม่ (เช่น `3.4 MB` $\rightarrow$ `89 KB`) พร้อมปุ่มดาวน์โหลดไฟล์ความละเอียดสูงสุด



---

## 🏗️ 8. Infrastructure as Code (Terraform)

#### `iac/main.tf`

```hcl
terraform {
  required_version = ">= 1.5.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = var.aws_region
}

# S3 Bucket สำหรับเก็บรูปภาพและเอกสารชั่วคราว
resource "aws_s3_bucket" "photo_storage" {
  bucket_prefix = "thai-gov-photos-"
  force_destroy = true
}

# S3 Lifecycle Rule: ลบไฟล์ทิ้งถาวรอัตโนมัติหลัง 24 ชั่วโมง
resource "aws_s3_bucket_lifecycle_configuration" "photo_lifecycle" {
  bucket = aws_s3_bucket.photo_storage.id

  rule {
    id     = "auto-delete-24h-temp-files"
    status = "Enabled"
    expiration {
      days = 1
    }
  }
}

# AWS ECR Repositories
resource "aws_ecr_repository" "backend_repo" {
  name                 = "thai-gov-processor-backend"
  image_tag_mutability = "MUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }
}

resource "aws_ecr_repository" "frontend_repo" {
  name                 = "thai-gov-processor-frontend"
  image_tag_mutability = "MUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }
}

```

#### `iac/security_group.tf`

```hcl
resource "aws_security_group" "k3s_sg" {
  name        = "k3s-cluster-sg"
  description = "Security group for single-node K3s cluster"

  ingress {
    description = "SSH access"
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    description = "K3s API"
    from_port   = 6443
    to_port     = 6443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    description = "HTTP Inbound"
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    description = "HTTPS Inbound"
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

```

#### `iac/ec2_k3s.tf`

```hcl
resource "aws_iam_role" "k3s_instance_role" {
  name = "k3s-instance-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action    = "sts:AssumeRole"
      Effect    = "Allow"
      Principal = { Service = "ec2.amazonaws.com" }
    }]
  })
}

resource "aws_iam_role_policy_attachment" "ecr_read" {
  role       = aws_iam_role.k3s_instance_role.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryPowerUser"
}

resource "aws_iam_role_policy" "s3_access" {
  name = "k3s-s3-access"
  role = aws_iam_role.k3s_instance_role.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Action = [
        "s3:PutObject",
        "s3:GetObject",
        "s3:DeleteObject"
      ]
      Resource = "${aws_s3_bucket.photo_storage.arn}/*"
    }]
  })
}

resource "aws_iam_instance_profile" "k3s_profile" {
  name = "k3s-instance-profile"
  role = aws_iam_role.k3s_instance_role.name
}

resource "aws_instance" "k3s_node" {
  ami                  = "ami-0b940e4f1a2380590" # Ubuntu 24.04 LTS (ap-southeast-1)
  instance_type        = "t3a.large"             # 2 vCPU, 8GB RAM
  key_name             = var.key_pair_name
  vpc_security_group_ids = [aws_security_group.k3s_sg.id]
  iam_instance_profile = aws_iam_instance_profile.k3s_profile.name

  root_block_device {
    volume_size = 35 # GB
    volume_type = "gp3"
  }

  user_data = <<-EOF
              #!/bin/bash
              apt-get update -y
              apt-get install -y curl snapd

              PUBLIC_IP=$(curl -s [http://169.254.169.254/latest/meta-data/public-ipv4](http://169.254.169.254/latest/meta-data/public-ipv4))
              
              # ติดตั้ง K3s 
              curl -sfL [https://get.k3s.io](https://get.k3s.io) | INSTALL_K3S_EXEC="--tls-san $PUBLIC_IP --write-kubeconfig-mode 644" sh -
              
              # ติดตั้ง Helm
              snap install helm --classic
              snap install aws-cli --classic
              EOF

  tags = {
    Name = "k3s-thai-gov-processor"
  }
}

```

---

## ☸️ 9. Kubernetes & Ingress Manifests

#### `k8s/base/ingress.yaml`

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: app-ingress
  namespace: production
  annotations:
    traefik.ingress.kubernetes.io/router.entrypoints: web
spec:
  rules:
  - host: photo-gov.your-ip.nip.io
    http:
      paths:
      - path: /api
        pathType: Prefix
        backend:
          service:
            name: backend-service
            port:
              number: 8080
      - path: /
        pathType: Prefix
        backend:
          service:
            name: frontend-service
            port:
              number: 3000

```

---

## 🔄 10. CI/CD Pipeline Configuration (`Jenkinsfile`)

ใช้ **Kubernetes Plugin** บน Jenkins เพื่อ Spawn Dynamic Pod Agent (Golang, Kaniko, Trivy) ขึ้นมารันงานเฉพาะตอน Build แล้วลบทิ้งเมื่อเสร็จสิ้น:

```groovy
pipeline {
  agent {
    kubernetes {
      yaml '''
apiVersion: v1
kind: Pod
metadata:
  labels:
    role: jenkins-agent
spec:
  serviceAccountName: jenkins-admin
  containers:
  - name: golang
    image: golang:1.22-alpine
    command: ['cat']
    tty: true
  - name: kaniko
    image: gcr.io/kaniko-project/executor:debug
    command: ['cat']
    tty: true
  - name: trivy
    image: aquasec/trivy:latest
    command: ['cat']
    tty: true
'''
    }
  }

  environment {
    AWS_REGION    = 'ap-southeast-1'
    ECR_REGISTRY  = 'xxxxxxxxxxxx.dkr.ecr.ap-southeast-1.amazonaws.com'
    IMAGE_BACKEND = "${ECR_REGISTRY}/thai-gov-processor-backend"
    TAG           = "${BUILD_NUMBER}-${GIT_COMMIT.take(7)}"
  }

  stages {
    stage('Unit Test & Benchmark') {
      steps {
        container('golang') {
          dir('backend') {
            sh 'go test -v -race -cover ./...'
          }
        }
      }
    }

    stage('Security Scan (SAST)') {
      steps {
        container('trivy') {
          sh 'trivy fs --severity HIGH,CRITICAL --exit-code 0 backend/'
        }
      }
    }

    stage('Build & Push to ECR via Kaniko') {
      steps {
        container('kaniko') {
          sh """
          /kaniko/executor \
            --context=dir://./backend \
            --dockerfile=backend/Dockerfile \
            --destination=${IMAGE_BACKEND}:${TAG} \
            --destination=${IMAGE_BACKEND}:latest
          """
        }
      }
    }

    stage('Scan Container Image (CVEs)') {
      steps {
        container('trivy') {
          sh "trivy image --severity CRITICAL --exit-code 0 ${IMAGE_BACKEND}:${TAG}"
        }
      }
    }

    stage('Deploy to K3s Production') {
      steps {
        sh "kubectl set image deployment/backend-deployment processor=${IMAGE_BACKEND}:${TAG} -n production"
        sh "kubectl rollout status deployment/backend-deployment -n production --timeout=120s"
      }
    }
  }

  post {
    always {
      cleanWs()
    }
    failure {
      echo "Deployment failed! Review build logs immediately."
    }
  }
}

```

---

## 💰 11. Cost Optimization & Budgeting

| Service / Resource | รายละเอียดการคำนวณ | ค่าใช้จ่าย (เปิดทิ้งไว้ 1 วัน) | ค่าใช้จ่าย (เปิดทิ้งไว้ 1 เดือน) |
| --- | --- | --- | --- |
| **AWS EC2 (t3a.large)** | 2 vCPU, 8GB RAM ($0.0752/ชม.) | ~60 บาท / วัน | ~1,850 บาท / เดือน |
| **Amazon S3** | Storage < 1GB + S3 Lifecycle 24 ชม. | ~0.00 บาท | ~0.50 บาท / เดือน |
| **Amazon ECR** | เก็บ Docker Image < 200MB | ~0.00 บาท | ~0.60 บาท / เดือน |
| **EKS Control Plane** | **ไม่ได้ใช้** (หันมาใช้ K3s บน EC2 แทน) | **ประหยัดได้ 100%** | **ประหยัดไป $73 (~2,600 บาท)** |
| **รวมโดยประมาณ** |  | **~60 บาท / วัน** | **~1,850 บาท / เดือน** |

> 💡 **Best Practice ในการประหยัดงบ:**
> ในการนำเสนอผลงาน ให้รัน `terraform apply` ก่อนการสัมภาษณ์หรืออัดวิดีโอสาธิต จากนั้นเมื่อเสร็จสิ้นให้สั่ง `terraform destroy` ทันที จะเสียค่าใช้จ่ายจริงเพียง **ไม่กี่บาทต่อการทดสอบ**

---
