
# 🧠 Quizzes Backend API

Backend service yang kuat dan skalabel untuk aplikasi Kuis Online. Dibangun menggunakan **Golang**, **Fiber**, dan **PostgreSQL**.

Project ini mendukung manajemen kuis hierarkis (Topik > Kuis > Soal), sistem autentikasi berbasis peran (RBAC), fitur sosial (Friends), dan gamifikasi (Leaderboard).

## 🚀 Tech Stack

-   **Language:** Golang (1.20+)
-   **Framework:** [Fiber v2](https://gofiber.io/) (Express-inspired, high performance)
-   **Database:** PostgreSQL (via [Neon Serverless](https://neon.tech/))
-   **ORM:** [GORM](https://gorm.io/)
-   **Auth:** JWT (JSON Web Token)
-   **Password Hashing:** Bcrypt

## ✨ Fitur Utama

### 🔐 Authentication & RBAC
-   **Multi-Auth:** Login terpisah untuk User (Peserta) dan Admin (Pengelola).
-   **Dynamic RBAC:** Role dinamis disimpan di database (`supervisor`, `admin`, `pengajar`).
-   **Middleware:** Proteksi route berdasarkan token JWT dan Role.

### 📚 Manajemen Kuis
-   **Hierarki:** Mata Kuliah (Topic) -> Kuis (Quiz) -> Soal (Question).
-   **Support:** Pilihan Ganda dengan opsi jawaban dinamis (Array).
-   **Randomizer:** Soal diacak secara otomatis saat diambil oleh user.

### 🎮 Gameplay & Gamification
-   **History:** Menyimpan skor, jawaban user (snapshot), dan total soal.
-   **Leaderboard:** Peringkat user berdasarkan total poin per Topik.
-   **Review:** User bisa melihat detail jawaban benar/salah setelah mengerjakan.

### 🤝 Social Feature
-   **Friendship:** Tambah teman, hapus teman, dan lihat daftar teman.

## 📂 Struktur Project

Project ini menggunakan **Modular Architecture** agar mudah dikembangkan (Scalable).

```text
quizzes-backend/
├── config/             # Konfigurasi Database & Auto Seeder
├── controllers/        # Logic bisnis (Auth, Quiz, Social, dll)
├── middleware/         # JWT Auth & Role Check
├── models/             # Definisi Struct Database (GORM)
├── routes/             # Grouping URL API
├── utils/              # Helper Response Standard
├── .env                # Environment Variables
└── main.go             # Entry Point
````

## 🛠️ Instalasi & Cara Menjalankan

### 1\. Clone Repository

```bash
git clone [https://github.com/ROFL1ST/quizzes-backend.git](https://github.com/ROFL1ST/quizzes-backend.git)
cd quizzes-backend
```

### 2\. Install Dependencies

```bash
go mod tidy
```

### 3\. Konfigurasi Environment

Buat file `.env` di root folder dan isi dengan konfigurasi berikut:

```env
PORT=8000
JWT_SECRET=rahasia_super_aman_sekali

# Connection String (Contoh menggunakan Neon Postgres)
DATABASE_URL="postgresql://user:password@host-name.aws.neon.tech/dbname?sslmode=require"
```

### 4\. Jalankan Aplikasi

```bash
go run main.go
```

Saat pertama kali dijalankan, aplikasi akan melakukan:

1.  **Auto Migration:** Membuat semua tabel database secara otomatis.
2.  **Auto Seeding:** Mengisi data Role dasar dan akun Super Admin.

## ⚡ Default Credentials (Seeding)

Jika database kosong, sistem akan otomatis membuat akun berikut:

| Role | Username | Password |
| :--- | :--- | :--- |
| **Supervisor** | `superadmin` | `123456` |

> Gunakan akun ini untuk login di endpoint `/api/admin/login` dan membuat admin atau pengajar lain.

## 📡 API Endpoints Overview

### Auth

| Method | Endpoint | Deskripsi |
| :--- | :--- | :--- |
| POST | `/api/register` | Register User Baru |
| POST | `/api/login` | Login User |
| POST | `/api/admin/login` | Login Admin/Pengajar |

### Gameplay (User)

| Method | Endpoint | Deskripsi |
| :--- | :--- | :--- |
| GET | `/api/topics` | Lihat semua Mata Kuliah |
| GET | `/api/topics/:slug/quizzes` | Lihat daftar Kuis di Topik tertentu |
| GET | `/api/quizzes/:id/questions` | Ambil soal acak untuk dikerjakan |
| POST | `/api/history` | Submit jawaban & simpan skor |

### Social & Leaderboard

| Method | Endpoint | Deskripsi |
| :--- | :--- | :--- |
| GET | `/api/leaderboard/:slug` | Top 10 User berdasarkan Topik |
| POST | `/api/friends/add` | Tambah teman by Username |
| GET | `/api/friends` | Lihat daftar teman |

### Admin Management (Protected)

| Method | Endpoint | Deskripsi |
| :--- | :--- | :--- |
| POST | `/api/admin/topics` | Buat Mata Kuliah baru |
| POST | `/api/admin/quizzes` | Buat Judul Kuis baru |
| POST | `/api/admin/questions` | Input butir soal |
| POST | `/api/admin/register` | Daftarkan Admin baru (perlu role\_id) |

## 🧪 Testing API (Postman)

Tersedia file koleksi Postman untuk pengujian mudah.
File: `quiz_collection.json` (Silakan import ke Postman).

-----

Made with ❤️ by [ROFL1ST](https://github.com/ROFL1ST)

