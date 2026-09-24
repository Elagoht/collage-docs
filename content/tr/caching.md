---
description: collage render edilmiş sayfaları ve onları oluşturan veriyi nasıl önbelleğe alır, neyi atması gerektiğini nasıl bilir.
---

# Önbellek

collage iki şeyi önbelleğe alır. **Sayfa önbelleği**, bir render'ın ürettiğini —
tek bir URL'nin HTML'ini — saklar; böylece o URL'nin bir sonraki okuyucusu sayfayı
render beklemeden alır. **Veri önbelleği**, `collage.Cached`, render'ların
yapıldığı malzemeyi saklar; böylece aynı yazarı gösteren otuz sayfa o yazarı bir kez
çeker.

İkisi de aynı **bağımlılık etiketleriyle** dizinlenir; tek bir çağrı bir içerik
parçasını ve ondan kurulmuş her şeyi birlikte atar.

Siz açmadıkça hiçbir şey önbelleğe alınmaz:

```go
app, err := collage.New(&collage.Config{
	Template: collage.TemplateConfig{Root: "templates"},
	Cache: collage.CacheConfig{
		Enabled:    true,
		Type:       "memory",
		DefaultTTL: 5 * time.Minute,
		MaxEntries: 10000,
	},
})
```

`Enabled` false iken, sayfalar ne bildirirse bildirsin her istek render edilir ve
hiçbir şey saklanmaz.

## Render stratejileri

Her sayfa, çıktısının nasıl yeniden kullanılabileceğini builder'ı üzerinde tek bir
çağrıyla söyler:

| Builder çağrısı | Ne olur | Gönderilen `Cache-Control` |
| --- | --- | --- |
| `Dynamic()` (varsayılan) | Her istekte render edilir, hiç saklanmaz | `no-store` |
| `Static()` | Bir kez render edilir, bir şey onu geçersiz kılana kadar sunulur | `public, max-age=0, must-revalidate` |
| `Incremental(ttl)` | Render'dan bu yana `ttl` geçene kadar önbellekten sunulur | `public, max-age=<saniye cinsinden ttl>` |

Bir formun `{{csrfToken}}`'ını taşıyan sayfa, son sütunun istisnasıdır: diğerleri
gibi önbelleğe alınır, ama her okuyucuya kendi token'ı gönderilir; bu yüzden yanıt,
strateji ne olursa olsun `private, no-store` olarak çıkar — bkz.
[Formlar ve action'lar](/docs/forms-and-actions#pages-with-forms-are-still-cached).

```go
page := collage.NewPage("blog-post").
	WithLayout(layout).
	WithContent(post).
	WithPath("en", "/blog/{slug}").
	Incremental(10 * time.Minute).
	WithDependency("blog:posts").
	Build()
```

`Dynamic()` varsayılandır, çünkü asla yanlış olamayacak tek stratejidir: üzerinde
düşünmeyi unuttuğunuz bir sayfa yavaş olur, bayat değil.

`Static()`'in pratikte bir son kullanma süresi yoktur ve `DefaultTTL` ona
uygulanmaz. Yalnızca etiketlerinden biri geçersiz kılındığında ya da önbellek yer
açmak için onu çıkardığında yeniden render edilir. Biri yayımladığında — ve yalnızca
o zaman — değişen içerik için doğru strateji budur. `Cache-Control`'ü tarayıcılara
ve CDN'lere sayfayı tutmalarını ama her seferinde sormalarını söyler;
[ETag](#etags-and-304) da sormayı ucuzlatır.

`Incremental(ttl)`, saate bağlı değişen ya da ne zaman değiştiğini size
söyleyemeyen bir yerden gelen içerik içindir. Sıfır TTL bir kayıt hatasıdır
(`collage.ErrMissingTTL`), negatif TTL de öyle (`collage.ErrInvalidTTL`) —
sessizce hiç süresi dolmayan bir sayfa değil.

[Document'lar](/docs/documents) — sitemap'ler, feed'ler, HTML olmayan her şey —
aynı üç çağrıyı alır; aynı şekilde saklanır, anahtarlanır, geçersiz kılınır ve —
v0.12.0'dan itibaren — [birleştirilir](#concurrent-misses-render-once).

## Ne, ne zaman önbelleğe alınır

Bir yanıt sayfa önbelleğine yalnızca şunların hepsi doğruysa girer:

- önbellek açıktır ve sayfa `Static()` ya da `Incremental(ttl)`'dir;
- istek bir `GET`'tir. Bir `HEAD` önbellekten *sunulabilir*, ama önbelleği asla
  doldurmaz — saklanacak bir gövde üretmemiştir;
- her fragment başarıyla render edilmiştir.

Bu son kural göründüğünden daha önemlidir. Hata veren bir fragment — isteğe bağlı
olsa bile, yedeği onun yerini doldurmuş olsa bile — render'ı **kusurlu** (degraded)
yapar; kusurlu bir render sunulur ama saklanmaz. Saklamak, tek bir isteğin geçici
hatasını TTL dolana kadar her okuyucunun önüne sabitlemek olurdu.

Hata yanıtları da asla önbelleğe alınmaz. 404 ya da 500 `no-store` olarak yazılır
ve bir hata sayfasının kendi render'ının bildirdiği etiketler atılır. Bir
[action'ın](/docs/forms-and-actions) yanıtı, render ettiği sayfa nasıl bildirilmiş
olursa olsun asla önbelleğe alınmaz: tek bir gönderimden üretilmiştir ve onu
gönderene aittir.

Bir [asset mount'undan](/docs/assets) sunulan dosyalar sayfa önbelleğinden hiç
geçmez. Tazelikleri yalnızca `Cache-Control` header'larıdır.

## Önbellek anahtarı

Önbellekteki bir sayfa, şunlardan oluşan bir anahtarla bulunur:

- isteğin yolu,
- çözümlenmiş locale,
- yakalanan yol parametreleri,
- query string (bkz. [aşağısı](#query-parameters-in-the-key)),
- middleware'inizin `collage.Vary` ile bildirdiği her değer.

Anahtarı aynı olan iki istek, collage açısından aynı sayfadır.

## Bağımlılık etiketleri

Etiket, bir içerik parçasını adlandıran herhangi bir dizedir: `post:hello-world`,
`author:ada`, `blog:posts`. Bir sayfanın etiketleri iki yerden gelir ve
birleştirilir:

```go
// From the page, for what it always depends on.
WithDependency("blog:posts")

// From a data handler, for what this particular render used.
return view, []string{"post:" + post.Slug, "author:" + post.AuthorID}, nil
```

Sayfanın kendi etiketleri, hiçbir şey render edilmeden önce bilinir. İşe yarayanlar
data handler'ınkilerdir, çünkü bu URL'nin sonunda hangi yazıyı ve hangi yazarı
gösterdiğini yalnızca handler bilir. Etiketler, handler ardından bir hata döndürse
bile toplanır: sayfanın neye bağlı olduğunu yine de söylemiştir.

Sayfa saklandığında etiketleri de onunla birlikte saklanır ve collage her etiketten
hangi önbellek anahtarlarının kurulduğunu hatırlar.

### Geçersiz kılmak

İçerik değiştiğinde onu adlandırın:

```go
if err := app.InvalidateTags(ctx, "post:"+slug, "blog:posts"); err != nil {
	log.Printf("invalidate: %v", err)
}
```

Bu etiketlerden herhangi birini taşıyan önbellekteki her sayfa atılır; aynı şekilde
`collage.Cached`'in onların altında sakladığı her değer de. Başka hiçbir şeye
dokunulmaz. Atılan her sayfanın bir sonraki okuyucusu taze bir render alır.

`InvalidateTagsN` aynısını yapar ve etiketlerin kaç önbellek anahtarına ulaştığını
söyler:

```go
reached, err := app.InvalidateTagsN(ctx, "author:ada")
```

Bu sayı, collage'ın kendi etiket dizininin çözümleyip attığı anahtarların sayısıdır;
kaldırılan canlı sayfaların kesin sayısı değildir ve her iki yönde de yanılır:
girdisinin süresi çoktan dolmuş bir anahtar yine de sayılır; deponun kendi etiket
dizini üzerinden ulaştığı bir girdi ise — [aşağıdaki](#the-tag-index-is-per-process-and-bounded)
sınırın ötesinde kalan ya da başka bir instance'ın yazdığı — sayılmadan kaldırılır.
Bir log satırı ya da bir metrik için iyidir, mantık kurmak için değil.

Önbellek bazı anahtarları atamazsa geri kalanlar yine atılır ve hatalar birleştirilip
error içinde döner. Başarılı olduğunu bildiren kısmi bir geçersiz kılma, bayat
sayfaların bir yayına almadan sağ çıkmasının yoludur.

Genellikle çağrılacağı yer, içeriğinizin değiştiği yerdir: bir CMS webhook'u, bir
yönetim formu. Bir action bunu bildirimsel olarak,
[sonucundaki `InvalidateTags`](/docs/forms-and-actions#invalidating-what-an-action-changed)
ile yapabilir; bu yanıt yazılmadan önce çalışır — böylece az önce değiştirdiği
sayfaya yönlendirilen bir okuyucu eski sürümü asla görmez.

### Etiket dizini süreç başınadır ve sınırlıdır

collage'ın hangi anahtarın hangi etikete ait olduğuna dair kaydı bellekte, onları
yazan süreçte yaşar. Bundan iki sonuç çıkar.

**Bir yük dengeleyicinin arkasında her instance yalnızca kendi anahtarlarını
bilir.** Yerleşik önbelleklerle bu sorun değildir, çünkü her instance'ın kendi
önbelleği de vardır. Kendinize ait paylaşılan bir depoyla — `Cache.Store`, örneğin
bir Redis — `collage.TaggedCache`'i de uygulayın ki depo etiketleri kendisi
dizinlesin; collage bu durumda ondan etikete göre geçersiz kılmasını da ister ve depo
başka bir instance'ın yazdığına da ulaşabilir. Yerleşik disk önbelleği bunu yapar;
etiketlerinin yeniden başlatmadan sonra da çalışması bu yüzdendir.

**Etiket başına kayıt `Cache.MaxKeysPerTag` ile sınırlıdır** (varsayılan 10000).
Query string anahtarın parçasıdır, dolayısıyla bir istemci tek bir sayfa için
istediği kadar anahtar üretebilir; sınır olmasa dizin sonsuza kadar büyürdü. Bir
etiket sınıra ulaştığında, altında kaydedilmiş en eski anahtar o dizin tarafından
unutulur. Yerleşik bellek ve disk önbelleklerinin ikisi de `collage.TaggedCache`'i
uygular ve girdi başına kendi etiket kayıtlarını tutar; bu yüzden onlarla etiketi
geçersiz kılmak yine her girdiye ulaşır — yalnızca `InvalidateTagsN`'in bildirdiği
sayı eksik kalır. `TaggedCache`'i uygulamayan kendi deponuzun elinde ise yalnızca
collage'ın dizini vardır: unutulan bir sayfa süresi dolana kadar depoda kalır ve
etiketi geçersiz kılmak artık ona ulaşmaz. Sınırı, bir etiketin gerçekten
kapsayabileceği önbellekteki URL sayısının üstüne ayarlayın; sınır istemiyorsanız
negatif verin.

## Anahtardaki query parametreleri

Varsayılan olarak ham query string'in tamamı anahtarın parçasıdır. Tek güvenli
varsayılan budur: data handler isteğin tamamını alır ve `rc.Request.URL.Query()`'yi
okuyabilir; yani `?page=2` farklı bir sayfa olabilir ve collage bunu bilemez.

Ama bu pahalıdır da. `?utm_source=newsletter` içeren bir bülten bağlantısı sayfanın
ikinci bir kopyasını saklar; varyantları deneyen bir tarayıcı botu da önbelleği
kimsenin istemediği kopyalarla doldurur ve gerçek olanları dışarı iter. Sayfanın
hangi parametreleri okuduğunu söyleyin:

```go
collage.NewPage("articles").
	WithLayout(layout).
	WithContent(list).
	WithPath("en", "/articles").
	WithCacheParams("page", "sort").
	Incremental(time.Minute).
	Build()
```

Artık anahtarda yalnızca `page` ve `sort` vardır ve sabit bir sıraya konurlar; böylece
`?page=2&sort=new` ile `?sort=new&page=2` tek bir girdiyi paylaşır. Query'deki geri
kalan her şey önbellek açısından yok sayılır — handler onları yine okuyabilir, ama
sayfanın gösterdiğini değiştirmemelidir.

Hiç ad verilmeden çağrılan `WithCacheParams()`, query'yi anahtardan tamamen çıkarır:
sayfa, query ne derse desin aynı render edilir.

Sayfanın okumadığı bir parametreyi adlandırmanın bir maliyeti yoktur. Okuduğu birini
adlandırmamak ise bir hatadır: iki farklı sayfa tek bir girdiyi paylaşır ve bir
okuyucuya başkasınınki sunulur. Bazı sayfalar da hiç önbelleğe alınmamalıdır — bir
arama sayfasının anahtarı arama terimi, yani okuyucunun yazdığı her neyse o olurdu.
Böyle bir sayfa `Dynamic()` ister.

Document'larda da aynı `WithCacheParams` vardır.

## Bellek mi, disk mi

Varsayılan olan `Type: "memory"`, sayfaları süreç içinde tutar. Hızlıdır ve her
yeniden başlatmadan sonra boştur. En fazla `MaxEntries` sayfa tutar (varsayılan
10000; sınırsız için negatif) ve dolduğunda ekleme sırasına göre en eskiyi atar —
bir sayfayı okumak onu gençleştirmez. Süresi dolmuş bir girdi, bir sonraki
aranışında atılır.

`Type: "disk"` sayfaları dosya olarak tutar; böylece bir yeniden başlatma her şeyi
yeniden render ettirmez:

```go
Cache: collage.CacheConfig{
	Enabled: true,
	Type:    "disk",
	Dir:     ".cache",
},
```

`Dir`'in varsayılanı yoktur: dosyaların nereye yazılacağını kendisi seçen bir
framework, onları kimsenin bakmadığı bir yere yazar. Onu `.gitignore`'a ekleyin.
v0.11.0'dan itibaren oluşturulamayan bir dizin — salt okunur bir dosya sistemi,
yazacak yeri olmayan bir container — başlamamak için bir neden değildir: collage bir
uyarı loglar ve bunun yerine bellekte önbelleğe alır.

### İsim alanı

Bir disk önbelleği, onu dolduran süreçten daha uzun yaşar; bu hem işin özü hem de
bir tehlikedir. Şablonu değişmiş yeni bir binary, eski binary'nin render ettiği
HTML'i sunmamalıdır.

Bu yüzden girdiler, build'e göre adlandırılmış bir alt dizinde yaşar.
`Cache.Version`'ı boş bırakırsanız bu ad, çalışan yürütülebilir dosyanın bir
hash'idir: tam olarak içine derlenen kodunuz ya da şablonlarınız değiştiğinde
değişir ve aynı build'in iki çalıştırması — ya da onu çalıştıran bir filodaki her
makine — tek bir önbelleği paylaşır. `Version`'ı yalnızca çıktının neye benzeyeceğine
binary'nin dışındaki bir şey karar veriyorsa, örneğin bir içerik revizyonu, kendiniz
ayarlayın.

İstek sahteciliği koruma anahtarı isim alanının parçası değildir; bu yüzden disk
önbelleği, `Security.CSRFKey` ayarlı olsun ya da olmasın bir yeniden başlatmadan sağ
çıkar. Anahtar yine de tek bir tür sayfa için önemlidir: içinde form olan önbellekteki
bir sayfa, her okuyucunun token'ının gideceği yerde anahtardan türetilmiş bir işaret
taşır (bkz.
[Formlar ve action'lar](/docs/forms-and-actions#pages-with-forms-are-still-cached))
ve bir anahtarla saklanmış sayfa başka bir anahtarla sunulamaz. Dolayısıyla başka bir
anahtarın işaretini taşıyan saklanmış bir sayfa — anahtar değişmeden önce ya da
kendi anahtarını üreten bir süreç tarafından render edilmiş — ıska sayılır: atılır ve
yeniden render edilir. Form içermeyen sayfalar anahtardan hiç etkilenmez.

Aynı `Dir`'i ve aynı build'i paylaşan her şey girdileri de paylaşır; tek bir test
binary'sindeki iki uygulama da buna dahildir. Her teste `t.TempDir()` ile kendi
dizinini verin — bkz. [Test](/docs/testing#isolate-the-disk-cache).

### Kendi deponuz

`Cache.Store` herhangi bir `collage.Cache` uygulamasını alır ve bu durumda `Type`
yok sayılır. Ana şalter yine `Enabled`'dır. Ona tipli bir nil pointer atamayın —
bir interface alanındaki nil `*myCache`, nil bir interface değildir ve collage
doğrudan onun üzerinden çağrı yapar.

## ETag'ler ve 304

Önbellekteki her sayfa, içeriğinin bir hash'i olan bir ETag ile saklanır ve
önbellekten sunulan her yanıt onu taşır. Sayfaya zaten sahip olan bir tarayıcı ya da
CDN onu `If-None-Match` içinde geri gönderir; hâlâ eşleşiyorsa collage gövdesiz bir
`304 Not Modified` ile yanıt verir.

`Static()` sayfaları yeniden doğrulamayı ucuz yapan budur: `must-revalidate`,
istemcinin her seferinde sorması demektir ve yanıt genellikle birkaç bayttır.

Form içeren bir sayfa yine istisnadır. Okuyucuya gönderilen, onun kendi sahtecilik
token'ını taşır; bu yüzden ETag'i saklanan kopyayı değil o okuyucunun kopyasını
adlandırır ve yanıt `private, no-store`'dur: onu geri gönderen bir tarayıcı yalnızca
kendisine verilen kopya için `304` alır ve paylaşılan hiçbir şey onu tutmaz.

## Eşzamanlı ıskalar tek bir kez render edilir

Popüler bir sayfanın süresi dolduğunda, ilk yeniden render bitmeden gelen her istek
bir ıskadır (miss). Kendi hâline bırakılsa her biri render ederdi: aynı sayfa, aynı
dış servis çağrıları, aynı anda ve trafikle birlikte artan sayıda.

collage buna izin vermez. Bir anahtar için ilk istek render eder; bu arada gelen
diğerleri onu bekler ve aynı baytlarla sunulur. Yapılandırılacak bir şey yoktur.

- **Yalnızca önbelleğe alınan route'lar birleştirilir.** Bir `Dynamic()` sayfanın
  önbellek anahtarı yoktur; dolayısıyla, sayfanın istediği gibi, iki istek iki
  render demektir. Önbelleğe alınan bir [document](/docs/documents) v0.12.0'dan
  itibaren bir sayfa gibi birleştirilir: birçok istemcinin yokladığı, süresi dolan
  bir feed handler'ını bir kez çalıştırır.
- **Bir okuyucunun vazgeçmesi diğerlerini başarısız kılmaz.** Bağlantısı kapanan bir
  istek beklemeyi bırakır. Render eden isteğin kendisi iptal edilirse, arkasında
  bekleyenler onun hatasını almak yerine yeniden dener.
- **Görünürdür.** Bu şekilde sunulan bir istek metriklerinize iki kez bildirilir:
  araması hiçbir şey bulamadığında bir `CacheMiss` olarak, ardından başka bir
  isteğin render'ıyla sunulduğunda `CacheCoalesced` olarak. Dolayısıyla bir
  anahtarın maliyeti olan render sayısı, ıskalarından birleştirilmiş olaylarının
  çıkarılmasıyla bulunur. Sürekli tırmanan bir birleştirme sayısı, bir sayfanın
  render edilebildiğinden daha hızlı süresinin dolduğu anlamına gelir; çok kısa bir
  `Incremental` TTL'i dışarıdan böyle görünür.

## Geliştirmede önbellek asla okunmaz

`Config.DevMode` (ya da `Template.DevMode`) açıkken — iskeleti oluşturulmuş bir
proje bunu `collage dev` altında açar — önbellekteki sayfalar asla sunulmaz.
Geliştirmede şablonlar her istekte diskten yeniden yüklenir ve önbellekteki bir sayfa,
az önce yaptığınız düzenlemeyi TTL'i boyunca gizlerdi. Sayfalar yine yazılır ve
etiketler yine izlenir; böylece hook'lar ve metrikler production'da olacağı gibi
davranır. Ortadan kalkan, düzenlemeden önce render edilmiş bir sayfanın sunulmasıdır.

Geliştirmede disk önbelleğinin yerini bir bellek önbelleği alır ve collage bunu
yaptığını loglar. `collage.Cached` de orada render'lar arasında hiçbir şey tutmaz
(aşağıya bakın).

## Veriyi sayfalar arasında önbelleğe almak

Sayfa önbelleği bütün sayfaları saklar. Her biri yazarını gösteren otuz farklı blog
yazısı için hiçbir şey yapmaz: her biri farklı bir URL'dir, her biri kendi başına
render edilir ve her biri yazarı çeker. Hepsini render eden bir statik dışa aktarma,
yazarı otuz kez çeker.

`collage.Cached` yazarı saklar:

```go
func authorCard(ctx context.Context, rc *collage.RenderContext) (Author, []string, error) {
	id := rc.Param("author")
	author, err := collage.Cached(rc, "author:"+id, time.Hour, []string{"author:" + id},
		func(ctx context.Context) (Author, error) {
			return api.Author(ctx, id)
		})
	return author, nil, err
}
```

```go
func Cached[T any](rc *RenderContext, key string, ttl time.Duration, tags []string,
	fetch func(context.Context) (T, error)) (T, error)
```

`author:ada`'yı isteyen ilk render `fetch`'i çağırır; sonraki her render, hangi
sayfada olursa olsun, saklanan değeri alır. `Cache.Enabled` açıkken iki yazarın otuz
yazısı artık — sunulsun ya da dışa aktarılsın — iki yazar isteği yapar. (Kapalıyken
ve geliştirmede depo yoktur; bkz. [aşağısı](#where-it-keeps-nothing).) Bir
[document'ın](/docs/documents) handler'ı da aynı depoyu paylaşır; dolayısıyla bu
yazarları okuyan bir sitemap ya da feed hiçbirini yeniden çekmez.

### İki önbellek için tek etiket kümesi

Verdiğiniz `tags`, data handler onları döndürmüş gibi sayfanın kendi etiketlerine
eklenir. Böylece tek bir çağrı —

```go
app.InvalidateTags(ctx, "author:ada")
```

— saklanan yazarı *ve* onu gösteren önbellekteki her sayfayı birlikte atar.

Bu olmadan yanlış yapması kolay olan kısım budur. Yazarları kendi API istemcisi
içinde memoize eden ve sayfaları etikete göre geçersiz kılan bir uygulamanın, aynı
anda temizlenmesi gereken iki önbelleği vardır. Yalnızca sayfaları temizlerseniz,
yerini almaları gereken bayat yazardan yeniden render edilirler.

### Nasıl davranır

- **Anahtar sizindir.** Değeri bütün uygulama genelinde adlandırır; bu yüzden onu
  fetch kadar belirgin yapın: `author` değil, `author:ada`. Tek bir anahtarı iki farklı
  tip olarak istemek `collage.ErrCachedTypeMismatch`'tir.
- **Anahtar başına aynı anda tek fetch.** Bir anahtar çekilirken onu isteyen
  render'lar kendi fetch'lerini başlatmak yerine o fetch'i bekler.
- **Hatalar saklanmaz.** Bekleyen herkes hatayı alır; bir sonraki render yeniden
  dener.
- **Geçersiz kılınmış bir fetch saklanmaz.** Bir anahtarın fetch'i hâlâ sürerken
  etiketleri geçersiz kılınırsa, sonuç bekleyenlere gider ama tutulmaz — tam da
  geçersiz kılmanın yerine koymak istediği şeydir.
- **`ttl` sayfanınkinden bağımsızdır.** Hiçbir şey geçersiz kılmadığında değerin ne
  kadar tutulacağını sınırlar; sıfır, bir şey geçersiz kılana kadar tutar. Her dakika
  yeniden render edilen bir sayfa bir saat önce çekilmiş bir yazarı yine
  kullanabilir; amaç da budur.
- **Bellekte, sınırlı.** Değerler, sayfalar diskte önbelleğe alınsa bile süreç
  içinde, `Cache.MaxEntries`'e kadar tutulur; önce en uzun süredir kullanılmayan
  gider. Çok instance'lı bir yayına almada her instance kendi değerlerini tutar.

### Hiçbir şey tutmadığı yerler

`Cache.Enabled` false iken, geliştirmede, middleware'i `collage.SkipCache` çağıran
bir istekte (bir [önizleme](/docs/previews)) ve bir action'ın kendi handler'ı
içinde `Cached` hiçbir şey saklamaz ve `collage.Once` gibi davranır: aynı render'ın
fragment'leri tek bir fetch'i paylaşır, bir sonraki render yeniden çeker. Bu yüzden
bir önizleme taze bir sayfanın yanı sıra taze veri de görür.

### Once, Cached ve sayfa önbelleği

| | Paylaşıldığı yer | Ömrü |
| --- | --- | --- |
| `collage.Once` | tek bir render'ın fragment'leri | o render |
| `collage.Cached` | süreçteki her render | `ttl`'i ya da etiketleri geçersiz kılınana kadar |
| sayfa önbelleği | tek bir URL için her istek | sayfanın stratejisi ya da etiketleri geçersiz kılınana kadar |

`Once`'ı tek bir sayfanın iki kez çektiği şeyler için, `Cached`'i birçok sayfanın
çektiği şeyler için, sayfa önbelleğini de sayfanın kendisi için kullanın. Birlikte
çalışırlar: önbellekteki bir sayfa render edilmez, dolayısıyla fetch'lerinin hiçbiri
çalışmaz. `Once` için bkz. [Data handler'lar](/docs/data-handlers).
