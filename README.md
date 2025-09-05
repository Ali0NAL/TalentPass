# TalentPass

TalentPass

TalentPass, yazılım geliştiricilerin ve profesyonellerin iş başvurularını daha düzenli bir şekilde takip etmesini sağlayan modern bir backend servisidir.
Go dilinde yazılmıştır, PostgreSQL veritabanı kullanır ve Docker Compose ile kolayca ayağa kaldırılabilir.

 Özellikler
 Kimlik Doğrulama

JWT tabanlı login/register akışı

Yetkilendirme middleware (RequireAuth)

 İş İlanları (Jobs)

İş ilanı oluşturma, listeleme, güncelleme ve silme

Tag ve filtreleme desteği

Alanlar: title, company, url, location, tags, created_at, updated_at

 Başvurular (Applications)

Bir ilana başvuru yapma

Kullanıcının kendi başvurularını listeleme

Başvuru durumunu güncelleme (applied, interview, offer, denied)

Opsiyonel takip tarihi (next_action_at)

 Organizasyonlar (Orgs) (yapım aşamasında)

Organizasyon oluşturma

Üyelik ve rol yönetimi

Org bazlı iş ilanı yayınlama

 Altyapı

PostgreSQL → SQLC ile strongly-typed sorgular

Redis (planlanan) → caching ve oturum yönetimi

Mailhog → test amaçlı e-posta yakalama

Zerolog → structured logging

Rate Limit Middleware → IP başına 120 istek/dk

Sağlık kontrolü endpoint: /healthz

# Kurulum

Gereksinimler

Go 1.22+

Docker & Docker Compose

PostgreSQL 15+

Goose (migration aracı)


# Ortam değişkenlerini ayarla
$env:DATABASE_URL="postgres://postgres:postgres@localhost:5432/talentpass?sslmode=disable"
$env:JWT_SECRET="dev-secret-change-me"
$env:PORT="8080"

# Servisleri ayağa kaldır
docker compose up -d

# Migration çalıştır
goose -dir ./migrations postgres "$env:DATABASE_URL" up

# API başlat
go run ./cmd/api
