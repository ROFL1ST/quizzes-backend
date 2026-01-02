# 🧠 Quizzes Backend API

Backend service yang kuat dan skalabel untuk aplikasi Kuis Online. Dibangun menggunakan **Golang**, **Fiber**, dan **PostgreSQL**.

Project ini mendukung manajemen kuis hierarkis (Topik > Kuis > Soal), sistem autentikasi berbasis peran (RBAC), fitur sosial (Friends), dan gamifikasi (Leaderboard).

## 🚀 Tech Stack

- **Language:** Golang (1.20+)
- **Framework:** [Fiber v2](https://gofiber.io/) (Express-inspired, high performance)
- **Database:** PostgreSQL (via [Neon Serverless](https://neon.tech/))
- **ORM:** [GORM](https://gorm.io/)
- **Auth:** JWT (JSON Web Token)
- **Password Hashing:** Bcrypt

## ✨ Fitur Utama

### 🔐 Authentication & RBAC

- **Multi-Auth:** Login terpisah untuk User (Peserta) dan Admin (Pengelola).
- **Dynamic RBAC:** Role dinamis disimpan di database (`supervisor`, `admin`, `pengajar`).
- **Middleware:** Proteksi route berdasarkan token JWT dan Role.

### 📚 Manajemen Kuis

- **Hierarki:** Mata Kuliah (Topic) -> Kuis (Quiz) -> Soal (Question).
- **Support:** Pilihan Ganda dengan opsi jawaban dinamis (Array).
- **Randomizer:** Soal diacak secara otomatis saat diambil oleh user.

### 🎮 Gameplay & Gamification

- **History:** Menyimpan skor, jawaban user (snapshot), dan total soal.
- **Leaderboard:** Peringkat user berdasarkan total poin per Topik.
- **Review:** User bisa melihat detail jawaban benar/salah setelah mengerjakan.

### 🤝 Social Feature

- **Friendship:** Tambah teman, hapus teman, dan lihat daftar teman.

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
```

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

| Role           | Username     | Password |
| :------------- | :----------- | :------- |
| **Supervisor** | `superadmin` | `123456` |

> Gunakan akun ini untuk login di endpoint `/api/admin/login` dan membuat admin atau pengajar lain.

## 📡 API Endpoints Overview

### 👤 User Routes (Peserta)

#### Auth & Profile

| Method | Endpoint                  | Deskripsi                        |
| :----- | :------------------------ | :------------------------------- |
| POST   | `/api/register`           | Register User Baru               |
| POST   | `/api/login`              | Login User                       |
| POST   | `/api/verify-email`       | Verifikasi Email                 |
| POST   | `/api/forgot-password`    | Request Reset Password           |
| POST   | `/api/reset-password`     | Reset Password Baru              |
| GET    | `/api/auth/me`            | Cek Session & Data User          |
| GET    | `/api/users/me`           | Lihat Profil & Statistik Sendiri |
| PUT    | `/api/users/me`           | Update Profil (Nama/Pass)        |
| GET    | `/api/users/:username`    | Lihat Profil User Lain           |
| GET    | `/api/users/search`       | Cari User                        |
| GET    | `/api/users/achievements` | Lihat Pencapaian Saya            |

#### Gameplay & Kuis

| Method | Endpoint                      | Deskripsi                           |
| :----- | :---------------------------- | :---------------------------------- |
| GET    | `/api/topics`                 | Lihat semua Mata Kuliah             |
| GET    | `/api/topics/:slug/quizzes`   | Lihat daftar Kuis di Topik tertentu |
| GET    | `/api/quizzes/:id/questions`  | Ambil soal acak untuk dikerjakan    |
| POST   | `/api/history`                | Submit jawaban & simpan skor        |
| GET    | `/api/history`                | Lihat history kuis saya             |
| GET    | `/api/history/:id`            | Detail history tertentu             |
| GET    | `/api/quizzes/remedial/start` | Mulai sesi remedial (soal salah)    |

#### Social & Features

| Method | Endpoint                 | Deskripsi                     |
| :----- | :----------------------- | :---------------------------- |
| GET    | `/api/leaderboard/:slug` | Top 10 User berdasarkan Topik |
| GET    | `/api/friends`           | Lihat daftar teman            |
| POST   | `/api/friends/request`   | Kirim pertemanan              |
| POST   | `/api/friends/confirm`   | Terima pertemanan             |
| DELETE | `/api/friends/:id`       | Hapus teman                   |
| GET    | `/api/feed`              | Activity Feed teman           |
| GET    | `/api/notifications`     | Lihat Notifikasi              |

#### Challenges (PvP)

| Method | Endpoint                     | Deskripsi            |
| :----- | :--------------------------- | :------------------- |
| POST   | `/api/challenges`            | Buat Challenge Baru  |
| GET    | `/api/challenges`            | List Challenge aktif |
| POST   | `/api/challenges/:id/accept` | Terima Tantangan     |
| POST   | `/api/challenges/:id/start`  | Mulai Game Realtime  |

#### Shop & Inventory

| Method | Endpoint              | Deskripsi            |
| :----- | :-------------------- | :------------------- |
| GET    | `/api/shop/items`     | Lihat Item di Toko   |
| POST   | `/api/shop/buy`       | Beli Item            |
| GET    | `/api/shop/inventory` | Lihat Inventory Saya |
| POST   | `/api/shop/equip`     | Pakai Item           |

#### User Routes

| Feature         | Method | Endpoint                     | Description                              |
| --------------- | ------ | ---------------------------- | ---------------------------------------- |
| **Auth**        | POST   | `/api/register`              | Register new user                        |
|                 | POST   | `/api/login`                 | Login user                               |
|                 | GET    | `/api/users/me`              | Get own profile                          |
| **Topics**      | GET    | `/api/topics`                | Get all topics                           |
|                 | GET    | `/api/topics/:slug/quizzes`  | Get quizzes by topic                     |
| **Quizzes**     | GET    | `/api/quizzes/:id/questions` | Play quiz (get questions)                |
|                 | POST   | `/api/history`               | Submit quiz result                       |
| **Reviews**     | POST   | `/api/quizzes/:id/reviews`   | Rate & review a quiz                     |
|                 | GET    | `/api/quizzes/:id/reviews`   | Get reviews for a quiz                   |
| **Classroom**   | GET    | `/api/classrooms`            | Get my classrooms (teaching/joined)      |
|                 | POST   | `/api/classrooms/join`       | Join class by code                       |
|                 | GET    | `/api/classrooms/:id`        | Get class details & assignments          |
| **Survival**    | POST   | `/api/survival/start`        | Start survival mode                      |
|                 | POST   | `/api/survival/answer`       | Answer survival question                 |
| **Social**      | GET    | `/api/friends`               | Get friend list                          |
|                 | POST   | `/api/friends/request`       | Send friend request                      |
| **Leaderboard** | GET    | `/api/leaderboard/global`    | **[NEW]** Global Leaderboard (Top 20 XP) |
|                 | GET    | `/api/leaderboard/:slug`     | Topic Leaderboard                        |
| **Reports**     | POST   | `/api/reports`               | Report a bug/user/question               |

---

### 🛡️ Admin & Pengajar Routes

#### Auth & Dashboard

| Method | Endpoint               | Deskripsi            |
| :----- | :--------------------- | :------------------- |
| POST   | `/api/admin/login`     | Login Admin/Pengajar |
| POST   | `/api/admin/register`  | Register Admin Baru  |
| GET    | `/api/admin/analytics` | Dashboard Analytics  |

#### Management

| Method        | Endpoint                     | Deskripsi                                |
| :------------ | :--------------------------- | :--------------------------------------- | -------------------- |
| GET           | `/api/admin/topics`          | List Topik Management                    |
| POST          | `/api/admin/topics`          | Buat Topik Baru                          |
| GET           | `/api/admin/quizzes`         | List Quiz Management                     |
| POST          | `/api/admin/quizzes`         | Buat Kuis Baru                           |
| GET           | `/api/admin/questions`       | List Bank Soal                           |
| POST          | `/api/admin/questions`       | Input Soal Manual                        |
| POST          | `/api/admin/questions/bulk`  | Upload Soal Bulk                         |
| GET           | `/api/admin/users`           | Manage Users                             |
| GET           | `/api/admin/roles`           | Manage Roles                             |
| GET           | `/api/admin/shop/items`      | Manage Shop Items                        |
| PUT           | `/api/admin/users/:id/ban`   | **[NEW]** Ban User                       |
| PUT           | `/api/admin/users/:id/unban` | **[NEW]** Unban User                     |
| POST          | `/api/admin/broadcast`       | **[NEW]** Create System Announcement     |
| POST          | `/api/admin/create-admin`    | **[NEW]** Create Admin (Superadmin Only) |
| **Reports**   | GET                          | `/api/admin/reports`                     | View all reports     |
|               | PUT                          | `/api/admin/reports/:id`                 | Resolve report       |
| **Classroom** | POST                         | `/api/classrooms`                        | Create new classroom |
|               | POST                         | `/api/classrooms/:id/assignments`        | Assign quiz to class |

## 🧪 Testing API (Postman)

Tersedia file koleksi Postman untuk pengujian mudah.
File: `quiz_collection.json` (Silakan import ke Postman).

---

Made with ❤️ by [ROFL1ST](https://github.com/ROFL1ST)
