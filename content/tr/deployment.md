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
CGO_ENABLED=0 GOOS=<os> GOARCH=<arch> go build -trimpath -ldflags="-s -w" -o bin/mysite .
```

- **CGO kapalıdır**, çünkü ne collage ne de standart kütüphane C'ye ihtiyaç duyar.
  Static bir binary, içinde başka hiçbir şey olmayan bir image'a konabilir.
- **`-trimpath`** sayesinde binary, kendisini build eden makinenin path'lerini
  taşımaz.
- **`-s -w`** debug tablolarını atar. Boyutun büyük kısmı bu tablolardır.
- **Varsayılan hedef bu makinedir** (v0.32.0'dan beri; öncesinde linux/amd64). Mac
  için build edilmiş bir binary Linux container'ında çalışmaz. O yüzden sunucu için
  hedefi belirtin: `collage build -os linux -arch amd64`.

| Flag | Varsayılan | Anlamı |
| --- | --- | --- |
| `-o path` | `bin/<module name>` | Binary'nin yazılacağı yer. |
| `-os name` | bu makineninki | Hedef işletim sistemi. Çoğu sunucu için `linux`. |
| `-arch name` | bu makineninki | Hedef mimari: `amd64`, Graviton ya da Ampere makineler için `arm64`. |
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
$ collage build -i -os linux -arch amd64
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

Bir değişkeni daha `main.go` değil, framework'ün kendisi okur: bir development
sunucusunun cevap vereceği ek host'ları adlandıran `COLLAGE_DEV_HOST` (v0.56.0'dan
beri). Production'da ayarlamayın. Bkz. [Development modu bu makinede kalır](#development-mode-stays-on-this-machine).

| Değişken | Varsayılan | Değeri |
| --- | --- | --- |
| `HOST` | `localhost` | Container'da `0.0.0.0`. `localhost` makinenin dışından gelen hiçbir bağlantıyı kabul etmez. Bu, yerel bir reverse proxy'nin arkasında doğrudur, başka her yerde yanlıştır. |
| `PORT` | `6060` | Platformunuz hangi port'u atıyorsa o. Binary'nin `-port` flag'i bu değeri ezer. |
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

## Graceful shutdown ve drain

`app.ListenAndServe`, `SIGINT` ve `SIGTERM` sinyallerini yakalar ve ikisinden
birinde graceful shutdown yapar. `app.Shutdown(ctx)`'i kendiniz çağırdığınızda da
aynısı olur. Sıra sabittir:

1. **Drain.** `ListenAndServe`'ün
   [`ServeHook`](/docs/writing-plugins#streams-and-shutdown) plugin'lerine
   `OnServe`'de verdiği ctx iptal edilir. Böylece bir scheduler yeni iş başlatmaz
   (v0.55.0'dan beri). [`DrainHook`](/docs/writing-plugins#streams-and-shutdown)'u
   implement eden her plugin'e bir kez haber verilir. Bir health plugin'i
   readiness check'ini burada false'a çevirir. Keep-alive'lar kapatılır: boşta
   bekleyen keep-alive bağlantıları hemen, meşgul olanlar ise o anki
   response'larından sonra kapanır. Bu bağlantıların client'ları load balancer
   üzerinden yeniden bağlanır. Port açık kalır ve request'lere
   `Server.DrainDelay` boyunca (v0.53.0'dan beri) normal şekilde cevap verilir.
2. **Stream'ler.** Development reload stream'leri ve plugin stream'leri kapatılır.
   Bunlar kendiliğinden hiç bitmez.
3. **Sunucu.** Port kapanır. Devam eden request'lerin bitmesi için en fazla
   `Server.ShutdownTimeout` kadar (varsayılan 10 saniye) beklenir.
4. **Plugin'ler.** Her plugin'in `Shutdown`'ı çalışır ve `ListenAndServe` `nil`
   döner.

`DrainDelay` varsayılan olarak 0'dır, yani hiç beklenmez. nginx, Caddy ya da
Cloudflare arkasındaki tek bir instance, v0.53.0'dan önce olduğu kadar hızlı durur.
Port kapanmadan önce bir load balancer'ın instance'ın ayrıldığını fark etmesi
gerekiyorsa bu değeri ayarlayın. Değer, load balancer'ın readiness check'inin
başarısız olması için geçen süreden birkaç saniye uzun olmalıdır:

```go
app, err := collage.New(&collage.Config{
	Server: collage.ServerConfig{
		DrainDelay:      10 * time.Second,
		ShutdownTimeout: 10 * time.Second,
	},
})
```

**İki süre toplanır.** Bir sinyalde `ShutdownTimeout` drain bittiğinde başlar. Bu
yüzden bir durdurma `DrainDelay + ShutdownTimeout` kadar sürebilir.
`app.Shutdown(ctx)`'i kendiniz çağırdığınızda tek bir ctx hem drain'i hem de devam
eden request'lerin beklenmesini sınırlar, `ShutdownTimeout` ise kullanılmaz. Bu
ctx'e en az `DrainDelay` artı request'lerinizin ihtiyaç duyduğu süre kadar bir
deadline verin. Aksi halde drain, request'lere kalacak süreyi tüketir. Bir
plugin'in `Shutdown`'ı bu deadline'ı biraz aşabilir:
[elagoht/jobs](/docs/plugins#elagohtjobs), ctx'ini dinlemeyen bir job için bir
saniye daha bekler. Bu yüzden onunla bir durdurma
`DrainDelay + ShutdownTimeout + 1s` kadar sürebilir.

Bekleme erken de bitebilir. İkinci bir `SIGINT` ya da `SIGTERM` (Ctrl-C'ye iki kez
basmak) beklemeyi hemen bitirir ve doğrudan sunucunun durdurulmasına geçer. Bitmiş
bir `Shutdown` ctx'i de aynısını yapar. Development modunda `DrainDelay` yok
sayılır, böylece restart'lar anında olur. `OnDrain` ise yine çağrılır. Henüz hiçbir
şey hizmet vermezken yapılan bir `Shutdown` plugin'lere haber verir ama beklemez,
çünkü drain edilecek trafik yoktur.

Process'i durduran şey, onu öldürmeden önce bu toplamdan daha uzun beklemelidir:

- **systemd:** `TimeoutStopSec`, `DrainDelay + ShutdownTimeout`'tan büyük olmalıdır.
  elagoht/jobs varsa buna bir saniye eklenir.
  `collage build -i`'ın yazdığı unit `TimeoutStopSec=30` ayarlar. Bu, varsayılan
  değerleri karşılar. `DrainDelay`'i artırdığınızda bunu da artırın.
- **Kubernetes:** `terminationGracePeriodSeconds` bu toplamdan büyük olmalıdır.
  Readiness probe'unu `/readyz`'ye, liveness probe'unu `/healthz`'ye yönlendirin.
  Bu iki endpoint'e [elagoht/health](/docs/plugins#elagohthealth) cevap verir.
  Böylece readiness drain başlar başlamaz başarısız olur ve pod, portu kapanmadan
  önce Service'ten çıkarılır:

```yaml
spec:
  terminationGracePeriodSeconds: 30   # > DrainDelay + ShutdownTimeout (+ 1s with elagoht/jobs)
  containers:
    - name: app
      readinessProbe:
        httpGet: { path: /readyz, port: 8080 }
        periodSeconds: 2
        failureThreshold: 1
      livenessProbe:
        httpGet: { path: /healthz, port: 8080 }
        periodSeconds: 10
        failureThreshold: 3
```

  Bu readiness probe'uyla 5 saniyelik bir `DrainDelay`, probe'un iki saniyelik
  periyodunu rahatça karşılar.

Sunucunun beklemesi için süre yetmezse, yani `ShutdownTimeout` dolduğunda hâlâ açık
bir request varsa, plugin'ler yine de kapatılır. `ListenAndServe` de bunu bildiren
bir hata döner (`collage: server shutdown: context deadline exceeded`).
`Shutdown`'ı başarısız olan bir plugin için de aynı şekilde hata döner. Scaffold'daki
`main.go` bu hatayı `log.Fatalf`'e verir. Bu yüzden process 0 yerine 1 status
koduyla çıkar. Durdurma sırasında sıfırdan farklı bir exit kodunu crash olarak gören
bir platform bunu crash olarak raporlar.

`DrainDelay`'i ya da `ShutdownTimeout`'u artırırsanız platformunuzun grace
period'unu da onunla birlikte artırın.

### Kendi sunucunuzla hizmet vermek

`OnServe`'ü yalnızca `ListenAndServe` çağırır. `app.Handler()` üzerinde kendi
`http.Server`'ını çalıştıran bir uygulama `OnServe` almaz. Bu yüzden bu tür
plugin'leri kendisi başlatmalı ve durmaya başladığında ctx'lerini iptal etmelidir.
[elagoht/jobs](/docs/plugins#elagohtjobs) için sunucu dinlemeye başlar başlamaz
onun `Start(ctx)`'ini çağırın. Uygulama, sunucusu durduktan sonra
`app.Shutdown(ctx)`'i de çağırmalıdır. Plugin'lerin `Shutdown`'ını çalıştıran bu
çağrıdır. O olmadan bir queue'da bekleyen item'lar loglanmadan bile kaybolur.

```go
// Init first, and its error: app.Handler() would only log it and answer 503,
// and the jobs plugin's Start panics before its Init has run.
if err := app.Start(); err != nil {
	return err
}
ln, err := net.Listen("tcp", ":8080")
if err != nil {
	return err
}
srv := &http.Server{Handler: app.Handler()}
go srv.Serve(ln)
ctx, stopJobs := context.WithCancel(context.Background())
j.Start(ctx) // the port is bound: start the jobs

// On shutdown:
stopJobs()                // the drain: no more triggers; Enqueue still accepts
srv.Shutdown(shutdownCtx) // requests in flight finish
app.Shutdown(shutdownCtx) // the plugins stop; jobs finishes its queues and refuses more
```

`app.Shutdown`, `DrainHook` plugin'lerine haber verir ama `DrainDelay` kadar
beklemez, çünkü `App` hizmet vermeye devam edecek bir sunucunuzdan haberdar
değildir. Onu çağırmadan önce sunucunuzu durdurun.

## Health check'ler

Liveness ve readiness, [elagoht/health](/docs/plugins#elagohthealth) plugin'inden
gelir. `/healthz`, process request'lere cevap verdiği sürece `200` döner. `/readyz`
ise sizin check'lerinizden biri başarısız olana ya da bir drain başlayana kadar
`200` döner, sonra `503` dönmeye başlar. Her platformu bu iki path'e yönlendirin:
yukarıdaki Kubernetes probe'larını ya da bir systemd unit'inin önündeki load
balancer'ın `/readyz`'deki health check'ini. Böylece readiness, drain başlar
başlamaz başarısız olur. Plugin'i, bir request'i reddedebilen ya da ona cevap
verebilen her plugin'den önce listeleyin. Nedenini plugin'in bölümünde
bulabilirsiniz.

`collage new --template demo` ile scaffold edilen proje, bir `/healthz`
[document](/docs/documents)'ı ile gelir: `documents/health.go`. `elagoht/health`
eklendiğinde `/healthz`'ye önce plugin'in middleware'i cevap verir ve document'a
hiç ulaşılmaz. Startup bunu yakalamaz, çünkü plugin document'ları değil page'leri
kontrol eder. Document'ı silin ya da plugin'in liveness endpoint'ini `livePath` ile
başka bir path'e taşıyın.

Plugin'i kullanmayan bir proje de probe'larına kendi document'larıyla cevap
verebilir. Bunlar yalnızca byte'lar ve bir content type'tır, template içermez. Bu
yüzden bir template bozuldu diye bir check başarısız olmaya başlamaz. Bunları
dynamic tutun: cache'ten sunulan bir health check, `ok` artık doğru olmaktan
çıktıktan çok sonra da `ok` cevabını verirdi. Liveness için demo'daki
`documents/health.go` kopyalanabilecek bir örnektir. Ancak böyle bir readiness
document'ı drain'den haberdar değildir, bu yüzden port kapanana kadar `200`
dönmeye devam eder. Minimal projede, yani `--template demo` olmadan `collage new`
ile oluşturulan projede hiç `/healthz` yoktur.

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

Bu cache her static ve incremental page'i tutar. Strateji tanımlamayan bir page
de, render ettiği hiçbir şeyin data handler'ı yoksa static'tir; bkz.
[Caching](/docs/caching#a-page-that-declares-none). Render edilmiş page'ler
restart'tan sağ çıkar. Böylece aynı build'i yeniden deploy
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

Öndeki bir CDN kendi kopyalarını tutar.
[elagoht/cdnpurge](/docs/plugins#elagohtcdnpurge), collage bir page'i invalidate
ettiğinde bu kopyaları purge eder. [elagoht/compress](/docs/plugins#elagohtcompress)
ise response'ları Brotli ve gzip ile, okuyucu başına değil cache'lenen page başına
bir kez sıkıştırır.

### Memory cache ve container'ın bellek sınırı

`"memory"` cache `MaxBytes` ile sınırlıdır. Siz ayarlamadıkça bu sınır saklanan
page'ler için 256 MiB'tır ve gerçekten işler: cache dolduğunda canlı heap büyümeyi
bırakır. Ama process orada durmaz. Go'nun garbage collector'ı yeniden çalışmadan
önce heap'in, son toplamadan sonra canlı kalan miktarın yaklaşık iki katına
çıkmasına izin verir. Bu yüzden dolu bir 256 MiB'lık cache, 512 MiB'ı epey aşan
bir process demek olabilir. Yaklaşık 14 KB'lık bir page 30.000 farklı URL ile
istendiğinde process yaklaşık 720 MB'ta dengelendi. Bu, bütün ayarlar varsayılan
değerinde bırakıldığında 512 MiB'lık bir container'ın öldürülmesine yeter.

Runtime'a ne kadar belleği olduğunu `GOMEMLIMIT` ile, container'ın sınırının biraz
altında bir değerle söyleyin. Böylece runtime sınıra yaklaştıkça öldürülmek yerine
belleği daha sıkı toplar:

```dockerfile
# In a 512 MiB container.
ENV GOMEMLIMIT=400MiB
```

Aynı ölçüm `GOMEMLIMIT=320MiB` ile yaklaşık 430 MB'ta kaldı. Bu sınır yumuşaktır:
runtime'ın hedeflediği bir değerdir, zorla uyguladığı bir tavan değildir. Bu yüzden
`MaxBytes`'ı bunun epey altında tutun. Cache'in tuttuğu her şey canlıdır ve runtime
ne kadar uğraşırsa uğraşsın toplanamaz. Boyutu sınırın üçte biri ile yarısı
arasında olan bir cache, render'ların kendisine de yer bırakır. 512 MiB'lık bir
container'da `MaxBytes`'ı varsayılanından 128–192 MiB'a indirin. Scaffold'un disk
cache'i page'leri diskte tuttuğu için buna ihtiyaç duymaz.

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
düz HTTP üzerinden göndermez. Header'ı değiştirmek yerine sonuna ekleyen bir proxy
bir liste gönderir: `https, http`. v0.34.0'dan beri geçerli olan, okuyucunun kendi
bağlantısını gösteren ilk entry'dir. Tarayıcının gönderdiği `Host`'u da iletin.
Forgery kontrolü eski bir tarayıcının `Origin`'ini onunla karşılaştırır ve `Host`
page cache'in key'inin bir parçasıdır. Sertifikayı da kendisi alan minimal bir Caddy
config'i şöyledir:

```
example.com {
	reverse_proxy 127.0.0.1:8080
}
```

nginx'te `location` bloğuna eklenen `proxy_set_header X-Forwarded-Proto $scheme;`
aynı işi görür. Güvenlik header'larının kendileri, yani HSTS, bir
Content-Security-Policy ve diğerleri, [elagoht/secure](/docs/plugins#elagohtsecure)
plugin'iyle binary'den de gelebilir.
[elagoht/ratelimit](/docs/plugins#elagohtratelimit) tek bir istemcinin form'ları ne
kadar hızlı gönderebileceğini sınırlar. [elagoht/basicauth](/docs/plugins#elagohtbasicauth)
ise bir staging deploy'unun önüne bir parola koyar.

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
handling artık sizin sorumluluğunuzdadır. Önce `app.Start()`'ı çağırırsınız,
`ServeHook` plugin'lerini kendiniz başlatırsınız ve plugin'lerin kapanması için
sunucunuz durduktan sonra `app.Shutdown(ctx)`'i çağırırsınız. Bkz.
[Kendi sunucunuzla hizmet vermek](#serving-with-your-own-server).


## Bir proxy'nin arkasında: `TrustedProxies`

Bir proxy'nin arkasında her request'in `RemoteAddr`'i proxy'ninkidir. Gerçek istemciyi
proxy `X-Forwarded-For`'da bildirir, ama bunu başkası da yapabilir: kendi
`X-Forwarded-For`'unu gönderen bir istemciye inanılmaz. `Server.TrustedProxies`
(v0.47.0'dan beri) kimin söylediğine inanılacağını belirler:

```go
Server: collage.ServerConfig{
	TrustedProxies: []string{"10.0.0.0/8", "127.0.0.1"},
},
```

Her girdi bir adres ya da CIDR aralığıdır; ikisinden de olmayan bir girdi
`collage.New`'ı başarısız kılar. `collage.ClientIP(r)` bundan sonra istemcinin kim
olduğunu söyler: `RemoteAddr`'in host'u; ta ki o güvenilen bir proxy olana kadar.
Öyleyse header sağdan okunur, güvenilen adresler atlanır ve güvenilmeyen ilk adres
istemcidir (hepsi güvenilense en soldaki). Boşken, yani varsayılanda, hiçbir header'a
güvenilmez ve `ClientIP` her zaman `RemoteAddr`'dir.

Header'daki bir girdi, bazı proxy'lerin yazdığı gibi port (`9.9.9.9:4567`,
`[2001:db8::1]:443`) ya da köşeli parantez (`[2001:db8::1]`) taşıyabilir; adres olarak
okunur. Yine de adres olmayan bir girdi, örneğin bazı proxy'lerin gönderdiği
`unknown`, okumayı bitirir. Ondan önce güvenilmeyen bir adres bulunmadıysa istemci
bilinmez ve `ClientIP` sıfır `netip.Addr`'dir: asla proxy değil, çünkü o her
ziyaretçiyi aynı istemci yapardı. İstemciye göre iş yapan bir plugin böyle bir
request'i atlar.

İstemciyle sunucu arasındaki her hop'u yazın: kendi proxy'lerinizi ve bir CDN'in
arkasındaysanız CDN'in yayımladığı aralıkları da. Unutulan bir hop istemci sanılır ve
onun üzerinden gelen her ziyaretçi tek bir istemci olur. Yalnızca kendi işlettiğiniz
proxy'leri ya da platformunuzun belgelediği aralıkları yazın: istemcinin request
gönderebildiği bir aralığa güvenmek, o istemcinin istediği adresi kendine vermesine
izin vermektir. Sıfır bitlik bir girdi (`0.0.0.0/0`, `::/0`) herkese güvenir;
`collage.New` onu kabul eder ama bunu söyleyen bir Warn log'lar.

Aynı liste, hangi proxy'nin `X-Forwarded-Proto`'suna inanılacağını da belirler
(v0.56.0'dan beri). Forgery cookie'si, request TLS üzerinden geldiğinde `Secure`
olur. TLS'i terminate eden bir proxy'nin arkasında bunu proxy o header ile söyler.
`TrustedProxies` set edilmişse `RemoteAddr`'ı listedeki bir proxy olmayan bir
request bu header'ı set edemez. Port'a düz http ile doğrudan ulaşan bir client, düz
http üzerinden geri gönderebileceği bir cookie alır. Liste boşsa header herkesten
kabul edilir. Liste var olmadan önce de böyleydi ve yalnızca proxy'nizden başka
hiçbir şeyin sunucuya ulaşamadığı durumda doğrudur.

## Timeout'lar

`ListenAndServe`, `Config.Server` içindeki timeout'ları uygular:

| Alan | Varsayılan | Neyi sınırlar |
| --- | --- | --- |
| `ReadTimeout` | 15s | Header'lar dahil request'in okunmasını. Header'ları yavaş gönderen bir client bağlantıyı açık tutamaz. |
| `ReadHeaderTimeout` | 0 | Yalnızca request header'larının okunmasını. Böylece header'ları bir byte'ta bir gönderen bir client erken düşürülür, büyük bir body ise `ReadTimeout`'un tamamını alır. Sıfır, `ReadTimeout`'u kullanır (v0.56.0'dan beri). |
| `WriteTimeout` | 30s | Response'un yazılmasını. Request'in header'ları okunduğunda başlar, bu yüzden bir upload'un body'sini de sınırlar. Verisi bundan uzun süren bir page yarıda kesilir. |
| `IdleTimeout` | 60s | Bir sonraki request'ini bekleyen keep-alive bağlantısını. |
| `DrainDelay` | 0 | Bir `SIGTERM`'den sonra, keep-alive'lar kapalı olarak hizmet vermeye devam edilen süreyi; böylece load balancer trafik göndermeyi bırakabilir (v0.53.0'dan beri). |
| `ShutdownTimeout` | 10s | Drain'den sonra, port kapandığında devam eden request'lerin bitmesini. |
| `MaxBodyBytes` | 4 MiB | Bir action'ın request body'sini, action kendi sınırını belirlemediyse. Negatif değer sınırsız demektir. |

```go
Server: collage.ServerConfig{
	Host:         envString("HOST", "localhost"),
	Port:         port,
	WriteTimeout: time.Minute,
},
```

`collage dev`'in proxy'si yalnızca development'ta geçerli, sabit 10 saniyelik kendi
header timeout'unu kullanır. Bunlardan herhangi birindeki negatif değer
`collage.ErrNegativeDuration`'dır.

Upload'lar için yalnızca `ReadTimeout`'u artırmak bir tuzaktır. `WriteTimeout`'tan
uzun süren bir upload okunur ve handler onu kaydeder, ama cevap yarıda kesilir.
Client bağlantının kapandığını görür ve yeniden dener. İkisini birden bütün sunucu
için artırmak ise yavaş her client'ın her endpoint'i o kadar süre tutmasına izin
verir. Bunun yerine upload action'ına `WithBodyTimeout` ile kendi deadline'ını
verin. [Upload'lar](#uploads) bölümüne bakın.

Bir data handler'ın ne kadar sürebileceği ayrı bir ayardır. Bunu her fragment'in
`WithTimeout`'u belirler. Timeout belirtmeyen fragment'ler ve document'lar için
`Template.Timeout` (5 saniye) geçerlidir. Bu süreyi `WriteTimeout`'un epey altında
tutun. Böylece yavaş bir upstream, page'in ortasında kapanan bir bağlantıya değil,
fragment'in fallback'ine dönüşür. [Config](/docs/configuration) sayfasına
bakın.

## Upload'lar

Bir upload ([Streaming body'ler](/docs/forms-and-actions#streaming-bodies)
bölümüne bakın), tarayıcı ile handler arasındaki her limitle karşılaşır. Bu
limitlerin çoğunun varsayılanları form'lara göre boyutlanmıştır.

**Sunucunun deadline'ları.** Yukarıdaki varsayılanlar her upload'u 15 saniyelik
okumadan sonra bitirir. Upload action'ındaki `WithBodyTimeout(d)` (v0.57.0'dan
beri), iki deadline'ı da yalnızca o action için şimdi artı `d` ile değiştirir. Bunu
page'in guard'larından sonra (streaming bir action'da ise forgery kontrolünden de
sonra) yapar. `d`'yi en az `MaxBodyBytes` bölü hizmet verdiğiniz en yavaş bağlantı
artı cevap verme süresi kadar tutun. Client'ı gitmiş ya da süresi dolmuş bir upload
sunucu hatası olarak değil, `400` ya da `408` olarak kaydedilir ve debug
seviyesinde log'lanır. Giden bir client hiçbir şey almaz. `WithBodyTimeout`
altında süresi dolan client da almaz, çünkü write deadline'ı read deadline'ı ile
birlikte geçmiştir.

**nginx**, `client_max_body_size`'ı (varsayılan 1 MiB) aşan her şeye collage onu
görmeden kendi `413`'ü ile cevap verir. Varsayılan olarak ayrıca request body'sinin
tamamını, bir kısmını bile iletmeden önce kendi diskine buffer'lar. Stream, kötü
bir dosyanın erkenden reddedilmesi ve upload'un ilerleyişi kaybolur. Dolan da
nginx'in temp dizini olur. Upload location'ı için:

```nginx
location /upload {
    client_max_body_size     2g;   # MaxBodyBytes, or a little more
    proxy_request_buffering  off;  # stream the body through
    proxy_send_timeout       30m;  # between two writes to collage
    proxy_read_timeout       30m;  # waiting for collage's answer
    proxy_pass               http://127.0.0.1:8080;
}
```

**Caddy** request body'lerini stream eder. Limiti, ayarlandığında,
`request_body { max_size 2GB }`'dir.

**Cloudflare ve diğer CDN'ler** request body'sini plana göre sınırlar (bu yazı
yazıldığında Free ve Pro planlarında 100 MB'tı; kendi planınız için Cloudflare'in
dokümantasyonuna bakın) ve daha büyüğüne kendileri cevap verir. Load balancer'ların
da kendi sınırları ve idle timeout'ları vardır. Daha büyük dosyalar için CDN'i
atlayan bir hostname ya da parça parça devam ettirilebilen upload'lar gerekir.
collage ikincisini yapmaz.

**[elagoht/health](/docs/plugins#elagohthealth)'in `maxInFlight`'ı**, upload'lar
dahil hizmet verilen her request'i sürdüğü boyunca sayar. Bunu aynı anda
beklediğiniz upload'lara göre boyutlandırın. Aksi halde yavaş upload'lar slot'ları
tutar ve sitenin geri kalanına `503` ile cevap verilir.

## Loglar

`Config.Logger` verilmezse collage, `slog`'un varsayılan handler'ına log yazar.
Terminalde çalışıyorsa insanların okuması için formatlanmış bir handler kullanır. Bir
sunucuda loglar bir makine tarafından okunur. Orada bir JSON handler verin:

```go
Logger: slog.New(slog.NewJSONHandler(os.Stdout, nil)),
```

collage her request'i değil, ters giden şeyleri log'lar.
[elagoht/accesslog](/docs/plugins#elagohtaccesslog) aynı logger üzerinden her
request için request id'li bir satır yazar.
[elagoht/prometheus](/docs/plugins#elagohtprometheus) framework'ün metric'lerini
`/metrics`'te sunar. [elagoht/otel](/docs/plugins#elagohtotel) ise span'lerini
OpenTelemetry trace'lerine dönüştürür.

## Development modu bu makinede kalır

`DevMode`, production'ın sunmaması gereken şeyleri sunar: stack'li ve kaynak kodlu
hata sayfaları, reload stream'i, bir development toolbar'ı. Bunları burada tutan iki
koruma vardır (v0.56.0'dan beri):

- **Yabancı bir `Host` reddedilir.** Development'ta sunucu, `Host`'u localhost (ya
  da `.localhost` altındaki bir ad), bir IP adresi, ayrılmış bir ad, `Server.Host`
  veya `COLLAGE_DEV_HOST` içindeki bir ad olmayan request'lere `403` ile cevap verir.
  Ayrılmış adlar `example.com`, `example.net`, `example.org` ve bunların altındaki
  adlar ile `.example`, `.test` ve `.invalid` altındaki adlardır. Bu kural uygulama
  doğrudan çalıştırıldığında da geçerlidir, yalnızca `collage dev` altında değil.
  Başka bir sitedeki bir page kendi adının `127.0.0.1`'e çözümlenmesini sağlayabilir
  (DNS rebinding) ve böylece development sunucusuyla same-origin olur, ama gönderdiği
  `Host` yine kendi adıdır. Ayrılmış adlara izin verilir, çünkü hiç kimse onlardan
  birini size yönlendirmek için register edemez. Test'lerin kullandığı adlar da
  bunlardır (httptest'in request'leri `example.com` içindir). Bir static build'in
  capture'ları kontrol edilmez.
- **Erişilebilir bir adres için uyarı verilir.** Development'ta `ListenAndServe`
  loopback dışında bir adrese (`0.0.0.0`, `::`, bir LAN adresi) bağlandığında,
  development sayfalarının uygulamanın iç yapısını açığa çıkardığını söyleyen bir
  Warn log'lar. `collage dev` de kendi proxy'si öyle bağlandığında aynısını log'lar.

**Yükseltme:** ayrılmış adların dışında özel bir `Host` kullanan bir `DevMode`
test'i ya da request'i, örneğin bir tenant domain'i, artık `403` alır. Adı
`Server.Host` olarak set edin, `COLLAGE_DEV_HOST` içine yazın ya da `tenant.test`
gibi ayrılmış bir ad kullanın.

### Development'a başka bir adla erişmek

`COLLAGE_DEV_HOST`, bir development sunucusunun cevap vereceği ek adların virgülle
ayrılmış listesidir. Şu durumlarda set edin:

- bir docker-compose servis adı, yani başka bir container'dan `http://app:6060`;
- ayrılmış adların dışında, `/etc/hosts` içindeki bir ad;
- tek bir development sunucusunda birkaç tenant host'u (ya da hiçbir ayar
  gerektirmeyen `*.localhost` adlarını kullanın);
- `0.0.0.0`'a bağlı bir sunucuya telefondan makinenin LAN adıyla erişmek. IP adresiyle
  erişmek için hiçbir şey gerekmez.

```
COLLAGE_DEV_HOST=app,mybox.lan
```

Development modundaki bir uygulama bunu kendi process ortamından okur. Doğrudan
çalıştırıyorsanız shell'de ya da process'in ortamını aldığı yerde set edin.
`collage dev` onu shell'den ya da `.env.development`'tan okur, kendi proxy'sine
uygular ve programa listeyi kendi `HOST`'uyla birlikte verir. Program loopback'te
dinler ama tarayıcının `Host`'unu görür. Reddedilen bir request'in `403`'ü `Host`'u
alıntılar ve ayarın adını verir. ngrok gibi bir tünel, adını listeye yazmadıkça
reddedilir. Bu bilerek böyledir, çünkü development sayfalarını internete açar.

İki koruma da production'da geçerli değildir. Orada `DevMode` kapalıdır ve `Host`,
DNS'inizin gönderdiği ne ise odur.

## Kontrol listesi

- `COLLAGE_CSRF_KEY` ayarlı ve her instance'ta aynı.
- Container'da `HOST=0.0.0.0`, `PORT` platformdan geliyor.
- `.cache` için yazılabilir bir çalışma dizini var ve eski `.cache/<hash>` dizinleri
  deploy sırasında siliniyor.
- Plugin'leri yapılandırıyorsanız `plugins-config.json` çalışma dizininde duruyor ya
  da gömülü.
- `COLLAGE_DEV` ayarlı değil.
- TLS proxy'de terminate ediliyor; `X-Forwarded-Proto` ve tarayıcının `Host`'u
  iletiliyor ve `Server.TrustedProxies` proxy'yi listeliyor.
- Form'ları buraya post eden diğer her origin (örneğin bir admin subdomain'i)
  `Security.CSRFTrustedOrigins`'te belirtilmiş.
- Upload action'ının kendi `WithMaxBodyBytes`'ı ve `WithBodyTimeout`'u var, öndeki
  proxy de o boyuta izin veriyor ve body'yi stream ederek iletiyor.
  [Upload'lar](#uploads) bölümüne bakın.
- Platformun stop grace period'u `Server.DrainDelay + Server.ShutdownTimeout`'tan
  uzun. elagoht/jobs varsa bir saniye daha uzun.
- [elagoht/health](/docs/plugins#elagohthealth) ile liveness check `/healthz`'ye,
  readiness check `/readyz`'ye bakıyor.
- Build'den önce CI'da `go test ./...` çalışıyor. [Test yazmak](/docs/testing) sayfasına
  bakın.
