# Parser (Syntactic Analysis)

1. Using GLR to parse lexical tokens into AST Tree
2. Parsing AST Tree into more 'humanized' struct

Productions in phrase_structure_grammar.txt from _ISO/IEC 9899:TC3 A.2 Phrase structure grammar_

## Output contract

`Parse()` returns `[]*entity.AstNode` with **length ≥ 1** on success. The parser
applies two cheap disambiguators internally:

- The dangling-else conflict is resolved at table-generation time by `shaveDFA`
  (prefers shift, i.e. `else` binds to the nearest `if`).
- The two `typedef_name` conflicts are resolved at fork time by consulting the
  current typedef stack (`pruneTypedefFork`). Identifiers known to be typedefs
  take the type path; others take the variable path.

Anything beyond these — true semantic ambiguity that depends on full type
information, scoped symbol tables, or C standard tie-breaker rules
(e.g. `(T)+x` cast-vs-paren) — is **not** resolved here. The intended consumer
is the future semantic analysis layer, which receives the candidate forest and
makes the final selection.

On failure, `Parse()` returns `nil` and the error from the branch that consumed
the most input (best-error heuristic). Errors from branches that died early
are suppressed.