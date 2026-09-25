---
description: Siteyi collage export ile static dosyalara render edin: nelerin yazıldığı, nelerin neden atlandığı, dinamik path'ler ve bir static host'ta yayımlama.
reference: NewBuilder, BuildOptions, BuildReport, PrintBuildReport, PathProvider, PathInstance, SkipRecord, ErrNotStatic
---

# Static export

Page'leri request'e bağlı olmayan bir sitenin sunucuya hiç ihtiyacı yoktur.
`collage export`, dosya olabilecek her page'i `dist/` içine render eder, static
dosyalarınızı yanlarına kopyalar ve sitenin kendi `404.html` dosyasını yazar. Bu
dizini herhangi bir static host'a koyabilirsiniz.

Siteyi sunan program da budur. Aynı template'ler, data handler'lar ve plugin'lerle
render eder. İkinci bir build yoktur, sunucuyla senkron tutmanız gereken bir şey de
yoktur. Şu an okuduğunuz sayfalar da bu yolla üretildi.

```sh
collage export          # -> dist/
collage export -clean   # empty dist/ first
collage serve           # look at dist/ the way a static host would serve it
```

## Nasıl çalışır

`collage export` uygulamanızı yüklemez. Yükleyemez de, çünkü uygulamanız sizin
kodunuzdur. Bunun yerine programınızı özel bir modda çalıştırır:

```sh
go run . -collage-build -out dist        # plus -clean when you passed it
```

| Flag | Varsayılan | Anlamı |
| --- | --- | --- |
| `-out dir` | `dist` | Dosyaların yazılacağı dizin. |
| `-clean` | kapalı | Önce dizinin mevcut içeriğini siler. |

Scaffold edilen `main.go` bu sözleşmeye uyar. `-collage-build` ile çalıştığında
uygulamayı her zamanki gibi kurar, ama onu sunmak yerine collage'ın builder'ına
verir ve ne olduğunu ekrana yazar. Bu fonksiyonun collage-docs'taki hâli şöyledir:

```go
func staticBuild(app *collage.App, outDir string, clean bool) error {
	loaded, err := site.Load(content.FS)
	if err != nil {
		return err
	}
	builder, err := collage.NewBuilder(app, collage.BuildOptions{
		OutDir:       outDir,
		Clean:        clean,
		PathProvider: docPaths{loaded},
	})
	if err != nil {
		return err
	}

	report, buildErr := builder.Build(context.Background())

	collage.PrintBuildReport(os.Stdout, report, buildErr)

	return buildErr
}
```

`Build`, hata döndürdüğünde bile bir rapor döndürür. Hata veren bir page diğerlerini
durdurmaz ve bütün hatalar `report.Errors` içinde yer alır. Dönen hata, bunların
hepsinin birleşimidir. `main` bu durumda sıfırdan farklı bir kodla çıkar, böylece
bir page yazılamadığında CI job'ı da başarısız olur. Atlanan page'ler ve uyarılar
build'i başarısız kılmaz.

`main.go`'yu yeniden yazarsanız `-collage-build`, `-out` ve `-clean` flag'lerini
koruyun. Aksi hâlde `collage export` işe yarar hiçbir şey yapmaz.

## Neler yazılır

| Ne | Nereye |
| --- | --- |
| `Static()` ya da `Incremental(ttl)` bir page | `dist/<path>/index.html`; `/` için `dist/index.html` |
| Aynı page'in varsayılan olmayan bir locale'deki hâli | Locale'in prefix'i altına: `dist/tr/<path>/index.html` |
| `Static()` ya da `Incremental(ttl)` bir document | Birebir kendi path'ine: `/sitemap.xml` için `dist/sitemap.xml` |
| Aynı document'ın varsayılan olmayan bir locale'deki hâli | Locale'in prefix'i altına: `dist/tr/sitemap.xml` |
| `AtRoot` ile kurulan bir document | Her config'de prefix'siz path'ine: `dist/robots.txt` |
| [`PrefixDefault`](/docs/links-and-locales#the-url-decides-the-locale) açıkken varsayılan locale'deki bir page ya da document | Bunlar da prefix'in altına yazılır: `dist/en/<path>/index.html` ve `dist/en/sitemap.xml`. `dist/index.html` ise okuyucuyu `/en/`'e yönlendirir |
| Not-found page | `dist/404.html`, ayrıca diğer her locale için `dist/<locale>/404.html` |
| Mount edilen her static dosya sistemi | Kendi prefix'i altına: `/static/app.css` için `dist/static/app.css` |

Bir page, içinde `index.html` olan bir dizine dönüşür. Çünkü her static host,
`/about` istendiğinde bu dosyayı arar. Bir [document](/docs/documents) ise tam olarak
kendi path'ine yazılır, çünkü `/robots.txt` isteyen bir crawler'ın bir dosya alması
gerekir.

**Not-found page**, `app.RegisterNotFoundPage` ile register edilen page'dir ve render
stratejisine bakılmaz. Bu page neredeyse her zaman `Dynamic()`'tir, çünkü
cache'lemeye hiç değmez. Yine de export'a dahil edilmesi gerekir. Host'lar iç içe bir
`tr/404.html` dosyasını arayıp aramama konusunda farklı davranır, bu yüzden ikisi de
yazılır. Not-found page'i olmayan bir site için bu dosya yazılmaz. Bilinmeyen bir
URL'de ne gösterileceğine de host karar verir.

**Mount edilen asset'ler** hem orijinal adlarıyla hem de `{{asset}}`'in link verdiği
content-addressed adlarla (`app.3a3663df.css`) kopyalanır. Böylece iki tür link de
çalışır. Production'da bir CDN'den sunulan ya da kopyalanamayacak kadar büyük bir
mount bu kopyalamadan çıkabilir:

```go
app.Mount("/media/", mediaFS, collage.WithoutBuildCopy())
```

Ayrıntılar için [Static asset'ler](/docs/assets) sayfasına bakın.

## Neler atlanır

Bazı page'ler dosya olamaz. Build bunları dışarıda bırakır ve her birini nedeniyle
birlikte raporda listeler:

- **`Dynamic()` page'ler ve document'lar.** Bunlar zaten her request'te render
  edilmek için vardır.
- **Form içeren page'ler.** Render'ı `{{csrfToken}}` içeren bir page atlanır. Bir
  form'un post edeceği bir sunucuya ihtiyacı vardır ve forgery token tek bir
  okuyucuya aittir. Page bilerek `Static()` yapılmış olabilir, yani cache'lenir ve
  post ettiği action tarafından invalidate edilir. Bu durumda da export edilmez,
  sunucudan sunulur. Scaffold'daki `/features` page'i buna bir örnektir.
- **Path provider'ı olmayan bir `{param}` pattern'i.** Hangi slug'ların var olduğunu
  bir şey söylemeden `/blog/{slug}` yazılamaz. Aşağıya bakın.
- **Aynı dosyaya düşen iki document.** Aynı path'i iki kez döndüren bir
  `DocumentPathProvider`, iki görevi tek bir dosyaya bağlar. İlki yazılır, geri
  kalanlar `collage.ErrDuplicateOutputPath` ile atlanır. İki locale'deki tek bir
  pattern bu duruma girmez, çünkü her locale kendi prefix'i altına yazılır. Bkz.
  [Document'lar](/docs/documents#documents-in-a-static-export).

Path olmadan register edilen error page'ler hiç listelenmez, çünkü bunlar birer URL
değildir.

Bazı şeyler page olmadığı için ne export edilir ne de raporda görünür:
[action'lar](/docs/forms-and-actions), `app.Handle` ile mount edilen handler'lar ve
middleware'ler. Export request olmadan render eder. Bu yüzden hiçbir middleware
çalışmaz ve `collage.Vary` hiç çağrılmaz. Her page, hiçbir tercihi olmayan bir
request'in alacağı hâliyle yazılır.

## Neler için uyarı verilir

`WithCacheParams` ile hangi query parametrelerini okuduğunu bildiren bir page, her
parametre değeri için farklı render edilir. Oysa bir dosyanın query string'i yoktur.
Static host, `/blog?page=2` isteğine `/blog` dosyasıyla cevap verir. Page query
olmadan yazılır ve rapor bunu açıkça belirtir. Böylece sayfalanmış bir arşiv
çalışıyormuş gibi görünmez. `WithCacheParams` kullanan bir document için de, örneğin
sayfalanmış bir feed için, aynı şekilde uyarı verilir. Bu v0.10.0'dan beri böyledir;
öncesinde yalnızca page'ler için uyarı veriliyordu.

Sayfalamanın export'ta da çalışması gerekiyorsa sayfa numarasını path'e koyun,
örneğin `/blog/page/{n}`. Ardından bu page'leri bir path provider ile listeleyin.

## Neler başarısız olur

Aşağıdakiler hatadır. Page yazılmaz, build bunu raporlar ve sıfırdan farklı bir
kodla çıkar.

- **Degraded bir render.** Bir fragment'i hata veren page,
  `collage.ErrDegradedRender` ile reddedilir ve hata o fragment'in adını içerir.
  Sunucudan sunulan bir page'de fragment hata verirse page gösterilir ama hiçbir
  zaman cache'lenmez. Bir dosyanın ise kendini toparlayacağı bir TTL'i yoktur, yani
  o hatayı bir sonraki export'a kadar taşır. Sidebar'ı eksik bir page, hiç page
  olmamasından iyiyse `BuildOptions.AllowDegraded` ayarını açın.
- **Boş bir render** (`collage.ErrEmptyRender`) ve boş bir document
  (`collage.ErrEmptyDocumentBody`). Sıfır byte'lık bir `index.html` hiçbir zaman
  yazılmaz.
- **Bir page'de oluşan panic** (`collage.ErrBuildPanic`). Panic recover edilir ve o
  page'e ait hata olarak kaydedilir. Build'in geri kalanı devam eder.
- **Form içeren bir not-found page** (`collage.ErrUnresolvedToken`). Bunun nedeni,
  static host'un bu page'e bir dosya olarak ihtiyaç duymasıdır.
- **Aynı output path'e düşen iki page** (`collage.ErrOutputPathCollision`). Bu,
  yalnızca sondaki slash ile ayrılan iki pattern ya da aynı path'i iki kez döndüren
  bir path provider olabilir. Bu durum tespit edildiğinde hiçbir page render edilmez.
  Document'lar, `404.html` page'leri ve mount edilen asset'ler yine yazılır, ama build
  yine de başarısız olur.

## Dinamik path'ler: `PathProvider`

`/blog/{slug}` adresindeki bir page, birçok URL'si olan tek bir page'dir. Builder bu
URL'leri bir `collage.PathProvider`'dan ister:

```go
type PathProvider interface {
	Paths(ctx context.Context, page *collage.Page, locale string) ([]collage.PathInstance, error)
}
```

Her dinamik page ve locale için bir kez çağrılır. Somut path'leri ve her birinin
yakaladığı parametreleri döndürür. Bu dokümantasyon, `/docs/{slug}` adresindeki tek
bir page'dir: `doc`. Provider, dokümantasyonun bütün sayfalarını listeler:

```go
// docPaths tells the static build which /docs/{slug} pages exist: every page of
// the documentation, and nothing else.
type docPaths struct{ site *site.Site }

func (d docPaths) Paths(_ context.Context, page *collage.Page, _ string) ([]collage.PathInstance, error) {
	if page.Name != "doc" {
		return nil, nil
	}
	var paths []collage.PathInstance
	for _, p := range d.site.Pages() {
		paths = append(paths, collage.PathInstance{Path: p.URL(), Params: map[string]string{"slug": p.Slug}})
	}
	return paths, nil
}
```

- **Page'i kontrol edin.** Bütün dinamik page'lere tek bir provider cevap verir.
  Tanımadığınız bir page için `nil` döndürürseniz o page için hiçbir şey yazılmaz.
  Bu bir hata sayılmaz.
- **Data handler'lar `Params`'ı görür.** `Params`, router'ın `Path`'ten yakaladığı
  değerlerin üzerine yazılır. Böylece `rc.Param("slug")`, canlı bir request'teki
  değerin aynısını döner.
- **`Path` locale prefix'i içermez.** Pattern'in path'ini, yani `/blog/hello`'yu
  döndürün. Varsayılan olmayan bir locale'i builder kendi dizini altına yazar.
- **Hata döndürürseniz** hata o page ve locale için kaydedilir ve build devam eder.

`{param}` içeren document'ların kendi interface'i vardır:
`BuildOptions.DocumentPathProvider`. Bu interface [Document'lar](/docs/documents)
sayfasında anlatılır.

Bir provider'ın döndürdüğü her path, dosyası yazılmadan önce kontrol edilir. Output
dizininin dışına çıkan bir path (`/../../etc`) `collage.ErrPathEscapesOutDir` ile
reddedilir. Dizinün dışına götüren bir symlink üzerinden yazma da aynı şekilde
reddedilir. Bu ret yalnızca o path'i başarısız kılar, build'in geri kalanını
etkilemez. Diğer page'ler yine render edilip yazılır ve hata raporda yer alır.

## Build seçenekleri

| Alan | Anlamı |
| --- | --- |
| `OutDir` | Dosyaların yazılacağı yer. Zorunludur. |
| `Clean` | Önce `OutDir`'in içeriğini siler (dizinin kendisini değil). |
| `Locales` | Yalnızca bu locale'leri build eder. Boş bırakılırsa page'lerin tanımladığı bütün locale'ler build edilir. |
| `Concurrency` | Aynı anda kaç page'in render edilip yazılacağı. `0` ya da `1` verilirse page'ler tek tek işlenir. Rapor her iki durumda da aynı sıradadır. |
| `PathProvider` | `{param}` içeren page'ler için somut path'ler. |
| `DocumentPathProvider` | `{param}` içeren document'lar için somut path'ler. |
| `AllowDegraded` | Render sırasında bir fragment'i hata veren page'leri de yazar. |

Builder, dosya sisteminin köküne çıkan bir `OutDir`'i reddeder. Bir repository'nin
kökü olan bir `OutDir`'i `Clean` etmeyi de reddeder (`collage.ErrDangerousOutDir`).
Aksi hâlde `-out .` ile `-clean` birlikte verildiğinde projeniz silinirdi.

### Export'ta plugin'ler

Export, sunucuyla aynı durumda render eder. Önce her plugin'in `Init`'i çalışır,
böylece plugin'ler aynı config'i okur. `OnBeforeRender`, `OnAfterRender` ve
`OnDocumentRendered` her page ve document için tetiklenir. Yani bir minifier ya da
structured data plugin'i sunulan bir page'e ne yapıyorsa dosyaya da onu yapar.
`OnPageResolved` ise tetiklenmez, çünkü export bir request değildir. Bkz.
[Plugin kullanmak](/docs/plugins).

## Raporu okumak

`collage.PrintBuildReport`, build'in ne yaptığını yazdırır. `collage new` ile scaffold
edilen projede çıktı şöyle görünür:

```sh
✓ 8 files written
    dist/404.html
    dist/index.html
    dist/static/app.3a3663df973f06fb.js
    dist/static/app.7050c2518057f5c5.css
    dist/static/app.css
    dist/static/app.js
    dist/static/favicon.936907e03c8c09ad.svg
    dist/static/favicon.svg

▲ 3 skipped
    hello  page uses the dynamic render strategy, which cannot be built statically
    features (en)  page carries {{csrfToken}}; a form needs a server to submit to, so it is served rather than exported
    health  document uses the dynamic render strategy, which cannot be built statically

8 written · 3 skipped · 0 failed · 2.5ms
```

Yazılan dosyalar on taneyi geçince özetlenir. Üç yüz dosya yazan bir build, atladığı
tek page'i bu listenin altına gömmemelidir. Atlananlar, uyarılar ve hatalar hiçbir
zaman kısaltılmaz. Son satırda bütün sayılar yer alır ve satır, aralarındaki en kötü
duruma göre renklendirilir. v0.10.0'dan beri atlananlar route olarak sayılır, yani
page'ler ve document'lar birlikte sayılır. Bu yüzden çıktıda `3 pages skipped` değil,
`3 skipped` yazar.

Raporu kodda kullanmak için `report.Skipped`, `report.Warnings` ve `report.Errors`
alanlarını kendiniz okuyun. Her atlama bir `collage.SkipRecord`'dur. Bu kayıtta
route'un adı (`Page`), `Locale`'i ve okuyan kişi için bir `Reason` bulunur.
v0.10.0'dan beri `errors.Is` ile eşleştirebileceğiniz bir `Err` de bulunur. Bkz.
[Hatalar](/docs/errors#static-builds).
[Test yazmak](/docs/testing#testing-the-export) sayfası bunu bir teste dönüştürür.

Renkler ve `✓ ▲ ✗` işaretleri yalnızca terminalde görünür. `NO_COLOR` tanımlıysa
terminalde de görünmez. CI log'larında bu işaretler düz ASCII olarak yazılır
(`+ ! x`).

## Göz atmak: `collage serve`

`dist/index.html`'i tarayıcıda doğrudan açmak işe yaramaz. Bir `file://` sayfasının
kökü yoktur, bu yüzden bütün mutlak link'ler ve stylesheet'ler kırılır.
`collage serve`, export çıktısını bir static host'un sunduğu şekilde sunar:

```sh
collage serve                  # http://localhost:4000
collage serve -dir public -port 8000
```

- `/about` isteğine `about/index.html` ile cevap verilir.
- `index.html` içermeyen bir dizin 404 döner. Dizin listelemesi yapılmaz.
- Bilinmeyen bir path, 404 status'uyla `404.html` döner.
- Hiçbir şey cache'lenmez. Yeniden export alıp sayfayı yenilediğinizde yeni çıktıyı
  görürsünüz.

Flag'leri `-dir` (varsayılan `dist`), `-host` (varsayılan `localhost`) ve `-port`'tur.
`-port`'un varsayılanı `4000`'dir, 3000 değildir. Böylece ikisini karşılaştırırken
`collage dev` ile yan yana çalışabilir. `collage serve` yalnızca dosya sunar,
projenizi çalıştırmaz.

## Hosting

Çıktı, mutlak link'ler içeren düz dosyalardan oluşur. Bu yüzden herhangi bir static
host onu sunabilir. Hangi host'u kullanırsanız kullanın üç şeyi kontrol edin:

- **Site, domain'in kökünde olmalıdır.** Link'ler ve asset URL'leri `/` ile başlar ve
  collage'ın bir base path ayarı yoktur. Bu yüzden bir alt path altında yayımlanan
  bir sitenin link'leri kırılır. `user.github.io/project/` adresindeki bir GitHub
  project page'i buna örnektir. Custom bir domain kullanın ya da siteye kendi
  domain'ini veren bir host seçin.
- **`404.html` köktedir.** Çoğu host bu dosyayı hiçbir config gerekmeden bu adla
  bulur.
- **[`TrailingSlash`](/docs/configuration#trailingslash) açık olmalıdır.** Bir page
  `<path>/index.html` olarak yazılır. Host onu `/about/` adresinde sunar ve
  `/about`'u oraya redirect eder. Bu ayar açıkken collage'ın ürettiği her link
  doğrudan host'un cevap verdiği adres olur, oraya giden bir redirect olmaz.

### GitHub Pages

Bu site, testleri çalıştıran, export alan ve `dist/`'i yükleyen bir workflow ile
yayımlanır:

```yaml
name: pages

on:
  push:
    branches: [main]

permissions:
  contents: read
  pages: write
  id-token: write

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.26'
      - run: go test ./...
      - run: go run . -collage-build -out dist -clean
      - uses: actions/configure-pages@v5
      - uses: actions/upload-pages-artifact@v3
        with:
          path: dist

  deploy:
    needs: build
    runs-on: ubuntu-latest
    environment:
      name: github-pages
      url: ${{ steps.deployment.outputs.page_url }}
    steps:
      - id: deployment
        uses: actions/deploy-pages@v4
```

Export adımı programı `collage export` üzerinden değil, doğrudan çalıştırır. Böylece
runner'a collage CLI'ı kurmanız gerekmez. Repository'nin Pages kaynağını GitHub
Actions olarak ayarlayın. `user.github.io` repository'niz değilse ona custom bir
domain de verin.

### Cloudflare Pages

CI'da export'u aynı şekilde alın ve dizini Wrangler ile yükleyin:

```sh
go run . -collage-build -out dist -clean
npx wrangler pages deploy dist --project-name mysite
```

Kökte bir `404.html` varsa Cloudflare Pages bilinmeyen path'ler için onu sunar.
Sitenin bir not-found page'i varsa export bu dosyayı her zaman yazar.

### Diğerleri

Netlify, CloudFront arkasındaki S3 ya da bir nginx dizini fark etmez. Hepsinin
ihtiyacı yalnızca `dist/`'in içeriğidir. Host bunu kendiliğinden yapmıyorsa
`404.html`'i error page olarak ayarlamanız da gerekir. Site form'lara, preview'lara
ya da her request'te render edilen page'lere ihtiyaç duyuyorsa bir sunucuya ihtiyacı
vardır. Bkz. [Deployment](/docs/deployment).
