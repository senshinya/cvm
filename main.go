package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	cvmruntime "shinya.click/cvm/runtime"
)

func main() {
	os.Exit(runMain(os.Args[1:]))
}

func runMain(args []string) int {
	if len(args) > 0 && args[0] == "run" {
		return runBytecode(args[1:])
	}
	return runCompileMode(args)
}

func runBytecode(args []string) int {
	cfg, err := parseRunBytecodeArgs(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, "Usage: cvm run [--stdin text] [--env NAME=VALUE] file.cvmbc [args...]")
		return 2
	}
	f, err := os.Open(cfg.file)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer f.Close()
	var stdin io.Reader
	if cfg.stdinSet {
		stdin = strings.NewReader(cfg.stdin)
	}
	reg := cvmruntime.DefaultExternRegistryWithIO(stdin, nil, nil)
	for _, env := range cfg.env {
		name, value, _ := strings.Cut(env, "=")
		reg.SetEnv(name, value)
	}
	progArgs := append([]string{cfg.file}, cfg.programArgs...)
	prog, err := cvmruntime.Load(f, cvmruntime.LoadOptions{Args: progArgs, Externs: reg})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	st, err := cvmruntime.Run(context.Background(), prog, cvmruntime.RunOptions{})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return st.Code
}

type runBytecodeConfig struct {
	file        string
	programArgs []string
	stdin       string
	stdinSet    bool
	env         []string
}

func parseRunBytecodeArgs(args []string) (runBytecodeConfig, error) {
	var cfg runBytecodeConfig
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--":
			i++
			if i >= len(args) {
				return cfg, fmt.Errorf("missing bytecode file")
			}
			cfg.file = args[i]
			cfg.programArgs = append([]string(nil), args[i+1:]...)
			return cfg, nil
		case arg == "--stdin":
			i++
			if i >= len(args) {
				return cfg, fmt.Errorf("missing value for --stdin")
			}
			cfg.stdin = args[i]
			cfg.stdinSet = true
		case strings.HasPrefix(arg, "--stdin="):
			cfg.stdin = strings.TrimPrefix(arg, "--stdin=")
			cfg.stdinSet = true
		case arg == "--env":
			i++
			if i >= len(args) {
				return cfg, fmt.Errorf("missing value for --env")
			}
			if err := validateRunEnv(args[i]); err != nil {
				return cfg, err
			}
			cfg.env = append(cfg.env, args[i])
		case strings.HasPrefix(arg, "--env="):
			env := strings.TrimPrefix(arg, "--env=")
			if err := validateRunEnv(env); err != nil {
				return cfg, err
			}
			cfg.env = append(cfg.env, env)
		default:
			cfg.file = arg
			cfg.programArgs = append([]string(nil), args[i+1:]...)
			return cfg, nil
		}
	}
	return cfg, fmt.Errorf("missing bytecode file")
}

func validateRunEnv(env string) error {
	name, _, ok := strings.Cut(env, "=")
	if !ok || name == "" {
		return fmt.Errorf("--env expects NAME=VALUE")
	}
	return nil
}

func runCompileMode(args []string) int {
	dumpIR := false
	dumpBytecode := false
	emitBytecode := ""
	files := make([]string, 0, 1)
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--dump-ir":
			dumpIR = true
		case "--dump-bytecode":
			dumpBytecode = true
		case "--emit-bytecode":
			i++
			if i >= len(args) {
				fmt.Println("Usage: cvm [--dump-ir|--dump-bytecode|--emit-bytecode out.cvmbc] [file]")
				return 2
			}
			emitBytecode = args[i]
		default:
			files = append(files, arg)
		}
	}
	if len(files) != 1 {
		fmt.Println("Usage: cvm [--dump-ir|--dump-bytecode|--emit-bytecode out.cvmbc] [file]")
		return 2
	}
	c := &Compiler{DumpIR: dumpIR, DumpBytecode: dumpBytecode, EmitBytecode: emitBytecode}
	if err := c.RunFile(files[0]); err != nil {
		c.handleError(err)
		return 1
	}
	return 0
}
