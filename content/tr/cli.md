---
description: collage CLI'ın bütün komutları (new, dev, build, export, serve, version ve help), flag'leri ve her birinin tam olarak neyi çalıştırdığı.
reference: DispatchCommands, Command, ErrUnknownCommand
---

# collage CLI

`collage` komutu yeni projeleri scaffold eder ve var olan projelerinizi yönetir.
Bir projeyi development modunda çalıştırır, deploy edeceğiniz binary'yi derler ve
projeyi static dosyalar olarak export eder.

```sh
go install github.com/Elagoht/collage/cmd/collage@latest
```

CLI, uygulamanızı kendi içine asla link etmez. Zaten edemez de, çünkü uygulamanız
sizin kodunuzdur. `dev`, `build` ve `export` komutları, `go` aracını bulunduğunuz
dizinde tıpkı elle çalıştıracağınız gibi çalıştırır. Bu sayfanın geri kalanı her
komutun tam olarak neyi çalıştırdığını anlatır.

## Kullanım

```sh
collage <command> [flags]
```

| Komut | Ne yapar |
| --- | --- |
| `new` | Yeni bir collage projesi scaffold eder |
| `dev` | Bulunduğunuz dizindeki projeyi development modunda çalıştırır |
| `build` | Bulunduğunuz dizindeki projeyi deploy edeceğiniz binary'ye derler |
| `export` | Bulunduğunuz dizindeki projeyi static dosyalara render eder |
| `serve` | Bir static export'u, bir static host'un sunacağı şekilde sunar |
| `version` | collage CLI'ın sürümünü yazdırır |
| `help` | Bir komutun yardımını gösterir ya da bütün komutları listeler |

`collage help <command>` o komutun kendi kullanım bilgisini yazdırır.
`collage <command> -h` de aynı şeyi yapar. Flag'ler Go'nun `flag` sözdizimini
kullanır: `-out dist` ile `-out=dist` aynıdır, `--out` da çalışır.

### Çıkış kodları

| Kod | Anlamı |
| --- | --- |
| `0` | Başarılı. `help`, `collage -h` (v0.11.0'dan beri) ve `collage <command> -h` de bu kodla çıkar |
| `1` | Komut doğru parse edildi ama işini yapamadı |
| `2` | Kullanım hatası: komut verilmedi, komut bilinmiyor, flag hatalı ya da beklenmeyen bir argüman var |

## collage new

```sh
collage new <name> [--template demo|minimal] [--dir path] [--module path] [--force]
```

`<name>` adında, çalıştırılmaya hazır yeni bir proje scaffold eder.

| Flag | Varsayılan | Anlamı |
| --- | --- | --- |
| `--template name` | `demo` | Scaffold edilecek proje: `demo` ya da `minimal` (v0.14.2'den beri; `-minimal`'in yerini alır) |
| `-dir path` | `./<name>` | Projenin scaffold edileceği dizin |
| `-module path` | `<name>` | `go.mod`'a yazılan module path |
| `-force` | kapalı | Dizin boş olmasa da scaffold eder |

```sh
collage new myblog                                   # into ./myblog, module "myblog"
collage new myblog --template minimal                # one page, nothing to delete
collage new myblog -module github.com/me/myblog
collage new myblog -dir . -force                     # into the current, non-empty directory
```

Flag'leri tek ya da çift tireyle yazabilirsiniz. Flag'ler addan önce de sonra da
gelebilir. Tam olarak bir ad vermeniz gerekir. Hiç ad vermemek ya da birden fazla
ad vermek kullanım hatasıdır. Hedef dizin varsa ve boş değilse, `-force`
vermediğiniz sürece komut reddedilir. `-force` verdiğinizde ise scaffold'un
yazdığı dosyalar, dizindeki aynı adlı dosyaların yerine geçer.

**Her projede** şunlar bulunur: `go.mod`; config'i, static mount'u,
[aşağıda](#the-contract-with-maingo) anlatılan CLI sözleşmesini ve
[plugin komutlarının](#plugin-commands) dispatch'ini içeren bir `main.go`; bütün
route'ları register eden bir `routes.go`; sitenin başlığını `rc.HoistTitle` ile
tanımlayan bir layout (böylece bir page'in kendi başlığı bu başlığın yerine
geçer); bir ana sayfa, `static/`, bir `.gitignore` ve bir README.

**Demo projesi** varsayılan projedir. Buna canlı demolardan oluşan bir page ekler:
JSON dönen bir action, kendi page'ine post eden bir form, kendi URL'si olan bir
fragment ve bir JSON document. Bunlar `pages/`, `fragments/`, `actions/`,
`documents/` ve `store/` dizinlerine dağıtılmıştır. Demo projesi ayrıca bir
not-found page'i, her biri için testleri, `plugins-config.json`'ı, bir favicon'u
ve bir `.env.example`'ı da ekler.

**`--template minimal`** bir projenin olabileceği en yalın hâldir. İçinde tek bir
page'i saran layout vardır. Bu page `<h1>Hello from {{.Name}}</h1>` satırından
ibarettir; `{{.Name}}`, template'e `collage.Data` ile verilen proje adıdır.
Bunun yanında arka plan ve metin rengini dark mode dahil ayarlayan bir stylesheet
bulunur. Başka hiçbir şey yoktur: test yoktur, not-found page'i de yoktur. Siz bir
not-found page'i register edene kadar collage bilinmeyen adreslere kendi sade 404'üyle
cevap verir.

İş bittiğinde sonraki adımları yazdırır. `cp` satırı yalnızca demo projesinde,
yani `.env.example` dosyası olan projede yazdırılır:

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

Bulunduğunuz dizindeki projeyi build eder ve `COLLAGE_DEV=1` ayarlanmış olarak
çalıştırır. Go kodu her değiştiğinde projeyi yeniden build edip yeniden başlatır.
Hiçbir flag ya da argüman almaz. Durdurmak için Ctrl-C'ye basın. Interrupt
sinyali programa da ulaşır ve program production'da nasıl kapanıyorsa öyle
kapanır.

Scaffold edilen `main.go`, `COLLAGE_DEV=1` ayarlı olduğunda development modunu
açar. Development modu template'leri ve static dosyaları her request'te diskten
okur. Bu yüzden onları düzenlediğinizde rebuild gerekmez ve rebuild yapılmaz;
bunun yerine tarayıcınızdaki page kendini yeniler. Programınızın diskten kendisi
okuduğu içerik de (örneğin Markdown), dizini
[`Config.DevWatch`](/docs/configuration#devwatch) içinde belirtildiğinde page'i
yeniler (v0.10.0'dan beri). Development modunun başka neleri değiştirdiğini
[Template'ler](/docs/templates#reloading-in-development) sayfasında
bulabilirsiniz.

### Rebuild'i ne tetikler

| İzlenir | İzlenmez |
| --- | --- |
| projenin herhangi bir yerindeki `.go` dosyaları | `_test.go` dosyaları |
| proje kökündeki `go.mod` ve `go.sum` | template'ler, static dosyalar ve `DevWatch` dizinleri (rebuild olmadan yeniden yüklenir) |
| aşağıda anlatılan ortam dosyası | gizli dizinler (`.git`, `.cache`, …) |
| | `bin`, `dist`, `node_modules`, `testdata`, `vendor` |

Atlanan dizinler, döngünün kendi kendini beslemesini engeller. Çalışan programın
yazdığı hiçbir şey (cache'i ya da bir export) rebuild tetikleyemez.

Rebuild şöyle ilerler:

- **Polling yapar.** Her 300 ms'de bir dosyaların değişiklik zamanlarını ve
  boyutlarını karşılaştırır. Dosya sistemi bildirimleri için bir kütüphane
  kullanmaz ve her platformda aynı şekilde davranır.
- **Art arda gelen yazmalar tek bir rebuild'dir.** Bir değişiklikten sonra,
  dosyalar bir polling aralığı boyunca değişmeyene kadar bekler. Böylece birkaç
  dosyayı yeniden yazan bir formatter yalnızca bir build'e yol açar.
- **Önce yeni build yapılır.** `go build`, binary'yi proje dışındaki geçici bir
  dizine yazar. Eski process ancak yeni binary derlendikten sonra durdurulur ve
  yenisi başlatılır. Eski process, Ctrl-C'de olduğu gibi elindeki request'leri
  bitirir. 10 saniye içinde çıkmazsa öldürülür.
- **Derlenmeyen bir değişiklikte son sağlam build çalışmaya devam eder.**
  Derleyicinin hatası ekranda görünür.
- **Kendiliğinden çıkan bir program** döngü içinde yeniden başlatılmaz. Örneğin
  açılışta panic olabilir ya da port zaten kullanımda olabilir. Program bir sonraki
  değişiklikte yeniden başlatılır.

### Ortam dosyaları

`collage dev`, bulunduğunuz dizindeki `.env.development` dosyasının
değişkenlerini programın ortamına ekler. `.env.development` yoksa `.env`
dosyasını kullanır. Her zaman tek bir dosya okunur, asla ikisi birden okunmaz:
`.env.development`, `.env`'in üzerine merge edilmez, onun yerine geçer.

```sh
# .env.development
PORT=3000
HOST=localhost
export COLLAGE_CSRF_KEY="0f1e2d3c4b5a69788796a5b4c3d2e1f00f1e2d3c4b5a69788796a5b4c3d2e1f0"
```

- Dosya `KEY=value` satırlarından, boş satırlardan ve `#` ile başlayan
  satırlardan oluşur. Satırın başında `export ` olabilir. Bir değer, eşleşen tek ya
  da çift tırnak içine de alınabilir. Tırnaklar kaldırılır ve içlerindeki hiçbir
  şey yorumlanmaz.
- Tırnaksız bir değer, boşluktan sonra gelen ilk `#` işaretinde biter. Yani
  `PORT=3000 # dev` satırının değeri `3000`'dir.
- Key'ler harf, rakam ve alt çizgiden oluşur ve rakamla başlayamaz.
- **Hatalı bir satır programın başlamasını engeller.** `collage dev` hatayı dosya
  adı ve satır numarasıyla (`.env.development:3`) bildirir. Satır düzeltilene
  kadar hiçbir şeyi başlatmaz ya da yeniden başlatmaz. Zaten çalışan bir program,
  başlatıldığı değerlerle çalışmaya devam eder. `collage dev` izlemeyi sürdürür;
  düzeltilmiş dosyayı kaydettiğinizde kaldığı yerden devam eder. Hatalı satırı
  atlamak, yazdığınız bir ayarın programa hiç ulaşmaması demek olurdu.
- **Shell'de zaten ayarlı olan bir değişken dosyadakine göre önceliklidir.** Bu
  yüzden `PORT=4000 collage dev` yine çalışır.
- Dosyada ne yazarsa yazsın `COLLAGE_DEV=1` her zaman ayarlanır.
- Dosyanın olmaması hata değildir. Bir dosya okunduğunda adı stderr'e yazdırılır.
- Dosya her yeniden başlatmada tekrar okunur. Dosya izlendiği için onu
  düzenlediğinizde program yeni değerlerle yeniden başlar.

Bu dosyaları yalnızca `collage dev` okur. `collage build`, `collage export` ve
derlenmiş binary, ortam değişkenlerini çalıştıkları ortamdan alır.

## collage build

```sh
collage build [-o path] [-os name] [-arch name] [-i]
```

Bulunduğunuz dizindeki projeyi deploy edeceğiniz binary'ye derler. Çalıştırdığı
komut şudur:

```sh
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o bin/<name> .
```

| Flag | Varsayılan | Anlamı |
| --- | --- | --- |
| `-o path` | `bin/<name>` | Binary'nin yazılacağı yer |
| `-os name` | `linux` | Hedef işletim sistemi (`GOOS`) |
| `-arch name` | `amd64` | Hedef mimari (`GOARCH`) |
| `-i` | kapalı | Binary'nin yanına hangi ek dosyaların yazılacağını sorar |

`<name>`, `go.mod`'daki module path'in son parçasıdır. Dizinde `go.mod` yoksa ya da
`go.mod` bir module tanımlamıyorsa komut hata verir. `-os windows` verildiğinde,
path zaten `.exe` ile bitmiyorsa sonuna `.exe` eklenir. Positional argümanlar
kullanım hatasıdır.

Bu ayarların nedenleri:

- **CGO kapalıdır**, çünkü ne collage ne de standart kütüphane C'ye ihtiyaç duyar.
  Static bir binary, içinde başka hiçbir şey olmayan bir image'a konabilir.
- **`-trimpath`** kullanılır, böylece binary onu derleyen makinedeki path'leri
  içermez.
- **`-s -w`** debug tablolarını çıkarır. Boyutun büyük kısmı bu tablolardır.
- **Varsayılan hedef**, üzerinde çalıştığınız makine değil **linux/amd64**'tür.
  Mac'te derlenen bir binary Linux container'ında çalışmaz. Bunu sunucuda
  "exec format error" alarak öğrenmek istemezsiniz.
- **Binary `dist/` yerine `bin/` dizinine yazılır.** `dist/`, `collage export`'un
  yazdığı yerdir ve `export -clean` onu boşaltır.

İş bittiğinde binary'nin path'ini, platformunu, boyutunu ve build süresini
yazdırır.

### -i ile ek dosyalar

`-i` verildiğinde build'den önce iki soru sorar ve cevapları standart input'tan
okur. `y` ya da `yes` evet anlamına gelir, diğer her şey hayır sayılır:

- **Dockerfile yazılsın mı?** İki aşamalı bir image'dır. İlk aşama bir `golang`
  build aşamasıdır ve CLI'ı derleyen Go'nun major ve minor sürümüne sabitlenir.
  İkinci aşama `gcr.io/distroless/static-debian12` tabanlıdır ve yalnızca binary'yi
  içerir. Bu aşamada `HOST=0.0.0.0` ve `PORT=8080` ayarlanır, 8080 portu expose
  edilir ve `COLLAGE_CSRF_KEY` yorum satırı olarak bulunur.
- **systemd unit yazılsın mı?** `/usr/local/bin/<name>` binary'sini `/srv/<name>`
  dizininden çalıştıran bir `<name>.service` dosyasıdır. İçinde `HOST=127.0.0.1`,
  `PORT=8080`, `Restart=on-failure` ve `TimeoutStopSec=30` vardır. Kurmadan önce
  kendi ihtiyacınıza göre düzenlemeniz gerekir.

İki dosya da binary'nin yanına yazılır: `bin/Dockerfile` ve `bin/<name>.service`.
Bunun nedeni, bu dosyaların üretilmiş olmasıdır; proje kökü ise insanların
yazdığı dosyalar içindir. Bu yerleşimin tek bedeli farklı bir komuttur ve rapor o
komutu yazdırır:

```sh
docker build -f bin/Dockerfile .
```

**Var olan bir dosyanın üzerine asla yazılmaz.** Dosya zaten varsa ilgili soru
sorulmaz. Bunlar projelerin düzenlediği dosyalardır. Bu dosyaları varsayılan
hâliyle değiştiren bir build komutu, birinin emeğini sessizce silmiş olurdu.

`-i` olmadan yalnızca binary yazılır. Ayrıntılar için
[Deployment](/docs/deployment) sayfasına bakın.

## collage export

```sh
collage export [-out dir] [-clean]
```

Bulunduğunuz dizindeki projeyi bir static host için static dosyalara render eder.
HTML olabilen her page için HTML üretir, mount edilmiş bütün asset'leri de çıktıya
ekler. Çalıştırdığı komut şudur:

```sh
go run . -collage-build -out <dir>          # plus -clean when you passed it
```

| Flag | Varsayılan | Anlamı |
| --- | --- | --- |
| `-out dir` | `dist` | Projenin render edildiği dizin |
| `-clean` | kapalı | Build'den önce dizinin mevcut içeriğini siler |

Programın çıktısı, yani build raporu, yeniden biçimlendirilmeden olduğu gibi
aktarılır. Positional argümanlar kullanım hatasıdır.

Neyin export edildiği, neyin hangi nedenle atlandığı ve çıktı dizini için yapılan
güvenlik kontrolleri [Static export](/docs/static-export) sayfasında anlatılır.
Form'ları ya da her request'te üretilen page'leri olan bir site için ihtiyacınız
olan komut `collage build`'dir.

## collage serve

```sh
collage serve [-dir dir] [-host name] [-port n]
```

Bir static export'u, bir static host'un sunacağı şekilde sunar. Böylece
burada gördüğünüz, deploy ettikten sonra göreceğinizle aynıdır. Yalnızca dosya
sunar, projenizi çalıştırmaz.

| Flag | Varsayılan | Anlamı |
| --- | --- | --- |
| `-dir dir` | `dist` | Sunulacak dizin |
| `-host name` | `localhost` | Dinlenecek interface |
| `-port n` | `4000` | Dinlenecek port |

Port 3000 değil 4000'dir, böylece `collage dev` ile yan yana çalışabilir. İkisini
karşılaştırmak istediğiniz an da tam olarak budur.

Sıradan bir dosya sunucusu gibi değil, bir static host gibi davranır:

- Uzantısı olmayan bir path'e `<path>/index.html` ile cevap verilir. Export'un
  yazdığı yapı da budur.
- Dizinler asla listelenmez.
- Hiçbir dosyaya karşılık gelmeyen bir path'e, export'un kendi `404.html`
  dosyası ve 404 status koduyla cevap verilir. Bu dosya yoksa sade bir 404 döner.
- Yalnızca `GET` ve `HEAD` request'lerine cevap verilir. Diğer her şey 405 alır.
- Her response `Cache-Control: no-store` header'ıyla gönderilir. Böylece yeniden
  export edip sayfayı yenilediğinizde eski çıktıyı değil yenisini görürsünüz.

Dizin yoksa ya da içinde hiç dosya yoksa komut hata verir ve önce
`collage export` çalıştırmanızı söyler. Path varsa ama dizin değilse de hata
verir ve tam olarak bunu söyler.

## collage version

```sh
collage version
```

`collage version <version>` yazdırır. Sürüm, binary'nin build bilgisinden okunur.
Bu yüzden `go install ...@v0.9.0` ile kurulan binary `0.9.0` bildirir. Bir git
checkout'unda derlenen binary ise bir pseudo-version bildirir. Örneğin v0.11.0
tag'inden sonraki bir commit için bu `0.11.1-0.<timestamp>-<commit>` olur.
Yalnızca hiç sürüm bilgisi olmayan bir binary `devel` bildirir. Bu, bir
repository dışında ya da `-buildvcs=false` ile derlenmiş bir binary'dir.

## collage help

```sh
collage help            # every command
collage help export     # one command's usage
```

Argümansız çalıştırıldığında bütün komutları listeler ve `0` ile çıkar. Bir komut
adı verildiğinde o komutun kullanım bilgisini yazdırır. Bilinmeyen bir ad
verildiğinde `collage: unknown command: "nope"` yazdırır ve `2` ile çıkar.

## main.go ile sözleşme

`dev` ve `export`, `main.go`'nuzun yaptığı iki şeye dayanır. Scaffold edilen
`main.go` ikisini de yapar. Bunlara ek olarak üçüncü bir şey daha yapar: plugin
komutlarını çalıştırır. Hiçbir `collage` komutu buna ihtiyaç duymaz ama
plugin'leriniz duyar:

| Komut | Çalıştırdığı | `main.go`'nuzun yapması gereken |
| --- | --- | --- |
| `collage dev` | önce `go build`, sonra `COLLAGE_DEV=1` ile binary | `COLLAGE_DEV` `1` olduğunda development modunu açmak |
| `collage export` | `go run . -collage-build -out <dir> [-clean]` | `-collage-build`, `-out` ve `-clean` flag'lerini parse etmek; `-collage-build` verildiğinde sunmak yerine `<dir>` dizinine render etmek |

`collage build`'in `main.go`'dan hiçbir beklentisi yoktur. Derleme, `go build`'in
kendisine hiçbir şey söylenmeden yaptığı bir iştir.

Scaffold edilen sürümün, yalnızca önemli kısmı bırakılmış hâli:

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

`main.go`'yu yeniden yazarsanız bunların hepsinin çalışmaya devam ettiğinden emin
olun. Aksi hâlde `collage dev`, `collage export` ve plugin'lerinizin komutları
projenizde işe yarar hiçbir şey yapmaz.

## Plugin komutları

Bir plugin, `Host.RegisterCommand` ile bir komut ekleyebilir. **`collage` CLI bu
komutu çalıştırmaz.** CLI uygulamanızı hiçbir zaman yüklemez. Bu yüzden ancak
plugin'leriniz başladıktan sonra var olan bir komuta CLI erişemez ve
`collage <plugin-command>` bilinmeyen bir komut olarak kalır.

Bu komutları kendi programınız `collage.DispatchCommands` ile dispatch eder.
Scaffold edilen `main.go` da bunu yapar (v0.10.0'dan beri; daha önce scaffold
edilmiş bir proje [yukarıdaki](#the-contract-with-maingo) bloğu kopyalayabilir).
Flag'ler parse edilip uygulama kurulduktan sonra geriye kalan bir kelime komut
olarak yorumlanır:

```go
if args := flag.Args(); len(args) > 0 {
	code, err := collage.DispatchCommands(context.Background(), app, args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	os.Exit(code)
}
```

`DispatchCommands` uygulamayı başlatır. Bunu yaparken her plugin'in `Init`'ini
çalıştırır; plugin'ler komutlarını `Init` içinde register eder. Ardından ilk argümanın
adını taşıyan komutu çalıştırır. Böylece bir plugin'in komutu, programınızın bir
alt komutu olarak ve varsa flag'lerden sonra çalışır:

```sh
go run . pages
./bin/myblog -port 4000 pages
```

Çıkış kodları CLI'ınkilerle aynı mantığı izler. Başarı için `0` döner. Açılış
hatası, çalışıp başarısız olan bir komut ya da `Run`'ı olmayan bir komut için `1`
döner. Nil bir app, argüman verilmemesi ya da hiçbir plugin'in register etmediği bir
kelime (`ErrUnknownCommand`) için `2` döner. Hiçbir komuta karşılık gelmeyen bir
kelime, yanlışlıkla başlatılmış bir sunucuya değil, kullanım hatasına yol açar.
Hiçbir komut eşleşmediğinde sunucuyu başlatmayı tercih eden bir program,
`errors.Is(err, collage.ErrUnknownCommand)` ile kontrol edip çıkmak yerine devam
edebilir. Kendi kullanım metninizi yazdırmak isterseniz `app.Commands()`
plugin'lerin register ettiği komutları listeler. Dispatch işlemi uygulamayı başlatır
ve bu da register aşamasını kapatır. Sonradan çağrılan `ListenAndServe` aynı başlatmayı
yeniden kullanır. Böyle bir komutun nasıl yazıldığı
[Plugin yazmak](/docs/writing-plugins#commands) sayfasında anlatılır.
