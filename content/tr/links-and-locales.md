---
description: Bir siteyi birkaç dilde sunmak ve page'lere adlarıyla link vermek; böylece link'ler page'i her locale'de takip eder.
reference: LocaleConfig, PageBuilder.WithPath, PageBuilder.WithFragmentPath, Vary, ErrNoPathInLocale, ErrUnknownRoute, ErrUnknownFragmentPath, ErrAmbiguousFragmentPath
---

# Link'ler ve locale'ler

Bir collage sitesi her page'i birkaç dilde, her dili kendi URL'sinde sunabilir.
collage locale'leri route eder, çeviri yapmaz. Bir kelimenin Türkçe karşılığı
içeriğinizde durur ve içeriğiniz zaten nerede tutuluyorsa orada kalır. collage'ın
işi, hangi URL'nin hangi dile ait olduğunu bilmek ve bunu bilen link'ler üretmektir.

## Locale'leri yapılandırmak

```go
app, err := collage.New(&collage.Config{
	Template: collage.TemplateConfig{Root: "templates"},
	Locale: collage.LocaleConfig{
		Default:   "en",
		Supported: []string{"en", "tr"},
	},
})
```

`Default`, locale prefix'i olmayan bir URL'nin locale'idir ve varsayılan değeri
`"en"`'dir. `Supported`, sitenin sunduğu bütün locale'leri listeler ve `Default`'u
içermek zorundadır (içermezse `collage.ErrLocaleDefaultNotSupported` döner). Boş
bırakılırsa yalnızca `Default`'tan oluşur.

Ardından her page, bulunduğu her locale için kendi path'ini tanımlar:

```go
collage.NewPage("about").
	WithLayout(layout).
	WithContent(about).
	WithPath("en", "/about").
	WithPath("tr", "/hakkinda").
	Build()
```

Path'ler buradaki gibi farklı olabilir ya da her locale'de aynı pattern olabilir.
Yalnızca tek bir locale'de path'i olan bir page yalnızca o locale'de vardır.

## Locale'i URL belirler

Locale prefix'i olmayan bir path `Default` locale'e aittir. İlk segment'i desteklenen
bir locale'in adı olan bir path o locale'e aittir ve path eşleştirilmeden önce bu
prefix çıkarılır:

| URL | Locale | Eşleştirildiği yer |
| --- | --- | --- |
| `/about` | `en` | `en` path'leri içinde `/about` |
| `/tr/hakkinda` | `tr` | `tr` path'leri içinde `/hakkinda` |
| `/tr/about` | `tr` | `tr` path'leri içinde `/about`; bulunamaz, 404 döner |
| `/fr/about` | `en` | `/fr/about`; `fr` desteklenmediği için sıradan bir segment'tir |
| `/en/about` | — | `/about`'a redirect edilir |
| `/TR/hakkinda` | — | `/tr/hakkinda`'ya redirect edilir |

Her page'in locale başına tek bir URL'si vardır. Bu yüzden bir prefix'in diğer
yazılışları kalıcı olarak o URL'ye redirect edilir. Bu yazılışlar iki tanedir:
varsayılan locale'in kendi prefix'i (varsayılan locale'in URL'leri bu prefix'i
taşımaz) ve desteklenen bir locale'in farklı büyük/küçük harfle yazılışı. Redirect, `GET` ve
`HEAD` için `301`, diğer bütün method'lar için `308`'dir. Böylece yanlış yazılışa
post edilen bir form `GET`'e dönüşmez, yeniden post edilir. Query string de redirect
ile birlikte taşınır. Bu redirect olmasaydı `/en/about`, arama motoru için
`/about`'un ikinci bir kopyası, cache'te de ikinci bir kayıt olurdu.

Varsayılan locale dahil bütün dilleri bir prefix altında sunmak isteyen bir site
`PrefixDefault: true` ayarını kullanır (v0.14.0'dan beri). Bu durumda `/en/about`,
`/tr/hakkinda`'nın yanında duran page olur. Prefix'siz bir URL ise hiçbir dile ait
değildir. Prefix olmadan istenen bir page, yani `/about` ya da kök, kalıcı olarak
varsayılan locale'deki adresine, yani sırasıyla `/en/about`'a ve `/en`'e redirect
edilir.
Adla üretilen her link de prefix'i taşır. Document'lar da aynı şekilde prefix alır ve
`/en/sitemap.xml`, `/tr/sitemap.xml`'in yanında durur. Tek istisna, sitenin
[`AtRoot`](/docs/documents#example-robotstxt) ile oluşturulan kendi dosyalarıdır:
`/robots.txt` kökte durur, başka hiçbir yerde durmaz (v0.14.1'den beri). Export,
varsayılan locale'i `en/` altına yazar ve köke de okuyucuyu `/en/`'e gönderen bir page
koyar.

Bir fragment bu sonucu `rc.Locale` olarak okur ve doğru içeriği çekmek için
kullanır:

```go
func aboutData(ctx context.Context, rc *collage.RenderContext) (any, []string, error) {
	text, err := cms.Page(ctx, "about", rc.Locale)
	if err != nil {
		return nil, nil, err
	}
	return aboutView{Body: text}, []string{"page:about:" + rc.Locale}, nil
}
```

`DisablePathLocale: true` prefix'leri tamamen kapatır ve bütün request'leri `Default`
locale'de bırakır. Bu ayar, tek dilli bir site ya da dili kendisi
[belirleyen](#negotiating-a-language-yourself) bir site içindir.

### Neden yalnızca URL

collage locale'i hiçbir zaman `Accept-Language` header'ından ya da bir cookie'den
seçmez.

İçeriği isteği yapan kişiye göre değişen bir URL, aslında birden fazla içeriği olan
tek bir URL'dir. URL saklayan her şey de bu durumda yanlış çalışır. Cache ilk
okuyucunun dilini herkese sunar. Arama motoru bir sürümü index'ler, diğerini hiç
görmez. Birinin paylaştığı bir link, onun göndermediği bir dilde açılır. Daha az göze
çarpan bir bozulma da vardır. Türkçe bir tarayıcı `/about`'a giden bir link'i takip
ettiğinde bu path Türkçe path'ler arasında aranır. `/about` orada olmadığı için
okuyucu 404 alır.

Locale URL'de olduğunda her URL, her okuyucu ve her cache için tek bir anlama gelir.
Okuyucunun tercihlerini kullanmak isterseniz bunu yine de yapabilirsiniz; yeter ki
bilerek yapın. Nasıl yapılacağını [aşağıda](#negotiating-a-language-yourself)
bulabilirsiniz.

## Adla link'ler

Path olarak yazılmış bir link, path değiştiğinde sessizce bozulur. Hangi locale'de
olduğunu da bilemez. Bunun yerine page'in register edildiği adla link verin:

```html
<a href="{{pageURL "about"}}">About</a>
<a href="{{pageURL "blog-post" "slug" .Slug}}">{{.Title}}</a>
<link rel="alternate" type="application/rss+xml" href="{{pageURL "feed"}}">
```

`pageURL` önce page'in adını, ardından path parametrelerini ad–değer çiftleri hâlinde
alır. Page'lerin yanı sıra [document'lar](/docs/documents) için de çalışır.

- **Render'ın locale'ini takip eder.** `/tr/hakkinda` üzerinde `{{pageURL "about"}}`
  sonucu `/tr/hakkinda`'dır, `/about` üzerinde ise `/about`'tur. Prefix sizin
  yerinize eklenir.
- **Varsayılan locale'e fallback yapar.** Geçerli locale'de path'i olmayan bir page'e
  link, varsayılan locale'deki path'i üzerinden verilir. Böylece yalnızca İngilizcede
  var olan bir page'e link veren Türkçe bir page yine de render edilir.
- **Katıdır.** Bilinmeyen bir ad, eksik ya da boş bir parametre veya pattern'de
  placeholder'ı olmayan bir parametre render'ı başarısız kılar. Üretilemeyen bir
  link, okuyucuya gösterilecek bir 404 değil, development'ta yakalanması
  gereken bir bug'dır.
- **Değerler escape edilir.** Değeri `.` ya da `..` olan bir parametre reddedilir.
- **Değerler string'dir.** Bir sayıyı `printf` üzerinden geçirin:
  `{{pageURL "user" "id" (printf "%d" .ID)}}`.
- **[`TrailingSlash`](/docs/configuration#trailingslash) açıkken `/` ile biter.** Bu,
  document link'leri için değil, page link'leri için geçerlidir: `/about/`, `/tr/`,
  ama `/feed.xml`.

### Belirli bir locale'de link

`pageURLIn` önce locale'i alır ve tam olarak o locale'e link verir. Fallback yapmaz:

```html
<a href="{{pageURLIn "tr" "about"}}" hreflang="tr">Hakkımızda</a>
```

O locale'de path'i olmayan bir page render'ı başarısız kılar. Sitenin desteklemediği
bir locale de aynı sonucu verir.

### Dil seçici

`localeURL`, render edilen page'in aynı path parametreleriyle başka bir locale'deki
URL'sidir. Page'in o locale'de path'i yoksa boş döner. Böylece `with`, page'in
çevrilmediği bir dili atlar:

```html
<nav class="languages">
  {{with localeURL "en"}}<a hreflang="en" href="{{.}}">English</a>{{end}}
  {{with localeURL "tr"}}<a hreflang="tr" href="{{.}}">Türkçe</a>{{end}}
</nav>
```

Bunu layout'a koyun; her page yalnızca var olan dilleri sunan bir dil seçiciye
kavuşur. Hiç desteklenmeyen bir locale ise yine hata verir. Bu durum çevrilmemiş bir
page değil, template'teki bir yazım hatasıdır.

Path parametreleri olduğu gibi taşınır. İki locale'de de `/blog/{slug}` path'inde
bulunan bir page, `/blog/hello`'dan `/tr/blog/hello`'ya geçer. Türkçe yazılarınızın
slug'ları da Türkçeyse dil seçici bunu bilemez. O link'i içeriğinizden kendiniz
üretirsiniz.

### Arama motorlarına çevirileri bildirmek

Dil seçici okuyucular içindir. Arama motorları bir page'in çevirilerini, head'indeki
`<link rel="alternate" hreflang="…">` element'lerinden öğrenir. Bu element'leri
`rc.HoistAlternate(hreflang, href)` (v0.10.0'dan beri) tanımlar, her dil için bir
tane. Bunu, page'in bulunduğu her locale için `app.URL` ile birlikte kullanın.
[Head ve SEO](/docs/head-and-seo#canonical-and-alternate-links) sayfasında bunu her
page için yapan bir layout var.

### Go'dan link'ler

Go kodunda `app.URL` kullanın. Adıyla belirtilen bir page'e redirect eden bir action,
bütün page'leri listeleyen bir sitemap ve bir page'in head'indeki alternate link'ler
bu duruma örnektir:

```go
target, err := app.URL("blog-post", "tr", map[string]string{"slug": post.Slug})
if err != nil {
	return nil, err
}
return collage.SeeOther(target), nil
```

```go
func (a *App) URL(name, locale string, params map[string]string) (string, error)
```

Boş bir locale, varsayılan locale anlamına gelir. `app.URL`, `pageURLIn` kadar katıdır.
Bilinmeyen bir ad `collage.ErrUnknownRoute` döner. Route'un path'i olmayan bir locale
`collage.ErrNoPathInLocale` döner. Pattern'i tam olarak doldurmayan parametreler
`collage.ErrRouteParams` döner. Hiçbir URL'nin taşıyamayacağı bir locale ise
`collage.ErrLocaleUnreachable` döner. Desteklenmeyen bir locale ya da
`DisablePathLocale` açıkken varsayılan dışındaki herhangi bir locale bu gruptadır.
Sonuç, prefix dahil bir path'tir. Mutlak bir URL gerektiğinde başına sitenizin
origin'ini ekleyin.

### Bir fragment'in URL'si

Bir page'in [`WithFragmentPath`](/docs/forms-and-actions#a-fragment-at-its-own-url)
ile kendi URL'sinde açtığı bir fragment'e de aynı şekilde, page'in ve fragment'in
adıyla link verilir (v0.18.0'dan beri):

```html
<div data-live="{{fragmentURL "home" "cpu-usage"}}">{{slot "cpu-usage"}}</div>
<div data-live="{{fragmentURL "post" "comments" "slug" .Slug}}">…</div>
```

`fragmentURL` render'ın locale'ini kullanır ve varsayılan locale'e fallback yapar.
`fragmentURLIn "tr" "home" "cpu-usage"` ise locale'i kendisi belirtir. Go
tarafındaki karşılığı `app.FragmentURL`'dir:

```go
func (a *App) FragmentURL(page, fragment, locale string, params map[string]string) (string, error)
```

`app.URL` kadar katıdır. Bilinmeyen bir page `collage.ErrUnknownRoute` döner. Page'in
açmadığı bir fragment `collage.ErrUnknownFragmentPath` döner. Bir locale'de iki
path'te açılmış bir fragment `collage.ErrAmbiguousFragmentPath` döner. Eksik
parametreler ise `collage.ErrRouteParams` döner. Bir template'te bunların her biri
render'ı başarısız kılar.

## Dili kendiniz belirlemek

Okuyucunun tarayıcısına göre dil seçmek, siteniz hakkında bir karardır. Bu yüzden bu
kararı siz verirsiniz ve bunu [middleware](/docs/middleware-and-apis) içinde
yaparsınız. Bunun iki dürüst yolu vardır.

**Redirect edin.** Varsayılan locale'in ana sayfasına Türkçe bir tarayıcıyla ya da kendi
tanımladığınız bir dil cookie'siyle gelen okuyucuyu `/tr`'ye gönderin. Her URL yine
tek bir anlama gelir. Yalnızca yeni bir okuyucunun nereden başlayacağını seçmiş
olursunuz.

```go
app.Use(func(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" && strings.HasPrefix(r.Header.Get("Accept-Language"), "tr") {
			http.Redirect(w, r, "/tr", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	})
})
```

**Tek bir URL'yi okuyucunun dilinde render edin.** Kararı middleware'de verin ve
request context'i üzerinden data handler'lara iletin. collage'a da `collage.Vary` ile
haber verin; böylece cache her dil için ayrı bir kopya tutar:

```go
type langKey struct{}

app.Use(func(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lang := "en"
		if strings.HasPrefix(r.Header.Get("Accept-Language"), "tr") {
			lang = "tr"
		}
		collage.Vary(r, "Accept-Language", lang)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), langKey{}, lang)))
	})
})
```

```go
func pageData(ctx context.Context, rc *collage.RenderContext) (any, []string, error) {
	lang, _ := ctx.Value(langKey{}).(string)
	// ...
}
```

`Vary`, çözümlediğiniz değeri page'in cache key'ine ekler. Bu değer tarayıcının
gönderdiği header'ın tamamı değil, `"tr"`'dir. Header'ın adını da response'un `Vary`
header'ına yazar. Böylece bir CDN de sürümleri birbirinden ayrı tutar. `Vary`
middleware'den çağrılmalıdır. v0.11.0'dan beri `Vary`, routing başladığında kapanır
ve bu, cache'lenen ya da cache'lenmeyen her route için geçerlidir. Bu yüzden bir data
handler'dan ya da routing'den sonraki herhangi bir yerden yapılan çağrı hiçbir şeyi
değiştirmez ve `collage.ErrVaryTooLate` döner. Bu yol,
[tek URL, tek içerik](#why-only-the-url) yaklaşımının faydalarından vazgeçer. O
yüzden onu bilerek seçin; genellikle `DisablePathLocale` ile birlikte kullanılır.

[Static export](/docs/static-export) request olmadan render eder, bu yüzden onun için
hiçbir middleware çalışmaz. Export edilen bir sitede locale'ler yalnızca URL'lerde
yer alabilir.
