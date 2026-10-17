---
description: Bir siteyi birkaç dilde sunmak ve page'lere adlarıyla link vermek; böylece link'ler page'i her locale'de takip eder.
reference: LocaleConfig, PageBuilder.WithPath, PageBuilder.WithFragmentPath, RenderContext, App, Vary, ErrLocaleUnreachable, ErrNoPathInLocale, ErrUnknownRoute, ErrUnknownFragmentPath, ErrAmbiguousFragmentPath
---

# Link'ler ve locale'ler

Bir collage sitesi her page'i birkaç dilde, her dili kendi URL'sinde sunabilir.
collage locale'leri route eder, çeviri yapmaz. Bir kelimenin Türkçe karşılığı
içeriğinizde durur ve içeriğiniz zaten nerede tutuluyorsa orada kalır. collage'ın
işi, hangi URL'nin hangi dile ait olduğunu bilmek ve bunu bilen link'ler üretmektir.
Template'lerinizdeki kelimeler için [elagoht/i18n](/docs/plugins#elagohti18n)
plugin'i her locale için bir katalog tutar ve page'in render edildiği locale'de
çeviri yapar.

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
	WithLayouts(layout).
	WithContent(about).
	WithPath("en", "/about").
	WithPath("tr", "/hakkinda").
	Build()
```

Path'ler buradaki gibi farklı olabilir ya da her locale'de aynı pattern olabilir.
Yalnızca tek bir locale'de path'i olan bir page yalnızca o locale'de vardır.

`WithPath`'teki locale bir etiket değildir. Path'in hangi locale'in route'larına
ekleneceğini seçer ve bir request yalnızca `Default` olan ya da `Supported`'da yer
alan bir locale'e ulaşabilir. `Default` siz ayarlamadıkça `"en"`'dir; bu yüzden
yalnızca Türkçe olan bir site onu ayarlar:

```go
Locale: collage.LocaleConfig{Default: "tr"},
```

ve `WithPath("tr", "/hakkinda")` yazar.

v0.35.0'dan beri hiçbir URL'nin ulaşamadığı bir locale'deki path, kayıt sırasında
`collage.ErrLocaleUnreachable` ile reddedilir. Hata, page'in adını ve locale
ayarını söyler. Önceden site hiçbir uyarı vermeden açılıyor, sonra her URL'de 404
dönüyordu. Aynı kural action'lar ve document'lar için de geçerlidir.

v0.60.0'dan beri `WithPath`, page, document ve action için aynı şekilde iki hatayı
daha reddeder:

- iki kez verilen bir locale: `collage.ErrDuplicateLocalePath`, ilk path korunur;
- boş bir locale: `collage.ErrInvalidLocale`.

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

<figure class="diagram-figure" data-search-ignore>
<svg class="diagram" viewBox="0 0 320 260" role="img" aria-labelledby="ll-url-title">
<title id="ll-url-title">Default en, Supported en ile tr iken ve PrefixDefault yokken önce request path'inin ilk segment'ine bakılır. Desteklenen bir locale'in adı değilse, /about ya da /fr/about gibi, locale en olur ve path geldiği gibi eşleştirilir. Yazıldığı gibi tr ise, /tr/hakkinda gibi, locale tr olur ve prefix çıkarılır. Varsayılan locale'in kendi prefix'i ya da başka bir büyük/küçük harf yazılışıysa, /en/about ya da /TR/hakkinda gibi, request herhangi bir page aranmadan tek yazılışa 301 ya da 308 ile redirect edilir. Kalan path yalnızca o locale'in path'leri içinde eşleştirilir; bu yüzden /tr/about 404 döner.</title>
<defs><marker id="ll-url-arrow" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto"><path class="d-head" d="M0 0 8 4 0 8z" /></marker></defs>
<rect class="d-pill" x="100" y="8" width="120" height="24" rx="12" /><text x="160" y="24" text-anchor="middle">request path</text>
<line class="d-flow" x1="160" y1="32" x2="160" y2="46" marker-end="url(#ll-url-arrow)" />
<rect class="d-frame" x="40" y="50" width="240" height="26" rx="6" /><text x="160" y="67" text-anchor="middle">ilk segment bir locale mi?</text>
<path class="d-flow" d="M100 76C100 92 54 90 54 106" marker-end="url(#ll-url-arrow)" />
<line class="d-flow" x1="160" y1="76" x2="160" y2="106" marker-end="url(#ll-url-arrow)" />
<path class="d-flow" d="M220 76C220 92 266 90 266 106" marker-end="url(#ll-url-arrow)" />
<text class="d-small" x="74" y="88" text-anchor="end">hayır</text>
<text class="d-small" x="166" y="96">tr</text>
<text class="d-small" x="250" y="88">en, TR</text>
<rect class="d-a" x="6" y="110" width="96" height="34" rx="5" /><text class="d-on-a" x="54" y="124" text-anchor="middle">Default: en</text><text class="d-on-a d-small" x="54" y="137" text-anchor="middle">path aynen kalır</text>
<rect class="d-b" x="112" y="110" width="96" height="34" rx="5" /><text class="d-on-b" x="160" y="124" text-anchor="middle">tr</text><text class="d-on-b d-small" x="160" y="137" text-anchor="middle">prefix çıkarılır</text>
<rect class="d-c" x="218" y="110" width="96" height="34" rx="5" /><text class="d-on-c" x="266" y="124" text-anchor="middle">redirect</text><text class="d-on-c d-small" x="266" y="137" text-anchor="middle">301 ya da 308</text>
<text class="d-small" x="54" y="160" text-anchor="middle">/about</text><text class="d-small" x="54" y="172" text-anchor="middle">/fr/about</text>
<text class="d-small" x="160" y="160" text-anchor="middle">/tr/hakkinda</text><text class="d-small" x="160" y="172" text-anchor="middle">/tr/about</text>
<text class="d-small" x="266" y="160" text-anchor="middle">/en/about</text><text class="d-small" x="266" y="172" text-anchor="middle">→ /about</text>
<text class="d-small" x="266" y="188" text-anchor="middle">/TR/hakkinda</text><text class="d-small" x="266" y="200" text-anchor="middle">→ /tr/hakkinda</text>
<line class="d-flow" x1="54" y1="178" x2="54" y2="208" marker-end="url(#ll-url-arrow)" />
<line class="d-flow" x1="160" y1="178" x2="160" y2="208" marker-end="url(#ll-url-arrow)" />
<rect class="d-frame" x="6" y="212" width="202" height="40" rx="6" /><text x="107" y="228" text-anchor="middle">o locale'de eşleştirilir</text><text class="d-small" x="107" y="242" text-anchor="middle">/tr/about: tr path'i yok, 404</text>
</svg>
<figcaption><code>Default: "en"</code> ve <code>Supported: []string{"en", "tr"}</code> ile, <code>PrefixDefault</code> olmadan, locale'i path'in ilk segment'i seçer. Orada desteklenen bir locale yoksa locale varsayılan locale'dir ve path geldiği gibi kalır. <code>tr</code> Türkçe demektir ve prefix çıkarılır. Varsayılan locale'in kendi prefix'i ya da farklı büyük/küçük harfle yazılmış bir locale, herhangi bir page aranmadan tek yazılışa redirect edilir. Geriye kalan path yalnızca o locale'in path'leri içinde eşleştirilir.</figcaption>
</figure>

Her page'in locale başına tek bir URL'si vardır. Bu yüzden bir prefix'in diğer
yazılışları kalıcı olarak o URL'ye redirect edilir. Bu yazılışlar iki tanedir:

- varsayılan locale'in kendi prefix'i (varsayılan locale'in URL'leri bu prefix'i
  taşımaz);
- desteklenen bir locale'in farklı büyük/küçük harfle yazılışı.

Redirect, `GET` ve `HEAD` için `301`, diğer bütün method'lar için `308`'dir.
Böylece yanlış yazılışa post edilen bir form `GET`'e dönüşmez, yeniden post edilir.
Query string de redirect ile birlikte taşınır. Bu redirect olmasaydı `/en/about`,
arama motoru için `/about`'un ikinci bir kopyası, cache'te de ikinci bir kayıt
olurdu.

Varsayılan locale dahil bütün dilleri bir prefix altında sunmak isteyen bir site
`PrefixDefault: true` ayarını kullanır (v0.14.0'dan beri). Bu durumda `/en/about`,
`/tr/hakkinda`'nın yanında duran page olur. Prefix'siz bir URL ise hiçbir dile ait
değildir:

- Prefix olmadan istenen bir page, yani `/about` ya da kök, kalıcı olarak
  varsayılan locale'deki adresine, yani sırasıyla `/en/about`'a ve `/en`'e
  redirect edilir.
- Adla üretilen her link de prefix'i taşır.
- Document'lar da aynı şekilde prefix alır ve `/en/sitemap.xml`,
  `/tr/sitemap.xml`'in yanında durur. Tek istisna, sitenin
  [`AtRoot`](/docs/documents#example-robotstxt) ile oluşturulan kendi
  dosyalarıdır: `/robots.txt` kökte durur, başka hiçbir yerde durmaz
  (v0.14.1'den beri).
- Export, varsayılan locale'i `en/` altına yazar ve köke de `/en/` home page'inin
  bir kopyasını koyar. Bu kopya, canonical adresi olarak `/en/`'i gösterir.

<figure class="diagram-figure" data-search-ignore>
<svg class="diagram" viewBox="0 0 320 258" role="img" aria-labelledby="ll-prefix-title">
<title id="ll-prefix-title">PrefixDefault true iken /en/about, about page'ini en'de, /tr/hakkinda ise tr'de sunar. Varsayılan locale'deki bir page'i prefix'siz isteyen /about, 301 ile /en/about'a, kök de /en'e redirect edilir. Static export en page'lerini en/ altına, tr page'lerini tr/ altına yazar; köke de canonical adresi /en/ olan bir en home page kopyası koyar.</title>
<defs><marker id="ll-prefix-arrow" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto"><path class="d-head" d="M0 0 8 4 0 8z" /></marker></defs>
<text class="d-small" x="6" y="14">bir request, PrefixDefault: true</text>
<rect class="d-chip" x="6" y="22" width="104" height="26" rx="13" /><text x="58" y="39" text-anchor="middle">/en/about</text>
<line class="d-flow" x1="110" y1="35" x2="150" y2="35" marker-end="url(#ll-prefix-arrow)" />
<rect class="d-a" x="154" y="22" width="160" height="26" rx="5" /><text class="d-on-a" x="234" y="39" text-anchor="middle">about, en'de</text>
<rect class="d-chip" x="6" y="54" width="104" height="26" rx="13" /><text x="58" y="71" text-anchor="middle">/tr/hakkinda</text>
<line class="d-flow" x1="110" y1="67" x2="150" y2="67" marker-end="url(#ll-prefix-arrow)" />
<rect class="d-b" x="154" y="54" width="160" height="26" rx="5" /><text class="d-on-b" x="234" y="71" text-anchor="middle">about, tr'de</text>
<rect class="d-chip" x="6" y="86" width="104" height="26" rx="13" /><text x="58" y="103" text-anchor="middle">/about</text>
<line class="d-flow" x1="110" y1="99" x2="150" y2="99" marker-end="url(#ll-prefix-arrow)" />
<rect class="d-c" x="154" y="86" width="160" height="26" rx="5" /><text class="d-on-c" x="234" y="103" text-anchor="middle">301 → /en/about</text>
<rect class="d-chip" x="6" y="118" width="104" height="26" rx="13" /><text x="58" y="135" text-anchor="middle">/</text>
<line class="d-flow" x1="110" y1="131" x2="150" y2="131" marker-end="url(#ll-prefix-arrow)" />
<rect class="d-c" x="154" y="118" width="160" height="26" rx="5" /><text class="d-on-c" x="234" y="135" text-anchor="middle">301 → /en</text>
<line class="d-rule" x1="6" y1="160" x2="314" y2="160" />
<text class="d-small" x="6" y="178">static export</text>
<rect class="d-a" x="6" y="186" width="96" height="40" rx="5" /><text class="d-on-a" x="54" y="202" text-anchor="middle">en/</text><text class="d-on-a d-small" x="54" y="217" text-anchor="middle">en page'leri</text>
<rect class="d-b" x="112" y="186" width="96" height="40" rx="5" /><text class="d-on-b" x="160" y="202" text-anchor="middle">tr/</text><text class="d-on-b d-small" x="160" y="217" text-anchor="middle">tr page'leri</text>
<rect class="d-frame" x="218" y="186" width="96" height="40" rx="5" /><text x="266" y="202" text-anchor="middle">kök</text><text class="d-small" x="266" y="217" text-anchor="middle">/en kopyası</text>
<text class="d-small" x="266" y="242" text-anchor="middle">canonical: /en/</text>
</svg>
<figcaption><code>PrefixDefault: true</code> iken her dilin bir prefix'i vardır. Prefix'siz adresinden istenen varsayılan locale page'i, kök dahil, 301 ile prefix'li adrese redirect edilir. Static bir host'un çıplak domain'i prefix'li home page'e gönderecek bir yolu yoktur; bu yüzden export varsayılan locale'in page'lerini <code>en/</code> altına yazar ve köke <code>/en/</code> home page'inin bir kopyasını koyar. Bu kopya canonical adresi olarak <code>/en/</code>'i gösterir.</figcaption>
</figure>

Bir fragment bu sonucu `rc.Locale` olarak okur ve doğru içeriği çekmek için
kullanır:

```go
func aboutData(ctx context.Context, rc *collage.RenderContext) (aboutView, []string, error) {
	text, err := cms.Page(ctx, "about", rc.Locale)
	if err != nil {
		return aboutView{}, nil, err
	}
	return aboutView{Body: text}, []string{"page:about:" + rc.Locale}, nil
}
```

Prefix'leri kapatan bir ayar yoktur: `DisablePathLocale` v0.60.0'da kaldırıldı. Tek
dilli bir site path'lerini prefix taşımayan `Default`'ta register eder; dili kendisi
[belirleyen](#negotiating-a-language-yourself) bir site ise bunu middleware'de
yapar.

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
- **Değerler string, tam sayı ya da `fmt.Stringer` olabilir** (son ikisi
  v0.36.0'dan beri): `{{pageURL "user" "id" .ID}}`.
- **[`TrailingSlash`](/docs/configuration#trailingslash) açıkken `/` ile biter.** Bu,
  document link'leri için değil, page link'leri için geçerlidir: `/about/`, `/tr/`,
  ama `/feed.xml`.

<figure class="diagram-figure" data-search-ignore>
<svg class="diagram" viewBox="0 0 320 200" role="img" aria-labelledby="ll-name-title">
<title id="ll-name-title">Tek bir template çağrısı, pageURL "about", page en'de /about üzerinde render edildiğinde /about, tr'de /tr/hakkinda üzerinde render edildiğinde /tr/hakkinda olur. tr path'i olmayan careers page'ine bir tr render'ından en path'i, /careers, üzerinden link verilir.</title>
<defs><marker id="ll-name-arrow" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto"><path class="d-head" d="M0 0 8 4 0 8z" /></marker></defs>
<rect class="d-cmd" x="70" y="8" width="180" height="24" rx="12" /><text class="d-cmd-text" x="160" y="24" text-anchor="middle">pageURL "about"</text>
<path class="d-flow" d="M130 32C130 42 79 38 79 46" marker-end="url(#ll-name-arrow)" />
<path class="d-flow" d="M190 32C190 42 241 38 241 46" marker-end="url(#ll-name-arrow)" />
<rect class="d-a" x="6" y="50" width="146" height="40" rx="5" /><text class="d-on-a" x="79" y="66" text-anchor="middle">/about üzerinde</text><text class="d-on-a d-small" x="79" y="81" text-anchor="middle">en render'ı</text>
<rect class="d-b" x="168" y="50" width="146" height="40" rx="5" /><text class="d-on-b" x="241" y="66" text-anchor="middle">/tr/hakkinda'da</text><text class="d-on-b d-small" x="241" y="81" text-anchor="middle">tr render'ı</text>
<line class="d-flow" x1="79" y1="90" x2="79" y2="104" marker-end="url(#ll-name-arrow)" />
<line class="d-flow" x1="241" y1="90" x2="241" y2="104" marker-end="url(#ll-name-arrow)" />
<rect class="d-chip" x="19" y="108" width="120" height="24" rx="12" /><text x="79" y="124" text-anchor="middle">/about</text>
<rect class="d-chip" x="181" y="108" width="120" height="24" rx="12" /><text x="241" y="124" text-anchor="middle">/tr/hakkinda</text>
<rect class="d-frame" x="6" y="148" width="308" height="44" rx="6" /><text x="160" y="165" text-anchor="middle">pageURL "careers", /tr/hakkinda'da</text><text class="d-small" x="160" y="181" text-anchor="middle">tr path'i yok: en path'i, /careers</text>
</svg>
<figcaption>Paylaşılan bir layout'taki aynı <code>{{pageURL "about"}}</code>, İngilizce page'de <code>/about</code>, Türkçe page'de <code>/tr/hakkinda</code> olur. Render'ın locale'inde path'i olmayan bir page'e, burada yalnızca İngilizcede var olan bir <code>careers</code> page'ine, varsayılan locale'deki path'i üzerinden link verilir.</figcaption>
</figure>

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

[Kendi URL'sindeki bir action'a](/docs/forms-and-actions#an-action-at-its-own-url)
post eden bir form da action'ı aynı şekilde adıyla belirtir:
`{{actionURL "logout"}}` (v0.36.0'dan beri).

### Go'dan link'ler

Bir data handler'da ya da action'da page'ler için `rc.URL`, action'lar için
`rc.ActionURL` kullanın (v0.37.0'dan beri). Adıyla belirtilen bir page'e redirect
etmek ya da canonical URL üretmek bu duruma örnektir. İkisi de URL'yi render'ın
kendi locale'inde üretir ve `pageURL` ile `actionURL` gibi varsayılan locale'e
fallback yapar. Böylece link gereken koda `*App`'i taşımak gerekmez:

```go
target, err := rc.URL("blog-post", map[string]string{"slug": post.Slug})
if err != nil {
	return nil, err
}
return collage.SeeOther(target), nil
```

```go
func (rc *RenderContext) URL(name string, params map[string]string) (string, error)
func (rc *RenderContext) ActionURL(name string, params map[string]string) (string, error)
```

Doğrudan request'ten gelen bir değer bile tek bir escape edilmiş path
segment'idir. `/` içeren ya da `.` veya `..` olan bir değer
`collage.ErrRouteParams` döner. Böylece bir değer path'i başka bir sitenin
URL'sine çeviremez. collage'ın oluşturmadığı bir `RenderContext`'te (örneğin bir
testin elle kurduğu) ikisi de `collage.ErrUnknownRoute` döner.

Render'ın kendi locale'i dışında bir locale için ya da render dışında (bütün
page'leri listeleyen bir sitemap, bir page'in head'indeki alternate link'ler)
locale'i parametre olarak alan `app.URL` kullanın:

```go
target, err := app.URL("blog-post", "tr", map[string]string{"slug": post.Slug})
```

```go
func (a *App) URL(name, locale string, params map[string]string) (string, error)
```

Boş bir locale, varsayılan locale anlamına gelir. `app.URL`, `pageURLIn` kadar
katıdır:

| Sorun | Hata |
| --- | --- |
| Bilinmeyen bir ad | `collage.ErrUnknownRoute` |
| Route'un path'i olmayan bir locale | `collage.ErrNoPathInLocale` |
| Pattern'i tam olarak doldurmayan parametreler | `collage.ErrRouteParams` |
| Hiçbir URL'nin taşıyamayacağı, yani ne varsayılan ne de desteklenen bir locale | `collage.ErrLocaleUnreachable` |

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

`app.URL` kadar katıdır:

| Sorun | Hata |
| --- | --- |
| Bilinmeyen bir page | `collage.ErrUnknownRoute` |
| Page'in açmadığı bir fragment | `collage.ErrUnknownFragmentPath` |
| Bir locale'de iki path'te açılmış bir fragment | `collage.ErrAmbiguousFragmentPath` |
| Eksik parametreler | `collage.ErrRouteParams` |

Bir template'te bunların her biri render'ı başarısız kılar.

## Dili kendiniz belirlemek

Okuyucunun tarayıcısına göre dil seçmek, siteniz hakkında bir karardır. Bu yüzden bu
kararı siz verirsiniz ve bunu [middleware](/docs/middleware-and-apis) içinde
yaparsınız. Bunun iki dürüst yolu vardır.

**Redirect edin.** Varsayılan locale'in home page'ine Türkçe bir tarayıcıyla ya da
kendi tanımladığınız bir dil cookie'siyle gelen okuyucuyu `/tr`'ye gönderin. Her
URL yine tek bir anlama gelir. Yalnızca yeni bir okuyucunun nereden başlayacağını
seçmiş olursunuz.

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
func pageData(ctx context.Context, rc *collage.RenderContext) (pageView, []string, error) {
	lang, _ := ctx.Value(langKey{}).(string)
	// ...
}
```

`Vary`, çözümlediğiniz değeri page'in cache key'ine ekler. Bu değer tarayıcının
gönderdiği header'ın tamamı değil, `"tr"`'dir. Header'ın adını da response'un `Vary`
header'ına yazar. Böylece bir CDN de sürümleri birbirinden ayrı tutar.

`Vary` middleware'den çağrılmalıdır. v0.11.0'dan beri `Vary`, routing
başladığında kapanır ve bu, cache'lenen ya da cache'lenmeyen her route için
geçerlidir. Bu yüzden bir data handler'dan ya da routing'den sonraki herhangi bir
yerden yapılan çağrı hiçbir şeyi değiştirmez ve `collage.ErrVaryTooLate` döner.

Bu yol, [tek URL, tek içerik](#why-only-the-url) yaklaşımının faydalarından
vazgeçer. O yüzden onu bilerek seçin; genellikle bütün page'lerinin path'i
`Default`'ta olan, yani hiçbir URL'nin bir locale adlandırmadığı bir sitede
kullanılır.

[Static export](/docs/static-export) request olmadan render eder, bu yüzden yazdığı
dosyalar için hiçbir middleware çalışmaz. Export edilen bir sitede locale'ler
yalnızca URL'lerde yer alabilir.

## Örnek: iki dilli bir ürün kataloğu

Bir mağaza İngilizce ve Türkçe satış yapıyor. İki page'inin her locale'de bir
path'i var. Ürün page'i stok kutusunu kendi URL'sinde de açıyor; böylece bir
script kutuyu yenileyebiliyor:

| Page | `en` | `tr` |
| --- | --- | --- |
| `catalogue` | `/products` | `/tr/urunler` |
| `product` | `/products/{id}` | `/tr/urunler/{id}` |
| `product` page'inin stok kutusu | `/products/{id}/stock` | `/tr/urunler/{id}/stok` |

Locale ayarı [yukarıdaki](#configuring-locales) ayardır: `Default: "en"`,
`Supported: []string{"en", "tr"}`. `catalog` paketi sizin kendi veri katmanınızın
yerini tutar. Her ürünün adını iki dilde de saklar ve handler'lar ondan render'ın
locale'indeki adı ister:

```go
func loadCatalogue(ctx context.Context, rc *collage.RenderContext) ([]catalog.Product, []string, error) {
	list, err := catalog.List(ctx, rc.Locale)
	if err != nil {
		return nil, nil, err
	}
	return list, []string{"catalog"}, nil
}

func loadProduct(ctx context.Context, rc *collage.RenderContext) (catalog.Product, []string, error) {
	id := rc.Param("id")
	p, err := catalog.Find(ctx, id, rc.Locale)
	if errors.Is(err, catalog.ErrNoProduct) {
		return catalog.Product{}, nil, fmt.Errorf("catalog: no product %q: %w", id, collage.ErrNotFound)
	}
	if err != nil {
		return catalog.Product{}, nil, err
	}
	return p, []string{"product:" + id}, nil
}

func loadStock(ctx context.Context, rc *collage.RenderContext) (int, error) {
	return catalog.Stock(ctx, rc.Param("id"))
}
```

Her page locale başına bir path tanımlar. Ürün page'i stok fragment'ini her
locale'de açar:

```go
layout := collage.NewFragment("layout", "layouts/default.html").Build()

stock := collage.NewFragment("product-stock", "fragments/stock.html").
	WithData(collage.Load(loadStock)).
	Build()

catalogue := collage.NewPage("catalogue").
	WithLayouts(layout).
	WithContent(collage.NewFragment("catalogue", "pages/catalogue.html").
		WithData(collage.LoadTagged(loadCatalogue)).
		Build()).
	WithPath("en", "/products").
	WithPath("tr", "/urunler").
	Build()

product := collage.NewPage("product").
	WithLayouts(layout).
	WithContent(collage.NewFragment("product", "pages/product.html").
		WithData(collage.LoadTagged(loadProduct)).
		WithSlotFragment("stock", stock).
		Build()).
	WithPath("en", "/products/{id}").
	WithPath("tr", "/urunler/{id}").
	WithFragmentPath("en", "/products/{id}/stock", stock).
	WithFragmentPath("tr", "/urunler/{id}/stok", stock).
	Build()

if err := app.Register(catalogue, product); err != nil {
	return err
}
```

Template'ler path değil, page adı yazar. Dil seçici layout'tadır. Katalog her
ürüne page adı ve `id`'siyle link verir. Ürün page'i de stok kutusunun URL'sini
adıyla belirtir:

```html
<!-- templates/layouts/default.html -->
<nav class="languages">
  {{with localeURL "en"}}<a hreflang="en" href="{{.}}">English</a>{{end}}
  {{with localeURL "tr"}}<a hreflang="tr" href="{{.}}">Türkçe</a>{{end}}
</nav>
{{slot "content"}}
```

```html
<!-- templates/pages/catalogue.html -->
<ul>
  {{range .}}
  <li><a href="{{pageURL "product" "id" .ID}}">{{.Name}}</a></li>
  {{end}}
</ul>
```

```html
<!-- templates/pages/product.html -->
<article>
  <h1>{{.Name}}</h1>
  <div data-live="{{fragmentURL "product" "product-stock" "id" .ID}}">{{slot "stock"}}</div>
</article>
```

```html
<!-- templates/fragments/stock.html -->
<p class="stock">{{.}}</p>
```

Türkçe page'lerde bütün link'ler Türkçe çıkar:

| Template'te | Sonuç |
| --- | --- |
| `localeURL "en"` | `/products/42` |
| `localeURL "tr"` | `/tr/urunler/42` |
| `pageURL "product" "id" .ID`, katalogda | `/tr/urunler/42` |
| `fragmentURL "product" "product-stock" "id" .ID` | `/tr/urunler/42/stok` |

`/products/42` üzerinde aynı template'ler `/products/42` ve `/products/42/stock`
üretir. Bunu mümkün kılan birkaç şey var:

- **İki pattern de parametreye `{id}` adını verir.** `localeURL`, render'ın
  parametrelerini adlarıyla taşır ve parametrelerin dolduramadığı bir pattern
  render'ı başarısız kılar. Id iki dilde de aynı sayı olduğu için dil seçicinin
  başka bir şeye ihtiyacı yoktur. Her dilde ayrı çevrilen bir slug olsaydı, o
  link'i [yukarıda](#a-language-switcher) anlatıldığı gibi içeriğinizden kendiniz
  üretmeniz gerekirdi.
- **`fragmentURL` slot'u değil, fragment'i adlandırır**: `NewFragment`'e verilen
  ad olan `product-stock`.
- **Stok kutusunun her locale'de bir path'i vardır.** Yalnızca İngilizce path
  olsaydı Türkçe page, [varsayılan locale fallback'i](#a-fragments-url) ile
  `/products/42/stock`'a link verirdi.
- **Yanlış yazılışlar sizin yerinize ele alınır.** `/en/products/42`,
  `/products/42`'ye redirect edilir. `/tr/products/42` ise 404 döner, çünkü
  `/products/{id}` Türkçe path'ler arasında değildir.
