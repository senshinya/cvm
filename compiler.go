package main

import (
	"errors"
	"fmt"
	"os"
	"shinya.click/cvm/common"
	"shinya.click/cvm/lexer"
	"shinya.click/cvm/parser"
	"strings"
)

type Compiler struct {
	FileName string
	Source   string
	Lines    []string
}

func (c *Compiler) RunSource(source string) {
	if c.FileName == "" {
		c.FileName = "main.c"
	}
	c.Source = source
	c.Lines = strings.Split(source, "\n")
	tokens, err := lexer.NewLexer(c.Source).ScanTokens()
	if err != nil {
		c.handleError(err)
		return
	}
	_, err = parser.NewParser(tokens).Parse()
	if err != nil {
		c.handleError(err)
		return
	}
}

func (c *Compiler) RunFile(fileName string) {
	c.FileName = fileName
	source, err := os.ReadFile(fileName)
	if err != nil {
		c.handleError(err)
		return
	}
	c.RunSource(string(source))
}

func (c *Compiler) handleError(err error) {
	var cvmError *common.CvmError
	switch {
	case errors.As(err, &cvmError):
		for _, message := range cvmError.Messages {
			fmt.Printf("%s:%d:%d: %s %s\n", c.FileName,
				message.SourcePos.Line, message.SourcePos.Column,
				common.IfElse(message.Level == common.MessageLevelError, common.RedText("error:"), common.GrayText("note:")),
				message.CustomMessage)
			fmt.Printf("    %d | %s\n", message.SourcePos.Line, c.Lines[message.SourcePos.Line-1])
			fmt.Printf("    %s | ", spaceByStringLength(fmt.Sprintf("%d", message.SourcePos.Line)))
			for i := 0; i < message.SourcePos.Column-1; i++ {
				fmt.Print(" ")
			}
			fmt.Printf("%s\n", common.GreenText("^"))
		}
	default:
		fmt.Println(err.Error())
	}
}

func spaceByStringLength(str string) string {
	spaces := ""
	for i := 0; i < len(str); i++ {
		spaces += " "
	}
	return spaces
}
