---
description: app.Use ile standart net/http middleware'leri, collage.Vary ile bir header'a bağlı önbellek anahtarları ve app.Handle ile kendi http.Handler'ınız.
---

# Middleware ve kendi API'niz

collage sayfa render eder. Bir Go programının HTTP üzerinden yaptığı geri kalan her
şeyi — bir API, kimlik doğrulama, dil müzakeresi, hız sınırlama — zaten yazacağınız
gibi yazarsınız ve iki noktadan birine takarsınız: her isteğin etrafında çalışan
**middleware** ve bir önekin altındaki her isteği yanıtlayan **kendi handler'ınız**.

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

Standart `func(http.Handler) http.Handler` biçimindedir; dolayısıyla `net/http`
için yazılmış her middleware değişmeden çalışır. İlk kaydedilen en dıştakidir. Her
kayıt gibi uygulama başlamadan önce yapılmalıdır; sonrasında
`collage.ErrAppStarted` döner.

Nerede çalıştığıyla ilgili iki şey:

- **Route belirlenmeden önce.** Her isteği görür — sayfaları, document'ları,
  action'ları, mount edilmiş statik dosyaları ve kendi handler'larınızı — 404 ile
  bitenler de dahil. Bir isteği kendisi, bir 401 ya da bir yönlendirmeyle
  yanıtlayabilir; o zaman arkasındaki hiçbir şey çalışmaz.
- **Framework'ün koruması içinde.** collage'ın span'i, metrikleri ve panic
  kurtarması içinde çalışır. `app.Handler()`'ı kendi middleware'inizle sarmaktan
  farkı budur: middleware'inizdeki bir panic, kopan bir bağlantı yerine normal hata
  yolunda sıradan bir 500 olur ve kendisinin yanıtladığı bir istek de diğerleri gibi
  sayılır.

### Data handler'lara değer geçirmek

Middleware'in isteğin context'ine koyduğu şey, data handler'ların `ctx` olarak
aldığı şeydir. Oturum açmış bir kullanıcı, bir özellik bayrağı ya da bir kiracı,
ona ihtiyaç duyan fragment'lere böyle ulaşır:

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

Önbelleğe dikkat edin. Kullanıcıya göre farklı render edilen bir sayfa URL'ye göre
önbelleğe alınmamalıdır; yoksa ilk okuyucunun sürümü herkesin sürümü olur: onu
`Dynamic()` yapın ya da neye göre değiştiğini aşağıdaki `collage.Vary` ile
önbelleğe söyleyin.

Statik dışa aktarma istek olmadan render eder, bu yüzden sırasında hiçbir
middleware çalışmaz. Bir context değerini okuyan data handler, o değerin yokluğuyla
başa çıkabilmelidir — oturum açmamış bir okuyucu için zaten başa çıkması gerekir.

## Bir header'a bağlı içerik: `collage.Vary`

Sayfa önbelleğinin anahtarı URL'dir. İçeriği bir istek header'ına —
`Accept-Language`'e, bir cihaz sınıfına, bir A/B grubuna — bağlı olan, önbellekteki
bir sayfa, ilk hangi sürüm render edildiyse onu herkese sunar. Middleware'den
çağrılan `collage.Vary`, önbellek anahtarına bir boyut ekler:

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

- **Anahtara ham header değil, sizin çözdüğünüz değer girer.** Tarayıcılar
  tercihlerini yüz farklı biçimde yazar — `tr-TR,tr;q=0.9`, `tr`, `tr-TR` — ve
  hepsi tek bir sayfadır. Header'ı, sayfalarınızın gerçekten farklılaştığı birkaç
  değere indirgeyin; önbellek de o kadar girdi tutar.
- **Header'ın adı yanıtın `Vary` header'ına girer**; böylece sizinle okuyucu
  arasındaki bir CDN ya da proxy de sürümleri ayrı tutar. Yalnızca herkese açık
  biçimde önbelleğe alınabilen yanıtlarda ayarlanır; `no-store` bir yanıtın ayrı
  tutacak bir şeyi yoktur.
- **Onu middleware'den çağırın.** Bildirimler, middleware bitip route belirleme
  başladığında, her route'ta kapanır — önbellekte olsun olmasın bir sayfa, bir
  document, bir action, bir mount, bir `app.Handle` handler'ı. Route belirlendikten
  sonra, örneğin bir data handler'dan çağrılırsa, `Vary` çalışıyormuş gibi yapmak
  yerine `collage.ErrVaryTooLate` döndürür (v0.11.0'dan itibaren; öncesinde bunu
  yalnızca önbellekteki bir sayfada yapar, başka yerlerde sessizce hiçbir şey
  yapmazdı). collage'ın sunmadığı bir istekte çağrılırsa
  `collage.ErrVaryOutsideRequest` döndürür.
- Aynı header'ı iki kez bildirmek son değeri tutar.

Dil müzakeresi de böyle yapılır: collage locale'i yalnızca URL'den seçer ve
`Accept-Language`'i okumayı size bırakır. Middleware'den tarayıcıyı `/tr`'ye
yönlendirin ya da her dil için bir URL render edip onu `Vary` ile bildirin. Bkz.
[Bağlantılar ve locale'ler](/docs/links-and-locales).

## Kendi handler'ınız: `app.Handle`

```go
api := http.NewServeMux()
api.HandleFunc("GET /api/users/{id}", getUser)
api.HandleFunc("POST /api/users", createUser)

if err := app.Handle("/api/", api); err != nil {
	return err
}
```

Yolu önekle başlayan her istek, yol değiştirilmeden handler'ınıza gider —
`getUser` `/api/users/42`'yi görür. Handler başka bir şey bekliyorsa onu
`http.StripPrefix` ile sarın. Herhangi bir `http.Handler` iş görür: `http.ServeMux`,
chi, bir gRPC gateway, bir reverse proxy.

**collage ona hiçbir şey yapmaz.** İstek sahteciliği denetimi yok, gövde boyutu
sınırı yok, önbellek yok. Neyi kabul edeceğine, ne kadar okuyacağına ve neyi
önbelleğe alacağına siz karar verirsiniz. Aldığı şey, her isteğin aldığıdır: span,
metrikler, panic koruması, `app.Use` ile kaydedilmiş middleware'ler ve düzgün
kapanışta bitmesinin beklenmesi. Verdiği bir 5xx yanıtı, plugin'lerin hata
hook'larına `collage.ErrHandlerFailed` olarak raporlanır; 4xx yanıtları ise kendi
bileceği iştir.

### Önekler ve çakışmalar

Önek `/` ile başlamalı ve bitmelidir, tek başına `/` olamaz
(`collage.ErrInvalidHandlerPrefix`). `/`'deki bir handler her sayfadan her isteği
alırdı; gerçekten istediğiniz buysa, bunun yerine `app.Handler()`'ı kendi mux'ınızın
içine koyun.

Bir önek, collage'ın yönlendirdiği hiçbir şeyi kapsayamaz. `/api/count`'taki bir
action'ın yanındaki `app.Handle("/api/", ...)`, `collage.ErrMountShadowsRoute` ile
reddedilir; çakışan iki handler ya da bir handler ile bir statik mount da
`collage.ErrMountConflict` ile reddedilir. Denetim uygulama başlarken çalışır,
dolayısıyla hangisinin önce kaydedildiğinin önemi yoktur; hatayı
`app.ListenAndServe` ve `app.Start` döndürür.

Denetim sabit yolları karşılaştırır. Önekin altında eşleşecek dinamik bir
pattern'i göremez: `/{rest...}`'te her şeyi yakalayan bir sayfa ve `/api/`'de bir
handler varken `/api/users` URL'si handler'a gider; bir öneki mount etmenin bedeli
budur.

### API'nizden geçersiz kılmak

Handler'ınız, önbellekteki sayfaları başka her şey gibi düşürebilir. Bir API
çağrısı içeriği değiştirdiğinde, o içerikten kurulan sayfaların bildirdiği
etiketleri geçersiz kılın:

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

`app.InvalidateTagsN` aynısını yapar ve ayrıca etiketlerin kaç önbellek girdisine
ulaştığını döndürür; bu, bir webhook'un yanıtında ya da bir log satırında işe yarar.

Olağan durum bir CMS webhook'udur: CMS, değişen girdiyle `/api/hooks/cms`'i çağırır,
handler webhook'un imzasını denetler ve o girdinin etiketini geçersiz kılar.
Yalnızca o girdinin sayfaları ve aynı etiketi taşıyan `collage.Cached` değerleri
yeniden render edilir. Bkz. [Önbellek](/docs/caching).

## Ne zaman bunun yerine bir action kullanmalı

`app.Handle`, sayfalarınızla ilgili olmayan kod içindir — kendi kimlik doğrulaması
olan bir API, proxy'lediğiniz bir servis, zaten sahip olduğunuz bir router. Siteye
ait bir endpoint için bir [action](/docs/forms-and-actions) kullanın:

- Bir form gönderimi ya da sayfalarınızdan birinden yapılan bir `fetch()`. Action
  istek sahteciliği token'ını denetler, handler denetlemez; dolayısıyla
  `app.Handle` üzerinden tarayıcıya açık bir `POST`, başka herhangi bir sitenin
  okuyucunuz adına yapabileceği bir istektir.
- Sınırlanması gereken her şey. Bir action'ın gövdesi `Server.MaxBodyBytes`
  (varsayılan olarak 4 MiB) ya da kendi `WithMaxBodyBytes`'ı ile sınırlanır.
- Bir fragment ya da sayfayla yanıt veren bir endpoint — bir formun değişen kısmı,
  bir doğrulama hatası — çünkü bir action `collage.RenderFragment` ya da
  `collage.RenderPage` döndürebilir.
- Başarılı olduğunda `ActionResult.InvalidateTags` üzerinden önbellekteki
  sayfaları düşüren bir endpoint.
- Sitenin her dilinde var olması gereken bir URL: bir action'ın yolları, bir
  sayfanınki gibi locale'e göre anahtarlanır.

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

Bu, `collage new`'un oluşturduğu projedeki sayaçtır: sahtecilik token'ına ihtiyaç
duyan ve sayacı gösteren sayfayı geçersiz kılan bir JSON endpoint'i. Token
gönderemeyen bir servisten gelen webhook için, `WithoutCSRF()` ile tanımlanmış bir
action yine de gövde sınırından yararlanır.
