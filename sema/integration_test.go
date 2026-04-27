package sema

import (
	"testing"

	"shinya.click/cvm/lexer"
	"shinya.click/cvm/parser"
)

func TestIntegrationHelloWorldAndFactorial(t *testing.T) {
	hello := `int printf(const char *fmt, ...);
int main() {
	printf("hello\n");
	return 0;
}`
	r := analyzeSource(t, hello)
	if len(r.Errors) != 0 {
		t.Fatalf("hello errors: %v", r.Errors)
	}
	if len(r.Program.Funcs) != 1 || r.Program.Funcs[0].Sym.Name != "main" {
		t.Fatalf("expected main, got %+v", r.Program.Funcs)
	}
	fact := `int factorial(int n) {
	if (n <= 1) return 1;
	return n * factorial(n - 1);
}`
	r = analyzeSource(t, fact)
	if len(r.Errors) != 0 {
		t.Fatalf("factorial errors: %v", r.Errors)
	}
}

func analyzeForestSource(t *testing.T, src string) *Program {
	t.Helper()
	tokens, err := lexer.NewLexer(src).ScanTokens()
	if err != nil {
		t.Fatal(err)
	}
	candidates, err := parser.NewParser(tokens).Parse()
	if err != nil {
		t.Fatal(err)
	}
	prog, err := Analyze(candidates)
	if err != nil {
		t.Fatalf("sema: %v", err)
	}
	return prog
}

func TestE2E_TypeDefDeclaration(t *testing.T) {
	analyzeForestSource(t, "volatile int (*const a(float))[2*3];")
}

func TestE2E_TypeName(t *testing.T) {
	analyzeForestSource(t, "int a[sizeof(int (*const [])(unsigned int, ...))];")
}

func TestE2E_FunctionDeclaration1(t *testing.T) {
	analyzeForestSource(t, "int (*fpfi(int (*)(long), int))(int, ...);")
}

func TestE2E_FunctionDeclaration2(t *testing.T) {
	analyzeForestSource(t, "int f(void), *fip(), (*pfi)();")
}

func TestE2E_FunctionDeclaration3(t *testing.T) {
	analyzeForestSource(t, "int (*apfi[3])(int *x, int *y);")
}

func TestE2E_SimpleFib(t *testing.T) {
	analyzeForestSource(t, `int Fibon1(int n){
    if (n == 1 || n == 2){
        return 1;
    } else{
        return Fibon1(n - 1) + Fibon1(n - 2);
    }
}
int scanf(const char *, ...);
int printf(const char *, ...);
int main(){
    int n = 0;
    int ret = 0;
    scanf("%d", &n);
    ret = Fibon1(n);
    printf("ret=%d", ret);
    return 0;
}`)
}

func TestE2E_Sqrt(t *testing.T) {
	analyzeForestSource(t, `float Q_rsqrt(float number)
{
  long i;
  float x2, y;
  const float threehalfs = 1.5F;

  x2 = number * 0.5F;
  y  = number;
  i  = * ( long * ) &y;
  i  = 0x5f3759df - ( i >> 1 );
  y  = * ( float * ) &i;
  y  = y * ( threehalfs - ( x2 * y * y ) );
  y  = y * ( threehalfs - ( x2 * y * y ) );

  return y;
}`)
}

func TestE2E_AmbiguousTypedefShadow(t *testing.T) {
	analyzeForestSource(t, `typedef int a;
int main() {
	int a;
	int b;
	a*b;
}`)
}

func TestE2E_SizeOf(t *testing.T) {
	analyzeForestSource(t, `int main() {
	int a;
	sizeof(a);
}`)
}

func TestE2E_TwoFunc(t *testing.T) {
	analyzeForestSource(t, `int main() {
}

int b;`)
}

func TestE2E_StructDefinition(t *testing.T) {
	analyzeForestSource(t, `struct Point {
int a;
int b;
};`)
}

func TestE2E_StructEnumDef(t *testing.T) {
	analyzeForestSource(t, `typedef struct Point {
    float x, y;
} Point;

typedef enum Color {
    RED,
    GREEN,
    BLUE,
    YELLOW = 10,
    WHITE,
    BLACK
} Color;`)
}

func TestE2E_ComplexTypeDef(t *testing.T) {
	analyzeForestSource(t, `typedef int (*(*a)(void))(void);`)
}
