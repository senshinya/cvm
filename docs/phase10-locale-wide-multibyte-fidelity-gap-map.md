# Phase 10 Locale Wide Multibyte Fidelity Gap Map

Phase 10 is closed on `codex/bytecode-runtime-phase-10`.

## Closed Surface

The runtime now has direct extern and source-level execution coverage for:

- Deterministic C-locale category storage through `setlocale` for every standard `LC_*` category.
- `localeconv` static per-memory `struct lconv` materialization, including stable pointer fields, C-locale separators, and `CHAR_MAX` numeric fields.
- Builtin `<locale.h>`, `<wchar.h>`, and `<wctype.h>` declarations for the Phase 10 runtime surface.
- `wchar_t`, `wint_t`, `mbstate_t`, `WEOF`, wide ctype descriptors, and wide case-conversion descriptors in builtin headers.
- Wide ctype classification for ASCII C-locale classes, EOF handling, and non-ASCII non-classification.
- Wide case conversion through `towlower`, `towupper`, `wctrans`, and `towctrans`.
- `wctype` and `iswctype` descriptor lookup for standard C-locale wide classes.
- Restartable C-locale multibyte conversion through `mbrlen`, `mbrtowc`, `wcrtomb`, `mbsrtowcs`, and `wcsrtombs`.
- Source-level restartable multibyte workflows covering length queries, NULL reset behavior, completion, truncation, and invalid high-byte/high-wide-character paths.

## Recheck Results

- Header declarations were rechecked for `<locale.h>`, `<wchar.h>`, and `<wctype.h>`.
- `DefaultExternRegistry` registrations were rechecked for `localeconv`, wide ctype, wide case conversion, descriptor lookup, and restartable multibyte externs.
- Direct extern and source-level runtime tests cover the Phase 10 locale, wide-character, descriptor, and restartable multibyte surfaces.
- `env GOCACHE=/private/tmp/cvm-go-build-cache go test ./runtime -run TestGCCExecutionGapReportIsCurrent -count=1` passed.
- Imported GCC accept scans for locale, wide-character, and restartable multibyte terms found no new low-risk runtime fixture beyond existing coverage.

## Residual Limits

- Locale behavior remains deterministic C-locale behavior; host locale names, locale databases, collation, and localized numeric formatting remain out of scope.
- `mbstate_t` is opaque and effectively stateless for the supported C-locale single-byte conversions.
- Restartable multibyte conversion supports ASCII and NUL bytes only; stateful encodings and UTF-8 multibyte sequences are not modeled.
- Wide ctype classification is C-locale ASCII classification; non-ASCII wide values are treated as unclassified and unchanged by case conversion.
- `wcrtomb`, `mbsrtowcs`, and `wcsrtombs` use deterministic C-locale invalid-character failures instead of platform-specific encoding behavior.
