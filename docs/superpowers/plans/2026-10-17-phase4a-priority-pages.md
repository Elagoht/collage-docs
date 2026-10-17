# Aşama 4a: Öncelikli beş sayfaya şema, kod ve senaryo Uygulama Planı

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** `fragments-and-slots`, `caching`, `forms-and-actions`, `middleware-and-apis` ve `plugins` sayfaları (EN + TR) şemalar, daha fazla kod ve gerçek dünya senaryolarıyla daha kolay okunur hale gelir; hiçbir bilgi kaybolmaz.

**Architecture:** Her sayfa bağımsız bir görevdir ve yalnızca kendi iki dosyasına (`content/<slug>.md`, `content/tr/<slug>.md`) dokunur. Şemalar satır içi SVG'dir ve mevcut `figure.diagram-figure` + `d-*` sınıflarını kullanır; CSS'e dokunulmaz. Senaryolar üç tekrar eden, kurgusal alandan gelir: bir blog, bir ürün kataloğu ve çok kullanıcılı, rol bazlı bir web uygulaması (ör. bir ekibin iş listesi).

**Tech Stack:** Markdown (goldmark, ham HTML açık), satır içi SVG, Go örnekleri (collage v0.65.0).

**Spec:** `docs/superpowers/specs/2026-10-17-docs-positioning-design.md` (Kararlar → 6)

## Global Constraints

- Bilgi silinmez: mevcut her olgu, tablo, hata adı ve kod örneği sayfada kalır. Yeniden düzenlenebilir, kısaltılabilir paragraflara bölünebilir.
- Her yeni iddia collage v0.65.0 kaynağına (`$(go env GOMODCACHE)/github.com/!elagoht/collage@v0.65.0`) ya da mevcut dokümana göre doğrulanır. Uydurma API, alan ya da davranış yok.
- Başka hiçbir framework, kütüphane, runtime ya da ürün adı yeni metne, koda ya da commit mesajına girmez. Mevcut sayfalardaki eski isimlere dokunulmaz.
- Yeni metinde "since vX" ya da "eskiden" anlatısı yok; mevcut olanlar kalır.
- Go örneklerinde `any` yok.
- Şema: `<figure class="diagram-figure" data-search-ignore>`, `<svg class="diagram" viewBox="0 0 320 …" role="img" aria-labelledby="<benzersiz-id>">`, `<title id>`, `<figcaption>`; yalnızca mevcut `d-*` sınıfları; `fill="#"`/`stroke="#"`/`style=` yok; figure bloğunda boş satır yok; genişlik 320 (telefonda okunur); `<text>` etiketleri kutularına sığar (~6px/karakter, `d-small` ~5.4px). Marker id'si sayfada benzersiz.
- Her sayfa için: en az bir şema (sayfanın ana kavramı), en az bir "gerçek dünya" bölümü ya da kutusu (üç alandan biri), mevcut uzun paragrafların okunur kısa paragraflara bölünmesi.
- TR dosyası EN ile aynı başlıkları seviye seviye taşır (yeni başlık eklenirse ikisine de). TR'de "parça" yok ("kısım"); terimler mevcut TR sayfalarındaki gibi İngilizce kalır; kod yorumları İngilizce.
- `description` en fazla 160 karakter.
- Her görevin sonunda: `go run ./tools/checkcontent`, `go test ./...` ve `go run . -collage-build -out <tmp> -clean` temiz (htmlcheck uyarısı yok).

## Review Focus

- Bir SVG etiketinin kutusundan taşması (dar ekran) — görsel kontrol.
- Bir şemanın davranışı yanlış göstermesi (ör. cache'lenmeyen bir şeyi cache'ten geçirmek) — kaynakla karşılaştırma.
- TR başlık yapısının EN'den kayması — `checkcontent`.
- Bir bağlantının olmayan bir başlığa gitmesi — testler.
- Senaryo kodunun derlenmeyecek bir imza kullanması — kaynakla karşılaştırma.

---

### Task 1: `fragments-and-slots`

Şema: layout → content → slot'lar ağacı; kardeş fragment'lerin aynı anda veri çekmesi ve sırayla yazılması; başarısızlık politikası (required / fallback / boş). Senaryo: ürün kataloğu sayfası (ürün, stok kutusu, "benzer ürünler" slot'u; stok servisi düşünce kutu boş kalır, page 500 olmaz).

### Task 2: `caching`

Şema 1: render stratejileri (Static / Incremental / Dynamic) karar akışı. Şema 2: tag'ler → page cache + data cache, `InvalidateTags` (stale + warming) ile `PurgeTags` (tamamen sil) farkı. Senaryo: blog (yazar profilini güncelleyen bir CMS webhook'u `author:<id>` tag'ini invalidate eder; GDPR silme isteği `PurgeTags` kullanır).

### Task 3: `forms-and-actions`

Şema: bir form gönderiminin yolu (CSRF + origin kontrolü → action → 422 render / 303 redirect / `Collage-Fetch` ile fragment). Senaryo: ekip iş listesinde "görev ekle" formu: doğrulama hatası 422 ile aynı formu gösterir, başarı 303 ile listeye döner, `fetch` ile yalnızca liste fragment'i döner. `update-part-of-a-page` rehberine link.

### Task 4: `middleware-and-apis`

Şema: middleware katmanları (plugin middleware'leri → `app.Use` sırası → router → page / action / `app.Handle`). Senaryo: rol bazlı web uygulaması — oturumdaki kullanıcıyı yükleyen middleware, mevcut bir JSON API'nin `app.Handle` ile takılması ve API değişikliğinin `InvalidateTags` çağırması.

### Task 5: `plugins`

Şema: bir plugin'in bağlandığı noktalar (middleware, render öncesi/sonrası, cache yazımı, build sonrası). Senaryo tablosu: "ne kuruyorsunuz → hangi plugin'ler" (blog, ürün kataloğu, çok kullanıcılı uygulama). Ayrıca "What a plugin can do" bölümündeki "plugin middleware'i uygulamanınkinden sonra" ifadesi, kaynağa (plugin middleware'leri `app.Use`'unkilerin dışında, liste sırasıyla) ve `middleware-and-apis`'e göre düzeltilir.

### Her görevin adımları

- [ ] Sayfayı ve ilgili kaynak kodunu oku; değişiklik listesini çıkar.
- [ ] EN sayfasını düzenle (şema + senaryo + paragraflar), iddiaları kaynağa göre doğrula.
- [ ] TR sayfasını aynı yapıyla düzenle.
- [ ] `go run ./tools/checkcontent`, `go test ./...`, export → temiz.
- [ ] Rapor: eklenen şemalar, senaryolar, doğrulanan iddialar ve kaynak satırları, emin olunamayanlar.

Görevler birbirinin dosyalarına dokunmadığı için paralel yürütülür. Ardından her sayfa `turkish-reviewer`'dan geçer, görsel kontrol yapılır ve tek bir son inceleme yapılır.
