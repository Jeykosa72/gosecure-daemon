# GoSecure-Daemon

**Secure File Monitor & Process Control Daemon** — layanan latar belakang berbasis Go untuk memantau direktori sensitif secara real-time, memproses perubahan berkas melalui worker pool berbasis CSP (*Communicating Sequential Processes*), dan menjalankan alat analisis keamanan dalam subprocess dengan batas waktu eksekusi.

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
       +--> sha256sum
       +--> ClamAV
```

## Fitur Keamanan dan Ketahanan

### 1. Principle of Least Privilege

Jika direktori target belum tersedia, daemon membuatnya secara otomatis dengan permission `0750` (`rwxr-x---`). Pengaturan ini membatasi akses direktori kepada pemilik dan grup yang memiliki izin.

### 2. Process Isolation dan Execution Timeout

Setiap tugas analisis dijalankan dalam subprocess terpisah menggunakan `os/exec` dan `context.WithTimeout`.

Jika proses analisis melewati batas waktu yang ditentukan, Go akan menghentikan proses yang dijalankan melalui context tersebut tanpa harus menghentikan daemon utama.

### 3. Graceful Shutdown

Daemon menangani sinyal `SIGINT` (Ctrl+C) dan `SIGTERM` (misalnya, dari systemd).

Saat menerima sinyal penghentian, daemon melakukan proses shutdown secara teratur dengan menghentikan watcher, mengelola antrean pekerjaan, dan menunggu worker yang sedang berjalan menyelesaikan tugasnya menggunakan `sync.WaitGroup`.

### 4. Thread Safety dan Resource Control

Daemon menggunakan worker pool dengan jumlah worker terbatas dan buffered channel berkapasitas tetap untuk mengendalikan konkurensi serta membantu mencegah penggunaan memori berlebihan akibat terlalu banyak goroutine.

## Persyaratan Sistem

- **Sistem operasi:** Linux atau macOS, bergantung pada dukungan implementasi filesystem watcher.
- **Go:** Versi 1.20 atau lebih baru.
- **Dependency:** Sesuai dengan yang didefinisikan dalam `go.mod`.
- **Tools analisis:** `sha256sum` dan/atau ClamAV, sesuai konfigurasi yang digunakan.

## Instalasi dan Penggunaan

### 1. Persiapkan Dependency

Pastikan Go sudah terpasang, lalu jalankan perintah berikut dari direktori utama proyek:

```bash
go mod tidy
```

Perintah tersebut menyelaraskan dependency proyek dengan kebutuhan kode sumber.

### 2. Build Binary

Kompilasi proyek menjadi binary:

```bash
mkdir -p bin
go build -o bin/gosecure-daemon .
```

Jika kompilasi berhasil, binary akan tersedia di:

```text
bin/gosecure-daemon
```

### 3. Jalankan Daemon

Jalankan binary menggunakan:

```bash
./bin/gosecure-daemon
```

Pastikan direktori target dan konfigurasi yang diperlukan sudah sesuai dengan implementasi proyek.

### 4. Uji Pemantauan File secara Real-Time

Buka terminal baru, lalu buat file uji di direktori pemantauan `/tmp/gosecure-target`:

```bash
mkdir -p /tmp/gosecure-target
echo "Sample Payload Data" > /tmp/gosecure-target/test_file.txt
```

Periksa log daemon untuk melihat apakah perubahan file terdeteksi dan pekerjaan analisis diproses oleh worker.

Jika implementasi menggunakan `sha256sum`, hasil analisis akan bergantung pada konfigurasi dan mekanisme logging yang diterapkan.

### 5. Jalankan Race Detector

Untuk memeriksa potensi data race pada kode yang diuji, jalankan:

```bash
go test -race ./...
```

Pastikan seluruh pengujian berjalan dengan sukses sebelum menggunakan daemon pada lingkungan yang sebenarnya.

## Catatan Implementasi

- Pastikan direktori yang dipantau sesuai dengan kebutuhan dan hak akses yang diharapkan.
- Batasi executable yang dapat dijalankan agar input yang tidak tepercaya tidak dapat menentukan perintah arbitrer.
- Gunakan `exec.CommandContext` untuk menghubungkan eksekusi subprocess dengan context timeout.
- Perhatikan bahwa penghentian melalui context tidak selalu menjamin seluruh proses turunan ikut dihentikan.
- Pastikan strategi shutdown menangani pekerjaan yang masih berada dalam antrean sesuai kebijakan pemrosesan yang diinginkan.
- Penggunaan buffered channel dan worker pool membantu membatasi konkurensi, tetapi penanganan antrean penuh tetap perlu dirancang secara eksplisit.

## Lisensi

Tambahkan informasi lisensi proyek sesuai lisensi yang dipilih.
