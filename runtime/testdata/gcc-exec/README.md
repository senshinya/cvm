# GCC Execution Fixtures

This directory is for GCC-derived execution fixtures that are deterministic
under the cvm bytecode runtime.

Each fixture must document:

- expected exit code;
- required externs;
- skip reason when unsupported;
- whether hosted C library behavior is required.

Do not copy the compile-only GCC fixture set wholesale. Add execution fixtures
only when the runtime implements the required bytecode and extern behavior.
