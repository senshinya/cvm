# GCC C99 Extra Fixtures

这个目录收录 `/tmp/cvm-gcc-c99/gcc/testsuite/gcc.dg` 下非 `c99*.c`、但在 DejaGNU
`dg-options` 或 `dg-additional-options` 中显式使用 C99/GNU99 方言的用例。

当前门禁只验证 `lexer -> parser -> sema.Analyze` 能直接判定的前端行为：

- `accept/`：GCC 期望接受，且当前前端也应接受。
- `reject/`：GCC 期望拒绝，且当前前端也应拒绝。
- `manifest.tsv`：完整清单。没有导入的用例必须保留明确 skip 原因。

截至本次导入，清单共 333 个显式 C99/GNU99 非 `c99*.c` 用例：

- 已导入 144 个：54 个 accept，90 个 reject。
- 跳过 189 个：主要原因是 GCC target/runtime harness、GNU 扩展、GCC builtin、GCC inline
  方言/代码生成、GCC
  warning/diagnostic 行为，以及少量尚未覆盖的 C99 边界语义。

这里的 skip 不表示这些语义永远不做；它们是后续 C99 合规工作的待办分类。主
`gcc-c99` 目录仍维持更严格策略：除预处理器/系统宏依赖外，`c99*.c` 用例都应进入门禁。
