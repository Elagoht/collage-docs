---
description: Binary'yi derleyin, bir container'da ya da systemd altında çalıştırın, production'ın gerektirdiklerini ayarlayın ve TLS'in arkasına koyun.
---

# Yayına alma

Bir collage sitesi bir Go programıdır; dolayısıyla yayına aldığınız şey derlenmiş bir
binary'dir. Şablonlar ve statik dosyalar onun içine gömülüdür, bu yüzden yanına hiçbir
şey kopyalanmadan her dizinden çalışır — plugin yapılandırıyorsanız
`plugins-config.json` hariç; bkz. [aşağısı](#plugin-configuration). Bu sayfa o
binary'yi bir sunucuda çalıştırmakla ilgilidir; hiç sunucusu olmayan bir site için
bkz. [Statik dışa aktarma](/docs/static-export).

## Derleme: `collage build`

```sh
collage build
./bin/mysite
```

`collage build`, aksi hâlde aklınızda tutmanız gereken `go build` komutunu sizin
yerinize çalıştırır:

```sh
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o bin/mysite .
```

- **CGO kapalı**, çünkü collage'ın ve standart kütüphanenin C'ye ihtiyacı yoktur ve
  statik bir binary, başka hiçbir şey içermeyen bir imaja konabilir.
- **`-trimpath`**, böylece binary kendisini derleyen makinenin yollarını taşımaz.
- **`-s -w`** debug tablolarını atar; boyutun çoğu onlardır.
- **Varsayılan olarak bu makine değil, linux/amd64.** Mac için derlenmiş bir binary
  bir Linux container'ında çalışmaz ve bunu öğrenmek için yanlış yer, sunucuda
  karşınıza çıkan bir `exec format error`'dır.

| Flag | Varsayılan | Anlamı |
| --- | --- | --- |
| `-o path` | `bin/<module name>` | Binary'nin yazılacağı yer. |
| `-os name` | `linux` | Hedef işletim sistemi. |
| `-arch name` | `amd64` | Hedef mimari — Graviton ya da Ampere makineler için `arm64`. |
| `-i` | kapalı | Binary'nin yanına bir Dockerfile ve bir systemd unit'i yazmayı önerir. |

Binary `dist/`'e değil, `bin/`'e gider: `dist/`, `collage export`'a aittir ve onun
`-clean`'i içinde duran bir binary'yi silerdi.

`collage start` diye bir şey yoktur. Derlenmiş bir binary'yi çalıştırmak için hiçbir
şeyi aklınızda tutmanız gerekmez; bir start komutu da ancak `go run .` çalıştırabilirdi
— bu da Go araç zincirini production imajınıza koyar ve her açılışta derleme yapar.

### `collage build -i`: bir Dockerfile ve bir systemd unit'i

`-i` ile `collage build`, binary'nin yanına bir `Dockerfile` ve bir systemd unit'i
yazıp yazmayacağını sorar:

```sh
$ collage build -i
Building mysite for linux/amd64.

  Write a Dockerfile? [y/N]: y
  Write a systemd unit? [y/N]: y

✓ bin/mysite
    linux/amd64 · 9.3 MB
    wrote bin/Dockerfile    docker build -f bin/Dockerfile .
    wrote bin/mysite.service

9.3 MB · linux/amd64 · 4.1s
```

Üretildikleri için `bin/`'e giderler; proje kökü sizin yazdıklarınız içindir. Var olan
bir dosyanın üzerine asla yazılmaz — bunlar düzenlemeniz beklenen dosyalardır — ve
iskelet olarak üretilen bir projenin `.gitignore`'u yalnızca binary'yi yok sayar;
bu yüzden onları diğer her dosya gibi commit'lersiniz.

Dockerfile iki aşamalıdır: Go imajında derler ve distroless bir taban üzerinde
yalnızca binary'yi taşır:

```dockerfile
FROM golang:1.26 AS build
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /mysite .

FROM gcr.io/distroless/static-debian12
WORKDIR /srv
COPY --from=build /mysite /usr/local/bin/mysite
ENV HOST=0.0.0.0 PORT=8080
# ENV COLLAGE_CSRF_KEY=
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/mysite"]
```

Onu proje kökünden `docker build -f bin/Dockerfile .` ile derleyin. Şablonlar ya da
statik dosyalar için bir `COPY` yoktur, çünkü onlar binary'nin içindedir. Son aşama
binary'yi kopyalar — ve v0.11.1'den itibaren, Dockerfile yazıldığı sırada projede
`plugins-config.json` varsa onu da; bkz. [Plugin yapılandırması](#plugin-configuration).
`WORKDIR`, sayfa önbelleği ve o dosya için önemlidir; aşağıya bakın.
`COLLAGE_CSRF_KEY`'i dosyada değil, platformunuzun gizli değerlerinden (secret)
ayarlayın.

systemd unit'i binary'yi `/usr/local/bin`'den, `/srv/<name>` içinde, aynı adı taşıyan
bir kullanıcı olarak çalıştırır ve önündeki bir reverse proxy için `127.0.0.1:8080`'i
dinler. `/etc/systemd/system/<name>.service` konumuna kurmadan önce kullanıcıyı,
dizini ve yolu düzenleyin:

```sh
systemctl daemon-reload && systemctl enable --now mysite
```

## Ortam

İskelet olarak üretilen `main.go` dört değişken okur. Production'da hiçbir şey `.env`
dosyalarını okumaz — onlar `collage dev` içindir — bu yüzden bunları binary nerede
çalışıyorsa orada ayarlayın.

| Değişken | Varsayılan | Ne olmalı |
| --- | --- | --- |
| `HOST` | `localhost` | Bir container'da `0.0.0.0`. `localhost` makinenin dışından hiçbir şey kabul etmez; bu, yerel bir reverse proxy'nin arkasında doğru, diğer her yerde yanlıştır. |
| `PORT` | `3000` | Platformunuz ne atıyorsa o. Binary'nin `-port` flag'i bunu geçersiz kılar. |
| `COLLAGE_CSRF_KEY` | üretilir | En az 32 rastgele bayt. **Bunu ayarlayın.** |
| `COLLAGE_DEV` | ayarlanmamış | **Hiçbir şey — ayarlamayın.** `collage dev` onu `1` yapar ve bu geliştirme modunu açar: şablonlar ve statik dosyalar diskten okunur, hiç okunmayan bellek içi bir önbellek kullanılır, hata sayfalarında hata zincirlerinin tamamı gösterilir. Bununla çalışan bir sunucu bir geliştirme sunucusudur. |

Bir anahtarı şöyle üretin:

```sh
openssl rand -hex 32
```

Anahtar yoksa collage her süreç için bir anahtar üretir ve başlangıçta uyarır. Bu
kendi makinenizde sorun değildir, bir sunucuda ise iki nedenle yanlıştır:

- **Formlar yeniden başlatmalar ve instance'lar arasında bozulur.** Bir yayına
  almadan önce render edilmiş bir form, sonrasında gönderildiğinde reddedilir; bir
  instance da diğerinin verdiğini reddeder. Anahtarı her instance'ta ve yeniden
  başlatmalar boyunca aynı tutun.
- **Form içeren önbellekteki sayfalar her başlatmadan sonra yeniden render edilir.**
  Önbellekteki bir sayfadaki sahtecilik token'ı yer tutucusu anahtardan türetilir;
  bu yüzden önceki sürecin anahtarıyla saklanmış bir sayfa ıskalama (miss) sayılır ve
  baştan render edilir. Form içermeyen sayfalar etkilenmez; disk önbelleğinin geri
  kalanı her durumda yeniden başlatmadan sağ çıkar. Bkz.
  [Önbellekleme](/docs/caching#the-namespace).

Anahtarı değiştirmek güvenlidir ama bedelsiz değildir: değişiklikten önce render
edilmiş formlar reddedilir ve form içeren önbellekteki sayfalar birer kez yeniden
render edilir.

## Gömülü şablonlar ve statik dosyalar

İskelet olarak üretilen bir proje ikisini de gömer:

```go
//go:embed all:templates
var templatesFS embed.FS

//go:embed all:static
var staticFS embed.FS
```

Şablonlar collage'a `Template.FS` olarak verilir, statik dosyalar ise
`fs.Sub(staticFS, "static")`'tan mount edilir. Geliştirme modunda diskteki dizinler
önceliklidir; böylece bir şablonu düzenlediğinizde değişiklik bir sonraki istekte
görünür. Production'da yalnızca gömülü kopya okunur; böylece süreç nerede başlarsa
başlasın, çalışan şey test ettiğiniz şeydir.

Sayfalarınızın çalışma zamanında okuduğu içeriği — Markdown dosyaları, bir JSON
kataloğu — aynı şekilde gömmek size kalmıştır. collage-docs, tam da bu nedenle kendi
`content/` dizinini gömer ve geliştirmede onu diskten okur. Böyle bir dizini
[`Config.DevWatch`](/docs/configuration#devwatch) içinde adlandırın (v0.10.0'dan
itibaren); içindeki bir dosyayı düzenlemek de tarayıcıyı yeniler.

## Plugin yapılandırması

İskelet olarak üretilen `main.go`, `plugins-config.json`'ı
`collage.LoadPluginConfig("plugins-config.json")` ile okur — binary'ye değil, çalışma
dizinine göre bir yol; gömülü de değil. Eksik bir dosya hata değildir: her plugin
sessizce varsayılanlarıyla çalışır. Yani dosyanın bulunmadığı bir yerde başlatılan
bir sunucu, plugin ayarlarınızı tek kelime etmeden yok sayar.

`collage build -i`'nin yazdığı Dockerfile, Dockerfile yazıldığı sırada projede bu
dosya varsa onu kopyalar (v0.11.1'den itibaren; öncekiler yalnızca binary'yi
kopyalıyordu). Dosyayı sonradan oluşturursanız satırı kendiniz ekleyin —
`collage build -i` bir Dockerfile'ın üzerine asla yazmaz — sürecin başladığı
`WORKDIR`'in yanına:

```dockerfile
WORKDIR /srv
COPY --from=build /mysite /usr/local/bin/mysite
COPY --from=build /src/plugins-config.json /srv/plugins-config.json
```

ya da onu gömün; böylece şablonlar gibi binary'nin içinde taşınır:

```go
//go:embed plugins-config.json
var pluginConfigJSON []byte

var pluginConfig map[string]json.RawMessage
if err := json.Unmarshal(pluginConfigJSON, &pluginConfig); err != nil {
	return nil, fmt.Errorf("plugin configuration: %w", err)
}
```

systemd ile dosyayı unit'in `WorkingDirectory`'sinde tutun.

## Düzgün kapanış

`app.ListenAndServe`, `SIGINT` ve `SIGTERM`'ü yakalar. İkisinden birinde yeni
bağlantı kabul etmeyi bırakır, süren isteklerin bitmesi için `Server.ShutdownTimeout`
kadar (varsayılan 10 saniye) bekler, her plugin'in `Shutdown`'ını çalıştırır ve `nil`
döner. `SIGTERM` gönderip bekleyen bir container çalışma ortamı, hiçbir şey eklemeden
temiz bir boşaltma elde eder. Üretilen systemd unit'i `TimeoutStopSec=30` ayarlar; bu,
kapanış zaman aşımından rahatça uzundur, böylece systemd hâlâ boşaltmakta olan bir
süreci öldürmez.

Boşaltmanın süresi dolarsa — zaman aşımı geçtiğinde hâlâ açık bir istek varsa —
plugin'ler yine de kapatılır ve `ListenAndServe` bunu söyleyen bir hata döndürür
(`collage: server shutdown: context deadline exceeded`); `Shutdown`'ı başarısız olan
bir plugin için de aynısını yapar. İskeletin `main.go`'su bunu `log.Fatalf`'e verir;
bu yüzden süreç 0 yerine 1 durum koduyla çıkar. Durdurma sırasında sıfırdan farklı bir
çıkışı çökme olarak değerlendiren bir platform bunu böyle bildirir.

`ShutdownTimeout`'u yükseltirseniz, platformunuzun bekleme süresini de onunla birlikte
yükseltin.

## Sağlık denetimleri

`collage new`'in iskelet olarak ürettiği proje, `/healthz`'ye `status`'ü `ok` olan
küçük bir JSON gövdesiyle yanıt verir. Bu bir sayfa değil, bir
[document](/docs/documents)'tır; dolayısıyla hiçbir şablon içermez ve bir şablon
bozuldu diye başarısız olmaya başlayamaz. `Dynamic()`'tir; böylece her denetim
gerçekten sürece ulaşır.

Platformunuzun liveness denetimini ona yönlendirin. Size sürecin ayakta olduğunu ve
sunum yaptığını söyler. Veritabanınıza ulaşılamadığında da başarısız olması gereken
bir readiness denetimi, aynı şekilde yazılmış kendi document'ınızdır; bir hata
döndürün, 500 ile yanıt verir. `collage new --template minimal` ile oluşturulan bir
projede `/healthz` yoktur; bir tane istiyorsanız demo iskeletinin
`documents/health.go`'sunu kopyalayın.

## Production'da sayfa önbelleği

İskelet bir disk önbelleği yapılandırır:

```go
Cache: collage.CacheConfig{
	Enabled:    true,
	Type:       "disk",
	Dir:        cacheDir, // ".cache"
	DefaultTTL: 5 * time.Minute,
},
```

Render edilmiş sayfalar yeniden başlatmadan sağ çıkar; böylece aynı build'in yeniden
yayına alınması, siteyi soğuk bir önbelleğe baştan render etmez. Yeni bir build'in ne
bulacağı, neyin değiştiğine bağlıdır:

- **Önbellek, binary'nin bir hash'iyle ad alanına ayrılır.** Yeni bir build farklı bir
  dizini, `.cache/<hash>`'i okur; bu yüzden önceki build'in render ettiği sayfaları
  asla sunmaz. Doğruluk için hiçbir şeyin temizlenmesi gerekmez — ama yer açmak için
  de hiçbir şey temizlemez: önceki build'lerin dizinleri asla silinmez; bu yüzden yeni
  bir container olarak değil de yerinde yeniden yayına alınan bir sunucu, her build
  için bir tane biriktirir. Eskilerini yayına alma betiğinizden silin. Sayfaların
  nasıl göründüğüne binary dışındaki bir şey karar veriyorsa `Cache.Version`'ı — bir
  commit, bir sürüm etiketi — ayarlayın.
- **Dizin, çalışma dizinine göredir.** İskelet olarak üretilen bir projede, nerede
  başlatıldığına bağlı olan tek şey budur. Container'da `WORKDIR`'i ya da unit'te
  `WorkingDirectory`'yi ayarlayın veya `Dir`'e mutlak bir yol verin ve sürecin oraya
  yazabildiğinden emin olun.
- **Yazamadığı bir önbellek hata değildir.** Başlangıçta oluşturulamayan bir dizin —
  salt okunur bir dosya sistemi, sürecin yazamayacağı bir çalışma dizini — collage'ın
  başlamayı reddetmek yerine bir uyarıyla bellek içi bir önbelleğe geri dönmesine yol
  açar (v0.11.0'dan itibaren). Çalışırken başarısız olan bir yazma loglanır, sayfa
  önbelleğe alınmadan sunulur ve bir sonraki istek onu yeniden render eder: yavaş,
  ama bozuk değil. Uyarıya bakın ve bir yayına almadan sonra dizinin dolduğunu kontrol
  edin.
- **Her instance'ın kendi önbelleği vardır.** Bir container'da dizin container'ın
  içindedir; bu yüzden her instance kendi önbelleğini, instance başına sayfa başına
  bir render ile doldurur.

### Birden çok instance ile geçersiz kılma

`app.InvalidateTags`, içinde çalıştığı sürecin önbelleğinden girdileri düşürür. Tek
bir instance varken, bir yazının etiketini geçersiz kılan bir CMS webhook'u siteyi
günceller. Birkaç instance varken ise yalnızca webhook'un ulaştığı instance'ı
günceller; diğerleri eski sayfayı TTL'i dolana kadar sunmaya devam eder.

Üç yanıttan birini seçin: değişen sayfalara `Incremental(ttl)` verin ki bayat bir
kopya kendiliğinden sona ersin, webhook'u her instance'a gönderin ya da
`Cache.Store` üzerinden `collage.TaggedCache`'i uygulayan paylaşımlı bir depolama
yapılandırın; bu, tek bir geçersiz kılmanın herhangi bir instance'ın yazdığı
girdilere ulaşmasını sağlar. Bkz. [Önbellekleme](/docs/caching).

Paylaşımlı bir depolama veriyi değil, sayfaları tutar.
[`collage.Cached`](/docs/caching#caching-data-across-pages) ile tutulan değerler,
`Cache.Store` ne olursa olsun her sürecin belleğinde yaşar ve bir geçersiz kılma
yalnızca içinde çalıştığı instance'a ulaşır. Bir webhook A instance'ına ulaştıktan
sonra B instance'ı, hâlâ tuttuğu eski değerden — paylaşımlı depolama sayfayı
düşürdüğü için — taze bir sayfa render eder. Bu değerlere katlanabileceğiniz kadar
kısa bir TTL verin ya da geçersiz kılmayı her instance'a gönderin.

## TLS, bir proxy'nin arkasında

collage düz HTTP sunar ve `ListenAndServeTLS`'i yoktur. TLS'i iyi sonlandırmak;
sertifikalar, yenileme, HTTP/2, HSTS ve 80 numaralı porttan bir yönlendirme demektir
ve bunun üzerinde çalıştığı her platform — bir load balancer, bir reverse proxy,
Cloudflare, Fly, Render — bunu bir framework flag'inin yapabileceğinden zaten daha iyi
yapar.

Binary'yi bunlardan birinin arkasına koyun ve proxy'nin `X-Forwarded-Proto: https`
gönderdiğinden emin olun. collage bunu sahtecilik token'ı cookie'sini `Secure` olarak
işaretlemek için kullanır; böylece HTTPS üzerinden sunulan bir site o cookie'yi asla
düz HTTP üzerinden göndermez. Sertifikayı da alan minimal bir Caddy yapılandırması:

```
example.com {
	reverse_proxy 127.0.0.1:8080
}
```

nginx ile `location` bloğundaki `proxy_set_header X-Forwarded-Proto $scheme;` aynı
işi görür.

TLS'i binary'nin kendisinin sonlandırmasını istiyorsanız, `app.Handler()` sıradan bir
`http.Handler`'dır:

```go
srv := &http.Server{
	Addr:              ":443",
	Handler:           app.Handler(),
	ReadHeaderTimeout: 15 * time.Second,
}
log.Fatal(srv.ListenAndServeTLS(certFile, keyFile))
```

Bunu yapmak, `ListenAndServe`'ün sizin için yaptığı zaman aşımlarını ve sinyal
işlemeyi artık sizin üstlendiğiniz anlamına gelir; plugin'lerin kapanması için
`app.Shutdown(ctx)`'i de kendiniz çağırırsınız.

## Zaman aşımları

`ListenAndServe`, `Config.Server` içindeki zaman aşımlarını uygular:

| Alan | Varsayılan | Sınırladığı |
| --- | --- | --- |
| `ReadTimeout` | 15s | Header'lar dahil bir isteğin okunması — header'ları yavaş gönderen bir istemci bir bağlantıyı açık tutamaz. |
| `WriteTimeout` | 30s | Yanıtın yazılması. Verisi bundan uzun süren bir sayfa yarıda kesilir. |
| `IdleTimeout` | 60s | Bir sonraki isteğini bekleyen bir keep-alive bağlantısı. |
| `ShutdownTimeout` | 10s | `SIGTERM` üzerine yapılan düzgün boşaltma. |
| `MaxBodyBytes` | 4 MiB | Action kendi sınırını belirlemedikçe, bir action'ın istek gövdesi. Negatif değer sınırsızdır. |

```go
Server: collage.ServerConfig{
	Host:         envString("HOST", "localhost"),
	Port:         port,
	WriteTimeout: time.Minute,
},
```

Bir data handler'ın ne kadar sürebileceği ayrı bir ayardır: her fragment'in
`WithTimeout`'u ya da hiçbir şey ayarlamayan fragment'ler ve document'lar için
`Template.Timeout` (5 saniye). Onu `WriteTimeout`'un epey altında tutun; böylece
yavaş bir dış servis, sayfanın ortasında kapanan bir bağlantıya değil, bir fragment'in
yedeğine dönüşür. Bkz. [Yapılandırma](/docs/configuration).

## Loglar

`Config.Logger` yoksa collage, `slog`'un varsayılan handler'ına — ya da bir
terminaldeyse insanlar için biçimlendirilmiş bir handler'a — log yazar. Logların bir
makine tarafından okunduğu bir sunucuda bir JSON handler'ı verin:

```go
Logger: slog.New(slog.NewJSONHandler(os.Stdout, nil)),
```

## Kontrol listesi

- `COLLAGE_CSRF_KEY` ayarlı ve her instance'ta aynı.
- Bir container'da `HOST=0.0.0.0`; `PORT` platformdan.
- `.cache` için yazılabilir bir çalışma dizini ve yayına almada silinen eski
  `.cache/<hash>` dizinleri.
- Plugin yapılandırıyorsanız, çalışma dizininde ya da gömülü bir
  `plugins-config.json`.
- `COLLAGE_DEV` ayarlanmamış.
- Proxy'de TLS ve iletilen `X-Forwarded-Proto`.
- Platformun durdurma bekleme süresi `Server.ShutdownTimeout`'tan uzun.
- `/healthz` üzerinde liveness denetimi.
- Build'den önce CI'da `go test ./...` — bkz. [Test](/docs/testing).
