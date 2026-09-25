---
description: collage nedir, hangi fikre dayanır ve ne zaman doğru araçtır.
reference: NewPage, NewFragment
---

# Giriş

collage, sunucuda render edilen web siteleri için yazılmış bir Go framework'üdür.
Page'leri, fragment denen küçük parçaları saran bir layout olarak yazarsınız. Her
fragment kendi verisini çeker ve kendi template'ini render eder. collage page'i bir
araya getirir, cache'ler ve içeriğiniz değiştiğinde neyi atması gerektiğini tam
olarak bilir.

collage tek bir Go programıdır. Aynı binary sitenizi sunar, form'larınıza ve API
çağrılarınıza cevap verir. İsterseniz sitenin tamamını static dosyalar olarak da
yazar. Client framework'ü yoktur, bundler yoktur, standart kütüphane dışında hiçbir
bağımlılığı da yoktur.

## Fikir

Çoğu page, birbirinden habersiz parçalardan oluşur. Bir blog yazısında yazının
kendisi, bir yazar kartı, ilgili yazıların listesi ve bir navigasyon çubuğu bulunur.
Her parça farklı bir veriye ihtiyaç duyar. Bu veriyi farklı bir yerden alır ve farklı
nedenlerle değişir.

collage'da bu parçaların her biri bir **fragment**'tir. Bir fragment; bir template,
template'in ihtiyaç duyduğu veriyi çeken bir fonksiyon ve başka fragment'lerin
yerleştiği isimli **slot**'lardan oluşur. Bir **page** ise içinde bir content
fragment'i bulunan bir layout fragment'i ve bir URL'dir.

```go
author := collage.NewFragment("author", "fragments/author.html").
	WithDataHandler(loadAuthor).
	Build()

post := collage.NewFragment("post", "pages/post.html").
	WithDataHandler(loadPost).
	WithSlotFragment("author", author).
	Build()

page := collage.NewPage("post").
	WithLayout(layout).
	WithContent(post).
	WithPath("en", "/blog/{slug}").
	Incremental(10 * time.Minute).
	Build()
```

Page'leri bu şekilde kurmanın üç sonucu vardır.

- **Veri eşzamanlı çekilir, çıktı sırayla yazılır.** Bir fragment render edilmeden
  önce, slot'larındaki bütün fragment'lerin data handler'ları aynı anda başlar.
  Birbirinden bağımsız parçalardan oluşan bir page, upstream çağrılarını birlikte
  yapar. Buna rağmen HTML her seferinde aynı çıkar.
- **Bir page neyden oluştuğunu bilir.** Her data handler hangi içerik parçalarını
  kullandığını bildirir: `post:hello-world`, `author:ada` gibi. Cache'teki page bu
  tag'leri taşır. Bir yazar değiştiğinde tek bir çağrı, o yazarı gösteren bütün
  page'leri cache'ten düşürür ve başka hiçbir şeye dokunmaz.
- **Bir parça, page'i de beraberinde götürmeden hata verebilir.** Bir fragment
  required olabilir, bir fallback'i olabilir ya da hiçbir şey render etmeyebilir.
  Bozuk bir sidebar, 500 hatası değil, yalnızca eksik bir sidebar olur.

## Neler sunar

- URL başına cache'lenen page'ler ve üç strateji. Page bir kez render edilip siz
  invalidate edene kadar tutulabilir, bir TTL sonunda yeniden render edilebilir ya da
  hiç cache'lenmeyebilir. Hiçbir veri çekmeyen bir page, söylenmesine gerek
  kalmadan bir kez render edilir.
- Page'ler arasında paylaşılan data cache. Cache açıkken aynı yazarın otuz yazısı,
  yazarın verisini yalnızca bir kez çeker.
- Kendi page'ine post edilen form'lar ve bunlara yerleşik request forgery koruması.
  Cache'lenmiş bir page de form taşıyabilir.
- Page isimlerinden üretilen link'ler. Bir page'in path'i değiştiğinde link'ler de
  sitenin bütün locale'lerinde bu değişikliği takip eder.
- Title, meta tag ve stylesheet gibi head elemanları. Bunları, page içinde nerede
  durursa dursun, onlara ihtiyaç duyan fragment tanımlar.
- Sitemap'ler, feed'ler ve `robots.txt` için HTML olmayan route'lar. Geri kalan her
  şey için kendi `http.Handler`'ınızı kullanabilirsiniz.
- Siteyi herhangi bir static hosting'e uygun dosyalara dönüştüren bir static export.
  Şu an okuduğunuz sayfalar da onunla üretildi.
- Her Go değişikliğinde yeniden build alan ve her template değişikliğinde tarayıcıyı
  yenileyen bir dev server. Render edilemeyen bir page, hatasını tarayıcıda gösterir;
  önce template'i, satırı ve nedeni verir. Derlenemeyen Go kodu hatasını terminalde
  gösterir ve bu sırada son sağlam build sunulmaya devam eder. Hiç başlayamayan bir
  program ise reddedilen bir bağlantı yerine yazdırdıklarını tarayıcıda gösterir.

## Ne zaman doğru araçtır

collage, page'leri çoğunlukla okunan sitelere uyar. Bir blog, dokümantasyon, bir
tanıtım sitesi, bir katalog, bir haber sitesi ya da bir uygulamanın herkese açık
tarafı buna örnektir. collage en iyi sonucu, içerik başka bir yerden geldiğinde
verir. Bu kaynak bir CMS, bir veritabanı ya da bir API olabilir. Page'ler de böyle
birkaç kaynaktan aynı anda bir araya getiriliyorsa collage tam yerindedir.

Ekranları çoğunlukla etkileşimden oluşan uygulamalara ise pek uymaz. Bir editör,
her saniye güncellenen bir dashboard ya da state'i tarayıcıda yaşayan herhangi bir
şey bu gruba girer. Bunlar için bir client framework'ü kullanın. İsterseniz
etraflarındaki page'ler için [`app.Handle`](/docs/middleware-and-apis) üzerinden
collage'ı kullanabilirsiniz.

## Bilerek dışarıda bıraktıkları

collage bir veritabanı katmanı, bir session store, bir ORM, bir çeviri kataloğu ya
da bir JavaScript toolchain'i ile gelmez. Go yazıyorsunuz ve Go'nun bunların hepsi
için zaten iyi çözümleri var. Sizin yerinize bunlardan birini seçen bir framework,
etrafından dolaşmanız gereken bir şey daha olurdu. collage'ın üstlendiği kısım,
page render etmeye özgü olan kısımdır: page'leri bir araya getirmek, cache'lemek ve
ne zaman stale olduklarını bilmek.

## Sırada ne var

[Kurulum](/docs/installation) bir dakikada bir projeyi çalışır hâle getirir.
[İlk page'iniz](/docs/your-first-page) sıfırdan bir page kurar.
[Page'ler ve layout'lar](/docs/pages-and-layouts) ise kavramların başladığı yerdir.
