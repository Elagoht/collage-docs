---
description: Binary'yi build edin, bir container'da ya da systemd altında çalıştırın, production'ın ihtiyaç duyduklarını ayarlayın ve TLS'in arkasına koyun.
reference: ServerConfig, CacheConfig, LoadPluginConfig
---

# Deployment

Bir collage sitesi bir Go programıdır. Bu yüzden deploy ettiğiniz şey derlenmiş bir
binary'dir. Template'ler ve static dosyalar binary'nin içine gömülüdür. Binary,
yanına hiçbir şey kopyalamadan herhangi bir dizinden çalışır. Tek istisna, plugin'leri
yapılandırıyorsanız `plugins-config.json` dosyasıdır; [aşağıya](#plugin-configuration)
bakın. Bu sayfa o binary'yi bir sunucuda çalıştırmayı anlatır. Hiç sunucusu olmayan
bir site için [Static export](/docs/static-export) sayfasına bakın.

## Build: `collage build`

```sh
collage build
./bin/mysite
```

`collage build`, normalde aklınızda tutmanız gereken `go build` komutunu sizin
yerinize çalıştırır:

```sh
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o bin/mysite .
```

- **CGO kapalıdır**, çünkü ne collage ne de standart kütüphane C'ye ihtiyaç duyar.
  Static bir binary, içinde başka hiçbir şey olmayan bir image'a konabilir.
- **`-trimpath`** sayesinde binary, kendisini build eden makinenin path'lerini
  taşımaz.
- **`-s -w`** debug tablolarını atar. Boyutun büyük kısmı bu tablolardır.
- **Varsayılan hedef bu makine değil, linux/amd64'tür.** Mac için build edilmiş bir
  binary Linux container'ında çalışmaz. Bunu sunucuda bir `exec format error` ile
  öğrenmek istemezsiniz.

| Flag | Varsayılan | Anlamı |
| --- | --- | --- |
| `-o path` | `bin/<module name>` | Binary'nin yazılacağı yer. |
| `-os name` | `linux` | Hedef işletim sistemi. |
| `-arch name` | `amd64` | Hedef mimari. Graviton ya da Ampere makineler için `arm64` kullanın. |
| `-i` | kapalı | Binary'nin yanına bir Dockerfile ve bir systemd unit'i yazmayı önerir. |

Binary `dist/`'e değil, `bin/`'e yazılır. `dist/` dizini `collage export`'a aittir
ve export'un `-clean` seçeneği orada duran bir binary'yi silerdi.

`collage start` diye bir komut yoktur. Derlenmiş bir binary'yi çalıştırmak için
hiçbir şeyi aklınızda tutmanız gerekmez. Bir start komutu olsaydı yapabileceği tek şey
`go run .` çalıştırmak olurdu. Bu da Go toolchain'ini production image'ınıza koyar ve
her açılışta yeniden derleme yapar.

### `collage build -i`: bir Dockerfile ve bir systemd unit'i

`-i` verildiğinde `collage build`, binary'nin yanına bir `Dockerfile` ve bir systemd
unit'i yazıp yazmayacağını sorar:

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

Bu dosyalar üretildikleri için `bin/`'e yazılır. Proje kökü sizin yazdığınız dosyalar
içindir. Var olan bir dosyanın üzerine asla yazılmaz, çünkü bunlar düzenlemeniz
beklenen dosyalardır. Scaffold edilen bir projenin `.gitignore`'u yalnızca binary'yi
yok sayar. Bu dosyaları da diğer dosyalar gibi commit edersiniz.

Dockerfile iki stage'den oluşur. Önce Go image'ında build eder, sonra distroless bir
base image üzerinde yalnızca binary'yi taşır:

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

Image'ı proje kökünden `docker build -f bin/Dockerfile .` ile build edin.
Template'ler ya da static dosyalar için bir `COPY` satırı yoktur, çünkü onlar
binary'nin içindedir. Son stage binary'yi kopyalar. v0.11.1'den itibaren, Dockerfile
yazıldığı sırada projede `plugins-config.json` varsa onu da kopyalar;
[Plugin config'i](#plugin-configuration) bölümüne bakın. `WORKDIR`, page cache ve bu
dosya için önemlidir; ayrıntıları aşağıda bulabilirsiniz. `COLLAGE_CSRF_KEY`'i dosyanın içinde değil,
platformunuzun secret'larından ayarlayın.

systemd unit'i binary'yi `/usr/local/bin`'den, `/srv/<name>` dizininde ve aynı adı
taşıyan bir kullanıcıyla çalıştırır. Binary, önüne konacak bir reverse proxy için
`127.0.0.1:8080`'i dinler. Unit'i `/etc/systemd/system/<name>.service` konumuna
kurmadan önce kullanıcıyı, dizini ve path'i kendinize göre düzenleyin:

```sh
systemctl daemon-reload && systemctl enable --now mysite
```

## Ortam değişkenleri

Scaffold edilen `main.go` dört değişken okur. Production'da `.env` dosyalarını okuyan
hiçbir şey yoktur, çünkü onlar `collage dev` içindir. Bu değişkenleri binary nerede
çalışıyorsa orada ayarlayın.

| Değişken | Varsayılan | Değeri |
| --- | --- | --- |
| `HOST` | `localhost` | Container'da `0.0.0.0`. `localhost` makinenin dışından gelen hiçbir bağlantıyı kabul etmez. Bu, yerel bir reverse proxy'nin arkasında doğrudur, başka her yerde yanlıştır. |
| `PORT` | `3000` | Platformunuz hangi port'u atıyorsa o. Binary'nin `-port` flag'i bu değeri ezer. |
| `COLLAGE_CSRF_KEY` | üretilir | En az 32 rastgele byte. **Bunu mutlaka ayarlayın.** |
| `COLLAGE_DEV` | ayarlı değil | **Hiçbir şey; ayarlamadan bırakın.** `collage dev` bu değişkeni `1` yapar ve development mode'u açar. Bu modda template'ler ve static dosyalar diskten okunur, cache bellekte tutulur ama hiç okunmaz, error page'lerde hata zincirinin tamamı gösterilir. Bu değişkenle çalışan bir sunucu bir development sunucusudur. |

Key'i şu komutla üretin:

```sh
openssl rand -hex 32
```

Key verilmezse collage her process için bir key üretir ve açılışta uyarı verir. Bu
kendi makinenizde sorun değildir. Bir sunucuda ise iki nedenle yanlıştır:

- **Form'lar restart'lar ve instance'lar arasında bozulur.** Deploy'dan önce render
  edilmiş bir form, deploy'dan sonra gönderildiğinde reddedilir. Bir instance da
  başka bir instance'ın verdiği token'ı reddeder. Key'i her instance'ta aynı tutun ve
  restart'lar arasında değiştirmeyin.
- **Form içeren cache'lenmiş page'ler her açılıştan sonra yeniden render edilir.**
  Cache'lenmiş bir page'deki forgery token placeholder'ı key'den türetilir. Bu yüzden
  önceki process'in key'iyle saklanmış bir page miss sayılır ve baştan render edilir.
  Form içermeyen page'ler bundan etkilenmez. Disk cache'inin geri kalanı her durumda
  restart'tan sağ çıkar. [Caching](/docs/caching#the-namespace) sayfasına bakın.

Key'i değiştirmek güvenlidir ama bedelsiz değildir. Değişiklikten önce render edilmiş
form'lar reddedilir ve form içeren cache'lenmiş page'ler birer kez yeniden render
edilir.

## Gömülü template'ler ve static dosyalar

Scaffold edilen bir proje ikisini de gömer:

```go
//go:embed all:templates
var templatesFS embed.FS

//go:embed all:static
var staticFS embed.FS
```

Template'ler collage'a `Template.FS` olarak verilir. Static dosyalar ise
`fs.Sub(staticFS, "static")` üzerinden mount edilir. Development mode'da diskteki
dizinler önceliklidir. Böylece bir template'i düzenlediğinizde değişiklik bir sonraki
request'te görünür. Production'da yalnızca gömülü kopya okunur. Böylece process nerede
başlatılırsa başlatılsın, çalışan şey test ettiğiniz şeydir.

Page'lerinizin runtime'da okuduğu içerikleri, örneğin Markdown dosyalarını ya da bir
JSON kataloğunu, aynı şekilde gömmek size kalmıştır. collage-docs da tam bu nedenle
kendi `content/` dizinini gömer ve development'ta bu dizini diskten okur. Böyle bir
dizini [`Config.DevWatch`](/docs/configuration#devwatch) içinde belirtirseniz
(v0.10.0'dan itibaren), içindeki bir dosyayı düzenlediğinizde tarayıcı da yenilenir.

## Plugin config'i

Scaffold edilen `main.go`, `plugins-config.json` dosyasını
`collage.LoadPluginConfig("plugins-config.json")` ile okur. Bu path binary'ye göre
değil, çalışma dizinine göredir ve dosya binary'ye gömülmez. Dosyanın bulunmaması bir
hata değildir. Bu durumda her plugin sessizce kendi varsayılanlarıyla çalışır. Yani
dosyanın olmadığı bir yerde başlatılan sunucu, plugin ayarlarınızı hiçbir uyarı
vermeden yok sayar.

`collage build -i`'nin yazdığı Dockerfile, yazıldığı sırada projede bu dosya varsa onu
da kopyalar (v0.11.1'den itibaren; önceki sürümler yalnızca binary'yi kopyalıyordu).
Dosyayı sonradan oluşturursanız satırı kendiniz ekleyin, çünkü `collage build -i` var
olan bir Dockerfile'ın üzerine asla yazmaz. Dosyayı process'in başladığı `WORKDIR`'e
koyun:

```dockerfile
WORKDIR /srv
COPY --from=build /mysite /usr/local/bin/mysite
COPY --from=build /src/plugins-config.json /srv/plugins-config.json
```

Ya da dosyayı gömün. Böylece dosya, template'ler gibi binary'nin içinde taşınır:

```go
//go:embed plugins-config.json
var pluginConfigJSON []byte

var pluginConfig map[string]json.RawMessage
if err := json.Unmarshal(pluginConfigJSON, &pluginConfig); err != nil {
	return nil, fmt.Errorf("plugin configuration: %w", err)
}
```

systemd kullanıyorsanız dosyayı unit'in `WorkingDirectory`'sinde tutun.

## Graceful shutdown

`app.ListenAndServe`, `SIGINT` ve `SIGTERM` sinyallerini yakalar. Bunlardan biri
geldiğinde yeni bağlantı kabul etmeyi bırakır. Devam eden request'lerin bitmesi için
en fazla `Server.ShutdownTimeout` kadar (varsayılan 10 saniye) bekler. Ardından her
plugin'in `Shutdown`'ını çalıştırır ve `nil` döner. `SIGTERM` gönderip bekleyen bir
container runtime, sizin hiçbir şey eklemenize gerek kalmadan temiz bir drain elde
eder. Üretilen systemd unit'i `TimeoutStopSec=30` ayarlar. Bu süre shutdown
timeout'undan rahatça uzundur, böylece systemd hâlâ drain eden bir process'i
öldürmez.

Drain için süre yetmezse, yani timeout dolduğunda hâlâ açık bir request varsa,
plugin'ler yine de kapatılır. `ListenAndServe` de bunu bildiren bir hata döner
(`collage: server shutdown: context deadline exceeded`). `Shutdown`'ı başarısız olan
bir plugin için de aynı şekilde hata döner. Scaffold'daki `main.go` bu hatayı
`log.Fatalf`'e verir. Bu yüzden process 0 yerine 1 status koduyla çıkar. Durdurma
sırasında sıfırdan farklı bir exit kodunu crash olarak gören bir platform bunu crash
olarak raporlar.

`ShutdownTimeout`'u artırırsanız platformunuzun grace period'unu da onunla birlikte
artırın.

## Health check'ler

`collage new` ile scaffold edilen proje, `/healthz` isteğine `status` alanı `ok` olan
küçük bir JSON body ile cevap verir. Bu bir page değil, bir
[document](/docs/documents)'tır. Bu yüzden hiçbir template kullanmaz ve bir template
bozuldu diye başarısız olmaya başlamaz. `Dynamic()` olduğu için her check gerçekten
process'e ulaşır.

Platformunuzun liveness check'ini bu adrese yönlendirin. Bu check size process'in
ayakta olduğunu ve request'lere cevap verdiğini söyler. Veritabanınıza
ulaşılamadığında da başarısız olması gereken bir readiness check istiyorsanız, onu aynı
şekilde kendi document'ınız olarak yazarsınız. Document bir hata dönerse 500 ile
cevap verir. `collage new --template minimal` ile oluşturulan projede `/healthz`
yoktur. İsterseniz demo scaffold'undaki `documents/health.go` dosyasını kopyalayın.

## Production'da page cache

Scaffold bir disk cache'i yapılandırır:

```go
Cache: collage.CacheConfig{
	Enabled:    true,
	Type:       "disk",
	Dir:        cacheDir, // ".cache"
	DefaultTTL: 5 * time.Minute,
},
```

Render edilmiş page'ler restart'tan sağ çıkar. Böylece aynı build'i yeniden deploy
ettiğinizde site boş bir cache'e baştan render edilmez. Yeni bir build'in ne
bulacağı, neyin değiştiğine bağlıdır:

- **Cache, binary'nin hash'iyle namespace'lenir.** Yeni bir build farklı bir dizini,
  `.cache/<hash>`'i okur. Bu yüzden önceki build'in render ettiği page'leri asla
  sunmaz. Doğruluk için hiçbir şeyi temizlemeniz gerekmez. Ama yer açmak için de
  hiçbir şey temizlik yapmaz: önceki build'lerin dizinleri hiç silinmez. Yeni bir
  container yerine yerinde yeniden deploy edilen bir sunucuda her build için bir dizin
  birikir. Eski dizinleri deploy script'inizden silin. Page'lerin nasıl görüneceğini
  binary dışındaki bir şey belirliyorsa `Cache.Version`'ı bir commit ya da release
  tag'i ile ayarlayın.
- **Dizin, çalışma dizinine göredir.** Scaffold edilen bir projede nerede
  başlatıldığına bağlı olan tek şey budur. Container'da `WORKDIR`'i ya da unit'te
  `WorkingDirectory`'yi ayarlayın veya `Dir`'e mutlak bir path verin. Process'in
  oraya yazabildiğinden de emin olun.
- **Cache'in yazamaması bir hata değildir.** Açılışta dizin oluşturulamazsa collage
  başlamayı reddetmez; bir uyarı verip in-memory cache'e geçer (v0.11.0'dan itibaren).
  Buna read-only bir dosya sistemi ya da process'in yazma izni olmayan bir çalışma
  dizini yol açabilir. Çalışırken başarısız olan bir yazma log'lanır, page cache'lenmeden
  sunulur ve bir sonraki request onu yeniden render eder. Sonuç yavaştır ama bozuk
  değildir. Bu uyarıyı takip edin ve deploy'dan sonra dizinin dolduğunu kontrol edin.
- **Her instance'ın kendi cache'i vardır.** Container'da dizin container'ın
  içindedir. Bu yüzden her instance kendi cache'ini doldurur. Her page, her instance'ta
  bir kez render edilir.

### Birden fazla instance ile invalidation

`app.InvalidateTags`, içinde çalıştığı process'in cache'inden entry'leri düşürür. Tek
instance varken, bir yazının tag'ini invalidate eden bir CMS webhook'u siteyi
günceller. Birden fazla instance varsa yalnızca webhook'un ulaştığı instance
güncellenir. Diğerleri eski page'i TTL'i dolana kadar sunmaya devam eder.

Üç çözümden birini seçin. Değişen page'lere `Incremental(ttl)` verin, böylece stale
kopya kendiliğinden expire olur. Ya da webhook'u her instance'a gönderin. Ya da
`Cache.Store` üzerinden `collage.TaggedCache`'i implement eden paylaşımlı bir store
yapılandırın. Böylece tek bir invalidation, herhangi bir instance'ın yazdığı
entry'lere ulaşır. [Caching](/docs/caching) sayfasına bakın.

Paylaşımlı bir store veriyi değil, page'leri tutar.
[`collage.Cached`](/docs/caching#caching-data-across-pages) ile tutulan değerler,
`Cache.Store` ne olursa olsun her process'in kendi belleğinde yaşar. Bir invalidation
da yalnızca içinde çalıştığı instance'a ulaşır. Webhook A instance'ına ulaştıktan
sonra, paylaşımlı store page'i düşürdüğü için B instance'ı page'i yeniden render eder.
Ama bunu hâlâ elinde tuttuğu eski değerle yapar. Bu değerlere katlanabileceğiniz
kadar kısa bir TTL verin ya da invalidation'ı her instance'a gönderin.

## TLS, bir proxy'nin arkasında

collage düz HTTP sunar ve `ListenAndServeTLS` sağlamaz. TLS'i düzgün terminate etmek
sertifikalar, yenileme, HTTP/2, HSTS ve 80 numaralı port'tan redirect demektir. collage'ın
çalıştığı her platform bunu zaten bir framework flag'inin yapabileceğinden daha iyi
yapar: load balancer, reverse proxy, Cloudflare, Fly, Render.

Binary'yi bunlardan birinin arkasına koyun ve proxy'nin `X-Forwarded-Proto: https`
gönderdiğinden emin olun. collage bu header'a bakarak forgery token cookie'sini
`Secure` olarak işaretler. Böylece HTTPS üzerinden sunulan bir site bu cookie'yi asla
düz HTTP üzerinden göndermez. Sertifikayı da kendisi alan minimal bir Caddy config'i
şöyledir:

```
example.com {
	reverse_proxy 127.0.0.1:8080
}
```

nginx'te `location` bloğuna eklenen `proxy_set_header X-Forwarded-Proto $scheme;`
aynı işi görür.

TLS'i binary'nin kendisinin terminate etmesini istiyorsanız, `app.Handler()` sıradan
bir `http.Handler`'dır:

```go
srv := &http.Server{
	Addr:              ":443",
	Handler:           app.Handler(),
	ReadHeaderTimeout: 15 * time.Second,
}
log.Fatal(srv.ListenAndServeTLS(certFile, keyFile))
```

Bu durumda `ListenAndServe`'ün sizin yerinize hallettiği timeout'lar ve signal
handling artık sizin sorumluluğunuzdadır. Plugin'lerin kapanması için
`app.Shutdown(ctx)`'i de kendiniz çağırırsınız.

## Timeout'lar

`ListenAndServe`, `Config.Server` içindeki timeout'ları uygular:

| Alan | Varsayılan | Neyi sınırlar |
| --- | --- | --- |
| `ReadTimeout` | 15s | Header'lar dahil request'in okunmasını. Header'ları yavaş gönderen bir client bağlantıyı açık tutamaz. |
| `WriteTimeout` | 30s | Response'un yazılmasını. Verisi bundan uzun süren bir page yarıda kesilir. |
| `IdleTimeout` | 60s | Bir sonraki request'ini bekleyen keep-alive bağlantısını. |
| `ShutdownTimeout` | 10s | `SIGTERM` geldiğinde yapılan graceful drain'i. |
| `MaxBodyBytes` | 4 MiB | Bir action'ın request body'sini, action kendi sınırını belirlemediyse. Negatif değer sınırsız demektir. |

```go
Server: collage.ServerConfig{
	Host:         envString("HOST", "localhost"),
	Port:         port,
	WriteTimeout: time.Minute,
},
```

Bir data handler'ın ne kadar sürebileceği ayrı bir ayardır. Bunu her fragment'in
`WithTimeout`'u belirler. Timeout belirtmeyen fragment'ler ve document'lar için
`Template.Timeout` (5 saniye) geçerlidir. Bu süreyi `WriteTimeout`'un epey altında
tutun. Böylece yavaş bir upstream, page'in ortasında kapanan bir bağlantıya değil,
fragment'in fallback'ine dönüşür. [Config](/docs/configuration) sayfasına
bakın.

## Loglar

`Config.Logger` verilmezse collage, `slog`'un varsayılan handler'ına log yazar.
Terminalde çalışıyorsa insanların okuması için formatlanmış bir handler kullanır. Bir
sunucuda loglar bir makine tarafından okunur. Orada bir JSON handler verin:

```go
Logger: slog.New(slog.NewJSONHandler(os.Stdout, nil)),
```

## Kontrol listesi

- `COLLAGE_CSRF_KEY` ayarlı ve her instance'ta aynı.
- Container'da `HOST=0.0.0.0`, `PORT` platformdan geliyor.
- `.cache` için yazılabilir bir çalışma dizini var ve eski `.cache/<hash>` dizinleri
  deploy sırasında siliniyor.
- Plugin'leri yapılandırıyorsanız `plugins-config.json` çalışma dizininde duruyor ya
  da gömülü.
- `COLLAGE_DEV` ayarlı değil.
- TLS proxy'de terminate ediliyor ve `X-Forwarded-Proto` iletiliyor.
- Platformun stop grace period'u `Server.ShutdownTimeout`'tan uzun.
- Liveness check `/healthz`'ye bakıyor.
- Build'den önce CI'da `go test ./...` çalışıyor. [Test yazmak](/docs/testing) sayfasına
  bakın.
