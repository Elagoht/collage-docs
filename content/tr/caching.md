---
description: collage'ın render edilmiş page'leri ve onları oluşturan veriyi nasıl cache'lediği, neyi atacağını nasıl bildiği.
reference: CacheConfig, Cached, Once, Key, TaggedCache, SkipCache, Vary, PageBuilder.Static, PageBuilder.Incremental, PageBuilder.Dynamic, FragmentBuilder.Static, FragmentBuilder.Shared, StrategyAuto, PathTag
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
	},
})
```

`Enabled` false olduğunda her request render edilir ve hiçbir şey saklanmaz.
Page'lerin ne bildirdiği bunu değiştirmez.

## Render stratejileri

Her page, çıktısının nasıl yeniden kullanılabileceğini builder'ındaki tek bir
çağrıyla belirtir:

| Builder çağrısı | Ne olur | Gönderilen `Cache-Control` |
| --- | --- | --- |
| `Dynamic()` | Her request'te render edilir, hiç saklanmaz | `no-store` |
| `Static()` | Bir kez render edilir, bir şey onu invalidate edene kadar sunulur | `public, max-age=0, must-revalidate` |
| `Incremental(ttl)` | Render'dan sonra `ttl` geçene kadar cache'ten sunulur | `public, max-age=<saniye cinsinden ttl>` |

Son sütunun istisnaları vardır. Biri
[guard'lı bir page](/docs/pages-and-layouts#private-pages-guards)'dir. Guard cache
okunmadan önce çalıştığı için okuyucuları bu cache'i paylaşır, ama sunucunun
önündeki bir CDN ya da proxy guard çalıştırmaz. Bu yüzden response, strateji ne
olursa olsun `private, no-cache` ile gider. Okuyucunun kendi tarayıcısı page'i
saklayıp ETag'iyle yeniden doğrulayabilir; yeniden doğrulama da guard'ın gördüğü
bir request'tir.

Diğeri, form'un `{{csrfToken}}` değerini taşıyan bir page'dir. Bu page de
diğerleri gibi cache'lenir, ama her okuyucuya kendi token'ı gönderilir.
Bu yüzden response, strateji ne olursa olsun `private, no-store` ile gider.
Ayrıntılar için
[Form'lar ve action'lar](/docs/forms-and-actions#pages-with-forms-are-still-cached)
sayfasına bakın.

v0.34.0'dan beri iki istisna daha vardır. Degraded bir render
([aşağıya](#what-is-cached-and-when) bakın) `no-store` ile gider. Bu sunucu onu
saklamaz; dolayısıyla önündeki bir CDN de onu page'in TTL'i boyunca saklamamalıdır.
Development'ta ise her page ve document `no-store` ile gider. Böylece tarayıcı,
düzenlemenizden önceki bir page'i tutmaz.

```go
page := collage.NewPage("blog-post").
	WithLayouts(layout).
	WithContent(post).
	WithPath("en", "/blog/{slug}").
	Incremental(10 * time.Minute).
	WithDependency("blog:posts").
	Build()
```

`Static()` için pratikte bir sona erme süresi yoktur ve `DefaultTTL` ona uygulanmaz.
Page yalnızca tag'lerinden biri invalidate edildiğinde ya da cache yer açmak için
onu çıkardığında yeniden render edilir. Yalnızca birisi yayımladığında değişen
içerik için doğru strateji budur. Gönderdiği `Cache-Control`, tarayıcılara
ve CDN'lere page'i tutmalarını ama her seferinde sormalarını söyler.
[ETag](#etags-and-304) de bu sormayı ucuz hâle getirir.

`Incremental(ttl)`, zamana bağlı olarak değişen içerik içindir. Ne zaman değiştiğini
size bildiremeyen bir kaynaktan gelen içerik için de uygundur. Sıfır TTL bir
register hatasıdır (`collage.ErrMissingTTL`), negatif TTL de öyledir
(`collage.ErrInvalidTTL`). İkisi de sessizce hiç sona ermeyen bir page'e dönüşmez.

[Document'lar](/docs/documents), yani sitemap'ler, feed'ler ve HTML olmayan her şey
aynı üç çağrıyı kullanır. Page'lerle aynı şekilde saklanır, key'lenir ve
invalidate edilir. v0.12.0'dan beri de aynı şekilde
[birleştirilir](#concurrent-misses-render-once).

### Hiçbirini tanımlamayan bir page

Bu üç çağrıdan hiçbirini yapmayan bir page'in stratejisi register edilirken
belirlenir. O ana kadar stratejisi `collage.StrategyAuto`'dur, sonrasında diğer
üçünden biri olur.

- Render ettiği herhangi bir şey her render'da veri çekiyorsa **dynamic** olur. Bu,
  bir data handler (`collage.DataHandler`, `collage.Load` ya da `collage.Effect`) ya
  da bir slot resolver'dır. Layout'unda, content'inde, bunların slot'larına
  bağlanan herhangi bir şeyde, fallback'lerinde ya da `WithFragmentPath` ile
  açılan bir fragment'te olması fark etmez.
- Aksi hâlde **static** olur. Yalnızca template'leri ve sabit değerleri
  (`collage.Value(v)`, `WithTitle(s)`) render eden bir page her okuyucu için aynı
  render edilir. Bu yüzden böyle page'lerden oluşan bir site, her birine
  `Static()` yazmadan cache'lenir ve [export edilir](/docs/static-export).

Handler'ın dynamic anlamına gelmesinin nedeni şudur: handler request'i, bir
cookie'yi ya da saati okuyabilir ve fonksiyonun dışından bunu yapıp yapmadığı
anlaşılamaz. Tahmin yanlış çıktığında bunun bedeli bir render olur, bir
okuyucuya başkasının page'inin sunulması asla olmaz. Üzerinde düşünmeyi
unuttuğunuz bir page yalnızca yavaş olur, stale olmaz. Bir handler'ın çıktısı
herkes için aynıysa (dosyadan okunan bir yazı gibi), page'i `Static()` ya da
`Incremental(ttl)` çağrısını kendisi yapar.

Bir page'i dynamic yapan şey, page'den page'e farklı olmak değildir. Path ve
parametreleri cache key'in parçasıdır ve export'ta her URL kendi dosyasıdır:
`/recipes/pancakes` ile `/recipes/omelette` zaten ayrı tutulur. Bir page'i dynamic
yapan şey, *aynı* URL'e gelen iki request arasında değişen çıktıdır: cookie'den
okunan bir kullanıcı ya da saat gibi. Handler'ı yalnızca path'in parametrelerini ve
locale'i okuyan bir fragment bunu kendisi söyleyebilir. Fragment'e `Static()`
yazıldığında, o fragment page'i dynamic yapmaz. v0.17.0'dan beri.

```go
more := collage.NewFragment("more-recipes", "fragments/more-recipes.html").
	WithData(collage.DataHandler(loadMore)). // every recipe but rc.Param("slug")
	Static().
	Build()
```

Bu, birçok page'in paylaştığı bir fragment içindir: son yazıların listesi ya da
içerikten oluşturulan bir navigasyon gibi. Böylece onu kullanan her page'in aynı
sözü tekrarlaması gerekmez. Bu söz, `Static()`'in bir page için verdiği sözün
aynısıdır ve onu bozan bir handler bir okuyucunun render'ını bir sonrakine sunar.
Yalnızca fragment'in kendi handler'ını kapsar. Başka bir fragment'in handler'ı
page'i yine dynamic yapar. Static bir fragment'te olsa bile bir slot resolver da
aynı şeyi yapar, çünkü döndüreceği fragment'ler bir render onları isteyene kadar
bilinmez.

`Static()` aynı anda iki söz verir: çıktı her okuyucu için aynıdır ve bir şey
invalidate edilene kadar da aynı kalır. Hiçbir cookie okumayan ama bir ölçüm
döndüren bir handler (CPU yükü, bir kuyruğun uzunluğu, saat gibi) ilk sözü tutar,
ikincisini tutmaz. Çıktısı her okuyucu için aynıdır ama zaman içinde aynı kalmaz.
Böyle bir fragment'i bunun yerine `Shared()` ile işaretleyin. v0.19.0'dan beri.

```go
cpu := collage.NewFragment("cpu", "fragments/cpu.html").
	WithData(collage.Load(cpuUsage)). // the same for everyone, different every second
	Shared().
	Build()
```

`Shared()` page'in stratejisine dokunmaz. Page dynamic kalır ve bir export, tek bir
anın ölçümünü page'e yazmaz. Fragment'i okuyuculara gönderen şeye de tek bir
render'ın hepsine gönderilebileceğini söyler (bkz.
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

Son kural göründüğünden daha önemlidir. Hata veren bir fragment render'ı
**degraded** yapar. Fragment optional olsa, hatta fallback'i onun yerini doldurmuş
olsa bile bu değişmez. Degraded bir render sunulur ama saklanmaz. Saklansaydı, tek
bir request'teki geçici bir hata TTL dolana kadar her okuyucunun karşısına çıkardı.

Hata response'ları da hiçbir zaman cache'lenmez. 404 ya da 500 `no-store` ile
yazılır ve bir error page'in kendi render'ında bildirilen tag'ler atılır. Bir
[action'ın](/docs/forms-and-actions) response'u da hiçbir zaman cache'lenmez. Render
ettiği page nasıl tanımlanmış olursa olsun bu değişmez, çünkü o response tek bir
gönderimden üretilmiştir ve onu gönderen kişiye aittir.

Bir [asset mount'undan](/docs/assets) sunulan dosyalar page cache'ten hiç geçmez.
Ne kadar taze kalacaklarını yalnızca `Cache-Control` header'ları belirler.

## Cache key

Cache'lenmiş bir page, şu parçalardan oluşan bir key ile bulunur:

- request path'i,
- çözümlenen locale,
- yakalanan path parametreleri,
- query string ([aşağıya](#query-parameters-in-the-key) bakın),
- middleware'inizin `collage.Vary` ile bildirdiği her değer (bir handler'da
  `collage.Varied` ile okunur),
- request'in host'u (v0.34.0'dan beri).

Key'i aynı olan iki request, collage açısından aynı page'dir.

Host key'de yer alır, çünkü bir render `r.Host`'tan mutlak bir URL oluşturabilir:
bir canonical link, bir `og:url`. Host key'de olmasaydı, başka bir host adı veren
tek bir request bu URL'yi her okuyucuya sunulan kopyaya yazardı.

### Paylaşılan bir render neyi görür

Cache'lenmiş bir render, request'i aynı key'e sahip her okuyucuya sunulur. Bu
yüzden v0.34.0'dan beri render'a yalnızca key'in içerdikleri verilir: path, host,
key'deki query parametreleri ve middleware'inizin `collage.Vary` ile bildirdiği
header'lar. Okuyucunun cookie'leri, `Authorization`'ı, adresi ya da client
sertifikası verilmez. `WithCacheParams`'ın dışarıda bıraktığı bir query parametresi
de verilmez.

Static ya da incremental bir page'deki bir handler cookie okumak istediğinde hiçbir
cookie bulamaz. Eskiden ilk okuyucunun cookie'sini bulur ve onu, ondan sonra gelen
herkesin aldığı kopyaya yazardı. Request'in context'i de aynı muameleyi görür:
middleware'inizin oraya koyduğu bir değer, yani giriş yapmış bir kullanıcı, bir
tenant ya da bir request id, cookie ile aynı nedenden ötürü paylaşılan bir render'dan
gizlenir. Middleware'inizin `collage.Vary` ile bildirdiği bir değeri
`collage.Varied(rc, header)` ile, elinizde yalnızca bir context varsa (bir page'in
`StaticParams`'ı, bir sitemap'in `LastMod`'u, bir document handler'ın `ctx`'i)
`collage.VariedContext(ctx, header)` ile (v0.42.0'dan beri) geri okuyun; orada güvenlidir, çünkü key'in içindedir
ve her değerin kendi cache'lenmiş page'i vardır. Bir cache boyutu olmayan bir değerin
paylaşılan bir render'da yeri yoktur: okuyucunun kendi request'ine ihtiyaç duyan bir
page `Dynamic()` olmalıdır ve dynamic bir page request'in tamamını görür.
Development'ta, bir handler'ın gerçekten okuduğu gizlenmiş bir context değeri
log'lanır, böylece hata görünür olur. Document'lar da aynı kurala uyar
(v0.41.1'den beri; öncesinde cacheable bir document'ın render'ı ilk okuyucunun
context değerlerini hâlâ görüyordu).

### Origin'ler

`collage.BaseURL(rc)` paylaşılan bir render'da güvenlidir (v0.42.0'dan beri). Host
cache key'in içindedir; bu yüzden bir kopyanın sunulduğu her okuyucu aynı host'u
istemiştir ve render'ın mutlak URL'leri kurduğu origin de onlarındır. Static build'in
request'i yoktur, o `Config.BaseURL`'ü alır. `OriginResolver` plugin'i olmayan bir
site hiçbir fark görmez; bkz. [Bir host'un origin'i](/docs/writing-plugins#which-origin-a-host-has-originresolver).
Bir invalidation entry düşürdüğünde `CacheInvalidateEvent.Entries` her birinin
path'ine ek olarak host'unu da adlandırır.

## Dependency tag'ler

Tag, bir içeriği adlandıran herhangi bir string'dir: `post:hello-world`,
`author:ada`, `blog:posts`. Bir page'in tag'leri iki yerden gelir ve birleştirilir:

```go
// From the page, for what it always depends on.
WithDependency("blog:posts")

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

Bu tag'lerden herhangi birini taşıyan her cache'lenmiş page atılır. `collage.Cached`
ile bu tag'ler altında saklanan her değer de atılır. Başka hiçbir şeye dokunulmaz.
Atılan her page'i sonra açan okuyucu yeni bir render alır.

`InvalidateTagsN` aynı işi yapar ve ayrıca tag'lerin kaç cache key'ine ulaştığını
söyler:

```go
reached, err := app.InvalidateTagsN(ctx, "author:ada")
```

Bu sayı, collage'ın kendi tag index'inin çözümleyip attığı key'lerin sayısıdır.
Gerçekten kaldırılan canlı page'lerin tam sayısı değildir ve iki yönde de sapabilir.
Entry'si zaten expire olmuş bir key yine sayılır. Öte yandan store'un kendi tag
index'i üzerinden ulaştığı bir entry, sayılmadan kaldırılır. Bu, örneğin
[aşağıda](#the-tag-index-is-per-process-and-bounded) anlatılan sınırın ötesinde
kalan ya da başka bir instance'ın yazdığı bir entry olabilir. Sayı bir log satırı ya
da bir metrik için uygundur, uygulama mantığı için değil.

Cache bazı key'leri atamazsa geri kalanlar yine atılır ve hatalar birleştirilip
tek bir hata olarak döner. Başarılı görünen kısmi bir invalidation, stale page'lerin bir
deploy'dan sağ çıkmasına yol açar.

Bunu genellikle içeriğinizin değiştiği yerde çağırırsınız: bir CMS webhook'unda ya
da bir admin form'unda. Bir action bunu declarative olarak da yapabilir. Bunun için
[result'ındaki `InvalidateTags`](/docs/forms-and-actions#invalidating-what-an-action-changed)
alanını kullanır. Bu invalidation response yazılmadan önce çalışır. Böylece az önce
değiştirdiği page'e redirect edilen okuyucu eski sürümü hiçbir zaman görmez.

### Path'e göre invalidate etmek

Cache'lenen her page ve document ayrıca render edildiği URL path'ini adlandıran bir
tag'e de bağlıdır: `collage.PathTag(path)` (v0.23.0'dan beri). Bu yüzden elinizde
bir tag değil de path varsa, entry'yi path'iyle atabilirsiniz:

```go
app.InvalidateTags(ctx, collage.PathTag("/blog/hello"))
```

Bu, o path için cache'lenmiş olanı, başka neye bağlı olursa olsun atar.

Bir invalidation ne attığını da söyler. Bir plugin her invalidation'ı
`OnCacheInvalidate` üzerinden duyar. Event tag'leri ve `Paths` içinde atılan
cache'lenmiş page ve document'ların URL path'lerini taşır. Bir CDN'in purge etmesi
ve bir arama motoruna değiştiği bildirilmesi gereken budur. Hiç cache'lenmemiş bir
page burada yer almaz. Bkz. [Plugin yazmak](/docs/writing-plugins#cacheinvalidatehook).
[elagoht/cdnpurge](/docs/plugins#elagohtcdnpurge) ve
[elagoht/indexnow](/docs/plugins#elagohtindexnow) bunun üzerine kuruludur.

### Tag index'i process başınadır ve sınırlıdır

collage'ın hangi key'in hangi tag'e ait olduğunu tuttuğu kayıt memory'de, yani o
key'leri yazan process'in içinde durur. Bunun iki sonucu vardır.

**Load balancer arkasında her instance yalnızca kendi key'lerini bilir.** Built-in
cache'lerde bu sorun olmaz, çünkü her instance'ın kendi cache'i de vardır. Kendi
paylaşılan store'unuzu kullanıyorsanız (`Cache.Store`, örneğin bir Redis),
`collage.TaggedCache`'i de implement edin. Böylece store tag'leri kendisi indeksler.
collage bu durumda tag'e göre invalidation'ı store'dan da ister ve store başka bir
instance'ın yazdığı entry'lere de ulaşabilir. Built-in disk cache bunu yapar.
Tag'lerinin restart sonrasında da çalışmaya devam etmesinin nedeni budur.

**Tag başına tutulan kayıt `Cache.MaxKeysPerTag` ile sınırlıdır** (varsayılan
10000). Query string key'in bir parçası olduğundan, bir client tek bir page için
istediği kadar key üretebilir. Bir sınır olmasaydı index sonsuza kadar büyürdü. Bir
tag sınıra ulaştığında, o tag altında kaydedilmiş en eski key bu index'ten silinir.
Built-in memory ve disk cache'lerinin ikisi de `collage.TaggedCache`'i implement eder
ve her entry'nin tag'lerini kendileri kaydeder. Bu yüzden onlarla tag'i invalidate
etmek yine her entry'ye ulaşır. Yalnızca `InvalidateTagsN`'in bildirdiği sayı eksik
kalır. `TaggedCache`'i implement etmeyen kendi store'unuzda ise elde yalnızca
collage'ın index'i vardır. Index'ten silinen bir page, expire olana kadar store'da
kalır ve tag'i invalidate etmek artık ona ulaşmaz. Sınırı, bir tag'in gerçekte
kapsayabileceği cache'lenmiş URL sayısının üzerinde bir değere ayarlayın. Sınır
istemiyorsanız negatif bir değer verin.

**Built-in bir cache'le index bir bütün olarak da sınırlıdır** (v0.40.0'dan beri),
`MaxEntries` kadar key'le. Cache'lenmiş her path kendi başına bir tag'dir. Bu yüzden
yalnızca tag başına bir sınır, index'in bu sınırın page sayısıyla çarpımına
ulaşmasına izin veriyordu: uydurma query'lerle doldurulan büyük bir sitede her page
için 10000 key, oysa cache'in kendisi toplamda 10000 entry tutuyordu. Artık
`MaxEntries`'ten fazla key yaşıyor olamaz, bu yüzden önce en eski yazılanlar gider.
Negatif bir `MaxEntries` bunu kapatır. Kendi store'unuz yalnızca tag başına sınırı
korur, çünkü collage onun boyutunu bilemez.

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
yürütmez. Sınırsız sayıda query varyantının sınırsız belleğe dönüşmesini engelleyen
şey key değil, cache'in sınırlarıdır: `Cache.MaxEntries`, `Cache.MaxBytes` ve
tracker'ın üst sınırları ([tag index'in sınırlı olması](#the-tag-index-is-per-process-and-bounded)
bölümüne bakın). Varyant üreten bir client entry'lerin dışarı itilmesine yol açar.
Bunun bedeli bellek değil, hit oranıdır.

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
atar. Bir page'in okunması onu daha yeni yapmaz. Expire olmuş bir entry, bir sonraki
lookup'ta atılır.

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
ekleyin. v0.11.0'dan beri oluşturulamayan bir dizin, uygulamanın başlamaması için
bir neden değildir. Salt okunur bir dosya sistemi ya da yazacak yeri olmayan bir
container buna örnektir. collage bu durumda bir uyarı log'lar ve onun yerine
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
| `0` | 256 MiB | 1 GiB |
| negatif | sınır yok | sınır yok |

Memory cache'te tutulan içeriği, diskte ise dosyaların boyutunu sayar. Go'nun
allocator'ı büyük bir değeri tam bellek sayfalarına yuvarlar. Bu yüzden 50 KB'lık page'lerle
dolu bir memory cache, `MaxBytes`'ın söylediğinden yaklaşık %10 daha fazla heap
kaplar. Process'in kendisi de heap'inden fazlasını kaplar. Bu yüzden bir
container'da `GOMEMLIMIT`'i de ayarlayın; ayrıntılar için
[Deployment](/docs/deployment#a-memory-cache-and-the-containers-limit) sayfasına bakın.
Kendi store'unuz kendini sınırlar ve bu alanı yok sayar. Page'in okuduğu
query parametrelerini [`WithCacheParams`](#query-parameters-in-the-key) ile
belirtmek işin diğer yarısıdır: query'yi yok sayan bir page binlerce değil, tek bir
kopya saklar.

### Namespace

Disk cache, onu dolduran process'ten daha uzun yaşar. Bu hem işin amacıdır hem de
bir risktir. Template'i değişmiş yeni bir binary, eski binary'nin render ettiği
HTML'i sunmamalıdır.

Bu yüzden entry'ler build'e göre adlandırılmış bir alt dizinde durur.
`Cache.Version`'ı boş bırakırsanız bu ad, çalışan executable'ın hash'i olur. Hash,
tam olarak binary'ye derlenen kodunuz ya da template'leriniz değiştiğinde değişir.
Aynı build'in iki çalıştırması, ya da onu çalıştıran bir filodaki her makine, tek
bir cache'i paylaşır. `Version`'ı kendiniz yalnızca çıktının nasıl görüneceğini
binary dışındaki bir şey belirliyorsa ayarlayın. Buna örnek olarak bir içerik
revizyonu verilebilir.

Forgery koruması için kullanılan key namespace'in bir parçası değildir. Bu yüzden
disk cache, `Security.CSRFKey` ayarlı olsun ya da olmasın bir restart'tan sağ çıkar.
Yine de bu key bir tür page için önemlidir. İçinde form bulunan cache'lenmiş bir
page, her okuyucunun token'ının geleceği yerde key'den türetilmiş bir marker taşır
([Form'lar ve action'lar](/docs/forms-and-actions#pages-with-forms-are-still-cached)
sayfasına bakın). Bir key ile saklanmış bir page, başka bir key ile sunulamaz. Bu
yüzden başka bir key'in marker'ını taşıyan saklanmış bir page miss sayılır, atılır
ve yeniden render edilir. Böyle bir page, key değişmeden önce ya da kendi key'ini
üreten bir process tarafından render edilmiş olabilir. Form içermeyen page'ler
key'den hiç etkilenmez.

Aynı `Dir`'i ve aynı build'i paylaşan her şey entry'leri de paylaşır. Tek bir test
binary'sindeki iki uygulama da buna dahildir. Her teste `t.TempDir()` ile kendi
dizinini verin. Ayrıntılar için [Test yazmak](/docs/testing#isolate-the-disk-cache)
sayfasına bakın.

### Kendi store'unuz

`Cache.Store` herhangi bir `collage.Cache` implementasyonunu kabul eder. Bu durumda
`Type` yok sayılır. Ana şalter yine `Enabled`'dır. Bu alana typed nil bir pointer
atamayın. Interface tipindeki bir alanda duran nil bir `*myCache`, nil bir interface
değildir ve collage doğrudan onun üzerinden çağrı yapar.

## ETag'ler ve 304

Cache'lenen her page, içeriğinin hash'i olan bir ETag ile saklanır. Cache'ten sunulan
her response da bu ETag'i taşır. Page'i zaten almış bir tarayıcı ya da CDN, ETag'i
`If-None-Match` içinde geri gönderir. ETag hâlâ eşleşiyorsa collage body içermeyen
bir `304 Not Modified` ile cevap verir.

`Static()` page'lerin yeniden doğrulanmasını ucuz yapan budur. `must-revalidate`,
client'ın her seferinde sorması anlamına gelir ve cevap genellikle birkaç byte'tır.

Form içeren bir page burada da istisnadır. Okuyucuya gönderilen kopya, o okuyucunun
kendi forgery token'ını taşır. Bu yüzden ETag, saklanan kopyayı değil, o okuyucunun
kopyasını tanımlar ve response `private, no-store` olur. ETag'i geri gönderen bir
tarayıcı yalnızca kendisine verilen kopya için `304` alır. Paylaşılan hiçbir cache
bu page'i tutmaz.

## Eşzamanlı miss'ler tek bir kez render edilir

Popüler bir page expire olduğunda, ilk yeniden render bitmeden gelen her request bir
miss olur. Hiçbir önlem alınmazsa her biri ayrı ayrı render ederdi. Aynı page, aynı
upstream çağrıları, aynı anda ve trafikle birlikte artan sayıda çalışırdı.

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
  sınırlanır. Stop'a basan tek bir okuyucu,
  aynı anda isteyen herkese parçaları "context canceled" ile başarısız olmuş bir
  page veremez.
- **Bu durum görünürdür.** Bu şekilde sunulan bir request metriklerinize iki kez
  bildirilir. Önce lookup hiçbir şey bulamadığında `CacheMiss` olarak, sonra başka bir
  request'in render'ıyla sunulduğunda `CacheCoalesced` olarak bildirilir. Yani bir
  key'in yol açtığı render sayısı, miss sayısından coalesced event sayısının
  çıkarılmasıyla bulunur. Sürekli artan bir coalesced sayısı, bir page'in render
  edilebildiğinden daha hızlı expire olduğu anlamına gelir. Çok kısa bir
  `Incremental` TTL'i dışarıdan böyle görünür.

## Development'ta cache hiç okunmaz

`Config.DevMode` (ya da `Template.DevMode`) açıkken cache'lenmiş page'ler hiçbir
zaman sunulmaz. Scaffold ile oluşturulan bir proje bu ayarı `collage dev` altında
açar. Development'ta template'ler her request'te diskten yeniden yüklenir.
Cache'lenmiş bir page ise az önce yaptığınız değişikliği TTL'i boyunca gizlerdi.
Page'ler yine yazılır ve tag'ler yine takip edilir. Böylece hook'lar ve metrikler
production'daki gibi davranır. Ortadan kalkan tek şey, değişiklikten önce render
edilmiş bir page'in sunulmasıdır.

Development'ta disk cache'in yerini bir memory cache alır ve collage bunu yaptığını
log'lar. `collage.Cached` de orada render'lar arasında hiçbir şey tutmaz (aşağıya
bakın).

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
render'ın fragment'leri de değerleri aynı türden key'lerle paylaşır. Key bir kez tanımlanan
bir isim ve bir tiptir; değişen kısmı `With` ekler. `authorKey.With("ada")`
key'inin adı `author:ada` olur. Key'in tipinden başka bir şey dönen bir `fetch`
derlenmez. Tag'ler düz string olarak kalır (burada `"author:" + id`), çünkü
`InvalidateTags` onlarla çağrılır. v0.50.0'dan önce key bir string'di.

### İki cache için tek tag kümesi

Verdiğiniz `tags`, data handler onları dönmüş gibi page'in kendi tag'lerine eklenir.
Böylece tek bir çağrı

```go
app.InvalidateTags(ctx, "author:ada")
```

saklanan yazarı *ve* onu gösteren her cache'lenmiş page'i birlikte atar.

Bu mekanizma olmadan en kolay hata burada yapılır. Yazarları kendi API client'ı
içinde memoize eden ve page'leri tag'e göre invalidate eden bir uygulamanın, birlikte
temizlenmesi gereken iki cache'i vardır. Yalnızca page'leri temizlerseniz, page'ler
yerine geçmeleri gereken stale yazar verisinden yeniden render edilir.

### Nasıl davranır

- **Key sizindir.** Key, değeri bütün uygulama genelinde adlandırır. Bu yüzden onu
  çektiğiniz şey kadar spesifik yapın: `authorKey` değil, `authorKey.With(id)`.
  Key, ismi ve tipinden oluşur. Bu yüzden adı aynı, tipi farklı iki key yanlış tiple
  okunan tek bir değer değil, iki ayrı değer tutar (v0.50.0'dan önce bu durum
  `collage.ErrCachedTypeMismatch` hatasıydı).
- **Key başına aynı anda tek bir çekme işlemi.** Bir key çekilirken onu isteyen
  render'lar kendi çekme işlemlerini başlatmaz, süren işlemi bekler.
- **Hatalar saklanmaz.** Bekleyen herkes hatayı alır. Bir sonraki render yeniden
  dener. v0.34.0'dan beri iki başarısızlık bekleyenlere aktarılmaz. Onu başlatan
  request ortadan kalktığı için başarısız olan bir çekme işlemini, hâlâ sunulmakta
  olan bir bekleyen yeniden çeker. Panic'e düşen bir çekme işlemi ise key'ini serbest
  bırakır ve bekleyenlerine `collage.ErrCachedFetchPanicked` ile cevap verir. Eskiden
  key, bir restart'a kadar dolu kalırdı.
- **Invalidate edilmiş bir çekme işlemi saklanmaz.** Bir key'in çekme işlemi hâlâ sürerken
  tag'leri invalidate edilirse, sonuç bekleyenlere iletilir ama saklanmaz. Çünkü bu
  sonuç, invalidation'ın yerine yenisini koymak istediği şeyin ta kendisidir.
- **`ttl`, page'in TTL'inden bağımsızdır.** Hiçbir şey invalidate etmediğinde değerin
  ne kadar tutulacağına üst sınır koyar. Sıfır verirseniz değer, bir şey invalidate
  edene kadar tutulur. Her dakika yeniden render edilen bir page, bir saat önce
  çekilmiş bir yazarı yine kullanabilir. Amaç da tam olarak budur.
- **Memory'de ve sınırlı.** Değerler process içinde, `Cache.MaxEntries` kadar
  tutulur. En uzun süredir kullanılmayan önce atılır. Page'ler diskte cache'lense bile
  bu değişmez. Birden çok instance'lı bir deploy'da her instance kendi değerlerini
  tutar.

### Hiçbir şey saklamadığı durumlar

`Cached` bazı durumlarda hiçbir şey saklamaz: `Cache.Enabled` false iken,
development'ta, middleware'i `collage.SkipCache` çağıran bir request'te (bir
[preview](/docs/previews)) ve bir action'ın kendi handler'ı içinde. Bu durumlarda
`collage.Once` gibi davranır. Aynı render'ın fragment'leri tek bir çekme işlemini
paylaşır ve bir sonraki render veriyi yeniden çeker. Bu yüzden bir preview, yeni bir page'in
yanında güncel veriyi de görür.

### Once, Cached ve page cache

| | Kimler arasında paylaşılır | Ne kadar yaşar |
| --- | --- | --- |
| `collage.Once` | tek bir render'ın fragment'leri | o render boyunca |
| `collage.Cached` | process'teki her render | `ttl` süresince ya da tag'leri invalidate edilene kadar |
| page cache | tek bir URL için gelen her request | page'in stratejisine göre ya da tag'leri invalidate edilene kadar |

Bir page'in iki kez çektiği şey için `Once`'ı, birçok page'in çektiği şey için
`Cached`'i, page'in kendisi için de page cache'i kullanın. Üçü birlikte çalışır.
Cache'lenmiş bir page render edilmez, bu yüzden veri çekme işlemlerinin hiçbiri
çalışmaz. `Once` için [Data handler'lar](/docs/data-handlers) sayfasına bakın.

Page cache host başınadır; `collage.Cached` değildir. Store'u process başına tektir
ve yalnızca verdiğiniz key ile anahtarlanır. Bu yüzden birden fazla müşteriye hizmet
veren bir sitede (örneğin elagoht/tenant) `Cached(rc, postsKey, ...)` acme'nin
post'larını bir kez çeker ve globex'e de verir. Müşteriyi key'e ve tag'lere koyun;
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
parça parça yenilenen bir page, her fragment path'i için bir tane olmak üzere birkaç
render'dır ve `Once` bunlar arasında hiçbir şey paylaşmaz. Ayrı ayrı yenilenen ve
aynı veriyi okuyan fragment'ler bir fetch'i yalnızca `Cached` ile paylaşır.
