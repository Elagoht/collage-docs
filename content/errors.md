---
description: Every exported error value in collage, grouped by where it comes from, with what it means and what to do about it.
reference: ErrCSRFCrossOrigin, ErrCSRFHeaderRequired, ErrStreamingBody, ErrStreamingBodyMethod, ErrNegativeBodyTimeout, ErrCachedFetchPanicked, PanicError, ErrUnknownSlot, ErrConflictingData, ErrTemplateType, TemplateTypeError, ErrNoDocumentHandler, ErrRouteParams, ErrUnknownFragmentPath, ErrAmbiguousFragmentPath, ErrInvalidGuardDecision, ErrGuardRefused, ErrGuarded, ErrConfigureClosed, ErrInvalidCommandName, ErrRedirectLoop, ErrRedirectShadowsPage, ErrLocaleUnreachable, ErrInvalidPattern, ErrDangerousOutDir, ErrOutputPathCollision, ErrUnknownAsset, ErrInvalidPrefix, ErrTooManyInlineSources
---

# Errors

collage reports failures with sentinel error values, exported from `pkg/collage`,
and wraps them with the details — the page, the path, the field. Match them with
`errors.Is`, never by comparing messages:

```go
if err := app.RegisterPage(page); errors.Is(err, collage.ErrDuplicateRoute) {
	// two pages claim one path
}
```

Most of them are reported **at startup**: `New` validates the configuration and
parses every template, registration validates every page, and starting the
application checks what only the whole set can reveal.

A mistake in how the site is put together is a program that refuses to start, not
a page that fails for the first reader who finds it. Under `collage dev` that
refusal is shown in the browser, as a 503 page with what the program printed — see
[CLI](/docs/cli#collage-dev).

<figure class="diagram-figure" data-search-ignore>
<svg class="diagram" viewBox="0 0 320 384" role="img" aria-labelledby="errs-title">
<title id="errs-title">When an error reaches you. collage.New returns configuration errors such as ErrTemplateRootMissing. A builder records its errors, such as ErrMissingContent, on the value it built, and registration refuses that value. Registration returns what validation finds, such as ErrUnknownSlot. Starting, with Start or ListenAndServe, returns what only the whole set shows, such as ErrMountShadowsRoute. After that, a failure during a request becomes a 404, a 500 or a fallback, reaches error hooks with its stage, and the production error page hides its cause. A static build with collage.Build returns its error or records it in the report: Errors, Skipped and Findings. errors.Is is the same test for every sentinel, however it is wrapped.</title>
<defs><marker id="errs-arrow" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto"><path class="d-head" d="M0 0 8 4 0 8z" /></marker></defs>
<rect class="d-frame" x="4" y="8" width="130" height="40" rx="6" /><text x="69" y="24" text-anchor="middle">collage.New</text><text class="d-small" x="69" y="38" text-anchor="middle">config, templates</text>
<line class="d-flow" x1="134" y1="28" x2="148" y2="28" marker-end="url(#errs-arrow)" />
<rect class="d-chip" x="152" y="8" width="164" height="40" rx="6" /><text class="d-small" x="234" y="24" text-anchor="middle">ErrTemplateRootMissing</text><text class="d-small" x="234" y="38" text-anchor="middle">New returns it</text>
<line class="d-flow" x1="69" y1="48" x2="69" y2="54" marker-end="url(#errs-arrow)" />
<rect class="d-frame" x="4" y="58" width="130" height="40" rx="6" /><text x="69" y="74" text-anchor="middle">builders</text><text class="d-small" x="69" y="88" text-anchor="middle">NewPage … Build()</text>
<line class="d-flow" x1="134" y1="78" x2="148" y2="78" marker-end="url(#errs-arrow)" />
<rect class="d-chip" x="152" y="58" width="164" height="40" rx="6" /><text class="d-small" x="234" y="74" text-anchor="middle">ErrMissingContent</text><text class="d-small" x="234" y="88" text-anchor="middle">refused at Register</text>
<line class="d-flow" x1="69" y1="98" x2="69" y2="104" marker-end="url(#errs-arrow)" />
<rect class="d-frame" x="4" y="108" width="130" height="40" rx="6" /><text x="69" y="124" text-anchor="middle">registration</text><text class="d-small" x="69" y="138" text-anchor="middle">Register, RegisterPage</text>
<line class="d-flow" x1="134" y1="128" x2="148" y2="128" marker-end="url(#errs-arrow)" />
<rect class="d-chip" x="152" y="108" width="164" height="40" rx="6" /><text class="d-small" x="234" y="124" text-anchor="middle">ErrUnknownSlot</text><text class="d-small" x="234" y="138" text-anchor="middle">Register returns it</text>
<line class="d-flow" x1="69" y1="148" x2="69" y2="154" marker-end="url(#errs-arrow)" />
<rect class="d-frame" x="4" y="158" width="130" height="40" rx="6" /><text x="69" y="174" text-anchor="middle">start</text><text class="d-small" x="69" y="188" text-anchor="middle">Start, ListenAndServe</text>
<line class="d-flow" x1="134" y1="178" x2="148" y2="178" marker-end="url(#errs-arrow)" />
<rect class="d-chip" x="152" y="158" width="164" height="40" rx="6" /><text class="d-small" x="234" y="174" text-anchor="middle">ErrMountShadowsRoute</text><text class="d-small" x="234" y="188" text-anchor="middle">Start returns it</text>
<path class="d-flow" d="M69 198C69 206 79 206 79 214" marker-end="url(#errs-arrow)" />
<path class="d-flow" d="M100 198C100 210 241 202 241 214" marker-end="url(#errs-arrow)" />
<rect class="d-c" x="4" y="218" width="150" height="26" rx="6" /><text class="d-on-c" x="79" y="235" text-anchor="middle">a request</text>
<rect class="d-c" x="166" y="218" width="150" height="26" rx="6" /><text class="d-on-c" x="241" y="235" text-anchor="middle">collage.Build</text>
<line class="d-flow" x1="79" y1="244" x2="79" y2="256" marker-end="url(#errs-arrow)" />
<line class="d-flow" x1="241" y1="244" x2="241" y2="256" marker-end="url(#errs-arrow)" />
<rect class="d-chip" x="4" y="260" width="150" height="54" rx="6" /><text class="d-small" x="79" y="276" text-anchor="middle">404, 500 or fallback</text><text class="d-small" x="79" y="290" text-anchor="middle">OnError gets the stage</text><text class="d-small" x="79" y="304" text-anchor="middle">production hides cause</text>
<rect class="d-chip" x="166" y="260" width="150" height="54" rx="6" /><text class="d-small" x="241" y="276" text-anchor="middle">Build returns it, or</text><text class="d-small" x="241" y="290" text-anchor="middle">the report: Errors,</text><text class="d-small" x="241" y="304" text-anchor="middle">Skipped, Findings</text>
<rect class="d-cmd" x="4" y="332" width="312" height="44" rx="6" /><text class="d-cmd-text" x="160" y="350" text-anchor="middle">errors.Is(err, collage.ErrUnknownSlot)</text><text class="d-cmd-text d-small" x="160" y="366" text-anchor="middle">the same test for every sentinel, however wrapped</text>
</svg>
<figcaption>When each kind of error reaches you. <code>New</code> refuses a bad configuration. What a builder records stays on the value it built, and registration refuses that value along with whatever validation finds. Starting checks what only the whole set shows: <code>Start</code> returns that failure, and <code>Handler</code> logs it and answers 503. After that, a failure during a request becomes a status, a fallback or an error page and reaches error hooks, and the production page says nothing about its cause; a static build returns its error or records it in its report. Whichever it is, match it with <code>errors.Is</code>.</figcaption>
</figure>

The tables below are grouped by where each error comes from. The message is the
text the sentinel carries before any wrapping adds detail. Three mistakes
registration catches, and how to read what collage prints for each, are worked
through in [Example: three mistakes registration catches](#example-three-mistakes-registration-catches).

A few failures carry no sentinel, because there is nothing an application could
do by matching one: `New(nil)`, a `nil` plugin or page, a plugin or command with
an empty name, a `{{dict}}` call with an odd number of arguments or a key that is
not a string, a disk cache with no version, a template `html/template` cannot
escape (see [Fragments and pages](#fragments-and-pages)), a `{{hoist}}` area name
that is not letters, digits, `_` and `-`, and a slot resolver's inline template
past the limit (see [Rendering](#rendering)). They are still refused, with a
message that says what is wrong.

## Configuration

Returned by `collage.New`, which checks the configuration before it builds the
application. See [Configuration](/docs/configuration#validation).

| Error | Message | Means | What to do |
| --- | --- | --- | --- |
| `ErrInvalidPort` | `collage: invalid port` | `Server.Port` is outside `1`–`65535`. | Leave it zero for `6060`, or set a valid port. |
| `ErrTemplateRootMissing` | `collage: template root missing` | `Template.Root` does not exist or is not a directory. | The most common startup failure: check the working directory, or embed the templates. |
| `ErrTemplateEscapesRoot` | `collage: template escapes root` | A template under `Template.Root` resolves outside it — a symlink leading out of the directory. | Copy the file in rather than linking to it. |
| `ErrInvalidCacheType` | `collage: invalid cache type` | `Cache.Enabled`, no `Store`, and `Type` is neither `collage.CacheMemory` (`"memory"`) nor `collage.CacheDisk` (`"disk"`). | Fix the spelling, or supply a `Store`. |
| `ErrEmptyCacheDir` | `collage: disk cache needs a directory` | `Cache.Type` is `"disk"` and `Cache.Dir` is empty. | Set `Dir`. |
| `ErrLocaleDefaultNotSupported` | `collage: default locale not in supported locales` | `Locale.Supported` does not include `Locale.Default`. | Add the default to `Supported`. |
| `ErrNegativeDuration` | `collage: negative duration` | One of the eight duration fields is negative; the message names which. | Use zero for the default. |
| `ErrInvalidOrigin` | `collage: invalid origin: want scheme://host[:port] with no path, query or fragment` | `BaseURL` is not a bare origin — it has a path, a query or a fragment, or no host (`https://:8080`), or an empty query or fragment marker (`https://example.com/?`). `collage.ParseOrigin` returns the same error. | Write `https://example.com`, with nothing after the host and port. |

## Plugins and commands

| Error | Message | Means | What to do |
| --- | --- | --- | --- |
| `ErrDuplicatePlugin` | `collage: duplicate plugin` | Two plugins share a name. | Register each plugin once. |
| `ErrConfigurerRegisteredLate` | `collage: plugin needs Configure and must be supplied in Config.Plugins` | `RegisterPlugin` was given a plugin with a `Configure` phase, which has already passed. | Move it to `Config.Plugins`. |
| `ErrDuplicateTemplateFunc` | `collage: duplicate plugin template function` | `AddTemplateFunc` was called for a name already added — by another plugin, or by the same one a second time — or for a built-in's name, such as `json`, or a name of the template language (`eq`, `len`, `html`, `urlquery`…); from `AddRenderFunc` as from `AddTemplateFunc`. | `AddTemplateFunc` returns it to the plugin's `Configure`; `New` fails only if `Configure` returns it. The plugins conflict: `Template.Funcs` cannot prevent it, so drop or rename one of them. A built-in's name is the plugin's to change. |
| `ErrUnknownPluginConfig` | `collage: plugin configuration names no registered plugin` | A `PluginConfig` key matches no registered plugin. Checked when the application starts. | Almost always a typo in the key. |
| `ErrConfigureClosed` | `collage: plugin configuration is closed: call it from Configure` | `AddTemplateFunc` or `AddRenderFunc` was called through a `ConfigHost` after the plugin's `Configure` returned; nothing was added. | Call it from `Configure`. |
| `ErrDuplicateCommand` | `collage: duplicate command name` | Two commands share a name — or the name is a `collage-` one, which `DispatchCommands` answers itself, or a plugin's command took one of the `collage` binary's (`build`, `serve`…). | Pick another name. |
| `ErrInvalidCommandName` | `collage: invalid command name` | A command's name is not one shell word: letters and digits of any script, `-` `_` `.` `:`, a letter or digit first. `collage add command` refuses such a `--name` too. | — |
| `ErrNilApp` | `collage: nil app` | `DispatchCommands` or `collage.Build` was given a nil `*App`. | — |
| `ErrUnknownCommand` | `collage: unknown command` | `DispatchCommands` got no arguments, or a name no plugin registered. | The scaffolded `main.go` exits `2` on it, as a usage error; a program that would rather serve can fall through on it instead. |
| `ErrStreamingBody` | `collage: the action's body is streamed; it is not parsed as a form` | `BeforeActionEvent.Form()` was called for an action built with `WithStreamingBody()`, and read nothing. | Treat it as "this action has no form to check", and let the request through. |

See [Writing a plugin](/docs/writing-plugins).

## Fragments and pages

Two places report these. A few are recorded by the builders as the chain runs —
`ErrDuplicateSlot` and `ErrSlotResolved` from `WithSlot`, `WithSlotResolver` and
`WithSlotFragment` (which also records `ErrNilFragment` and
`ErrSlotOccupied`), `ErrInvalidTimeout` from `WithTimeout`, `ErrDuplicateLocalePath`
and `ErrInvalidLocale` from `WithPath`, `ErrMissingLayout`,
`ErrConflictingLayout`, `ErrNilFragment` and `ErrFragmentCycle` from `WithLayouts`,
`ErrMissingContent` from a page's `Build`, and `ErrNoDocumentHandler` from a
document's — and you can read them with `BuildErr()`.

What a builder recorded stays on the value it built, and `RegisterPage` and `RegisterDocument` refuse a value
carrying any — a page's own, or those of any fragment in its tree — wrapped as
`collage: page %q was built with errors: %w` (`collage: document %q was built with
errors: %w` for a document), whether or not `BuildErr()` was called.

The rest are found by validation when the page is registered: `RegisterPage` checks
every fragment in the tree, its names, templates, slots — each binding against the
slots its template calls — data, timeouts, TTLs, paths and error pages, and returns
the first failure.

Every fragment's root template is also escaped once, as its first render would
escape it. A template `html/template` cannot escape — a `<script>` string that
never closes, `{{if}}` branches that leave an attribute in different states — is
refused with an error naming the page, the fragment and the template, rather than
failing on the first request. It carries no sentinel of its own; see
[Templates](/docs/templates#loaded-once-checked-early).

A fragment opened with `WithFragmentPath` is part of the page for this: its
template, its builder's mistakes and its validation are checked at registration
like the rest. Either way a malformed page is refused before it serves anything.

| Error | Message | Means |
| --- | --- | --- |
| `ErrEmptyName` | `collage: empty name` | A fragment or page has no name. |
| `ErrEmptyTemplatePath` | `collage: empty template path` | A fragment names no template — or `NewInlineFragment` was given an empty one. |
| `ErrConflictingTemplate` | `collage: fragment has both a template path and an inline template` | A fragment built by hand sets both a template file and an inline template. |
| `ErrSourceConflict` | `collage: inline template conflicts with another template` | An [inline template](/docs/fragments-and-slots#inline-templates) `{{define}}`s or `{{block}}`s a template of its own, which would replace a file template of that name for every page. |
| `ErrNilFragment` | `collage: nil fragment` | A `nil` fragment was used where one is required — bound to a slot, returned by a slot resolver, or listed in `WithLayouts`. |
| `ErrDuplicateSlot` | `collage: slot already declared` | `WithSlot` was called twice with one name. |
| `ErrUnknownSlot` | `collage: unknown slot` | A fragment is bound into a slot its template never calls — a typo on either side of the binding. The message names the slot and the slots the template does call. A slot the template calls and nothing fills is not an error: it renders empty. |
| `ErrInvalidSlotDefinition` | `collage: invalid slot definition` | A slot has an empty name, a map key that does not match its own name, or `WithSlot` was given a rule other than `SlotRequired` and `SlotSingle`. |
| `ErrSlotOccupied` | `collage: slot already occupied` | A second fragment was bound to a slot that holds one — or a resolver returned several for it, or a layout in a page's chain arrived with its `content` slot already filled. |
| `ErrSlotResolved` | `collage: slot is filled by a resolver` | One slot was given both a resolver and bound fragments. |
| `ErrRequiredSlotUnfilled` | `collage: required slot has no fill` | A slot declared required has nothing bound to it. |
| `ErrFragmentCycle` | `collage: fragment cycle detected` | A fragment is reachable from itself — or `WithLayouts` names one layout twice. |
| `ErrMissingContent` | `collage: missing content` | A page has no content fragment. |
| `ErrMissingLayout` | `collage: missing layout` | `WithLayouts` was called with no layouts. A page without a layout does not call it. |
| `ErrConflictingLayout` | `collage: layout chain already declared` | `WithLayouts` was called twice on one builder. |
| `ErrConflictingData` | `collage: conflicting data` | A fragment's data is set twice — `WithData` called twice with data — or a document sets both `WithBody` and `WithHandler`. The message names the fragment: `fragment "x" has its data set twice`. |
| `ErrTemplateType` | `collage: template does not fit its data` | A template reads something its fragment's data type cannot give it: `{{.Titel}}` on a type with no such field, say. Each finding is a `*TemplateTypeError` naming the page, the fragment, the file, the line and the column; see [How templates are checked](/docs/data-handlers#how-templates-are-checked). |
| `ErrInvalidTimeout` | `collage: invalid timeout` | A fragment's or a document's timeout is negative. |
| `ErrInvalidTTL` | `collage: invalid cache ttl` | A page's or document's TTL is negative, or `Incremental` was given a zero TTL. |
| `ErrInvalidPath` | `collage: invalid path` | A path pattern does not start with `/` — or a redirect's destination would leave the site: `//host`, `/\host`. |
| `ErrInvalidRedirectStatus` | `collage: invalid redirect status code` | A redirect status other than `0`, `301`, `302`, `307` or `308` — or, for a `RedirectSource` plugin's rule in a static build, other than those and `410`. |
| `ErrInvalidRedirect` | `collage: invalid redirect` | A redirect's `From` or `To` holds a control character — `\r`, `\n`, any other, or the Unicode line and paragraph separators — from `WithRedirect` or `WithPermanentRedirect`, or a hand-built `Redirect` checked by the page's `Validate`. In a static build, also a malformed or control-character rule from a `RedirectSource` plugin, or one whose `From` or `To` holds a space; each such rule is reported. |
| `ErrSelfErrorPage` | `collage: page cannot reference itself as an error page` | A page is its own not-found or error page. |
| `ErrDuplicateLocalePath` | `collage: locale already has a path` | `WithPath` named a locale the page, document or action already has a path in — or a document's `AtRoot` was called twice. The first path is kept. |
| `ErrInvalidLocale` | `collage: invalid locale` | `WithPath` was given an empty locale. A document that belongs outside every locale says `AtRoot(p)`. |

See [Pages and layouts](/docs/pages-and-layouts) and
[Fragments and slots](/docs/fragments-and-slots).

## Registration and routing

Returned by `RegisterPage`, `RegisterNotFoundPage`, `RegisterErrorPage`,
`RegisterDocument` and `Register`, or when the application starts.

| Error | Message | Means | What to do |
| --- | --- | --- | --- |
| `ErrAppStarted` | `collage: application already started` | A registration method was called after the application started — including `RegisterPlugin` after any start that failed, and `RegisterCommand`. | Register everything before `Handler`, `ListenAndServe`, `Start`, `RenderPath` or `DispatchCommands`. |
| `ErrNilRegistrable` | `collage: nil passed to Register` | `Register` was given a `nil` where a page, a document or an action was expected; the message names its position. | — |
| `ErrDuplicatePage` | `collage: duplicate page name` | Two pages share a name. | Names are how links find pages; make them unique. |
| `ErrTemplateNotFound` | `collage: template not found` | A page's fragment names a template that was not loaded. | Check the path relative to `Template.Root`, extension included. |
| `ErrUnregisteredErrorPage` | `collage: error page not registered` | A page names a not-found or error page that was never registered — including a second `Build()` of the one that was: each `Build()` is a page of its own. Checked at start. | Register it with `RegisterNotFoundPage` or `RegisterErrorPage`, and name that same `*Page`; an unregistered one would render empty when needed. |
| `ErrInvalidPattern` | `collage: invalid pattern` | A path or redirect source is malformed: no leading `/`, an empty segment, an empty placeholder name, a catch-all that is not last or has text around it, two placeholders in one segment, such as `/{name}.{ext}`, or one placeholder name used twice, such as `/a/{id}/b/{id}`, where a link could not say which `id` it fills. Text around a single placeholder, such as `/feeds/{category}.xml`, is a route. | Fix the pattern. Register `/{name}.md` and `/{name}.json` as two routes, and give each placeholder a name of its own. |
| `ErrLocaleUnreachable` | `collage: no URL reaches that locale` | A page, action or document has a path in a locale that is neither `Locale.Default` nor in `Locale.Supported`. Or a page, document, action or redirect source begins with literal text naming `Locale.Default` or a supported locale, in any case — `/tr/guide.txt` — which a request takes for the locale prefix and strips. Either way no request could reach it. | Set `Locale.Default` to the locale you wrote, or add it to `Supported`. A site in one language that is not English needs `Default` set. Leave the prefix off a path: the locale is added to the URL for you. |
| `ErrDuplicateRoute` | `collage: duplicate route` | A path or redirect source is already registered in that locale — or an action answers `GET` or `HEAD` on a page's or a document's path, in either registration order. | A fragment path spelled like its page's path is the usual cause; give it a path of its own. |
| `ErrAmbiguousParameterName` | `collage: ambiguous parameter name` | Two patterns use different parameter names at one position, such as `/blog/{slug}` and `/blog/{id}/edit`. | Use one name at that position. |
| `ErrOverlappingPattern` | `collage: overlapping pattern` | Two placeholders with text around them at one position cross: a segment matches both and neither is more specific, such as `/a{x}` and `/{x}b`, which `/aXb` matches. The message names such a segment. | Make one's text begin or end the other's, so it is the more specific, or give them text no segment can carry both of. |
| `ErrRedirectShadowsPage` | `collage: redirect shadows a registered page` | A redirect's source matches a page's or a document's path, in either registration order — not only one spelled the same: `/{slug}` takes a document at `/about.txt`, and `/old/{rest...}` takes `/old/keep.txt`. The message names both. | One of the two would be unreachable, since a redirect wins over a route. Narrow the redirect's source, or remove one. |
| `ErrRedirectLoop` | `collage: redirect loops back to its own source` | A redirect's destination is a path its own source matches: `/loop` to `/loop`, `/docs` to `/docs/` (a source matches with or without its trailing slash), `/blog/{slug}` to `/blog/{slug}`. The query and fragment are not compared. | A client following it would never arrive anywhere: point it somewhere else, or drop it. |
| `ErrUnsubstitutedPlaceholder` | `collage: redirect placeholder not captured by from pattern` | A redirect's destination uses a `{name}` its source does not capture. | Capture it in the source, or remove it. |

## Documents

| Error | Message | Means |
| --- | --- | --- |
| `ErrNilDocument` | `collage: nil document` | A `nil` document was registered. |
| `ErrEmptyContentType` | `collage: empty content type` | A document declares no content type. It is required and never guessed. |
| `ErrNoDocumentHandler` | `collage: document has no handler or body` | A document has neither a handler nor a fixed body (`WithBody`). Unlike a page, it has no template to fall back on. |
| `ErrConflictingData` | see [Fragments and pages](#fragments-and-pages) | A document has both `WithHandler` and `WithBody`. |
| `ErrDuplicateDocument` | `collage: duplicate document name` | Two documents share a name. |
| `ErrDocumentNotFound` | `collage: no document at path` | `RenderDocumentPath` found no document at the path — including when a page or a redirect is there. |
| `ErrEmptyDocumentBody` | `collage: document handler produced an empty body` | A handler succeeded with an empty body: a 500 when served, and not written by a build. A handler that really means "empty" can return a single newline. |

See [Documents](/docs/documents).

## Mounts and handlers

| Error | Message | Means |
| --- | --- | --- |
| `ErrUnknownAsset` | `collage: unknown asset` | `rc.Asset`, `rc.HoistStylesheet`, `{{asset}}` or `{{stylesheet}}` was given a path no mount serves — including a name the mount would not serve: a dotfile such as `.env`, or one that is not a valid path within the mount. No link that would answer 404 is made. In a template, the render fails. |
| `ErrInvalidPrefix` | `collage: invalid prefix` | A mount prefix does not begin and end with `/`, is `/` alone, or begins with `//` — or `App.Handle` was given a prefix that does not begin with `/`, is `/` alone, or holds `//` or a `.` or `..` segment, which the clean-path redirect removes before any handler is asked; the message says which. A segment that only begins with a dot is fine, so `App.Handle` takes `/.well-known/`. A mount at `/` would swallow every route. |
| `ErrNilFS` | `collage: nil mount file system` | A mount was given no filesystem. |
| `ErrMountConflict` | `collage: mount prefixes overlap` | Two mounts — or a mount and an `App.Handle` prefix — claim overlapping prefixes. |
| `ErrMountShadowsRoute` | `collage: mount shadows a route` | A mount prefix, or an `App.Handle` prefix, would swallow a route's path — a page's, a document's, a redirect's or an action's. Checked at start, whatever the registration order. |
| `ErrNilHandler` | `collage: nil handler` | `App.Handle` or `App.Use` was given nothing to run. |

See [Static assets](/docs/assets) and
[Middleware and your own API](/docs/middleware-and-apis).

## Rendering

| Error | Message | Means | What to do |
| --- | --- | --- | --- |
| `ErrNotFound` | `collage: not found` | **You return this one.** A data handler that wraps it says the content does not exist, rather than failed to load. | `fmt.Errorf("post %q: %w", slug, collage.ErrNotFound)` makes a required fragment's failure a 404 with the not-found page, not a 500. |
| `ErrRequiredSlotEmpty` | `collage: required slot is empty` | At render time, a required slot holds nothing — typically a resolver returned no fragments. | Subject to the fragment's failure policy. |
| `ErrMaxDepthExceeded` | `collage: max fragment depth exceeded` | The fragment tree nests deeper than the engine allows. | Almost always a fragment bound into its own slot, directly or not. |
| `ErrPageNotFound` | `collage: no page at path` | `App.RenderPath` found no page at the path. | — |
| `ErrCachedFetchPanicked` | `collage: data fetch panicked` | A `collage.Cached` call waited on another render's fetch, and that fetch panicked. The panic itself goes to the render whose fetch it was. | The key is free again: the next render fetches. Fix what panicked. |
| `ErrCSRFDisabled` | `collage: csrfToken used but request-forgery protection is disabled` | A template calls `{{csrfToken}}` in an application with `Security.DisableCSRF` set. | Remove the call, or turn protection back on. |

A [`collage.Key`](/docs/data-handlers#collagekey) is its name and its type, so one
name asked for as two types is two keys holding two values, not an error.

Three render failures to know about:

- **A slot resolver's inline templates past the limit.** The engine takes in at
  most 256 distinct inline templates that slot resolvers' fragments bring while the
  program runs; the next fails its fragment with `collage.ErrTooManyInlineSources`,
  which `errors.Is` recognises. An inline template is a constant, never built from data — see
  [Inline templates](/docs/fragments-and-slots#inline-templates).
- **A `{{hoist}}` area name** that is not letters, digits, `_` and `-` fails the
  render: `collage: hoist area "…": a name is letters, digits, '_' and '-'`.
- **A `collage.Once` fetch that panicked.** Every fragment waiting on the key is
  released at once with `collage: Once "…": the fetch panicked: …`, and the panic
  goes to the fragment whose fetch it was.

Concurrent requests that miss the page cache for one cached page share one
render. If that render panics, the requests waiting on it are answered with a 500
too, and the next request renders again.

A panic in a data handler, a slot resolver or a template function does not take
the process down: it becomes a `collage.PanicError`, and the fragment fails like
any other. Reach it with `errors.As` to recover the panic value and its stack:

```go
var panicked *collage.PanicError
if errors.As(err, &panicked) {
	log.Printf("panic: %v", panicked.Value)
}
```

See [Data handlers](/docs/data-handlers) for how a failure becomes a 404, a 500, a
fallback or nothing.

In development, the built-in 500 page leads with the cause: for a template, the
template, line and column of the call that failed and what it returned, above the
chain of templates the error passed through on its way up. The production page is
unchanged, and says nothing about the cause.

## Links and URLs

Returned by `App.URL`, `App.ActionURL` and `App.FragmentURL`, and failed renders from `{{pageURL}}`, `{{actionURL}}`,
`{{pageURLIn}}`, `{{localeURL}}`, `{{fragmentURL}}` and `{{fragmentURLIn}}`. See [Links and locales](/docs/links-and-locales).

| Error | Message | Means |
| --- | --- | --- |
| `ErrUnknownRoute` | `collage: no page or document by that name` | No page or document is registered under the name — or both are, and the link is ambiguous. |
| `ErrNoPathInLocale` | `collage: no path in that locale` | The route has no path in the locale asked for. `{{pageURL}}` falls back to the default locale instead, and `{{localeURL}}` renders the empty string. |
| `ErrRouteParams` | `collage: route parameters do not match the pattern` | A parameter is missing or empty, names no placeholder, is `.` or `..`, is neither a string, an integer nor a `fmt.Stringer`, or the template passed an odd number of arguments. |
| `ErrLocaleUnreachable` | `collage: no URL reaches that locale` | The locale is neither `Locale.Default` nor in `Locale.Supported`. Registering a path in one fails with it too; see [Registration and routing](#registration-and-routing). |
| `ErrUnknownFragmentPath` | `collage: the page opened no fragment path by that name` | The page opened no fragment by that name with `WithFragmentPath`. `RenderFragment` asked by name returns it too for parameters no request could carry, a locale the fragment has no path in, or a path that is not in its clean spelling. |
| `ErrAmbiguousFragmentPath` | `collage: the fragment is opened at more than one path` | The page opened the fragment at two paths in that locale, so the link cannot pick one. |

## Actions

Returned by `RegisterAction`. `RegisterPage` checks an action attached to a page
too — `ErrNilAction`, `ErrNoActionMethods`, `ErrNoActionHandler` and
`ErrDuplicateAction` — and refuses a `nil` fragment path with
`ErrNilFragmentPath`.

There is one `ErrNoActionHandler`, the same value at registration and on a
request. An action that recorded errors while it was built — a repeated locale in
`WithPath`, say — is refused at registration too; `ActionBuilder.BuildErr()` reads
them first.

| Error | Message | Means | What to do |
| --- | --- | --- | --- |
| `ErrNilAction` | `collage: nil action` | A `nil` action was registered, or attached to a page. | — |
| `ErrEmptyActionName` | `collage: action has no name` | An action has no name. | Give it one; it names the action in logs and errors. |
| `ErrDuplicateAction` | `collage: duplicate action` | A second action was registered under a name already taken. | — |
| `ErrNoActionPaths` | `collage: action has no paths` | A standalone action has no `WithPath`. | An action on a page takes the page's paths; one on its own needs its own. |
| `ErrNoActionHandler` | `collage: action has no handler` | An action has no `WithHandler`. | — |
| `ErrInvalidActionMethod` | `collage: action method is reserved for the server` | An action declares `OPTIONS`, `TRACE` or `CONNECT`, which the server answers itself; the message names the action and the method. | Remove it from `WithMethods`. A CORS preflight belongs in a middleware. |
| `ErrStreamingBodyMethod` | `collage: a streaming body needs POST, PUT or PATCH` | An action built with `WithStreamingBody()` answers none of `POST`, `PUT` or `PATCH`; the message names the action and its methods. | Give it one of them, or drop `WithStreamingBody()`. |
| `ErrNegativeBodyTimeout` | `collage: an action's body timeout must not be negative` | `WithBodyTimeout` was given a negative duration. | Zero keeps the server's deadlines. |
| `ErrNoActionMethods` | `collage: action declares no methods` | An action answers no method. | `WithMethods(http.MethodPost)`, or `WithAction` on a page. |
| `ErrNilFragmentPath` | `collage: fragment path has no fragment` | `WithFragmentPath` was given a `nil` fragment. | — |
| `ErrUnregisteredPage` | `collage: action answered with a page that was never registered` | An action's `RenderPage` returned a page that was not registered — or a second `Build()` of one that was, which is another page. The request fails with a 500. | Register the page, and answer with that same value rather than building one in the handler. |
| `ErrAmbiguousActionResult` | `collage: action result sets more than one of Location, Fragment, Page, Body` | An action's `ActionResult` set two or more of `Location`, `Fragment`, `Page` and `Body`; the message names the fields. The request fails with a 500, and error hooks receive it under the stage `"render"`. | Answer with one of them. A result with only `Status` or `Header` is fine. |

See [Forms and actions](/docs/forms-and-actions).

## Vary and SkipCache

Returned by `collage.Vary` and `collage.SkipCache`, which must be called from
middleware. See [Caching](/docs/caching) and [Previews](/docs/previews).

| Error | Message | Means | What to do |
| --- | --- | --- | --- |
| `ErrVaryTooLate` | `collage: Vary or SkipCache called after routing; call it from middleware` | `Vary` or `SkipCache` was called after routing began — from a data handler, say. It is refused on every route. | Call it from middleware registered with `App.Use`. |
| `ErrVaryOutsideRequest` | `collage: Vary or SkipCache called on a request collage is not serving` | `Vary` or `SkipCache` was called on a request that did not come through collage's handler. | — |

## Request forgery

A submission that fails the forgery check is answered with a **403** before the
action's handler runs, and error hooks receive the reason under the stage
`"route"`, wrapped with the action's name:

| Error | Message | Means |
| --- | --- | --- |
| `ErrCSRFMissing` | `collage: no csrf token` | The submission carried no token, or no cookie. |
| `ErrCSRFMismatch` | `collage: csrf token does not match` | The token is not the one in the cookie. |
| `ErrCSRFInvalid` | `collage: csrf token is not valid` | The token was not signed with this application's key. |
| `ErrCSRFCrossOrigin` | `collage: cross-origin request` | The browser marked the submission as sent from another origin, by `Sec-Fetch-Site` or by an `Origin` that is not the `Host` — refused whatever token it carried. |
| `ErrCSRFHeaderRequired` | `collage: no csrf token from the header` | A submission to an action built with `WithStreamingBody()` carried no token in `X-CSRF-Token` (or `Security.CSRFHeaderName`), the only place such an action looks. It wraps `ErrCSRFMissing`, so `errors.Is` with either matches. |

The usual causes are a form with no `{{csrfToken}}`, and a key generated per
process — set `Security.CSRFKey` so tokens survive restarts and work across
instances. A plain HTML form posting to a streaming action is refused with
`ErrCSRFHeaderRequired`: it cannot set the header, so send it with `fetch()`. A body
over the size limit is a **413** instead, even when reading the token is what hit
the limit — and so is a urlencoded form over 10 MB, the most net/http reads,
however high `MaxBodyBytes` is. See
[Forms and actions](/docs/forms-and-actions#forgery-protection).

A page whose form needs a token that cannot be issued — the system's random source
failed — is answered **500**, rather than served with a form every submission of
would be refused.

`{{csrfToken}}` in an application with `Security.DisableCSRF` set is
`ErrCSRFDisabled`; see [Rendering](#rendering).

## Reported to error hooks

These reach a plugin's `OnError` in `ErrorEvent.Err`, so it can tell failures
apart without reading messages. See
[Writing a plugin](/docs/writing-plugins#errorhook).

| Error | Message | Stage | Means |
| --- | --- | --- | --- |
| `ErrNoRoute` | `collage: no route matched the request` | `not_found` | No route matched the request — a link or routing problem. Deliberately not `ErrNotFound`: both are a 404, and they have different causes. |
| `ErrNotFound` | see [Rendering](#rendering) | `render` | A required fragment's content does not exist — a content problem. |
| `ErrMethodNotAllowed` | `collage: method not allowed` | `route` | The path exists but answers no such method: a 405, with an `Allow` header naming what it does answer. On a document's URL the 405 is plain text. |
| `ErrEmptyRender` | `collage: page rendered no markup` | `render` | A page rendered successfully but produced no markup, served or answered by an action: a 500. The same sentinel a static build records. |
| `ErrAmbiguousActionResult` | see [Actions](#actions) | `render` | An action answered with a result that asks for two responses. |
| `ErrCSRFMissing`, `ErrCSRFMismatch`, `ErrCSRFInvalid`, `ErrCSRFCrossOrigin`, `ErrCSRFHeaderRequired` | see [above](#request-forgery) | `route` | A submission refused by the forgery check. |
| `ErrInvalidGuardDecision` | `collage: invalid guard decision` | `guard` | A page's [guard](/docs/pages-and-layouts#private-pages-guards) answered with a decision that cannot be written — a redirect with no location, a `200` — and the request failed with a 500. A guard that returns an error is reported at the same stage, with its own error. |
| `ErrGuardRefused` | `collage: the guard refused this reader` | — | Returned, not reported: `RenderFragment` asked for a fragment whose guard refuses the reader, or a middleware answered its request instead of passing it on — a 401, a 429, a redirect. A plugin streaming fragments sends that reader nothing. |
| `ErrEmptyErrorPage` | `collage: error page rendered empty` | `error_page` | A registered error page rendered successfully but produced no markup, so the built-in page was served instead. |
| `ErrPanic` | `collage: panic recovered while serving the request` | `panic` | Something panicked while serving — a `Cache`, `Metrics` or `Tracer` implementation, a router, a plugin hook — and was recovered into a 500. Panics in data handlers and templates are `PanicError` instead. |
| `ErrAssetFailed` | `collage: asset request failed` | `asset` | A mounted file request answered with a status of 400 or above: one sentinel for every such status. |
| `ErrHandlerFailed` | `collage: mounted handler failed` | `handler` | A handler mounted with `App.Handle` answered with a server error. |
| `ErrUnsafeRedirectTarget` | `collage: unsafe redirect target` | `route` | A redirect's destination, after substitution, is not a single-slash relative path — `//host`, `/\host`, or one with a control character. Answered with a 500, not a `Location` header. A destination written that way is refused at registration with `ErrInvalidPath`, and a captured value is escaped so it cannot make one. |

An action whose handler read past its body limit and then failed is
reported as a `413`, whatever error it returned: the hook receives the handler's
own error, and `errors.As` finds a `*http.MaxBytesError` on it. An action whose
body was cut off — the client went away, or the read deadline passed — is answered
`400` or `408` and reaches no error hook; it is logged at debug. See
[Streaming bodies](/docs/forms-and-actions#streaming-bodies).

`"error_page"` (`collage.StageErrorPage`; every stage has a `Stage*` constant) is the stage worth alerting on: the page that reports failures
failed, and the reader still saw a plausible page, so nothing else would tell you.

## Static builds

Returned by `collage.Build`, or recorded in the `BuildReport` it returns. See [Static export](/docs/static-export).

| Error | Message | Where | Means |
| --- | --- | --- | --- |
| `ErrNilApp` | see [Plugins and commands](#plugins-and-commands) | `Build` | The app is `nil`. |
| `ErrEmptyOutDir` | `collage: empty output directory` | `Build` | `BuildOptions.OutDir` is empty. |
| `ErrDangerousOutDir` | `collage: refusing to use a dangerous output directory` | `Build` | `OutDir` resolves, after symlinks are followed, to a filesystem root. With `Clean`, also when it holds the project itself — a repository root, the working directory or the module root, or any directory above them (`-out ..`) — or is a directory that is not empty and has no `.collage-export` marker: `Clean` empties only a directory a build created, or one it found empty. Checked before anything is deleted or written. To let `Clean` empty an output directory that has no marker, create it once: `touch dist/.collage-export`. |
| `ErrOutputPathCollision` | `collage: two builds target one output path` | `Build` | Two pages, two documents, or a page and a document would be written to one file — patterns differing only by a trailing slash or only in case (`/Go` and `/go` are one file on macOS and Windows, and collide on every platform), `WithStaticParams` listing the same values twice, or a document at a page's `index.html`. The build stops before `OutDir` is touched: nothing is cleaned or written, and `BuildFinishedHook` is not called. |
| `ErrPathEscapesOutDir` | `collage: resolved path escapes the output directory` | report error | An output path, or a symlink on the way to it, leads outside `OutDir`. Checked before that one file is written; only that path fails. |
| `ErrDynamicPathUnresolved` | `collage: a path pattern with a {param} needs WithStaticParams to be built` | skip | A page's or document's path has a `{param}` and no `WithStaticParams`. |
| `ErrRouteParams` | see [Links and URLs](#links-and-urls) | report error | A map `WithStaticParams` returned does not fill the pattern exactly — a name missing, or one the pattern does not have — or a value holds a backslash or a NUL (on Windows a backslash is a separator). Only that file fails; the rest are built. |
| `ErrNotStatic` | `collage: a Dynamic() route cannot be built statically` | skip | A page or document is dynamic — declared `Dynamic()`, or declaring no strategy and rendering a data handler — so there is nothing to export. |
| `ErrGuarded` | `collage: a guarded page cannot be built statically` | skip | A page has a [guard](/docs/pages-and-layouts#private-pages-guards) on its layout chain or content fragment. An export has no reader to ask it about, and a file on a static host is served to anyone. |
| `ErrDuplicateRedirect` | `collage: two redirects from one path` | report error | Two redirects the router takes for one — `/old` and `/old/`, `/blog/{slug}` and `/blog/{x}` — from pages, documents or a `RedirectSource` plugin. A static host would pick one by its own precedence. |
| `ErrRedirectShadowsFile` | `collage: a redirect shadows a written file` | report error | A redirect matches the path of a file the build wrote — `/about` or `/about/` beside `about/index.html`, `/docs/{rest...}` beside `docs/x/index.html`. Narrow the pattern, or drop the page. |
| `ErrDegradedRender` | `collage: refusing to write a degraded render` | report error | A page rendered with a failed fragment and `AllowDegraded` is off. No file is written. |
| `ErrEmptyRender` | `collage: page rendered no markup` | report error | A page rendered no markup at all. Refused even with `AllowDegraded`. One sentinel with serving's, above. |
| `ErrUnresolvedToken` | `collage: refusing to write a page whose forgery token was never resolved` | report error, skip | The not-found page carries a `{{csrfToken}}`: a report error. Any other page with one is skipped instead: it needs a server. |
| `ErrBuildPanic` | `collage: panic while building a page` | report error | Rendering or writing one page panicked, or a `WithStaticParams` function did, which fails that route's locale; the build recovered and went on with the rest. |
| `ErrEmptyDocumentBody` | see [Documents](#documents) | report error | A document produced an empty body. |
| `ErrBuildFindings` | `collage: checks reported errors` | report error | A plugin reported an error-level finding from `BuildFinishedHook`. The pages were written; what was found is in `BuildReport.Findings`. |

Report errors are in `BuildReport.Errors`, a slice of errors that match with
`errors.Is`. A skip is a `SkipRecord` in `BuildReport.Skipped`: its `Reason` is a
sentence for people, and its `Err` is the sentinel for code —
`ErrNotStatic`, `ErrGuarded`, `ErrDynamicPathUnresolved` or `ErrUnresolvedToken` —
so match it with `errors.Is(skip.Err, …)` rather than by
reading `Reason`.

## Example: three mistakes registration catches

A new project registers its pages in `routes.go`, and the scaffolded `main.go`
stops on the first error, printing it after the time and the project's name (the
time is left out below). Here the project is called `blog`. `Register` stops at the first item it refuses, so these arrive
one per run: fix it, run again, read the next.

Each line reads from the outside in:

- `blog:` is `main.go`'s;
- `register page "home":` is `Register` naming the item it refused;
- `collage: unknown slot` is the sentinel's message, the one in the tables above;
- the rest is the detail: the page, the fragment, the template.

### A slot name with a typo

The home page's content binds an aside into a slot, and spells it differently from
the template:

```go
func Home() *collage.Fragment {
	return collage.NewFragment("home", "pages/home.html").
		WithSlotFragment("sidebr", Aside()).
		Build()
}
```

```html
<h1>Home</h1>
{{slot "sidebar"}}
```

```text
blog: register page "home": collage: unknown slot: page "home" fragment "home" binds into slot "sidebr", but template "pages/home.html" never calls it; it calls [sidebar]
```

The end of the line is the answer: the template calls `sidebar`, and the Go side
binds `sidebr`. Correct whichever side is wrong. Unchecked, the aside would render
nowhere, on every request.

A template that works a slot's name out as it renders may call any slot, so its
bindings are not checked this way.

### A field the data does not have

The post's content, in `fragments/pages/post.go` (package `fragments`), hands its
template a `postView`, and the template misspells a field:

```go
type postView struct {
	Title string
}

func Post() *collage.Fragment {
	return collage.NewFragment("post", "pages/post.html").
		WithData(collage.Value(postView{Title: "Hello"})).
		Build()
}
```

```html
<h1>{{.Titel}}</h1>
```

```text
blog: register page "post": collage: page "post": fragment "post" (pages/post.html:1:6): {{.Titel}}: type fragments.postView has no field or method Titel (did you mean Title?)
```

This one does not carry its sentinel's message, `collage: template does not fit
its data`. Each finding is a `*collage.TemplateTypeError`, which prints the page,
the fragment, the template's file, line and column, the expression, what is wrong
with it and, when one is close, the name you probably meant. Write `{{.Title}}`.

`errors.Is(err, collage.ErrTemplateType)` matches it all the same. Every finding
in the page's templates is reported at once, one per line — when nothing else on
the page was refused first: a missing template or a wrong slot binding stops the
check before its findings are returned.
`errors.As` reaches the first finding's fields:

```go
var typeErr *collage.TemplateTypeError
if errors.As(err, &typeErr) {
	fmt.Printf("%s:%d:%d: %s\n", typeErr.Template, typeErr.Line, typeErr.Col, typeErr.Expr)
}
```

### A fragment path on its page's path

The search page opens its results at a URL of their own, so a script can refresh
them, and gives them the page's path:

```go
func Search() *collage.Page {
	return collage.NewPage("search").
		WithLayouts(layouts.Master()).
		WithContent(fragments.Search()).
		WithPath("en", "/search").
		WithFragmentPath("en", "/search", fragments.SearchResults()).
		Build()
}
```

```text
blog: register page "search": collage: duplicate route: action "search:search-results" and page "search" both answer GET "/search" for locale "en"
```

The action named there is not one you wrote. collage serves a fragment path with
an action that answers `GET`, named after the page and the fragment —
`fragments.SearchResults()` builds the one named `search-results`. Its path is the page's own, so two routes answer
`GET /search`. Give the fragment a path of its own, such as `/search/results`.

Two pages on one path say so in a message of the same shape:
`page "team" and page "about" both claim "/about" for locale "en"`.

### Catch them in a test

All three are returned by registration, so `newApp` returns them, and a test that
calls it fails on the same error before anything is deployed. `Start` adds what is
checked when the application starts:

```go
func TestAppStarts(t *testing.T) {
	app, err := newApp(collage.Env{CacheDir: t.TempDir()})
	if err != nil {
		t.Fatalf("newApp() = %v, want nil", err)
	}
	t.Cleanup(func() { _ = app.Shutdown(context.Background()) })
	if err := app.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}
}
```

See [Testing](/docs/testing#test-the-application-main-builds).
