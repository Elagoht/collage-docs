---
description: Fragment'ler, sundukları slot'lar, bir fragment başarısız olduğunda ne olduğu ve render sırasında içerikten doldurulan slot'lar.
reference: NewFragment, NewInlineFragment, InlineHTML, FragmentBuilder, Fragment, FragmentBuilder.WithFallback, FragmentBuilder.WithSlot, FragmentBuilder.WithData, FragmentBuilder.WithoutTypeCheck, FragmentBuilder.Static, FragmentBuilder.Shared, Data, Value, SlotResolverFunc, SlotRule, SlotRequired, SlotSingle, ErrUnknownSlot
---

# Fragment'ler ve slot'lar

**Fragment**, bir page'in yapı taşıdır. Üç kısımdan oluşur:

- bir template;
- template'in render ettiği veriyi çeken, isteğe bağlı bir data handler;
- başka fragment'lerin yerleştiği isimli **slot**'lar.

Layout da bir fragment'tir. Bir page'in içeriği, bir sidebar, bir yazar kartı ve bir
yorum listesi de birer fragment'tir. Bir page, kökünde layout'un durduğu bir fragment
ağacıdır.

<figure class="diagram-figure" data-search-ignore>
<svg class="diagram" viewBox="0 0 320 250" role="img" aria-labelledby="fs-tree-title">
<title id="fs-tree-title">Ağaç olarak bir ürün page'i: layout'un content slot'u, required olan product fragment'ini tutar. Product fragment'inin iki slot'u vardır: stock, failure policy'si olmayan stok kutusunu; similar ise fallback'i olan benzer ürünleri tutar. Product başarısız olursa page 500 ya da 404 ile başarısız olur. Stok kutusu başarısız olursa dışarıda kalır. Benzer ürünler başarısız olursa fallback render edilir. Bu iki durumda page 200 döner.</title>
<rect class="d-frame" x="4" y="4" width="312" height="242" rx="8" />
<text x="14" y="21">layout</text>
<rect class="d-bar" x="70" y="12" width="236" height="10" rx="3" />
<rect class="d-slot d-slot-a" x="14" y="30" width="292" height="206" rx="5" />
<text class="d-slot-text-a d-small" x="24" y="44">slot "content"</text>
<rect class="d-frame d-frame-a" x="24" y="52" width="272" height="174" rx="6" />
<text class="d-slot-text-a" x="34" y="69">product</text>
<text class="d-small" x="86" y="69">hata: 500 / 404</text>
<rect class="d-chip-hit" x="196" y="57" width="90" height="18" rx="9" /><text class="d-hit-text d-small" x="241" y="69" text-anchor="middle">Required()</text>
<rect class="d-bar" x="34" y="80" width="160" height="6" rx="3" />
<rect class="d-bar" x="34" y="90" width="110" height="6" rx="3" />
<rect class="d-slot d-slot-b" x="34" y="102" width="122" height="116" rx="4" />
<text class="d-small" x="95" y="116" text-anchor="middle">slot "stock"</text>
<rect class="d-b" x="44" y="124" width="102" height="30" rx="4" /><text class="d-on-b" x="95" y="143" text-anchor="middle">stock</text>
<rect class="d-chip" x="44" y="162" width="102" height="18" rx="9" /><text class="d-small" x="95" y="174" text-anchor="middle">policy yok</text>
<text class="d-small" x="95" y="196" text-anchor="middle">hata: boş kalır</text>
<text class="d-small" x="95" y="209" text-anchor="middle">page: 200</text>
<rect class="d-slot d-slot-c" x="164" y="102" width="122" height="116" rx="4" />
<text class="d-small" x="225" y="116" text-anchor="middle">slot "similar"</text>
<rect class="d-c" x="174" y="124" width="102" height="30" rx="4" /><text class="d-on-c" x="225" y="143" text-anchor="middle">similar</text>
<rect class="d-chip" x="174" y="162" width="102" height="18" rx="9" /><text class="d-small" x="225" y="174" text-anchor="middle">WithFallback</text>
<text class="d-small" x="225" y="196" text-anchor="middle">hata: fallback</text>
<text class="d-small" x="225" y="209" text-anchor="middle">page: 200</text>
</svg>
<figcaption>Bir ürün page'i bir ağaçtır. Layout'un <code>content</code> slot'u ürünü tutar; ürünün template'i de kendine ait iki slot'u çağırır. Her fragment'in kendi failure policy'si vardır: ürün required'dır, bu yüzden başarısız olursa page de başarısız olur. Stok kutusunun policy'si yoktur, bu yüzden dışarıda kalır. Benzer ürünlerin bir fallback'i vardır. Page'in tamamı <a href="#example-a-product-page">Örnek: bir ürün page'i</a> bölümünde kurulur.</figcaption>
</figure>

```go
author := collage.NewFragment("author", "fragments/author.html").
	WithData(collage.LoadTagged(loadAuthor)).
	WithTimeout(time.Second).
	WithFallback(anonymousAuthor).
	Build()
```

## Fragment oluşturmak

`collage.NewFragment(name, templatePath)` bir builder başlatır. `Build()` ise
`*collage.Fragment`'i döner.

- **İsim**, fragment'in raporlarda nasıl görüneceğini belirler: hata mesajlarında,
  development error panel'inde, render metadata'sında ve trace'lerde. İsim,
  fragment'in page'in hangi kısmı olduğunu anlatmalıdır: `"list"` değil,
  `"post-comments"`.
- **Template path'i** template root'una göredir ve uzantıyı da içerir:
  `"fragments/author.html"`, `templates/fragments/author.html` dosyasıdır. Template
  engine'in yüklemediği bir path, fragment'i içeren page register edilirken
  `ErrTemplateNotFound` hatası verir. v0.11.0'dan beri bu kural, page'in
  `WithFragmentPath` ile kendi URL'sinde açtığı fragment'ler için de geçerlidir.
  Yalnızca bir [slot resolver](#slots-filled-per-render)'ın döndüğü fragment daha
  sonra, render edilirken kontrol edilir. Bkz. [Template'ler](/docs/templates).

Geri kalan her şey isteğe bağlıdır:

| Metot | Neyi ayarlar |
| --- | --- |
| `WithData(d)` | Template'in verisini: `collage.LoadTagged` ya da `collage.Load` ile bir handler'ı, `collage.Value` ile program başlarken sabitlenen bir değeri ya da `collage.Effect`'i. Bkz. [Data handler'lar](/docs/data-handlers) |
| `WithoutTypeCheck()` | Template'i, verisinin tipine göre yapılan kontrolün dışında bırakır. Bkz. [Template'ler nasıl kontrol edilir](/docs/data-handlers#how-templates-are-checked). v0.49.0'dan beri |
| `WithTitle(s)` | Page'in `<title>`'ını, handler olmadan. Bkz. [Head ve SEO](/docs/head-and-seo) |
| `WithSlot(name, rules...)` | Bir slot'u required yapar ya da tek fragment'le sınırlar |
| `WithSlotFragment(slot, child)` | Bir child fragment'i bir slot'a bağlar |
| `WithSlotResolver(slot, resolve)` | Slot'u bunun yerine her render'da doldurur |
| `Required()` | Bu fragment başarısız olursa page de başarısız olur |
| `WithFallback(f)` | Bu fragment başarısız olduğunda neyin render edileceğini |
| `WithTimeout(d)` | Data handler'ının ne kadar sürebileceğini |
| `Static()` | Data handler'ının aynı URL'e gelen her request'e aynı sonucu döndürdüğünü, böylece page'i dynamic yapmadığını. v0.17.0'dan beri |
| `Shared()` | Data handler'ının belli bir anda her okuyucuya aynı sonucu döndürdüğünü, ama bu sonucun zamanla değiştiğini. Böylece render'ı birçok okuyucuya gönderilebilir ve page'in stratejisine dokunulmaz. v0.19.0'dan beri |

Page builder'da olduğu gibi fragment builder da hataları zinciri durdurmadan kaydeder
ve `BuildErr()` bunları döner.

Kaydedilen hatalar, builder'ın oluşturduğu fragment'in üzerinde kalır. Bu yüzden
iki kez kısıtlanmış bir slot, teke sınırlanmış bir slot'a bağlanan ikinci bir child
ya da negatif bir timeout gibi hataların her biri, ağacında bu fragment bulunan her
page'i register sırasında durdurur. Birinin `BuildErr()`'ü
çağırıp çağırmadığı fark etmez. Hatayı `RegisterPage`'de değil de ona yol açan
satırda almak istediğinizde `BuildErr()`'ü çağırın. Bkz. [Page'ler ve
layout'lar](/docs/pages-and-layouts#building-a-page).

Data handler'ı olmayan bir fragment, template'ini veri olmadan ya da
`collage.Value(v)`'nin her render'da verdiği değerle render eder. Bu, hiç değişmeyen
markup için doğru seçimdir: bir footer, static bir duyuru ya da bir link listesi
gibi. Tek işi slot'ları yerleştirmek olan bir layout için de doğrudur.

Page'i cache'lenebilir de tutar: strateji tanımlamayan bir page, render ettiği bir
şeyin data handler'ı ya da slot resolver'ı olmadıkça static'tir. Bkz.
[Caching](/docs/caching#a-page-that-declares-none). İki söz, bir handler'ın nasıl
hesaba katılacağını değiştirir:

- Handler'ı yalnızca path'in parametrelerini ve locale'i okuyan bir fragment
  `Static()` der ve handler'ı artık hesaba katılmaz.
- Handler'ı her okuyucu için aynı olan ama zaman içinde aynı kalmayan bir fragment
  (bir ölçüm gibi) `Shared()` der. Bu, page hakkında hiçbir şey söylemez ve handler
  yine hesaba katılır.

İki sözden birini bozan bir handler, bir okuyucunun verisini başka bir okuyucuya
gönderir. Bir fragment'in verisini iki kez ayarlamak, register sırasında
`ErrConflictingData` hatası verir.

### Inline template'ler

Küçük bir fragment, bir dosya belirtmek yerine template'ini kendisi taşıyabilir
(v0.29.0'dan beri). `collage.NewInlineFragment(name, html)` aynı builder'ı döner,
bu yüzden zincirin geri kalanı değişmez:

```go
row := collage.NewInlineFragment("post-row", `
  <tr>
    <td>{{.Title}}</td>
    <td>{{template "partials/date.html" .Date}}</td>
  </tr>`).
	WithData(collage.LoadTagged(loadRow)).
	Build()
```

Bir dosya template'i gibi render edilir: slot'lar, `hoist`, her
[template fonksiyonu](/docs/template-functions) ve template dizinine yapılan
`{{template}}` çağrıları aynen çalışır. Register işlemi de onu bir dosya template'i
gibi parse edip kontrol eder. Bir parse hatası ya da template'in hiç çağırmadığı bir
slot, page'i ve fragment'i adıyla belirterek startup'ı durdurur.

İki fragment aynı ismi paylaşıp farklı template'ler taşıyabilir; her biri kendi
template'ini render eder. Development error panel'i onu `inline template of fragment "post-row"`
olarak adlandırır.

Daha uzun bir template, `collage.InlineHTML` olarak tanımlanmış ayrı bir sabit
olabilir (v0.30.0'dan beri). Bu, `string` için bir addır ve
[VS Code extension'ı](/docs/installation#editor-support) onu HTML olarak
renklendirir ve düzenler:

```go
const loginForm collage.InlineHTML = `
  <form method="post">
    {{csrfToken}}
    <input type="email" name="email" required>
  </form>`

content := collage.NewInlineFragment("login", loginForm).Build()
```

Inline template'i, bir page'in birkaç satırlık markup'tan oluşan ve onu besleyen
handler'ın hemen yanında duran kısımları için kullanın: bir tablo satırı, bir
buton, bir form alanı. Layout'lar ve bütün page'ler dosya olarak daha okunaklıdır.

Template koddur, bu yüzden sabit olmalıdır. Onu asla veriden oluşturmayın:
`NewInlineFragment("row", "<p>"+post.Title+"</p>")`, başlığın içerdiği her
`{{…}}` ifadesini çalıştırır ve her farklı string, program kapanana kadar saklanan
bir template'e dönüşür.

Bu en çok, fragment'leri her request'te oluşturan bir [slot
resolver](#slots-filled-per-render)'da önemlidir. Resolver'ın fragment'leri
inline olabilir, ama template'leri sabittir ve veri onlara `collage.Value` ya da bir
data handler üzerinden ulaşır.

Bunun dışında üç sınırı daha vardır:

- Go'nun raw string'i backtick içeremez. Bu yüzden JavaScript template literal
  içeren bir template dosyada kalır.
- Inline bir template, kendine ait template'leri `{{define}}` ya da `{{block}}` ile
  tanımlayamaz. Böyle bir tanım, aynı isimli bir dosya template'inin yerini her
  page'de alacağı için register sırasında `ErrSourceConflict` hatası verir.
- Bir dosya template'inden inline bir template'i `{{template}}` ile çağırmak
  desteklenmez: template kümesindeki ismi internal'dır.

Bir fragment ya bir dosya belirtir ya da bir template taşır; ikisi birden olamaz
(`ErrConflictingTemplate`). Boş bir template ise `ErrEmptyTemplatePath`
hatası verir. [`collage inspect`](/docs/cli#collage-inspect) inline bir fragment'i
`"inline": true` ile işaretler ve `template` alanını boş bırakır.

## Slot'lar

Slot, bir fragment'in template'inde `{{slot "name"}}` ile yazılan isimli bir
konumdur. Slot'u tanımlamak için template'te onu çağırmak yeterlidir. Fragment,
child'ları ona ismiyle bağlar:

```go
post := collage.NewFragment("post", "pages/post.html").
	WithData(collage.LoadTagged(loadPost)).
	WithSlot("author", collage.SlotRequired, collage.SlotSingle).
	WithSlotFragment("author", author).
	WithSlotFragment("related", relatedPosts).
	WithSlotFragment("related", popularPosts).
	Build()
```

```html
<!-- templates/pages/post.html -->
<article>
  <h1>{{.Title}}</h1>
  {{slot "author"}}
  <div class="body">{{.Body}}</div>
</article>
<aside>{{slot "related"}}</aside>
```

Bir slot istediği sayıda fragment tutar. Bu fragment'ler bağlandıkları sırayla, art
arda render edilir. Yukarıdaki `related` iki fragment tutar. Hiçbir şey bağlanmamış
bir slot hiçbir şey render etmez. Bir layout'ta da tanımlama gerekmez: register
işlemi page'in içeriğini layout'un `"content"` slot'una yerleştirir.

`WithSlot(name, rules...)`, bu davranışı istemediğiniz durumlar içindir. Yukarıdaki
`author` doldurulmak zorundadır ve tek fragment tutar. `WithSlot`, kısıtladığı
bağlamalardan önce de sonra da gelebilir. Her kural bir `collage.SlotRule`'dur:

- **`collage.SlotRequired`**: Slot'ta bir şey olmak zorundadır. Hiçbir şey bağlanmamış required
  bir slot, register sırasında reddedilir (`ErrRequiredSlotUnfilled`). Template
  çalışmadan önce de tekrar kontrol edilir. Böylece template bu slot'u hiç
  kullanmasa bile hata yakalanır.
- **`collage.SlotSingle`**: Slot en fazla bir fragment tutar. İkinci bir fragment
  bağlamak `ErrSlotOccupied` kaydeder.

Başka herhangi bir değer `ErrInvalidSlotDefinition` kaydeder. v0.60.0'dan önce bu
ikisi boolean'dı: `WithSlot(name, required, allowMultiple)`.

`WithSlotFragment`, nil bir child için `ErrNilFragment` kaydeder. `WithSlot`'u aynı
isimle iki kez çağırmak `ErrDuplicateSlot` kaydeder.

Template'te `{{slot "name"}}`, slot'un tuttuğu içeriği HTML olarak render eder ve bu
HTML tekrar escape edilmez. Çünkü child'lar kendi değerlerini render edilirken zaten
escape etmiştir.

Bir **bağlamanın iki tarafından birindeki yazım hatası**, örneğin
`{{slot "sidebar"}}` karşısında `WithSlotFragment("sidbar", ...)`, register işlemini
`ErrUnknownSlot` ile başarısız kılar. Hata, slot'u ve template'in gerçekten çağırdığı
slot'ları adlarıyla belirtir. Template'inin hiç çağırmadığı bir slot'a bağlanan bir
fragment zaten hiçbir zaman render edilemez.

Template'in include ettiği bir template'teki ya da tanımladığı bir block'taki
çağrılar da sayılır. Bir slot'u literal dışında bir şeyle, `{{slot .Which}}` ile adlandıran bir template herhangi
birini çağırabilir. Bu yüzden o fragment'in bağlamaları kontrol edilmez.

### Her fragment'in kendi verisi vardır

Bir child, parent'ının verisini görmez. `fragments/author.html` içinde `.`, post
değil, `loadAuthor`'ın döndüğü değerdir. Parent'ın çektiği bir şeye ihtiyacı olan
bir child, bunu parent'ının set ettiği bir `collage.Key` üzerinden okur. Ya da veriyi `collage.Once`
veya `collage.Cached` ile kendisi ister, böylece veri yalnızca bir kez çekilir.
Bkz. [Data handler'lar](/docs/data-handlers#sharing-data-between-fragments).

### Fragment'leri yeniden kullanmak

Bir `*collage.Fragment`, birçok page'deki birçok slot'a bağlanabilir. Yukarıdaki
`author` hem post page'inde hem de bir arama sonuçları page'inde görünebilir. İki kez
bağlanan bir fragment iki kez render edilir ve data handler'ı her bağlama için bir kez
çalışır.

Bir fragment kendisini içeremez. Bir fragment'i doğrudan ya da bir child zinciri
üzerinden kendi slot'una bağlarsanız, bu register sırasında `ErrFragmentCycle` ile
reddedilir.

## Bir fragment başarısız olduğunda

Bir fragment şu durumlarda başarısız olur:

- data handler'ı bir hata döndüğünde;
- template'i çalıştırılamadığında;
- required bir slot'u boş kaldığında;
- data handler veya template panic ettiğinde. Bir data handler'daki ya da bir
  template fonksiyonundaki panic recover edilir ve bir `*collage.PanicError`'a
  çevrilir. Bu durum çöken bir process olarak değil, bir başarısızlık olarak ele
  alınır.

Bundan sonra ne olacağını fragment'in **failure policy**'si belirler. Üç policy
vardır.

```go
postContent := collage.NewFragment("post", "pages/post.html").
	WithData(collage.LoadTagged(loadPost)).
	Required().
	Build()

comments := collage.NewFragment("comments", "fragments/comments.html").
	WithData(collage.LoadTagged(loadComments)).
	WithFallback(collage.NewFragment("comments-unavailable", "fragments/comments-unavailable.html").Build()).
	Build()

related := collage.NewFragment("related", "fragments/related.html").
	WithData(collage.LoadTagged(loadRelated)).
	Build()
```

- **`Required()`**: Fragment başarısız olursa page de başarısız olur. Bunu, page'in
  asıl göstermek için var olduğu içerik için kullanın. Page'in error page'i 500 ile
  sunulur. Hata `collage.ErrNotFound`'u wrap ediyorsa not-found page'i 404 ile
  sunulur. Başarısızlık, üstündeki optional fragment'lerden geçerek yukarı taşınır.
  Optional bir sidebar'ın içindeki required bir fragment de page'i başarısız kılar.
- **`WithFallback(f)`**: Fragment'in yerine `f` render edilir. Fallback de başarısız
  olursa fragment hiçbir şey render etmez ve page yine başarılı olur. Fallback bir
  başarısızlığı sınırlamak için vardır. Bu yüzden içindeki bir şey required olarak
  işaretlenmiş olsa bile fallback'in kendi başarısızlığı da sınırlanır.
- **Hiçbiri**: Fragment hiçbir şey render etmez ve page onsuz sunulur.

`collage.ErrNotFound` yalnızca required bir fragment'te "404" anlamına gelir. Optional
bir fragment'te diğerleri gibi sıradan bir başarısızlıktır, çünkü page onsuz da
render edilebilir.

Fallback'li ya da fallback'siz, başarısız olmuş herhangi bir fragment'le sunulan
page **degraded** sayılır. Degraded page okuyucuya gönderilir ama asla cache'lenmez.
Böylece sonraki request, hatayı bir TTL boyunca sunmak yerine yeniden dener.
Static export, açıkça söylenmedikçe degraded bir page'i yazmaz.

Development'ta bir hata asla sessiz kalmaz:

- Hiçbir şey render etmeyen, başarısız olmuş bir fragment, çıktısının olacağı yere
  ismini ve hatasını içeren bir HTML yorumu bırakır.
- Page, başarısız olan her fragment'i hatasıyla birlikte listeleyen bir panel taşır.
  Template hatalarında panel dosyayı ve satırı da gösterir. Bir fallback fragment'in
  yerini tutmuş olsa bile panel yine görünür.
- Required bir fragment bütün page'i başarısız kıldığında, development error page'i
  o fragment'in ismini verir. Yani hatanın içinden geçerek yukarı çıktığı layout'u
  değil, başladığı yeri gösterir.

Development dışında bunların hiçbiri yoktur.

v0.36.0'dan beri, page'in onsuz sunulduğu bir hata her mode'da
[`Config.Logger`](/docs/configuration) üzerinden bir uyarı olarak log'lanır:
`collage: fragment failed; the page is served without it`. Log satırı page'i,
fragment'i, locale'i, bir fallback'in yerini tutup tutmadığını ve hatayı söyler.

Böyle bir page yine 200 döner; bu yüzden production'da kaybolan bir formun
göründüğü yer log'dur.

### Template'in atladığı bir slot

Bir fragment'in child'ları, template çalışmadan önce veri çekmeye başlar. Böylece
verileri art arda değil, aynı anda çekerler.

Bu yüzden template'in render etmemeye karar verdiği bir slot'taki fragment'in de
handler'ı başlatılmış olur: `{{if .ShowComments}}{{slot "comments"}}{{end}}`.
Template biter bitmez bu fragment'in context'i cancel edilir ve hatası yok sayılır. Hiç render edilmemiş bir
fragment, `Required()` olsa bile page'i başarısız kılamaz.

## Timeout'lar

`WithTimeout(d)`, fragment'in data handler'ına bir süre sınırı koyar. Timeout
ayarlamayan bir fragment, `Config.Template.Timeout` değerini alır. Bu değerin
varsayılanı beş saniyedir. Negatif bir süre `ErrInvalidTimeout` kaydeder.

```go
recommendations := collage.NewFragment("recommendations", "fragments/recommendations.html").
	WithData(collage.LoadTagged(loadRecommendations)).
	WithTimeout(300 * time.Millisecond).
	WithFallback(nothingToRecommend).
	Build()
```

Timeout, fallback ile doğal olarak birlikte kullanılır. Yavaş bir öneri servisi,
page'e okuyucunun sabrına değil, 300 milisaniyeye ve bir fallback'e mal olur.

Timeout handler'ı değil, handler'ın aldığı **context**'i sınırlar. `ctx`'i yok sayan
bir handler istediği kadar çalışır. Bekleyebilecek her çağrıya `ctx`'i geçin. Bkz.
[Data handler'lar](/docs/data-handlers#timeouts-and-the-context).

## Örnek: bir ürün page'i

Bir mağazanın ürün page'i üç şey gösterir ve bunların önemi farklıdır:

| Kısım | Nereden gelir | Başarısız olunca |
| --- | --- | --- |
| Ürün: adı, fiyatı, açıklaması | Katalog | Page'in gösterecek bir şeyi olmaz: `Required()` |
| Bir stok kutusu: "12 adet stokta" | Zaman zaman erişilemeyen bir stok servisi | Kutu dışarıda kalır; policy yok |
| "Benzer ürünler" | Katalog | Yerine kataloğa giden bir link gösterilir: `WithFallback` |

Handler'lar ürünün id'sini path'ten okur. Katalog ve envanter paketleri sizin veri
katmanınızın yerini tutar:

```go
type product struct {
	ID    string
	Name  string
	Price string
}

type stockLevel struct {
	Count int
}

func loadProduct(ctx context.Context, rc *collage.RenderContext) (product, []string, error) {
	id := rc.Param("id")
	p, err := catalog.Product(ctx, id)
	if errors.Is(err, catalog.ErrNoProduct) {
		return product{}, nil, fmt.Errorf("catalog: no product %q: %w", id, collage.ErrNotFound)
	}
	if err != nil {
		return product{}, nil, err
	}
	return p, []string{"product:" + id}, nil
}

func loadStock(ctx context.Context, rc *collage.RenderContext) (stockLevel, error) {
	n, err := inventory.Count(ctx, rc.Param("id"))
	if err != nil {
		return stockLevel{}, err
	}
	return stockLevel{Count: n}, nil
}

func loadSimilar(ctx context.Context, rc *collage.RenderContext) ([]product, []string, error) {
	list, err := catalog.Similar(ctx, rc.Param("id"))
	if err != nil {
		return nil, nil, err
	}
	return list, []string{"catalog"}, nil
}
```

Her kısım kendi policy'si olan bir fragment'tir ve ürün, diğer ikisini kendi
slot'larına bağlar:

```go
stock := collage.NewFragment("product-stock", "fragments/stock.html").
	WithData(collage.Load(loadStock)).
	WithTimeout(300 * time.Millisecond).
	Build()

similar := collage.NewFragment("similar-products", "fragments/similar.html").
	WithData(collage.LoadTagged(loadSimilar)).
	WithFallback(collage.NewFragment("browse-catalogue", "fragments/browse-catalogue.html").Build()).
	Build()

content := collage.NewFragment("product", "pages/product.html").
	WithData(collage.LoadTagged(loadProduct)).
	WithSlotFragment("stock", stock).
	WithSlotFragment("similar", similar).
	Required().
	Build()

page := collage.NewPage("product").
	WithLayouts(layout).
	WithContent(content).
	WithPath("en", "/products/{id}").
	Build()
```

```html
<!-- templates/pages/product.html -->
<article>
  <h1>{{.Name}}</h1>
  <p class="price">{{.Price}}</p>
  {{slot "stock"}}
</article>
<aside>{{slot "similar"}}</aside>
```

`/products/42` için bir request geldiğinde:

1. Ürünün handler'ı çalışır.
2. Handler döner dönmez stock ve similar handler'ları birlikte başlar ve ürünün
   template'i yazmaya başlar.
3. Template `{{slot "stock"}}`'a ulaşınca stok kutusunu bekler. Ardından
   `{{slot "similar"}}` gelir; onun handler'ı o sırada genellikle bitmiştir.

<figure class="diagram-figure" data-search-ignore>
<svg class="diagram" viewBox="0 0 320 216" role="img" aria-labelledby="fs-time-title">
<title id="fs-time-title">Önce product handler'ı çalışır. O dönünce stock ve similar handler'ları birlikte başlar. Stock handler'ı 300 milisaniyede başarısız olur, similar handler'ı biter. Ürünün template'i sırasıyla ürünü, stok kutusu için hiçbir şeyi ve benzer ürünleri yazar; page stok kutusu olmadan 200 döner.</title>
<defs><marker id="fs-time-arrow" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto"><path class="d-head" d="M0 0 8 4 0 8z" /></marker></defs>
<text x="8" y="37">product</text>
<rect class="d-a" x="84" y="24" width="64" height="18" rx="4" /><text class="d-on-a d-small" x="116" y="36" text-anchor="middle">handler</text>
<text x="8" y="65">stock</text>
<rect class="d-b" x="150" y="52" width="120" height="18" rx="4" /><text class="d-on-b d-small" x="210" y="64" text-anchor="middle">300 ms'de hata</text>
<text x="8" y="93">similar</text>
<rect class="d-c" x="150" y="80" width="70" height="18" rx="4" /><text class="d-on-c d-small" x="185" y="92" text-anchor="middle">handler</text>
<line class="d-rule" x1="149" y1="18" x2="149" y2="104" />
<line class="d-flow" x1="84" y1="108" x2="306" y2="108" marker-end="url(#fs-time-arrow)" />
<text class="d-small" x="149" y="122" text-anchor="middle">birlikte başlar</text>
<text class="d-small" x="310" y="122" text-anchor="end">zaman</text>
<text x="8" y="153">yazılan</text>
<rect class="d-a" x="150" y="136" width="50" height="26" rx="4" /><text class="d-on-a d-small" x="175" y="153" text-anchor="middle">product</text>
<g class="d-dropped"><rect class="d-frame" x="204" y="136" width="50" height="26" rx="4" /><text class="d-small" x="229" y="153" text-anchor="middle">boş</text></g>
<rect class="d-c" x="258" y="136" width="52" height="26" rx="4" /><text class="d-on-c d-small" x="284" y="153" text-anchor="middle">similar</text>
<line class="d-flow" x1="230" y1="162" x2="230" y2="180" marker-end="url(#fs-time-arrow)" />
<rect class="d-chip-hit" x="155" y="184" width="150" height="24" rx="12" /><text class="d-hit-text d-small" x="230" y="200" text-anchor="middle">200 · stok kutusuz page</text>
</svg>
<figcaption>Stok servisi çökmüş. Handler'ı <code>ctx</code>'i aktardığı için 300 milisaniyelik timeout o context'i cancel edince vazgeçer. Failure policy'si olmayan stok kutusu hiçbir şey render etmez. Benzer ürünler aynı anda çekildiği için bu hata page'e bir 500'e değil, 300 milisaniyeye ve bir kutuya mal olur.</figcaption>
</figure>

Her hatanın bu page'e etkisi:

- **Stok servisi çökmüş ya da yavaş.** Kutu hiçbir şey render etmez, page 200 döner
  ve hata bir uyarı olarak log'lanır. Development'ta kutunun olacağı yeri bir HTML
  yorumu işaretler.
- **Benzer ürünler çağrısı başarısız.** Onların yerine `browse-catalogue` render
  edilir ve page 200 döner.
- **42 numaralı ürün yok.** Hata, required bir fragment'te `collage.ErrNotFound`'u
  wrap ettiği için okuyucu not-found page'ini 404 ile alır.
- **Kataloğun kendisi başarısız.** Ürün required olduğu için page'in error page'i
  500 ile sunulur.

Stok kutusu olmadan sunulan bir page degraded'dır, bu yüzden asla cache'lenmez:
page `Incremental(ttl)` tanımlamış olsaydı bile sonraki request, page'i bir TTL
boyunca kutusuz tutmak yerine stok servisine yeniden sorardı. Strateji
tanımlanmadığında handler'lar page'i dynamic yapar. Bkz.
[Caching](/docs/caching#a-page-that-declares-none).

## Her render'da doldurulan slot'lar

Buraya kadar anlatılan her şey, child'ları program başlarken bağlar. Bazı page'lerin
düzenini ise içerikleri belirler. Örneğin bölümlerini bir editörün CMS'te seçip
sıraladığı bir landing page ya da kullanıcının seçtiği widget'lardan oluşan bir
dashboard.

Bunları başlangıçta bağlamak, her değişiklikte yeniden başlatma gerektirir. Ayrıca
neyin nereye gideceği konusunda veriyle uyumlu kalması gereken bir kod yazmanız
gerekir.

`WithSlotResolver`, bir slot'u her render'da doldurur. Resolver, render'ın
context'ini ve `RenderContext`'ini alır (context'i v0.60.0'dan beri) ve slot'un bu render'da tutacağı fragment'leri döner:

```go
type block struct {
	Kind    string
	Heading string
	Text    string
}

var blocksKey = collage.NewKey[[]block]("blocks")

landing := collage.NewFragment("landing", "pages/landing.html").
	WithData(collage.Effect(func(ctx context.Context, rc *collage.RenderContext) error {
		blocks, err := cms.Blocks(ctx, "landing")
		if err != nil {
			return err
		}
		blocksKey.Set(rc, blocks)
		return nil
	})).
	WithSlot("blocks", collage.SlotRequired).
	WithSlotResolver("blocks", func(ctx context.Context, rc *collage.RenderContext) ([]*collage.Fragment, error) {
		blocks, _ := blocksKey.Get(rc)
		fragments := make([]*collage.Fragment, 0, len(blocks))
		for i, b := range blocks {
			f, err := blockFragment(i, b)
			if err != nil {
				return nil, err
			}
			fragments = append(fragments, f)
		}
		return fragments, nil
	}).
	Required().
	Build()
```

Bir resolver, program boyunca var olan fragment'leri dönebilir ya da onları o anda
oluşturabilir. Her block'un kendi verisini alması, fragment'leri o anda oluşturarak
sağlanır:

```go
func blockFragment(i int, b block) (*collage.Fragment, error) {
	switch b.Kind {
	case "hero", "text":
		return collage.NewFragment(fmt.Sprintf("block-%d-%s", i, b.Kind), "blocks/"+b.Kind+".html").
			WithData(collage.Value(b)).
			Build(), nil
	}
	return nil, fmt.Errorf("landing: unknown block kind %q", b.Kind)
}
```

Kurallar şunlardır:

- **Resolver, kendi fragment'inin data handler'ından sonra çalışır.** Böylece o
  handler'ın çektiği veriyi okuyabilir. **Döndüğü fragment'ler kendi handler'larını
  başlatmadan önce çalışır.** Bu handler'lar da ardından, diğer child'larda olduğu
  gibi eşzamanlı çalışır.
- **Resolver'ın döndüğü değer slot'un kurallarına tabidir.** Slot birden fazla
  fragment'e izin vermiyorsa en fazla bir fragment dönebilir. Slot required ise en az
  bir fragment dönmelidir (`ErrRequiredSlotEmpty`). Resolver'ın döndüğü bir hata ya
  da resolver'daki bir panic, resolver'ın ait olduğu fragment'i başarısız kılar ve
  o fragment'in failure policy'si uygulanır.
- **Yalnızca bir resolver'a bağlanan slot isteğe bağlıdır ve istediği sayıda fragment
  tutar.** Ondan önce ya da sonra çağrılan `WithSlot`, onu required ya da tek
  fragment'lik yapar. Yukarıdaki `blocks` bu şekilde required'dır. Bir slot ya bir
  resolver ile ya da `WithSlotFragment` ile doldurulur, ikisiyle birden asla
  doldurulmaz. İkisini karıştırmak `ErrSlotResolved` kaydeder.
- **Bir resolver, tıpkı bir data handler gibi, strateji tanımlamayan page'i dynamic
  yapar.** Resolver'ın döndüğü şey request'e bağlı olabilir.
- **Resolver'ın fragment'leri yalnızca render edilirken kontrol edilir.** Register
  işlemi bu fragment'leri göremez. Bu yüzden template'i olmayan bir fragment
  döndürülürse uygulama başlarken değil, o render sırasında hata oluşur. Builder'ın
  kaydettiği bir hata da aynı şekilde yakalanır. Hatalarla oluşturulmuş bir fragment
  döndürülürse, slot'un sahibi olan fragment başarısız olur. Bu durumda o fragment'in
  failure policy'si uygulanır ve hata, döndürülen fragment'in ismini içerir. Hatayı
  resolver'ın içinde ele almak isterseniz `BuildErr()`'ü kendiniz kontrol edin.
  Register işlemi diğer bütün fragment'lerin template'lerini verilerinin tipine
  göre kontrol eder, bunlarınkini ise etmez. Template'lerin hepsi başlangıçta
  yüklenir. Bu yüzden bir resolver'ın kullanabileceği block türleri yine de
  program tarafından belirlenir.

Bölümleri içerikten gelen bir page, o içeriğin tag'lerini de bildirmelidir. Buradaki
örnekte landing handler'ı, `collage.Effect` yerine `collage.LoadTagged` ile
yazılıp `"landing"`'i tag olarak dönebilir. Böylece bir editör page'i
yeniden sıraladığında cache'teki page atılır. Bkz. [Caching](/docs/caching).

## İç içe geçme sınırı

Bir fragment ağacı, root dahil en fazla **32 seviye** derinliğinde olabilir. Daha
derine inen bir page, fragment'lerinin failure policy'leri ne olursa olsun
`ErrMaxDepthExceeded` ile render edilemez. Hata, sınıra ulaşan fragment zincirini
isim isim gösterir.

Gerçek page'ler bu sınırın yakınına bile gelmez. Sınır, register işleminin önceden
engelleyemediği tek bir durum için vardır: döndüğü bir fragment'in, aşağıda bir
yerde yine kendisine resolve olduğu bir resolver. Bağlanmış bir döngü zaten register
sırasında `ErrFragmentCycle` olarak reddedilir.

## Kendi URL'si olan fragment'ler

Bir fragment, etrafındaki page olmadan da getirilebilir. Örneğin bir `fetch()` ile
yenilenen arama sonuçları ya da bir form gönderildikten sonra yerine konan bir panel.
Bu, page üzerinde `WithFragmentPath(locale, pattern, fragment)` ile tanımlanır.
Tanımlanmayan hiçbir şeye bu yolla erişilemez. Böyle bir fragment, register sırasında
page ile birlikte kontrol edilir: template'i, builder'ının kaydettiği hatalar ve
validation'ı. Bkz. [Form'lar ve action'lar](/docs/forms-and-actions).
