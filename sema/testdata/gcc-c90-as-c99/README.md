# GCC C89/C90 As C99 Fixtures

这个目录收录 `/tmp/cvm-gcc-c99/gcc/testsuite/gcc.dg` 中显式使用 C89/C90
方言的用例，并按 C99 前端语义重新分类。

筛选范围：

- `-std=c89`
- `-std=c90`
- `-std=iso9899:1990`
- `-ansi`

暂不包含 GNU89/GNU90 专属扩展测试。原始 GCC DejaGNU 期望不能直接复用，因为许多
C90 reject 用例在 C99 下反而应当 accept，例如混合声明和语句、`for` 初始化声明、复合字面量、
VLA、`long long` 等。

当前门禁只验证 `lexer -> parser -> sema.Analyze` 能直接判定的前端行为：

- `accept/`：按 C99 语义应接受，且当前前端也应接受。
- `reject/`：按 C99 语义应拒绝，且当前前端也应拒绝。
- `manifest.tsv`：完整清单。没有导入的用例必须保留明确 skip 原因。

截至本次导入，清单共 112 个显式 C89/C90 用例：

- 已导入 90 个：32 个 accept，58 个 reject。
- 跳过 22 个：主要原因是 GCC target/runtime harness、GNU 扩展、GCC builtin、
  GCC warning/diagnostic 行为，以及当前前端尚未覆盖的 C99 边界诊断。
