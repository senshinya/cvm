来源：GCC testsuite `gcc/testsuite/gcc.dg`。

导入提交：`b3ba4c3336c341596094672c635bd99971f36d8c`

当前纳入当前前端可稳定覆盖的 C99 用例子集：

- `accept/`：27 个 GCC 期望 C99 编译接受，且当前 `lexer -> parser -> sema.Analyze` 接受的用例。
- `reject/`：35 个 GCC 期望 C99 编译拒绝，且当前 `lexer -> parser -> sema.Analyze` 拒绝的用例。
- `manifest.tsv`：记录从 `gcc.dg/c99*.c` 扫描到的 135 个候选，以及每个候选的导入/跳过原因。

测试 runner 会去掉 DejaGNU 指令注释和普通 C 块注释，然后走 cvm 当前的
lexer/parser/Sema 聚合入口。尚未纳入依赖预处理器、系统头、GCC 扩展、
代码生成或运行时执行结果的用例；这些候选会保留在 `manifest.tsv` 中并标注
skip 原因。

严格 skip 策略：

- 只有依赖预处理器、系统头、GCC predefined macro/type 的用例可以保留为 `skipped`。
- GCC warning-only 用例按 accept 处理，因为当前 cvm 没有 warning 通道。
- 没有 `dg-error` 的 GCC C99 用例必须被 `lexer -> parser -> sema.Analyze` 接受。
- 有 `dg-error` 的 GCC C99 用例必须被 `lexer -> parser -> sema.Analyze` 拒绝。
