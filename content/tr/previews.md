---
description: Editörlere yayımlanmamış taslakları gerçek sitede gösterin. Cache editöre yayımlanmış page'i, başka hiç kimseye de taslağı sunmaz.
reference: SkipCache, Cached, NewAction, ErrVaryTooLate
---

# Preview'lar

Bir CMS'te "preview" düğmesine basan editör, taslağı gerçek sitede ve etrafındaki
gerçek layout'la görmek ister. Cache burada iki kez engel çıkarır. Editöre, cache'te
zaten duran yayımlanmış page sunulmamalıdır. Ona gösterilen taslak da asla
cache'lenmemelidir. Aksi hâlde diğer bütün okuyuculara giden page o taslak olur.

collage bunun için tek bir primitive sunar: `collage.SkipCache`. Gerisini size bırakır.
Kimin preview yapabileceği sizin kullanıcılarınızla ilgili bir sorudur ve collage'ın
session'lar konusunda bir görüşü yoktur.

## `collage.SkipCache`

```go
func SkipCache(r *http.Request) error
```

Middleware'den çağrıldığında tek bir request'i preview olarak işaretler. Bu request'te
şunlar olur:

- **Taze bir render.** Page cache okunmaz. Bu yüzden yayımlanmış page'in cache'teki
  kopyası asla sunulmaz.
- **Hiçbir şey saklanmaz.** Render sonucu page cache'e yazılmaz.
- **`Cache-Control: private, no-store`**. Böylece onu ne tarayıcı cache'i, ne bir
  proxy ne de bir CDN saklar.
- **Taze veri.** O render'daki her `collage.Cached` çağrısı, saklanmış bir değer
  dönmek yerine veriyi yeniden çeker ve hiçbir şey saklamaz. Taslak page'i,
  yayımlanmış verinin cache'teki kopyasıyla birlikte gören bir editör aslında
  taslağı görmüş olmaz.

Document'lar da buna uyar. Preview sırasında istenen bir document ne cache'ten okunur
ne de cache'e yazılır.

Diğer request'ler bundan etkilenmez. Onlar her zamanki gibi cache'ten sunulur ve
cache yayımlanmış sürümü tutmaya devam eder.

Bir preview'ın render'ı da yalnızca ona aittir. Cache'lenen bir page'in bir okuyucu
için yapılan render'ı, o page'i aynı anda isteyen bütün okuyucularla paylaşılır. Bu
yüzden ona okuyucunun cookie'leri de, middleware'inizin request context'ine koyduğu
değerler de verilmez (bkz.
[paylaşılan bir render neyi görür](/docs/caching#what-a-shared-render-sees)).
Preview ise hiçbir zaman paylaşılmaz: editörün kendi request'i ve context'iyle render
edilir. Middleware'inizin context'e koyduğu bir işaret, data handler'a ve
`collage.Cached` içindeki fetch'e bu sayede ulaşır.

Bir preview'ın response'unda `X-Collage-Cache` header'ı da bulunmaz. Bu header
yalnızca page cache'ten geçen bir response'a konur. Preview ise page cache'ten hiç
geçmez.

<figure class="diagram-figure" data-search-ignore>
<svg class="diagram" viewBox="0 0 320 384" role="img" aria-labelledby="preview-flow-title">
<title id="preview-flow-title">Cache'lenen aynı page için iki request. Okuyucunun middleware'i hiçbir şey çağırmaz: hit'te page cache cevap verir, miss'te diğer okuyucularla paylaşılan ve cookie ya da context değeri görmeyen bir render saklanır ve public bir Cache-Control ile gönderilir. Editörün middleware'i routing'den önce SkipCache'i çağırır ve bir drafts işareti koyar: page cache okunmaz, render editörün kendisine aittir, Cached veriyi taze çeker, hiçbir şey saklanmaz ve taslak private, no-store olarak gönderilir.</title>
<defs><marker id="preview-arrow" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto"><path class="d-head" d="M0 0 8 4 0 8z" /></marker><marker id="preview-hit" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto"><path class="d-head-hit" d="M0 0 8 4 0 8z" /></marker></defs>
<rect class="d-a" x="10" y="8" width="140" height="24" rx="5" /><text class="d-on-a" x="80" y="24" text-anchor="middle">okuyucu</text>
<rect class="d-b" x="170" y="8" width="140" height="24" rx="5" /><text class="d-on-b" x="240" y="24" text-anchor="middle">editör + cookie</text>
<line class="d-flow" x1="80" y1="32" x2="80" y2="46" marker-end="url(#preview-arrow)" />
<line class="d-flow" x1="240" y1="32" x2="240" y2="46" marker-end="url(#preview-arrow)" />
<rect class="d-frame" x="10" y="50" width="140" height="40" rx="6" /><text x="80" y="66" text-anchor="middle">middleware</text><text class="d-small" x="80" y="81" text-anchor="middle">geçerli cookie yok</text>
<rect class="d-cmd" x="170" y="50" width="140" height="40" rx="6" /><text class="d-cmd-text" x="240" y="66" text-anchor="middle">SkipCache(r)</text><text class="d-small" x="240" y="81" text-anchor="middle">+ ctx'te drafts</text>
<line class="d-flow" x1="80" y1="90" x2="80" y2="104" marker-end="url(#preview-arrow)" />
<line class="d-flow" x1="240" y1="90" x2="240" y2="104" marker-end="url(#preview-arrow)" />
<rect class="d-frame" x="10" y="108" width="300" height="24" rx="6" /><text class="d-small" x="160" y="124" text-anchor="middle">routing: bundan sonra SkipCache geç kalır</text>
<line class="d-flow" x1="80" y1="132" x2="80" y2="146" marker-end="url(#preview-arrow)" />
<line class="d-flow" x1="240" y1="132" x2="240" y2="146" marker-end="url(#preview-arrow)" />
<rect class="d-frame" x="10" y="150" width="140" height="40" rx="6" /><text x="80" y="166" text-anchor="middle">page cache</text><text class="d-small" x="80" y="181" text-anchor="middle">HIT: olduğu gibi</text>
<g class="d-dropped"><rect class="d-frame" x="170" y="150" width="140" height="40" rx="6" /><text x="240" y="166" text-anchor="middle">page cache</text><text class="d-small" x="240" y="181" text-anchor="middle">okunmaz</text></g>
<line class="d-flow" x1="80" y1="190" x2="80" y2="204" marker-end="url(#preview-arrow)" /><text class="d-small" x="88" y="201">miss</text>
<line class="d-flow" x1="240" y1="190" x2="240" y2="204" marker-end="url(#preview-arrow)" />
<rect class="d-frame" x="10" y="208" width="140" height="52" rx="6" /><text x="80" y="224" text-anchor="middle">paylaşılan render</text><text class="d-small" x="80" y="238" text-anchor="middle">cookie, ctx gizli</text><text class="d-small" x="80" y="251" text-anchor="middle">Cached: saklananı okur</text>
<rect class="d-frame" x="170" y="208" width="140" height="52" rx="6" /><text x="240" y="224" text-anchor="middle">kendi render'ı</text><text class="d-small" x="240" y="238" text-anchor="middle">request'i ve ctx'i</text><text class="d-small" x="240" y="251" text-anchor="middle">Cached: taze çeker</text>
<line class="d-flow" x1="80" y1="260" x2="80" y2="274" marker-end="url(#preview-arrow)" />
<line class="d-flow" x1="240" y1="260" x2="240" y2="274" marker-end="url(#preview-arrow)" />
<rect class="d-frame" x="10" y="278" width="140" height="40" rx="6" /><text x="80" y="294" text-anchor="middle">saklanır</text><text class="d-small" x="80" y="309" text-anchor="middle">sonraki okuyucu için</text>
<g class="d-dropped"><rect class="d-frame" x="170" y="278" width="140" height="40" rx="6" /><text x="240" y="294" text-anchor="middle">hiçbir şey</text><text class="d-small" x="240" y="309" text-anchor="middle">saklanmaz: page, data</text></g>
<line class="d-flow" x1="80" y1="318" x2="80" y2="332" marker-end="url(#preview-arrow)" />
<line class="d-flow" x1="240" y1="318" x2="240" y2="332" marker-end="url(#preview-arrow)" />
<path class="d-hitline" d="M10 170 H3 V356 H6" marker-end="url(#preview-hit)" />
<rect class="d-soft-a" x="10" y="336" width="140" height="40" rx="6" /><text x="80" y="352" text-anchor="middle">yayımlanmış page</text><text class="d-small" x="80" y="367" text-anchor="middle">public Cache-Control</text>
<rect class="d-soft-a" x="170" y="336" width="140" height="40" rx="6" /><text x="240" y="352" text-anchor="middle">taslak page</text><text class="d-small" x="240" y="367" text-anchor="middle">private, no-store</text>
</svg>
<figcaption>Cache'lenen aynı blog yazısı için iki request. Okuyucununkini hiçbir şey işaretlemez. Bu yüzden ona page cache cevap verir. Miss'te ise diğer okuyucularla paylaşılan (ve onların cookie'lerini ya da context değerlerini görmeyen) bir render saklanır ve public bir <code>Cache-Control</code> ile gönderilir. Editörün middleware'i routing'den önce <code>collage.SkipCache</code>'i çağırmıştır: bu request page cache'i atlar, editörün request'i ve context'iyle kendi başına render edilir, <code>collage.Cached</code>'in veriyi taze çekmesini sağlar, hiçbir şey saklamaz ve <code>private, no-store</code> olarak gönderilir.</figcaption>
</figure>

`collage.Vary` gibi bu fonksiyon da routing'den önce çağrılmalıdır. Yani onu `app.Use`
ile register edilmiş bir middleware'den çağırmanız gerekir. Routing'den sonra, örneğin bir
data handler'dan çağrılırsa her route'ta `collage.ErrVaryTooLate` döner (v0.11.0'dan
beri). collage'ın sunmadığı bir request üzerinde çağrılırsa
`collage.ErrVaryOutsideRequest` döner.

## Eksiksiz bir preview akışı

Genelde kullanılan düzen şudur. Aşağıdaki örnek de bu düzeni izler:

1. CMS'in preview düğmesi `/api/preview?secret=…&slug=…` adresini yeni bir sekmede
   açar.
2. Bu URL'deki bir [action](/docs/forms-and-actions) secret'ı kontrol eder, imzalı bir
   cookie set eder ve yazıya redirect eder.
3. Middleware sonraki her request'te cookie'yi görür, `collage.SkipCache`'i çağırır ve
   data handler'lara CMS'ten taslakları istemelerini söyler.
4. Editörün işi bittiğinde başka bir action cookie'yi siler.

Burada iki secret kullanılır ve ikisi de ortam değişkenlerinden okunur.
`PREVIEW_SECRET` CMS ile paylaşılır, böylece CMS bir preview başlatabilir.
`PREVIEW_KEY` ise yalnızca sunucunuzun bildiği anahtardır ve cookie'yi imzalar.

### Cookie'yi imzalamak

Cookie, ne zaman sona ereceğini ve bu değer üzerinden hesaplanmış bir imzayı tutar.
Anahtara sahip olmayan hiç kimse geçerli bir cookie üretemez. Süresi dolmuş bir cookie
ise tarayıcı onu hâlâ gönderse bile reddedilir.

```go
package preview

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"strconv"
	"strings"
	"time"
)

// CookieName is the cookie a preview is carried in.
const CookieName = "preview"

// key signs preview cookies. At least 32 random bytes; empty turns previews off.
var key = []byte(os.Getenv("PREVIEW_KEY"))

// sign returns a cookie value that is valid until expires.
func sign(expires time.Time) string {
	stamp := strconv.FormatInt(expires.Unix(), 10)
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(stamp))
	return stamp + "." + hex.EncodeToString(mac.Sum(nil))
}

// Valid reports whether value is a preview cookie this server signed and that has
// not expired.
func Valid(value string) bool {
	if len(key) == 0 {
		return false
	}
	stamp, _, ok := strings.Cut(value, ".")
	if !ok {
		return false
	}
	unix, err := strconv.ParseInt(stamp, 10, 64)
	if err != nil {
		return false
	}
	expires := time.Unix(unix, 0)
	return time.Now().Before(expires) && hmac.Equal([]byte(value), []byte(sign(expires)))
}
```

### Preview'ı başlatan action

```go
// Start is the URL the CMS's preview button opens:
// GET /api/preview?secret=...&slug=...
func Start(app *collage.App) *collage.Action {
	return collage.NewAction("preview-start").
		WithPath("en", "/api/preview").
		WithMethods(http.MethodGet).
		WithHandler(func(_ context.Context, rc *collage.RenderContext) (*collage.ActionResult, error) {
			query := rc.Request.URL.Query()
			secret := os.Getenv("PREVIEW_SECRET")
			if len(key) == 0 || secret == "" ||
				!hmac.Equal([]byte(query.Get("secret")), []byte(secret)) {
				return collage.Status(http.StatusUnauthorized), nil
			}

			// Built from the page's name, so a slug cannot send the editor anywhere
			// but a blog post.
			target, err := app.URL("blog-post", "", map[string]string{"slug": query.Get("slug")})
			if err != nil {
				return collage.Status(http.StatusBadRequest), nil
			}

			expires := time.Now().Add(time.Hour)
			cookie := &http.Cookie{
				Name:     CookieName,
				Value:    sign(expires),
				Path:     "/",
				Expires:  expires,
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteLaxMode,
			}

			result := collage.SeeOther(target)
			result.Header = http.Header{}
			result.Header.Add("Set-Cookie", cookie.String())
			return result, nil
		}).
		Build()
}
```

Bir `GET` action'ı forgery token gerektirmez, çünkü yalnızca güvenli olmayan
method'lar kontrol edilir. CMS'in bu action'ı düz bir link olarak açabilmesini
sağlayan da budur. `ActionResult.Header`, redirect'ten önce response'a yazılır. Cookie
de tam olarak oraya konur.

`Secure: true` production için doğru ayardır. Chrome ve Firefox secure bir cookie'yi
`http://localhost` üzerinde de kabul eder. Bu yüzden `collage dev` orada bu ayarla,
hiçbir değişiklik yapmadan çalışır. Safari ise kabul etmeyebilir ve düz `http`
üzerinde cookie'yi yok sayar. Safari'de geliştirme yapıyorsanız `Secure` değerini
request'in TLS üzerinden gelip gelmediğine göre belirleyin ya da preview'ı başka bir
tarayıcıda açın.

### Preview'ı bitiren action

```go
// Exit clears the preview cookie: GET /api/preview/exit
func Exit() *collage.Action {
	return collage.NewAction("preview-exit").
		WithPath("en", "/api/preview/exit").
		WithMethods(http.MethodGet).
		WithHandler(func(context.Context, *collage.RenderContext) (*collage.ActionResult, error) {
			cookie := &http.Cookie{Name: CookieName, Path: "/", MaxAge: -1}
			result := collage.SeeOther("/")
			result.Header = http.Header{}
			result.Header.Add("Set-Cookie", cookie.String())
			return result, nil
		}).
		Build()
}
```

### Preview'ı uygulayan middleware

```go
type draftsKey struct{}

// Drafts reports whether this request is a preview, and data handlers should
// fetch unpublished content.
func Drafts(ctx context.Context) bool {
	drafts, _ := ctx.Value(draftsKey{}).(bool)
	return drafts
}

// Middleware turns a request carrying a valid preview cookie into a preview.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if cookie, err := r.Cookie(CookieName); err == nil && Valid(cookie.Value) {
			if err := collage.SkipCache(r); err == nil {
				r = r.WithContext(context.WithValue(r.Context(), draftsKey{}, true))
			}
		}
		next.ServeHTTP(w, r)
	})
}
```

Context değeri yalnızca `SkipCache` başarılı olduğunda set edilir. Bütün bu düzen tek
bir sonucu önlemek için vardır: hâlâ cache'lenebilen bir request'e taslak render
edilmesi. Bu yüzden ikisi ya birlikte gerçekleşir ya da hiç gerçekleşmez.

### Hepsini bağlamak

```go
if err := app.Use(preview.Middleware); err != nil {
	return err
}
for _, action := range []*collage.Action{preview.Start(app), preview.Exit()} {
	if err := app.RegisterAction(action); err != nil {
		return fmt.Errorf("register action %q: %w", action.Name, err)
	}
}
```

### Taslakları çeken data handler'lar

Bir data handler, request bir preview ise taslakları ister:

```go
var postKey = collage.NewKey[cms.Post]("post")

func postData(ctx context.Context, rc *collage.RenderContext) (postView, []string, error) {
	slug := rc.Param("slug")
	tags := []string{"post:" + slug}

	post, err := collage.Cached(rc, postKey.With(slug), time.Hour, tags,
		func(ctx context.Context) (cms.Post, error) {
			return cmsClient.Post(ctx, slug, cms.Options{Drafts: preview.Drafts(ctx)})
		})
	if errors.Is(err, cms.ErrNotFound) {
		return postView{}, tags, fmt.Errorf("post %q: %w", slug, collage.ErrNotFound)
	}
	if err != nil {
		return postView{}, tags, err
	}
	return postView{Post: post, Preview: preview.Drafts(ctx)}, tags, nil
}
```

`collage.Cached` çağrısı için özel bir durum yazmanız gerekmez. Sıradan bir okuyucu
için cache'te saklanan yayımlanmış yazıyı döner. Preview'da ise veriyi taslaklarla
birlikte yeniden çeker ve hiçbir şey saklamaz. Böylece taslak, bir sonraki okuyucunun
alacağı değere sızamaz.

View'daki `Preview` alanı, template'in `/api/preview/exit` linkini içeren bir banner
göstermesini sağlar. Bu, tam da page hiçbir zaman cache'lenmediği için güvenlidir.
Banner hiçbir okuyucunun karşısına çıkamaz.

## Örnek: bir editör taslak bir yazıyı preview ediyor

Bir blog yazılarını bir CMS'te tutar ve her yazının `/blog/{slug}` adresinde bir
page'i vardır. `hello` yazısı yayımlanmıştır ve okuyuculara page cache'ten sunulur.
Editör Ada, CMS'te yazının girişini yeniden yazar ve yayımlamadan önce sonucu sitede
görmek ister.

Site yukarıdaki gibi bağlanmıştır: preview middleware'i, iki action ve `postData`.
Yazının page'i bir saat cache'lenir:

```go
postPage := collage.NewPage("blog-post").
	WithLayouts(layout).
	WithContent(collage.NewFragment("post", "pages/post.html").
		WithData(collage.LoadTagged(postData)).
		Required(). // an unknown slug is a 404
		Build()).
	WithPath("en", "/blog/{slug}").
	Incremental(time.Hour).
	Build()
```

Template'i banner'ı yalnızca bir preview'da gösterir:

```html
{{if .Preview}}
<p class="preview-banner">Draft preview. <a href="/api/preview/exit">Leave preview</a></p>
{{end}}
<h1>{{.Post.Title}}</h1>
```

Development dışında request'ler sırasıyla şunları alır:

| # | Kim | Request | Aldığı | `Cache-Control` | `X-Collage-Cache` |
| --- | --- | --- | --- | --- | --- |
| 1 | Bir okuyucu | `GET /blog/hello` | Yayımlanmış yazı, render edilip saklanır | `public, max-age=3600` | `MISS` |
| 2 | Başka bir okuyucu | `GET /blog/hello` | Saklanan kopya | `public, max-age=3600` | `HIT` |
| 3 | Ada, CMS'in preview düğmesinden | `GET /api/preview?secret=…&slug=hello` | `/blog/hello`'ya bir `303` ve preview cookie'si | — | — |
| 4 | Ada | `GET /blog/hello` | Banner'lı taslak | `private, no-store` | yok |
| 5 | Ada, bir düzenlemeden sonra | `GET /blog/hello` | Yeni taslak: `collage.Cached` veriyi yeniden çekti | `private, no-store` | yok |
| 6 | Bir okuyucu | `GET /blog/hello` | Hâlâ yayımlanmış kopya | `public, max-age=3600` | `HIT` |
| 7 | Ada, banner'dan | `GET /api/preview/exit` | `/`'ye bir `303` ve silinen cookie | — | — |

Dördüncü ve beşinci satırlar yalnızca Ada için render edilir. Bu render'ların çektiği hiçbir şey
cache'e ulaşmaz. Bu yüzden 6. satırdaki okuyucuya, 2. satırdaki okuyucuyla aynı
saklanmış kopya sunulur.

Ada yazıyı yayımladığında CMS'in webhook'u `postData`'nın döndüğü `post:hello`
tag'ini invalidate eder. Sonraki okuyucu yeni sürümü alır (bkz.
[Middleware ve kendi API'niz](/docs/middleware-and-apis#invalidating-from-your-api)).

## Bilmeniz gerekenler

**Bir CDN yine de önce cevap verebilir.** `SkipCache` collage'ın cache'lerini kontrol
eder ve preview response'unu `no-store` olarak işaretler. Ancak CDN'in kendi
cache'inden cevapladığı bir request sunucunuza hiç ulaşmaz.

- Static page'ler (ya `Static()` olarak tanımlanmış ya da strateji tanımlamayıp
  hiçbir şey çekmeyen page'ler) `max-age=0, must-revalidate` ile gönderilir, bu
  yüzden CDN her seferinde sunucuya sorar.
- Bir `Incremental(ttl)` page ise TTL'i boyunca CDN'den sunulabilir.

CDN'i, preview cookie'sini taşıyan request'lerde kendi cache'ini atlayacak şekilde
yapılandırın.

**CMS'in içinde preview.** CMS preview'ı kendi domain'indeki bir `<iframe>` içinde
gösteriyorsa, `SameSite=Lax` bir cookie iframe'deki siteye gönderilmez.

- Bu durumda `Secure: true` ile birlikte `SameSite: http.SameSiteNoneMode` kullanın.
  Yine de bunun yetmeyebileceğini bilin. O iframe'in içinde cookie'niz third-party
  bir cookie'dir ve Safari ile Firefox bu tür cookie'leri varsayılan olarak engeller.
  Bu yüzden CMS'in iframe'indeki bir preview cookie'yi hiç almayabilir.
- Buna ek olarak `Partitioned: true` ayarlarsanız (CMS'in sitesine göre partition
  edilen bir CHIPS cookie'si), cookie partitioning'i destekleyen tarayıcılarda
  iletilir.
- Her yerde çalışan düzen ise yukarıdakidir: preview'ı yeni bir sekmede açmak.

**Static export hiçbir zaman taslak görmez.** Export bir request olmadan render
edilir. Bu yüzden hiçbir middleware çalışmaz, `preview.Drafts` her page için false
olur ve yalnızca yayımlanmış içerik yazılır. Preview action'ları da export edilmez,
çünkü bir action'ın çalışması için sunucu gerekir. Static dosyalar olarak deploy
edilmiş bir site de sunucuyu editörlerin erişebileceği bir yerde çalıştırarak preview
sunabilir. Ayrıntılar için [Static export](/docs/static-export) sayfasına bakın.

**Invalidation yine önemlidir.** Preview taslağı gösterir. Siteyi değiştiren ise
taslağın yayımlanmasıdır. CMS bir yazıyı yayımladığında, webhook'u o yazının tag'ini
invalidate etmelidir. Böylece okuyucular yeni sürümü alır. Ayrıntılar için
[Middleware ve kendi API'niz](/docs/middleware-and-apis#invalidating-from-your-api)
sayfasına bakın.
