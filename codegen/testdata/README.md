# GCC Bytecode Compile Fixtures

`gcc-bytecode-compile.tsv` lists GCC accept fixtures that currently compile through:

`preprocessor -> parser -> sema -> codegen -> bytecode validator`

The source files stay under `sema/testdata/gcc-*`; this manifest only defines the
bytecode-compile subset. GCC reject cases remain frontend-only because they should
not reach codegen. GCC accept cases omitted from this file are still accepted by
the frontend, but require codegen features that are not part of the current
bytecode generation coverage.
