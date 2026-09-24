---
description: collage CLI'nin her komutu — new, dev, build, export, serve, version ve help — flag'leri ve tam olarak neyi çalıştırdığıyla.
---

# collage CLI

`collage` komutu projelerin iskelesini oluşturur ve elinizdeki projeleri yönetir:
birini geliştirme modunda çalıştırmak, yayına aldığınız binary'yi derlemek ve
statik dosyalar olarak dışa aktarmak.

```sh
go install github.com/Elagoht/collage/cmd/collage@latest
```

Uygulamanızı asla kendi içine bağlamaz — bağlayamaz da, çünkü uygulamanız sizin
kodunuzdur. `dev`, `build` ve `export`, `go` aracını geçerli dizinde, tam da elle
yapacağınız gibi çalıştırır; bu sayfanın geri kalanı her birinin tam olarak neyi
çalıştırdığını anlatır.

## Kullanım

```sh
collage <command> [flags]
```

| Komut | Ne yapar |
| --- | --- |
| `new` | Yeni bir collage projesinin iskelesini oluşturur |
| `dev` | Geçerli dizindeki projeyi geliştirme modunda çalıştırır |
| `build` | Geçerli dizindeki projeyi yayına aldığınız binary'ye derler |
| `export` | Geçerli dizindeki projeyi statik dosyalara render eder |
| `serve` | Statik bir dışa aktarmayı, statik barındırma hizmetinin sunacağı gibi sunar |
| `version` | collage CLI sürümünü yazdırır |
| `help` | Bir komutun yardımını gösterir ya da tüm komutları listeler |

`collage help <command>` o komutun kendi kullanımını yazdırır; `collage <command> -h`
de aynısını yapar. Flag'ler Go'nun `flag` sözdizimini kullanır: `-out dist` ile
`-out=dist` aynıdır, `--out` da çalışır.

### Çıkış kodları

| Kod | Anlamı |
| --- | --- |
| `0` | Başarı; `help`, `collage -h` (v0.11.0'dan itibaren) ve `collage <command> -h` dahil |
| `1` | Komut doğru ayrıştırıldı ama işini yapamadı |
| `2` | Bir kullanım sorunu: komut yok, bilinmeyen bir komut, hatalı bir flag ya da beklenmeyen bir argüman |

## collage new

```sh
collage new <name> [--template demo|minimal] [--dir path] [--module path] [--force]
```

`<name>` adında, çalışmaya hazır yeni bir projenin iskelesini oluşturur.

| Flag | Varsayılan | Anlamı |
| --- | --- | --- |
| `--template name` | `demo` | İskelesi oluşturulacak proje: `demo` ya da `minimal` (v0.14.2'den itibaren; `-minimal`'ın yerini alır) |
| `-dir path` | `./<name>` | İskelenin oluşturulacağı dizin |
| `-module path` | `<name>` | `go.mod`'a yazılan modül yolu |
| `-force` | kapalı | Boş olmayan bir dizine yine de iskele oluşturur |

```sh
collage new myblog                                   # into ./myblog, module "myblog"
collage new myblog --template minimal                # one page, nothing to delete
collage new myblog -module github.com/me/myblog
collage new myblog -dir . -force                     # into the current, non-empty directory
```

Flag'ler tek ya da çift tireyle yazılabilir ve addan önce ya da sonra gelebilir.
Tam olarak bir ad gereklidir; hiç ad vermemek ya da birden fazla vermek bir kullanım hatasıdır. Var olan ve boş olmayan
bir hedef dizin, `-force` vermediğiniz sürece reddedilir — `-force` ile de
iskelenin yazdığı dosyalar aynı addaki dosyaların yerini alır.

**Her proje** şunları alır: `go.mod`; yapılandırmayı, statik mount'u,
[aşağıda](#the-contract-with-maingo) anlatılan CLI sözleşmesini ve
[plugin komutlarının](#plugin-commands) dağıtımını barındıran bir `main.go`; her
route'u kaydeden bir `routes.go`; sitenin başlığını `rc.HoistTitle` ile bildiren
bir layout (böylece bir sayfanın kendi başlığı onun yerini alır); bir ana sayfa,
`static/`, bir `.gitignore` ve bir README.

**Demo projesi**, yani varsayılan, buna canlı demolardan oluşan bir sayfa ekler —
JSON ile yanıt veren bir action, kendi sayfasına gönderilen bir form, kendi URL'si
olan bir fragment, bir JSON document — ve bunları `pages/`, `fragments/`,
`actions/`, `documents/` ve `store/` altına bölüştürür; ayrıca bir bulunamadı
sayfası, her biri için testler, `plugins-config.json`, bir favicon ve bir
`.env.example` ekler.

**`--template minimal`** bir projenin olabileceği en az şeydir: tek bir sayfayı,
`<h1>Hello from collage</h1>`'i saran layout ve karanlık mod dahil arka planı ve
metin rengini ayarlayan bir stil dosyası. Başka hiçbir şey yok — test yok,
bulunamadı sayfası da yok: siz bir tane kaydedene kadar collage bilinmeyen bir
adrese kendi sade 404'üyle yanıt verir.

Bittiğinde sonraki adımları yazdırır — `cp` satırını yalnızca demo projesi için,
yani `.env.example`'ı olan proje için:

```sh
cd myblog
go mod tidy
cp .env.example .env.development
collage dev
```

## collage dev

```sh
collage dev
```

Geçerli dizindeki projeyi derler, `COLLAGE_DEV=1` ayarlıyken çalıştırır ve Go
kodu her değiştiğinde yeniden derleyip yeniden başlatır. Hiçbir flag ya da argüman
almaz. Durdurmak için Ctrl-C'ye basın; program da kesmeyi alır ve production'da
nasıl kapanıyorsa öyle kapanır.

İskelesi oluşturulan `main.go`, `COLLAGE_DEV=1` ayarlı olduğunda geliştirme modunu
açar. Geliştirme modu şablonları ve statik dosyaları her istekte diskten okur; bu
yüzden onları düzenlemek yeniden derleme gerektirmez ve derleme olmaz — onun
yerine tarayıcınızdaki sayfa kendini yeniler. Programınızın diskten kendisinin
okuduğu içerik, örneğin Markdown, dizini
[`Config.DevWatch`](/docs/configuration#devwatch) içinde adlandırıldığında sayfayı
o da yeniler (v0.10.0'dan itibaren). Geliştirme modunun başka neleri değiştirdiği
için bkz. [Şablonlar](/docs/templates#reloading-in-development).

### Yeniden derlemeyi ne tetikler

| İzlenir | İzlenmez |
| --- | --- |
| projenin herhangi bir yerindeki `.go` dosyaları | `_test.go` dosyaları |
| proje kökündeki `go.mod` ve `go.sum` | şablonlar, statik dosyalar ve `DevWatch` dizinleri (yeniden derlemeden yenilenir) |
| aşağıdaki ortam dosyası | gizli dizinler (`.git`, `.cache`, …) |
| | `bin`, `dist`, `node_modules`, `testdata`, `vendor` |

Atlanan dizinler, döngünün kendi kendini beslemesini önleyen şeydir: çalışan
programın yazdığı hiçbir şey — önbelleği, bir dışa aktarma — yeniden derlemeyi
tetikleyemez.

Bir yeniden derleme şöyle ilerler:

- **Her 300 ms'de bir yoklar** (polling), değiştirilme zamanlarını ve boyutları
  karşılaştırarak. Dosya sistemi bildirim kütüphanesi yok, her platformda aynı
  davranış.
- **Bir yazma patlaması tek bir yeniden derlemedir.** Bir değişiklikten sonra
  dosyalar bir aralık boyunca sessiz kalana kadar bekler; böylece birkaç dosyayı
  yeniden yazan bir formatlayıcı tek bir derlemeye mal olur.
- **Yeni build önce yapılır.** `go build`, proje dışındaki geçici bir dizine bir
  binary yazar; eski süreç ancak bu derlendikten sonra kesilir — Ctrl-C'de olduğu
  gibi isteklerini boşaltır ve yalnızca 10 saniye içinde çıkmamışsa öldürülür — ve
  yenisi başlatılır.
- **Derlenmeyen bir değişiklik son sağlam build'i sunmaya devam ettirir**;
  derleyicinin hatası ekranda kalır.
- **Kendiliğinden çıkan bir program** — başlangıçta bir panic, zaten kullanımda olan
  bir port — döngü içinde yeniden başlatılmaz. Bir sonraki değişiklik onu yeniden
  başlatır.

### Ortam dosyaları

`collage dev`, geçerli dizindeki `.env.development`'ın değişkenlerini, ya da
`.env.development` yoksa `.env`'in değişkenlerini programın ortamına ekler. Tek bir
dosya, asla ikisi birden: `.env.development`, `.env`'in üzerine birleştirilmez,
onun yerini alır.

```sh
# .env.development
PORT=3000
HOST=localhost
export COLLAGE_CSRF_KEY="0f1e2d3c4b5a69788796a5b4c3d2e1f00f1e2d3c4b5a69788796a5b4c3d2e1f0"
```

- Biçim `KEY=value` satırları, boş satırlar ve `#` ile başlayan satırlardır. Bir
  `export ` öneki kabul edilir; bir değerin etrafındaki eşleşen tek ya da çift
  tırnak çifti de kabul edilir ve içindeki hiçbir şey yorumlanmadan kaldırılır.
- Tırnaksız bir değer, boşluktan sonra gelen bir `#`'te biter; yani
  `PORT=3000 # dev`, `3000`'dir.
- Anahtarlar harf, rakam ve alt çizgiden oluşur ve rakamla başlamaz.
- **Bozuk bir satır programın başlamasını engeller**: `collage dev` onu dosya ve
  satırıyla (`.env.development:3`) bildirir ve düzeltilene kadar hiçbir şeyi
  başlatmaz ya da yeniden başlatmaz — zaten çalışan bir program, başlatıldığı
  değerlerle çalışmaya devam eder. İzlemeyi sürdürür; bu yüzden düzeltilmiş dosyayı
  kaydetmek işi kaldığı yerden devam ettirir. Atlanan bir satır, sizin yazdığınız
  ama programın hiç görmediği bir ayar olurdu.
- **Kabukta zaten ayarlı bir değişken dosyaya karşı kazanır**; bu yüzden
  `PORT=4000 collage dev` yine çalışır.
- Dosya ne derse desin `COLLAGE_DEV=1` her zaman ayarlanır.
- Dosyanın olmaması bir hata değildir. Bir dosya okunduğunda adı stderr'e yazdırılır.
- Dosya her yeniden başlatmada yeniden okunur ve izlenir; bu yüzden onu düzenlemek
  programı yeni değerlerle yeniden başlatır.

Bu dosyaları yalnızca `collage dev` okur. `collage build`, `collage export` ve
derlenmiş binary ortamlarını nerede çalışıyorlarsa oradan alır.

## collage build

```sh
collage build [-o path] [-os name] [-arch name] [-i]
```

Geçerli dizindeki projeyi yayına aldığınız binary'ye derler. Şunu çalıştırır:

```sh
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o bin/<name> .
```

| Flag | Varsayılan | Anlamı |
| --- | --- | --- |
| `-o path` | `bin/<name>` | Binary'nin yazılacağı yer |
| `-os name` | `linux` | Hedef işletim sistemi (`GOOS`) |
| `-arch name` | `amd64` | Hedef mimari (`GOARCH`) |
| `-i` | kapalı | Binary'nin yanına hangi ek dosyaların yazılacağını sorar |

`<name>`, `go.mod`'daki modül yolunun son öğesidir; `go.mod`'u olmayan ya da hiçbir
modül bildirmeyen bir dizin hatadır. `-os windows` için, yolda yoksa sonuna `.exe`
eklenir. Konumsal argümanlar bir kullanım hatasıdır.

Neden bu ayarlar:

- **CGO kapalı**, çünkü collage'ın ve standart kütüphanenin C'ye ihtiyacı yok ve
  statik bir binary, içinde başka hiçbir şey olmayan bir imaja konabilir.
- **`-trimpath`**, böylece binary onu derleyen makinenin yollarını taşımaz.
- **`-s -w`** hata ayıklama tablolarını atar; boyutun büyük kısmı onlardır.
- Üzerinde olduğunuz makine yerine **varsayılan olarak linux/amd64**: bir Mac'te
  derlenen binary bir Linux container'ında çalışmaz ve bunu öğrenmenin yeri bir
  sunucudaki "exec format error" değildir.
- **`dist/` değil, `bin/`**: `dist/`, `collage export`'un yazdığı yerdir ve
  `export -clean` onu boşaltır.

Bittiğinde binary'nin yolunu, platformunu, boyutunu ve derleme süresini yazdırır.

### -i ile ek dosyalar

`-i` ile, derlemeden önce standart girdiden yanıt okuyarak iki soru sorar — `y` ya
da `yes` evet demektir, geri kalan her şey hayır:

- **Dockerfile yazılsın mı?** İki aşamalı bir imaj: CLI'yi derleyen Go'nun major ve
  minor sürümüne sabitlenmiş bir `golang` derleme aşaması ve yalnızca binary'yi
  barındıran, `HOST=0.0.0.0`, `PORT=8080`, dışa açılmış 8080 portu ve yorum satırı
  hâline getirilmiş bir `COLLAGE_CSRF_KEY` içeren bir
  `gcr.io/distroless/static-debian12` aşaması.
- **systemd unit'i yazılsın mı?** `/usr/local/bin/<name>`'i `/srv/<name>`'den
  `HOST=127.0.0.1`, `PORT=8080`, `Restart=on-failure` ve `TimeoutStopSec=30` ile
  çalıştıran, kurmadan önce düzenlenmek üzere bir `<name>.service`.

İkisi de binary'nin yanına yazılır — `bin/Dockerfile`, `bin/<name>.service` —
çünkü üretilmiş dosyalardır ve proje kökü bir insanın yazdıkları içindir. Rapor, bu
yerleşimin bedeli olan tek komutu yazdırır:

```sh
docker build -f bin/Dockerfile .
```

**Var olan bir dosyanın üzerine asla yazılmaz.** Dosya zaten oradaysa soru
sorulmaz. Bunlar bir projenin düzenlediği dosyalardır ve birini varsayılanla
değiştiren bir derleme komutu, birinin emeğini sessizce geri alırdı.

`-i` olmadan yalnızca binary yazılır. Bkz. [Yayına alma](/docs/deployment).

## collage export

```sh
collage export [-out dir] [-clean]
```

Geçerli dizindeki projeyi statik barındırma hizmeti için statik dosyalara render
eder — HTML olabilecek her sayfa için HTML, artı mount edilmiş her asset. Şunu
çalıştırır:

```sh
go run . -collage-build -out <dir>          # plus -clean when you passed it
```

| Flag | Varsayılan | Anlamı |
| --- | --- | --- |
| `-out dir` | `dist` | Projenin render edildiği dizin |
| `-clean` | kapalı | Derlemeden önce dizinin mevcut içeriğini siler |

Programın çıktısı — build raporu — yeniden biçimlendirilmeden doğrudan aktarılır.
Konumsal argümanlar bir kullanım hatasıdır.

Neyin dışa aktarıldığı, neyin atlandığı ve neden, ve çıktı dizinine yönelik
güvenlik denetimleri [Statik dışa aktarma](/docs/static-export) sayfasındadır.
Formları ya da istek başına sayfaları olan bir site için istediğiniz
`collage build`'dir.

## collage serve

```sh
collage serve [-dir dir] [-host name] [-port n]
```

Statik bir dışa aktarmayı statik barındırma hizmetinin sunacağı gibi sunar; böylece
gördüğünüz, yayına aldıktan sonra alacağınız şeydir. Dosya sunar; projenizi
çalıştırmaz.

| Flag | Varsayılan | Anlamı |
| --- | --- | --- |
| `-dir dir` | `dist` | Sunulacak dizin |
| `-host name` | `localhost` | Dinlenecek arayüz |
| `-port n` | `4000` | Dinlenecek port |

Port 3000 değil 4000'dir, böylece `collage dev`'in yanında çalışabilir — ikisini
karşılaştırdığınız an da tam olarak budur.

Bir dosya sunucusu gibi değil, statik barındırma hizmeti gibi davranır:

- Uzantısı olmayan bir yola `<path>/index.html` ile yanıt verilir; bir dışa
  aktarmanın yazdığı yapı budur.
- Dizinler asla listelenmez.
- Hiçbir şeye çözümlenmeyen bir yola dışa aktarmanın kendi `404.html`'i ve 404
  durumuyla, o yoksa düz bir 404 ile yanıt verilir.
- Yalnızca `GET` ve `HEAD` yanıtlanır; geri kalan her şey 405'tir.
- Her yanıt `Cache-Control: no-store` ile gönderilir; böylece yeniden dışa aktarıp
  yenilemek eski çıktıyı değil yenisini gösterir.

Var olmayan ya da hiç dosya barındırmayan bir dizin, önce `collage export`
çalıştırmanızı söyleyen bir hatadır. Var olan ama dizin olmayan bir yol da tam
olarak bunu söyleyen bir hatadır.

## collage version

```sh
collage version
```

`collage version <version>` yazdırır. Sürüm binary'nin build bilgisinden okunur;
bu yüzden `go install ...@v0.9.0` `0.9.0` bildirir, bir git checkout'unda derlenen
binary ise bir pseudo-version bildirir — v0.11.0 etiketinden sonraki bir commit
için `0.11.1-0.<timestamp>-<commit>`.
Yalnızca hiç sürüm bilgisi olmayan bir binary — bir deponun dışında ya da
`-buildvcs=false` ile derlenmiş — `devel` bildirir.

## collage help

```sh
collage help            # every command
collage help export     # one command's usage
```

Argümansız çalıştırıldığında her komutu listeler ve `0` ile çıkar. Bir komut adıyla
o komutun kullanımını yazdırır; bilinmeyen bir ad
`collage: unknown command: "nope"` yazdırır ve `2` ile çıkar.

## main.go ile sözleşme

`dev` ve `export`, `main.go`'nuzun yaptığı iki şeye dayanır ve iskelesi oluşturulan
`main.go` ikisini de yapar — üçüncüsüyle birlikte: hiçbir `collage` komutunun
ihtiyaç duymadığı ama plugin'lerinizin duyduğu, plugin komutlarını çalıştırmak:

| Komut | Çalıştırdığı | `main.go`'nuz şunu yapmalı |
| --- | --- | --- |
| `collage dev` | `go build`, sonra `COLLAGE_DEV=1` ile binary | `COLLAGE_DEV` `1` olduğunda geliştirme modunu açmak |
| `collage export` | `go run . -collage-build -out <dir> [-clean]` | `-collage-build`, `-out` ve `-clean`'i ayrıştırmak ve `-collage-build` verildiğinde sunmak yerine `<dir>`'e render etmek |

`collage build`'in `main.go`'dan hiçbir beklentisi yoktur: derlemek, `go build`'in
kendisine hiçbir şey söylenmeden yaptığı bir iştir.

İskelesi oluşturulan sürüm, önemli kısmına indirgenmiş hâliyle:

```go
func main() {
	buildFlag := flag.Bool("collage-build", false, "render the app to static files instead of serving it")
	outFlag := flag.String("out", "dist", "output directory for -collage-build")
	cleanFlag := flag.Bool("clean", false, "remove -out's existing contents before building")
	portFlag := flag.Int("port", envInt("PORT", 3000), "port to listen on (env PORT)")
	flag.Parse()

	devMode := os.Getenv("COLLAGE_DEV") == "1"

	app, err := newApp(devMode, *portFlag)
	if err != nil {
		log.Fatal(err)
	}

	// A word after the flags is a plugin's command: go run . <command>
	if args := flag.Args(); len(args) > 0 {
		code, err := collage.DispatchCommands(context.Background(), app, args)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(code)
	}

	if *buildFlag {
		if err := staticBuild(app, *outFlag, *cleanFlag); err != nil {
			log.Fatalf("static build: %v", err)
		}
		return
	}

	if err := app.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
```

`main.go`'yu yeniden yazarsanız bunların hepsini çalışır hâlde tutun; yoksa
`collage dev`, `collage export` ve plugin'lerinizin komutları projenizde işe
yarar hiçbir şey yapmaz olur.

## Plugin komutları

Bir plugin, `Host.RegisterCommand` üzerinden bir komut ekleyebilir. **`collage`
CLI onu çalıştırmaz.** CLI uygulamanızı hiç yüklemez; bu yüzden ancak plugin'leriniz
başladıktan sonra var olan bir komut onun erişimi dışındadır ve
`collage <plugin-command>` bilinmeyen bir komuttur.

Onları kendi programınız `collage.DispatchCommands` ile dağıtır; iskelesi
oluşturulan `main.go` da bunu yapar (v0.10.0'dan itibaren — daha önce iskelesi
oluşturulmuş bir proje [yukarıdaki](#the-contract-with-maingo) bloğu
kopyalayabilir). Flag'ler ayrıştırılıp uygulama kurulduktan sonra artakalan bir
kelime bir komuttur:

```go
if args := flag.Args(); len(args) > 0 {
	code, err := collage.DispatchCommands(context.Background(), app, args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	os.Exit(code)
}
```

`DispatchCommands` uygulamayı başlatır — komutlarını kaydeden şey olan her
plugin'in `Init`'ini çalıştırarak — ve ilk argümanın adlandırdığı komutu
çalıştırır. Böylece bir plugin'in komutu, programınızın bir alt komutu olarak,
flag'lerden sonra çalışır:

```sh
go run . pages
./bin/myblog -port 4000 pages
```

Çıkış kodu CLI'ninkini izler: başarı için `0`; bir başlangıç hatası, çalışıp
başarısız olan bir komut ya da `Run`'ı olmayan bir komut için `1`; nil bir app,
argüman olmaması ya da hiçbir plugin'in kaydetmediği bir kelime
(`ErrUnknownCommand`) için `2`. Sahiplenilmemiş bir kelime, yanlışlıkla başlatılmış
bir sunucu değil, bir kullanım hatasıdır. Hiçbir komut eşleşmediğinde sunmayı
tercih eden bir program `errors.Is(err, collage.ErrUnknownCommand)`'ı kontrol edip
çıkmak yerine devam edebilir. Kendi kullanım metninizi yazdırmak isterseniz
`app.Commands()` plugin'lerin kaydettiklerini listeler. Dağıtım uygulamayı başlatır,
bu da kaydı kapatır; sonraki bir `ListenAndServe` o başlatmayı yeniden kullanır.
Böyle bir komut yazmak [Plugin yazmak](/docs/writing-plugins#commands) sayfasında
anlatılıyor.
