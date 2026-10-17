# Aşama 1: Kökte anasayfa (PrefixDefault export) Uygulama Planı

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** `PrefixDefault: true` olan bir sitenin static export'unda `dist/index.html` boş bir meta-refresh sayfası yerine varsayılan dilin anasayfası olsun; collage-docs bu sürüme geçsin.

**Architecture:** Collage'ın `internal/build/builder.go` içindeki `writeRootRedirect`, anasayfanın zaten yazılmış dosyasını (`dist/<default>/index.html`) köke kopyalar. Sayfada `rel="canonical"` yoksa `<head>`'in hemen ardına `/<default>/`'i gösteren bir canonical ekler; böylece arama motorları iki kopyayı tek adres sayar. Sunucu tarafındaki 301 (`/` → `/en/`) değişmez.

**Tech Stack:** Go 1.26, collage (`~/Desktop/collage`, `github.com/Elagoht/collage`), collage-docs.

**Spec:** `docs/superpowers/specs/2026-10-17-docs-positioning-design.md` (Kararlar → 1. Kök `/` sayfası)

## Global Constraints

- `PrefixDefault` kapatılmaz; collage-docs'un URL'leri değişmez.
- Go kodunda `any` kullanılmaz.
- Metinler bugünkü davranışı anlatır; "eskiden yönlendirmeydi" gibi geçmiş anlatısı yazılmaz (CHANGELOG girdisi hariç; o zaten değişikliği anlatır).
- Collage commit'leri `~/Desktop/collage` reposunda, `main` yerine bir branch'te yapılır; push ve tag Furkan'ın onayıyla.

## Review Focus

- Anasayfası `<head>` içermeyen ya da büyük harf `<HEAD>` kullanan bir şablon: canonical yine eklenmeli (büyük/küçük harf duyarsız eşleşme; `<head>` yoksa dokunmadan kopyala).
- Anasayfası zaten canonical taşıyan site (meta plugin'i): ikinci bir canonical eklenmemeli.
- `<head data-x="…">` gibi öznitelikli head: canonical, açılış etiketinin kapanışından sonra girer.
- Varsayılan dilin anasayfası build edilmemişse kökte hiçbir şey yazılmaz (bugünkü davranış korunur).
- Sayfanın kendi `noindex`'i yoktur ve eklenmez: kök artık gerçek bir sayfa, canonical yeterli.

---

### Task 1: Collage — kökte anasayfanın kopyası

**Files:**
- Modify: `~/Desktop/collage/internal/build/builder.go` (`writeRootRedirect`, ~1150–1210; adı `writeRootHome` olur, çağrı yeri ~493 ve `rootWritten` kullanımları aynı kalır)
- Modify: `~/Desktop/collage/pkg/collage/prefix_test.go` (`TestPrefixDefault_Export`, ~137–165)
- Modify: `~/Desktop/collage/pkg/collage/config.go:399-410` (`PrefixDefault` yorumu: "at its root a page that sends the reader to Default's home" → "at its root a copy of Default's home")
- Modify: `~/Desktop/collage/docs/deployment.md:551`, `~/Desktop/collage/internal/build/capture_test.go` yorumları ("root redirect" → "root home")
- Modify: `~/Desktop/collage/CHANGELOG.md` (en üste `## Unreleased` → `### Changed`)

**Interfaces:**
- Produces: `func (b *Builder) writeRootHome(outDirResolved string, tasks []buildTask, written []string) ([]string, error)` — dönen dilim yazılan kök dosyanın yolu; `capture` mantığı onu bugünkü gibi "builder'ın kendi dosyası" sayar.
- Produces: `func withCanonical(page []byte, href string) []byte` (aynı dosyada, paket içi) — `rel="canonical"` yoksa `<head…>` açılışından hemen sonra `<link rel="canonical" href="…">` ekler; `<head` yoksa sayfayı değiştirmeden döner.

- [ ] **Step 1: Testi yeni davranışa göre yaz**

`TestPrefixDefault_Export` içinde kök kontrolünü değiştir:

```go
root, err := os.ReadFile(filepath.Join(out, "index.html"))
if err != nil {
	t.Fatalf("no root page: %v", err)
}
home, err := os.ReadFile(filepath.Join(out, "en", "index.html"))
if err != nil {
	t.Fatalf("no home page: %v", err)
}
want := strings.Replace(string(home), "<head>", `<head><link rel="canonical" href="/en/">`, 1)
if !strings.Contains(string(home), `rel="canonical"`) && string(root) != want {
	t.Errorf("root is not the home page with a canonical:\n%s", root)
}
if strings.Contains(string(root), `http-equiv="refresh"`) {
	t.Errorf("root still refreshes:\n%s", root)
}
if !strings.Contains(string(root), `<link rel="canonical" href="/en/">`) {
	t.Errorf("root names no canonical /en/:\n%s", root)
}
```

(`prefixApp`'in anasayfası zaten canonical taşıyorsa `root == home` olmalı; testi o duruma göre `bytes.Equal` ile sıkılaştır.)

`builder_internal_test.go` (yoksa `internal/build/rootpage_test.go` oluştur) içine tablo testi `TestWithCanonical`:

| giriş | beklenen |
|---|---|
| `<html><head><title>x</title></head>` | `<html><head><link rel="canonical" href="/en/"><title>x</title></head>` |
| `<HTML><HEAD lang="en"><title>x</title>` | `<HTML><HEAD lang="en"><link rel="canonical" href="/en/"><title>x</title>` |
| `<head><link rel="canonical" href="/en/"></head>` | değişmez |
| `<head><link href="/en/" rel="canonical"></head>` | değişmez |
| `<p>no head</p>` | değişmez |
| `<header>x</header>` | değişmez (`<header>`, `<head>` değildir) |

- [ ] **Step 2: Testlerin düştüğünü gör**

Run: `cd ~/Desktop/collage && go test ./pkg/collage -run TestPrefixDefault_Export && go test ./internal/build -run TestWithCanonical`
Expected: FAIL — kök hâlâ refresh sayfası; `withCanonical` tanımsız.

- [ ] **Step 3: `writeRootHome` ve `withCanonical`'ı yaz**

`writeRootHome`, `written[i]` ile anasayfa görevinin dosya yolunu bulur (koşul bugünkü `built` döngüsüyle aynı), o dosyayı okur, `withCanonical(data, "/"+defaultLocale+"/")` uygular, `resolveTarget` + `verifyNoSymlinksBeneath` ile köke yazar. Fonksiyonun doc comment'i yeni davranışı anlatır: static host'ta yönlendirme yapılandırılamaz; kökteki dosya anasayfanın kendisidir, ziyaretçi bir ara sayfa görmez, canonical tek adresi söyler. `withCanonical`: `rel="canonical"` aramasını `bytes.Contains` + küçük harfe çevrilmiş kopya üzerinde yapar; `<head` ardından boşluk ya da `>` gelen ilk eşleşmenin `>`'sinden sonra ekler.

- [ ] **Step 4: Testleri ve tüm paketi çalıştır**

Run: `cd ~/Desktop/collage && go test ./... && go vet ./...`
Expected: PASS (capture testleri kökü hâlâ builder'ın kendi dosyası sayar).

- [ ] **Step 5: Yorumları, docs'u ve CHANGELOG'u güncelle**

CHANGELOG girdisi: "**The root of a `PrefixDefault` export is the home page.** `dist/index.html` is a copy of the default locale's home, with a canonical link to `/<default>/` added when the page has none, instead of a page that refreshes to it. A visitor to the bare domain on a static host sees the site at once. The server's 301 from `/` is unchanged."

- [ ] **Step 6: Commit (branch)**

```bash
cd ~/Desktop/collage && git switch -c feat/root-home
git add -A && git commit -m "feat(build): the root of a PrefixDefault export is the home page"
```

Push, sürüm (`v0.65.0`) ve tag için Furkan'a sor.

---

### Task 2: collage-docs — yeni sürüme geç ve dokümanı düzelt

Task 1 yayımlanıp `v0.65.0` tag'i alındıktan sonra.

**Files:**
- Modify: `go.mod`, `go.sum` (`github.com/Elagoht/collage v0.65.0`)
- Modify: `main_test.go` (yeni test)
- Modify: `content/static-export.md:83` ve `:560`, `content/links-and-locales.md:94` + TR karşılıkları (`content/tr/…` aynı satırlar)
- Modify: `main.go:204-206` yorumu ("'/' sends the reader to /en/" → sunucuda 301, export'ta kökte anasayfanın kopyası)

**Interfaces:**
- Consumes: Task 1'in davranışı (`dist/index.html` = anasayfa + canonical).

- [ ] **Step 1: Failing test yaz** — `TestExportRootIsHome` (`main_test.go`): `newApp(false, 0)` + `collage.Build` geçici bir dizine; `index.html` içinde `class="home"` var, `http-equiv="refresh"` yok, `rel="canonical"` ve `/en/` var. Mevcut testlerin `cacheDir`'i nasıl ayırdığına uy.
- [ ] **Step 2:** `go test ./... -run TestExportRootIsHome` → FAIL (v0.64.0).
- [ ] **Step 3:** `go get github.com/Elagoht/collage@v0.65.0 && go mod tidy`.
- [ ] **Step 4:** `go test ./...` → PASS; `go run . -collage-build -out dist -clean` → `dist/index.html` anasayfa.
- [ ] **Step 5:** Doküman satırlarını EN ve TR'de bugünkü davranışa göre yaz: "`dist/index.html` is a copy of the `/en/` home, with a canonical link to it". `static-export.md:560` maddesi: "the `404.html` pages and the root copy of the home page". `go run ./tools/checkcontent` geçer.
- [ ] **Step 6:** `turkish-reviewer` agent'ını `static-export links-and-locales` slug'larıyla çalıştır.
- [ ] **Step 7: Commit** — `git commit -m "deps: collage v0.65.0, the export's root is the home page"`
