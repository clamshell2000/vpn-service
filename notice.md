# GitHub Repository Guidelines - VPN Service

## Những gì NÊN đưa lên GitHub

1. **Mã nguồn**:
   - Tất cả các file mã nguồn trong `backend/src/core`, `backend/src/utils`, `backend/api`
   - Cấu trúc thư mục dự án
   - Unit tests trong `backend/tests/unit` và integration tests trong `backend/tests/e2e`
   - Tài liệu Swagger API trong `backend/docs/swagger`

2. **Cấu hình template**:
   - Docker Compose templates trong `infrastructure/docker` (không chứa thông tin nhạy cảm)
   - Nginx configuration templates trong `infrastructure/nginx/conf.d` (không chứa thông tin nhạy cảm)
   - Prometheus và Grafana configuration templates trong `infrastructure/monitoring`
   - WireGuard configuration templates trong `backend/vpn/wireguard/config_templates` (không chứa keys)

3. **Tài liệu**:
   - README.md với hướng dẫn cài đặt và sử dụng
   - development-structure.md (tài liệu kiến trúc)
   - Hướng dẫn API
   - Tài liệu troubleshooting
   - Sơ đồ kiến trúc (PlantUML)

4. **Scripts**:
   - Script kiểm tra trạng thái (`scripts/check-status.sh`)
   - Script tạo peer (`scripts/create-peer.sh`) (không chứa keys)
   - Script bảo trì (`scripts/maintenance.sh`)
   - Script troubleshooting (`scripts/troubleshoot.sh`)

5. **CI/CD**:
   - GitHub Actions workflows
   - Cấu hình CI/CD
   - Cấu hình linting và testing

## Những gì KHÔNG NÊN đưa lên GitHub

1. **Thông tin xác thực**:
   - Private keys và WireGuard keys trong `backend/data/wg_configs`
   - Mật khẩu và tokens
   - JWT signing keys
   - SSH keys
   - API keys (AWS, third-party services)

2. **Cấu hình với thông tin nhạy cảm**:
   - File .env chứa biến môi trường
   - Cấu hình production với thông tin thật
   - Connection strings với thông tin xác thực database trong `backend/config`
   - Thông tin IP thật (như 54.254.241.55, 116.106.201.170) trong cấu hình

3. **Dữ liệu người dùng**:
   - Database dumps
   - Thông tin người dùng trong `backend/db/models`
   - Logs chứa thông tin người dùng trong `backend/logs`
   - Thông tin IP của người dùng trong `backend/logs/usage_analytics.log`

4. **Thông tin cơ sở hạ tầng**:
   - IP addresses thật của máy chủ EC2
   - Thông tin cấu hình firewall
   - Thông tin cấu hình AWS security groups
   - Chi tiết về cấu trúc mạng nội bộ

5. **Dữ liệu nhạy cảm khác**:
   - Certificates và private keys SSL trong `infrastructure/nginx/ssl`
   - Báo cáo đánh giá bảo mật
   - Thông tin về lỗ hổng bảo mật

## Sử dụng .gitignore

Tạo file `.gitignore` với các mục sau:

```
# Environment variables
.env
*.env
.env.*

# Keys and certificates
*.key
*.pem
*.crt
private/
secrets/
infrastructure/nginx/ssl/

# WireGuard specific
wg-private-key
wg-preshared-key
*.conf
backend/data/wg_configs/peer*/

# Logs
*.log
logs/
backend/logs/

# Database
*.db
*.sqlite
*.sqlite3
postgres-data/

# Build artifacts
bin/
dist/
build/

# Docker volumes
volumes/
data/

# IDE specific
.idea/
.vscode/
*.swp
*.swo

# OS specific
.DS_Store
Thumbs.db
```

## Sử dụng Secrets Management

1. **Sử dụng GitHub Secrets** cho CI/CD workflows
2. **Sử dụng AWS Secrets Manager** hoặc HashiCorp Vault cho quản lý secrets trong production
3. **Sử dụng template files** với placeholders thay vì hardcoded secrets

## Quy trình an toàn

1. **Kiểm tra code trước khi commit** để đảm bảo không có secrets bị lộ
2. **Sử dụng git-secrets** để tự động quét secrets trước khi commit
3. **Thực hiện code review** để đảm bảo không có thông tin nhạy cảm
4. **Sử dụng branch protection** để yêu cầu approval trước khi merge

## Xử lý thông tin nhạy cảm trong dự án VPN

### Cách xử lý WireGuard keys
- Sử dụng environment variables hoặc secrets manager để lưu trữ keys
- Tạo script để tự động generate keys khi deploy
- Không bao giờ commit keys vào repository

### Cách xử lý thông tin server
- Sử dụng placeholders cho IP addresses (ví dụ: `SERVER_IP`)
- Lưu trữ thông tin thật trong file cấu hình riêng không được commit
- Sử dụng CI/CD để tự động thay thế placeholders khi deploy

### Cách xử lý thông tin database
- Sử dụng environment variables cho database credentials
- Tạo database migrations scripts không chứa dữ liệu thật
- Sử dụng mock data cho development và testing

## Tài liệu liên quan

- [GitHub Security Best Practices](https://docs.github.com/en/code-security/getting-started/github-security-best-practices)
- [Git - gitignore Documentation](https://git-scm.com/docs/gitignore)
- [Vault by HashiCorp](https://www.vaultproject.io/)
- [AWS Secrets Manager](https://aws.amazon.com/secrets-manager/)

---

**Lưu ý**: Đây là hướng dẫn cơ bản. Hãy luôn cập nhật và tuân thủ các best practices mới nhất về bảo mật khi làm việc với mã nguồn và thông tin nhạy cảm.

**Ngày cập nhật**: 2025-03-25
