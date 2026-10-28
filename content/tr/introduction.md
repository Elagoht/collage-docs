---
description: collage neden var, neyin üzerine kurulu, size neyi bırakır ve ne zaman doğru araçtır.
reference: NewPage, NewFragment, PageBuilder.WithFragmentPath, FragmentBuilder.WithGuard, App.InvalidateTags, App.PurgeTags, App.Use, App.Handle
---

# Giriş

collage, sunucuda render edilen web siteleri ve web uygulamaları için yazılmış bir
Go framework'üdür. Page'leri, fragment denen küçük parçaları saran bir layout
olarak yazarsınız. Her fragment kendi verisini çeker ve kendi template'ini render
eder. collage page'i bir araya getirir, cache'ler ve içeriğiniz değiştiğinde neyi
atması gerektiğini tam olarak bilir. Page'in bir parçasının tarayıcıda değişmesi
gerektiğinde sunucu o parçayı yeniden render edip gönderir. Başka hiçbir şey
göndermez.

collage tek bir Go programıdır. Aynı binary sitenizi sunar, form'larınıza ve API
çağrılarınıza cevap verir. İsterseniz sitenin tamamını static dosyalar olarak da
yazar. Client framework'ü yoktur, bundler yoktur, standart kütüphane dışında hiçbir
bağımlılığı da yoktur.

## Neden collage

Çoğu web sitesi bir uygulama değildir. Bir landing page, bir ürün sayfası, bir
dokümantasyon sitesi, sayfalarından birinde form olan birkaç sayfalık bir site:
bunlar tıklandığından çok daha fazla okunur. Yine de böyle bir siteyi kurmanın
alışılmış yolu sunucuda bir JavaScript runtime'ı, bir client framework'ü, bir
bundler ve bir hydration adımıyla başlar. Bu da kurulması ve ayakta tutulması
gereken bir runtime, deploy edilmesi gereken bir bağımlılık ağacı ve sunucunun
gönderebileceği HTML'i oluşturabilsin diye her tarayıcıya gönderilen bir program
demektir. Böyle bir runtime Go'dan yavaştır ve aynı page'ler için sunucuda daha
fazla bellek ve disk tüketir.

Bu işin çoğunu tarayıcı zaten yapar. Page'i render eden de form'ları gönderen de
HTML'dir. Kalanını birkaç satır sade JavaScript halleder. Sunucuda bir şey
değiştiğinde de bütün projeyi bir single-page app'e çevirmeye gerek yoktur:
`fetch` ve `XMLHttpRequest`, server-sent events ve WebSocket'ler, değişen parçayı
getirip yerine koymak için zaten vardır.

collage bunun üzerine kuruludur. Sunucu state'i tutar ve onu fragment fragment
HTML olarak render eder. Tarayıcı da ihtiyaç duyduğu fragment'i ister.

## Component değil, fragment

Çoğu page, birbirinden habersiz parçalardan oluşur. Bir blog yazısında yazının
kendisi, bir yazar kartı, ilgili yazıların listesi ve bir navigasyon çubuğu bulunur.
Her parça farklı bir veriye ihtiyaç duyar. Bu veriyi farklı bir yerden alır ve farklı
nedenlerle değişir.

collage'da bu parçaların her biri bir **fragment**'tir. Bir fragment; bir template,
template'in ihtiyaç duyduğu veriyi çeken bir fonksiyon ve başka fragment'lerin
yerleştiği isimli **slot**'lardan oluşur. Bir **page** ise içinde bir content
fragment'i bulunan bir layout fragment'i ve bir URL'dir.

```go
author := collage.NewFragment("author", "fragments/author.html").
	WithData(collage.LoadTagged(loadAuthor)).
	Build()

post := collage.NewFragment("post", "pages/post.html").
	WithData(collage.LoadTagged(loadPost)).
	WithSlotFragment("author", author).
	Build()

page := collage.NewPage("post").
	WithLayouts(layout).
	WithContent(post).
	WithPath("en", "/blog/{slug}").
	Incremental(10 * time.Minute).
	Build()
```

Page'leri bu şekilde kurmanın üç sonucu vardır.

- **Veri eşzamanlı çekilir, HTML sırayla yazılır.** Bir fragment render edilmeden
  önce slot'larındaki her fragment'in data handler'ı aynı anda başlar. Birbirinden
  bağımsız parçalardan oluşan bir page, upstream çağrılarını birlikte yapar. HTML
  ise yine de her seferinde aynı şekilde çıkar.
- **Bir page neyden oluştuğunu bilir.** Her data handler hangi içerik parçalarını
  kullandığını söyler (`post:hello-world`, `author:ada`) ve cache'lenen page bu
  tag'leri taşır. Bir yazar değiştiğinde tek bir çağrı onu gösteren her page'i
  invalidate eder, başka hiçbir şeye dokunmaz.
- **Bir parça, page'i de beraberinde götürmeden başarısız olabilir.** Bir fragment
  zorunlu olabilir, bir fallback'i olabilir ya da hiçbir şey render etmeyebilir.
  Bozuk bir sidebar 500 hatası değil, eksik bir sidebar demektir.

Bir fragment kendi URL'sinde de açılabilir. Bir page'in bir parçası client
framework'ü olmadan böyle yenilenir: sunucu fragment'i render eder, üç satır
JavaScript de onu yerine koyar.

```go
collage.NewPage("search").
	WithLayouts(layout).
	WithContent(searchPage).
	WithPath("en", "/search").
	WithFragmentPath("en", "/search/results", results).
	Build()
```

```js
input.addEventListener("input", async () => {
  const response = await fetch("/search/results?q=" + encodeURIComponent(input.value));
  document.querySelector("#results").outerHTML = await response.text();
});
```

Yalnızca tanımlanan şeylere erişilebilir ve bir fragment path, page'iyle aynı
guard'lardan geçer. Bir action da bir form'a yalnızca değişen fragment'le cevap
verebilir. Ayrıntılar
[Kendi URL'sinde bir fragment](/docs/forms-and-actions#a-fragment-at-its-own-url)
bölümünde anlatılır. Belirli aralıklarla ya da sunucu bir değişikliği push
ettiğinde yenilenen fragment'ler için [elagoht/live](/docs/plugins#elagohtlive)
vardır.

## Bir page neyden oluştuğunu bilir

Cache'lenen her page ve `collage.Cached` ile cache'lenen her değer, oluşturulduğu
içeriğin tag'lerini taşır. O içerik değiştiğinde adını söylemeniz yeterlidir:

```go
// An author edited their profile: every page that shows them renders again.
app.InvalidateTags(ctx, "author:ada")

// An account was deleted: nothing built from it may be served again.
app.PurgeTags(ctx, "user:42")
```

`InvalidateTags`, bir tag'i kullanan page'leri stale olarak işaretler ve arka planda
yeniden render eder. Yeni render başarısız olursa cevap vermek için eski page'i
saklar. `PurgeTags` ise onları diskten de tamamen siler. Bu, bir kez daha bile
gösterilmemesi gereken içerik içindir. Tag'i kullanmayan page'lere dokunulmaz.
Render stratejileri, data cache, warming ve stale page'lerle birlikte konunun
tamamı [Caching](/docs/caching#invalidating) sayfasında anlatılır.

## İstediğiniz kadar middleware ve plugin

Size tek bir middleware dosyası veren bir framework'te, bir request'in ihtiyaç
duyduğu her şeyi (authentication, redirect'ler, header'lar, loglama) onun içinde
elle zincirlersiniz. collage'da `app.Use` herhangi
bir standart `func(http.Handler) http.Handler` alır ve istediğiniz kadar
çağrılabilir. İlk register edilen en dıştakidir:

```go
app.Use(requestID)
app.Use(auditLog)
app.Use(loadUser)
```

Plugin'ler kendi middleware'lerini, template fonksiyonlarını, page'lerini,
mount'larını ve render ile cache'e bağlanan hook'larını getirir ve listelediğiniz
sırayla çalışır. Çekirdek bilerek küçük tutulur. Bir sitenin özel olarak ihtiyaç
duyduğu şey bir plugin'dir:

- **Form'lar ve state:** validation, spam'e karşı honeypot, flash mesajları,
  session'lar, upload'lar.
- **Güvenlik:** güvenlik header'ları, rate limit, basic auth, OAuth, fail2ban, bir
  consent banner'ı.
- **Canlı güncelleme:** server-sent events ya da WebSocket üzerinden yenilenen
  fragment'ler ve bir fragment'i yerine koyan küçük bir script.
- **SEO ve içerik:** meta tag'ler, structured data, sitemap'ler, feed'ler, Open Graph
  görselleri, Markdown, syntax highlighting, arama, çeviriler.
- **Dağıtım ve operasyon:** küçültme, sıkıştırma, görsel optimizasyonu, deploy
  hedefleri, metrikler, tracing, health check'ler, zamanlanmış job'lar ve
  kuyruklar.

Her biri import ettiğiniz bir Go modülüdür. Ne bir loader ne de bir registry
vardır. Kendi plugin'iniz, ihtiyaç duyduğu hook'ları uygulayan bir Go tipidir.
Bkz. [Plugin kullanmak](/docs/plugins), [Plugin yazmak](/docs/writing-plugins) ve
[Middleware ve kendi API'niz](/docs/middleware-and-apis).

## Hâlâ Go yazıyorsunuz

İkinci bir dil de ikinci bir process de yoktur. Veri katmanınız Go kodunuzdur
(veritabanı, kuyruk, dış API) ve data handler'lardan ve action'lardan herhangi bir
fonksiyon gibi çağrılır. Elinizdeki bir API ya da servis, collage'ın kendi
mux'ına, page'lerin yanına mount edilir:

```go
api := http.NewServeMux()
api.HandleFunc("GET /api/users/{id}", users.Get)
api.HandleFunc("POST /api/users", users.Create)

if err := app.Handle("/api/", api); err != nil {
	return err
}
```

Herhangi bir `http.Handler` işinizi görür: bir `ServeMux`, halihazırda kullandığınız
bir router, bir reverse proxy. Bir şeyi değiştirdiğinde cache'lenen page'leri invalidate
edebilir ve siteyle aynı binary'de yer alır. Kendi middleware'ini kendisi getirir:
sitenin middleware'leri onun için yalnızca `collage.WithMiddleware()` ile mount
edildiğinde çalışır. Bkz.
[Kendi handler'ınız](/docs/middleware-and-apis#your-own-handler-apphandle).

## Siteler ve web uygulamaları

collage yalnızca okunan page'ler için değildir. Sunucuda render edilen bir front
end'in ihtiyaç duyduğu her şey buradadır. Bu yüzden çok kullanıcılı, rol bazlı bir
web uygulaması da bir landing page kadar collage'a uygundur:
[session'lar](/docs/plugins#elagohtsession), sahtesi yapılamayan form'lar, spam'e
karşı bir [honeypot](/docs/plugins#elagohthoneypot), page'in canlı güncellenen
parçaları ve bir page'i kimin görebileceğine o render edilmeden önce karar veren
guard'lar.

```go
func requireRole(role string) collage.GuardFunc {
	return func(ctx context.Context, r *http.Request) (*collage.GuardDecision, error) {
		if users.HasRole(ctx, role) {
			return nil, nil // allowed
		}
		return &collage.GuardDecision{Status: http.StatusForbidden}, nil
	}
}

admin := collage.NewFragment("admin", "layouts/admin.html").
	WithGuard(requireRole("admin")).
	Build()
```

`admin` layout'una sarılan her page, her locale'de ve ona ulaşan her URL'de,
fragment path'leri de dahil, yöneticiler içindir. Bkz.
[Private page'ler](/docs/pages-and-layouts#private-pages-guards) ve
[Form'lar ve action'lar](/docs/forms-and-actions).

### Dışarıda bıraktıkları

collage batteries-included değildir. ORM'i, migration'ları ve veritabanı katmanı
yoktur. Session'lar, çeviriler ve bir JavaScript toolchain'i de çekirdekte değildir.
Go yazıyorsunuz ve Go'nun bunların hepsine iyi cevapları zaten var. Bunlardan birini
sizin yerinize seçen bir framework, etrafından dolaşmanız gereken bir şey daha
olurdu. collage'ın üstlendiği kısım web front end'ine özgü olandır: page'leri bir
araya getirmek, cache'lemek, ne zaman stale olduklarını bilmek ve üzerlerindeki
form'lara cevap vermek. Session'lar ve çeviriler plugin'dir.

### Ne zaman doğru araç değildir

collage, state'i tarayıcıda yaşayan ekranlar için uygun değildir: zengin bir metin
editörü, bir çizim canvas'ı, çevrimdışı çalışması gereken bir uygulama. Bunlar için
bir client framework'ü kullanın. İsterseniz etraflarındaki page'ler için de
[`app.Handle`](/docs/middleware-and-apis#your-own-handler-apphandle) üzerinden
collage'ı kullanın.

## Bir request nasıl ilerler

<figure class="diagram-figure" data-search-ignore>
<svg class="diagram" viewBox="0 0 320 400" role="img" aria-labelledby="flow-title">
<title id="flow-title">Bir request plugin middleware'lerinden ve app.Use'dan geçer, routing de onu bir page'e, bir fragment path'e, bir action'a ya da kendi handler'ınıza gönderir. Yalnızca page tag'lerle cache'lenir; fragment path ve action her seferinde yeniden render edilir. Hepsi veri katmanınızdan okur. Kendi handler'ınıza gelen bir request, handler WithMiddleware ile mount edilmedikçe middleware'leri atlar.</title>
<defs><marker id="flow-arrow" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto"><path class="d-head" d="M0 0 8 4 0 8z" /></marker></defs>
<rect class="d-pill" x="68" y="8" width="100" height="26" rx="13" /><text x="118" y="25" text-anchor="middle">tarayıcı</text>
<line class="d-flow" x1="118" y1="34" x2="118" y2="54" marker-end="url(#flow-arrow)" />
<path class="d-flow" d="M118 42H270V54" marker-end="url(#flow-arrow)" />
<rect class="d-frame" x="6" y="58" width="224" height="28" rx="6" /><text x="118" y="76" text-anchor="middle">plugin middleware'leri</text>
<rect class="d-frame" x="6" y="92" width="224" height="28" rx="6" /><text x="118" y="110" text-anchor="middle">app.Use</text>
<rect class="d-frame" x="234" y="58" width="72" height="62" rx="5" /><text class="d-small" x="270" y="86" text-anchor="middle">app.Handle</text><text class="d-small" x="270" y="99" text-anchor="middle">düz</text>
<line class="d-flow" x1="118" y1="120" x2="118" y2="140" marker-end="url(#flow-arrow)" />
<rect class="d-cmd" x="68" y="144" width="100" height="26" rx="13" /><text class="d-cmd-text" x="118" y="161" text-anchor="middle">routing</text>
<path class="d-flow" d="M98 170C98 185 42 182 42 198" marker-end="url(#flow-arrow)" />
<line class="d-flow" x1="118" y1="170" x2="118" y2="198" marker-end="url(#flow-arrow)" />
<path class="d-flow" d="M138 170C138 185 194 182 194 198" marker-end="url(#flow-arrow)" />
<rect class="d-a" x="6" y="202" width="72" height="34" rx="5" /><text class="d-on-a" x="42" y="223" text-anchor="middle">page</text>
<rect class="d-b" x="82" y="202" width="72" height="34" rx="5" /><text class="d-on-b d-small" x="118" y="216" text-anchor="middle">fragment</text><text class="d-on-b d-small" x="118" y="228" text-anchor="middle">path</text>
<rect class="d-c" x="158" y="202" width="72" height="34" rx="5" /><text class="d-on-c" x="194" y="223" text-anchor="middle">action</text>
<line class="d-flow" x1="42" y1="236" x2="42" y2="250" marker-end="url(#flow-arrow)" />
<rect class="d-chip-hit" x="6" y="254" width="72" height="24" rx="12" /><text class="d-hit-text d-small" x="42" y="270" text-anchor="middle">cache · tag</text>
<line class="d-flow" x1="42" y1="278" x2="42" y2="294" marker-end="url(#flow-arrow)" />
<line class="d-flow" x1="118" y1="236" x2="118" y2="294" marker-end="url(#flow-arrow)" />
<line class="d-flow" x1="194" y1="236" x2="194" y2="294" marker-end="url(#flow-arrow)" />
<rect class="d-soft-a" x="6" y="298" width="224" height="30" rx="6" /><text x="118" y="317" text-anchor="middle">render: fragment'ler aynı anda çeker</text>
<line class="d-flow" x1="118" y1="328" x2="118" y2="346" marker-end="url(#flow-arrow)" />
<line class="d-flow" x1="270" y1="120" x2="270" y2="346" marker-end="url(#flow-arrow)" />
<rect class="d-frame" x="6" y="350" width="300" height="44" rx="6" /><text x="156" y="369" text-anchor="middle">Go ile yazılmış veri katmanınız</text><text class="d-small" x="156" y="384" text-anchor="middle">veritabanı · API · CMS · kuyruk</text>
</svg>
<figcaption>Her request önce, register edildikleri sırayla middleware'lerden geçer; kendi handler'larınızdan birine gelen bir request ise o handler WithMiddleware ile mount edilmedikçe bunları atlar. Ardından routing onu bir page'e, bir page'in fragment path'lerinden birine, bir action'a ya da kendi handler'larınızdan birine verir. Bir page, oluşturulduğu içeriğin tag'leriyle cache'lenir; fragment path ve action ise her seferinde yeniden render edilir. Hepsi veri katmanınızdan okur, kendi handler'ınız da onunla doğrudan konuşur.</figcaption>
</figure>

## Neler sunar

- URL başına cache'lenen page'ler ve üç strateji: bir kez render edilip siz
  invalidate edene kadar saklanan, bir TTL sonunda yeniden render edilen ya da hiç
  cache'lenmeyen. Hiçbir şey çekmeyen bir page, söylenmesine gerek kalmadan bir kez
  render edilir.
- Page'ler arasında cache'lenen veri. Cache açıkken aynı yazarın otuz yazısı, yazarı
  yalnızca bir kez çeker.
- Etkilenen page'leri arka planda yeniden render eden tag tabanlı invalidation ve
  onları tamamen silen purge.
- Bir page'in bir parçasını birkaç satır JavaScript ile yenilemek için kendi
  URL'lerindeki fragment'ler.
- Kendi page'ine post eden ve built-in request forgery korumasıyla gelen
  form'lar. Cache'lenmiş bir page de yine böyle bir form taşıyabilir.
- Hiçbir şey render edilmeden önce bir page'i kimin görebileceğine karar veren
  guard'lar.
- Page adlarından oluşturulan link'ler. Böylece bir page'in path'i değiştiğinde,
  sitenin her locale'inde link'ler onu takip eder.
- Title'lar, meta tag'ler, stylesheet'ler gibi head elemanları. Bunları, page'in
  neresinde olursa olsun ihtiyaç duyan fragment tanımlar.
- Sitemap'ler, feed'ler ve `robots.txt` için HTML olmayan route'lar, `app.Use` ile
  middleware ve geri kalan her şey için `app.Handle` ile kendi `http.Handler`'ınız.
- Siteyi herhangi bir static host için dosyalara çeviren bir static export. Şu an
  okuduğunuz sayfalar da onunla üretildi.
- Her Go değişikliğinde yeniden build eden ve her template değişikliğinde tarayıcıyı
  yeniden yükleyen bir development sunucusu. Render'ı başarısız olan bir page,
  hatasını tarayıcıda gösterir ve en başta template, satır ve sebep yer alır.
  Derlenemeyen Go kodu, derleyicinin hatasını tarayıcıda gösterir. Hiç başlayamayan bir program ise reddedilen bir bağlantı
  yerine yazdırdığı çıktıyı tarayıcıda gösterir.

## Sırada ne var

[Kurulum](/docs/installation) bir projeyi bir dakikada çalışır hale getirir.
[İlk page'iniz](/docs/your-first-page) sıfırdan bir page kurar.
[Page'ler ve layout'lar](/docs/pages-and-layouts) ise kavramların başladığı yerdir.
