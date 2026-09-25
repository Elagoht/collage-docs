// Package ui holds the words of the site that are not documentation: the header,
// the page furniture, the home page. The documentation itself is content, in
// content/ and its translations; these are the few strings around it.
package ui

// Text is every string the templates show, in one language.
type Text struct {
	// Name is the language's name in itself, for the language switcher.
	Name string

	Docs              string
	GoReference       string
	SearchPlaceholder string
	SearchLabel       string
	SearchResults     string
	// SearchEmpty and SearchFailed are read by static/search.js, which puts the
	// query for {query} and the error for {error}.
	SearchEmpty  string
	SearchFailed string
	LanguageNav  string
	Footer       string

	Contents      string
	Documentation string
	OnThisPage    string
	Pages         string
	Previous      string
	Next          string

	HomeTitle       string
	HomeDescription string
	HomeHeadline    string
	HomeTagline     string
	ReadGuide       string
	ViewOnGitHub    string
	WhatYouGet      string
	Points          []Point
	TheGuide        string

	NotFoundTitle     string
	NotFoundHeadline  string
	NotFoundBefore    string
	NotFoundGuideLink string
	NotFoundAfter     string
}

// Point is one of the home page's four claims.
type Point struct {
	Title string
	Text  string
}

var texts = map[string]Text{
	"en": {
		Name:              "English",
		Docs:              "Docs",
		GoReference:       "Go reference",
		SearchPlaceholder: "Search the docs",
		SearchLabel:       "Search the documentation",
		SearchResults:     "Results",
		SearchEmpty:       "Nothing matches “{query}”.",
		SearchFailed:      "Search is unavailable ({error}).",
		LanguageNav:       "Language",
		Footer:            "collage is MIT-licensed. These pages are a collage site, exported to static files.",

		Contents:      "Contents",
		Documentation: "Documentation",
		OnThisPage:    "On this page",
		Pages:         "Pages",
		Previous:      "Previous",
		Next:          "Next",

		HomeTitle:       "collage — server-rendered pages from cached fragments, in Go",
		HomeDescription: "A Go framework for server-rendered pages composed from cached fragments. No dependencies.",
		HomeHeadline:    "Pages composed from cached fragments.",
		HomeTagline:     "collage is a Go framework for server-rendered sites. A page is a layout around fragments; each fragment fetches its own data, and the page is cached and invalidated by what it was built from. No client framework, no JavaScript build step, no dependencies.",
		ReadGuide:       "Read the guide",
		ViewOnGitHub:    "View on GitHub",
		WhatYouGet:      "What you get",
		Points: []Point{
			{"Fragments, not components", "A template, a data handler and the slots it exposes. Siblings fetch concurrently; the page is written in order."},
			{"Cached by what it depends on", "Pages and data carry tags. Change an author and one call drops the author and every page that shows them."},
			{"Forms without JavaScript", "Actions answer POST with forgery protection built in, and a cached page can still carry a form."},
			{"Served or exported", "One program is a server and a static site generator. This site is the second."},
		},
		TheGuide: "The guide",

		NotFoundTitle:     "Not found — collage",
		NotFoundHeadline:  "There is nothing here.",
		NotFoundBefore:    "The page may have moved. The ",
		NotFoundGuideLink: "guide",
		NotFoundAfter:     " starts at the beginning.",
	},
	"tr": {
		Name:              "Türkçe",
		Docs:              "Dokümantasyon",
		GoReference:       "Go referansı",
		SearchPlaceholder: "Dokümantasyonda ara",
		SearchLabel:       "Dokümantasyonda ara",
		SearchResults:     "Sonuçlar",
		SearchEmpty:       "“{query}” ile eşleşen bir sonuç yok.",
		SearchFailed:      "Arama şu anda kullanılamıyor ({error}).",
		LanguageNav:       "Dil",
		Footer:            "collage MIT lisanslıdır. Bu sayfalar, static dosyalara export edilmiş bir collage sitesidir.",

		Contents:      "İçindekiler",
		Documentation: "Dokümantasyon",
		OnThisPage:    "Bu sayfada",
		Pages:         "Sayfalar",
		Previous:      "Önceki",
		Next:          "Sonraki",

		HomeTitle:       "collage — Go ile cache'lenmiş fragment'lerden sunucuda render edilen page'ler",
		HomeDescription: "Cache'lenmiş fragment'lerden oluşan, sunucuda render edilen page'ler için bir Go framework'ü. Hiçbir bağımlılığı yoktur.",
		HomeHeadline:    "Cache'lenmiş fragment'lerden oluşan page'ler.",
		HomeTagline:     "collage, sunucuda render edilen siteler için bir Go framework'üdür. Bir page, fragment'leri saran bir layout'tur. Her fragment kendi verisini çeker. Page ise neyden oluştuysa ona göre cache'lenir ve invalidate edilir. Client framework'ü, JavaScript build adımı ya da bağımlılık gerekmez.",
		ReadGuide:       "Rehberi oku",
		ViewOnGitHub:    "GitHub'da görüntüle",
		WhatYouGet:      "Neler sunuyor",
		Points: []Point{
			{"Component değil, fragment", "Bir fragment; bir template, bir data handler ve dışarı açtığı slot'lardan oluşur. Kardeş fragment'ler verilerini eşzamanlı çeker, page yine de sırasıyla yazılır."},
			{"Neye bağlıysa ona göre cache'lenir", "Page'ler ve veriler tag taşır. Bir yazarı değiştirdiğinizde tek bir çağrı hem yazarı hem de onu gösteren her page'i cache'ten düşürür."},
			{"JavaScript'siz form'lar", "Action'lar POST request'lerini yerleşik forgery korumasıyla karşılar. Cache'lenmiş bir page de form içerebilir."},
			{"Serve edin ya da export edin", "Tek bir program hem sunucu hem de static site generator'dır. Bu site ikinci yolla üretildi."},
		},
		TheGuide: "Rehber",

		NotFoundTitle:     "Bulunamadı — collage",
		NotFoundHeadline:  "Burada bir şey yok.",
		NotFoundBefore:    "Aradığınız sayfa taşınmış olabilir. ",
		NotFoundGuideLink: "Rehberi",
		NotFoundAfter:     " en baştan okuyabilirsiniz.",
	},
}

// For returns the strings for locale, or English for a locale the site has no
// strings for.
func For(locale string) Text {
	if text, ok := texts[locale]; ok {
		return text
	}
	return texts["en"]
}
