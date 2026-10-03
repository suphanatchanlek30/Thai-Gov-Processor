# Architecture

สองส่วนหลักของระบบคือ **Application layer** (Next.js + Go + S3) กับ **DevOps layer** (Terraform → K3s → Jenkins → Argo CD) แยกกันชัดเจน — layer แรกทำหน้าที่ประมวลผลไฟล์ ส่วนอีก layer ทำหน้าที่เอาโค้ดขึ้น production โดยอัตโนมัติ

ภาพรวมทั้งระบบอยู่ในไฟล์ [`Architecture diagram/Project CICD FUll.drawio-2.svg`](../Architecture%20diagram/Project%20CICD%20FUll.drawio-2.svg) ส่วนด้านล่างนี้คือ diagram แยกตาม tier ที่ตรงกับ implementation ปัจจุบัน (BuildKit แทน Kaniko, Argo CD แทน `kubectl rollout`, ไม่เปิด SSH, frontend/backend อยู่ namespace เดียวกับ Ingress)

## Diagram แยกตาม tier

```mermaid
flowchart TB
%% ================= STYLES =================
    classDef client fill:#0284c7,stroke:#38bdf8,stroke-width:2px,color:#ffffff
    classDef cicd fill:#1e293b,stroke:#818cf8,stroke-width:2px,color:#ffffff
    classDef gate fill:#7f1d1d,stroke:#f87171,stroke-width:2px,color:#ffffff
    classDef cd fill:#14532d,stroke:#4ade80,stroke-width:2px,color:#ffffff
    classDef k8s fill:#0f172a,stroke:#38bdf8,stroke-width:2px,color:#ffffff
    classDef storage fill:#7c2d12,stroke:#fb923c,stroke-width:2px,color:#ffffff
    classDef security fill:#581c87,stroke:#c084fc,stroke-width:2px,color:#ffffff

%% ================= 1. CLIENT & UI FLOW =================
    subgraph CLIENT_TIER ["💻 Client & Frontend Tier"]
        User(["👤 Job Applicant / End User"]):::client
        subgraph NEXTJS_APP ["Next.js 15 Web Application (UI)"]
            UI_Home["🏠 Drag & Drop File Upload"]:::client
            UI_Presets["🎯 Presets Selector<br/>• OCSC 200x230 px, ≤ 100 KB, JPG<br/>• Passport 2x2 inch, white background<br/>• PDF Merger ≤ 500 KB"]:::client
            UI_Crop["✂️ Crop & Rotate (locked ratio)"]:::client
            UI_Preview["👁️ Before / After Preview & Download"]:::client
        end
    end

%% ================= 2. CI/CD TIER =================
    subgraph DEVOPS_TIER ["⚙️ CI/CD Automation Tier"]
        Dev["👨‍💻 Developer"]:::cicd
        GitHub["🐙 GitHub Monorepo<br/>/frontend /backend /k8s /iac"]:::cicd
        subgraph JENKINS_NS ["Jenkins (namespace: jenkins)"]
            Trigger["⚡ Webhook Receiver<br/>HMAC verified"]:::cicd
            subgraph DYNAMIC_AGENT ["Ephemeral Agent Pod (spawn on demand)"]
                Stage_Lint["1. Lint & Unit Test<br/>golangci-lint · go test -race · eslint"]:::cicd
                Stage_Build["2. Build Image<br/>BuildKit rootless → .tar"]:::cicd
                Stage_Scan{{"3. Security Gate<br/>Trivy: CRITICAL = fail"}}:::gate
                Stage_Push["4. Push to ECR<br/>crane · tag = git SHA"]:::cicd
                Stage_Tag["5. Update manifest tag<br/>commit with skip-ci marker"]:::cicd
            end
        end
        subgraph ARGO_NS ["Argo CD (namespace: argocd)"]
            Argo["🔄 GitOps Sync<br/>auto-prune · self-heal"]:::cd
            Smoke{{"🧪 PostSync Smoke Test"}}:::gate
        end
    end

%% ================= 3. AWS CLOUD & RUNTIME =================
    subgraph AWS_CLOUD ["☁️ AWS Cloud (ap-southeast-1 Singapore)"]
        subgraph IAC_BOX ["🏗️ Infrastructure as Code"]
            Terraform["🟣 Terraform<br/>VPC · EC2 · S3 · ECR · IAM<br/>state: S3 + lockfile"]:::security
        end
        ECR[("📦 Amazon ECR<br/>thai-gov-processor-backend · -frontend<br/>IMMUTABLE · scan on push")]:::storage
        subgraph EC2_INSTANCE ["🖥️ EC2 m7i-flex.large (2 vCPU, 8 GB) · IMDSv2"]
            SG["🛡️ Security Group<br/>80 / 443 public · 6443 admin IP only<br/>no SSH (use SSM)"]:::security
            IAM["🔑 IAM Instance Profile<br/>ECR push/pull 2 repos · S3 one bucket"]:::security
            subgraph K3S_CLUSTER ["☸️ K3s Single-Node Cluster"]
                Traefik["🌐 Traefik Ingress :443<br/>TLS via cert-manager + Let's Encrypt"]:::k8s
                subgraph NS_PROD ["Namespaces: thai-gov (prod) · thai-gov-staging"]
                    FE_Svc["🔀 frontend Service :3000"]:::k8s
                    UIPod["📱 Next.js Pods × 2"]:::k8s
                    BE_Svc["🔀 backend Service :8080"]:::k8s
                    subgraph PODS_HPA ["HPA: 2–4 pods at CPU 70%"]
                        Pod1["⚡ Go Worker Pod 1<br/>bimg · pdfcpu"]:::k8s
                        Pod2["⚡ Go Worker Pod 2"]:::k8s
                    end
                end
            end
        end
        subgraph STORAGE_TIER ["🪣 Storage & Data Privacy"]
            S3[("Amazon S3<br/>processed/ (ผลลัพธ์เท่านั้น)")]:::storage
            S3_Lifecycle["⏳ Lifecycle: expire after 1 day<br/>ต้นฉบับไม่ถูกเก็บลง S3 (ประมวลผลในหน่วยความจำ)<br/>(PDPA)"]:::security
            S3 --- S3_Lifecycle
        end
    end

%% ================= PIPELINE CONNECTIONS =================
    Dev -->|"1 · git push / PR"| GitHub
    GitHub -->|"2 · webhook"| Trigger
    Trigger -->|"3 · spawn pod"| Stage_Lint
    Stage_Lint --> Stage_Build --> Stage_Scan
    Stage_Scan -->|"pass (branch dev เท่านั้น)"| Stage_Push
    Stage_Push -->|"push image"| ECR
    Stage_Push --> Stage_Tag
    Stage_Tag -->|"4 · commit new tag"| GitHub
    Argo -->|"5 · watch repo"| GitHub
    Argo -->|"6 · sync"| NS_PROD
    Argo --> Smoke
    Smoke -.->|"health + selftest"| BE_Svc
    ECR -.->|"pull image"| NS_PROD

%% ================= USER APPLICATION CONNECTIONS =================
    User -->|"open website"| UI_Home
    UI_Home --> UI_Presets --> UI_Crop
    UI_Crop -->|"POST /api/v1/photos/preset"| Traefik
    Traefik -->|"app.domain /"| FE_Svc
    FE_Svc --> UIPod
    Traefik -->|"app.domain /api"| BE_Svc
    Traefik -->|"jenkins.domain /github-webhook/ เท่านั้น"| Trigger
    BE_Svc --> Pod1
    BE_Svc --> Pod2
    Pod1 -->|"PutObject ผลลัพธ์"| S3
    Pod2 -->|"PutObject ผลลัพธ์"| S3
    S3 -.->|"presigned URL (1 h)"| UI_Preview
    UI_Preview -->|"download processed file"| User

%% ================= IAC PROVISIONING =================
    Terraform -.->|"provision"| EC2_INSTANCE
    Terraform -.->|"provision"| S3
    Terraform -.->|"provision"| ECR
```

## Runtime บน AWS

Region หลักคือ `ap-southeast-1` (Singapore) ทั้งระบบรันอยู่บน EC2 เครื่องเดียวที่ติดตั้ง K3s — ตั้งใจทำแบบ production-like เพื่อเรียนรู้ ไม่ใช่ high-availability production เต็มรูปแบบ

```mermaid
flowchart TB
    user(["👤 ผู้ใช้"])
    gh["🐙 GitHub"]
    le["🔐 Let's Encrypt"]

    subgraph AWS["☁️ AWS · ap-southeast-1 (Singapore)"]
        subgraph VPC["VPC 10.0.0.0/16 → public subnet 10.0.1.0/24"]
            subgraph EC2["🖥️ EC2 m7i-flex.large · 2 vCPU / 8 GB · Ubuntu 24.04 · IMDSv2 · Elastic IP<br/>SG: 80, 443 ทุกที่ · 6443 เฉพาะ IP ตัวเอง · ไม่เปิด 22 (ใช้ SSM)"]
                subgraph K3S["☸️ K3s single-node"]
                    subgraph KS["ns: kube-system"]
                        traefik["Traefik ingress<br/>:80 → :443"]
                        cm["cert-manager"]
                        ms["metrics-server"]
                    end
                    subgraph PROD["ns: thai-gov (prod) · thai-gov-staging"]
                        fe["Frontend<br/>Next.js × 2"]
                        be["Backend<br/>Go + Gin + libvips<br/>HPA 2–4"]
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
        ecr[("📦 ECR<br/>thai-gov-processor-backend<br/>thai-gov-processor-frontend")]
        s3[("🪣 S3 ไฟล์ผลลัพธ์<br/>processed/<br/>หมดอายุ 1 วัน")]
        state[("🪣 S3 Terraform state")]
        iam["🔑 IAM role<br/>least privilege"]
    end

    user -->|HTTPS| traefik
    traefik -->|"app.example.com /"| fe
    traefik -->|"app.example.com /api"| be
    traefik -->|"jenkins.IP.sslip.io /github-webhook/"| jc
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

### เส้นทางของผู้ใช้ 1 คน

```mermaid
sequenceDiagram
    autonumber
    actor U as ผู้ใช้
    participant T as Traefik
    participant FE as Frontend (Next.js)
    participant BE as Backend (Go + Gin)
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

## หลักการแบ่งหน้าที่

- **Jenkins** ทำ CI เท่านั้น — ตรวจโค้ด, test, build, security scan แล้วหยุดแค่ commit เปลี่ยน image tag กลับเข้า Git ไม่มีสิทธิ์ deploy เข้า production โดยตรง
- **Argo CD** ทำ CD — อ่าน desired state จาก Git แล้ว sync เข้า K3s เอง (pull-based, ไม่ใช่ push-based)
- **Terraform** รับผิดชอบ infrastructure เท่านั้น (VPC, EC2, S3, ECR, IAM, Budgets) ไม่แตะ application deployment
- **Build ครั้งเดียว** บน `dev` แล้ว promote tag เดิมผ่าน `staging` ไป `main` (ดู [cicd-pipeline.md](cicd-pipeline.md))
- มี environment จริงสองตัวในเครื่องเดียว: `thai-gov` (prod, backend HPA 2–4) และ `thai-gov-staging` (1 replica ต่อ service ไม่มี HPA) `dev` ไม่มี environment ของตัวเอง

รายละเอียดของแต่ละส่วนดูได้ที่ [cicd-pipeline.md](cicd-pipeline.md) และ [infrastructure.md](infrastructure.md)
