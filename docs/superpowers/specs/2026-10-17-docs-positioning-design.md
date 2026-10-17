# Dokümantasyonun konumlandırılması, anasayfa ve öğretici yeniden tasarımı

## Neden

Anasayfa ve Giriş sayfası Collage'ı "önbelleklenen statik fragment'lardan sayfa
kuran bir araç" olarak anlatıyor. Furkan'ın Collage'ı yazma sebebini anlatmıyor.
Dokümanlar eksiksiz ama okuması yorucu: uzun düzyazı, az şema, az kod, gerçek
kullanım örneği yok. İlk kez gelen biri de önce içinde sadece `/en/` linki olan
boş bir sayfa görüyor.

## Anlatılacak felsefe (Furkan'ın sözleri, kaynak bu)

- Landing page'ler, ürün tanıtım sayfaları, birkaç sayfalık ve etkileşimi az
  siteler; devasa JavaScript runtime'ı sunucuları ve overengineer edilmiş client
  framework'leri gerektirmez. Böyle bir runtime Go'dan yavaştır; sunucuda çok daha fazla RAM
  ve disk tüketir.
- Biraz etkileşim için bütün projeyi bir client framework'üne çevirmeye gerek yok. HTML işin
  çoğunu zaten görür, kalanı vanilla JS görür. Sunucuda değişen bir şey varsa
  sadece o kısım çekilip HTML'e yerleştirilir: fetch/XHR, SSE ve WebSocket tam
  da bunun için var.
- Collage sunucudaki state'i fragment fragment sunar ve günceller. Formları
  tarayıcı yönetir. Güvenlik önlemleri (CSRF, honeypot…) düşünülmüştür.
- Tek middleware dosyası sunan framework'lerde her şey o dosyaya sıkıştırılıyor.
  Collage'da istediğin kadar middleware ve istediğin kadar plugin eklenebilir.
- Çekirdek minimal; geri kalan her şey plugin'le eklenir.
- Hâlâ Go yazıyorsun. Ayrı bir backend projen varsa handler'larını doğrudan
  Collage'ın HTTP mux'una takabilirsin.
- **Konumlandırma:** frontend odaklı bir sunucu framework'ü. Batteries-included
  değil: ORM, migration ve kuyruk yok. Action'larla backend olma potansiyeli var
  ama amacı bu değil. Data katmanı Go ile yazıldığında her şey tek bir projede,
  tek bir binary'de birleşir.
- Sadece web sitesi değil, web uygulaması da yazılabilir: çok kullanıcılı, rol
  bazlı bir web-app Collage ile geliştirilebilir (session, honeypot, CSRF,
  formlar, canlı fragment'lar). Bu, genel bir yetenek olarak söylenir; belirli
  bir proje, benzetme veya proje ayrıntısı verilmez.
- Yalnızca "statik cache'lenen fragment" anlatılmayacak. Öne çıkacaklar:
  component yerine fragment, cache + tag tabanlı invalidation + purge, zengin
  plugin desteği.

## Hedefler

1. İlk ziyaretçi `/` adresinde boş bir yönlendirme sayfası görmez.
2. Anasayfa, yukarıdaki felsefeyi ilk ekranda ve sırasıyla anlatır.
3. Giriş sayfası aynı konumlandırmayı ayrıntılandırır ve onunla çelişmez.
4. Pluginsiz, sadece `WithFragmentPath` + vanilla `fetch` ile sayfadaki bir
   alanı güncelleyen, baştan sona yazılan bir öğretici ve gerçekten çalışan
   örnek bir proje olur.
5. Mevcut sayfalar şema, daha çok kod ve gerçek dünya senaryolarıyla daha
   okunur hale gelir; hiçbir bilgi kaybolmaz.

## Hedef olmayanlar

- Collage çekirdeğinde yeni bir özellik geliştirmek (kök yönlendirme sorusu
  hariç, aşağıya bkz.).
- Sitenin görsel kimliğini sıfırdan değiştirmek: mevcut tipografi, renkler ve
  `d-*` şema sistemi korunur.
- Cookbook bölümü eklemek tek başına yeterli sayılmaz. Öğretici ve örnek proje
  şart.

## Anti-hedefler (teknik olarak çalışsa bile başarısızlık sayılır)

- Siteye istemci tarafı diyagram kütüphanesi veya başka büyük bir JS kütüphanesi eklemek.
  Sayfa "client framework'üne ihtiyacınız yok" derken 2 MB JS yüklemiş olur.
- Uydurma performans rakamları. JavaScript runtime'larıyla kıyas genel ifadelerle yapılır. Rakam
  verilecekse bu sitenin kendi binary'sinden ölçülür (boyutu, boştaki RSS) ve
  nasıl ölçüldüğü yazılır.
- Plugin özelliklerini çekirdek özelliği gibi sunmak. Session, honeypot, live
  (SSE), websocket ve morph plugin'dir. CSRF, fragment path, cache/tag/purge,
  action ve `app.Handle` çekirdektir. Metin bu ayrımı açıkça yapar; "küçük
  çekirdek + plugin" mesajı da zaten bunu gerektirir.
- Başka bir framework, kütüphane ya da ürünün adını anmak; belirli bir projeye
  atıf yapmak ya da gerçek bir uygulamanın
  ayrıntılarını (kullanıcı sayısı, kullandığı plugin'ler) anlatmak.
- Eski davranışı veya geçmişteki geçici çözümleri anlatmak. Metin her zaman
  bugünkü davranışı anlatır.
- Go örneklerinde `any` kullanmak (Furkan'ın genel kuralı, dokümandaki kod
  parçaları dahil).

## Kısıtlar (repo kuralları)

- `elagoht/i18n` `strict: true` çalışır: her yeni anasayfa anahtarı hem
  `locales/en.json` hem `locales/tr.json` içinde olmalı.
- Yeni bir doküman sayfası her iki `nav.json` dosyasına girer. TR dosyası,
  orijinalin başlıklarını seviye seviye aynen izler, yoksa testler düşer.
- Linkler `/docs/<slug>#<id>` biçiminde olur; kırık link testleri düşürür.
- TR değişikliklerinden sonra `turkish-reviewer` agent'ı çalıştırılır.
- `go test ./...` ve `go run . -collage-build -out dist -clean` geçmelidir.

## Kararlar

### 1. Kök `/` sayfası

Durum: `PrefixDefault: true` olduğu için export, `dist/index.html`'e
`meta refresh` içeren çıplak bir sayfa yazıyor (`collage` v0.64.0,
`internal/build/builder.go` `writeRootRedirect`). Bunu builder her seferinde
kendisi yazıyor; uygulama bu dosyayı değiştiremiyor. Dev/serve modunda `/` 301
döndüğü için sorun yalnızca export'ta (GitHub Pages'te) görünüyor.

`PrefixDefault` kapatılmaz; kapatılırsa bütün URL'ler değişir ve dışarıdan
gelen linkler kırılır.

**Karar (Furkan): düzeltme Collage'da yapılır** (`~/Desktop/collage`,
`github.com/Elagoht/collage`). Böylece `PrefixDefault` kullanan her site düzelir.
Yönü:
- Kök sayfa artık çıplak bir yönlendirme değildir. Builder, varsayılan dilin
  anasayfasını (`/<default>/` olarak zaten render edilmiş sayfa) kökte de
  yazar. O sayfanın canonical'ı zaten `/<default>/`'i gösterdiği için arama
  motorları tek bir adresi görür. Kök anasayfanın build edilmediği durumda
  bugünkü yönlendirme sayfası yedek olarak kalır; ama anlık
  `location.replace` ve düzgün bir gövdeyle gelir.
- Kesin davranış (örneğin opsiyon mu, varsayılan mı; dil algılama scripti
  çekirdekte mi plugin'de mi) Collage'ın kendi kuralları ve testleriyle
  builder tarafından belirlenir. Değişiklik Collage'ın CHANGELOG'una ve kendi
  dokümanına (bu sitedeki `static-export` ve `links-and-locales`) yansır.
- Bu repo yeni Collage sürümüne geçirilir; `dist/index.html`'in artık
  anasayfa olduğu bir testle doğrulanır.

### 2. Şemalar: satır içi SVG, diyagram kütüphanesi değil (benim kararım)

Neden: Hazır bir diyagram kütüphanesi ya istemcide büyük bir JS kütüphanesi ister
(anti-hedef) ya da JavaScript runtime'ı gerektiren bir build adımı ister (felsefeyle çelişir). Anasayfada zaten
temaya duyarlı bir SVG şema sistemi var (`static/app.css` içindeki `d-*`
sınıfları). Goldmark `WithUnsafe` ile çalıştığı için ham HTML Markdown içinde
geçerli.

Uygulama kuralları:
- Şema bir `<figure>` içinde durur, `<svg role="img" aria-labelledby=…>` ve
  `<figcaption>` taşır. Bilgi taşıdığı için `aria-hidden` kullanılmaz.
- Goldmark'ın ham HTML bloğu boş satırda bittiği için SVG bloklarının içinde
  boş satır bırakılmaz.
- SVG içindeki etiketler TR kopyada çevrilir.
- Dar ekranda okunabilmesi için ya `viewBox` dar tasarlanır ya da anasayfadaki
  `only-wide`/`only-narrow` ikilisi kullanılır.
- Gerekirse şema sınıfları anasayfaya özel CSS'ten ortak bir bölüme taşınır.

### 3. Anasayfa mesaj sırası

Furkan onayladı:

1. **Hero:** "Sitenizin client framework'üne ihtiyacı yok." HTML + biraz vanilla JS; tek
   bir Go binary'si. Kurulum komutları ve wordmark kalır.
2. **Component değil, fragment:** sunucudaki state parça parça render edilir;
   bir parça kendi URL'inden çekilip yerine konur. Kısa bir kod + şema
   (sayfa → fragment path → `fetch` → yalnızca o alan değişir).
3. **Cache, tag'ler, invalidation, purge:** mevcut invalidate şeması kalır,
   metni güçlendirilir.
4. **Sınırsız middleware ve plugin:** tek middleware dosyasıyla çalışan
   framework'lerle, isim verilmeden yapılan karşılaştırma. Gerçek plugin listesinden bir ızgara gösterilir (session,
   honeypot, live, websocket, validate, flash, i18n, meta, sitemap, search…).
5. **Hâlâ Go yazıyorsun:** mevcut `http.Handler`'larını `app.Handle` ile
   Collage mux'una takma örneği; data katmanı ile arayüz tek projede.
6. **Küçük çekirdek, web sitesi de web uygulaması da:** "batteries-included
   değil" dürüstçe söylenir (ORM yok, migration yok). Yanında, çok kullanıcılı
   ve rol bazlı web uygulamalarının da yazılabildiği söylenir (session, CSRF,
   honeypot, action'lar, canlı fragment'lar). Belirli bir proje anılmaz.
7. **Geliştirme ve dağıtım:** mevcut dev/build şemaları kalır.
8. **Rehber haritası:** mevcut haliyle kalır.

Mevcut `home.points.*` ve `home.diagrams.*` anahtarları bu sıraya göre yeniden
düzenlenir, gerektiğinde yeni anahtarlar eklenir (iki dilde).

### 4. Giriş sayfası (`introduction`)

Anasayfayla aynı sırayla yeniden yazılır. Mevcut metindeki iki ifade
konumlandırmayla çelişiyor ve düzeltilir:
- "Etkileşimin yoğun olduğu uygulamalar için uygun değil" → session, CSRF,
  honeypot ve live plugin'leriyle uygulama yazılabildiği; asıl uygun olmadığı
  yerin state'i tamamen tarayıcıda yaşayan ekranlar (editör, canvas) olduğu
  anlatılır.
- "Session store yok" → çekirdekte yok, `elagoht/session` plugin'i ile gelir.

Eklenecekler: "Ne değil" bölümü (ORM, migration ve kuyruk yok; data katmanı
senin Go kodun), bir mimari şema (tarayıcı ↔ Collage binary'si: sayfalar,
fragment path'leri, action'lar, `app.Handle` ile takılan handler'lar,
middleware/plugin katmanı, senin data katmanın) ve gerçek dünya senaryoları
(landing page, katalog, doküman sitesi, çok kullanıcılı rol bazlı bir web-app).

### 5. Öğretici ve örnek proje

- Yeni doküman sayfası, "Getting started" bölümünde `your-first-page`'den
  sonra gelir. Çalışma adı: `live-region` / "Bir alanı sunucudan güncellemek".
  Kesin adı builder seçer.
- Pluginsiz, sadece çekirdek kullanılır: bir sayfa, `WithFragmentPath` ile
  açılan bir fragment ve bunu çekip `replaceWith`/`innerHTML` ile yerine koyan
  birkaç satır vanilla JS. Bunun yanına aynı fragment'ı form + action ile
  güncelleyen bir adım eklenir (CSRF'in nasıl taşındığı gösterilir).
  JS kapalıyken çalışan hali, yani tam sayfa render'ı, de gösterilir.
- İstemci kodu, `forms-and-actions#a-fragment-at-its-own-url` bölümünde
  belgelenen protokole (`FetchHeader` vb.) ve Collage kaynağına bakılarak
  yazılır, ezberden yazılmaz.
- Morph, SSE ve websocket sonda "bir adım ötesi" olarak plugin'lere
  (`elagoht/live`, `elagoht/websocket`) link verir; çekirdekmiş gibi
  sunulmaz.
- Örnek proje `examples/` altında, `go test ./...`'in derleyip en az bir
  testle (`collagetest`) fragment path'ini çağırdığı bir yerde durur. Böylece
  CI onu doğru tutar. Ayrı bir Go modülü mü yoksa bu modülün bir paketi mi
  olacağına builder karar verir; ölçüt `go test ./...`'in onu kapsamasıdır.
- Doküman sitesi statik olduğu için öğretici "bunu yerelde çalıştır" der;
  sayfada canlı demo vaat edilmez.

### 6. Mevcut sayfaların gözden geçirilmesi

Her sayfa için: önce bir şema (sayfanın ana kavramı), her kavram için bir
"gerçek dünya" senaryosu (blog, ürün kataloğu ve çok kullanıcılı bir
web-app gibi tekrar eden 2–3 örnek alan; hepsi kurgusal), düzyazı yerine kısa paragraflar ve daha çok kod.
Bilgi silinmez, sadece yeniden düzenlenir. Öncelik sırası:
`fragments-and-slots`, `caching`, `forms-and-actions`, `middleware-and-apis`,
`plugins`, sonra geri kalanı. Her sayfa EN ve TR birlikte güncellenir.

## Aşamalar (her biri ayrı teslim edilebilir; sıra Furkan tarafından onaylandı)

1. Kök `/` düzeltmesi (Collage repo'sunda, sonra bu repo yeni sürüme geçer).
2. Anasayfa + Giriş sayfası konumlandırması (EN + TR).
3. Öğretici + örnek proje (EN + TR).
4. Mevcut sayfaların şema ve örnek geçişi; sayfa sayfa, öncelik sırasıyla.

## Benim kararlarım (Furkan onayına açık)

- Şemalar diyagram kütüphanesi yerine satır içi SVG.
- Örnek proje `examples/` altında ve `go test ./...` kapsamında.
- Kök düzeltmesinin yönü: kökte varsayılan dilin anasayfası, yedek olarak
  anlık yönlendirme.

## Builder'a bırakılanlar

- Şemaların tam görsel tasarımı (mevcut `d-*` dili içinde).
- Öğretici sayfasının slug'ı ve örnek projenin konusu (sayaç, stok durumu,
  bildirim sayısı gibi küçük bir şey).
- Metnin kelimesi kelimesine yazımı, yukarıdaki sıra ve kurallar içinde.
