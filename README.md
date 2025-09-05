# TalentPass

TalentPass, yazılım geliştiricilerin ve profesyonellerin iş
başvurularını daha düzenli bir şekilde takip etmesini sağlayan modern
bir backend servisidir.\
Go dilinde yazılmıştır, PostgreSQL veritabanı kullanır ve Docker Compose
ile kolayca ayağa kaldırılabilir.

------------------------------------------------------------------------

## 🚀 Özellikler

### 🔐 Kimlik Doğrulama

-   JWT tabanlı login/register akışı\
-   Yetkilendirme middleware (`RequireAuth`)

### 💼 İş İlanları (Jobs)

-   İş ilanı oluşturma, listeleme, güncelleme ve silme\
-   Tag ve filtreleme desteği\
-   Alanlar: `title`, `company`, `url`, `location`, `tags`,
    `created_at`, `updated_at`

### 📄 Başvurular (Applications)

-   Bir ilana başvuru yapma\
-   Kullanıcının kendi başvurularını listeleme\
-   Başvuru durumunu güncelleme (`applied`, `interview`, `offer`,
    `denied`)\
-   Opsiyonel takip tarihi: `next_action_at`

### 🏢 Organizasyonlar (Orgs) *(yapım aşamasında)*

-   Organizasyon oluşturma\
-   Üyelik ve rol yönetimi\
-   Org bazlı iş ilanı yayınlama

### ⚙️ Altyapı

-   **PostgreSQL** → SQLC ile strongly-typed sorgular\
-   **Redis** (planlanan) → caching ve oturum yönetimi\
-   **Mailhog** → test amaçlı e-posta yakalama\
-   **Zerolog** → structured logging\
-   **Rate Limit Middleware** → IP başına 120 istek/dk\
-   **Sağlık kontrolü** endpoint: `/healthz`

------------------------------------------------------------------------

## 🛠️ Kurulum

### Gereksinimler

-   Go 1.22+\
-   Docker & Docker Compose\
-   PostgreSQL 15+\
-   Goose (migration aracı)

### Adımlar

``` powershell
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
```

------------------------------------------------------------------------

## 📡 API Endpointleri

### Auth

-   `POST /v1/auth/register` → kullanıcı kaydı\
-   `POST /v1/auth/login` → giriş yap ve JWT token al

### Jobs

-   `POST /v1/jobs` → iş ilanı oluştur\
-   `GET /v1/jobs` → ilanları listele\
-   `GET /v1/jobs/{id}` → ilan detaylarını getir\
-   `PUT /v1/jobs/{id}` → ilan güncelle\
-   `DELETE /v1/jobs/{id}` → ilan sil

### Applications

-   `POST /v1/applications` → başvuru yap\
-   `GET /v1/applications` → kendi başvurularını listele\
-   `PATCH /v1/applications/{id}:status` → başvuru durumunu güncelle

### Health

-   `GET /healthz` → servis durumu\
    \`\`\`

------------------------------------------------------------------------
