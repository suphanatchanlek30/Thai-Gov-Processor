# REST API

| Method | Path | ใช้ทำอะไร |
| --- | --- | --- |
| `GET` | `/healthz` | liveness/readiness probe |
| `GET` | `/api/v1/selftest` | แปลงรูปตัวอย่างที่ฝังไว้ในแอป แล้วเช็คว่าได้ขนาดตรง preset (ใช้ใน smoke test) |
| `GET` | `/api/v1/presets` | รายการ preset ทั้งหมด |
| `POST` | `/api/v1/photos/preset` | แปลงรูปตาม preset |
| `POST` | `/api/v1/documents/merge-pdf` | รวมหลายไฟล์ (รูปภาพ และ/หรือ PDF หลายไฟล์) เป็น PDF เดียว |

## `POST /api/v1/photos/preset` (multipart/form-data)

| field ที่ส่งเข้า | ค่า |
| --- | --- |
| `file` | JPG, PNG, HEIC, WEBP (สูงสุด 15 MB) |
| `preset` | `ocsc` \| `passport` \| `teacher` \| `custom` |
| `width`, `height`, `max_kb` | ใช้เฉพาะ `custom` |

| field ที่ตอบกลับ | ความหมาย |
| --- | --- |
| `filename` | ชื่อไฟล์ผลลัพธ์ |
| `width`, `height` | ขนาดภาพจริงหลังแปลง |
| `size_kb` | ขนาดไฟล์หลังบีบ |
| `quality` | ค่า JPEG quality ที่หาได้ |
| `download_url` | presigned URL ของ S3 |
| `expires_in` | อายุลิงก์เป็นวินาที (3600) |

## `POST /api/v1/documents/merge-pdf` (multipart/form-data)

| field ที่ส่งเข้า | ค่า |
| --- | --- |
| `files[]` | รูปภาพ, ไฟล์ PDF (หลายไฟล์) หรือผสมกันก็ได้ — รวมออกมาเป็น PDF เดียวตามลำดับที่ส่งเข้ามา |
| `target_max_kb` | ค่าเริ่มต้น `500` |

ถ้าไฟล์ใน `files[]` เป็น PDF ที่มีหลายหน้าอยู่แล้ว ทุกหน้าจะถูกดึงมารวมต่อกันตามลำดับไฟล์ ไม่ใช่แค่รวมหน้าแรก

ตอบกลับ `filename`, `total_pages`, `size_kb`, `download_url` และ `expires_in`

**Error:** `400` ไฟล์ไม่รองรับ · `413` ไฟล์ใหญ่เกิน · `422` บีบให้ต่ำกว่าเกณฑ์ไม่ได้โดยไม่เสียคุณภาพ
