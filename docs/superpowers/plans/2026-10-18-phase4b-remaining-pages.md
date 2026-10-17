# Aşama 4b–4e: Kalan sayfalara şema, kod ve senaryo Uygulama Planı

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Kalan 20 doküman sayfası (EN + TR), Aşama 4a'daki beş sayfa gibi şema, kod ve gerçek dünya örnekleriyle daha kolay okunur hale gelir; hiçbir bilgi kaybolmaz.

**Architecture:** Aşama 4a'nın yöntemi aynen: her sayfa yalnızca kendi iki dosyasına dokunan bağımsız bir görev; beşerli gruplar paralel yürür. Her grup: kaynağa göre doğrulama → Türkçe kontrol → tarayıcıda şema kontrolü → son inceleme → yayın.

**Spec:** `docs/superpowers/specs/2026-10-17-docs-positioning-design.md` (Kararlar → 6)

## Global Constraints

`docs/superpowers/plans/2026-10-17-phase4a-priority-pages.md` içindeki Global Constraints aynen geçerlidir. Ek olarak:

- Aşama 4a'dan öğrenilenler: bir şema, kaynakta olmayan bir yolu göstermez (ör. sayfalara GET/HEAD dışında istek gitmez; fragment path ve action cache'lenmez; mount'lar ve `app.Handle` router'dan önce eşleşir; plugin middleware'i `app.Use`'unkilerin dışındadır); zaman çizelgeleri olayları gerçek sırasıyla çizer.
- Örnek bölüm başlıkları "Example: …" / "Örnek: …" biçimindedir.
- Referans sayfalarında (`cli`, `configuration`, `template-functions`, `errors`) "senaryo", bir görev odaklı tarif olabilir (ör. "bir siteyi ilk kez deploy etmek için hangi komutlar"); tablolar korunur.

## Gruplar

- **B:** your-first-page, pages-and-layouts, data-handlers, templates, links-and-locales
- **C:** head-and-seo, assets, documents, previews, static-export
- **D:** deployment, security, testing, writing-plugins, plugin-authors
- **E:** installation, cli, configuration, template-functions, errors

## Her görevin adımları

- [ ] Sayfayı ve ilgili kaynağı oku.
- [ ] EN: en az bir şema, bir örnek, kısa paragraflar; her yeni iddia kaynağa göre doğrulanır, örnek kod v0.65.0'a karşı derlenir.
- [ ] TR: aynı yapı.
- [ ] `go run ./tools/checkcontent`, `go test ./...`, export → temiz.
