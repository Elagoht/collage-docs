// Command checkcontent loads content/ the way the site does and reports the first
// problem, or how much of each translation there is.
//
//	go run ./tools/checkcontent              everything
//	go run ./tools/checkcontent tr caching   the original and only tr/caching.md
//
// Naming pages checks those alone, so a translator is not stopped by another
// translation in progress.
package main

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"strings"
	"testing/fstest"

	"github.com/Elagoht/collage-docs/site"
)

func main() {
	var fsys fs.FS = os.DirFS("content")
	if len(os.Args) > 2 {
		only, err := subset(fsys, os.Args[1], os.Args[2:])
		if err != nil {
			fail(err)
		}
		fsys = only
	}
	set, err := site.LoadSet(fsys, site.Original, site.Translations...)
	if err != nil {
		fail(err)
	}
	for _, locale := range set.Locales() {
		fmt.Printf("%s: %d pages\n", locale, len(set.Site(locale).Pages()))
	}
}

// subset is content/ with every page of locale's translation hidden but slugs.
func subset(fsys fs.FS, locale string, slugs []string) (fs.FS, error) {
	keep := map[string]bool{}
	for _, slug := range slugs {
		keep[path.Join(locale, strings.TrimSuffix(slug, ".md")+".md")] = true
	}
	out := fstest.MapFS{}
	err := fs.WalkDir(fsys, ".", func(name string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		if strings.HasPrefix(name, locale+"/") && strings.HasSuffix(name, ".md") && !keep[name] {
			return nil
		}
		data, err := fs.ReadFile(fsys, name)
		if err != nil {
			return err
		}
		out[name] = &fstest.MapFile{Data: data}
		return nil
	})
	return out, err
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
