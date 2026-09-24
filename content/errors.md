---
description: Every exported error value in collage, grouped by where it comes from, with what it means and what to do about it.
reference: PanicError
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
application checks what only the whole set can reveal. A mistake in how the site is
put together is a program that refuses to start, not a page that fails for the
first reader who finds it.

The tables below are grouped by where each error comes from. The message is the
text the sentinel carries before any wrapping adds detail.

Since v0.10.0 every error below is exported. Before it, the action, request-forgery,
method, asset, `{{dict}}` and template-escapes-root errors, and `ErrNotStatic`,
existed but could only be told apart by their messages. (`ErrTemplateRootMissing`
was exported already.)

## Configuration

Returned by `collage.New`, from `Config.Validate` or while building the
application. See [Configuration](/docs/configuration#validation).

| Error | Message | Means | What to do |
| --- | --- | --- | --- |
| `ErrNilConfig` | `collage: nil config` | `New(nil)`. | Pass a `*Config`; its zero value is fine. |
| `ErrInvalidPort` | `collage: invalid port` | `Server.Port` is outside `1`–`65535`. | Leave it zero for `3000`, or set a valid port. |
| `ErrEmptyTemplateRoot` | `collage: empty template root` | `Template.Root` is empty and `Template.FS` is nil. | Only reachable by calling `Validate` yourself; `New` defaults `Root` first. |
| `ErrTemplateRootMissing` | `collage: template root missing` | `Template.Root` does not exist or is not a directory. | The most common startup failure: check the working directory, or embed the templates. |
| `ErrTemplateEscapesRoot` | `collage: template escapes root` | A template under `Template.Root` resolves outside it — a symlink leading out of the directory. | Copy the file in rather than linking to it. |
| `ErrInvalidCacheType` | `collage: invalid cache type` | `Cache.Enabled`, no `Store`, and `Type` is neither `"memory"` nor `"disk"`. | Fix the spelling, or supply a `Store`. |
| `ErrEmptyCacheDir` | `collage: disk cache needs a directory` | `Cache.Type` is `"disk"` and `Cache.Dir` is empty. | Set `Dir`. |
| `ErrEmptyCacheVersion` | `collage: disk cache needs a version` | A disk cache was built with no version. | Normally unreachable: an empty `Version` is derived from the binary. |
| `ErrUnsupportedCache` | `collage: unsupported cache type` | The cache type names nothing the framework can build. | Normally caught earlier as `ErrInvalidCacheType`. |
| `ErrEmptyLocaleDefault` | `collage: empty default locale` | `Locale.Default` is empty. | Only reachable through `Validate`; `New` defaults it to `"en"`. |
| `ErrLocaleDefaultNotSupported` | `collage: default locale not in supported locales` | `Locale.Supported` does not include `Locale.Default`. | Add the default to `Supported`. |
| `ErrNegativeDuration` | `collage: negative duration` | One of the six duration fields is negative; the message names which. | Use zero for the default. |

## Plugins and commands

| Error | Message | Means | What to do |
| --- | --- | --- | --- |
| `ErrNilPlugin` | `collage: nil plugin` | A `nil` plugin was registered. | — |
| `ErrEmptyPluginName` | `collage: empty plugin name` | A plugin's `Name()` is empty. | Give it a name like `acme/stamp`. |
| `ErrDuplicatePlugin` | `collage: duplicate plugin` | Two plugins share a name. | Register each plugin once. |
| `ErrConfigurerRegisteredLate` | `collage: plugin needs Configure and must be supplied in Config.Plugins` | `RegisterPlugin` was given a plugin with a `Configure` phase, which has already passed. | Move it to `Config.Plugins`. |
| `ErrDuplicateTemplateFunc` | `collage: duplicate plugin template function` | `AddTemplateFunc` was called for a name already added — by another plugin, or by the same one a second time. | `AddTemplateFunc` returns it to the plugin's `Configure`; `New` fails only if `Configure` returns it. The plugins conflict: `Template.Funcs` cannot prevent it, so drop or rename one of them. |
| `ErrUnknownPluginConfig` | `collage: plugin configuration names no registered plugin` | A `PluginConfig` key matches no registered plugin. Checked when the application starts. | Almost always a typo in the key. |
| `ErrEmptyCommandName` | `collage: empty command name` | `RegisterCommand` with no name. | — |
| `ErrDuplicateCommand` | `collage: duplicate command name` | Two commands share a name. | — |
| `ErrNilApp` | `collage: nil app` | `DispatchCommands` was given a nil `*App`. | — |
| `ErrUnknownCommand` | `collage: unknown command` | `DispatchCommands` got no arguments, or a name no plugin registered. | The scaffolded `main.go` exits `2` on it, as a usage error; a program that would rather serve can fall through on it instead. |

See [Writing a plugin](/docs/writing-plugins).

## Fragments and pages

Two places report these. A few are recorded by the builders as the chain runs —
`ErrDuplicateSlot`, `ErrUnknownSlot` and `ErrSlotResolved` from `WithSlot`,
`WithSlotResolver` and `WithSlotFragment` (which also records `Bind`'s
`ErrNilFragment` and `ErrSlotOccupied`), `ErrInvalidTimeout` from `WithTimeout`,
`ErrMissingContent` from a page's `Build`, and `ErrNoDocumentHandler` from a
document's — and you can read them with `BuildErr()`. What a builder recorded stays
on the value it built, and `RegisterPage` and `RegisterDocument` refuse a value
carrying any — a page's own, or those of any fragment in its tree — wrapped as
`collage: page %q was built with errors: %w` (`collage: document %q was built with
errors: %w` for a document), whether or not `BuildErr()` was called.

The rest are found by validation when the page is registered: `RegisterPage` checks
every fragment in the tree, its names, templates, slots, timeouts, TTLs, paths and
error pages, and returns the first failure. A fragment opened with
`WithFragmentPath` is part of the page for this (since v0.11.0): its template, its
builder's mistakes and its validation are checked at registration like the rest. Either way a malformed page is refused
before it serves anything.

| Error | Message | Means |
| --- | --- | --- |
| `ErrEmptyName` | `collage: empty name` | A fragment or page has no name. |
| `ErrEmptyTemplatePath` | `collage: empty template path` | A fragment names no template. |
| `ErrNilFragment` | `collage: nil fragment` | A `nil` fragment was used where one is required — bound to a slot, or returned by a slot resolver. |
| `ErrDuplicateSlot` | `collage: slot already declared` | `WithSlot` was called twice with one name. |
| `ErrUnknownSlot` | `collage: unknown slot` | A fragment was bound to, or a template called `{{slot}}` for, a slot the fragment never declared. |
| `ErrInvalidSlotDefinition` | `collage: invalid slot definition` | A slot has an empty name, or a map key that does not match its own name. |
| `ErrSlotOccupied` | `collage: slot already occupied` | A second fragment was bound to a slot that holds one — or a resolver returned several for it. |
| `ErrSlotResolved` | `collage: slot is filled by a resolver` | One slot was given both a resolver and bound fragments. |
| `ErrRequiredSlotUnfilled` | `collage: required slot has no fill` | A slot declared required has nothing bound to it. |
| `ErrFragmentCycle` | `collage: fragment cycle detected` | A fragment is reachable from itself. |
| `ErrMissingContent` | `collage: missing content` | A page has no content fragment. |
| `ErrInvalidTimeout` | `collage: invalid timeout` | A fragment's timeout is negative. |
| `ErrMissingTTL` | `collage: missing cache ttl for incremental strategy` | `Incremental` was given a zero TTL. |
| `ErrInvalidTTL` | `collage: invalid cache ttl` | A page's TTL is negative. |
| `ErrInvalidPath` | `collage: invalid path` | A path pattern does not start with `/`. |
| `ErrInvalidRedirectStatus` | `collage: invalid redirect status code` | A redirect status other than `0`, `301`, `302`, `307` or `308`. |
| `ErrSelfErrorPage` | `collage: page cannot reference itself as an error page` | A page is its own not-found or error page. |

See [Pages and layouts](/docs/pages-and-layouts) and
[Fragments and slots](/docs/fragments-and-slots).

## Registration and routing

Returned by `RegisterPage`, `RegisterNotFoundPage`, `RegisterErrorPage` and
`RegisterDocument`, or when the application starts.

| Error | Message | Means | What to do |
| --- | --- | --- | --- |
| `ErrAppStarted` | `collage: application already started` | A registration method was called after the application started — including `RegisterPlugin` after any start that failed (since v0.12.0; after one that failed in a plugin's `Init` since v0.11.0). | Register everything before `Handler`, `ListenAndServe`, `Start`, `RenderPath` or `DispatchCommands`. |
| `ErrNilPage` | `collage: nil page` | A `nil` page was registered. | — |
| `ErrDuplicatePage` | `collage: duplicate page name` | Two pages share a name. | Names are how links find pages; make them unique. |
| `ErrTemplateNotFound` | `collage: template not found` | A page's fragment names a template that was not loaded. | Check the path relative to `Template.Root`, extension included. |
| `ErrUnregisteredErrorPage` | `collage: error page not registered` | A page names a not-found or error page that was never registered. Checked at start. | Register it with `RegisterNotFoundPage` or `RegisterErrorPage`; an unregistered one would render empty when needed. |
| `ErrInvalidPattern` | `collage: invalid pattern` | A path or redirect source is malformed: no leading `/`, an empty segment, an empty placeholder name, a catch-all that is not last, or — since v0.11.0 — a placeholder inside a segment, such as `/feeds/{category}.xml`. | Fix the pattern. A placeholder is a whole segment: `/feeds/{category}/rss.xml`. |
| `ErrDuplicateRoute` | `collage: duplicate route` | A path or redirect source is already registered in that locale. | — |
| `ErrAmbiguousParameterName` | `collage: ambiguous parameter name` | Two patterns use different parameter names at one position, such as `/blog/{slug}` and `/blog/{id}/edit`. | Use one name at that position. |
| `ErrRedirectShadowsPage` | `collage: redirect shadows a registered page` | A redirect's source is also a page's path. | One of the two would be unreachable; remove one. |
| `ErrUnsubstitutedPlaceholder` | `collage: redirect placeholder not captured by from pattern` | A redirect's destination uses a `{name}` its source does not capture. | Capture it in the source, or remove it. |

## Documents

| Error | Message | Means |
| --- | --- | --- |
| `ErrNilDocument` | `collage: nil document` | A `nil` document was registered. |
| `ErrEmptyContentType` | `collage: empty content type` | A document declares no content type. It is required and never guessed. |
| `ErrNoDocumentHandler` | `collage: document has no handler` | A document has no handler. Unlike a page, it has no template to fall back on. |
| `ErrDuplicateDocument` | `collage: duplicate document name` | Two documents share a name. |
| `ErrDocumentNotFound` | `collage: no document at path` | `RenderDocumentPath` found no document at the path — including when a page or a redirect is there. |
| `ErrEmptyDocumentBody` | `collage: document handler produced an empty body` | A handler succeeded with an empty body: a 500 when served, and not written by a build. A handler that really means "empty" can return a single newline. |

See [Documents](/docs/documents).

## Mounts and handlers

| Error | Message | Means |
| --- | --- | --- |
| `ErrUnknownAsset` | `collage: unknown asset` | `rc.Asset`, `rc.HoistStylesheet`, `{{asset}}` or `{{stylesheet}}` was given a path no mount serves. In a template, the render fails. |
| `ErrNoMountForAsset` | `collage: no mount serves that asset` | Wrapped inside `ErrUnknownAsset`, when no mount's prefix covers the path at all — as opposed to a mount that has no such file. |
| `ErrInvalidPrefix` | `collage: invalid mount prefix` | A mount prefix does not begin and end with `/`, is `/` alone, or begins with `//`. A mount at `/` would swallow every route. |
| `ErrNilFS` | `collage: nil mount file system` | A mount was given no filesystem. |
| `ErrMountConflict` | `collage: mount prefixes overlap` | Two mounts — or a mount and an `App.Handle` prefix — claim overlapping prefixes. |
| `ErrMountShadowsRoute` | `collage: mount shadows a route` | A mount prefix, or an `App.Handle` prefix, would swallow a route's path — a page's, a document's, a redirect's or an action's. Checked at start, whatever the registration order. |
| `ErrInvalidHandlerPrefix` | `collage: handler prefix must begin and end with "/" and not be "/"` | `App.Handle` was given a bad prefix. |
| `ErrNilHandler` | `collage: nil handler` | `App.Handle` or `App.Use` was given nothing to run. |

See [Static assets](/docs/assets) and
[Middleware and your own API](/docs/middleware-and-apis).

## Rendering

| Error | Message | Means | What to do |
| --- | --- | --- | --- |
| `ErrNotFound` | `collage: not found` | **You return this one.** A data handler that wraps it says the content does not exist, rather than failed to load. | `fmt.Errorf("post %q: %w", slug, collage.ErrNotFound)` makes a required fragment's failure a 404 with the not-found page, not a 500. |
| `ErrRequiredSlotEmpty` | `collage: required slot is empty` | At render time, a required slot holds nothing — typically a resolver returned no fragments. | Subject to the fragment's failure policy. |
| `ErrMaxDepthExceeded` | `collage: max fragment depth exceeded` | The fragment tree nests deeper than the engine allows. | Almost always a fragment bound into its own slot, directly or not. |
| `ErrNoRootFragment` | `collage: page has no root fragment` | A page has neither a layout nor content. | — |
| `ErrPageNotFound` | `collage: no page at path` | `App.RenderPath` found no page at the path. | — |
| `ErrOnceTypeMismatch` | `collage: once key fetched as two different types` | Two `collage.Once` calls asked for one key as different types in one render. | The keys collide; namespace them. |
| `ErrCachedTypeMismatch` | `collage: Cached key holds a value of a different type` | Two `collage.Cached` calls asked for one key as different types. | As above. |
| `ErrDictOddArgs` | `collage: dict requires an even number of arguments` | `{{dict}}` was given a key with no value. | Pair every key with a value. |
| `ErrDictKeyNotString` | `collage: dict key must be a string` | A `{{dict}}` key is not a string. | Quote the key. |
| `ErrCSRFDisabled` | `collage: csrfToken used but request-forgery protection is disabled` | A template calls `{{csrfToken}}` in an application with `Security.DisableCSRF` set. | Remove the call, or turn protection back on. |

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

## Links and URLs

Returned by `App.URL` and failed renders from `{{pageURL}}`, `{{pageURLIn}}` and
`{{localeURL}}`. See [Links and locales](/docs/links-and-locales).

| Error | Message | Means |
| --- | --- | --- |
| `ErrUnknownRoute` | `collage: no page or document by that name` | No page or document is registered under the name — or both are, and the link is ambiguous. |
| `ErrNoPathInLocale` | `collage: no path in that locale` | The route has no path in the locale asked for. `{{pageURL}}` falls back to the default locale instead, and `{{localeURL}}` renders the empty string. |
| `ErrRouteParams` | `collage: route parameters do not match the pattern` | A parameter is missing or empty, names no placeholder, is `.` or `..`, or the template passed an odd number of arguments. |
| `ErrLocaleUnreachable` | `collage: no URL reaches that locale` | The locale is not in `Locale.Supported`, or it is not the default and path locales are disabled. |

## Actions

Returned by `RegisterAction`. `RegisterPage` checks an action attached to a page
too — `ErrNilAction`, `ErrNoMethods`, and since v0.10.0 `ErrNoActionHandler` and
`ErrDuplicateAction` — and refuses a `nil` fragment path with
`ErrNilFragmentPath`. Since v0.11.0 there is one `ErrNoActionHandler`, the same
value at registration and on a request.

| Error | Message | Means | What to do |
| --- | --- | --- | --- |
| `ErrNilAction` | `collage: nil action` | A `nil` action was registered, or attached to a page. | — |
| `ErrEmptyActionName` | `collage: action has no name` | An action has no name. | Give it one; it names the action in logs and errors. |
| `ErrDuplicateAction` | `collage: duplicate action` | A second action was registered under a name already taken. | — |
| `ErrNoActionPaths` | `collage: action has no paths` | A standalone action has no `WithPath`. | An action on a page takes the page's paths; one on its own needs its own. |
| `ErrNoActionHandler` | `collage: action has no handler` | An action has no `WithHandler`. | — |
| `ErrNoMethods` | `collage: action declares no methods` | An action answers no method. | `WithMethods(http.MethodPost)`, or `WithAction` on a page. |
| `ErrNilFragmentPath` | `collage: fragment path has no fragment` | `WithFragmentPath` was given a `nil` fragment. | — |
| `ErrUnregisteredPage` | `collage: action answered with a page that was never registered` | An action's `RenderPage` returned a page that was not registered. The request fails with a 500. | Register the page, and answer with that same value rather than building one in the handler. |

See [Forms and actions](/docs/forms-and-actions).

## Vary and SkipCache

Returned by `collage.Vary` and `collage.SkipCache`, which must be called from
middleware. See [Caching](/docs/caching) and [Previews](/docs/previews).

| Error | Message | Means | What to do |
| --- | --- | --- | --- |
| `ErrVaryTooLate` | `collage: Vary or SkipCache called after routing; call it from middleware` | `Vary` or `SkipCache` was called after routing began — from a data handler, say. Since v0.11.0 on every route; before, only on a cached page, and a late call elsewhere did nothing. | Call it from middleware registered with `App.Use`. |
| `ErrVaryOutsideRequest` | `collage: Vary called on a request collage is not serving` | `Vary` or `SkipCache` was called on a request that did not come through collage's handler. | — |

## Request forgery

A submission that fails the forgery check is answered with a **403** before the
action's handler runs, and error hooks receive the reason under the stage
`"route"`, wrapped with the action's name:

| Error | Message | Means |
| --- | --- | --- |
| `ErrCSRFMissing` | `collage: no csrf token` | The submission carried no token, or no cookie. |
| `ErrCSRFMismatch` | `collage: csrf token does not match` | The token is not the one in the cookie. |
| `ErrCSRFInvalid` | `collage: csrf token is not valid` | The token was not signed with this application's key. |

The usual causes are a form with no `{{csrfToken}}`, and a key generated per
process — set `Security.CSRFKey` so tokens survive restarts and work across
instances. A body over the size limit is a **413** instead, even when reading the
token is what hit the limit. See
[Forms and actions](/docs/forms-and-actions#forgery-protection).

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
| `ErrMethodNotAllowed` | `collage: method not allowed` | `route` | The path exists but answers no such method: a 405, with an `Allow` header naming what it does answer. On a document's URL the 405 is plain text (since v0.11.0). |
| `ErrEmptyRender` | `collage: page rendered no markup` | `render` | A page rendered successfully but produced no markup, served or answered by an action: a 500. The same sentinel a static build records. |
| `ErrCSRFMissing`, `ErrCSRFMismatch`, `ErrCSRFInvalid` | see [above](#request-forgery) | `route` | A submission refused by the forgery check. |
| `ErrEmptyErrorPage` | `collage: error page rendered empty` | `error_page` | A registered error page rendered successfully but produced no markup, so the built-in page was served instead. |
| `ErrPanic` | `collage: panic recovered while serving the request` | `panic` | Something panicked while serving — a `Cache`, `Metrics` or `Tracer` implementation, a router, a plugin hook — and was recovered into a 500. Panics in data handlers and templates are `PanicError` instead. |
| `ErrAssetFailed` | `collage: asset request failed` | `asset` | A mounted file request answered with a status of 400 or above: one sentinel for every such status. |
| `ErrHandlerFailed` | `collage: mounted handler failed` | `handler` | A handler mounted with `App.Handle` answered with a server error. |
| `ErrUnsafeRedirectTarget` | `collage: unsafe redirect target` | `route` | A redirect's destination, after substitution, is not a single-slash relative path — `//host`, `/\host`, or one with a control character. Answered with a 500, not a `Location` header. |

`"error_page"` is the stage worth alerting on: the page that reports failures
failed, and the reader still saw a plausible page, so nothing else would tell you.

## Static builds

Returned by `collage.NewBuilder` and `Builder.Build`, or recorded in the
`BuildReport`. See [Static export](/docs/static-export).

| Error | Message | Where | Means |
| --- | --- | --- | --- |
| `ErrNilRenderer` | `collage: nil renderer` | `NewBuilder` | The app is `nil`. |
| `ErrInvalidOutDir` | `collage: invalid output directory` | `NewBuilder` | `BuildOptions.OutDir` is empty. |
| `ErrDangerousOutDir` | `collage: refusing to use a dangerous output directory` | `Build` | `OutDir` resolves to a filesystem root — or, with `Clean`, to a repository root. |
| `ErrOutputPathCollision` | `collage: two builds target one output path` | `Build` | Two pages would be written to one file — patterns differing only by a trailing slash, or a path provider returning a path twice. Reported before any page renders, and then no page is: documents, `404.html` and assets are still written. |
| `ErrPathEscapesOutDir` | `collage: resolved path escapes the output directory` | report error | An output path, or a symlink on the way to it, leads outside `OutDir`. Checked before that one file is written; only that path fails. |
| `ErrDynamicPathUnresolved` | `collage: dynamic path pattern requires a path provider` | skip | A page's or document's path has a `{param}` and there is no path provider. |
| `ErrNotStatic` | `collage: a Dynamic() route cannot be built statically` | skip | A page or document is `Dynamic()`, so there is nothing to export. |
| `ErrDuplicateOutputPath` | `collage: two build tasks write the same output path` | skip | Two document tasks resolve to one file — a `DocumentPathProvider` returning one path twice. The first is built, the rest skipped. |
| `ErrDegradedRender` | `collage: refusing to write a degraded render` | report error | A page rendered with a failed fragment and `AllowDegraded` is off. No file is written. |
| `ErrEmptyRender` | `collage: page rendered no markup` | report error | A page rendered no markup at all. Refused even with `AllowDegraded`. One sentinel with serving's, above. |
| `ErrUnresolvedToken` | `collage: refusing to write a page whose forgery token was never resolved` | report error, skip | The not-found page carries a `{{csrfToken}}`: a report error. Any other page with one is skipped instead: it needs a server. |
| `ErrBuildPanic` | `collage: panic while building a page` | report error | Rendering or writing one page panicked; the build recovered and went on with the rest. |
| `ErrEmptyDocumentBody` | see [Documents](#documents) | report error | A document produced an empty body. |

Report errors are in `BuildReport.Errors`, a slice of errors that match with
`errors.Is`. A skip is a `SkipRecord` in `BuildReport.Skipped`: its `Reason` is a
sentence for people, and its `Err` (since v0.10.0) is the sentinel for code —
`ErrNotStatic`, `ErrDynamicPathUnresolved`, `ErrUnresolvedToken` or
`ErrDuplicateOutputPath` — so match it with `errors.Is(skip.Err, …)` rather than by
reading `Reason`.
