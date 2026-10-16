# ==============================================================================
# GoSecure-Daemon Makefile (Build & Deployment Automation)
# ==============================================================================

BINARY_NAME=gosecure-daemon
BUILD_DIR=bin
CMD_DIR=./cmd/daemon
SERVICE_NAME=gosecure.service
SYSTEMD_PATH=/etc/systemd/system
INSTALL_PATH=/usr/local/bin

.PHONY: all build test race run clean install-service uninstall-service logs status

all: test build

## build: Kompilasi binary daemon
build:
	@echo "==> Mengompilasi binary $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)
	@echo "==> Build selesai: $(BUILD_DIR)/$(BINARY_NAME)"

## test: Jalankan unit test standar
test:
	@echo "==> Jalankan unit testing..."
	go test -v ./...

## race: Jalankan race detector untuk menguji konkurensi
race:
	@echo "==> Jalankan race detector..."
	go test -race -v ./...

## run: Menjalankan daemon dalam mode pengembangan
run: build
	@echo "==> Menjalankan $(BINARY_NAME) secara lokal..."
	./$(BUILD_DIR)/$(BINARY_NAME)

## clean: Bersihkan direktori build
clean:
	@echo "==> Membersihkan file kompilasi..."
	rm -rf $(BUILD_DIR)

## install-service: Otomatisasi instalasi dan deployment ke Systemd (butuh sudo)
install-service: build
	@echo "==> Memasang binary ke $(INSTALL_PATH)..."
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_PATH)/
	sudo chmod +x $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "==> Memasang unit service ke $(SYSTEMD_PATH)..."
	sudo cp deployments/$(SERVICE_NAME) $(SYSTEMD_PATH)/
	@echo "==> Reloading systemd daemon..."
	sudo systemctl daemon-reload
	@echo "==> Mengaktifkan dan menjalankan $(SERVICE_NAME)..."
	sudo systemctl enable --now $(SERVICE_NAME)
	@echo "==> Deployment berhasil! Cek status dengan 'make status'"

## uninstall-service: Menghentikan dan menghapus daemon dari Systemd (butuh sudo)
uninstall-service:
	@echo "==> Menghentikan dan menonaktifkan service..."
	-sudo systemctl disable --now $(SERVICE_NAME)
	@echo "==> Menghapus file service dan binary..."
	-sudo rm -f $(SYSTEMD_PATH)/$(SERVICE_NAME)
	-sudo rm -f $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "==> Reloading systemd daemon..."
	sudo systemctl daemon-reload
	@echo "==> Uninstallation selesai."

## logs: Memantau log real-time dari systemd journal
logs:
	sudo journalctl -u $(SERVICE_NAME) -f -o json-pretty

## status: Memeriksa status operasional service systemd
status:
	sudo systemctl status $(SERVICE_NAME)