来源：GCC testsuite `gcc/testsuite/gcc.dg`。

导入提交：`b3ba4c3336c341596094672c635bd99971f36d8c`

当前只纳入 Plan B 语义层可稳定覆盖的 C99 用例子集：

- `accept/`：GCC 期望 C99 编译接受的用例。
- `reject/`：GCC 期望 C99 编译拒绝的约束用例。

测试 runner 会去掉 DejaGNU 指令注释和普通 C 块注释，然后走 cvm 当前的
lexer/parser/PreFilter/Sema 流水线。尚未纳入依赖预处理器、系统头、GCC 扩展、
代码生成或运行时执行结果的用例。
