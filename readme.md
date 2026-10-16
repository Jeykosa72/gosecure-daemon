# GoSecure-Daemon

**Secure File Monitor & Process Control Daemon**

GoSecure-Daemon adalah layanan latar belakang (*system daemon*) berbasis Go yang dirancang untuk memantau direktori sensitif secara *real-time*, memproses perubahan berkas melalui *worker pool* berbasis CSP (*Communicating Sequential Processes*), serta menjalankan analisis keamanan melalui subprocess dengan batas waktu eksekusi (*execution timeout*).

Daemon ini juga menyediakan pencatatan log terstruktur berbasis JSON untuk mendukung observabilitas dan integrasi dengan sistem manajemen log di Linux.

## Arsitektur Sistem

```text
[OS Filesystem]
       |
       | Inotify / fsnotify
       v
[FileWatcher Goroutine]
       |
       | Job Channel (CSP)
       v
[Worker Pool (4 Workers)]
       |
       | context.WithTimeout
       v
[Secure Process Runner]
       |
       +--> sha256sum / ClamAV
       |
       v
[Structured JSON Logger]
       |
       | log/slog
       v
[systemd Journal / journald]
```

## Fitur Utama dan Keamanan

### 1. Automated Build and Deployment

Dilengkapi `Makefile` untuk mengotomatisasi berbagai aktivitas pengembangan dan operasional, termasuk:

- Kompilasi binary daemon.
- Eksekusi unit test dan race detector.
- Menjalankan daemon dalam mode pengembangan.
- Instalasi dan penghapusan systemd service.
- Pemeriksaan status service.
- Pemantauan log menggunakan `journalctl`.

### 2. Structured Logging dengan `log/slog`

Menggunakan package standar Go `log/slog` untuk menghasilkan log terstruktur dalam format JSON.

Fitur logging meliputi:

- Level log `DEBUG`, `INFO`, `WARN`, dan `ERROR`.
- Pemfilteran level log melalui variabel lingkungan `LOG_LEVEL`.
- Integrasi dengan systemd journal melalui standard output atau standard error.
- Potensi integrasi dengan sistem agregasi log seperti Fluentd, Elasticsearch, dan Grafana Loki.

### 3. Systemd Integration and Hardening

Menyediakan unit file `gosecure.service` untuk menjalankan daemon sebagai service Linux.

Konfigurasi service dapat mencakup:

- Pengelolaan proses oleh systemd.
- Kebijakan restart ketika proses mengalami kegagalan.
- Pembatasan resource, seperti jumlah file descriptor melalui `LimitNOFILE`.
- Pembatasan hak akses dan kapabilitas proses melalui konfigurasi systemd.

Ketersediaan setiap fitur bergantung pada konfigurasi aktual dalam unit file.

### 4. Principle of Least Privilege

Jika direktori pengawasan belum tersedia, daemon dapat membuat direktori tersebut dengan permission `0750` (`rwxr-x---`).

Permission ini membatasi akses berdasarkan pemilik dan grup direktori. Hak akses efektif tetap dipengaruhi oleh konfigurasi pengguna, grup, dan permission sistem.

### 5. Process Isolation and Execution Timeout

Setiap pekerjaan analisis dijalankan melalui subprocess menggunakan package `os/exec` dan mekanisme `context.WithTimeout`.

Pendekatan ini memungkinkan daemon membatasi durasi eksekusi proses analisis. Jika batas waktu tercapai, proses yang dijalankan melalui `exec.CommandContext` akan dihentikan sesuai mekanisme penghentian yang didukung oleh sistem operasi.

Penghentian proses utama tidak selalu menjamin seluruh proses turunannya ikut berhenti.

### 6. Graceful Shutdown

Daemon menangani sinyal penghentian seperti `SIGINT` (Ctrl+C) dan `SIGTERM` (systemd).

Proses shutdown dirancang untuk mengelola penghentian file watcher, penutupan antrean pekerjaan, dan penyelesaian pekerjaan yang masih berjalan melalui mekanisme sinkronisasi seperti `sync.WaitGroup`.

Jaminan penyelesaian seluruh pekerjaan bergantung pada implementasi shutdown dan kebijakan penanganan antrean.

### 7. Concurrency Control and Race Detection

Menggunakan goroutine, channel, dan worker pool untuk mengatur pemrosesan pekerjaan secara konkuren.

Pendekatan ini membantu membatasi jumlah pekerjaan yang diproses secara bersamaan dan mengendalikan penggunaan resource. Penanganan antrean penuh tetap perlu ditentukan agar sistem memiliki perilaku yang jelas ketika kapasitas antrean tercapai.

Race detector Go dapat digunakan untuk mendeteksi potensi *data race* pada jalur eksekusi yang diuji.

## Struktur Proyek

```text
gosecure-daemon/
├── cmd/
│   └── daemon/
│       └── main.go           # Entry point, signal handling, dan shutdown
├── internal/
│   ├── config/
│   │   └── config.go         # Konfigurasi dan parsing environment variables
│   ├── logger/
│   │   └── logger.go         # Inisialisasi structured logger
│   ├── monitor/
│   │   └── watcher.go        # File watcher berbasis fsnotify
│   ├── process/
│   │   └── runner.go         # Eksekusi subprocess dengan timeout
│   └── worker/
│       ├── pool.go           # Worker pool berbasis channel dan WaitGroup
│       └── pool_test.go      # Pengujian worker pool dan konkurensi
├── gosecure.service          # Unit file systemd
├── Makefile                  # Otomatisasi build, test, dan deployment
├── go.mod                    # Definisi modul dan dependency Go
└── README.md                 # Dokumentasi proyek
```

## Persyaratan Sistem

- **Sistem operasi:** Linux dengan systemd untuk deployment sebagai service.
- **Go:** Versi yang sesuai dengan deklarasi pada `go.mod`.
- **Build tools:** `make` dan toolchain Go.
- **Filesystem watcher:** Dependency `fsnotify` atau implementasi yang digunakan proyek.
- **Tools analisis:** `sha256sum` dan/atau ClamAV jika digunakan oleh konfigurasi runner.

> Catatan: Penggunaan systemd tidak berlaku untuk semua lingkungan macOS. Mode pengembangan lokal bergantung pada dukungan filesystem watcher dan dependency yang digunakan.

## Environment Variables

| Variabel | Default | Deskripsi |
|---|---|---|
| `WATCH_DIR` | `/tmp/gosecure-target` | Path direktori yang dipantau daemon. |
| `LOG_LEVEL` | `INFO` | Level logging: `DEBUG`, `INFO`, `WARN`, atau `ERROR`. |

Nilai default di atas mengikuti konfigurasi yang didokumentasikan dan perlu disesuaikan jika implementasi pada kode sumber berbeda.

## Instalasi dan Penggunaan

### 1. Clone Repository

Jika proyek tersedia di repository Git, jalankan:

```bash
git clone <REPOSITORY_URL>
cd gosecure-daemon
```

Ganti `<REPOSITORY_URL>` dengan URL repository yang sebenarnya.

Jika proyek sudah berada di komputer lokal, langsung masuk ke direktori proyek.

### 2. Persiapkan Dependency

Jalankan:

```bash
go mod tidy
```

Perintah ini menyelaraskan dependency yang dibutuhkan dengan kode sumber proyek.

### 3. Build Binary

Kompilasi daemon menggunakan `Makefile`:

```bash
make build
```

Berdasarkan konfigurasi `Makefile`, binary diharapkan tersedia di:

```text
bin/gosecure-daemon
```

Jika ingin mengompilasi secara langsung, sesuaikan path entry point dengan struktur proyek:

```bash
mkdir -p bin
go build -o bin/gosecure-daemon ./cmd/daemon
```

### 4. Jalankan dalam Mode Pengembangan

Jalankan daemon menggunakan:

```bash
make run
```

Untuk mengatur direktori pengawasan dan level log:

```bash
WATCH_DIR=/tmp/gosecure-target LOG_LEVEL=DEBUG make run
```

Perintah tersebut mengasumsikan target `run` pada `Makefile` meneruskan environment variables ke proses daemon.

### 5. Uji Pemantauan File secara Real-Time

Buka terminal terpisah dan buat file baru pada direktori yang dipantau:

```bash
mkdir -p /tmp/gosecure-target
echo "Sample Payload Data" > /tmp/gosecure-target/sample.txt
```

Periksa log daemon untuk memastikan file terdeteksi dan pekerjaan analisis diteruskan ke worker pool.

Hasil pengujian bergantung pada implementasi watcher, runner, dan konfigurasi logging.

## Deployment Produksi dengan systemd

### 1. Instalasi Service

Untuk mengompilasi dan memasang daemon sebagai service systemd, jalankan:

```bash
sudo make install-service
```

Target `install-service` harus didefinisikan pada `Makefile` dan menangani instalasi binary serta unit file dengan benar.

Pastikan konfigurasi service menggunakan lokasi binary dan direktori kerja yang sesuai.

### 2. Periksa Status Service

```bash
make status
```

Jika target tersebut belum tersedia atau memerlukan hak administratif, periksa status secara langsung:

```bash
sudo systemctl status gosecure.service
```

### 3. Pantau Log Service

Untuk memantau log secara real-time:

```bash
make logs
```

Alternatif menggunakan `journalctl` secara langsung:

```bash
sudo journalctl -u gosecure.service -f
```

Jika logger dikonfigurasi menggunakan JSON handler dan output diarahkan ke journal, log akan tersedia melalui systemd journal.

### 4. Pengujian pada Service yang Berjalan

Buat file uji:

```bash
echo "Sample Payload Data" | sudo tee /tmp/gosecure-target/sample.txt
```

Periksa log untuk memverifikasi bahwa file terdeteksi dan proses analisis dijalankan.

Pastikan permission direktori dan hak akses pengguna service memungkinkan daemon membaca direktori target.

### 5. Menghapus Service

Jika target `uninstall-service` sudah didefinisikan pada `Makefile`, jalankan:

```bash
sudo make uninstall-service
```

Pastikan target tersebut hanya menghapus resource yang memang dikelola oleh proyek.

## Pengujian dan Race Detector

### Menjalankan Unit Test

Untuk menjalankan seluruh unit test:

```bash
go test ./...
```

### Menjalankan Race Detector

Gunakan target `race` pada `Makefile`:

```bash
make race
```

Atau jalankan secara langsung:

```bash
go test -race ./...
```

Race detector membantu menemukan akses konkuren yang tidak tersinkronisasi pada jalur program yang dijalankan selama pengujian.

Hasil pengujian yang sukses tidak menjamin seluruh kemungkinan *data race* telah ditemukan.

## Catatan Keamanan

- Batasi executable yang dapat dipanggil oleh process runner. Jangan membiarkan input tidak tepercaya menentukan perintah arbitrer.
- Gunakan timeout untuk mengendalikan durasi pekerjaan analisis.
- Tentukan kebijakan ketika antrean pekerjaan penuh, misalnya menolak pekerjaan baru, melakukan backpressure, atau mencatat pekerjaan yang dilewati.
- Pastikan shutdown menangani pekerjaan aktif dan pekerjaan yang masih berada dalam antrean sesuai kebutuhan sistem.
- Gunakan akun service dengan hak akses minimum yang diperlukan.
- Tinjau konfigurasi systemd, termasuk direktori kerja, permission, restart policy, dan batas resource.
- Hindari menyimpan informasi sensitif di dalam log.
- Uji konfigurasi dan perilaku daemon sebelum menggunakannya pada lingkungan produksi.

## Lisensi
