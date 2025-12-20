package config

import (
	"fmt"
	"github.com/ROFL1ST/quizzes-backend/models"
)

func SeedExamData() {
	var count int64
	DB.Model(&models.Topic{}).Where("slug = ?", "jarkom").Count(&count)
	if count > 0 {
		fmt.Println("Exam data already seeded. Skipping...")
		return
	}

	fmt.Println("Seeding Exam Data (IMK & JarKom)...")

	topicJarkom := models.Topic{Slug: "jarkom", Title: "Jaringan Komputer", Description: "TCP/IP, HTTP, DNS, dan Layer OSI"}
	DB.Create(&topicJarkom)

	quizJarkom := models.Quiz{TopicID: topicJarkom.ID, Title: "UTS JarKom (Latihan)", Description: "Kumpulan soal UTS Jaringan Komputer"}
	DB.Create(&quizJarkom)

	questionsJarkom := []models.Question{
		{
			QuizID:        quizJarkom.ID,
			QuestionText:  "Fungsi utama lapisan aplikasi dalam model TCP/IP adalah...",
			Options:       []string{"Mengatur pengalamatan IP antar host", "Menyediakan layanan langsung bagi pengguna atau aplikasi", "Mengirimkan data dalam bentuk bit", "Menentukan rute terbaik bagi paket data"},
			CorrectAnswer: "Menyediakan layanan langsung bagi pengguna atau aplikasi",
			Hint:          "Berhubungan langsung dengan user interface.",
		},
		{
			QuizID:        quizJarkom.ID,
			QuestionText:  "Protokol yang digunakan untuk mengirim email antar server adalah...",
			Options:       []string{"POP3", "IMAP", "SMTP", "HTTP"},
			CorrectAnswer: "SMTP",
			Hint:          "Simple Mail Transfer Protocol.",
		},
		{
			QuizID:        quizJarkom.ID,
			QuestionText:  "Port standar untuk protokol HTTPS adalah...",
			Options:       []string{"21", "25", "80", "443"},
			CorrectAnswer: "443",
			Hint:          "HTTP aman.",
		},
		{
			QuizID:        quizJarkom.ID,
			QuestionText:  "Koneksi persistent pada HTTP berarti...",
			Options:       []string{"Setiap request membuka koneksi baru", "Koneksi TCP tetap terbuka untuk beberapa permintaan dan respons", "Server langsung menutup koneksi setelah satu permintaan", "Data dikirim tanpa perlu koneksi"},
			CorrectAnswer: "Koneksi TCP tetap terbuka untuk beberapa permintaan dan respons",
			Hint:          "Agar tidak perlu handshake berulang kali.",
		},
		{
			QuizID:        quizJarkom.ID,
			QuestionText:  "Fungsi utama DNS adalah...",
			Options:       []string{"Mengirimkan data antara client dan server", "Menerjemahkan nama domain menjadi IP", "Mengatur pengiriman file", "Mengelola sesi koneksi web"},
			CorrectAnswer: "Menerjemahkan nama domain menjadi IP",
			Hint:          "Seperti buku telepon internet.",
		},
		{
			QuizID:        quizJarkom.ID,
			QuestionText:  "Komponen DNS yang menyimpan data otoritatif adalah...",
			Options:       []string{"Recursive Resolver", "TLD Server", "Authoritative Name Server", "Root Server"},
			CorrectAnswer: "Authoritative Name Server",
			Hint:          "Sumber kebenaran terakhir.",
		},
		{
			QuizID:        quizJarkom.ID,
			QuestionText:  "Record DNS untuk server email adalah...",
			Options:       []string{"A Record", "AAAA Record", "MX Record", "CNAME Record"},
			CorrectAnswer: "MX Record",
			Hint:          "Mail Exchange.",
		},
		{
			QuizID:        quizJarkom.ID,
			QuestionText:  "Dalam jaringan P2P, komputer berfungsi sebagai...",
			Options:       []string{"Client saja", "Server saja", "Client sekaligus server", "Gateway"},
			CorrectAnswer: "Client sekaligus server",
			Hint:          "Peer to Peer.",
		},
		{
			QuizID:        quizJarkom.ID,
			QuestionText:  "DASH berfungsi untuk...",
			Options:       []string{"Mengirim video dalam satu file besar", "Menyesuaikan kualitas video dengan kecepatan jaringan", "Mengompresi file video", "Menyimpan cache DNS"},
			CorrectAnswer: "Menyesuaikan kualitas video dengan kecepatan jaringan",
			Hint:          "Streaming adaptif.",
		},
		{
			QuizID:        quizJarkom.ID,
			QuestionText:  "Fungsi utama CDN adalah...",
			Options:       []string{"Mengatur distribusi email", "Mengirim konten dari server terdekat", "Mengompresi data", "Menyimpan file lokal"},
			CorrectAnswer: "Mengirim konten dari server terdekat",
			Hint:          "Content Delivery Network.",
		},
		{
			QuizID:        quizJarkom.ID,
			QuestionText:  "Lapisan Transport OSI berfungsi untuk...",
			Options:       []string{"Mengatur alamat IP", "Menentukan rute", "Pengiriman data yang andal dan efisien", "Enkripsi data"},
			CorrectAnswer: "Pengiriman data yang andal dan efisien",
			Hint:          "TCP/UDP ada di sini.",
		},
		{
			QuizID:        quizJarkom.ID,
			QuestionText:  "Protokol connection-oriented dan reliable adalah...",
			Options:       []string{"IP", "UDP", "TCP", "HTTP"},
			CorrectAnswer: "TCP",
			Hint:          "Menjamin urutan data.",
		},
		{
			QuizID:        quizJarkom.ID,
			QuestionText:  "Urutan Three-Way Handshake adalah...",
			Options:       []string{"ACK -> SYN -> SYN-ACK", "SYN -> ACK -> SYN-ACK", "SYN -> SYN-ACK -> ACK", "FIN -> ACK -> FIN-ACK"},
			CorrectAnswer: "SYN -> SYN-ACK -> ACK",
			Hint:          "Ajak kenalan, dibalas, oke.",
		},
		{
			QuizID:        quizJarkom.ID,
			QuestionText:  "Fungsi Flow Control adalah...",
			Options:       []string{"Menentukan rute", "Mengatur kecepatan pengiriman agar tidak membanjiri penerima", "Mencegah kehilangan data", "Mengatur ukuran jendela UDP"},
			CorrectAnswer: "Mengatur kecepatan pengiriman agar tidak membanjiri penerima",
			Hint:          "Agar penerima tidak kewalahan.",
		},
		{
			QuizID:        quizJarkom.ID,
			QuestionText:  "Yang bukan fitur RDT TCP adalah...",
			Options:       []string{"Sequence Number", "Acknowledgment", "Checksum", "Encryption"},
			CorrectAnswer: "Encryption",
			Hint:          "TCP standar tidak mengenkripsi.",
		},
		{
			QuizID:        quizJarkom.ID,
			QuestionText:  "Sebutkan port default untuk layanan SSH!",
			Type:          "short_answer",
			Options:       []string{}, // Kosongkan untuk isian
			CorrectAnswer: "22",
			Hint:          "Angka dua kembar.",
		},
		{
			QuizID:        quizJarkom.ID,
			QuestionText:  "Singkatan dari protokol pengiriman file adalah?",
			Type:          "short_answer",
			Options:       []string{},
			CorrectAnswer: "FTP",
			Hint:          "File Transfer Protocol.",
		},

		// --- TIPE: BENAR / SALAH (BOOLEAN) ---
		{
			QuizID:        quizJarkom.ID,
			QuestionText:  "UDP adalah protokol yang reliable dan menjamin urutan data.",
			Type:          "boolean",
			Options:       []string{"Benar", "Salah"},
			CorrectAnswer: "Salah",
			Hint:          "UDP itu connectionless dan best-effort.",
		},
		{
			QuizID:        quizJarkom.ID,
			QuestionText:  "IPv6 menggunakan alamat 128-bit.",
			Type:          "boolean",
			Options:       []string{"Benar", "Salah"},
			CorrectAnswer: "Benar",
			Hint:          "IPv4 cuma 32-bit.",
		},

		// --- TIPE: MULTI SELECT (JAWABAN > 1) ---
		// PENTING: CorrectAnswer harus format JSON Array string
		{
			QuizID:        quizJarkom.ID,
			QuestionText:  "Manakah protokol yang berada di Application Layer?",
			Type:          "multi_select",
			Options:       []string{"HTTP", "TCP", "DNS", "IP"},
			CorrectAnswer: `["HTTP", "DNS"]`, // Perhatikan format string JSON ini
			Hint:          "TCP di Transport, IP di Network.",
		},
	}

	for _, q := range questionsJarkom {
		DB.Create(&q)
	}

	topicIMK := models.Topic{Slug: "imk", Title: "Interaksi Manusia & Komputer", Description: "Usability, UX, dan Design Principle"}
	DB.Create(&topicIMK)

	quizIMK := models.Quiz{TopicID: topicIMK.ID, Title: "Kuis IMK Dasar", Description: "Konsep dasar IMK"}
	DB.Create(&quizIMK)

	questionsIMK := []models.Question{
		{
			QuizID:        quizIMK.ID,
			QuestionText:  "IMK (Interaksi Manusia dan Komputer) berfokus pada hubungan antara:",
			Options:       []string{"Sistem operasi dan perangkat keras", "Pengguna dan antarmuka komputer", "Komputer dan jaringan", "Database dan server"},
			CorrectAnswer: "Pengguna dan antarmuka komputer",
			Hint:          "Manusia (User) dengan Interface.",
		},
		{
			QuizID:        quizIMK.ID,
			QuestionText:  "Tujuan utama IMK adalah:",
			Options:       []string{"Mempercepat komputasi", "Membuat sistem mudah digunakan", "Mengurangi ukuran aplikasi", "Memperbesar kapasitas memori"},
			CorrectAnswer: "Membuat sistem mudah digunakan",
			Hint:          "Usability.",
		},
		{
			QuizID:        quizIMK.ID,
			QuestionText:  "Yang bukan terminologi dalam IMK adalah:",
			Options:       []string{"Usability", "User Experience", "Compiler Optimization", "User Interface"},
			CorrectAnswer: "Compiler Optimization",
			Hint:          "Istilah teknis backend/mesin.",
		},
		{
			QuizID:        quizIMK.ID,
			QuestionText:  "Dalam UCD (User-Centered Design), fokus utama adalah pada:",
			Options:       []string{"Teknologi", "Programmer", "Pengguna", "Desainer"},
			CorrectAnswer: "Pengguna",
			Hint:          "User-Centered.",
		},
		{
			QuizID:        quizIMK.ID,
			QuestionText:  "Prototyping dalam IMK bertujuan untuk:",
			Options:       []string{"Meningkatkan performa CPU", "Menguji ide desain sebelum dikembangkan penuh", "Menghemat RAM", "Menentukan spesifikasi hardware"},
			CorrectAnswer: "Menguji ide desain sebelum dikembangkan penuh",
			Hint:          "Trial sebelum final.",
		},
		{
			QuizID:        quizIMK.ID,
			QuestionText:  "Dialog style yang menggunakan menu pilihan termasuk:",
			Options:       []string{"Command line", "Form-fill", "Menu-driven", "Question & Answer"},
			CorrectAnswer: "Menu-driven",
			Hint:          "Memilih dari daftar.",
		},
		{
			QuizID:        quizIMK.ID,
			QuestionText:  "Faktor terpenting dalam tipografi untuk keterbacaan adalah:",
			Options:       []string{"Warna casing laptop", "Jenis dan ukuran font", "Kecepatan internet", "Versi browser"},
			CorrectAnswer: "Jenis dan ukuran font",
			Hint:          "Berhubungan dengan huruf.",
		},
		{
			QuizID:        quizIMK.ID,
			QuestionText:  "Penanganan kesalahan (error handling) sebaiknya:",
			Options:       []string{"Menyalahkan pengguna", "Menghentikan aplikasi secara tiba-tiba", "Memberikan pesan yang jelas dan solusi", "Menyembunyikan pesan error"},
			CorrectAnswer: "Memberikan pesan yang jelas dan solusi",
			Hint:          "Helpful error message.",
		},
		{
			QuizID:        quizIMK.ID,
			QuestionText:  "Yang termasuk pihak terlibat dalam IMK adalah:",
			Options:       []string{"UI designer, user, programmer", "Admin server saja", "Teknisi jaringan saja", "Hardware engineer saja"},
			CorrectAnswer: "UI designer, user, programmer",
			Hint:          "Ada desainer, pembuat, dan pengguna.",
		},
		{
			QuizID:        quizIMK.ID,
			QuestionText:  "UI yang lebih kaya animasi namun mengganggu pengguna tergolong:",
			Options:       []string{"High usability", "Cognitive overload", "Perfect UI", "Hardware-driven"},
			CorrectAnswer: "Cognitive overload",
			Hint:          "Otak terlalu penuh informasi.",
		},
		{
			QuizID:       quizIMK.ID,
			QuestionText: "Apa kepanjangan dari GUI?",
			Type:         "short_answer",
			Options:      []string{},
			CorrectAnswer: "Graphical User Interface",
			Hint:         "Antarmuka Grafis.",
		},
		{
			QuizID:       quizIMK.ID,
			QuestionText: "Hukum yang memprediksi waktu untuk menunjuk target disebut Hukum ...",
			Type:         "short_answer",
			Options:      []string{},
			CorrectAnswer: "Fitts",
			Hint:         "Nama orang, diawali huruf F.",
		},

		// --- TIPE: BENAR / SALAH (BOOLEAN) ---
		{
			QuizID:       quizIMK.ID,
			QuestionText: "Prototyping sebaiknya dilakukan setelah produk selesai dikoding sepenuhnya.",
			Type:         "boolean",
			Options:      []string{"Benar", "Salah"},
			CorrectAnswer: "Salah",
			Hint:         "Justru dilakukan di awal untuk testing ide.",
		},
		{
			QuizID:       quizIMK.ID,
			QuestionText: "Konsistensi adalah salah satu prinsip utama dalam desain antarmuka.",
			Type:         "boolean",
			Options:      []string{"Benar", "Salah"},
			CorrectAnswer: "Benar",
			Hint:         "Agar user tidak bingung.",
		},

		// --- TIPE: MULTI SELECT (JAWABAN > 1) ---
		{
			QuizID:       quizIMK.ID,
			QuestionText: "Pilih komponen Usability menurut Nielsen:",
			Type:         "multi_select",
			Options:      []string{"Learnability", "Efficiency", "Coding Speed", "Memorability"},
			CorrectAnswer: `["Learnability", "Efficiency", "Memorability"]`, // Format JSON String
			Hint:         "Kecepatan ngoding bukan urusan user.",
		},
		{
			QuizID:       quizIMK.ID,
			QuestionText: "Manakah yang termasuk jenis ragam dialog?",
			Type:         "multi_select",
			Options:      []string{"Command Line", "Direct Manipulation", "Backend Processing", "Menu Selection"},
			CorrectAnswer: `["Command Line", "Direct Manipulation", "Menu Selection"]`,
			Hint:         "Cara user berinteraksi, bukan proses belakang layar.",
		},
	}

	for _, q := range questionsIMK {
		DB.Create(&q)
	}

	fmt.Println("✅ Exam Data Seeded Successfully!")
}
