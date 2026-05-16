package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"shinya.click/cvm/bytecode"
	"shinya.click/cvm/codegen"
	"shinya.click/cvm/common"
	"shinya.click/cvm/entity"
	"shinya.click/cvm/parser"
	"shinya.click/cvm/preprocessor"
	"shinya.click/cvm/sema"
	"strings"
)

type Compiler struct {
	FileName     string
	Source       string
	Lines        []string
	Sources      *preprocessor.SourceManager
	DumpIR       bool
	DumpBytecode bool
	EmitBytecode string
	Output       io.Writer
}

func (c *Compiler) RunSource(source string) error {
	if c.FileName == "" {
		c.FileName = "main.c"
	}
	if err := c.validateDumpModes(); err != nil {
		return err
	}
	c.Source = source
	c.Lines = strings.Split(source, "\n")
	pp, err := preprocessor.PreprocessSource(c.FileName, source, preprocessor.Options{})
	if err != nil {
		return err
	}
	c.Sources = pp.Sources
	candidates, err := parser.NewParser(pp.Tokens).Parse()
	if err != nil {
		return err
	}
	prog, err := sema.Analyze(candidates)
	if err != nil {
		return err
	}
	if c.DumpIR {
		fmt.Fprint(c.output(), sema.PrintProgram(prog))
		return nil
	}
	if c.DumpBytecode {
		mod, err := codegen.Generate(prog)
		if err != nil {
			return err
		}
		fmt.Fprint(c.output(), bytecode.PrintModule(mod))
		return nil
	}
	if c.EmitBytecode != "" {
		mod, err := codegen.Generate(prog)
		if err != nil {
			return err
		}
		f, err := os.Create(c.EmitBytecode)
		if err != nil {
			return err
		}
		if err := bytecode.EncodeModule(f, mod); err != nil {
			_ = f.Close()
			return err
		}
		return f.Close()
	}
	return nil
}

func (c *Compiler) output() io.Writer {
	if c.Output != nil {
		return c.Output
	}
	return os.Stdout
}

func (c *Compiler) validateDumpModes() error {
	selected := 0
	for _, enabled := range []bool{c.DumpIR, c.DumpBytecode, c.EmitBytecode != ""} {
		if enabled {
			selected++
		}
	}
	if selected > 1 {
		return fmt.Errorf("--dump-ir, --dump-bytecode, and --emit-bytecode are mutually exclusive")
	}
	return nil
}

func (c *Compiler) RunFile(fileName string) error {
	c.FileName = fileName
	if err := c.validateDumpModes(); err != nil {
		return err
	}
	source, err := os.ReadFile(fileName)
	if err != nil {
		return err
	}
	return c.RunSource(string(source))
}

func (c *Compiler) handleError(err error) {
	var cvmError *common.CvmError
	switch {
	case errors.As(err, &cvmError):
		for _, message := range cvmError.Messages {
			file, line, column, text := c.displayErrorLocation(message.SourcePos)
			fmt.Printf("%s:%d:%d: %s %s\n", file,
				line, column,
				common.IfElse(message.Level == common.MessageLevelError, common.RedText("error:"), common.GrayText("note:")),
				message.CustomMessage)
			fmt.Printf("    %d | %s\n", line, text)
			fmt.Printf("    %s | ", spaceByStringLength(fmt.Sprintf("%d", line)))
			for i := 0; i < column-1; i++ {
				fmt.Print(" ")
			}
			fmt.Printf("%s\n", common.GreenText("^"))
		}
	default:
		fmt.Println(err.Error())
	}
}

func (c *Compiler) displayErrorLocation(pos entity.SourcePos) (string, int, int, string) {
	if c.Sources != nil {
		display := c.Sources.DisplayLocation(pos)
		file := display.File
		if file == "" {
			file = c.FileName
		}
		if display.Line > 0 && display.Column > 0 {
			return file, display.Line, display.Column, display.Text
		}
	}
	line, column := pos.Line, pos.Column
	if line <= 0 {
		line = 1
	}
	if column <= 0 {
		column = 1
	}
	text := ""
	if line-1 >= 0 && line-1 < len(c.Lines) {
		text = c.Lines[line-1]
	}
	return c.FileName, line, column, text
}

func spaceByStringLength(str string) string {
	spaces := ""
	for i := 0; i < len(str); i++ {
		spaces += " "
	}
	return spaces
}
