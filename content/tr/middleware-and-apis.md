---
description: app.Use ile standart net/http middleware'leri, collage.Vary ile bir header'a bağlı cache key'leri ve app.Handle ile kendi http.Handler'ınız.
reference: Vary, SkipCache, ErrVaryTooLate, ErrMountShadowsRoute
---

# Middleware ve kendi API'niz

collage page render eder. Bir Go programının HTTP üzerinden yaptığı diğer işleri,
yani API'yi, authentication'ı, dil seçimini ve rate limiting'i her zamanki gibi
kendiniz yazarsınız. Bu kodu iki noktadan birine bağlarsınız:

- Her request'in etrafında çalışan **middleware**'e.
- Bir prefix'in altındaki her request'e cevap veren **kendi handler'ınıza**.

## Middleware: `app.Use`

```go
if err := app.Use(func(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Frame-Options", "DENY")
		next.ServeHTTP(w, r)
	})
}); err != nil {
	return err
}
```

Middleware standart `func(http.Handler) http.Handler` imzasına sahiptir. Bu yüzden
`net/http` için yazılmış her middleware hiç değişmeden çalışır. İlk register edilen
middleware en dıştakidir.

Diğer bütün register işlemleri gibi bu da uygulama başlamadan önce yapılmalıdır.
Uygulama başladıktan sonra `app.Use` `collage.ErrAppStarted` döner.

Middleware'in nerede çalıştığıyla ilgili iki nokta var:

- **Routing'den önce.** Middleware her request'i görür: page'leri, document'ları,
  action'ları, mount edilmiş static dosyaları ve kendi handler'larınızı. 404 ile
  biten request'ler de buna dahildir. Middleware bir request'e 401 ya da redirect
  ile kendisi cevap verebilir. Bu durumda arkasındaki hiçbir şey çalışmaz.
- **Framework'ün koruması içinde.** Middleware, collage'ın span'i, metric'leri ve
  panic recovery'si içinde çalışır. `app.Handler()`'ı kendi middleware'inizle
  sarmaktan farkı budur. Middleware'inizdeki bir panic bağlantıyı koparmaz, normal
  hata yolunda sıradan bir 500'e dönüşür. Middleware'in kendisinin cevap verdiği
  bir request de diğer request'ler gibi sayılır.

<figure class="diagram-figure" data-search-ignore>
<svg class="diagram" viewBox="0 0 320 360" role="img" aria-labelledby="mw-layers-title">
<title id="mw-layers-title">Bir request, collage'ın span, metric ve panic recovery'den oluşan koruması içinde önce Config.Plugins içindeki plugin'lerin middleware'lerinden liste sırasıyla, sonra app.Use middleware'lerinden register sırasıyla geçer. Routing önce static mount'lara ve app.Handle prefix'lerine bakar, geri kalan her şeyi router'a verir; router da bir page, onun fragment path'lerinden birini, bir document ya da bir action seçer.</title>
<defs><marker id="mw-arrow" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto"><path class="d-head" d="M0 0 8 4 0 8z" /></marker></defs>
<rect class="d-pill" x="110" y="8" width="100" height="26" rx="13" /><text x="160" y="25" text-anchor="middle">tarayıcı</text>
<line class="d-flow" x1="160" y1="34" x2="160" y2="52" marker-end="url(#mw-arrow)" />
<rect class="d-slot d-slot-a" x="6" y="56" width="308" height="296" rx="8" />
<text class="d-slot-text-a d-small" x="18" y="72">collage koruması</text><text class="d-small" x="302" y="72" text-anchor="end">span · metric · panic recovery</text>
<rect class="d-frame" x="30" y="84" width="260" height="38" rx="6" /><text x="160" y="100" text-anchor="middle">plugin middleware'leri</text><text class="d-small" x="160" y="114" text-anchor="middle">Config.Plugins, liste sırasıyla</text>
<line class="d-flow" x1="160" y1="122" x2="160" y2="136" marker-end="url(#mw-arrow)" />
<rect class="d-frame" x="30" y="140" width="260" height="38" rx="6" /><text x="160" y="156" text-anchor="middle">app.Use middleware'leri</text><text class="d-small" x="160" y="170" text-anchor="middle">register sırasıyla</text>
<line class="d-flow" x1="160" y1="178" x2="160" y2="192" marker-end="url(#mw-arrow)" />
<rect class="d-cmd" x="110" y="196" width="100" height="26" rx="13" /><text class="d-cmd-text" x="160" y="213" text-anchor="middle">routing</text>
<path class="d-flow" d="M135 222C135 234 61 230 61 242" marker-end="url(#mw-arrow)" />
<line class="d-flow" x1="160" y1="222" x2="160" y2="242" marker-end="url(#mw-arrow)" />
<path class="d-flow" d="M185 222C185 234 259 230 259 242" marker-end="url(#mw-arrow)" />
<rect class="d-b" x="15" y="246" width="92" height="30" rx="5" /><text class="d-on-b d-small" x="61" y="265" text-anchor="middle">static mount</text>
<rect class="d-frame" x="114" y="246" width="92" height="30" rx="5" /><text class="d-small" x="160" y="265" text-anchor="middle">app.Handle</text>
<rect class="d-cmd" x="213" y="246" width="92" height="30" rx="15" /><text class="d-cmd-text" x="259" y="265" text-anchor="middle">router</text>
<text class="d-small" x="110" y="292" text-anchor="middle">önce prefix ile eşleşir</text>
<path class="d-flow" d="M245 276C245 290 210 288 210 302" marker-end="url(#mw-arrow)" />
<path class="d-flow" d="M273 276C273 290 279 288 279 302" marker-end="url(#mw-arrow)" />
<rect class="d-a" x="176" y="306" width="68" height="34" rx="5" /><text class="d-on-a d-small" x="210" y="320" text-anchor="middle">page ya da</text><text class="d-on-a d-small" x="210" y="332" text-anchor="middle">document</text>
<rect class="d-c" x="249" y="306" width="60" height="34" rx="5" /><text class="d-on-c" x="279" y="327" text-anchor="middle">action</text>
</svg>
<figcaption>Her request collage'ın koruması içinde çalışır: span'i, metric'leri ve bir panic'i 500'e çeviren panic recovery. Bunun içinde önce Config.Plugins içindeki plugin'lerin middleware'leri liste sırasıyla, sonra app.Use middleware'leri register edildikleri sırayla çalışır; app.RegisterPlugin ile eklenen bir plugin, çağrının yapıldığı yerde durur. Bunlardan herhangi biri request'e kendisi cevap verip onu orada durdurabilir. Geçen request route edilir: önce bir static mount'a ya da bir app.Handle prefix'ine bakılır, gerisini router alır.</figcaption>
</figure>

### Plugin'lerin middleware'i ve sizinki

v0.38.0'dan beri bir plugin'in middleware'i, plugin'in register edildiği yerde
register edilmiş sayılır. `Config.Plugins` içindeki plugin'ler `collage.New`'da
register edilir. Bu yüzden onların middleware'i `app.Use` ile eklenen her şeyin
dışındadır.

Sizin middleware'iniz de bu sayede onların request context'ine koyduklarını okur:
session'ı, bir flash mesajını. Giriş yapmış kullanıcıyı her request'te bir kez
yüklemek tek bir middleware'dir:

```go
app.Use(func(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := session.FromContext(r.Context()).Get(session.UserKey)
		user, err := users.Active(r.Context(), id) // devre dışı bir hesap için nil
		if err != nil {
			http.Error(w, "unavailable", http.StatusServiceUnavailable)
			return
		}
		next.ServeHTTP(w, r.WithContext(withUser(r.Context(), user)))
	})
})
```

Sizin middleware'inizin koyduğunu okuması gereken bir plugin, örneğin o
kullanıcıya göre sınırlama yapan bir rate limit, ondan sonra `app.RegisterPlugin`
ile register edilir. Middleware'i çağrının yapıldığı yere girer: kendisinden önce
`app.Use` ile eklenenlerin içine, sonra eklenenlerin dışına.

```go
app.Use(loadUser)
app.RegisterPlugin(ratelimit.New(ratelimit.Options{KeyFunc: userID}))
```

`RegisterPlugin`, template fonksiyonu ekleyen bir plugin'i reddeder. Böyle bir
plugin `Config.Plugins` içinde olmalıdır.

v0.38.0'dan önce her plugin'in middleware'i uygulamanınkinin içinde çalışıyordu ve
`app.Use` middleware'inde `session.FromContext` `nil` dönüyordu.

### Path'ler önce temizlenir

Dot segment'ler ya da çift slash içeren bir path (`/a/../b`, `/a//b`, `/a/./b`),
middleware ve plugin'ler dahil hiçbir şey onu okumadan önce temiz yazımına redirect
edilir. `GET` ve `HEAD` için 301, body taşıyan bir metot için 308 kullanılır; query
korunur (v0.24.0'dan beri).

Bu yüzden `/_collage/`'ı atlayan ya da `/admin/`'i koruyan bir middleware
`/_collage/../admin` ile hiç karşılaşmaz. Bir prefix kontrolü, router'ın route
edeceği path'in kontrolüdür.

Encode edilmiş bir slash router için bir ayırıcı değildir. Bu yüzden v0.34.0'dan
beri:

- `%2F` içeren bir path hiçbir route'a ulaşmaz ve router ona 404 ile cevap verir.
- Aynı zamanda temiz olmayan bir path (`/blog/%2Fhello`) `/blog/hello`'ya
  temizlenmez; middleware'iniz çalışmadan önce 404 ile cevaplanır.
- Middleware'iniz decode edilmiş `r.URL.Path`'i okur ve orada `/public%2Fsecret`,
  `/public/secret`'tir. Eskiden router orayı tek bir segment olarak okurdu. Böylece
  `/public/`'i geçiren bir middleware, request'i `/{slug}`'a geçirmiş olurdu.

### Body middleware'den önce sınırlanır

v0.34.0'dan beri bir request'in body'si, ilk middleware çalışmadan önce sınırlanır.
Sınır, request'in yönlendirildiği
[action](/docs/forms-and-actions#request-bodies-are-bounded)'ın limitidir; diğer her
şey için `Server.MaxBodyBytes`'tır.

- Body'yi okuyan bir middleware (bir signature kontrolü, bir logger, form'u parse
  eden bir method override) onu sınırlı olarak okur.
- Body'si bir middleware tarafından limitinin ötesine kadar okunmuş bir action `413`
  ile cevap verir.
- `app.Handle` ile mount edilmiş bir handler'a dokunulmaz.

### Data handler'lara değer aktarmak

Middleware'in request'in context'ine koyduğu değerleri data handler'lar `ctx`
olarak alır. Oturum açmış bir kullanıcı, bir feature flag ya da bir tenant, ona
ihtiyaç duyan fragment'lere bu yolla ulaşır:

```go
type userKey struct{}

app.Use(func(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if user, ok := sessions.User(r); ok {
			r = r.WithContext(context.WithValue(r.Context(), userKey{}, user))
		}
		next.ServeHTTP(w, r)
	})
})
```

```go
func accountData(ctx context.Context, rc *collage.RenderContext) (accountView, []string, error) {
	user, ok := ctx.Value(userKey{}).(User)
	if !ok {
		return accountView{}, nil, nil // signed out: the fragment renders its signed-out state
	}
	return accountView{Name: user.Name}, nil, nil
}
```

Cache'e dikkat edin. Kullanıcıya göre farklı render edilen bir page URL'ye göre
cache'lenmemelidir. Aksi hâlde ilk okuyucunun gördüğü versiyon herkese gösterilir.
İki yol var:

- Böyle bir page'i dynamic bırakın, yani ona `Static()` ya da `Incremental(ttl)`
  vermeyin (data handler'ı olan ve strateji tanımlamayan bir page zaten dynamic'tir).
- Ya da neye göre değiştiğini aşağıda anlatılan `collage.Vary` ile cache'e bildirin.

Cache'lenmiş bir render ne okuyucunun cookie'lerini ve credential'larını, ne de
middleware'inizin context'e koyduğu değerleri görür
([paylaşılan bir render neyi görür](/docs/caching#what-a-shared-render-sees)):
v0.39.0'dan beri, yanlışlıkla cache'lediğiniz kullanıcıya özel bir page — `userKey`
değeri gizlendiği için — ilk okuyucunun hesabını herkese göstermek yerine herkese
oturum kapalıymış gibi render edilir. Onu dynamic bırakın.

Static export request olmadan render eder, bu yüzden yazdığı dosyalar için hiçbir
middleware çalışmaz. Context'ten değer okuyan bir data handler, bu değer olmadığında
da doğru çalışmalıdır. Oturum açmamış bir okuyucu için bunu zaten yapması gerekir.

v0.52.0'dan beri build ardından [her dosyanın
header'larını](/docs/static-export#headers-and-redirects) handler üzerinden
yakalayabilir ve bu request'ler middleware'den geçer:

- Trafiği sayan ya da sınırlayan bir middleware, `collage.IsCapture(r.Context())`
  true olduğunda request'i atlar.
- Header set eden middleware ise onu her zamanki gibi işler.

## Bir header'a bağlı içerik: `collage.Vary`

Page cache'in key'i URL'dir. İçeriği bir request header'ına bağlı olan cache'lenmiş bir
page, ilk hangi versiyonu render edildiyse onu herkese sunar. Bu header
`Accept-Language`, bir cihaz sınıfı ya da bir A/B grubu olabilir. Middleware'den
çağrılan `collage.Vary`, cache key'ine yeni bir boyut ekler:

```go
app.Use(func(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lang := "en"
		if strings.HasPrefix(r.Header.Get("Accept-Language"), "tr") {
			lang = "tr"
		}
		if err := collage.Vary(r, "Accept-Language", lang); err != nil {
			app.Logger().Error("vary", "error", err)
		}
		next.ServeHTTP(w, r)
	})
})
```

Bir fragment'in data handler'ı değeri `collage.Varied` ile geri okur:

```go
func languageData(ctx context.Context, rc *collage.RenderContext) (string, []string, error) {
	lang, _ := collage.Varied(ctx, "Accept-Language") // "tr" or "en"
	return greeting(lang), nil, nil
}
```

- **Key'e ham header değil, sizin çözümlediğiniz değer girer.** Tarayıcılar
  tercihlerini yüz farklı şekilde yazar: `tr-TR,tr;q=0.9`, `tr`, `tr-TR`. Bunların
  hepsi aynı page'dir. Header'ı, page'lerinizin gerçekten farklılaştığı birkaç
  değere indirin. Cache de o sayıda entry tutar. Asla `Cookie`'nin ya da
  `Authorization`'ın kendisine göre vary etmeyin: v0.64.0'dan beri bu değer
  memory'de tutulur ve her [warm render](/docs/caching#warming-after-an-invalidation)
  ile yeniden gönderilir.
- **Değeri context'ten değil, `collage.Varied` ile okuyun.** Cache'lenebilir bir
  page'de tek bir render birçok okuyucuya sunulur. Bu yüzden middleware'inizin
  `context.WithValue` ile request context'ine koyduğu bir değer o render'dan
  gizlenir: aksi hâlde ilk okuyucunun değeri olur ve geri kalan herkes için
  donar. Bir `Vary` değeri okunabilir çünkü key'in içindedir ve her değerin kendi
  cache'lenmiş page'i vardır. Bkz.
  [paylaşılan bir render neyi görür](/docs/caching#what-a-shared-render-sees).
- **Header'ın adı response'un `Vary` header'ına eklenir.** Böylece sizinle okuyucu
  arasındaki bir CDN ya da proxy de versiyonları birbirinden ayrı tutar. Bu header
  yalnızca public olarak cache'lenebilen response'larda set edilir. `no-store` bir
  response'ta ayrı tutulacak bir şey yoktur. v0.34.0'dan beri header'ın adı,
  middleware'inizin yazdığı bir `Vary`'nin yanına eklenir. Eskiden onun yerine
  geçerdi.
- **`Vary`'yi middleware'den çağırın.** Middleware bittiğinde ve routing
  başladığında bildirimler kapanır. Bu, her route için geçerlidir: cache'lenen ya
  da cache'lenmeyen bir page, bir document, bir action, bir mount ya da bir `app.Handle`
  handler'ı. Routing'den sonra, örneğin bir data handler'dan çağrılırsa `Vary`
  çalışıyormuş gibi yapmaz, `collage.ErrVaryTooLate` döner. Bu davranış v0.11.0'dan
  beri böyledir. Öncesinde bu hatayı yalnızca cache'lenen bir page'de dönüyor, diğer
  yerlerde sessizce hiçbir şey yapmıyordu. collage'ın sunmadığı bir request'te
  çağrılırsa `collage.ErrVaryOutsideRequest` döner.
- Aynı header'ı iki kez bildirirseniz son değer geçerli olur.

Dil seçimi de bu yolla yapılır. collage locale'i yalnızca URL'den seçer,
`Accept-Language`'ı okumayı size bırakır. Tarayıcıyı middleware'den `/tr`'ye
redirect edebilirsiniz. Ya da her dil için tek bir URL render edip bunu `Vary` ile
bildirebilirsiniz. Ayrıntılar için [Link'ler ve locale'ler](/docs/links-and-locales)
sayfasına bakın.

## Kendi handler'ınız: `app.Handle`

```go
api := http.NewServeMux()
api.HandleFunc("GET /api/users/{id}", getUser)
api.HandleFunc("POST /api/users", createUser)

if err := app.Handle("/api/", api); err != nil {
	return err
}
```

Path'i prefix ile başlayan her request, path'i değiştirilmeden handler'ınıza gider.
Yani `getUser` `/api/users/42`'yi görür. Handler'ınız farklı bir path bekliyorsa onu
`http.StripPrefix` ile sarın.

Herhangi bir `http.Handler` kullanabilirsiniz: `http.ServeMux`, chi, bir gRPC
gateway ya da bir reverse proxy.

**collage bu handler'a hiçbir şey eklemez.** Handler'ın neyi kabul edeceğine, ne
kadarını okuyacağına ve neyi cache'leyeceğine siz karar verirsiniz.

| Handler'ın almadıkları | Handler'ın her request gibi aldıkları |
| --- | --- |
| request forgery kontrolü | span ve metric'ler |
| body boyutu limiti | panic koruması |
| cache | `app.Use` ile register edilen middleware'ler |
| | graceful shutdown sırasında bitmesinin beklenmesi |

Handler'ın döndüğü bir 5xx, plugin'lerin error hook'larına
`collage.ErrHandlerFailed` olarak bildirilir. 4xx cevapları ise handler'ın kendi
sorumluluğundadır.

### Prefix'ler ve çakışmalar

Prefix `/` ile başlamalıdır ve tek başına `/` olamaz
(`collage.ErrInvalidPrefix`).

- **`/` ile biterse** altındaki her path'i üstlenir.
- **Sonunda `/` yoksa** tek bir tam path'tir ve yalnızca o path'e cevap verir:
  `app.Handle("/metrics", h)`, `/metrics`'i sunar, `/metrics/x`'i sunmaz
  (v0.24.0'dan beri; öncesinde prefix'in `/` ile bitmesi gerekiyordu).

`/`'deki bir handler bütün page'lerin bütün request'lerini alırdı. Gerçekten
istediğiniz buysa `app.Handler()`'ı kendi mux'ınızın içine koyun.

Bir prefix, collage'ın route ettiği hiçbir şeyi kapsayamaz:

- `/api/count`'ta bir action varken `app.Handle("/api/", ...)` çağrısı
  `collage.ErrMountShadowsRoute` ile reddedilir.
- Çakışan iki handler ya da çakışan bir handler ile bir static mount da
  `collage.ErrMountConflict` ile reddedilir.

Bu kontrol uygulama başlarken çalışır. Bu yüzden hangisinin önce register edildiği
önemli değildir. Hatayı `app.ListenAndServe` ve `app.Start` döner.

Kontrol yalnızca literal path'leri karşılaştırır. Prefix'in altında eşleşebilecek
dinamik bir pattern'i göremez. Örneğin `/{rest...}`'te catch-all bir page ve
`/api/`'de bir handler varsa `/api/users` URL'si handler'a gider. Bir prefix mount
ettiğinizde bu ödünü kabul etmiş olursunuz.

### API'nizden invalidate etmek

Handler'ınız da cache'teki page'leri diğer her şey gibi invalidate edebilir. Bir API çağrısı
içeriği değiştirdiğinde, o içerikten oluşturulan page'lerin bildirdiği tag'leri
invalidate edin:

```go
func updatePost(app *collage.App, posts *store.Posts) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("slug")
		if err := posts.Update(r.Context(), slug, r.Body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := app.InvalidateTags(r.Context(), "post:"+slug, "blog:posts"); err != nil {
			app.Logger().Error("invalidate", "slug", slug, "error", err)
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
```

`app.InvalidateTagsN` aynı işi yapar ve ek olarak tag'lerin kaç cache entry'sine
ulaştığını döner. Bu sayı bir webhook'un response'unda ya da bir log satırında işe
yarar.

En yaygın örnek bir CMS webhook'udur:

1. CMS, değişen entry ile `/api/hooks/cms`'i çağırır.
2. Handler webhook'un imzasını kontrol eder ve o entry'nin tag'ini invalidate eder.

Yalnızca o entry'nin page'leri ve aynı tag'i taşıyan `collage.Cached` değerleri
yeniden render edilir. Ayrıntılar için [Caching](/docs/caching) sayfasına
bakın.

## Ne zaman bunun yerine action kullanmalı

`app.Handle`, page'lerinizle ilgisi olmayan kod içindir: kendi authentication'ı olan
bir API, proxy'lediğiniz bir servis ya da zaten kullandığınız bir router. Siteye ait
bir endpoint için bir [action](/docs/forms-and-actions) kullanın:

- Bir form post'u ya da page'lerinizden birinden yapılan bir `fetch()`. Action
  request forgery token'ını kontrol eder, handler etmez. Bu yüzden `app.Handle`
  üzerinden tarayıcıya açılan bir `POST`'u, başka herhangi bir site okuyucunuz adına
  gönderebilir.
- Sınırlanması gereken her şey. Bir action'ın body'si `Server.MaxBodyBytes`
  (varsayılan 4 MiB) ya da action'ın kendi `WithMaxBodyBytes` değeri ile sınırlanır.
- Bir fragment ya da page ile cevap veren bir endpoint, örneğin bir form'un değişen
  kısmı ya da bir validation hatası. Action `collage.RenderFragment` ya da
  `collage.RenderPage` dönebilir.
- Başarılı olduğunda `ActionResult.InvalidateTags` ile cache'teki page'leri invalidate eden
  bir endpoint.
- Sitenin her locale'inde bulunması gereken bir URL. Action'ın path'leri, page'in
  path'leri gibi locale'e göre tanımlanır.

```go
count := collage.NewAction("count").
	WithPath("en", "/api/count").
	WithMethods(http.MethodPost).
	WithHandler(func(context.Context, *collage.RenderContext) (*collage.ActionResult, error) {
		result, err := collage.JSON(http.StatusOK, countResponse{Count: store.Increment()})
		if err != nil {
			return nil, err
		}
		result.InvalidateTags = []string{store.CountTag}
		return result, nil
	}).
	Build()
```

Bu, `collage new`'un oluşturduğu projedeki sayaçtır. Forgery token'ı isteyen ve
sayacı gösteren page'i invalidate eden bir JSON endpoint'idir. Token gönderemeyen bir
servisten gelen webhook için `WithoutCSRF()` ile tanımlanmış bir action da body
limitinden yararlanmaya devam eder.

## Örnek: bir ekibin iş listesi

Her üyenin giriş yaptığı ve bir rolü olduğu bir ekip iş listesi düşünün: admin
ekibin her görevini, member ise yalnızca kendisine atanmış görevleri değiştirebilir.
Page'ler ekibin görevlerini gösterir. Ekibin komut satırı aracının zaten konuştuğu
bir JSON API de olduğu gibi kalır. Bu sayfanın üç konusu burada buluşur:

- kimin sorduğunu her request'te bir kez bulan tek bir middleware;
- `app.Handle` ile mount edilen mevcut API;
- API'nin, değiştirdiği listenin tag'ini invalidate etmesi.

### Kim soruyor, her request'te bir kez

Middleware ya bir API token'ını ya da session cookie'sini okur ve kullanıcıyı,
kendine ait bir tipteki key ile request'in context'ine koyar:

```go
type User struct {
	ID      string
	TeamID  string
	Role    string // "admin" or "member"
	ByToken bool   // signed in with an API token, not the session cookie
}

// Users is your own data layer.
type Users interface {
	BySession(ctx context.Context, cookie string) (User, bool, error)
	ByToken(ctx context.Context, token string) (User, bool, error)
}

type userKey struct{}

// UserFrom returns the signed-in user; false for a reader who is not.
func UserFrom(ctx context.Context) (User, bool) {
	user, ok := ctx.Value(userKey{}).(User)
	return user, ok
}

// loadUser resolves who is asking once per request. Pages, actions and the
// API behind it read the result with UserFrom.
func loadUser(users Users) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var (
				user User
				ok   bool
				err  error
			)
			if token, bearer := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer "); bearer {
				user, ok, err = users.ByToken(r.Context(), token)
				user.ByToken = true
			} else if cookie, cerr := r.Cookie("sid"); cerr == nil {
				user, ok, err = users.BySession(r.Context(), cookie.Value)
			}
			if err != nil {
				http.Error(w, "unavailable", http.StatusServiceUnavailable)
				return
			}
			if ok {
				r = r.WithContext(context.WithValue(r.Context(), userKey{}, user))
			}
			next.ServeHTTP(w, r)
		})
	}
}
```

Middleware routing'den önce çalıştığı için aynı sorgu bir page'e, bir action'a ve
mount edilmiş API'ye aynı şekilde hizmet eder. Arkasındaki hiçbir şey okuyucunun
kim olduğunu veritabanına yeniden sormaz.

### Olduğu gibi mount edilen API

```go
type Task struct {
	ID         string
	TeamID     string
	AssigneeID string
	Title      string
	Done       bool
}

// Tasks is your own data layer.
type Tasks interface {
	Get(ctx context.Context, id string) (Task, bool, error)
	SetDone(ctx context.Context, id string, done bool) error
	ByTeam(ctx context.Context, teamID string) ([]Task, error)
}

// taskAPI is the JSON API the team's command-line tool already talks to.
func taskAPI(app *collage.App, tasks Tasks) http.Handler {
	api := http.NewServeMux()
	api.HandleFunc("PATCH /api/tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
		user, ok := UserFrom(r.Context())
		if !ok || !user.ByToken {
			http.Error(w, "token required", http.StatusUnauthorized)
			return
		}
		task, found, err := tasks.Get(r.Context(), r.PathValue("id"))
		if err != nil {
			http.Error(w, "unavailable", http.StatusServiceUnavailable)
			return
		}
		if !found || task.TeamID != user.TeamID {
			http.NotFound(w, r)
			return
		}
		if user.Role != "admin" && task.AssigneeID != user.ID {
			http.Error(w, "not your task", http.StatusForbidden)
			return
		}
		var body struct {
			Done bool `json:"done"`
		}
		// collage does not bound a mounted handler's body; this is ours to do.
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&body); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		if err := tasks.SetDone(r.Context(), task.ID, body.Done); err != nil {
			http.Error(w, "unavailable", http.StatusServiceUnavailable)
			return
		}
		if err := app.InvalidateTags(r.Context(), "team:"+task.TeamID+":tasks"); err != nil {
			app.Logger().Error("invalidate", "task", task.ID, "error", err)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	return api
}
```

İçindeki her kontrol, `app.Handle`'ın yaptığı ve yapmadığı şeylerden gelir:

- **Cookie değil, token.** Handler'da request forgery kontrolü yoktur. Bu yüzden
  handler'ın session cookie'sine dayanarak kabul ettiği bir yazma işlemi
  korumasızdır. Onun üzerinden tarayıcıya açık bir `POST`'u da başka herhangi bir
  site okuyucu adına gönderebilir. API yalnızca token ile giriş yapılmış
  request'leri kabul eder; page'ler görevleri forgery token'ını kontrol eden bir
  [action](/docs/forms-and-actions) üzerinden değiştirir.
- **Body elle sınırlanır.** collage mount edilmiş bir handler'ın body'sine
  dokunmaz, bu yüzden handler onu kendisi `http.MaxBytesReader` ile sarar.
- **Rol, verinin olduğu yerde kontrol edilir.** Middleware yalnızca kimin
  sorduğunu söyler; o kullanıcının neyi değiştirebileceğine API karar verir.
- **Action'lar `/api/` dışında durur.** Bir prefix bir route'u kapsayamaz. Bu
  yüzden bu mount'un yanında `/api/tasks/new`'de bir action olsaydı uygulama
  `collage.ErrMountShadowsRoute` ile başlamazdı.

### API'nin invalidate ettiği liste

İş listesi fragment'i kullanıcıyı context'ten okur ve ekibin listesini, ekibe ait
bir tag ile [data cache](/docs/caching#caching-data-across-pages)'te tutar:

```go
var teamTasksKey = collage.NewKey[[]Task]("team-tasks")

// taskListData feeds the task list fragment. The page has no declared
// strategy, so it is dynamic and sees the user; the list itself is shared
// by everyone on the team and fresh until the API changes it.
func taskListData(tasks Tasks) func(context.Context, *collage.RenderContext) ([]Task, []string, error) {
	return func(ctx context.Context, rc *collage.RenderContext) ([]Task, []string, error) {
		user, ok := UserFrom(ctx)
		if !ok {
			return nil, nil, nil // signed out: the fragment links to the sign-in page
		}
		list, err := collage.Cached(rc, teamTasksKey.With(user.TeamID), 0,
			[]string{"team:" + user.TeamID + ":tasks"},
			func(ctx context.Context) ([]Task, error) { return tasks.ByTeam(ctx, user.TeamID) })
		return list, nil, err
	}
}

func taskList(tasks Tasks) *collage.Fragment {
	return collage.NewFragment("task-list", "fragments/task-list.html").
		WithData(collage.LoadTagged(taskListData(tasks))).
		Build()
}
```

Page kullanıcıya göre render edildiği için dynamic kalır. Arkasındaki liste bütün
ekip için aynıdır, bu yüzden bir kez çekilir ve paylaşılır: sıfır `ttl`, onu tag'i
invalidate edilene kadar taze tutar.

Komut satırı aracı bir görevi tamamlandı olarak işaretlediğinde API'nin
`InvalidateTags` çağrısı o ekibin listesini stale olarak işaretler. Sonraki render'a
eski liste hemen sunulur; bu sırada arka planda tek bir çekme işlemi onu yeniler. Bu
işlem bittikten sonraki bir request değişikliği görür
([nasıl davranır](/docs/caching#how-it-behaves)). Diğer ekiplerin listelerine
dokunulmaz.

Hepsini bağlamak için iki çağrı yeter:

```go
if err := app.Use(loadUser(users)); err != nil {
	return err
}
if err := app.Handle("/api/", taskAPI(app, tasks)); err != nil {
	return err
}
```
