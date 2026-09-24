---
description: Bir siteyi birkaç dilde sunmak ve sayfalara adlarıyla bağlantı vererek bağlantıların onları her locale'de izlemesini sağlamak.
---

# Bağlantılar ve locale'ler

Bir collage sitesi her sayfayı birkaç dilde, her birini kendi URL'sinde sunabilir.
collage locale'leri yönlendirir (route eder); çeviri yapmaz. Bir kelimenin Türkçede
ne olması gerektiği içeriğinizde, o içerik zaten nerede duruyorsa orada yer alır — collage'ın
payına düşen, hangi URL'nin hangi dile ait olduğunu bilmek ve bunu bilen bağlantılar
kurmaktır.

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

`Default`, locale öneki taşımayan bir URL'nin locale'idir ve varsayılan değeri
`"en"`'dir. `Supported` sitenin sunduğu her locale'i listeler ve `Default`'u
içermelidir (aksi hâlde `collage.ErrLocaleDefaultNotSupported`); boş bırakılırsa
yalnızca `Default`'tan oluşur.

Ardından bir sayfa, var olduğu her locale'deki yolunu bildirir:

```go
collage.NewPage("about").
	WithLayout(layout).
	WithContent(about).
	WithPath("en", "/about").
	WithPath("tr", "/hakkinda").
	Build()
```

Yollar buradaki gibi farklı olabilir ya da her locale'de aynı pattern olabilir.
Yalnızca bir locale'de yolu olan bir sayfa yalnızca o locale'de vardır.

## Locale'i URL belirler

Locale öneki olmayan bir yol `Default` locale'dedir. İlk segment'i desteklenen bir
locale'i adlandıran bir yol o locale'dedir ve yol eşleştirilmeden önce önek
kaldırılır:

| URL | Locale | Eşleştirildiği yer |
| --- | --- | --- |
| `/about` | `en` | `en` yollarındaki `/about` |
| `/tr/hakkinda` | `tr` | `tr` yollarındaki `/hakkinda` |
| `/tr/about` | `tr` | `tr` yollarındaki `/about` — bulunamaz, 404 |
| `/fr/about` | `en` | `/fr/about` — `fr` desteklenmiyor, bu yüzden sıradan bir segment |
| `/en/about` | — | `/about`'a yönlendirilir |
| `/TR/hakkinda` | — | `/tr/hakkinda`'ya yönlendirilir |

Her sayfanın locale başına tek bir URL'si vardır; bu yüzden bir önekin diğer
yazımları kalıcı olarak o URL'ye yönlendirilir: varsayılan locale'in, kendi
URL'lerinde bulunmayan öneki ve desteklenen bir locale'in farklı büyük/küçük harfle
yazılışı. Yönlendirme `GET` ya da `HEAD` için `301`, diğer her şey için `308`'dir;
böylece yanlış yazılışa gönderilen bir form `GET`'e dönüştürülmek yerine yeniden
gönderilir ve query string de beraberinde gider. Bu yönlendirme olmasaydı
`/en/about`, bir arama motoru için `/about`'un ikinci bir kopyası, önbellekte de
ikinci bir girdi olurdu.

Bir fragment sonucu `rc.Locale` olarak okur ve doğru içeriği çekmek için kullanır:

```go
func aboutData(ctx context.Context, rc *collage.RenderContext) (aboutView, []string, error) {
	text, err := cms.Page(ctx, "about", rc.Locale)
	if err != nil {
		return aboutView{}, nil, err
	}
	return aboutView{Body: text}, []string{"page:about:" + rc.Locale}, nil
}
```

`DisablePathLocale: true` önekleri tamamen kapatır ve her isteği `Default` locale'de
bırakır — tek dilli bir site ya da dili kendi başına
[belirleyen](#negotiating-a-language-yourself) bir site için.

### Neden yalnızca URL

collage asla `Accept-Language` header'ından ya da bir çerezden locale seçmez.

İçeriği kimin sorduğuna bağlı olan bir URL, birden çok içeriği olan tek bir URL'dir
ve URL saklayan her şey onu yanlış anlar: bir önbellek ilk okuyucunun dilini herkese
sunar, bir arama motoru bir sürümü dizine ekler ve diğerini hiç görmez, birinin
paylaştığı bir bağlantı da onun göndermediği bir dilde açılır. Daha az belirgin bir
şekilde de bozulur — `/about`'a giden bir bağlantıyı izleyen Türkçe bir tarayıcı,
`/about`'un bulunmadığı Türkçe yollar arasında aranır ve bir 404 alır.

Locale URL'de olduğunda her URL, her okuyucu ve her önbellek için tek bir anlama
gelir. Okuyucunun tercihlerini kullanmak isterseniz bunu yine de bilinçli olarak
yapabilirsiniz — [aşağıya](#negotiating-a-language-yourself) bakın.

## Adla bağlantılar

Yol olarak yazılmış bir bağlantı, yol değiştiğinde sessizce bozulur ve hangi
locale'de olduğunu bilemez. Sayfanın kaydedildiği adla bağlantı verin:

```html
<a href="{{pageURL "about"}}">About</a>
<a href="{{pageURL "blog-post" "slug" .Slug}}">{{.Title}}</a>
<link rel="alternate" type="application/rss+xml" href="{{pageURL "feed"}}">
```

`pageURL` önce sayfanın adını, ardından yol parametrelerini ad–değer çiftleri olarak
alır. Sayfaların yanı sıra [document](/docs/documents)'lar için de çalışır.

- **Render'ın locale'ini izler.** `/tr/hakkinda` üzerinde `{{pageURL "about"}}`
  `/tr/hakkinda`'dır; `/about` üzerinde `/about`'tur. Önek sizin için eklenir.
- **Varsayılan locale'e geri düşer.** Geçerli locale'de yolu olmayan bir sayfa için
  varsayılan locale'deki yola bağlantı verilir; böylece yalnızca İngilizcede var olan
  bir sayfaya bağlantı veren Türkçe bir sayfa yine de render edilir.
- **Katıdır.** Bilinmeyen bir ad, eksik ya da boş bir parametre veya pattern'de
  karşılığı olan placeholder bulunmayan bir parametre render'ı başarısız kılar.
  Kurulamayan bir bağlantı, okuyucu için bir 404 değil, geliştirmede bulunacak bir
  hatadır.
- **Değerler escape edilir**; `.` ya da `..` değeri reddedilir.
- **Değerler dizedir.** Bir sayıyı `printf`'ten geçirin:
  `{{pageURL "user" "id" (printf "%d" .ID)}}`.
- **[`TrailingSlash`](/docs/configuration#trailingslash) açıkken `/` ile biter** —
  bir document'in değil, bir sayfanın bağlantısı: `/about/`, `/tr/`, ama `/feed.xml`.

### Belirli bir locale'de bağlantı

`pageURLIn` önce locale'i alır ve tam olarak o locale'e bağlantı verir — geri düşme
yoktur:

```html
<a href="{{pageURLIn "tr" "about"}}" hreflang="tr">Hakkımızda</a>
```

O locale'de yolu olmayan bir sayfa render'ı başarısız kılar; sitenin desteklemediği
bir locale de öyle.

### Dil değiştirici

`localeURL`, render edilmekte olan sayfanın, aynı yol parametreleriyle, başka bir
locale'deki hâlidir. Sayfanın o locale'de yolu yoksa boştur; bu yüzden `with`, sayfanın
çevrilmediği bir dili atlar:

```html
<nav class="languages">
  {{with localeURL "en"}}<a hreflang="en" href="{{.}}">English</a>{{end}}
  {{with localeURL "tr"}}<a hreflang="tr" href="{{.}}">Türkçe</a>{{end}}
</nav>
```

Onu layout'a koyun; her sayfa yalnızca var olanı sunan bir dil değiştiriciye sahip
olur. Hiç desteklenmeyen bir locale ise yine hatadır — bu çevrilmemiş bir sayfa
değil, şablondaki bir yazım hatasıdır.

Yol parametreleri olduğu gibi aktarılır. Her iki locale'de `/blog/{slug}` yolunda
olan bir sayfa `/blog/hello`'dan `/tr/blog/hello`'ya geçer; Türkçe yazılarınızın
Türkçe slug'ları varsa değiştirici bunu bilemez ve o bağlantıyı içeriğinizden
kendiniz kurarsınız.

### Arama motorlarına çevirileri bildirmek

Dil değiştirici okuyucular içindir. Arama motorları bir sayfanın çevirilerini, head'indeki
`<link rel="alternate" hreflang="…">` öğelerinden öğrenir; bunları her dil için bir
tane olmak üzere `rc.HoistAlternate(hreflang, href)` (v0.10.0'dan itibaren) bildirir.
Bunu, sayfanın var olduğu her locale için `app.URL` ile birlikte kullanın —
[Head ve SEO](/docs/head-and-seo#canonical-and-alternate-links) sayfasında bunu her
sayfa için yapan bir layout var.

### Go'dan bağlantılar

Go'da — adı verilen bir sayfaya yönlendiren bir action, her sayfayı listeleyen bir
sitemap, bir sayfanın head'indeki alternatifler — `app.URL` kullanın:

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

Boş bir locale varsayılan locale demektir. `pageURLIn` kadar katıdır: bilinmeyen bir
ad `collage.ErrUnknownRoute`, route'un yolu olmadığı bir locale
`collage.ErrNoPathInLocale`, pattern'i tam olarak doldurmayan parametreler
`collage.ErrRouteParams`, hiçbir URL'nin taşıyamayacağı bir locale — desteklenmeyen
ya da `DisablePathLocale` açıkken varsayılan dışındaki herhangi biri — ise
`collage.ErrLocaleUnreachable` hatasıdır. Sonuç, önek dahil bir yoldur; mutlak bir
URL gerektiğinde sitenizin origin'ini ekleyin.

## Dili kendiniz belirlemek

Okuyucunun tarayıcısına göre dil seçmek siteniz hakkında bir karardır; bu yüzden bu
karar sizindir ve [middleware](/docs/middleware-and-apis)'de verilir. Bunu yapmanın
iki dürüst yolu var.

Varsayılan locale'in ana sayfasına Türkçe bir tarayıcıyla ya da size ait bir dil
çereziyle gelen okuyucuyu `/tr`'ye **yönlendirin**. Her URL yine tek bir anlama
gelir; yalnızca yeni bir okuyucunun nereden başlayacağını seçmiş olursunuz.

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

**Tek bir URL'yi okuyucunun dilinde render edin.** Kararı middleware'de verin,
istek context'i üzerinden data handler'lara iletin ve önbelleğin dil başına bir kopya
tutması için bunu collage'a `collage.Vary` ile bildirin:

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
func pageData(ctx context.Context, rc *collage.RenderContext) (view, []string, error) {
	lang, _ := ctx.Value(langKey{}).(string)
	// ...
}
```

`Vary`, çözümlediğiniz değeri — tarayıcının header'ının tamamını değil, `"tr"`'yi —
sayfanın önbellek anahtarına, header adını da yanıtın `Vary` header'ına koyar;
böylece bir CDN de sürümleri birbirinden ayrı tutar. Middleware'den çağrılmalıdır.
v0.11.0'dan itibaren `Vary`, routing başladığında her route'ta — önbelleğe alınsın ya
da alınmasın — kapanır; bu yüzden bir data handler'dan ya da routing'den sonraki
herhangi bir yerden yapılan çağrı hiçbir şeyi değiştirmez ve `collage.ErrVaryTooLate`
döner. Bu, [tek URL, tek içerik](#why-only-the-url) ilkesinin faydalarından vazgeçen
yoldur; bu yüzden onu bilerek seçin — genellikle `DisablePathLocale` ile birlikte.

Bir [statik dışa aktarma](/docs/static-export) istek olmadan render eder, dolayısıyla
onun için hiçbir middleware çalışmaz: dışa aktarılmış bir sitenin locale'leri
yalnızca URL'lerinde olabilir.
