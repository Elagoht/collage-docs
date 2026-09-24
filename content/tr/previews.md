---
description: Editörlere yayımlanmamış taslakları gerçek sitede gösterin; önbellek onlara yayımlanmış sayfayı, başkalarına da taslağı sunmasın.
---

# Önizlemeler

Bir CMS'te "önizle"ye basan editör, taslağı gerçek sitede, etrafında gerçek layout'la
görmek ister. Önbellekleme burada iki kez engel olur. Editöre önbelleğin zaten tuttuğu
yayımlanmış sayfa sunulmamalıdır; ona gösterilen taslak da asla önbelleğe
alınmamalıdır — yoksa diğer her okuyucunun aldığı sayfa o olur.

collage bunun için tek bir temel yapı sunar, `collage.SkipCache`; gerisini size
bırakır: kimin önizleme yapabileceği kullanıcılarınızla ilgili bir sorudur ve
collage'ın oturumlar konusunda bir görüşü yoktur.

## `collage.SkipCache`

```go
func SkipCache(r *http.Request) error
```

Middleware'den çağrıldığında tek bir isteği önizleme olarak işaretler. O istek şunları
alır:

- **Taze bir render.** Sayfa önbelleği okunmaz; bu yüzden yayımlanmış sayfanın
  önbellekteki bir kopyası asla sunulmaz.
- **Hiçbir şey saklanmaz.** Render sayfa önbelleğine yazılmaz.
- **`Cache-Control: private, no-store`**; böylece ne bir tarayıcı önbelleği, ne bir
  proxy, ne de bir CDN onu tutar.
- **Taze veri.** O render'daki her `collage.Cached` çağrısı saklanan bir değeri dönmek
  yerine veriyi çeker ve hiçbir şey saklamaz. Yayımlanmış verinin önbellekteki bir
  kopyasıyla çevrelenmiş taslak sayfayı gören bir editöre taslak gösterilmiş olmaz.

Document'ler de buna uyar: bir önizlemede istenen document ne önbellekten okunur ne de
önbelleğe yazılır. Diğer hiçbir isteğe dokunulmaz: her zamanki gibi önbellekten sunulur
ve önbellek yayımlanmış sürümü tutmaya devam eder.

`collage.Vary` gibi routing'den önce çağrılmalıdır; bu da `app.Use` ile kaydedilmiş bir
middleware'den çağrılması demektir. Routing'den sonra — örneğin bir data handler'dan —
çağrılırsa her route'ta `collage.ErrVaryTooLate` döner (v0.11.0'dan itibaren).

## Eksiksiz bir önizleme akışı

Olağan düzen, aşağıdaki de budur:

1. CMS'in önizleme düğmesi `/api/preview?secret=…&slug=…` adresini yeni bir sekmede
   açar.
2. O URL'deki bir [action](/docs/forms-and-actions) gizli değeri denetler, imzalı bir
   cookie ayarlar ve yazıya yönlendirir.
3. Middleware sonraki her istekte cookie'yi görür, `collage.SkipCache`'i çağırır ve
   data handler'lara CMS'ten taslakları istemelerini söyler.
4. Editörün işi bittiğinde başka bir action cookie'yi temizler.

İşin içinde, ikisi de ortamdan gelen iki gizli değer var: CMS'in önizleme
başlatabilmesi için onunla paylaşılan `PREVIEW_SECRET` ve yalnızca sunucunuzun bildiği,
cookie'yi imzalayan `PREVIEW_KEY`.

### Cookie'yi imzalamak

Cookie ne zaman sona ereceğini ve bunun üzerine atılmış bir imzayı tutar. Anahtara
sahip olmayan hiç kimse bir tane üretemez; süresi dolmuş olan da tarayıcı hâlâ
gönderiyor olsa bile reddedilir.

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

### Önizlemeyi başlatan action

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
				return collage.NoContent(http.StatusUnauthorized), nil
			}

			// Built from the page's name, so a slug cannot send the editor anywhere
			// but a blog post.
			target, err := app.URL("blog-post", "", map[string]string{"slug": query.Get("slug")})
			if err != nil {
				return collage.NoContent(http.StatusBadRequest), nil
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

Bir `GET` action'ı sahtecilik token'ı gerektirmez — yalnızca güvenli olmayan metotlar
denetlenir — CMS'in onu düz bir bağlantı olarak açabilmesini sağlayan da budur.
`ActionResult.Header`, yönlendirmeden önce yanıta yazılır; cookie de buraya konur.

`Secure: true` production için doğrudur. Chrome ve Firefox `http://localhost`
üzerinde de secure bir cookie'yi kabul eder; bu yüzden `collage dev` orada onunla
değişiklik gerekmeden çalışır. Safari kabul etmeyebilir ve düz `http` üzerinde
cookie'yi düşürür. Safari'de geliştiriyorsanız `Secure`'u isteğin TLS üzerinden gelip
gelmediğine göre ayarlayın ya da önizlemeyi başka bir tarayıcıda yapın.

### Önizlemeyi bitiren action

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

### Önizlemeye uyan middleware

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

Context değeri yalnızca `SkipCache` başarılı olduğunda ayarlanır. Hâlâ önbelleğe
alınabilir bir isteğe render edilmiş bir taslak, bütün bu düzenin önlemek için var
olduğu tek sonuçtur; bu yüzden ikisi ya birlikte olur ya hiç olmaz.

### Bağlamak

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

Bir data handler, istek bir önizlemeyse taslakları ister:

```go
func postData(ctx context.Context, rc *collage.RenderContext) (postView, []string, error) {
	slug := rc.Param("slug")
	tags := []string{"post:" + slug}

	post, err := collage.Cached(rc, "post:"+slug, time.Hour, tags,
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

`collage.Cached` çağrısı özel bir durum gerektirmez. Sıradan bir okuyucu için saklanan
yayımlanmış yazıyı döner; bir önizlemede ise — taslaklarla birlikte — veriyi çeker ve
hiçbir şey saklamaz; böylece taslak, bir sonraki okuyucunun aldığı değere sızamaz.

Görünümdeki `Preview`, şablonun `/api/preview/exit` bağlantısı içeren bir şerit
göstermesini sağlar. Bu tam da sayfa hiçbir zaman önbelleğe alınmadığı için güvenlidir:
şerit bir okuyucunun önüne çıkamaz.

## Bilinmesi gerekenler

**Bir CDN yine de önce yanıt verebilir.** `SkipCache` collage'ın önbelleklerini
kontrol eder ve önizleme yanıtını `no-store` olarak işaretler; ama CDN'in kendi
önbelleğinden yanıtladığı bir istek sunucunuza hiç ulaşmaz. `Static()` sayfalar
`max-age=0, must-revalidate` ile gönderilir, bu yüzden CDN her seferinde geri döner
ve denetler; bir `Incremental(ttl)` sayfa ise TTL'i boyunca CDN'den sunulabilir.
CDN'i, önizleme cookie'sini taşıyan istekler için kendi önbelleğini atlayacak şekilde
yapılandırın.

**CMS'in içinde önizleme.** CMS önizlemeyi kendi alan adındaki bir `<iframe>` içinde
gösteriyorsa, `SameSite=Lax` bir cookie çerçevelenen siteye gönderilmez. Bu durumda
`Secure: true` ile birlikte `SameSite: http.SameSiteNoneMode` kullanın — ve bunun yine
de yetmeyebileceğini bilin. O çerçevenin içinde cookie'niz üçüncü taraf bir cookie'dir;
Safari ve Firefox bunları varsayılan olarak engeller, bu yüzden CMS'in iframe'indeki bir
önizleme onu hiç almayabilir. Ayrıca `Partitioned: true` ayarlamak (CMS'in sitesine
göre bölümlenmiş bir CHIPS cookie'si) bölümlemeyi destekleyen tarayıcılarda onu
geçirir; her yerde çalışan düzen ise yukarıdakidir: önizlemeyi yeni bir sekmede açmak.

**Statik dışa aktarma hiçbir zaman taslak görmez.** Dışa aktarma bir istek olmadan
render eder; bu yüzden hiçbir middleware çalışmaz, `preview.Drafts` her sayfa için
false'tur ve yalnızca yayımlanmış içerik yazılır. Önizleme action'ları da dışa
aktarılmaz — bir action'ın sunucuya ihtiyacı vardır. Statik dosyalar olarak yayına
alınmış bir site, sunucuyu editörlerin erişebileceği bir yerde çalıştırarak yine de
önizleme sunabilir; bkz. [Statik dışa aktarma](/docs/static-export).

**Geçersiz kılma yine önemlidir.** Önizleme taslağı gösterir; siteyi değiştiren şey
onu yayımlamaktır. CMS yayımladığında, okuyucuların yeni sürümü alması için webhook'u
yazının etiketini geçersiz kılmalıdır. Bkz.
[Middleware ve kendi API'niz](/docs/middleware-and-apis#invalidating-from-your-api).
