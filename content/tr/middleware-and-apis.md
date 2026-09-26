---
description: app.Use ile standart net/http middleware'leri, collage.Vary ile bir header'a bağlı cache key'leri ve app.Handle ile kendi http.Handler'ınız.
reference: Vary, SkipCache, ErrVaryTooLate, ErrMountShadowsRoute
---

# Middleware ve kendi API'niz

collage page render eder. Bir Go programının HTTP üzerinden yaptığı diğer işleri,
yani API'yi, authentication'ı, dil seçimini ve rate limiting'i her zamanki gibi
kendiniz yazarsınız. Bu kodu iki noktadan birine bağlarsınız: her request'in
etrafında çalışan **middleware**'e ya da bir prefix'in altındaki her request'e
cevap veren **kendi handler'ınıza**.

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
middleware en dıştakidir. Diğer bütün register işlemleri gibi bu da uygulama
başlamadan önce yapılmalıdır. Uygulama başladıktan sonra `app.Use`
`collage.ErrAppStarted` döner.

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

### Path'ler önce temizlenir

Dot segment'ler ya da çift slash içeren bir path (`/a/../b`, `/a//b`, `/a/./b`),
middleware ve plugin'ler dahil hiçbir şey onu okumadan önce temiz yazımına redirect
edilir. `GET` ve `HEAD` için 301, body taşıyan bir metot için 308 kullanılır; query
korunur (v0.24.0'dan beri). Bu yüzden `/_collage/`'ı atlayan ya da `/admin/`'i
koruyan bir middleware `/_collage/../admin` ile hiç karşılaşmaz. Bir prefix kontrolü,
router'ın route edeceği path'in kontrolüdür.

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
func accountData(ctx context.Context, rc *collage.RenderContext) (any, []string, error) {
	user, ok := ctx.Value(userKey{}).(User)
	if !ok {
		return accountView{}, nil, nil // signed out: the fragment renders its signed-out state
	}
	return accountView{Name: user.Name}, nil, nil
}
```

Cache'e dikkat edin. Kullanıcıya göre farklı render edilen bir page URL'ye göre
cache'lenmemelidir. Aksi hâlde ilk okuyucunun gördüğü versiyon herkese gösterilir.
Böyle bir page'i dynamic bırakın, yani ona `Static()` ya da `Incremental(ttl)`
vermeyin (data handler'ı olan ve strateji tanımlamayan bir page zaten dynamic'tir).
Ya da neye göre değiştiğini aşağıda anlatılan `collage.Vary` ile cache'e bildirin.

Static export request olmadan render eder, bu yüzden export sırasında hiçbir
middleware çalışmaz. Context'ten değer okuyan bir data handler, bu değer olmadığında
da doğru çalışmalıdır. Oturum açmamış bir okuyucu için bunu zaten yapması gerekir.

## Bir header'a bağlı içerik: `collage.Vary`

Page cache'in key'i URL'dir. İçeriği bir request header'ına bağlı olan cache'lenmiş bir
page, ilk hangi versiyonu render edildiyse onu herkese sunar. Bu header
`Accept-Language`, bir cihaz sınıfı ya da bir A/B grubu olabilir. Middleware'den
çağrılan `collage.Vary`, cache key'ine yeni bir boyut ekler:

```go
type langKey struct{}

app.Use(func(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lang := "en"
		if strings.HasPrefix(r.Header.Get("Accept-Language"), "tr") {
			lang = "tr"
		}
		if err := collage.Vary(r, "Accept-Language", lang); err != nil {
			app.Logger().Error("vary", "error", err)
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), langKey{}, lang)))
	})
})
```

- **Key'e ham header değil, sizin çözümlediğiniz değer girer.** Tarayıcılar
  tercihlerini yüz farklı şekilde yazar: `tr-TR,tr;q=0.9`, `tr`, `tr-TR`. Bunların
  hepsi aynı page'dir. Header'ı, page'lerinizin gerçekten farklılaştığı birkaç
  değere indirin. Cache de o sayıda entry tutar.
- **Header'ın adı response'un `Vary` header'ına eklenir.** Böylece sizinle okuyucu
  arasındaki bir CDN ya da proxy de versiyonları birbirinden ayrı tutar. Bu header
  yalnızca public olarak cache'lenebilen response'larda set edilir. `no-store` bir
  response'ta ayrı tutulacak bir şey yoktur.
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
`http.StripPrefix` ile sarın. Herhangi bir `http.Handler` kullanabilirsiniz:
`http.ServeMux`, chi, bir gRPC gateway ya da bir reverse proxy.

**collage bu handler'a hiçbir şey eklemez.** Request forgery kontrolü, body boyutu
limiti ya da cache yoktur. Handler'ın neyi kabul edeceğine, ne kadarını okuyacağına
ve neyi cache'leyeceğine siz karar verirsiniz. Handler, her request'in aldığı şeyleri
alır: span, metric'ler, panic koruması, `app.Use` ile register edilen middleware'ler ve
graceful shutdown sırasında bitmesinin beklenmesi. Handler'ın döndüğü bir 5xx,
plugin'lerin error hook'larına `collage.ErrHandlerFailed` olarak bildirilir. 4xx
cevapları ise handler'ın kendi sorumluluğundadır.

### Prefix'ler ve çakışmalar

Prefix `/` ile başlamalıdır ve tek başına `/` olamaz
(`collage.ErrInvalidHandlerPrefix`). `/` ile biterse altındaki her path'i üstlenir.
Sonunda `/` yoksa tek bir tam path'tir ve yalnızca o path'e cevap verir:
`app.Handle("/metrics", h)`, `/metrics`'i sunar, `/metrics/x`'i sunmaz (v0.24.0'dan
beri; öncesinde prefix'in `/` ile bitmesi gerekiyordu).

`/`'deki bir handler bütün page'lerin bütün request'lerini alırdı. Gerçekten
istediğiniz buysa `app.Handler()`'ı kendi mux'ınızın içine koyun.

Bir prefix, collage'ın route ettiği hiçbir şeyi kapsayamaz. `/api/count`'ta bir
action varken `app.Handle("/api/", ...)` çağrısı `collage.ErrMountShadowsRoute` ile
reddedilir. Çakışan iki handler ya da çakışan bir handler ile bir static mount da
`collage.ErrMountConflict` ile reddedilir. Bu kontrol uygulama başlarken çalışır. Bu
yüzden hangisinin önce register edildiği önemli değildir. Hatayı `app.ListenAndServe` ve
`app.Start` döner.

Kontrol yalnızca literal path'leri karşılaştırır. Prefix'in altında eşleşebilecek
dinamik bir pattern'i göremez. Örneğin `/{rest...}`'te catch-all bir page ve
`/api/`'de bir handler varsa `/api/users` URL'si handler'a gider. Bir prefix mount
ettiğinizde bu ödünü kabul etmiş olursunuz.

### API'nizden invalidate etmek

Handler'ınız da cache'teki page'leri diğer her şey gibi düşürebilir. Bir API çağrısı
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

En yaygın örnek bir CMS webhook'udur. CMS, değişen entry ile `/api/hooks/cms`'i
çağırır. Handler webhook'un imzasını kontrol eder ve o entry'nin tag'ini invalidate
eder. Yalnızca o entry'nin page'leri ve aynı tag'i taşıyan `collage.Cached`
değerleri yeniden render edilir. Ayrıntılar için [Caching](/docs/caching) sayfasına
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
- Başarılı olduğunda `ActionResult.InvalidateTags` ile cache'teki page'leri düşüren
  bir endpoint.
- Sitenin her locale'inde bulunması gereken bir URL. Action'ın path'leri, page'in
  path'leri gibi locale'e göre tanımlanır.

```go
count := collage.NewAction("count").
	WithPath("en", "/api/count").
	WithMethods(http.MethodPost).
	WithHandler(func(context.Context, *collage.RenderContext) (*collage.ActionResult, error) {
		result, err := collage.JSONOf(http.StatusOK, countResponse{Count: store.Increment()})
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
