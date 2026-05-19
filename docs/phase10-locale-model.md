# Phase 10 Deterministic Locale Model

Phase 10 keeps locale behavior deterministic and C-locale oriented. The runtime does not consult the host locale.

## Category State

- Supported categories are the current `<locale.h>` constants: `LC_ALL`, `LC_COLLATE`, `LC_CTYPE`, `LC_MONETARY`, `LC_NUMERIC`, and `LC_TIME`.
- Every category's effective locale is `"C"`.
- `setlocale(category, NULL)` queries the effective locale and returns `"C"` for supported categories.
- `setlocale(category, "C")` accepts the request and leaves the effective locale as `"C"`.
- `setlocale(category, "")` accepts the request as the deterministic hosted default and leaves the effective locale as `"C"`.
- Unsupported category numbers return `NULL`.
- Unsupported locale strings return `NULL` and do not mutate category state.

## Storage

- Returned locale names are per-memory static strings managed by `ExternRegistry.staticCString`.
- Repeated successful `setlocale` calls may return the same stable pointer for a given memory.
- Different `Memory` instances receive independent static storage.

## Follow-On Surfaces

- `localeconv` should expose a deterministic C-locale `struct lconv` whose string fields use per-memory static C strings.
- Wide-character classification and multibyte conversion should continue to model the C locale only: ASCII values are valid single-byte characters and high-bit bytes/wide characters are rejected where the existing byte-oriented C-locale helpers reject them.
