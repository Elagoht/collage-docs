---
description: Siteyi collage export ile statik dosyalara render edin — neler yazılır, neler neden atlanır, dinamik yollar ve statik barındırma hizmetinde yayımlama.
---

# Statik dışa aktarma

Sayfaları gelen isteğe göre değişmeyen bir sitenin sunucuya hiç ihtiyacı yoktur.
`collage export`, dosya olabilecek her sayfayı `dist/` içine render eder, statik
dosyalarınızı yanlarına kopyalar ve sitenin kendi `404.html`'ini yazar. Dizini
herhangi bir statik barındırma hizmetine koyun.

Siteyi sunan programın ta kendisidir; aynı şablonlar, data handler'lar ve plugin'ler
üzerinden render eder. İkinci bir build yoktur, sunucuyla uyumlu tutulması gereken
bir şey de yoktur. Okuduğunuz sayfalar bu yolla üretildi.

```sh
collage export          # -> dist/
collage export -clean   # empty dist/ first
collage serve           # look at dist/ the way a static host would serve it
```

## Nasıl çalışır

`collage export` uygulamanızı yüklemez — yükleyemez de, çünkü uygulamanız sizin
kodunuzdur. Bunun yerine programınızı özel bir modda çalıştırır:

```sh
go run . -collage-build -out dist        # plus -clean when you passed it
```

| Flag | Varsayılan | Anlamı |
| --- | --- | --- |
| `-out dir` | `dist` | Dosyaların yazılacağı yer. |
| `-clean` | kapalı | Önce dizinin mevcut içeriğini kaldırır. |

İskeleti oluşturulan `main.go` bu sözleşmeye uyar: `-collage-build` verildiğinde
uygulamayı her zamanki gibi kurar ve onu sunmak yerine collage'ın builder'ına
verir, ardından ne olduğunu yazdırır. Bu fonksiyonun collage-docs sürümü şudur:

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

`Build`, bir hata döndürdüğünde bile bir rapor döndürür: hata veren bir sayfa
diğerlerini durdurmaz ve her hata `report.Errors` içindedir. Dönen hata bunların
hepsinin birleşimidir ve `main` onunla sıfırdan farklı bir kodla çıkar; böylece bir
sayfa yazılamadığında CI işi başarısız olur. Atlanan sayfalar ve uyarılar build'i
başarısız kılmaz.

`main.go`'yu yeniden yazarsanız `-collage-build`, `-out` ve `-clean` flag'lerini
koruyun; yoksa `collage export` işe yarar bir şey yapmaz olur.

## Neler yazılır

| Ne | Nereye |
| --- | --- |
| `Static()` ya da `Incremental(ttl)` bir sayfa | `dist/<path>/index.html`; `/` için `dist/index.html` |
| Aynı sayfa, varsayılan olmayan bir locale'de | Locale'in önekinin altına: `dist/tr/<path>/index.html` |
| `Static()` ya da `Incremental(ttl)` bir document | Birebir kendi yoluna: `/sitemap.xml` için `dist/sitemap.xml` |
| Aynı document, varsayılan olmayan bir locale'de | Locale'in önekinin altına: `dist/tr/sitemap.xml` |
| Bulunamadı sayfası | `dist/404.html` ve diğer her locale için `dist/<locale>/404.html` |
| Mount edilmiş her statik dosya sistemi | Kendi önekinin altına: `/static/app.css` için `dist/static/app.css` |

Bir sayfa, içinde `index.html` olan bir dizin olur; her statik barındırma hizmeti
`/about` istendiğinde bunu arar. Bir [document](/docs/documents) ise tam olarak kendi
yoluna yazılır, çünkü `/robots.txt` isteyen bir tarayıcı botu bir dosya almalıdır.

**Bulunamadı sayfası** `app.RegisterNotFoundPage`'den gelir ve render stratejisine
bakılmaz — neredeyse her zaman `Dynamic()`'tir, çünkü önbelleğe almaya asla değmez;
yine de dışa aktarmada yeri vardır. Barındırma hizmetleri iç içe bir `tr/404.html`
arayıp aramamakta farklılaşır, bu yüzden ikisi de yazılır. Bulunamadı sayfası olmayan
bir site dosya almaz ve bilinmeyen bir URL, barındırma hizmeti ne gösterirse onu
gösterir.

**Mount edilmiş asset'ler** hem özgün adlarıyla hem de `{{asset}}`'in bağlandığı
içerik adresli adlarla (`app.3a3663df.css`) kopyalanır; böylece iki tür bağlantı da
çalışır. Production'da bir CDN'den sunulan ya da çoğaltılamayacak kadar büyük bir
mount bunun dışında kalmayı seçebilir:

```go
app.Mount("/media/", mediaFS, collage.WithoutBuildCopy())
```

Bkz. [Statik asset'ler](/docs/assets).

## Neler atlanır

Bazı sayfalar dosya olamaz. Build onları dışarıda bırakır ve her birini raporda,
nedeniyle birlikte adlandırır:

- **`Dynamic()` sayfalar ve document'lar.** İstek başına render edilmek için
  vardırlar.
- **Form içeren sayfalar.** Render'ı `{{csrfToken}}` içeren bir sayfa atlanır: bir
  formun gönderileceği bir sunucuya ihtiyacı vardır ve sahtecilik token'ı tek bir
  okuyucuya aittir. Sayfa bilerek `Static()` olabilir — önbelleğe alınmış ve
  gönderildiği action tarafından geçersiz kılınan — ve dışa aktarılmak yerine
  sunulur. İskeletteki `/features` sayfası böyledir.
- **Path provider'ı olmayan bir `{param}` pattern'i.** `/blog/{slug}`, bir şey
  hangi slug'ların var olduğunu söyleyene kadar yazılamaz. Aşağıya bakın.
- **Tek bir dosyaya düşen iki document.** Bir yolu iki kez döndüren bir
  `DocumentPathProvider`, iki görevi tek bir dosyaya çözer; ilki yazılır, gerisi
  `collage.ErrDuplicateOutputPath` ile atlanır. İki locale'deki tek bir pattern bu
  durum değildir — her locale kendi önekinin altına yazılır. Bkz.
  [Document'lar](/docs/documents#documents-in-a-static-export).

Yol olmadan kaydedilmiş hata sayfaları hiç listelenmez: URL değildirler.

Sayfa olmadıkları için dışa aktarılmayan ve raporlanmayanlar:
[action'lar](/docs/forms-and-actions), `app.Handle` ile mount edilen handler'lar
ve middleware. Dışa aktarma istek olmadan render eder; bu yüzden hiçbir middleware
çalışmaz ve `collage.Vary` hiç çağrılmaz — her sayfa, hiçbir tercihi olmayan bir
isteğin alacağı sürümüyle yazılır.

## Neler için uyarılır

`WithCacheParams` ile hangi query parametrelerini okuduğunu bildiren bir sayfa, her
biri için farklı render edilir. Bir dosyanın query string'i yoktur: statik
barındırma hizmeti `/blog?page=2`'ye `/blog` dosyasıyla yanıt verir. Sayfa query
olmadan yazılır ve sayfalanmış bir arşivin çalışıyormuş gibi görünmesine izin
vermek yerine rapor bunu söyler. `WithCacheParams` içeren bir document — sayfalanmış
bir feed — için de aynı şekilde uyarılır (v0.10.0'dan itibaren; öncesinde yalnızca
sayfalar için uyarılıyordu).

Sayfalamanın dışa aktarmada çalışması gerekiyorsa sayfa numarasını yola koyun —
`/blog/page/{n}` — ve sayfaları bir path provider ile listeleyin.

## Neler başarısız olur

Bunlar hatadır: sayfa yazılmaz, build bunu raporlar ve sıfırdan farklı bir kodla
çıkar.

- **Kusurlu bir render.** Bir fragment'in hata verdiği bir sayfa, fragment'i
  adlandıran `collage.ErrDegradedRender` ile reddedilir. Hata veren fragment'i olan
  sunulmuş bir sayfa gösterilir ama asla önbelleğe alınmaz; bir dosyanın ise
  toparlanabileceği bir TTL'i yoktur, bu yüzden hatayı bir sonraki dışa aktarmaya
  kadar taşırdı. Eksik bir kenar çubuğu olan bir sayfa hiç sayfa olmamasından iyiyse
  `BuildOptions.AllowDegraded`'ı ayarlayın.
- **Boş bir render**, `collage.ErrEmptyRender`, ve boş bir document,
  `collage.ErrEmptyDocumentBody`. Sıfır baytlık bir `index.html` asla yazılmaz.
- **Bir sayfadaki panic**, `collage.ErrBuildPanic`. Yakalanır ve o sayfanın hanesine
  kaydedilir; build'in geri kalanı devam eder.
- **Form içeren bir bulunamadı sayfası**, `collage.ErrUnresolvedToken`, çünkü statik
  barındırma hizmetinin o sayfaya bir dosya olarak ihtiyacı vardır.
- **Tek bir çıktı yoluna düşen iki sayfa**, `collage.ErrOutputPathCollision` —
  yalnızca sondaki eğik çizgiyle ayrılan iki pattern ya da bir yolu iki kez döndüren
  bir path provider. Bu bulunduğunda hiçbir sayfa render edilmez; document'lar,
  `404.html` sayfaları ve mount edilmiş asset'ler yine yazılır ve build yine başarısız
  olur.

## Dinamik yollar: `PathProvider`

`/blog/{slug}`'daki bir sayfa, birçok URL'si olan tek bir sayfadır. Builder bunları
bir `collage.PathProvider`'a sorar:

```go
type PathProvider interface {
	Paths(ctx context.Context, page *collage.Page, locale string) ([]collage.PathInstance, error)
}
```

Her dinamik sayfa ve locale için bir kez çağrılır; somut yolları ve her birinin
yakaladığı parametreleri döndürür. Bu belgeler, `/docs/{slug}` üzerindeki tek bir
sayfadır, `doc`; provider belgelerin her sayfasını listeler:

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

- **Sayfayı kontrol edin.** Tek bir provider her dinamik sayfaya yanıt verir.
  Tanımadığınız bir sayfa için `nil` döndürmek onun için hiçbir şey yazmaz; bu bir
  hata değildir.
- **`Params`, data handler'ların gördüğüdür.** Router'ın `Path`'ten yakaladığının
  üzerine bindirilir; böylece `rc.Param("slug")` canlı bir isteğin alacağı değere
  sahip olur.
- **`Path` locale öneki olmadan verilir.** Pattern'in yolunu, `/blog/hello`'yu
  döndürün; builder varsayılan olmayan bir locale'i kendi dizininin altına yazar.
- **Bir hata döndürmek** onu o sayfa ve locale'in hanesine kaydeder ve devam eder.

`{param}` içeren document'ların kendi interface'i vardır,
`BuildOptions.DocumentPathProvider`; [Document'lar](/docs/documents) sayfasında
anlatılır.

Bir provider'ın döndürdüğü her yol, kendi dosyası yazılmadan önce kontrol edilir:
çıktı dizininin dışına çözülecek bir yol — `/../../etc` —
`collage.ErrPathEscapesOutDir` ile reddedilir; dizinin dışına çıkan bir symlink
üzerinden yazma da öyle. Ret yalnızca o yolu başarısız kılar, etrafındaki build'i
değil: diğer sayfalar yine render edilip yazılır ve hata raporda yer alır.

## Build seçenekleri

| Alan | Anlamı |
| --- | --- |
| `OutDir` | Nereye yazılacağı. Zorunlu. |
| `Clean` | Önce `OutDir`'in içeriğini (dizinin kendisini değil) kaldırır. |
| `Locales` | Yalnızca bu locale'leri derler. Boşsa bir sayfanın bildirdiği her locale derlenir. |
| `Concurrency` | Aynı anda kaç sayfanın render edilip yazılacağı. `0` ya da `1` birer birer demektir. Rapor her iki durumda da aynı sıradadır. |
| `PathProvider` | `{param}` içeren sayfalar için somut yollar. |
| `DocumentPathProvider` | `{param}` içeren document'lar için somut yollar. |
| `AllowDegraded` | Render'ında hata veren bir fragment olan sayfaları yazar. |

Builder, dosya sisteminin köküne çözülen bir `OutDir`'i reddeder ve bir deponun
kökü olan bir `OutDir`'i `Clean` etmeyi reddeder (`collage.ErrDangerousOutDir`) —
aksi hâlde `-clean` ile `-out .` projenizi silerdi.

### Dışa aktarmada plugin'ler

Dışa aktarma, sunucunun render ettiği durumda render eder. Önce her plugin'in `Init`'i
çalışır, böylece plugin aynı yapılandırmayı okur; `OnBeforeRender`, `OnAfterRender`
ve `OnDocumentRendered` her sayfa ve document için tetiklenir, dolayısıyla bir
minifier ya da structured data plugin'i sunulan bir sayfaya ne yapıyorsa dosyaya da
onu yapar. `OnPageResolved` tetiklenmez, çünkü dışa aktarma bir istek değildir. Bkz.
[Plugin kullanmak](/docs/plugins).

## Raporu okumak

`collage.PrintBuildReport` build'in ne yaptığını yazdırır. `collage new`'un iskeletini
oluşturduğu proje için şöyle görünür:

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

Yazılan dosyalar ondan sonra özetlenir, çünkü üç yüz dosya yazan bir build atladığı
tek sayfayı gömmemelidir. Atlananlar, uyarılar ve hatalar asla kısaltılmaz. Son
satırda bütün sayılar vardır ve aralarındaki en kötüsüne göre renklendirilir.
v0.10.0'dan itibaren atlananlar route olarak, sayfalar ve document'lar birlikte
sayılır — `3 pages skipped` değil, `3 skipped`.

Rapora göre kodda işlem yapmak için `report.Skipped`, `report.Warnings` ve
`report.Errors`'u kendiniz okuyun. Her atlama, route'un adını (`Page`), `Locale`'ini,
insanlar için bir `Reason`'ı ve v0.10.0'dan itibaren `errors.Is` ile eşleştirilecek
bir `Err`'i içeren bir `collage.SkipRecord`'dur — bkz.
[Hatalar](/docs/errors#static-builds).
[Test](/docs/testing#testing-the-export) bunu bir teste dönüştürür.

Renk ve `✓ ▲ ✗` işaretleri yalnızca bir terminalde görünür, `NO_COLOR` ayarlıyken de
görünmez. Bir CI logunda işaretler düz ASCII'dir (`+ ! x`).

## Göz atmak: `collage serve`

`dist/index.html`'i bir tarayıcıda açmak işe yaramaz: bir `file://` sayfasının kökü
yoktur, bu yüzden her mutlak bağlantı ve stil dosyası bozuk olur. `collage serve`
dışa aktarılan siteyi bir statik barındırma hizmetinin sunduğu gibi sunar:

```sh
collage serve                  # http://localhost:4000
collage serve -dir public -port 8000
```

- `/about` isteğine `about/index.html` ile yanıt verilir.
- `index.html` içermeyen bir dizin 404'tür — dizin listelemesi yoktur.
- Bilinmeyen bir yol, 404 durumuyla `404.html`'i alır.
- Hiçbir şey önbelleğe alınmaz; yeniden dışa aktarıp sayfayı yenilemek yeni çıktıyı
  gösterir.

Flag'leri `-dir` (varsayılan `dist`), `-host` (varsayılan `localhost`) ve `-port`
(varsayılan `4000` — 3000 değil, böylece karşılaştırırken `collage dev`'in yanında
çalışabilir). Dosya sunar; projenizi çalıştırmaz.

## Barındırma

Çıktı, mutlak bağlantılara sahip düz dosyalardır; dolayısıyla herhangi bir statik
barındırma hizmeti onu sunar. Hangisinde olursa olsun kontrol edilecek üç şey var:

- **Site, alan adının kökünde olmalıdır.** Bağlantılar ve asset URL'leri `/` ile
  başlar ve collage'ın bir base-path ayarı yoktur; dolayısıyla bir alt yol altında
  yayımlanan bir site — `user.github.io/project/` adresindeki bir GitHub proje sayfası
  gibi — bozuk bağlantılara sahip olur. Özel bir alan adı ya da siteye kendi alan
  adını veren bir barındırma hizmeti kullanın.
- **`404.html` köktedir.** Çoğu barındırma hizmeti onu hiçbir yapılandırma
  gerekmeden bu adla bulur.
- **[`TrailingSlash`](/docs/configuration#trailingslash) açıktır.** Bir sayfa
  `<path>/index.html` olarak yazılır; barındırma hizmeti onu `/about/` adresinde
  sunar ve `/about`'u oraya yönlendirir. Ayar açıkken collage'ın kurduğu her
  bağlantı, oraya giden bir yönlendirme değil, doğrudan hizmetin yanıt verdiği
  adres olur.

### GitHub Pages

Bu site, testleri çalıştıran, dışa aktaran ve `dist/`'i yükleyen bir workflow ile
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

Dışa aktarma adımı programı `collage export` üzerinden değil doğrudan çalıştırır;
böylece runner'da collage CLI'ının kurulu olması gerekmez. Deponun Pages kaynağını
GitHub Actions olarak ayarlayın ve `user.github.io` deponuz değilse ona özel bir alan
adı verin.

### Cloudflare Pages

CI'da aynı şekilde dışa aktarın ve dizini Wrangler ile yükleyin:

```sh
go run . -collage-build -out dist -clean
npx wrangler pages deploy dist --project-name mysite
```

Cloudflare Pages, kökte bir `404.html` varsa bilinmeyen yollar için onu sunar;
sitenin bir bulunamadı sayfası olduğunda dışa aktarma bu dosyayı her zaman yazar.

### Diğerleri

Netlify, CloudFront arkasında S3, bir nginx dizini — her birinin ihtiyacı yalnızca
`dist/`'in içeriği ve, zaten yapmıyorsa, hata sayfası olarak yapılandırılmış
`404.html`'dir. Site formlara, önizlemelere ya da istek başına sayfalara ihtiyaç
duyduğunda ise bir sunucuya ihtiyaç duyar: bkz. [Yayına alma](/docs/deployment).
