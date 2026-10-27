---
description: collage'ın render edilmiş page'leri ve onların dayandığı veriyi nasıl cache'lediğini, neyi atacağını nasıl bildiğini anlatır.
reference: CacheConfig, Cached, App.PurgeTags, CacheStale, CacheWarm, Once, Key, TaggedCache, StaleCache, CacheStatusHeader, SkipCache, Vary, PageBuilder.Static, PageBuilder.Incremental, PageBuilder.Dynamic, FragmentBuilder.Static, FragmentBuilder.Shared, StrategyAuto, PathTag, SharedCache, Reached, WriteOutcome, WriteFresh, WriteStale, WriteSkipped
---

# Caching

collage iki şeyi cache'ler. **Page cache**, bir render'ın ürettiği çıktıyı, yani tek
bir URL'nin HTML'ini saklar. Böylece o URL'yi sonra açan okuyucu, render
beklemeden page'i alır. **Data cache** (`collage.Cached`) ise render'ların
dayandığı veriyi saklar. Böylece aynı yazarı gösteren otuz page o yazarı yalnızca
bir kez çeker.

İkisi de aynı **dependency tag**'lerle indekslenir. Bu sayede tek bir çağrı, bir
içeriği ve ondan üretilmiş her şeyi birlikte atar.

Siz açana kadar hiçbir şey cache'lenmez:

```go
app, err := collage.New(&collage.Config{
	Template: collage.TemplateConfig{Root: "templates"},
	Cache: collage.CacheConfig{
		Enabled:    true,
		Type:       "memory",
		DefaultTTL: 5 * time.Minute,
		MaxEntries: 10000,
		MaxBytes:   256 << 20, // 0 means the type's default; negative, no limit
		WarmLimit:  0,         // pages re-rendered after an invalidation; 0 means 1000, negative, off
	},
})
```

`Enabled` false olduğunda her request render edilir ve hiçbir şey saklanmaz.
Page'lerin ne tanımladığı bunu değiştirmez.

## Render stratejileri

Her page, çıktısının nasıl yeniden kullanılabileceğini builder'ındaki tek bir
çağrıyla belirtir:

| Builder çağrısı | Ne olur | Gönderilen `Cache-Control` |
| --- | --- | --- |
| `Dynamic()` | Her request'te render edilir, hiç saklanmaz | `no-store` |
| `Static()` | Bir kez render edilir, bir şey onu invalidate edene kadar sunulur | `public, max-age=0, must-revalidate` |
| `Incremental(ttl)` | Render'dan sonra `ttl` geçene kadar cache'ten sunulur | `public, max-age=<saniye cinsinden ttl>` |

<figure class="diagram-figure" data-search-ignore>
<svg class="diagram" viewBox="0 0 320 312" role="img" aria-labelledby="cache-strategy-title">
<title id="cache-strategy-title">Bir page register edildiğinde tanımladığı strateji korunur: Dynamic her request'te render edilir ve hiç saklanmaz, Static bir tag invalidate edilene kadar, Incremental ttl'i dolana kadar tutulur. Hiçbirini tanımlamayan bir page, render ettiği herhangi bir yerde bir data handler ya da slot resolver varsa dynamic olur (Static bir fragment'in kendi handler'ı hariç), yoksa static olur.</title>
<defs><marker id="cache-strategy-arrow" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto"><path class="d-head" d="M0 0 8 4 0 8z" /></marker></defs>
<rect class="d-pill" x="90" y="8" width="140" height="24" rx="12" /><text x="160" y="24" text-anchor="middle">page register edilir</text>
<line class="d-flow" x1="160" y1="32" x2="160" y2="46" marker-end="url(#cache-strategy-arrow)" />
<rect class="d-frame" x="70" y="50" width="180" height="26" rx="6" /><text x="160" y="67" text-anchor="middle">strateji tanımlı mı?</text>
<path class="d-flow" d="M130 76C130 92 42 90 42 106" marker-end="url(#cache-strategy-arrow)" />
<path class="d-flow" d="M150 76C150 92 118 90 118 106" marker-end="url(#cache-strategy-arrow)" />
<path class="d-flow" d="M170 76C170 92 194 90 194 106" marker-end="url(#cache-strategy-arrow)" />
<path class="d-flow" d="M190 76C190 92 270 90 270 106" marker-end="url(#cache-strategy-arrow)" />
<rect class="d-c" x="6" y="110" width="72" height="34" rx="5" /><text class="d-on-c" x="42" y="131" text-anchor="middle">Dynamic()</text>
<rect class="d-a" x="82" y="110" width="72" height="34" rx="5" /><text class="d-on-a" x="118" y="131" text-anchor="middle">Static()</text>
<rect class="d-b" x="158" y="110" width="72" height="34" rx="5" /><text class="d-on-b d-small" x="194" y="124" text-anchor="middle">Incremental</text><text class="d-on-b d-small" x="194" y="137" text-anchor="middle">(ttl)</text>
<rect class="d-frame" x="234" y="110" width="72" height="34" rx="5" /><text x="270" y="124" text-anchor="middle">yok</text><text class="d-small" x="270" y="137" text-anchor="middle">StrategyAuto</text>
<text class="d-small" x="42" y="160" text-anchor="middle">her request</text><text class="d-small" x="42" y="172" text-anchor="middle">saklanmaz</text>
<text class="d-small" x="118" y="160" text-anchor="middle">invalidate</text><text class="d-small" x="118" y="172" text-anchor="middle">edilene kadar</text>
<text class="d-small" x="194" y="160" text-anchor="middle">ttl dolana</text><text class="d-small" x="194" y="172" text-anchor="middle">kadar</text>
<line class="d-flow" x1="270" y1="144" x2="270" y2="192" marker-end="url(#cache-strategy-arrow)" />
<rect class="d-frame" x="6" y="196" width="300" height="52" rx="6" /><text class="d-small" x="156" y="212" text-anchor="middle">bir data handler ya da slot resolver</text><text class="d-small" x="156" y="225" text-anchor="middle">page'in render ettiği bir yerde mi?</text><text class="d-small" x="156" y="238" text-anchor="middle">(Static() fragment'in kendi handler'ı: hayır)</text>
<line class="d-flow" x1="90" y1="248" x2="90" y2="272" marker-end="url(#cache-strategy-arrow)" /><text class="d-small" x="98" y="264">evet</text>
<line class="d-flow" x1="222" y1="248" x2="222" y2="272" marker-end="url(#cache-strategy-arrow)" /><text class="d-small" x="230" y="264">hayır</text>
<rect class="d-c" x="30" y="276" width="120" height="28" rx="5" /><text class="d-on-c" x="90" y="294" text-anchor="middle">dynamic</text>
<rect class="d-a" x="162" y="276" width="120" height="28" rx="5" /><text class="d-on-a" x="222" y="294" text-anchor="middle">static</text>
</svg>
<figcaption>Bir page bir strateji tanımladıysa o stratejiyi korur. Hiçbirini tanımlamayan bir page'in stratejisi register edilirken belirlenir: her render'da veri çeken herhangi bir şey onu dynamic yapar, template'lerden ve sabit değerlerden oluşan bir page ise static olur. Ayrıntılar <a href="#a-page-that-declares-none">aşağıdadır</a>. <code>Cache.Enabled</code> false iken strateji ne olursa olsun hiçbir şey saklanmaz.</figcaption>
</figure>

Son sütunun istisnaları vardır. Bunların her birinde response, strateji ne olursa
olsun kendi `Cache-Control`'üyle gider:

- **[Guard'lı bir page](/docs/pages-and-layouts#private-pages-guards)**
  `private, no-cache` ile gider. Guard cache okunmadan önce çalıştığı için
  okuyucuları bu cache'i paylaşır, ama sunucunun önündeki bir CDN ya da proxy guard
  çalıştırmaz. Okuyucunun kendi tarayıcısı page'i saklayıp ETag'iyle yeniden
  doğrulayabilir; yeniden doğrulama da guard'ın gördüğü bir request'tir.
- **Form'un `{{csrfToken}}` değerini taşıyan bir page** `private, no-store` ile
  gider. Bu page de diğerleri gibi cache'lenir, ama her okuyucuya kendi token'ı
  gönderilir. Ayrıntılar için
  [Form'lar ve action'lar](/docs/forms-and-actions#pages-with-forms-are-still-cached)
  sayfasına bakın.
- **Degraded bir render** ([aşağıya](#what-is-cached-and-when) bakın) v0.34.0'dan
  beri `no-store` ile gider. Bu sunucu onu saklamaz; dolayısıyla önündeki bir CDN
  de onu page'in TTL'i boyunca saklamamalıdır. Başarısız bir render'ın yerine
  sunulan stale bir page de öyle gider (v0.63.0'dan beri).
- **Development'ta** her page ve document, v0.34.0'dan beri `no-store` ile gider.
  Böylece tarayıcı, düzenlemenizden önceki bir page'i tutmaz.

```go
page := collage.NewPage("blog-post").
	WithLayouts(layout).
	WithContent(post).
	WithPath("en", "/blog/{slug}").
	Incremental(10 * time.Minute).
	WithDependencyTags("blog:posts").
	Build()
```

`Static()` için pratikte bir sona erme süresi yoktur ve `DefaultTTL` ona uygulanmaz.
Page yalnızca tag'lerinden biri invalidate edildiğinde ya da cache yer açmak için
onu çıkardığında yeniden render edilir. Bu render başarısız olursa cevabı
invalidate edilmiş kopya verir (bkz. [Stale page'ler](#stale-pages-and-x-collage-cache)).

Yalnızca birisi yayımladığında değişen içerik için doğru strateji budur.
Gönderdiği `Cache-Control`, tarayıcılara ve CDN'lere page'i tutmalarını ama her
seferinde sormalarını söyler. [ETag](#etags-and-304) de bu sormayı ucuz hâle
getirir.

`Incremental(ttl)`, zamana bağlı olarak değişen içerik içindir. Ne zaman değiştiğini
size bildiremeyen bir kaynaktan gelen içerik için de uygundur. Sıfır TTL bir
register hatasıdır (`collage.ErrInvalidTTL`), negatif TTL de öyledir. İkisi de
sessizce hiç sona ermeyen bir page'e dönüşmez.

[Document'lar](/docs/documents), yani sitemap'ler, feed'ler ve HTML olmayan her şey
aynı üç çağrıyı kullanır. Page'lerle aynı şekilde saklanır, key'lenir ve
invalidate edilir. v0.12.0'dan beri de aynı şekilde
[birleştirilir](#concurrent-misses-render-once).

### Hiçbirini tanımlamayan bir page

Bu üç çağrıdan hiçbirini yapmayan bir page'in stratejisi register edilirken
belirlenir. O ana kadar stratejisi `collage.StrategyAuto`'dur, sonrasında diğer
üçünden biri olur.

- Render ettiği herhangi bir şey her render'da veri çekiyorsa **dynamic** olur. Bu,
  bir data handler (`collage.LoadTagged`, `collage.Load` ya da `collage.Effect`) ya
  da bir slot resolver'dır. Layout'unda, content'inde, bunların slot'larına
  bağlanan herhangi bir şeyde, fallback'lerinde ya da `WithFragmentPath` ile
  açılan bir fragment'te olması fark etmez.
- Aksi hâlde **static** olur. Yalnızca template'leri ve sabit değerleri
  (`collage.Value(v)`, `WithTitle(s)`) render eden bir page her okuyucu için aynı
  render edilir. Bu yüzden böyle page'lerden oluşan bir site, her birine
  `Static()` yazmadan cache'lenir ve [export edilir](/docs/static-export).

Handler'ın dynamic anlamına gelmesinin nedeni şudur: handler request'i, bir
cookie'yi ya da saati okuyabilir ve fonksiyonun dışından bunu yapıp yapmadığı
anlaşılamaz. Tahmin yanlış çıktığında bedeli bir render olur; bir okuyucuya
hiçbir zaman başkasının page'i sunulmaz. Üzerinde düşünmeyi
unuttuğunuz bir page yalnızca yavaş olur, stale olmaz.

Bir handler'ın çıktısı herkes için aynıysa (dosyadan okunan bir yazı gibi), page'i
`Static()` ya da `Incremental(ttl)` çağrısını kendisi yapar.

Bir page'i dynamic yapan şey, page'den page'e farklı olmak değildir. Path ve
parametreleri cache key'in bir kısmıdır ve export'ta her URL kendi dosyasıdır:
`/recipes/pancakes` ile `/recipes/omelette` zaten ayrı tutulur. Bir page'i dynamic
yapan şey, *aynı* URL'ye gelen iki request arasında değişen çıktıdır: cookie'den
okunan bir kullanıcı ya da saat gibi.

Handler'ı yalnızca path'in parametrelerini ve locale'i okuyan bir fragment bunu
kendisi söyleyebilir. Fragment'e `Static()` yazıldığında, o fragment page'i dynamic
yapmaz. Bu, v0.17.0'dan beri böyledir.

```go
more := collage.NewFragment("more-recipes", "fragments/more-recipes.html").
	WithData(collage.LoadTagged(loadMore)). // every recipe but rc.Param("slug")
	Static().
	Build()
```

Bu, birçok page'in paylaştığı bir fragment içindir: son yazıların listesi ya da
içerikten oluşturulan bir navigasyon gibi. Böylece onu kullanan her page'in aynı
sözü tekrarlaması gerekmez. Bu söz, `Static()`'in bir page için verdiği sözün
aynısıdır ve onu bozan bir handler bir okuyucunun render'ını bir sonrakine sunar.

Yalnızca fragment'in kendi handler'ını kapsar:

- başka bir fragment'in handler'ı page'i yine dynamic yapar;
- static bir fragment'te olsa bile bir slot resolver da aynı şeyi yapar, çünkü
  döndüreceği fragment'ler bir render onları isteyene kadar bilinmez.

`Static()` aynı anda iki söz verir: çıktı her okuyucu için aynıdır ve bir şey
invalidate edilene kadar da aynı kalır. Hiçbir cookie okumayan ama bir ölçüm
döndüren bir handler (CPU yükü, bir kuyruğun uzunluğu, saat gibi) ilk sözü tutar,
ikincisini tutmaz. Çıktısı her okuyucu için aynıdır ama zaman içinde aynı kalmaz.
Böyle bir fragment'i bunun yerine `Shared()` ile işaretleyin. `Shared()`
v0.19.0'dan beri vardır.

```go
cpu := collage.NewFragment("cpu", "fragments/cpu.html").
	WithData(collage.Load(cpuUsage)). // the same for everyone, different every second
	Shared().
	Build()
```

`Shared()` page'in stratejisine dokunmaz. Page dynamic kalır ve bir export, tek bir
anın ölçümünü page'e yazmaz. Fragment'i okuyuculara gönderen şeyin de onu her
değişiklikte bir kez render etmesini sağlar: v0.78.0'dan beri collage, middleware'i
ve guard'ları kabul eden her okuyucuya o tek render'ın bir kopyasını verir (bkz.
[Plugin yazmak](/docs/writing-plugins#pushing-fragments)). `Static()` bunu zaten
içerir. `Static()`'te olduğu gibi, sözünü bozan bir handler bir okuyucunun
verisini başka bir okuyucuya gönderir.

Bir page'in tanımlanmış stratejisi hiçbir yönde sorgulanmaz: `Dynamic()` diyen bir
page, fragment'leri ne derse desin dynamic kalır. v0.16.0'a kadar hiçbir şey
tanımlamayan bir page dynamic'ti. Buna güvenen ve onu dynamic yapacak bir
handler'ı olmayan bir page artık `Dynamic()` çağrısını yapmalıdır.

Bir [document](/docs/documents#a-fixed-body) da aynı şekilde belirlenir:
handler'ı varsa dynamic, sabit bir body'si varsa static olur.

## Ne, ne zaman cache'lenir

Bir response, page cache'e yalnızca şu koşulların hepsi sağlandığında girer:

- cache açıktır ve page'in stratejisi, ister tanımlanmış ister belirlenmiş olsun,
  static ya da incremental'dır;
- request bir `GET`'tir. Bir `HEAD` cache'ten *sunulabilir*, ama cache'i hiçbir
  zaman doldurmaz, çünkü saklanacak bir body üretmemiştir;
- her fragment başarıyla render edilmiştir.

Son kural göründüğünden daha önemlidir. Başarısız olan bir fragment render'ı
**degraded** yapar. Fragment optional olsa, hatta fallback'i onun yerini doldurmuş
olsa bile bu değişmez. Degraded bir render sunulur ama saklanmaz. Saklansaydı, tek
bir request'teki geçici bir hata TTL dolana kadar her okuyucunun karşısına çıkardı.
Page'in stale bir kopyası tutuluyorsa onun yerine o kopya sunulur (bkz.
[Stale page'ler](#stale-pages-and-x-collage-cache)).

Hata response'ları da hiçbir zaman cache'lenmez. 404 ya da 500 `no-store` ile
yazılır ve bir error page'in kendi render'ında bildirilen tag'ler atılır. Bir 404,
request'in bulduğu stale entry'yi de düşürür.

Bir [action'ın](/docs/forms-and-actions) response'u da hiçbir zaman cache'lenmez. Render
ettiği page nasıl tanımlanmış olursa olsun bu değişmez, çünkü o response tek bir
gönderimden üretilmiştir ve onu gönderen kişiye aittir.

Bir [asset mount'undan](/docs/assets) sunulan dosyalar page cache'ten hiç geçmez.
Ne kadar taze kalacaklarını yalnızca `Cache-Control` header'ları belirler.

## Cache key

Cache'lenmiş bir page, şu kısımlardan oluşan bir key ile bulunur:

- request path'i,
- çözümlenen locale,
- yakalanan path parametreleri,
- query string ([aşağıya](#query-parameters-in-the-key) bakın),
- middleware'inizin `collage.Vary` ile bildirdiği her değer (bir handler'da
  `collage.Varied` ile okunur),
- request'in host'u (v0.34.0'dan beri), normalleştirilmiş olarak (v0.79.0'dan
  beri): küçük harfle, portsuz ve sondaki noktasız; bir IPv6 adresi de köşeli
  parantez içinde, kanonik biçimiyle.

Key'i aynı olan iki request, collage açısından aynı page'dir.

Host key'de yer alır, çünkü bir render `r.Host`'tan mutlak bir URL oluşturabilir:
bir canonical link, bir `og:url`. Host key'de olmasaydı, başka bir host adı veren
tek bir request bu URL'yi her okuyucuya sunulan kopyaya yazardı.

Normalleştirilir, çünkü nasıl yazılacağını client seçer. `EXAMPLE.com:8080`,
`example.com.` ve `example.com` tek bir sitedir; `[0:0::1]:9` ve `[::1]` de öyle.
v0.79.0'dan önce her yazım kendine ait bir render ve entry'ydi: herkesin okuduğu
page'leri sınırlı bir cache'ten dışarı itmenin ucuz bir yolu. Paylaşılan render'a
da aynı normalleştirilmiş host verilir; böylece bir request'in kullandığı yazım,
bir başkasına sunulana asla yazılmaz. **Cacheable bir page'in data handler'ı artık
`r.Host`'ta portu görmez.** Scheme'in varsayılanı olmayan bir portta sunulan site,
mutlak URL'lerini portu koruyan `Config.BaseURL`'den, `collage.BaseURL(rc)` ile
kurar. Dynamic bir page `Host`'u gönderildiği gibi görür, static bir build'in header
yakalaması da. Diskte zaten duran entry'ler okunabilir kalır.

### Paylaşılan bir render neyi görür

Cache'lenmiş bir render, request'i aynı key'e sahip her okuyucuya sunulur. Bu
yüzden v0.34.0'dan beri render'a yalnızca key'in içerdikleri verilir: path, host
(key'de durduğu gibi, normalleştirilmiş), key'deki query parametreleri ve middleware'inizin `collage.Vary` ile bildirdiği
header'lar. Okuyucunun cookie'leri, `Authorization`'ı, adresi ya da client
sertifikası verilmez. `WithCacheParams`'ın dışarıda bıraktığı bir query parametresi
de verilmez.

Static ya da incremental bir page'deki bir handler cookie okumak istediğinde hiçbir
cookie bulamaz. Eskiden ilk okuyucunun cookie'sini bulur ve onu, ondan sonra gelen
herkesin aldığı kopyaya yazardı. Request'in context'i de aynı muameleyi görür:
middleware'inizin oraya koyduğu bir değer, yani giriş yapmış bir kullanıcı, bir
tenant ya da bir request id, cookie ile aynı nedenden ötürü paylaşılan bir render'dan
gizlenir.

Middleware'inizin `collage.Vary` ile bildirdiği bir değeri
`collage.Varied(ctx, header)` ile geri okuyun. `ctx`, bir data handler'ın, bir page'in
`StaticParams`'ının, bir sitemap'in `LastMod`'unun ya da bir document handler'ın
aldığı context'tir (v0.42.0'dan beri; v0.60.0'dan önce `RenderContext` alıyordu ve
context alan sürüm `VariedContext` idi). Orada güvenlidir, çünkü key'in içindedir
ve her değerin kendi cache'lenmiş page'i vardır.

Bir cache boyutu olmayan bir değerin paylaşılan bir render'da yeri yoktur:
okuyucunun kendi request'ine ihtiyaç duyan bir page `Dynamic()` olmalıdır ve
dynamic bir page request'in tamamını görür.
Development'ta, bir handler'ın gerçekten okuduğu gizlenmiş bir context değeri
log'lanır, böylece hata görünür olur. Document'lar da aynı kurala uyar
(v0.41.1'den beri; öncesinde cacheable bir document'ın render'ı ilk okuyucunun
context değerlerini hâlâ görüyordu).

### Origin'ler

`collage.BaseURL(rc)` paylaşılan bir render'da güvenlidir (v0.42.0'dan beri). Host
cache key'in içindedir; bu yüzden bir kopyanın sunulduğu her okuyucu aynı host'u
istemiştir ve render'ın mutlak URL'leri kurduğu origin de onlarındır.

Static bir build'de request yoktur; bu yüzden `Config.BaseURL` kullanılır.
`OriginResolver` plugin'i olmayan bir site hiçbir fark görmez; bkz.
[Bir host'un origin'i](/docs/writing-plugins#which-origin-a-host-has-originresolver).
Bir invalidation entry'lere ulaştığında `CacheInvalidateEvent.Entries` her birinin
path'ine ek olarak host'unu da, key'de durduğu gibi normalleştirilmiş olarak,
adlandırır.

## Dependency tag'ler

Tag, bir içeriği adlandıran herhangi bir string'dir: `post:hello-world`,
`author:ada`, `blog:posts`. Bir page'in tag'leri iki yerden gelir ve birleştirilir:

```go
// From the page, for what it always depends on.
WithDependencyTags("blog:posts")

// From a data handler, for what this particular render used.
return view, []string{"post:" + post.Slug, "author:" + post.AuthorID}, nil
```

Page'in kendi tag'leri daha hiçbir şey render edilmeden bilinir. Asıl işe yarayanlar
data handler'dan gelenlerdir, çünkü bu URL'nin sonunda hangi yazıyı ve hangi yazarı
gösterdiğini yalnızca handler bilir. Handler sonrasında bir hata dönse bile tag'ler
toplanır, çünkü handler page'in neye bağlı olduğunu yine de söylemiştir.

Page saklandığında tag'leri de onunla birlikte saklanır. collage ayrıca her tag'den
hangi cache key'lerinin üretildiğini kaydeder.

### Invalidate etmek

İçerik değiştiğinde, değişen içeriğin adını verin:

```go
if err := app.InvalidateTags(ctx, "post:"+slug, "blog:posts"); err != nil {
	log.Printf("invalidate: %v", err)
}
```

Bu tag'lerden herhangi birini taşıyan her cache'lenmiş page stale olarak
işaretlenir (v0.63.0'dan beri; önceden atılırdı). Artık taze diye sunulmaz:

- hemen arka planda yeniden render edilir (v0.64.0'dan beri;
  [aşağıya](#warming-after-an-invalidation) bakın). Bu henüz olmadıysa onu gelen
  bir sonraki request render eder. Başarılı bir render eskisinin yerine saklanır;
- ama eski page tutulur, çünkü render başarısız olur ya da degraded çıkarsa cevap
  onunla verilir. Bkz. [Stale page'ler](#stale-pages-and-x-collage-cache).

`collage.Cached` ile bu tag'ler altında saklanan her değer de stale olarak
işaretlenir. Değer tutulur ve arka plandaki bir çekme işlemi onun yerini alırken bir
kez daha sunulur (bkz. [Nasıl davranır](#how-it-behaves)). Başka hiçbir şeye
dokunulmaz.

Bu, built-in store'lar için, yani memory ve disk için geçerlidir. Kendi store'unuz
[`collage.StaleCache`](#stale-pages-in-your-own-store)'i implement etmedikçe
`InvalidateTags`'te entry'leri düşürmeye devam eder.

`App.PurgeTags` aynı tag'leri alır ve onlara bağlı her şeyi, diskteki kopyalar
dahil, stale olarak sunulacak bir şey bırakmadan siler:

```go
if err := app.PurgeTags(ctx, "post:"+slug); err != nil {
	log.Printf("purge: %v", err)
}
```

Bunu, bir kez bile olsa bir daha gösterilmemesi gereken veri için kullanın. Bedeli
şudur: geri dönülecek bir şey kalmadığı için çekme işlemi ya da render başarısız
olursa page de başarısız olur.

<figure class="diagram-figure" data-search-ignore>
<svg class="diagram" viewBox="0 0 320 332" role="img" aria-labelledby="cache-tags-title">
<title id="cache-tags-title">Tek bir tag, author:ada, hem page cache'e hem data cache'e ulaşır. InvalidateTags ulaştığını stale olarak işaretler ve tutar; PurgeTags ise onu diskteki kopyası dahil tamamen siler. İki durumda da ulaşılan page'ler arka planda warm edilir. Sonrasında bir render başarısız olursa invalidate edilmiş bir page stale olarak sunulur, purge edilmiş bir page ise başarısız olur.</title>
<defs><marker id="cache-tags-arrow" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto"><path class="d-head" d="M0 0 8 4 0 8z" /></marker><marker id="cache-tags-hit" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto"><path class="d-head-hit" d="M0 0 8 4 0 8z" /></marker></defs>
<rect class="d-chip-hit" x="110" y="8" width="100" height="24" rx="12" /><text class="d-hit-text" x="160" y="24" text-anchor="middle">author:ada</text>
<path class="d-hitline" d="M140 32 84 58" marker-end="url(#cache-tags-hit)" />
<path class="d-hitline" d="M180 32 236 58" marker-end="url(#cache-tags-hit)" />
<rect class="d-a" x="10" y="62" width="140" height="40" rx="5" /><text class="d-on-a" x="80" y="78" text-anchor="middle">page cache</text><text class="d-on-a d-small" x="80" y="93" text-anchor="middle">/authors/ada, yazılar</text>
<rect class="d-b" x="170" y="62" width="140" height="40" rx="5" /><text class="d-on-b" x="240" y="78" text-anchor="middle">data cache</text><text class="d-on-b d-small" x="240" y="93" text-anchor="middle">Cached author:ada</text>
<text class="d-small" x="160" y="122" text-anchor="middle">tek çağrı ikisine de ulaşır</text>
<rect class="d-cmd" x="10" y="132" width="140" height="24" rx="12" /><text class="d-cmd-text" x="80" y="148" text-anchor="middle">InvalidateTags</text>
<rect class="d-cmd" x="170" y="132" width="140" height="24" rx="12" /><text class="d-cmd-text" x="240" y="148" text-anchor="middle">PurgeTags</text>
<line class="d-flow" x1="80" y1="156" x2="80" y2="172" marker-end="url(#cache-tags-arrow)" />
<line class="d-flow" x1="240" y1="156" x2="240" y2="172" marker-end="url(#cache-tags-arrow)" />
<rect class="d-frame" x="10" y="176" width="140" height="40" rx="6" /><text class="d-small" x="80" y="193" text-anchor="middle">stale işaretlenir,</text><text class="d-small" x="80" y="206" text-anchor="middle">fallback olarak tutulur</text>
<g class="d-dropped"><rect class="d-frame" x="170" y="176" width="140" height="40" rx="6" /><text class="d-small" x="240" y="193" text-anchor="middle">tamamen silinir</text><text class="d-small" x="240" y="206" text-anchor="middle">diskten de</text></g>
<line class="d-flow" x1="80" y1="216" x2="80" y2="232" marker-end="url(#cache-tags-arrow)" />
<line class="d-flow" x1="240" y1="216" x2="240" y2="232" marker-end="url(#cache-tags-arrow)" />
<rect class="d-soft-a" x="10" y="236" width="140" height="28" rx="6" /><text class="d-small" x="80" y="254" text-anchor="middle">arka planda warm edilir</text>
<rect class="d-soft-a" x="170" y="236" width="140" height="28" rx="6" /><text class="d-small" x="240" y="254" text-anchor="middle">arka planda warm edilir</text>
<line class="d-flow" x1="80" y1="264" x2="80" y2="280" marker-end="url(#cache-tags-arrow)" />
<line class="d-flow" x1="240" y1="264" x2="240" y2="280" marker-end="url(#cache-tags-arrow)" />
<rect class="d-frame" x="10" y="284" width="140" height="40" rx="6" /><text class="d-small" x="80" y="301" text-anchor="middle">render başarısız:</text><text class="d-small" x="80" y="314" text-anchor="middle">eski kopya, STALE</text>
<rect class="d-frame" x="170" y="284" width="140" height="40" rx="6" /><text class="d-small" x="240" y="301" text-anchor="middle">render başarısız:</text><text class="d-small" x="240" y="314" text-anchor="middle">page başarısız olur</text>
</svg>
<figcaption>Page'ler ve <code>collage.Cached</code> değerleri aynı tag'leri taşır, bu yüzden tek bir çağrı ikisine de ulaşır. İki çağrı arasındaki fark geride ne kaldığıdır: <code>InvalidateTags</code> eski page'i ve değeri geri dönülecek bir şey olarak tutar, <code>PurgeTags</code> hiçbir şey tutmaz. İkisi de ulaştığı page'leri warm eder. <code>collage.StaleCache</code> olmayan kendi store'unuz <code>InvalidateTags</code>'te de entry'lerini düşürür.</figcaption>
</figure>

`InvalidateTagsN` aynı işi yapar ve ayrıca tag'lerin kaç cache key'ine ulaştığını
söyler:

```go
reached, err := app.InvalidateTagsN(ctx, "author:ada")
```

Bu sayı, collage'ın kendi tag index'inin çözümleyip invalidate ettiği key'lerin
sayısıdır. Gerçekten ulaşılan canlı page'lerin tam sayısı değildir ve iki yönde de
sapabilir:

- entry'si zaten expire olmuş bir key yine sayılır;
- store'un kendi tag index'i üzerinden ulaştığı bir entry ise sayılmadan
  invalidate edilir. Bu, örneğin
  [aşağıda](#the-tag-index-is-per-process-and-bounded) anlatılan sınırın ötesinde
  kalan ya da başka bir instance'ın yazdığı bir entry olabilir. Bir
  [`collage.SharedCache`](#a-cache-a-fleet-shares) bunun istisnasıdır: onun kendi
  index'inin ulaştığı page'ler, hangi instance yazmış olursa olsun, sayılır
  (v0.66.0'dan beri).

Sayı bir log satırı ya da bir metrik için uygundur, uygulama mantığı için değil.

Cache bazı key'leri invalidate edemezse geri kalanlar yine invalidate edilir ve
hatalar birleştirilip tek bir hata olarak döner. Başarılı görünen kısmi bir
invalidation, stale page'lerin bir deploy'dan sağ çıkmasına yol açar.

Bunu genellikle içeriğinizin değiştiği yerde çağırırsınız: bir CMS webhook'unda ya
da bir admin form'unda (baştan sona bir webhook örneği
[aşağıdadır](#example-author-pages-on-a-blog)).

Bir action bunu declarative olarak da yapabilir. Bunun için
[result'ındaki `InvalidateTags`](/docs/forms-and-actions#invalidating-what-an-action-changed)
alanını kullanır. Bu invalidation response yazılmadan önce çalışır. Böylece az önce
değiştirdiği page'e redirect edilen okuyucuya eski page taze bir page gibi
sunulmaz. v0.64.0'dan beri redirect'in kendi page'i, redirect yazılmadan önce
yeniden render edilir (bkz. [Warming](#warming-after-an-invalidation)). Eski page
yalnızca bu render başarısız olursa ya da yavaş kalırsa sunulur. Bkz.
[oradaki not](/docs/forms-and-actions#invalidating-what-an-action-changed).

`app.InvalidateTags`'i handler'dan kendiniz çağırırsanız invalidation'ın page'leri
yalnızca arka planda warm edilir. Redirect'in page'i, redirect yazılmadan önce
render edilmez.

### Path'e göre invalidate etmek

Cache'lenen her page ve document ayrıca render edildiği URL path'ini adlandıran bir
tag'e de bağlıdır: `collage.PathTag(path)` (v0.23.0'dan beri). Bu yüzden elinizde
bir tag değil de path varsa, entry'yi path'iyle invalidate edebilirsiniz:

```go
app.InvalidateTags(ctx, collage.PathTag("/blog/hello"))
```

Bu, o path için cache'lenmiş olanı, başka neye bağlı olursa olsun invalidate eder.

Bir invalidation neye ulaştığını da söyler. Bir plugin her invalidation'ı
`OnCacheInvalidate` üzerinden duyar. Event tag'leri ve `Entries` içinde invalidate
edilen cache'lenmiş page ve document'ları, her biri bir host ve bir path olarak
taşır. Bir CDN'in purge etmesi ve bir arama motoruna değiştiği bildirilmesi gereken
budur. Hiç cache'lenmemiş bir page burada yer almaz. Onu implement eden her plugin,
kendisinden önceki biri başarısız olsa bile çağrılır (v0.78.0'dan beri) ve hatalar
`InvalidateTags`'ten join edilmiş olarak döner. Invalidation'ı tetiklemek uygulamanın
işidir: v0.78.0'dan beri bir plugin'in kendi `InvalidateTags`'i yoktur. Bkz. [Plugin
yazmak](/docs/writing-plugins#cacheinvalidatehook).
[elagoht/cdnpurge](/docs/plugins#elagohtcdnpurge) ve
[elagoht/indexnow](/docs/plugins#elagohtindexnow) bunun üzerine kuruludur.

### Tag index'i process başınadır ve sınırlıdır

collage'ın hangi key'in hangi tag'e ait olduğunu tuttuğu kayıt memory'de, yani o
key'leri yazan process'in içinde durur. Bunun iki sonucu vardır.

**Load balancer arkasında her instance yalnızca kendi key'lerini bilir.** Built-in
cache'lerde bu sorun olmaz, çünkü her instance'ın kendi cache'i de vardır.

Kendi paylaşılan store'unuzu kullanıyorsanız (`Cache.Store`, örneğin bir Redis),
`collage.TaggedCache`'i de implement edin. Böylece store tag'leri kendisi indeksler.
collage bu durumda tag'e göre invalidation'ı store'dan da ister ve store başka bir
instance'ın yazdığı entry'lere de ulaşabilir. Built-in disk cache bunu yapar.
Tag'lerinin restart sonrasında da çalışmaya devam etmesinin nedeni budur.

[`collage.SharedCache`](#a-cache-a-fleet-shares) olan bir store (v0.66.0'dan
beri) bu açığı tamamen kapatır: onun kendi tag index'i herhangi bir instance'ın
yazdığı her page'i ve veri değerini listeler. Böylece sayı,
`CacheInvalidateEvent.Entries` ve warming yalnızca bu process'in değil, filonun
page'lerini kapsar.

**Tag başına tutulan kayıt `Cache.MaxEntries` ile sınırlıdır** (varsayılan
10000; v0.60.0'dan önce bu ayrı bir alandı: `MaxKeysPerTag`). Query string key'in
bir kısmı olduğundan, bir client tek bir page için istediği kadar key üretebilir.
Bir sınır olmasaydı index sonsuza kadar büyürdü. Bir tag sınıra ulaştığında, o tag
altında kaydedilmiş en eski key bu index'ten silinir. Bunun bedeli store'a göre
değişir:

- Built-in memory ve disk cache'lerinin ikisi de `collage.TaggedCache`'i implement
  eder ve her entry'nin tag'lerini kendileri kaydeder. Bu yüzden onlarla tag'i
  invalidate etmek yine her entry'ye ulaşır. Yalnızca `InvalidateTagsN`'in
  bildirdiği sayı eksik kalır.
- `TaggedCache`'i implement etmeyen kendi store'unuzda ise elde yalnızca
  collage'ın index'i vardır. Index'ten silinen bir page, expire olana kadar
  store'da kalır ve tag'i invalidate etmek artık ona ulaşmaz.

`MaxEntries`'i, bir tag'in gerçekte kapsayabileceği cache'lenmiş URL sayısının
üzerinde bir değere ayarlayın. Negatif bir `MaxEntries` bu sınırı 10000'de tutar.

**Built-in bir cache'le index bir bütün olarak da sınırlıdır** (v0.40.0'dan beri),
`MaxEntries` kadar key'le. Cache'lenmiş her path kendi başına bir tag'dir. Bu yüzden
yalnızca tag başına bir sınır varken index, bu sınırın page sayısıyla çarpımına
ulaşabiliyordu. Uydurma query'lerle doldurulan büyük bir sitede bu, her page için
10000 key demekti; oysa cache'in kendisi toplamda 10000 entry tutuyordu.

Hâlâ canlı olabilecek key sayısı `MaxEntries`'i geçemez; bu yüzden önce en eski
yazılanlar gider. Negatif bir `MaxEntries` bunu kapatır. Kendi store'unuz yalnızca
tag başına sınırı korur, çünkü collage onun boyutunu bilemez.

## Bir invalidation'dan sonra warming

Invalidate edilmiş bir page, bir şey onu yeniden render edene kadar stale'dir. Bu
render'ın bedelini de ilk ziyaretçi öderdi. Bu yüzden v0.64.0'dan beri her
`InvalidateTags` ve `PurgeTags` çağrısından sonra, invalidation'ın ulaştığı page ve
document'lar arka planda taze veriyle yeniden render edilir ve saklanır. Çağrının
sizden, bir plugin'den ya da bir action'dan gelmesi fark etmez. Böylece sonraki
ziyaretçiye taze bir `HIT` verilir.

Bu işi dört worker yapar. İlk key bir probe olarak tek başına önce gider. Worker'lar
ancak ilk başarılı warm'dan sonra başlar. Bu yüzden atlanan key'ler varken tur, biri
başarılı olana kadar key key ilerler.

- **Yalnızca sunulmuş olan warm edilir.** Cache'e yazılan bir entry, kendisinin
  yeniden nasıl isteneceğini kaydeder: path'ini, query'sini, locale'ini ve `Vary`
  header değerlerini. Bir ziyaretçiye gerçekten sunulmuş her varyant warm edilir,
  kimsenin istemediği bir varyant warm edilmez. Bu tarifler yalnızca memory'de durur.
  Restart'tan sonra page'ler yeniden yazılana kadar hiç tarif yoktur ve ilk
  invalidation bu durumda hiçbir şeyi warm etmez.
- **Yalnızca hâlâ okunan yeniden warm edilir.** Bir page, son warm'ından beri bir
  ziyaretçiye ulaştıysa warm edilir: ziyaretçiye taze ya da stale sunulduysa veya
  ziyaretçinin request'i onu render ettiyse. Warm edilip sonra okunmayan bir page
  sonraki invalidation'da atlanır ve bir sonraki request'inde warming yokmuş gibi
  render edilir. O ziyaretçi, page'i yeniden warm etmeye değer kılar. Birinin bir kez
  istediği bir varyant sonsuza kadar warm tutulmaz.
- **Cache'in bıraktığı geri gelmez.** Bir `InvalidateTags`'in ulaştığı ama entry'sini
  cache'in artık stale olarak bile tutmadığı bir page geri getirilmez. Böyle bir
  entry `MaxEntries`'te ya da `MaxBytes`'ta evict edilmiş ya da düşürülmüştür.
  `PurgeTags` ise entry'lerini kendisi siler ve onları yine de warm eder.
  (`StaleCache` olmayan kendi store'unuz invalidate ettiği her entry'yi düşürür. Bu
  yüzden bu kontrol ona uygulanmaz.)
- **`Cache.WarmLimit` tek bir invalidation'ı sınırlar.** Sıfır 1000 demektir ve en
  son yazılan page'ler önce gelir. Negatif bir değer warming'i kapatır. Sınırın
  ötesindeki page'ler eskisi gibi ilk request'lerinde render edilir.
- **Bir tur ilk gerçek render başarısızlığında durur** ve warn seviyesinde tek bir
  satır log'lar (`collage: warming stopped, a page failed to render`). Böylece kapalı
  bir upstream'e kalan her page için yeniden sorulmaz. Kalan page'ler stale bırakılır
  ve onlara her zamanki gibi [stale sunma](#stale-pages-and-x-collage-cache)
  uygulanır. Atlanan page'ler turu durdurmaz: 404 dönen bir page (tarifi düşürülür),
  guard'lı bir page (tarifi düşürülür ve onu ilk okuyucusu render eder), zaten taze
  olan bir page ve yukarıdaki iki durum.
- **Başka bir invalidation'ın çakıştığı bir warm yeniden warm edilir.** Başka bir
  `InvalidateTags` ya da `PurgeTags` ile çakışan render'lar saklanmaz. Bu kural warm
  render'lar için de geçerlidir. Warmer o page'i, süren bir invalidation kalmadığında
  yeniden kuyruğa alır. Bunu, onunla çakışan her invalidation için en fazla bir kez
  yapar. Böylece bir invalidation fırtınası tek bir page'i durmadan render ettiremez.
- **Warm render bir request değildir.** `App.Use` ve plugin middleware'lerini,
  request hook'larını, kişiselleştirmeyi, error hook'larını ve request başına
  log'ları atlar. `BeforeRender`, `AfterRender` ve `DocumentRendered` ise saklanan
  body'yi şekillendirdikleri için yine çalışır. Warm render kendi flight'ında
  çalışır. Bu yüzden hiçbir ziyaretçi bir warm render'ı beklemez ve warm edilirken
  bir page iki kez render edilebilir.
- **Warming development'ta, static build'de ve cache kapalı olduğunda çalışmaz.**
  `Shutdown` warming'i durdurur. Hâlihazırda süren render'lar shutdown context'iyle
  tamamlanır.

Bu kurallardan şunlar çıkar:

- **Bir kimlik bilgisi header'ını asla `Vary` header'ı olarak bildirmeyin.** Bir
  warm request hiçbir ziyaretçinin `Cookie`'sini ya da `Authorization`'ını taşımaz.
  Cache key'in dayandığı header'ların değerlerini taşır. Orada `Cookie` ya da
  `Authorization` bildirirseniz değeri tarifte, memory'de tutulur ve o page'in her
  warm render'ında yeniden gönderilir. Bunun yerine bir middleware'in ondan
  çıkardığı değere göre (bir plan, bir rol) vary edin ya da page'i `Dynamic` yapın.
- **Middleware warm render'ı engellemez.** Bir request'i reddeden bir middleware,
  bir warm'ı hiç görmez. Silinmiş bir tenant'ın host'u için 404 dönen bir tenant
  plugin'i ya da bir auth kontrolü buna örnektir. Bu yüzden `PurgeTags`'ten sonra
  warming, artık hiçbir ziyaretçinin ulaşmadığı bir entry yazabilir. Bu zararsızdır,
  çünkü kimseye sunulmaz ve her entry gibi evict edilir. Böyle bir host'un
  page'lerini, middleware onun için cevap vermeyi bıraktıktan sonra purge edin ya
  da süreleri dolana kadar bekleyin.
- **Bir turun dışında kalan page, yeniden yazılana kadar izlenmez.** Bir
  invalidation ulaştığı key'leri unutur. Bir turun warm etmediği bir key,
  `WarmLimit`'in ötesinde kalmış, bir başarısızlıkta ya da `Shutdown`'da duran bir
  turdan geriye kalmış ya da atlanmış olabilir. Bir ziyaretçinin request'i onu
  yeniden render edip saklayana kadar sonraki hiçbir invalidation ona ulaşmaz. Bu
  arada stale entry'si her zamanki gibi sunulur (ya da yeniden render edilir).
- **Gereksiz query varyantlarını `WithCacheParams` dışarıda tutar.** Hiçbirini
  tanımlamayan bir page'de her farklı query bir key'dir. Bir ziyaretçiye sunulan
  her key de hâlâ okunduğu sürece warm edilir. Query'si her şeyi taşıyabilen
  (tracking parametreleri, cache-buster'lar) bir page,
  [okuduğu parametrelerin adını vermelidir](#query-parameters-in-the-key). Böylece
  gereksiz bir query, temiz URL ile aynı key olur.

Warm edilen key'ler yeniden izlenir. Bu yüzden sonraki bir invalidation'ın
`OnCacheInvalidate`'i onları yine adlandırır. Bir tur, info seviyesinde tek bir
satır log'lar (`collage: warmed pages`; warm edilen, atlanan ve başarısız olan
sayılarıyla). Warm edilen her key bir `CacheWarm` metrics event'idir.

Bir warm request yine `CacheHit`, `CacheMiss` ve `CacheSet`'te sayılır. Render'ı da
`RenderDuration`'da ve fragment metriklerinde sayılır. Bu yüzden bir invalidation'dan
sonra bir miss, set ve render patlaması bekleyin. Warming, bir invalidation'ın hemen
ardından CPU da harcar: binlerce page'i aynı anda invalidate eden bir site onları
dörder dörder yeniden render eder. Bu önemliyse `WarmLimit`'i düşürün.

Bir invalidation'ın hemen ardından `STALE` geldiğini doğrulayan bir test warmer ile
yarışır. Böyle bir testte `WarmLimit: -1` ayarlayın. Invalidation'dan sonra `HIT`
geldiğini doğrulayan bir test ise önce `collagetest`'in `site.WaitWarmed()`'ı ile
warmer'ı bekler (bkz. [Test yazmak](/docs/testing#invalidation-and-warming)).

Invalidate edip redirect eden bir action, önce redirect'in page'ini warm eder. Bu
bekleme 5 saniyeyle ve response'unun write deadline'ıyla sınırlıdır. Bkz.
[Form'lar ve action'lar](/docs/forms-and-actions#invalidating-what-an-action-changed).

## Stale page'ler ve `X-Collage-Cache`

v0.63.0'dan beri stale olan bir page ya da document, built-in store'lar tarafından
atılmaz. Stale olması `InvalidateTags` ile işaretlendiği ya da TTL'i geçtiği
anlamına gelir. (Bir invalidation'dan sonra böyle bir page normalde, biri onu
istemeden önce [warm edilir](#warming-after-an-invalidation). Aşağıda anlatılanlar,
warm render başarısız olduğunda, atlandığında ya da henüz çalışmadığında bir
ziyaretçinin karşılaştığı durumdur.)

Onu isteyen bir request page'i yeniden render eder. Bu render, her miss'te olduğu
gibi bekleyen request'ler arasında paylaşılır
([aşağıya](#concurrent-misses-render-once) bakın). Sonrasında ne olacağını bu
render belirler:

- **Başarılı bir render** stale page'in yerine saklanır ve sunulur. Response bir
  `MISS`'tir.
- **Başarısız olan ya da degraded çıkan bir render** saklanmaz. Onun yerine stale
  page sunulur. Response `X-Collage-Cache: STALE` taşır ve hiçbir tarayıcı ya da
  CDN onu tutmasın diye `Cache-Control: no-store` ile gider. Okuyucu 500 değil,
  eski page'i görür.
  Stale bir `collage.Cached` değerini okumak render'ı degraded yapar. Warming
  yoksa bir invalidation'ın hemen ardından en sık neden de budur: arka plandaki
  çekme işlemi onu değiştirene kadar page'in verisi stale'dir.
- **404 dönen bir render** entry'yi düşürür, çünkü içerik artık yoktur.

v0.63.0'dan önce bir page invalidate edildiğinde tamamen giderdi. Bu yüzden bir
invalidation'ın ardından başarısız olan bir render 500 demekti. `PurgeTags` ise yine
geri dönülecek bir şey bırakmaz. Bu yüzden purge edilmiş bir page render edilemezse
başarısız olur.

Expire olmuş bir page de aynı anlamda stale'dir. Bu yüzden upstream'i kapalı olan
bir `Incremental` page, TTL'i dolduktan sonra da son başarılı render'ıyla cevap
vermeyi sürdürür. Stale bir page için bir üst yaş sınırı yoktur. Onu yalnızca
`MaxEntries` ya da `MaxBytes`'ta yapılan eviction ortadan kaldırır.

Herhangi bir `InvalidateTags` ya da `PurgeTags` çağrısıyla çakışan bir render,
tag'leri ne olursa olsun saklanmaz. Çakışan render, bir çağrı sürerken başlamış ya
da kendisi sürerken bir çağrının başladığı render'dır. Böyle bir render
değişiklikten önceki veriyi okumuş olabilir. Onu saklamak, eski page'i taze diye
geri getirirdi.

**Stale sunumlar raporlanır; gizledikleri başarısızlık ise yalnızca bir kez
raporlanır.** Her stale response `CacheStale` metrics event'i olarak sayılır.
Gizlediği başarısızlık ise başarısız her render için bir kez raporlanır. O
render'ı kaç request'in beklediği bunu değiştirmez:

- warn seviyesinde bir kez log'lanır
  (`collage: serving a stale page, its render failed`, path ve hatayla birlikte).
  Hiçbir error page sunulmadığı için ayrıca error seviyesinde bir
  `collage: request failed` olarak log'lanmaz;
- plugin'lerin `OnError` hook'una da bir kez iletilir. Hiçbir hata response'u
  gönderilmediği için `Status` 0'dır.

Yalnızca stale `collage.Cached` verisi okuduğu için degraded olan bir render
bir kesinti değildir, çünkü verisi yenilenmektedir. Bu yüzden page'in stale
olarak sunulması debug seviyesinde log'lanır
(`collage: serving a stale page while its data refreshes`) ve `OnError`'a hiçbir
şey ulaşmaz. Document'lar da aynı şekilde raporlanır.

**Hangisi olduğunu `X-Collage-Cache` söyler.** `collage.CacheStatusHeader`,
`X-Collage-Cache`'tir. Bir page ya da document ile cevap veren cache'lenebilir bir
`GET` ya da `HEAD` şunlardan birini taşır:

| Değer | Anlamı |
| --- | --- |
| `HIT` | Render edilmeden sunulan taze bir entry. |
| `MISS` | Render edildi: hiç entry yoktu ya da yalnızca stale bir entry vardı. |
| `STALE` | Yerine geçecek render başarısız ya da degraded olduğu için sunulan stale entry. |

Conditional bir request'e verilen `304` cevabı da bu header'ı taşır. Header şu
durumlarda gönderilmez: cache'in okunmadığı development'ta, bir hata response'unda,
cache'lenemeyen bir page'de (`Dynamic()`, `collage.SkipCache`) ve static build'de.
`StaleCache` olmayan kendi store'unuz da hiçbir zaman `STALE` cevabını vermez.
Header, dışarıdan debug eden kişi içindir; collage'ın kendisi onu okumaz.

## Key'deki query parametreleri

Varsayılan olarak ham query string'in tamamı key'e dahildir. Güvenli olan tek
varsayılan budur. Data handler request'in tamamını alır ve
`rc.Request.URL.Query()`'yi okuyabilir. Yani `?page=2` farklı bir page olabilir ve
collage bunu bilemez.

Bu varsayılan bilerek böyledir ve değişmeyecektir. Query'yi varsayılan olarak
çıkarmak ya da yalnızca bir izin listesini tutmak, tanımlamadığı bir parametreyi
okuyan her page'in o parametrenin tüm değerleri için tek bir entry'yi paylaşmasına
yol açardı: `?page=2` birinci sayfa diye sunulur, bir ziyaretçinin filtresi
sonrakine sunulurdu. Üstelik hiçbir şey bunu söylemezdi. Bir page daha küçük bir
key'i `WithCacheParams` ile kendisi seçer. Framework onun yerine asla tahmin
yürütmez.

Sınırsız sayıda query varyantının sınırsız belleğe dönüşmesini engelleyen
şey key değil, cache'in sınırlarıdır: `Cache.MaxEntries`, `Cache.MaxBytes` ve
tracker'ın üst sınırları
([tag index'in sınırlı olması](#the-tag-index-is-per-process-and-bounded) bölümüne
bakın). Varyant üreten bir client entry'lerin dışarı itilmesine yol açar.
Bunun bedeli bellek değil, hit oranıdır.

v0.79.0'dan beri bir client, host'u farklı yazarak cache'i bölemez
([Cache key](#the-cache-key) bölümüne bakın); ona kalan boyut query'dir. Hiçbir şey
bildirmeyen cacheable bir page, ham query'nin tamamını key'inde tutar; bu yüzden bir
client'ın uydurduğu her değer kendine ait bir entry olur. Query okuyan her cacheable
page'de ve document'ta `WithCacheParams`'ı, hiç okumayan birinde de adsız
`WithCacheParams()`'ı bildirin; böylece `?utm_source=` ve bir crawler'ın varyantları
onu bölmez.

Ama bu yaklaşım pahalıdır da. `?utm_source=newsletter` içeren bir bülten linki,
page'in ikinci bir kopyasının saklanmasına yol açar. Varyantları deneyen bir crawler
da cache'i kimsenin istemediği kopyalarla doldurur ve gerçek entry'leri dışarı iter.
Bunun yerine page'in hangi parametreleri okuduğunu belirtin:

```go
collage.NewPage("articles").
	WithLayouts(layout).
	WithContent(list).
	WithPath("en", "/articles").
	WithCacheParams("page", "sort").
	Incremental(time.Minute).
	Build()
```

Artık key'de yalnızca `page` ve `sort` bulunur ve bunlar sabit bir sıraya konur.
Böylece `?page=2&sort=new` ile `?sort=new&page=2` tek bir entry'yi paylaşır.
Query'deki diğer her şey caching açısından yok sayılır. Handler bunları yine
okuyabilir, ama bunlar page'in gösterdiği içeriği değiştirmemelidir.

`WithCacheParams()` isim verilmeden çağrılırsa query'yi key'den tamamen çıkarır.
Bu durumda page, query ne olursa olsun aynı şekilde render edilir.

Page'in okumadığı bir parametreyi listelemenin bir maliyeti yoktur. Okuduğu bir
parametreyi listelememek ise bir bug'dır. İki farklı page tek bir entry'yi paylaşır
ve bir okuyucuya başka birinin page'i sunulur. Bazı page'ler de hiç
cache'lenmemelidir. Örneğin bir arama page'inin key'i arama terimi olurdu, yani
okuyucu ne yazdıysa o. Böyle bir page için doğru seçim `Dynamic()`'tir.

Document'larda da aynı `WithCacheParams` bulunur.

## Memory mi, disk mi

Varsayılan olan `Type: "memory"`, page'leri process içinde tutar. Hızlıdır, ama her
restart'tan sonra boş başlar. En fazla `MaxEntries` kadar page tutar (varsayılan
10000, limitsiz için negatif) ve bunların toplamı en fazla `MaxBytes` kadar olabilir
(varsayılan 256 MiB). İkisinden biri dolduğunda eklenme sırasına göre en eskisini
atar. Bir page'in okunması onu daha yeni yapmaz. Expire olmuş bir entry taze diye
sunulmaz, ama eviction'ın yere ihtiyacı olana kadar
[stale](#stale-pages-and-x-collage-cache) olarak tutulur.

`Type: "disk"` page'leri dosya olarak tutar. Böylece bir restart her şeyin yeniden
render edilmesine yol açmaz:

```go
Cache: collage.CacheConfig{
	Enabled: true,
	Type:    "disk",
	Dir:     ".cache",
},
```

`Dir`'in varsayılan bir değeri yoktur. Dosyaları nereye yazacağını kendisi seçen bir
framework, onları kimsenin bakmadığı bir yere yazar. Bu dizini `.gitignore`'a
ekleyin. `collage.Cached` değerleri de `Dir/data/` altında tutulur. Bkz.
[Restart'tan sonra](#across-a-restart).

v0.11.0'dan beri oluşturulamayan bir dizin, uygulamanın başlamaması için bir neden
değildir. Salt okunur bir dosya sistemi ya da yazacak yeri olmayan bir container
buna örnektir. collage bu durumda bir uyarı log'lar ve onun yerine
memory'de cache'ler.

v0.39.0'dan beri disk cache de en fazla `MaxEntries` kadar entry tutar. Varsayılan
yine 10000'dir, sınırsız bırakmak istediğiniz bir cache için negatif bir değer verin.
Dolduğunda dosya değiştirilme zamanına göre en eskisini atar. Cache key request'in
host'unu ve varsayılan olarak query'sinin tamamını içerdiği için, bu sınır olmasa
ikisinden birini değiştiren anonim bir çağıran, her seferinde hiç atılmayan bir dosya
ekleyerek diski doldurabilirdi.

v0.40.0'dan beri eviction taraması, yazmaların aldığı lock'un dışında çalışır.
Böylece yeni key'lerden oluşan bir sel artık her yazanı dizinin tamamının
taranmasını beklemeye zorlamaz. Bir tarama hâlâ yer açarken dolu bir cache'e ulaşan
yeni bir entry sunulur ama saklanmaz, böylece sınır korunur.

### Byte cinsinden de sınırlı

`MaxEntries` entry'leri sayar, ne kadar yer kapladıklarını değil. Key de varsayılan
olarak query'yi içerir: binlerce uydurma query altında (`?utm=1`, `?utm=2`, ...)
istenen tek bir büyük page, cache'i onun kopyalarıyla doldurur. 57 KB'lık bir
dokümantasyon page'inin on bin kopyası yarım gigabyte'tan fazla memory demektir;
500 KB'lık bir listeleme page'inin ise beş gigabyte. `MaxBytes` (v0.40.0'dan beri)
bu ağırlığı sınırlar: aşıldığında, `MaxEntries`'te olduğu gibi en eski entry'ler
evict edilir. Sınırın tamamından büyük bir page ise sunulur ama saklanmaz.

| `MaxBytes` | Memory | Disk |
| --- | --- | --- |
| `0` | 256 MiB | 1 GiB (memory katmanı: 256 MiB) |
| negatif | sınır yok | sınır yok |

Memory cache'te tutulan içeriği, diskte ise dosyaların boyutunu sayar. Bir disk
cache'in memory katmanı ([aşağıda](#memory-in-front-of-the-disk)) 256 MiB tutar.
`MaxBytes` bundan küçükse katman da `MaxBytes` kadar tutar.

Go'nun allocator'ı büyük bir değeri tam bellek sayfalarına yuvarlar. Bu yüzden
50 KB'lık page'lerle dolu bir memory cache, `MaxBytes`'ın söylediğinden yaklaşık %10
daha fazla heap kaplar. Process'in kendisi de heap'inden fazlasını kaplar. Bu yüzden
bir container'da `GOMEMLIMIT`'i de ayarlayın; ayrıntılar için
[Deployment](/docs/deployment#a-memory-cache-and-the-containers-limit) sayfasına
bakın. Kendi store'unuz kendini sınırlar ve bu alanı yok sayar.

Page'in okuduğu query parametrelerini
[`WithCacheParams`](#query-parameters-in-the-key) ile belirtmek işin diğer yarısıdır:
query'yi yok sayan bir page binlerce değil, tek bir kopya saklar.

### Diskin önünde memory

v0.63.0'dan beri bir `"disk"` cache iki katmandır: dizinin önünde bir
[memory cache](#memory-or-disk). Bir page'i dosyadan okumak her request'te bir
açma, bir okuma ve bir decode demektir. Bu yüzden bir hit, bunun yerine memory'den
sunulur.

- Bir yazma önce diske, sonra memory'ye gider.
- Bir okuma ise önce memory'ye sorar ve diskte bulduğunu giderken memory'ye
  kopyalar.
- `MaxEntries` iki katmanı da sınırlar. `MaxBytes` diski sınırlar. Memory katmanı
  256 MiB tutar, `MaxBytes` bundan küçükse de `MaxBytes` kadar tutar.

Bu katman diske ek olarak process'inizin belleğini kullanır. Bu yüzden küçük bir
container'da bir memory cache'teki gibi `GOMEMLIMIT`'i ayarlayın. Katmanın daha azını
tutması gerekiyorsa `MaxBytes`'ı ya da `MaxEntries`'i düşürün (bkz.
[Deployment](/docs/deployment#a-memory-cache-and-the-containers-limit)).

**Restart'tan sonra cache boş başlamaz.** Açılışta memory katmanı, geçerli build'in
dizininden en yeni entry'lerden başlanarak doldurulur. Bu, bir sonraki entry katmanı
evict etmeye zorlayana kadar sürer. En son yazılan page'ler hemen memory'den `HIT`
olur, geri kalanlar için de tek bir dosya okuması yeter. Stale entry'ler de katmana
yüklenir. Böylece bir upstream kesintisinin ortasındaki bir restart, sunacak stale
entry'leri yine bulur. v0.62 ya da daha önceki sürümlerin yazdığı dosyaların
header'ında key yoktur. Bu yüzden onlar açılışta yüklenmez, ama onları okuyan ilk
request'te yine sunulur.

**Her instance için bir `Dir`.** Bir `Dir`'i paylaşan instance'lar bir cache'i
paylaşmaz. Her birinin kendi memory katmanı vardır. Memory bir page'i tuttuğu
sürece o page için verdiği cevap kesindir.

Bir instance'taki invalidation diski ve kendi memory'sini işaretler. Diğer her
instance kendi memory'sindeki kopyayı, o kopya expire olana kadar taze bir `HIT`
olarak sunmaya devam eder. TTL'i olmayan bir `Static` page için bu, instance restart
olana ya da kopyayı evict edene kadar sürer. Bu arada page'i yeniden çeken bir CDN'e
eski kopya verilebilir ve CDN onu tutabilir.

Her instance'a kendi dizinini verin. Instance'larının birbirinin invalidation'larını
görmesi gereken bir filonun, `Type: "disk"` yerine `Cache.Store`'da paylaşılan bir
store'a ihtiyacı vardır: [`collage.SharedCache`](#a-cache-a-fleet-shares)'i
implement eden bir store, örneğin collage-redis, page'leri ve veriyi filo genelinde
tutarlı tutar.

### Namespace

Disk cache, onu dolduran process'ten daha uzun yaşar. Bu hem işin amacıdır hem de
bir risktir. Template'i değişmiş yeni bir binary, eski binary'nin render ettiği
HTML'i sunmamalıdır.

Bu yüzden entry'ler build'e göre adlandırılmış bir alt dizinde durur.
`Cache.Version`'ı boş bırakırsanız bu ad, çalışan executable'ın hash'i olur. Hash,
tam olarak binary'ye derlenen kodunuz ya da template'leriniz değiştiğinde değişir.
Aynı build'in iki çalıştırması, ya da onu çalıştıran bir filodaki her makine, aynı
dizini türetir.

Bu, bir build'in başka bir build'in page'lerini okumasını engeller. Ama bir filonun
tek bir cache'i paylaşmasını sağlamaz ([yukarıya](#memory-in-front-of-the-disk)
bakın). `Version`'ı kendiniz yalnızca çıktının nasıl görüneceğini binary dışındaki
bir şey belirliyorsa ayarlayın. Buna örnek olarak bir içerik revizyonu verilebilir.

Forgery koruması için kullanılan key namespace'in bir kısmı değildir. Bu yüzden
disk cache, `Security.CSRFKey` ayarlı olsun ya da olmasın bir restart'tan sağ çıkar.
Yine de bu key bir tür page için önemlidir. İçinde form bulunan cache'lenmiş bir
page, her okuyucunun token'ının geleceği yerde key'den türetilmiş bir marker taşır
([Form'lar ve action'lar](/docs/forms-and-actions#pages-with-forms-are-still-cached)
sayfasına bakın). Bir key ile saklanmış bir page, başka bir key ile sunulamaz.

Bu yüzden başka bir key'in marker'ını taşıyan saklanmış bir page miss sayılır, atılır
ve yeniden render edilir. Böyle bir page, key değişmeden önce ya da kendi key'ini
üreten bir process tarafından render edilmiş olabilir. Form içermeyen page'ler
key'den hiç etkilenmez.

Aynı `Dir`'i ve aynı build'i paylaşan her şey diskteki dosyaları da paylaşır. Tek
bir test binary'sindeki iki uygulama da buna dahildir. Her teste `t.TempDir()` ile kendi
dizinini verin. Ayrıntılar için [Test yazmak](/docs/testing#isolate-the-disk-cache)
sayfasına bakın.

### Kendi store'unuz

`Cache.Store` herhangi bir `collage.Cache` implementasyonunu kabul eder. Bu durumda
`Type` yok sayılır. Ana şalter yine `Enabled`'dır. Bu alana typed nil bir pointer
atamayın. Interface tipindeki bir alanda duran nil bir `*myCache`, nil bir interface
değildir ve collage doğrudan onun üzerinden çağrı yapar.

`InvalidateKey` ve `Invalidate` siler. Yalnızca bunları implement eden bir store
v0.63.0 öncesindeki gibi davranır: invalidate edilen bir page gider ve başarısız olan
bir yeniden render request'i de başarısız kılar.

### Kendi store'unuzda stale page'ler

Store'unuz bir entry'yi invalidation'ından ya da expire olmasından sonra da
tutabiliyorsa `collage.StaleCache`'i de implement edin. Framework bunun için type
assertion yapar ve bu interface'i implement eden bir store'dan entry'leri silmek
yerine stale olarak işaretlemesini ister. Başarısız bir yeniden render'a son
başarılı render'ın page'iyle cevap verilebilmesini sağlayan budur:

```go
type StaleCache interface {
	Cache
	GetStale(ctx context.Context, key string) (content []byte, etag string, stale, found bool)
	MarkStaleKey(ctx context.Context, key string) error
	MarkStale(ctx context.Context, tags []string) error
}
```

- `GetStale`, invalidation'ını ya da expire süresini geçmiş bir entry'yi de
  `stale` true olarak dönen bir `Get`'tir. `found` yalnızca key'in hiç entry'si
  yoksa false olur.
- `Get` ise stale bir entry'yi de expire olmuş bir entry'yi de bildirmez. Bunlar
  eskisi gibi miss'tir. Framework, `StaleCache` olan bir store'da bir page'i
  yalnızca `GetStale` ile arar. Böylece sayaçlarınızda bir arama, tek bir hit ya da
  tek bir miss olur.
- `MarkStaleKey` entry'yi içeriği, ETag'i, expire süresi ve tag'leriyle birlikte
  tutar ve işaretler. Entry'si olmayan bir key'i işaretlemek hata değildir. Key için
  sonra yapılan bir `Set` ya da `SetTagged` taze bir entry saklar ve işareti
  kaldırır.
- `MarkStale` aynı şeyi tag'lerden herhangi birini taşıyan her entry için yapar.
  Bunu yalnızca tag'leri indeksleyen bir store yapabilir. Aynı zamanda
  `TaggedCache` olan bir `StaleCache`'ten, tag'e göre invalidate etmesi yerine
  işaretlemesi istenir.
- Bir entry'nin ne kadar süre stale kalacağını hiçbir şey sınırlamaz. Onu kendi
  sınırlarınıza göre kendiniz evict edin.

`PurgeTags` yine `InvalidateKey`'i ve `Invalidate`'i çağırır. Zaten stale olarak
işaretlenmiş bir page, collage'ın kendi tag index'i tarafından unutulmuştur. Bu
yüzden onu tag'e göre yalnızca aynı zamanda `TaggedCache` olan bir store purge
edebilir. `TaggedCache` olmayan bir `StaleCache`, böyle bir page'i evict edilene ya
da key'i purge edilene kadar tutar.

```go
var _ collage.StaleCache = (*myCache)(nil)
```

## Bir filonun paylaştığı cache

Bir load balancer arkasında her instance'ın kendi page cache'i, kendi
`collage.Cached` değerleri ve kendi [tag index'i](#the-tag-index-is-per-process-and-bounded)
vardır. Bir CMS webhook'u bunlardan birine ulaşır ve onun `InvalidateTags`'i
yalnızca o instance'a ulaşır. Diğerleri eski page'i taze bir `HIT` olarak sunmaya
ve eski değerden render etmeye, o kopya expire olana kadar devam eder. Bir `Static`
page için bu, instance restart olana ya da kopyayı evict edene kadar sürer.
`Dir`'ini paylaştıkları bir `"disk"` cache bunu değiştirmez
([yukarıya](#memory-in-front-of-the-disk) bakın).

v0.66.0'dan beri `collage.SharedCache`'i implement eden bir store, filonun
tamamı tarafından paylaşılır. Page'ler de `collage.Cached` verisi de onda tutulur ve
önünde instance başına hiçbir şey tutulmaz. Böylece herhangi bir instance'taki bir
`InvalidateTags` ya da `PurgeTags`, her instance'taki bir sonraki okumaya ulaşır.

<figure class="diagram-figure" data-search-ignore>
<svg class="diagram" viewBox="0 0 320 276" role="img" aria-labelledby="cache-fleet-title">
<title id="cache-fleet-title">Bir load balancer arkasındaki üç instance tek bir store'u paylaşır. Bir webhook A instance'ına ulaşır ve A, paylaşılan store'a tek bir DropTags çağrısı yapar. Store page'leri, Cached verisini, tag index'ini ve epoch'u tutar. B ve C instance'ları aynı store'u okur ve ona yazar, bu yüzden sonraki okumaları invalidation'ı görür. A, invalidation'ın ulaştığı page'leri kendi WarmLimit'i içinde warm eder.</title>
<defs><marker id="cache-fleet-arrow" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto"><path class="d-head" d="M0 0 8 4 0 8z" /></marker><marker id="cache-fleet-hit" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto"><path class="d-head-hit" d="M0 0 8 4 0 8z" /></marker></defs>
<rect class="d-chip-hit" x="10" y="8" width="90" height="24" rx="12" /><text class="d-hit-text" x="55" y="24" text-anchor="middle">webhook</text>
<line class="d-hitline" x1="55" y1="32" x2="55" y2="50" marker-end="url(#cache-fleet-hit)" />
<rect class="d-a" x="10" y="54" width="90" height="34" rx="5" /><text class="d-on-a" x="55" y="75" text-anchor="middle">instance A</text>
<rect class="d-a" x="115" y="54" width="90" height="34" rx="5" /><text class="d-on-a" x="160" y="75" text-anchor="middle">instance B</text>
<rect class="d-a" x="220" y="54" width="90" height="34" rx="5" /><text class="d-on-a" x="265" y="75" text-anchor="middle">instance C</text>
<line class="d-flow" x1="55" y1="88" x2="55" y2="100" marker-end="url(#cache-fleet-arrow)" />
<rect class="d-cmd" x="10" y="104" width="90" height="24" rx="12" /><text class="d-cmd-text" x="55" y="120" text-anchor="middle">DropTags</text>
<line class="d-flow" x1="55" y1="128" x2="55" y2="150" marker-end="url(#cache-fleet-arrow)" />
<line class="d-flow" x1="160" y1="88" x2="160" y2="150" marker-end="url(#cache-fleet-arrow)" />
<line class="d-flow" x1="265" y1="88" x2="265" y2="150" marker-end="url(#cache-fleet-arrow)" />
<text class="d-small" x="212" y="122" text-anchor="middle">okur, yazar</text>
<rect class="d-b" x="10" y="154" width="300" height="48" rx="5" /><text class="d-on-b" x="160" y="173" text-anchor="middle">paylaşılan store (collage-redis)</text><text class="d-on-b d-small" x="160" y="189" text-anchor="middle">page'ler, Cached veri, tag index, epoch</text>
<line class="d-flow" x1="82" y1="202" x2="82" y2="222" marker-end="url(#cache-fleet-arrow)" />
<line class="d-flow" x1="237" y1="202" x2="237" y2="222" marker-end="url(#cache-fleet-arrow)" />
<rect class="d-soft-a" x="10" y="226" width="145" height="42" rx="6" /><text class="d-small" x="82" y="243" text-anchor="middle">A page'leri warm eder</text><text class="d-small" x="82" y="256" text-anchor="middle">kendi WarmLimit'i içinde</text>
<rect class="d-frame" x="165" y="226" width="145" height="42" rx="6" /><text class="d-small" x="237" y="243" text-anchor="middle">B, C: sonraki okuma</text><text class="d-small" x="237" y="256" text-anchor="middle">invalidation'ı görür</text>
</svg>
<figcaption>Filo için tek bir store. Bir webhook'un ulaştığı instance tek bir <code>DropTags</code> çağrısı yapar. Bu çağrı, tag'lerin altındaki her page'i ve değeri, hangi instance yazmış olursa olsun stale olarak işaretler (purge'te ise siler) ve ulaştığı page'leri sayı, plugin event'i ve warming için listeler. Diğer her instance aynı store'u okur, bu yüzden sonraki okuması değişikliği görür. Yalnızca invalidate eden instance, kendi <code>WarmLimit</code>'i içinde warm eder.</figcaption>
</figure>

### collage-redis ile kurulum

[collage-redis](/docs/plugins#elagohtredis), Redis üzerinde bir `SharedCache`'tir.
Plugin değil, library'dir: register edilecek bir şey yoktur. Store
`Cache.Store`'a verilir ve bu durumda `Type` yok sayılır.

```sh
go get github.com/Elagoht/collage-redis
```

```go
import "github.com/Elagoht/collage-redis" // package redis

store, err := redis.New(redis.Config{URL: os.Getenv("REDIS_URL")})
if err != nil {
	log.Fatal(err)
}
defer store.Close()

app, err := collage.New(&collage.Config{
	Cache: collage.CacheConfig{Enabled: true, Store: store},
})
```

`New` Redis'e hiç bağlanmaz. Yalnızca hatalı bir config ya da URL'de başarısız
olur. Bu yüzden Redis kapalıyken de uygulama açılır ve cache'siz sunar. TLS için
`rediss://` kullanın. Parola ve veritabanı numarası da URL'den gelir. v0.1.0,
collage v0.66.0 ya da sonrasını ve Redis 6 ya da sonrasını gerektirir.

Yalnızca `URL` (ya da `Client`) gereklidir:

| Alan | Varsayılan | Anlamı |
| --- | --- | --- |
| `URL` | yok | Bir `redis://` ya da `rediss://` URL'si. `Client` verilmediyse zorunludur; ikisi birden verilemez. |
| `Client` | yok | Yeniden kullanılacak mevcut bir go-redis client'ı. `Close` onu açık bırakır: onu sahibi kapatır. |
| `Prefix` | `collage:` | Her key'in başına gelir. |
| `Version` | çalışan executable'ın hash'i | `Prefix`'ten sonra key'leri namespace'ler. Boş, yalnızca boşluktan oluşan ya da `:` içeren bir değer reddedilir. |
| `MaxAge` | 7 gün | Bir key'in en uzun yaşama süresi ve bir entry'nin en uzun taze kalma süresi. |
| `Timeout` | 250ms | Her Redis komutunu sınırlar. |
| `Logger` | `slog.Default()` | Store'un uyarılarını alır. |

`URL`'den kurulan bir client başarısız bir komutu yeniden denemez, çünkü bir
başarısızlık yalnızca bir miss'tir ve yeniden denemek Redis kapalıyken her
request'i yavaşlatırdı. Kendi `Client`'ınızı verirseniz, `Timeout`'un ağda beklemekte
olan bir komutu da sınırlaması için onda `ContextTimeoutEnabled`'ı açın ve
`MaxRetries`'i -1 yapın. Bir URL'nin tarif edemediği bir şeye, örneğin Sentinel'e,
ulaşmanın yolu da bir `Client`'tır.

Bunun için Redis'i şöyle çalıştırın:

- **`maxmemory-policy noeviction`**, yaklaşık `MaxAge` süresince yapılan yazmalara
  yetecek bellekle. Her key kendiliğinden expire olur; eviction yalnızca store'un hâlâ
  güvendiği entry'leri siler. Dolu bir Redis'te her yazma bir bütün olarak başarısız
  olur (bir miss: page sunulur ama saklanmaz), okumalar ve invalidation'lar ise
  çalışmaya devam eder.
- **Persistence (RDB ya da AOF)**, ya da ara sıra taze diye sunulan stale bir
  page'e razı olun: bir invalidation ile onun çakıştığı bir render'ın yazması
  arasındaki bir `FLUSHALL` ya da persistence olmadan bir restart, o yazmanın ayırt
  edilmesini sağlayan bilgiyi siler.
- **Tek bir node ya da Sentinel.** Redis Cluster v0.1.0'da desteklenmez.
- **Senkronize saatler** (örneğin NTP ile): tazelik, her instance'ın kendi saatiyle
  değerlendirilir.

### Ne paylaşılır, ne paylaşılmaz

Her instance tarafından paylaşılanlar:

- render edilmiş page'ler ve tag'leri;
- [stale fallback](#stale-pages-and-x-collage-cache): bir yeniden render başarısız
  olduğunda sunulan son başarılı kopya;
- `collage.Cached` değerleri. Her `Cached` çağrısı store'u okur: her yüklemede bir
  gidiş dönüş ve bir JSON decode. JSON'un olduğu gibi geri getiremeyeceği bir değer
  (bkz. [Restart'tan sonra](#across-a-restart)) render'lar arasında hiç
  cache'lenmez ve `Once` ile olduğu gibi davranır, çünkü onu memory'de tutmak
  instance başına kopyayı geri getirirdi;
- invalidation'lar ve ulaştıkları yer: `InvalidateTagsN`'in döndüğü sayı, bir CDN
  purge'ünün üzerinde çalıştığı `CacheInvalidateEvent.Entries` ve warming, herhangi
  bir instance'ın yazdığı page'leri kapsar.

Instance başına kalanlar:

- `Cached` key'i başına tek çekme işlemi, arka plandaki yenileme ve başarısızlık
  bekleme süresi: tek bir key'i aynı anda kaçıran N instance en fazla N çekme
  işlemi yapar;
- warming: yalnızca invalidation'ı çalıştıran instance yeniden render eder.

### Neyi garanti eder

- **Bir invalidation'ın çakıştığı render, her instance'ta stale olarak saklanır.**
  Store, yalnızca artan bir sayaç olan bir epoch'a sahiptir. Bir render ya da bir
  veri çekme işlemi başlamadan önce onu okur ve yazarken geri verir. O zamandan beri
  invalidate edilmiş bir tag taşıyan yazma stale olarak saklanır. Böylece B'nin,
  A'nın invalidate etmekte olduğu veriden render ettiği bir page fallback olarak
  tutulur, kalıcı olacak taze bir page olarak asla saklanmaz (bir `Static` page için
  bu, sonsuza kadar demektir). O zamandan beri purge edilmiş bir tag taşıyan yazma
  ise hiç saklanmaz.
- **Geç kalan, çakışmış bir yazma daha yeni bir page'in yerini almaz.** Bir
  invalidation'dan sonraki warm taze bir page sakladıysa, invalidation'dan önce
  başlayıp ondan sonra biten bir render, onun üzerine kendi stale page'ini yazmaz.
  Aynısı veri değerleri için de geçerlidir.
- **Bir invalidation tek bir çağrıdır.** `InvalidateTags` ve `PurgeTags`, tag ya da
  entry sayısı ne olursa olsun, beş saniye içinde birer `DropTags` çağrısı yapar.
  Bu çağrı tag'lerin altındaki her page'i ve değeri stale olarak işaretler (ya da
  siler) ve ulaştığı page'leri tek bir adımda döner. Başarısız olursa hata dönülür
  ve hiçbir şey sayılmaz.
- **Her instance'ın page'lerine ulaşır.** `DropTags`'in döndüğü, herhangi bir
  instance'ın yazdığı page'ler sayılır, `CacheInvalidateEvent.Entries` içinde
  listelenir (böylece [elagoht/cdnpurge](/docs/plugins#elagohtcdnpurge) B'nin
  sunduğunu da purge eder) ve invalidate eden instance'ın `WarmLimit`'i içinde warm
  edilir. Bu sınırın ötesindeki bir page'i, bir sonraki ziyaretçisini alan instance
  render eder; bir kez `MISS` olur. Plugin event'i yine her entry'yi alır.

### Store başarısız olduğunda

Hata veren ya da 500 milisaniyeden uzun süren bir store bir miss'tir: page render
edilir ve saklanmadan sunulur, değer çekilir ve dönülür. Bu asla başarısız bir
request değildir, bu yüzden öyle raporlanmaz: `"cache_write"` hata log'u da
`OnError` event'i de yoktur. Aksi hâlde dolu bir Redis bunları cache'lenebilir her
miss'te gönderirdi.

Page'lerin ve verinin başarısızlıkları birlikte App başına tek bir sağlık durumuna
gider. Bu durum warn seviyesinde, en fazla 30 saniyede bir, log'lanmayan
başarısızlıkların sayısıyla birlikte log'lanır
(`collage: shared page and data cache unavailable`) ve bir çağrı yeniden başarılı
olduğunda bir kez info seviyesinde log'lanır (`… available again`). collage-redis de
aynı kesintiyi kendi tarafında, en fazla 10 saniyede bir log'lar ve Redis yeniden
cevap verdiğinde bir kez info seviyesinde log'lar.

### Bilinmesi gerekenler

- **Bir deploy cache boş başlar.** collage-redis'in `Version`'ı executable'ın
  hash'idir. Bu yüzden yeni bir build boş bir namespace alır ve bir rolling
  deploy'un eski ve yeni build'leri birbirinin page'lerini asla okumaz. Eski build'in
  key'leri silinmez, `MaxAge` içinde expire olur. Entry'lerin binary'ler arasında
  paylaşılması için `Version`'ı kendiniz ayarlayın.
- **`MaxAge` stale fallback'i sınırlar.** `MaxAge` (varsayılan yedi gün) içinde
  yeniden render edilmeyen bir page ya da değer, bir upstream kesintisi sırasında
  bile gider. Built-in store'larda böyle bir sınır yoktur.
- **Key başına tek çekme işlemi instance başınadır.** Yukarıda olduğu gibi, tek bir
  key'i aynı anda kaçıran N instance upstream'e en fazla N çekme işlemi yapar.
- **Invalidate edilmiş ve henüz yeniden render edilmemiş bir page, store'un tag
  index'inde kalır.** Tag'lerinin o yeniden render'dan önce ikinci kez invalidate
  edilmesi onu yeniden sayar, plugin event'inde yine listeler ve warming için
  yeniden sunar.

`SharedCache` olmayan bir store, memory cache ve `Type: "disk"` eskisi gibi
davranır.

### Kendi SharedCache'inizi yazmak

Her store bir SharedCache olabilir. Framework, `TaggedCache` ve `StaleCache` için
yaptığı gibi bunun için de type assertion yapar ve interface ikisini de içerir:

```go
type SharedCache interface {
	TaggedCache
	StaleCache
	Epoch(ctx context.Context) (uint64, error)
	SetSince(ctx context.Context, key string, content []byte, ttl time.Duration,
		tags []string, meta []byte, since uint64) (etag string, outcome WriteOutcome, err error)
	DropTags(ctx context.Context, tags []string, purge bool) ([]Reached, error)
	GetData(ctx context.Context, key string) (value []byte, stale, found bool, err error)
	SetData(ctx context.Context, key string, value []byte, ttl time.Duration,
		tags []string, since uint64) (WriteOutcome, error)
	DeleteData(ctx context.Context, key string) error
}
```

- `Epoch`, store'un sayacını döner. Her tag'in bir generation'ı, yani son
  invalidation'ındaki epoch'u vardır. Aynı şekilde bir purge generation'ı da vardır.
- `SetSince` ve `SetData`, render'ın ya da çekme işleminin başladığı epoch'u,
  `since`'i alır ve ne yaptıklarını söyler: `since`'ten sonra bir tag purge
  edildiyse `collage.WriteSkipped`, invalidate edildiyse `collage.WriteStale`, aksi
  hâlde `collage.WriteFresh`.
- `DropTags` epoch'u ilerletir, generation'ları kaydeder, tag'lerin altındaki her
  page'i ve değeri stale olarak işaretler (ya da `purge` ile siler) ve ulaştığı her page'i
  bir `collage.Reached` olarak döner: key'i ve onunla yazıldığı meta. collage
  page'in host'unu ve path'ini bu meta'dan bilir.

`collagetest.SharedCacheSuite`, sözleşmenin tamamını bir store'a karşı, tek bir
backend üzerindeki iki handle aracılığıyla çalıştırır; tıpkı iki instance'ın tek bir
Redis'e iki bağlantı tutması gibi:

```go
func TestMySharedStore(t *testing.T) {
	collagetest.SharedCacheSuite(t, func(t *testing.T) (a, b collage.SharedCache) {
		// Two handles on one empty store.
		return newHandle(t, backend), newHandle(t, backend)
	})
}
```

Önce ilk handle üzerinde `collagetest.CacheSuite`'i çalıştırır: `Cache`,
`TaggedCache` ve `StaleCache` sözleşmeleri. Paylaşılmayan bir store için çağrılacak
olan da budur. Ardından bir handle üzerinden yazar ve diğeri üzerinden kontrol eder:
epoch sıralaması, çakışmış yazmaların stale saklanması ya da atlanması, `DropTags`'in
neye ulaştığı, veri değerleri ve `-race` altında anlam taşıyan eşzamanlı kullanım.
collage-redis bunu geçer. Filo olarak çalışan bir uygulamayı test etmek için
[Test yazmak](/docs/testing#a-fleet-on-one-shared-cache) sayfasına bakın.

## ETag'ler ve 304

Cache'lenen her page, içeriğinin hash'i olan bir ETag ile saklanır. Cache'ten sunulan
her response da bu ETag'i taşır. Page'i zaten almış bir tarayıcı ya da CDN, ETag'i
`If-None-Match` içinde geri gönderir. ETag hâlâ eşleşiyorsa collage body içermeyen
bir `304 Not Modified` ile cevap verir.

`Static()` page'lerin yeniden doğrulanmasını ucuz yapan budur. `must-revalidate`,
client'ın her seferinde sorması anlamına gelir ve cevap genellikle birkaç byte'tır.
Bir `304`, tam cevabın taşıyacağı [`X-Collage-Cache`](#stale-pages-and-x-collage-cache)
header'ını taşır.

Form içeren bir page burada da istisnadır. Okuyucuya gönderilen kopya, o okuyucunun
kendi forgery token'ını taşır. Bu yüzden ETag, saklanan kopyayı değil, o okuyucunun
kopyasını tanımlar ve response `private, no-store` olur. ETag'i geri gönderen bir
tarayıcı yalnızca kendisine verilen kopya için `304` alır. Paylaşılan hiçbir cache
bu page'i tutmaz.

## Eşzamanlı miss'ler tek bir kez render edilir

Popüler bir page expire olduğunda, ilk yeniden render bitmeden gelen her request bir
miss olur. Hiçbir önlem alınmazsa her biri ayrı ayrı render ederdi. Aynı page, aynı
upstream çağrılarıyla, aynı anda ve trafikle birlikte artan sayıda render edilirdi.

collage buna izin vermez. Bir key için gelen ilk request render eder. Bu sırada gelen
diğer request'ler onu bekler ve aynı byte'larla sunulur. Yapılandırmanız gereken
bir şey yoktur.

- **Yalnızca cache'lenen route'lar birleştirilir.** Dynamic bir page'in cache
  key'i yoktur. Bu yüzden iki request, page'in istediği gibi iki render demektir.
  Cache'lenen bir [document](/docs/documents) v0.12.0'dan beri bir page gibi
  birleştirilir. Böylece birçok client'ın yokladığı ve expire olan bir feed,
  handler'ını yalnızca bir kez çalıştırır.
- **Bir okuyucunun vazgeçmesi diğerlerini başarısız kılmaz.** Bağlantısı kapanan bir
  request beklemeyi bırakır. Render'ın kendisi onu bekleyen request'lerin hiçbirine
  ait değildir. Bu yüzden ilk request vazgeçtiğinde render durmaz: v0.18.1'den beri
  o request'in iptalinden bağımsız, v0.39.0'dan beri de okuyucuya özgü context
  değerleri olmadan çalışır ve yalnızca fragment'lerin kendi timeout'larıyla
  sınırlanır.
  Durdur'a basan tek bir okuyucu, aynı anda isteyen herkese kısımları
  "context canceled" ile başarısız olmuş bir page sunulmasına yol açamaz.
- **Bu durum görünürdür.** Bu şekilde sunulan bir request metriklerinize iki kez
  bildirilir. Önce lookup hiçbir şey bulamadığında `CacheMiss` olarak, sonra başka bir
  request'in render'ıyla sunulduğunda `CacheCoalesced` olarak bildirilir. Yani bir
  key'in yol açtığı render sayısı, miss sayısından coalesced event sayısının
  çıkarılmasıyla bulunur. Sürekli artan bir coalesced sayısı, bir page'in render
  edilebildiğinden daha hızlı expire olduğu anlamına gelir. Çok kısa bir
  `Incremental` TTL'i dışarıdan böyle görünür.

## Development'ta cache hiç okunmaz

`Config.DevMode` açıkken cache'lenmiş page'ler hiçbir zaman sunulmaz. Scaffold ile
oluşturulan bir proje bu ayarı `collage dev` altında açar. Development'ta
template'ler her request'te diskten yeniden yüklenir. Cache'lenmiş bir page ise az
önce yaptığınız değişikliği TTL'i boyunca gizlerdi.

Page'ler yine yazılır ve tag'ler yine takip edilir. Böylece warming dışında
hook'lar ve metrikler production'daki gibi davranır. Ortadan kalkan tek şey,
değişiklikten önce render edilmiş bir page'in sunulmasıdır.

Development'ta disk cache'in yerini bir memory cache alır ve collage bunu yaptığını
log'lar. Orada hiçbir response `X-Collage-Cache` taşımaz. `collage.Cached` de
orada render'lar arasında hiçbir şey tutmaz (aşağıya bakın).

## Veriyi page'ler arasında cache'lemek

Page cache, page'leri bütün olarak saklar. Her biri kendi yazarını gösteren otuz
farklı blog yazısı için ise hiçbir şey yapmaz. Her biri farklı bir URL'dir, her biri
ayrı render edilir ve her biri yazarı ayrıca çeker. Hepsini render eden bir
static export, yazarı otuz kez çeker.

`collage.Cached` yazarı saklar:

```go
var authorKey = collage.NewKey[Author]("author")

func authorCard(ctx context.Context, rc *collage.RenderContext) (Author, []string, error) {
	id := rc.Param("author")
	author, err := collage.Cached(rc, authorKey.With(id), time.Hour, []string{"author:" + id},
		func(ctx context.Context) (Author, error) {
			return api.Author(ctx, id)
		})
	return author, nil, err
}
```

```go
func Cached[T any](rc *RenderContext, key Key[T], ttl time.Duration, tags []string,
	fetch func(context.Context) (T, error)) (T, error)
```

`author:ada`'yı isteyen ilk render `fetch`'i çağırır. Sonraki her render, hangi
page'de olursa olsun, saklanan değeri alır. `Cache.Enabled` açıkken iki yazarın otuz
yazısı artık yalnızca iki yazar request'i yapar. Page'ler ister sunulsun ister export
edilsin, sonuç aynıdır. (Kapalıyken ve development'ta bir store yoktur;
[aşağıya](#where-it-keeps-nothing) bakın.) Bir [document'ın](/docs/documents)
handler'ı da aynı store'u paylaşır. Böylece bu yazarları okuyan bir sitemap ya da
feed, hiçbirini yeniden çekmez.

Key olarak bir [`collage.Key`](/docs/data-handlers#collagekey) verilir. Tek bir
render'ın fragment'leri de değerleri aynı türden key'lerle paylaşır. Key bir kez
tanımlanan bir isim ve bir tiptir; değişen kısmı `With` ekler. `authorKey.With("ada")`
key'inin adı `author:ada` olur. Key'in tipinden başka bir şey dönen bir `fetch`
derlenmez. Tag'ler düz string olarak kalır (burada `"author:" + id`), çünkü
`InvalidateTags` onlarla çağrılır. v0.50.0'dan önce key bir string'di.

Değişen kısmı uygulamanın seçtiği bir şeyden kurun: kendi veritabanında bulduğu
bir id, üst sınır koyduğu bir sayfa numarası gibi. Ham request girdisinden
kurmayın. Store `MaxBytes` ile değil `MaxEntries` ile sınırlıdır. Boyutu
uygulamaya ait Go değerleri tutar. Bu yüzden bir query parametresinden alınan
key, bir ziyaretçinin uydurduğu her değer için bir entry demektir ve `"disk"`
cache'te bunlar process'ten sonra da kalır.

### İki cache için tek tag kümesi

Verdiğiniz `tags`, data handler onları dönmüş gibi page'in kendi tag'lerine eklenir.
Böylece tek bir çağrı

```go
app.InvalidateTags(ctx, "author:ada")
```

saklanan yazarı *ve* onu gösteren her cache'lenmiş page'i birlikte stale olarak
işaretler.

Bu mekanizma olmadan en kolay hata burada yapılır. Yazarları kendi API client'ı
içinde memoize eden ve page'leri tag'e göre invalidate eden bir uygulamanın, birlikte
temizlenmesi gereken iki cache'i vardır. Yalnızca page'leri temizlerseniz, page'ler
yerine geçmeleri gereken eski yazar verisinden yeniden render edilir.

### Nasıl davranır

- **Key sizindir.** Key, değeri bütün uygulama genelinde adlandırır. Bu yüzden onu
  çektiğiniz şey kadar spesifik yapın: `authorKey` değil, `authorKey.With(id)`. Key,
  ismi ve tipinden oluşur. Bu yüzden adı aynı, tipi farklı iki key yanlış tiple
  okunan tek bir değer değil, iki ayrı değer tutar (v0.50.0'dan önce bu durum
  `collage.ErrCachedTypeMismatch` hatasıydı).
- **Bir key için aynı anda tek bir çekme işlemi çalışır.** Bir key çekilirken onu
  isteyen render'lar kendi çekme işlemlerini başlatmaz, süren işlemi bekler.
- **Hatalar saklanmaz.** Geri dönülecek stale bir değeri olmayan bir çekme işlemi
  hatasını bekleyen herkese iletir. Bir sonraki render yeniden dener. (Stale bir
  değer varsa başarısız bir çekme işlemi, başarısız bir yenilemedir: aşağıya bakın.)
  v0.34.0'dan beri iki başarısızlık bekleyenlere aktarılmaz. Onu başlatan request
  ortadan kalktığı için başarısız olan bir çekme işlemini, hâlâ sunulmakta olan bir
  bekleyen yeniden çeker. Panic'e düşen bir çekme işlemi ise key'ini serbest bırakır
  ve bekleyenlerine `collage.ErrCachedFetchPanicked` ile cevap verir. Eskiden key,
  bir restart'a kadar dolu kalırdı.
- **Invalidation eski değeri tutar, yenisi gelene kadar sunar.** `InvalidateTags` ve
  bir değerin `ttl`'inin dolması, değeri silmez, stale olarak işaretler. Page'e
  hiçbir warm render ulaşmadıysa (warming kapalıysa, `WarmLimit` aşıldıysa, tur bir
  başarısızlıkta durduysa ya da bir ziyaretçi warmer'dan önce geldiyse; bkz.
  [Warming](#warming-after-an-invalidation)) sonraki render'a stale değer hemen
  sunulur ve arka planda tek bir çekme işlemi onun yerini alır. Bu çekme işleminin
  timeout'u sabit 30 saniyedir ve request'ten bağımsız çalışır. Çekme işlemi
  bittikten sonra gelen request değişikliği görür. Tag'leri invalidate edildiğinde
  hâlâ süren bir çekme işlemi, sonucunu bekleyenlere iletir. Sonuç da saklanır, ama
  stale olarak. Araya `PurgeTags` girmişse sonuç hiç saklanmaz. Invalidation'dan
  sonra başlayıp böyle bir çekme işlemine katılan bir render da değeri stale olarak
  alır (v0.66.0'dan beri, her store için; önceden taze olarak alırdı ve page'i,
  invalidation'ın yerine yenisini getirdiği eski veriden taze olarak
  saklanabilirdi). Bu yüzden o
  render `no-store` ile sunulur ve saklanmaz. Bir warm render ise taze veriyi bekler
  ve taze bir page saklar; böyle bir çekme işlemine katılan warm render veriyi
  yeniden çeker.
- **Başarısız bir yenileme, stale değeri sunmayı sürdürür.** Arka plandaki çekme
  işlemi başarısız olursa stale değer kalır. Upstream ne kadar uzun süre kapalı
  kalırsa kalsın sunulmaya devam eder. Key'e bir bekleme süresi konur: ilk
  başarısızlıktan sonra bir saniye boyunca onun için yeni bir çekme işlemi başlamaz.
  Sonraki her başarısızlıkta bu süre iki, dört saniye diye ikiye katlanarak en fazla
  bir dakikaya çıkar. Böylece kapalı bir upstream'e her request'te sorulmaz. Başarılı
  bir çekme işlemi bekleme süresini bitirir. Key'in invalidate edilmesi de bitirir,
  çünkü bir webhook upstream'in cevap verdiğinin haberidir.
- **Stale veriden oluşan bir render cache'lenmez.** Plugin'ler bunu degraded olarak
  görür (`AfterRenderEvent.Degraded`). Render hiçbir yerde saklanmaz, böylece değer
  taze olunca page yeniden render edilir. Page cache'te page'in stale bir kopyası
  varsa, render'ın yerine o eski page `STALE` olarak sunulur. Bir invalidation'ın
  hemen ardından olan budur (bkz. [Stale
  page'ler](#stale-pages-and-x-collage-cache)). Kopya yoksa degraded render
  `Cache-Control: no-store` ile sunulur.
- **Sunulan stale değerler raporlanır.** Sunulan her stale değer, bekleme süresi
  içinde olsun olmasın, `CacheStale` metrics event'i olarak raporlanır. Page cache de
  sunduğu her stale page için aynı event'i kendi key'iyle raporlar. Başarısız her
  yenileme hatasıyla birlikte warn seviyesinde log'lanır. Başarısızlıklardan sonraki
  ilk başarı ise kaç başarısızlık olduğuyla birlikte info seviyesinde log'lanır.
- **Bulunamayan kayıt istisnadır.** Hatası `collage.ErrNotFound`'ı saran bir çekme
  işlemi değeri düşürür. Böylece upstream'de silinen bir kayıt siteden kaybolur ve
  page'i 404 döner.
- **Static export hiçbir zaman stale veri kullanmaz.** `collage.Build` upstream'i
  bekler ve upstream başarısız olursa başarısız olur. Bir invalidation'dan sonra
  yeniden render edilmiş bir fragment'i push etmek için plugin'in çağırdığı
  `App.RenderFragment` de aynı şekilde davranır. Push edilen şey bir daha
  istenmeyeceği için yeni değer olmak zorundadır.
- **`ttl`, page'in TTL'inden bağımsızdır.** Hiçbir şey invalidate etmediğinde bir
  değerin ne kadar süre taze sayılacağına üst sınır koyar. Süre dolunca değer
  silinmez, stale olur. Sıfır verirseniz değer, bir şey invalidate edene kadar taze
  kalır. Her dakika yeniden render edilen bir page, bir saat önce çekilmiş bir yazarı
  yine kullanabilir. Amaç da tam olarak budur.
- **Değerler memory'de tutulur ve sınırlıdır.** Değerler process içinde,
  `Cache.MaxEntries` kadar tutulur. En uzun süredir kullanılmayan önce atılır.
  Page'ler diskte cache'lense bile bu değişmez. Birden çok instance'lı bir deploy'da
  her instance kendi değerlerini tutar; `Cache.Store` bir
  [`collage.SharedCache`](#a-cache-a-fleet-shares) ise bu geçerli değildir: değerler
  o store'da, page'lerin yanında tutulur ve hiçbir instance bir kopya tutmaz.

### Restart'tan sonra

`Cache.Type` `"disk"` iken değerler ayrıca `<Cache.Dir>/data/` altına JSON olarak
yazılır. Yazma işini arka planda bir writer yapar ve `App.Shutdown` onu bekler.
Değerler başlangıçta geri okunur. Böylece uygulama restart'tan sonra boş bir data
cache'le başlamaz: diske yazılmış değerleri upstream'i beklemeden hemen sunar ve her
birini ilk kullanımında arka planda yeniler, çünkü process kapalıyken neyin
değiştiğini söyleyen bir şey yoktur.

- **Build'e göre namespace'lenmez.** Page'lerin aksine bu dizin build'ler arasında
  paylaşılır. Bu yüzden upstream'i kapalıyken yapılan yeni bir deploy, bir önceki
  deploy'un verisinden render eder. Dizinden yalnızca store'un kendi dosyaları
  silinir, başka hiçbir dosyaya dokunulmaz.
- **JSON'la geri gelmeyen bir tip memory'de kalır.** JSON'un olduğu gibi geri
  getiremeyeceği bir tip yalnızca memory'de tutulur. Unexported bir alanı ya
  da `json:"-"` etiketli bir alanı olan, ya da herhangi bir yerinde interface
  bulunan tipler buna girer. Tipin o kısmı, `time.Time` gibi hem
  `json.Marshaler` hem `json.Unmarshaler` implement ediyorsa bu kural uygulanmaz.
  Tipinin şekli değişmiş bir değer geri okunmaz.
- **Şifrelenmez.** Dosyalar düz JSON'dur. Bu yüzden disk cache ile `Cached`
  üzerinden kişisel veri geçirmeyin.
- **Aynı `Cache.Dir`'i paylaşan instance'lar data dizinini de paylaşır.** Her biri
  başlangıçta dizini kendi `MaxEntries` değerine göre budar.

### Hiçbir şey saklamadığı durumlar

`Cached` bazı durumlarda hiçbir şey saklamaz: `Cache.Enabled` false iken,
development'ta, middleware'i `collage.SkipCache` çağıran bir request'te (bir
[preview](/docs/previews)) ve bir action'ın kendi handler'ı içinde. Bu durumlarda
`collage.Once` gibi davranır. Aynı render'ın fragment'leri tek bir çekme işlemini
paylaşır ve bir sonraki render veriyi yeniden çeker. Dönülecek stale bir değer de
yoktur. Bu yüzden bir preview, yeni bir page'in yanında güncel veriyi de görür.

### Once, Cached ve page cache

| | Kimler arasında paylaşılır | Ne kadar yaşar |
| --- | --- | --- |
| `collage.Once` | tek bir render'ın fragment'leri | o render boyunca |
| `collage.Cached` | process'teki her render (bir [`SharedCache`](#a-cache-a-fleet-shares) ile filodaki her render) | `ttl` süresince ya da tag'leri invalidate edilene kadar taze, sonra yenisi gelene kadar stale sunulur |
| page cache | tek bir URL için gelen her request | page'in stratejisine göre ya da tag'leri invalidate edilene kadar taze, sonra yalnızca yeniden render'ı başarısız ya da degraded olursa stale sunulur |

Bir page'in iki kez çektiği şey için `Once`'ı, birçok page'in çektiği şey için
`Cached`'i, page'in kendisi için de page cache'i kullanın. Üçü birlikte çalışır.
Cache'lenmiş bir page render edilmez, bu yüzden veri çekme işlemlerinin hiçbiri
çalışmaz. `Once` için [Data handler'lar](/docs/data-handlers) sayfasına bakın.

Page cache host başınadır; `collage.Cached` değildir. Store'u process başına (ya da
filonun paylaşılan store'u olarak) tektir ve yalnızca verdiğiniz key ile key'lenir.
Bu yüzden birden fazla müşteriye hizmet veren bir sitede (örneğin elagoht/tenant)
`Cached(rc, postsKey, ...)` acme'nin yazılarını bir kez çeker ve globex'e de verir. Müşteriyi key'e ve tag'lere koyun;
böylece bir müşterinin verisini invalidate etmek ötekilere dokunmaz:

```go
var postsKey = collage.NewKey[[]Post]("posts")

func posts(ctx context.Context, rc *collage.RenderContext) ([]Post, []string, error) {
	id, _ := tenant.ID(rc)
	list, err := collage.Cached(rc, postsKey.With(id), time.Hour, []string{"posts:" + id},
		func(ctx context.Context) ([]Post, error) { return db.Posts(ctx, id) })
	return list, nil, err
}
```

[Fragment path'leri](/docs/forms-and-actions#a-fragment-at-its-own-url) üzerinden
kısım kısım yenilenen bir page, her fragment path'i için bir tane olmak üzere birkaç
render'dır ve `Once` bunlar arasında hiçbir şey paylaşmaz. Ayrı ayrı yenilenen ve
aynı veriyi okuyan fragment'ler bir çekme işlemini yalnızca `Cached` ile paylaşır.

## Örnek: bir blogda yazar page'leri

Bir blog yazarlarını bir CMS'te tutar. Her yazarın `/authors/{id}` adresinde bir
page'i vardır ve her yazı, yazarının adını ve biyografisini gösteren bir kart
içerir. Site, CMS'te olan iki şeyi takip etmek zorundadır:

- bir yazar profilini düzenler;
- bir yazar hesabını siler.

Profil `collage.Cached` ile, onu gösteren tüm page'ler için bir kez çekilir ve
`author:<id>` ile tag'lenir. Bir yazının yazar kartı da profili aynı key ve tag
üzerinden çeker.

```go
var authorKey = collage.NewKey[Author]("author")

func loadAuthor(ctx context.Context, rc *collage.RenderContext) (Author, []string, error) {
	id := rc.Param("id")
	author, err := collage.Cached(rc, authorKey.With(id), time.Hour, []string{"author:" + id},
		func(ctx context.Context) (Author, error) {
			return cms.Author(ctx, id) // wraps collage.ErrNotFound for a deleted author
		})
	return author, nil, err
}

authorPage := collage.NewPage("author").
	WithLayouts(layout).
	WithContent(collage.NewFragment("author", "pages/author.html").
		WithData(collage.LoadTagged(loadAuthor)).
		Required(). // a deleted author is a 404, not a page without its content
		Build()).
	WithPath("en", "/authors/{id}").
	Static(). // the same for every reader, until the CMS says otherwise
	Build()
```

Handler yalnızca path parametresini okur ve profil yalnızca CMS öyle dediğinde
değişir. Bu yüzden page `Static()` çağrısını kendisi yapar (bu çağrı olmasa
handler'ı onu dynamic yapardı) ve onu yeniden render ettiren şey aşağıdaki
webhook'tur.

CMS her değişiklikte bir webhook çağırır. Bu,
[`app.Handle`](/docs/middleware-and-apis) ile mount edilen sıradan bir `net/http`
handler'ıdır:

```go
type cmsEvent struct {
	Type     string `json:"type"`      // "author.updated" or "author.deleted"
	AuthorID string `json:"author_id"` // the CMS's id for the author
}

func cmsWebhook(app *collage.App) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// app.Handle adds no forgery check and no body limit: check the CMS's
		// signature here, and bound what you read.
		var event cmsEvent
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&event); err != nil {
			http.Error(w, "bad payload", http.StatusBadRequest)
			return
		}
		tag := "author:" + event.AuthorID
		var err error
		switch event.Type {
		case "author.updated":
			err = app.InvalidateTags(r.Context(), tag)
		case "author.deleted":
			// The CMS has already deleted the author, so the author's page
			// warms as a 404, and posts lose the card, instead of showing
			// the old profile.
			err = app.PurgeTags(r.Context(), tag)
		default:
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if err != nil {
			log.Printf("cms webhook %s %s: %v", event.Type, event.AuthorID, err)
			http.Error(w, "cache not updated", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}

// After collage.New:
if err := app.Handle("/hooks/cms", cmsWebhook(app)); err != nil {
	log.Fatal(err)
}
```

**Profil düzenlemesi `InvalidateTags` çağırır.** Tag, saklanan yazara ve onu
gösteren her cache'lenmiş page'e ulaşır: yazar page'ine ve kartı içeren her
yazıya. Bunun nedeni, `Cached`'in kendi tag'lerini page'in tag'lerine eklemesidir
([yukarıya](#one-set-of-tags-for-both-caches) bakın).

- Bunlar silinmez, stale olarak işaretlenir ve page'ler arka planda taze veriyle
  [warm edilir](#warming-after-an-invalidation).
- Warm render'dan önce gelen bir okuyucuya, değer yenilenirken eski page `STALE`
  olarak sunulur.
- CMS kapalıysa okuyucular 500 yerine son başarılı page'i görmeye devam eder.

**Hesap silme `PurgeTags` çağırır.** Burada bir kez bile olsa bir daha
gösterilmemesi gereken şey eski profilin kendisidir. `InvalidateTags` onu tutardı:
warmer'dan önce gelen bir okuyucuya, stale page'den ya da stale değerden yine
sunulabilirdi. `PurgeTags` diskteki kopyalar dahil hiçbir şey tutmaz.

Yazarı CMS'te purge işleminden sonra değil, önce silin. Bir purge, bir invalidation gibi
ulaştığı page'leri warm eder. Yazar silinmişse `/authors/{id}` için yapılan warm
render, required bir fragment'te `collage.ErrNotFound` ile karşılaşır ve 404 döner;
böylece hiçbir şey saklanmaz. Yazar hâlâ duruyorsa warm render profili yeniden
saklar.

Örnek bir memory cache varsayar. Profil kişisel veridir ve `Type: "disk"` ile
`Cached` değerlerini `Dir/data/` altına düz JSON olarak yazar (bkz.
[Restart'tan sonra](#across-a-restart)).
