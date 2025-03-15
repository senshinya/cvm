package main

import (
	"testing"
)

func TestError(t *testing.T) {
	(&Compiler{}).RunSource(`typedef int a;
int main() {
	int a;
	int b;
	a*b;
}`)
}

func TestAbstractDeclaration(t *testing.T) {

	(&Compiler{}).RunSource(`void func2(int (*)[10]);

int main()
{
  func2();
  return 0;
}`)
}

func TestEnumDeclaration(t *testing.T) {
	(&Compiler{}).RunSource(`enum a { b, c };
int main() {
	enum a d = b;
}`)
}

func TestStructDeclaration(t *testing.T) {
	(&Compiler{}).RunSource(`struct a { int b; };
int main() {
	int ccc = b;
	struct a c;
	c.b = 1;
	return 0;
}`)
}

func TestFunctionDeclaration(t *testing.T) {
	(&Compiler{}).RunSource(`int (*f)();
int main() {
	f();
}`)
}

func TestKAndRFunctionDeclaration(t *testing.T) {
	(&Compiler{}).RunSource(`void example(a, b, c)
    int a;
    char b;
    float c;
{
	int ccc = a;
}`)
}

func TestFunctionParameterShadow(t *testing.T) {
	(&Compiler{}).RunSource(`typedef int a;
void example(int a) {
	a*b;
}`)
}
