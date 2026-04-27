// Package sema 实现 cvm 编译器的 C99 语义分析。
//
// Sema 接收 parser 包产出的候选解析森林，先运行 PreFilter 丢弃明显错误的
// typedef 解释，再对每个候选运行语义分析 walker（Plan B/C）并为 codegen
// 产出带类型的 Program IR。
//
// 参见 docs/superpowers/specs/2026-04-27-sema-layer-design.md。
package sema
