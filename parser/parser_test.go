package parser

import (
	"shinya.click/cvm/lexer"
	"testing"
)

func TestTypeDefDeclaration(t *testing.T) {
	tokens, err := lexer.NewLexer("volatile int (*const a(float))[2*3];").ScanTokens()
	if err != nil {
		panic(err)
	}
	NewParser(tokens).Parse()
}

func TestTypeName(t *testing.T) {
	tokens, err := lexer.NewLexer("int a[sizeof(int (*const [])(unsigned int, ...))];").ScanTokens()
	if err != nil {
		panic(err)
	}
	NewParser(tokens).Parse()
}

func TestFunctionDeclaration1(t *testing.T) {
	tokens, err := lexer.NewLexer("int (*fpfi(int (*)(long), int))(int, ...);").ScanTokens()
	if err != nil {
		panic(err)
	}
	NewParser(tokens).Parse()
}

func TestFunctionDeclaration2(t *testing.T) {
	tokens, err := lexer.NewLexer("int f(void), *fip(), (*pfi)();").ScanTokens()
	if err != nil {
		panic(err)
	}
	NewParser(tokens).Parse()
}

func TestFunctionDeclaration3(t *testing.T) {
	tokens, err := lexer.NewLexer("int (*apfi[3])(int *x, int *y);").ScanTokens()
	if err != nil {
		panic(err)
	}
	NewParser(tokens).Parse()
}

func TestSimpleFib(t *testing.T) {
	tokens, err := lexer.NewLexer(`int Fibon1(int n){
    if (n == 1 || n == 2){
        return 1;
    } else{
        return Fibon1(n - 1) + Fibon1(n - 2);
    }
}
int main(){
    int n = 0;
    int ret = 0;
    scanf("%d", &n);
    ret = Fibon1(n);
    printf("ret=%d", ret);
    return 0;
}`).ScanTokens()
	if err != nil {
		panic(err)
	}
	p := NewParser(tokens)
	_, err = p.Parse()
	if err != nil {
		panic(err)
	}
}

func TestSqrt(t *testing.T) {
	tokens, err := lexer.NewLexer(`float Q_rsqrt(float number)
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
}`).ScanTokens()
	if err != nil {
		panic(err)
	}
	p := NewParser(tokens)
	_, err = p.Parse()
	if err != nil {
		panic(err)
	}
}

func TestAmbiguous(t *testing.T) {
	tokens, err := lexer.NewLexer(`int main() {
	a*b;
}`).ScanTokens()
	if err != nil {
		panic(err)
	}
	NewParser(tokens).Parse()
}

func TestSizeOf(t *testing.T) {
	tokens, err := lexer.NewLexer(`int main() {
	sizeof(a);
}`).ScanTokens()
	if err != nil {
		panic(err)
	}
	NewParser(tokens).Parse()
}

func TestTwoFunc(t *testing.T) {
	tokens, err := lexer.NewLexer(`int main() {
}

int b;`).ScanTokens()
	if err != nil {
		panic(err)
	}
	NewParser(tokens).Parse()
}

func TestTypeDef(t *testing.T) {
	tokens, err := lexer.NewLexer(`typedef int a;

typedef struct abc{
	int c;
} ccc;`).ScanTokens()
	if err != nil {
		panic(err)
	}
	NewParser(tokens).Parse()
}
