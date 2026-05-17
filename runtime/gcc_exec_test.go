package runtime

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"shinya.click/cvm/bytecode"
	"shinya.click/cvm/codegen"
	"shinya.click/cvm/entity"
	"shinya.click/cvm/parser"
	"shinya.click/cvm/preprocessor"
	"shinya.click/cvm/sema"
)

type gccExecCase struct {
	path     string
	exitCode int
	category string
	reason   string
}

func TestGCCExecutionManifestParses(t *testing.T) {
	content := "path\texit\tcategory\treason\n" +
		"sema/testdata/gcc-c99/accept/c99-main-1.c\t0\tarithmetic\treturns zero\n"
	cases := parseGCCExecManifest(t, content)
	if len(cases) != 1 {
		t.Fatalf("case count = %d, want 1", len(cases))
	}
	if got := cases[0]; got.path != "sema/testdata/gcc-c99/accept/c99-main-1.c" || got.exitCode != 0 || got.category != "arithmetic" || got.reason != "returns zero" {
		t.Fatalf("case = %#v", got)
	}
}

func TestGCCExecutionManifestRejectsEscapingPath(t *testing.T) {
	paths := []string{
		"sema/testdata/gcc-c99/accept/../reject/foo.c",
		"sema/testdata/gcc-c99/accept/..\\reject/foo.c",
	}
	for _, path := range paths {
		content := "path\texit\tcategory\treason\n" +
			path + "\t0\tarithmetic\tescapes accept root\n"
		_, err := parseGCCExecManifestContent(content)
		if err == nil {
			t.Fatalf("expected traversal path to be rejected: %s", path)
		}
	}
}

func TestGCCExecutionManifestRejectsInvalidRows(t *testing.T) {
	const validRow = "sema/testdata/gcc-c99/accept/c99-main-1.c\t0\tarithmetic\treturns zero\n"
	tests := []struct {
		name    string
		content string
		wantErr string
	}{
		{
			name:    "bad header",
			content: "path\texit\treason\n" + validRow,
			wantErr: "manifest header is missing or malformed",
		},
		{
			name:    "malformed row",
			content: "path\texit\tcategory\treason\nsema/testdata/gcc-c99/accept/c99-main-1.c\t0\tarithmetic\n",
			wantErr: "manifest line 2 malformed",
		},
		{
			name:    "duplicate path",
			content: "path\texit\tcategory\treason\n" + validRow + validRow,
			wantErr: "duplicates path",
		},
		{
			name:    "absolute path",
			content: "path\texit\tcategory\treason\n/sema/testdata/gcc-c99/accept/c99-main-1.c\t0\tarithmetic\tabsolute path\n",
			wantErr: "outside supported GCC accept fixture roots",
		},
		{
			name:    "unsupported root",
			content: "path\texit\tcategory\treason\nsema/testdata/gcc-c99/reject/not-accept.c\t0\tarithmetic\twrong root\n",
			wantErr: "outside supported GCC accept fixture roots",
		},
		{
			name:    "non C source",
			content: "path\texit\tcategory\treason\nsema/testdata/gcc-c99/accept/not-c.txt\t0\tarithmetic\tnot source\n",
			wantErr: "outside supported GCC accept fixture roots",
		},
		{
			name:    "invalid exit",
			content: "path\texit\tcategory\treason\nsema/testdata/gcc-c99/accept/c99-main-1.c\tzero\tarithmetic\tbad exit\n",
			wantErr: "invalid exit code",
		},
		{
			name:    "empty category",
			content: "path\texit\tcategory\treason\nsema/testdata/gcc-c99/accept/c99-main-1.c\t0\t\tmissing category\n",
			wantErr: "empty category",
		},
		{
			name:    "empty reason",
			content: "path\texit\tcategory\treason\nsema/testdata/gcc-c99/accept/c99-main-1.c\t0\tarithmetic\t \n",
			wantErr: "empty reason",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseGCCExecManifestContent(tt.content)
			if err == nil {
				t.Fatalf("parseGCCExecManifestContent succeeded, want error containing %q", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %q, want substring %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestHasGCCRunDirectiveAcceptsBlockCommentRun(t *testing.T) {
	source := "/* { dg-do run } */\nint main(void) { return 0; }\n"
	if !hasGCCRunDirective(source) {
		t.Fatalf("hasGCCRunDirective(%q) = false, want true", source)
	}
}

func TestHasGCCRunDirectiveAcceptsTargetQualifiedRun(t *testing.T) {
	source := "/* { dg-do run { target *-*-solaris2* } } */\nint main(void) { return 0; }\n"
	if !hasGCCRunDirective(source) {
		t.Fatalf("hasGCCRunDirective(%q) = false, want true", source)
	}
}

func TestHasGCCRunDirectiveAcceptsC99RuntimeRequirement(t *testing.T) {
	source := "/* { dg-require-effective-target c99_runtime } */\nint main(void) { return 0; }\n"
	if !hasGCCRunDirective(source) {
		t.Fatalf("hasGCCRunDirective(%q) = false, want true", source)
	}
}

func TestHasGCCRunDirectiveRejectsNonRunCases(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name:   "compile only",
			source: "/* { dg-do compile } */\nint main(void) { return 0; }\n",
		},
		{
			name:   "no directive",
			source: "int main(void) { return 0; }\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if hasGCCRunDirective(tt.source) {
				t.Fatalf("hasGCCRunDirective(%q) = true, want false", tt.source)
			}
		})
	}
}

func TestGCCExecutionFixtures(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("testdata", "gcc-exec", "manifest.tsv"))
	if err != nil {
		t.Fatalf("read GCC execution manifest: %v", err)
	}
	cases := parseGCCExecManifest(t, string(content))
	const minGCCExecCases = 6
	if len(cases) < minGCCExecCases {
		t.Fatalf("GCC execution suite too small: got %d cases, want >= %d", len(cases), minGCCExecCases)
	}
	for _, c := range cases {
		c := c
		t.Run(filepath.Base(c.path), func(t *testing.T) {
			sourcePath := filepath.Join("..", c.path)
			source, err := os.ReadFile(sourcePath)
			if err != nil {
				t.Fatalf("read fixture %s: %v", c.path, err)
			}
			if !hasGCCRunDirective(string(source)) {
				t.Fatalf("%s is not a GCC runtime fixture: missing { dg-do run } directive", c.path)
			}
			st := runGCCExecFixture(t, sourcePath, string(source))
			if st.Code != c.exitCode {
				t.Fatalf("%s exit code = %d, want %d", c.path, st.Code, c.exitCode)
			}
		})
	}
}

func TestTgmathSinFloatExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <tgmath.h>

float foo(float x)
{
  return sin(x);
}

int main(void)
{
  return foo(0.0f) == 0.0f ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "tgmath-sin-float.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestTgmathExpPowExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <tgmath.h>

int main(void)
{
  if (exp(0.0) != 1.0)
    return 1;
  if (pow(2.0f, 3.0f) != 8.0f)
    return 2;
  return 0;
}
`
	st := runGCCExecFixture(t, "tgmath-exp-pow.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestTgmathSqrtExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <tgmath.h>

int main(void)
{
  if (sqrt(9.0f) != 3.0f)
    return 1;
  return sqrt(16.0L) == 4.0L ? 0 : 2;
}
`
	st := runGCCExecFixture(t, "tgmath-sqrt-real.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestTgmathCosExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <tgmath.h>

int main(void)
{
  if (cos(0.0f) != 1.0f)
    return 1;
  return cos(0.0L) == 1.0L ? 0 : 2;
}
`
	st := runGCCExecFixture(t, "tgmath-cos-real.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestTgmathTanExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <tgmath.h>

int main(void)
{
  if (tan(0.0f) != 0.0f)
    return 1;
  return tan(0.0L) == 0.0L ? 0 : 2;
}
`
	st := runGCCExecFixture(t, "tgmath-tan-real.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestTgmathLogExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <tgmath.h>

int main(void)
{
  if (log(1.0f) != 0.0f)
    return 1;
  return log(1.0L) == 0.0L ? 0 : 2;
}
`
	st := runGCCExecFixture(t, "tgmath-log-real.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestTgmathSinhExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <tgmath.h>

int main(void)
{
  if (sinh(0.0f) != 0.0f)
    return 1;
  return sinh(0.0L) == 0.0L ? 0 : 2;
}
`
	st := runGCCExecFixture(t, "tgmath-sinh-real.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestTgmathCoshExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <tgmath.h>

int main(void)
{
  if (cosh(0.0f) != 1.0f)
    return 1;
  return cosh(0.0L) == 1.0L ? 0 : 2;
}
`
	st := runGCCExecFixture(t, "tgmath-cosh-real.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestTgmathTanhExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <tgmath.h>

int main(void)
{
  if (tanh(0.0f) != 0.0f)
    return 1;
  return tanh(0.0L) == 0.0L ? 0 : 2;
}
`
	st := runGCCExecFixture(t, "tgmath-tanh-real.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestTgmathComplexSinExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <tgmath.h>

int main(void)
{
  complex double z = __builtin_complex(0.0, 0.0);
  complex double r = sin(z);
  return __builtin_cabs(r) == 0.0 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "tgmath-complex-sin.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestTgmathComplexTanhExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <tgmath.h>

int main(void)
{
  complex double z = __builtin_complex(0.0, 0.0);
  complex double r = tanh(z);
  return __builtin_cabs(r) == 0.0 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "tgmath-complex-tanh.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestTgmathComplexCoshExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <tgmath.h>

int main(void)
{
  complex double z = __builtin_complex(0.0, 0.0);
  complex double r = cosh(z);
  return __builtin_cabs(r) == 1.0 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "tgmath-complex-cosh.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestTgmathComplexSinhExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <tgmath.h>

int main(void)
{
  complex double z = __builtin_complex(0.0, 0.0);
  complex double r = sinh(z);
  return __builtin_cabs(r) == 0.0 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "tgmath-complex-sinh.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestTgmathComplexLogExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <tgmath.h>

int main(void)
{
  complex double z = __builtin_complex(1.0, 0.0);
  complex double r = log(z);
  return __builtin_cabs(r) == 0.0 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "tgmath-complex-log.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestTgmathComplexTanExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <tgmath.h>

int main(void)
{
  complex double z = __builtin_complex(0.0, 0.0);
  complex double r = tan(z);
  return __builtin_cabs(r) == 0.0 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "tgmath-complex-tan.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestTgmathComplexCosExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <tgmath.h>

int main(void)
{
  complex double z = __builtin_complex(0.0, 0.0);
  complex double r = cos(z);
  return __builtin_cabs(r) == 1.0 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "tgmath-complex-cos.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestTgmathComplexSqrtExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <tgmath.h>

int main(void)
{
  complex double z = __builtin_complex(3.0, 4.0);
  complex double r = sqrt(z);
  return __builtin_cabs(r) == __builtin_cabs(__builtin_complex(2.0, 1.0)) ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "tgmath-complex-sqrt.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestTgmathComplexFloatSinExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <tgmath.h>

int main(void)
{
  complex float z = __builtin_complex(0.0f, 0.0f);
  complex float r = sin(z);
  return __builtin_cabsf(r) == 0.0f ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "tgmath-complex-float-sin.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestTgmathComplexExpExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <tgmath.h>

int main(void)
{
  complex double z = __builtin_complex(0.0, 0.0);
  complex double r = exp(z);
  return __builtin_cabs(r) == 1.0 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "tgmath-complex-exp.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestTgmathComplexFloatExpExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <tgmath.h>

int main(void)
{
  complex float z = __builtin_complex(0.0f, 0.0f);
  complex float r = exp(z);
  return __builtin_cabsf(r) == 1.0f ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "tgmath-complex-float-exp.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestTgmathComplexPowExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <tgmath.h>

int main(void)
{
  complex double z = __builtin_complex(2.0, 0.0);
  complex double r = pow(z, 3.0);
  return __builtin_cabs(r) == 8.0 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "tgmath-complex-pow.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestTgmathComplexFloatPowExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <tgmath.h>

complex double foo(complex float x, float y)
{
  return pow(x, y);
}

int main(void)
{
  complex float z = __builtin_complex(2.0f, 0.0f);
  complex double r = foo(z, 3.0f);
  return __builtin_cabs(r) == 8.0 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "tgmath-complex-float-pow.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCVLAStructAndUnionMembersExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

static int A(int i)
{
  struct S { int ar[1][i]; } s;

  s.ar[0][0] = 7;
  s.ar[0][i - 1] = 9;
  return s.ar[0][0] + s.ar[0][i - 1];
}

static int C(int i)
{
  union U { int ar[1][i]; } u;

  u.ar[0][0] = 11;
  u.ar[0][i - 1] = 13;
  return u.ar[0][0] + u.ar[0][i - 1];
}

int main(void)
{
  return A(23) == 16 && C(23) == 24 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "vla-struct-union-members-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCVLAParameterDynamicStrideExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

void ed(int n, float s[3][n])
{
  for (int i = 0; i < n; i++)
    s[1][i] = i + 1;
}

int main(void)
{
  int n = 4;
  float s[3][n];
  ed(n, s);
  return s[1][3] == 4.0f ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "vla-parameter-dynamic-stride-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCIntegerConversionWarningsExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <limits.h>

int fuc(unsigned char uc)
{
  return uc;
}

unsigned fui(unsigned int ui)
{
  return ui + -1;
}

int main(void)
{
  int x = 0;
  unsigned int ui = 3;
  unsigned char uc = 3;

  uc = -1;
  if (uc != UCHAR_MAX)
    return 1;

  ui = -1;
  if (ui != UINT_MAX)
    return 2;

  uc = x ? 1U : -1;
  if (uc != UCHAR_MAX)
    return 3;

  ui = x ? 1U : -1;
  if (ui != UINT_MAX)
    return 4;

  if (fuc(-1) != UCHAR_MAX)
    return 5;

  if (fui(3) != 2U)
    return 6;

  ui = 1U * -1;
  if (ui != UINT_MAX)
    return 7;

  return 0;
}
`
	st := runGCCExecFixture(t, "Wconversion-integer-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCSignedUnsignedConditionalConversionsExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <limits.h>

int fsi(int si)
{
  return si;
}

unsigned fui(unsigned int ui)
{
  return ui;
}

int main(void)
{
  int x = 1;
  unsigned int ui = 0;
  unsigned char uc = 0;
  signed char sc = 0;

  uc = x ? SCHAR_MIN : 1U;
  if (uc != 128)
    return 1;

  ui = x ? INT_MIN : 1U;
  if (ui != (unsigned) INT_MIN)
    return 2;

  sc = (unsigned char) -1;
  if (sc != -1)
    return 3;

  uc = sc;
  if (uc != UCHAR_MAX)
    return 4;

  if (fsi(UINT_MAX / 3U) != 1431655765)
    return 5;

  if (fui(-1) != UINT_MAX)
    return 6;

  return 0;
}
`
	st := runGCCExecFixture(t, "Wsign-conversion-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCLongDoubleArithmeticExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

int main(void)
{
  volatile long double x = 1.25L;
  long double y = x + 2.75L;
  y *= 2.0L;
  y /= 4.0L;
  return y == 2.0L ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "long-double-arithmetic-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCLongDoubleArgumentReturnExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

long double add(long double x, long double y)
{
  return x + y;
}

int main(void)
{
  long double z = add(1.25L, 2.75L);
  return z == 4.0L ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "long-double-argument-return-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCFloatCompoundDoubleRHSExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

int main(void)
{
  float f = 1.25f;
  if ((f += 2.25) != 3.5f)
    return 1;
  return f == 3.5f ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "float-compound-double-rhs-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCLongDoubleFieldCompoundExpressionExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

struct box {
  long double x;
};

int main(void)
{
  struct box b = { 1.25L };
  if ((b.x += 2.75L) != 4.0L)
    return 1;
  return b.x == 4.0L ? 0 : 2;
}
`
	st := runGCCExecFixture(t, "long-double-field-compound-expression-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCFloatingAssignmentAndLogicalExpressionsExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

struct box {
  long double x;
};

int main(void)
{
  float f = 0.0f;
  struct box b = { 0.0L };

  if ((f = 2.0f) != 2.0f)
    return 1;
  if ((b.x = -3.0L) != -3.0L)
    return 2;
  if (!(f && b.x))
    return 3;
  if (0.0L || 0.0f)
    return 4;
  return 0;
}
`
	st := runGCCExecFixture(t, "floating-assignment-logical-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCFloatIncDecExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

int main(void)
{
  float f = 1.5f;
  if (f++ != 1.5f)
    return 1;
  if (f != 2.5f)
    return 2;
  if (++f != 3.5f)
    return 3;
  return 0;
}
`
	st := runGCCExecFixture(t, "float-incdec-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCLongDoubleFieldIncDecExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

struct box {
  long double x;
};

int main(void)
{
  struct box b = { 4.0L };
  if (b.x-- != 4.0L)
    return 1;
  if (b.x != 3.0L)
    return 2;
  if (--b.x != 2.0L)
    return 3;
  return 0;
}
`
	st := runGCCExecFixture(t, "long-double-field-incdec-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCBitFieldCompoundAndIncDecExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

struct bits {
  unsigned int a : 3;
};

int main(void)
{
  struct bits s = { 1 };
  if ((s.a += 3) != 4)
    return 1;
  if (s.a++ != 4)
    return 2;
  if (s.a != 5)
    return 3;
  if (--s.a != 4)
    return 4;
  return 0;
}
`
	st := runGCCExecFixture(t, "bit-field-compound-incdec-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCBoolBitFieldCompoundAndIncDecExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

struct bits {
  _Bool b : 1;
};

int main(void)
{
  struct bits s = { 0 };
  if ((s.b += 2) != 1)
    return 1;
  if (s.b-- != 1)
    return 2;
  if (s.b != 0)
    return 3;
  if (--s.b != 1)
    return 4;
  return 0;
}
`
	st := runGCCExecFixture(t, "bool-bit-field-compound-incdec-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCBitFieldExpressionValuesAreStoredValuesExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

struct bits {
  unsigned int a : 3;
};

int main(void)
{
  struct bits s = { 0 };
  if ((s.a = 9) != 1)
    return 1;
  if ((s.a += 8) != 1)
    return 2;
  s.a = 7;
  if (++s.a != 0)
    return 3;
  if (s.a != 0)
    return 4;
  return 0;
}
`
	st := runGCCExecFixture(t, "bit-field-expression-stored-values-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCPointerFieldCompoundAssignmentExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

struct box {
  int *p;
};

int main(void)
{
  int values[3] = { 10, 20, 30 };
  struct box b;
  b.p = values;

  if (*(b.p += 1) != 20)
    return 1;
  if (*b.p != 20)
    return 2;
  if (*(b.p -= 1) != 10)
    return 3;
  return b.p == values ? 0 : 4;
}
`
	st := runGCCExecFixture(t, "pointer-field-compound-assignment-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCPointerFieldInitializerExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

struct box {
  int *p;
};

int main(void)
{
  int values[2] = { 10, 20 };
  struct box b = { values };
  if (b.p[1] != 20)
    return 1;
  return *(b.p += 1) == 20 ? 0 : 2;
}
`
	st := runGCCExecFixture(t, "pointer-field-initializer-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCPointerArrayElementCompoundAssignmentExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

int main(void)
{
  int values[2] = { 10, 20 };
  int *ptrs[1] = { values };

  if (*(ptrs[0] += 1) != 20)
    return 1;
  if (ptrs[0] != values + 1)
    return 2;
  return 0;
}
`
	st := runGCCExecFixture(t, "pointer-array-element-compound-assignment-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCAddressablePointerIncDecExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

struct box {
  int *p;
};

int main(void)
{
  int values[3] = { 10, 20, 30 };
  struct box b;
  int *ptrs[1] = { values + 1 };

  b.p = values;
  if (*(b.p++) != 10)
    return 1;
  if (*b.p != 20)
    return 2;
  if (*(--b.p) != 10)
    return 3;
  if (*(ptrs[0]--) != 20)
    return 4;
  return ptrs[0] == values ? 0 : 5;
}
`
	st := runGCCExecFixture(t, "addressable-pointer-incdec-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCStaticPointerFieldInitializerExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

struct box {
  int *p;
};

static int values[2] = { 10, 20 };
static struct box b = { values };

int main(void)
{
  if (b.p[1] != 20)
    return 1;
  b.p += 1;
  return *b.p == 20 ? 0 : 2;
}
`
	st := runGCCExecFixture(t, "static-pointer-field-initializer-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCStaticPointerArrayInitializerExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

static int values[3] = { 10, 20, 30 };
static int *ptrs[2] = { values, values + 1 };

int main(void)
{
  if (*ptrs[0] != 10 || *ptrs[1] != 20)
    return 1;
  ptrs[0] += 2;
  return *ptrs[0] == 30 ? 0 : 2;
}
`
	st := runGCCExecFixture(t, "static-pointer-array-initializer-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCFunctionPointerFieldCallExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

struct ops {
  int (*fn)(int);
};

int inc(int x)
{
  return x + 1;
}

int main(void)
{
  struct ops o = { inc };
  return o.fn(41) == 42 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "function-pointer-field-call-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCStaticFunctionPointerFieldCallExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

struct ops {
  int (*fn)(int);
};

int inc(int x)
{
  return x + 1;
}

static struct ops o = { inc };

int main(void)
{
  return o.fn(41) == 42 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "static-function-pointer-field-call-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCFunctionPointerArrayCallExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

int inc(int x)
{
  return x + 1;
}

int dec(int x)
{
  return x - 1;
}

int main(void)
{
  int (*ops[2])(int) = { inc, dec };
  return ops[0](41) == 42 && ops[1](43) == 42 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "function-pointer-array-call-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCFunctionPointerReturnCallExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

int inc(int x)
{
  return x + 1;
}

int dec(int x)
{
  return x - 1;
}

int (*choose(int flag))(int)
{
  return flag ? &inc : &dec;
}

int main(void)
{
  return choose(1)(41) == 42 && choose(0)(43) == 42 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "function-pointer-return-call-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCFunctionDesignatorReturnConvertsToPointerExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

int inc(int x)
{
  return x + 1;
}

int dec(int x)
{
  return x - 1;
}

int (*choose(int flag))(int)
{
  return flag ? inc : dec;
}

int main(void)
{
  return choose(1)(41) == 42 && choose(0)(43) == 42 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "function-designator-return-pointer-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCFunctionDesignatorConditionalInitializerExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

int inc(int x)
{
  return x + 1;
}

int dec(int x)
{
  return x - 1;
}

int main(void)
{
  int flag = 0;
  int (*fn)(int) = flag ? inc : dec;
  return fn(43) == 42 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "function-designator-conditional-initializer-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCFunctionDesignatorAssignmentAndCommaExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

int inc(int x)
{
  return x + 1;
}

int dec(int x)
{
  return x - 1;
}

int main(void)
{
  int (*fn)(int);
  fn = inc;
  if (fn(41) != 42)
    return 1;
  fn = (inc, dec);
  return fn(43) == 42 ? 0 : 2;
}
`
	st := runGCCExecFixture(t, "function-designator-assignment-comma-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCVariadicDirectCallExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

int first(int n, ...)
{
  return n;
}

int main(void)
{
  return first(42, 7, 8) == 42 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "variadic-direct-call-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCVariadicFunctionPointerCallExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

int first(int n, ...)
{
  return n;
}

int main(void)
{
  int (*fn)(int, ...) = first;
  return fn(42, 7, 8) == 42 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "variadic-function-pointer-call-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCFunctionPointerStructReturnExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

struct pair {
  int tag;
  __complex__ double value;
};

struct pair make(void)
{
  struct pair p = { 7, __builtin_complex(3.0, 4.0) };
  return p;
}

int main(void)
{
  struct pair (*fn)(void) = make;
  struct pair p = fn();
  return p.tag == 7 && __builtin_cabs(p.value) == 5.0 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "function-pointer-struct-return-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCFunctionPointerStructArgumentExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

struct pair {
  int tag;
  __complex__ double value;
};

double norm_pair(struct pair p)
{
  p.tag = 99;
  return __builtin_cabs(p.value);
}

int main(void)
{
  double (*fn)(struct pair) = norm_pair;
  struct pair p = { 7, __builtin_complex(3.0, 4.0) };
  if (fn(p) != 5.0)
    return 1;
  return p.tag == 7 ? 0 : 2;
}
`
	st := runGCCExecFixture(t, "function-pointer-struct-argument-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCFunctionPointerArgumentCallExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

int inc(int x)
{
  return x + 1;
}

int apply(int (*fn)(int), int x)
{
  return fn(x);
}

int main(void)
{
  return apply(inc, 41) == 42 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "function-pointer-argument-call-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCComplexReciprocalImaginaryExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

__complex__ double foo(__complex__ double x)
{
  return 1.0 / x * -1.0i;
}

int main(void)
{
  __complex__ double z = __builtin_complex(0.0, 2.0);
  __complex__ double r = foo(z);
  return __builtin_cabs(r + __builtin_complex(0.5, 0.0)) == 0.0 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "complex-4-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCIntegerImaginaryExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

int main(void)
{
  __complex__ double z = 1i;
  return __builtin_cabs(z - __builtin_complex(0.0, 1.0)) == 0.0 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "integer-imaginary-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCImaginaryToRealLocalInitializerExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

int main(void)
{
  double d = 3.0i;
  return d == 0.0 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "imaginary-to-real-local-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCImaginaryToBoolLocalInitializerExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

int main(void)
{
  _Bool b = 3.0i;
  return b ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "imaginary-to-bool-local-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCComplexEqualityExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

int main(void)
{
  __complex__ double z = __builtin_complex(1.0, 2.0);
  __complex__ double w = __builtin_complex(1.0, 2.0);
  if (z != w)
    return 1;
  if (z == __builtin_complex(1.0, 3.0))
    return 2;
  return 0;
}
`
	st := runGCCExecFixture(t, "complex-equality-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCScalarExplicitCastToComplexExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

int main(void)
{
  __complex__ double z = (__complex__ double)3.0;
  return __builtin_cabs(z - __builtin_complex(3.0, 0.0)) == 0.0 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "scalar-explicit-cast-to-complex-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCBuiltinComplexDirectArgumentExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

int main(void)
{
  return __builtin_cabs(__builtin_complex(3.0, 4.0)) == 5.0 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "builtin-complex-direct-argument-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCComplexArgumentExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

double norm(__complex__ double z)
{
  return __builtin_cabs(z);
}

int main(void)
{
  return norm(__builtin_complex(3.0, 4.0)) == 5.0 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "complex-argument-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCComplexArgumentReturnExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

__complex__ double ident(__complex__ double z)
{
  return z;
}

int main(void)
{
  return __builtin_cabs(ident(__builtin_complex(3.0, 4.0))) == 5.0 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "complex-argument-return-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCComplexArgumentIsPassedByValueExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

void mutate(__complex__ double z)
{
  z += 2.0;
}

int main(void)
{
  __complex__ double z = __builtin_complex(3.0, 4.0);
  mutate(z);
  return __builtin_cabs(z) == 5.0 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "complex-argument-by-value-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCComplexFieldArgumentIsPassedByValueExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

struct pair {
  int tag;
  __complex__ double value;
};

void mutate(__complex__ double z)
{
  z += 2.0;
}

int main(void)
{
  struct pair p = { 7, __builtin_complex(3.0, 4.0) };
  mutate(p.value);
  return p.tag == 7 && __builtin_cabs(p.value) == 5.0 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "complex-field-argument-by-value-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCComplexStructArgumentIsPassedByValueExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

struct pair {
  int tag;
  __complex__ double value;
};

void mutate(struct pair p)
{
  p.tag = 9;
  p.value = __builtin_complex(5.0, 12.0);
}

int main(void)
{
  struct pair p = { 7, __builtin_complex(3.0, 4.0) };
  mutate(p);
  return p.tag == 7 && __builtin_cabs(p.value) == 5.0 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "complex-struct-argument-by-value-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCComplexStructReturnExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

struct pair {
  int tag;
  __complex__ double value;
};

struct pair make(void)
{
  struct pair p = { 7, __builtin_complex(5.0, 12.0) };
  return p;
}

int main(void)
{
  struct pair p = make();
  return p.tag == 7 && __builtin_cabs(p.value) == 13.0 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "complex-struct-return-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCComplexStructReturnAssignmentExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

struct pair {
  int tag;
  __complex__ double value;
};

struct pair make(void)
{
  struct pair p = { 7, __builtin_complex(5.0, 12.0) };
  return p;
}

int main(void)
{
  struct pair p = { 1, __builtin_complex(0.0, 0.0) };
  p = make();
  return p.tag == 7 && __builtin_cabs(p.value) == 13.0 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "complex-struct-return-assignment-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCComplexStructConditionalReturnInitializerExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

struct pair {
  int tag;
  __complex__ double value;
};

struct pair make_a(void)
{
  struct pair p = { 7, __builtin_complex(3.0, 4.0) };
  return p;
}

struct pair make_b(void)
{
  struct pair p = { 9, __builtin_complex(5.0, 12.0) };
  return p;
}

int main(void)
{
  int flag = 0;
  struct pair p = flag ? make_a() : make_b();
  return p.tag == 9 && __builtin_cabs(p.value) == 13.0 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "complex-struct-conditional-return-initializer-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCComplexStructConditionalReturnAssignmentExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

struct pair {
  int tag;
  __complex__ double value;
};

struct pair make_a(void)
{
  struct pair p = { 7, __builtin_complex(3.0, 4.0) };
  return p;
}

struct pair make_b(void)
{
  struct pair p = { 9, __builtin_complex(5.0, 12.0) };
  return p;
}

int main(void)
{
  int flag = 1;
  struct pair p = { 0, __builtin_complex(0.0, 0.0) };
  p = flag ? make_a() : make_b();
  return p.tag == 7 && __builtin_cabs(p.value) == 5.0 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "complex-struct-conditional-return-assignment-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCComplexStructCommaReturnAssignmentExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

struct pair {
  int tag;
  __complex__ double value;
};

struct pair make_a(void)
{
  struct pair p = { 7, __builtin_complex(3.0, 4.0) };
  return p;
}

struct pair make_b(void)
{
  struct pair p = { 9, __builtin_complex(5.0, 12.0) };
  return p;
}

int main(void)
{
  struct pair p = { 0, __builtin_complex(0.0, 0.0) };
  p = (make_a(), make_b());
  return p.tag == 9 && __builtin_cabs(p.value) == 13.0 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "complex-struct-comma-return-assignment-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCComplexStructCommaReturnInitializerExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

struct pair {
  int tag;
  __complex__ double value;
};

struct pair make_a(void)
{
  struct pair p = { 7, __builtin_complex(3.0, 4.0) };
  return p;
}

struct pair make_b(void)
{
  struct pair p = { 9, __builtin_complex(5.0, 12.0) };
  return p;
}

int main(void)
{
  struct pair p = (make_a(), make_b());
  return p.tag == 9 && __builtin_cabs(p.value) == 13.0 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "complex-struct-comma-return-initializer-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCComplexStructConditionalReturnArgumentExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

struct pair {
  int tag;
  __complex__ double value;
};

struct pair make_a(void)
{
  struct pair p = { 7, __builtin_complex(3.0, 4.0) };
  return p;
}

struct pair make_b(void)
{
  struct pair p = { 9, __builtin_complex(5.0, 12.0) };
  return p;
}

double norm_pair(struct pair p)
{
  return p.tag + __builtin_cabs(p.value);
}

int main(void)
{
  int flag = 0;
  return norm_pair(flag ? make_a() : make_b()) == 22.0 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "complex-struct-conditional-return-argument-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCNestedComplexStructArgumentIsPassedByValueExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

struct inner {
  int tag;
  __complex__ double value;
};

struct outer {
  int prefix;
  struct inner in;
};

void mutate(struct outer o)
{
  o.prefix = 2;
  o.in.tag = 9;
  o.in.value = __builtin_complex(5.0, 12.0);
}

int main(void)
{
  struct outer o = { 1, { 7, __builtin_complex(3.0, 4.0) } };
  mutate(o);
  return o.prefix == 1 && o.in.tag == 7 && __builtin_cabs(o.in.value) == 5.0 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "nested-complex-struct-argument-by-value-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCComplexStructAssignmentExpressionMemberExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

struct pair {
  int tag;
  __complex__ double value;
};

struct pair make(void)
{
  struct pair p = { 9, __builtin_complex(5.0, 12.0) };
  return p;
}

int main(void)
{
  struct pair p = { 1, __builtin_complex(3.0, 4.0) };
  if ((p = make()).tag != 9)
    return 1;
  return __builtin_cabs((p = make()).value) == 13.0 ? 0 : 2;
}
`
	st := runGCCExecFixture(t, "complex-struct-assignment-expression-member-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCComplexStructAssignmentExpressionArgumentExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

struct pair {
  int tag;
  __complex__ double value;
};

struct pair make(void)
{
  struct pair p = { 9, __builtin_complex(5.0, 12.0) };
  return p;
}

double norm_pair(struct pair p)
{
  return p.tag + __builtin_cabs(p.value);
}

int main(void)
{
  struct pair p = { 1, __builtin_complex(3.0, 4.0) };
  if (norm_pair(p = make()) != 22.0)
    return 1;
  return p.tag == 9 && __builtin_cabs(p.value) == 13.0 ? 0 : 2;
}
`
	st := runGCCExecFixture(t, "complex-struct-assignment-expression-argument-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCComplexArrayInitializerExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

int main(void)
{
  __complex__ double values[2] = {
    __builtin_complex(3.0, 4.0),
    __builtin_complex(5.0, 12.0)
  };
  return __builtin_cabs(values[1]) == 13.0 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "complex-array-initializer-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCStaticComplexArrayInitializerExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

const __complex__ double values[2] = {
  __builtin_complex(3.0, 4.0),
  __builtin_complex(5.0, 12.0)
};

int main(void)
{
  return __builtin_cabs(values[1]) == 13.0 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "static-complex-array-initializer-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCComplexStructFieldInitializerExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

struct pair {
  int tag;
  __complex__ double value;
};

int main(void)
{
  struct pair p = { 7, __builtin_complex(5.0, 12.0) };
  return p.tag == 7 && __builtin_cabs(p.value) == 13.0 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "complex-struct-field-initializer-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCStaticComplexStructFieldInitializerExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

struct pair {
  int tag;
  __complex__ double value;
};

const struct pair p = { 7, __builtin_complex(5.0, 12.0) };

int main(void)
{
  return p.tag == 7 && __builtin_cabs(p.value) == 13.0 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "static-complex-struct-field-initializer-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCComplexStructDesignatedInitializerExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

struct pair {
  int tag;
  __complex__ double value;
};

int main(void)
{
  struct pair p = { .value = __builtin_complex(5.0, 12.0), .tag = 7 };
  return p.tag == 7 && __builtin_cabs(p.value) == 13.0 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "complex-struct-designated-initializer-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCStaticComplexStructDesignatedInitializerExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

struct pair {
  int tag;
  __complex__ double value;
};

const struct pair p = { .value = __builtin_complex(5.0, 12.0), .tag = 7 };

int main(void)
{
  return p.tag == 7 && __builtin_cabs(p.value) == 13.0 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "static-complex-struct-designated-initializer-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCNestedComplexStructInitializerExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

struct inner {
  int tag;
  __complex__ double value;
};

struct outer {
  int prefix;
  struct inner in;
};

int main(void)
{
  struct outer o = { 1, { 7, __builtin_complex(5.0, 12.0) } };
  return o.prefix == 1 && o.in.tag == 7 && __builtin_cabs(o.in.value) == 13.0 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "nested-complex-struct-initializer-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCNestedComplexStructDesignatedInitializerExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

struct inner {
  int tag;
  __complex__ double value;
};

struct outer {
  int prefix;
  struct inner in;
};

int main(void)
{
  struct outer o = { .in.value = __builtin_complex(5.0, 12.0), .in.tag = 7, .prefix = 1 };
  return o.prefix == 1 && o.in.tag == 7 && __builtin_cabs(o.in.value) == 13.0 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "nested-complex-struct-designated-initializer-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCComplexStructFieldAssignmentExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

struct pair {
  int tag;
  __complex__ double value;
};

int main(void)
{
  struct pair p = { 7, __builtin_complex(0.0, 0.0) };
  p.value = __builtin_complex(5.0, 12.0);
  return p.tag == 7 && __builtin_cabs(p.value) == 13.0 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "complex-struct-field-assignment-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCComplexStructFieldCompoundAssignmentExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

struct pair {
  int tag;
  __complex__ double value;
};

int main(void)
{
  struct pair p = { 7, __builtin_complex(1.0, 4.0) };
  p.value += 2.0;
  return p.tag == 7 && __builtin_cabs(p.value) == 5.0 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "complex-struct-field-compound-assignment-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCComplexStructArrowFieldAssignmentExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

struct pair {
  int tag;
  __complex__ double value;
};

int main(void)
{
  struct pair p = { 7, __builtin_complex(0.0, 0.0) };
  struct pair *q = &p;
  q->value = __builtin_complex(5.0, 12.0);
  return q->tag == 7 && __builtin_cabs(q->value) == 13.0 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "complex-struct-arrow-field-assignment-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCComplexStructArrowFieldCompoundAssignmentExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

struct pair {
  int tag;
  __complex__ double value;
};

int main(void)
{
  struct pair p = { 7, __builtin_complex(1.0, 4.0) };
  struct pair *q = &p;
  q->value += 2.0;
  return q->tag == 7 && __builtin_cabs(q->value) == 5.0 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "complex-struct-arrow-field-compound-assignment-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCComplexPointerAssignmentToStructFieldExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

struct pair {
  int tag;
  __complex__ double value;
};

int main(void)
{
  struct pair p = { 7, __builtin_complex(0.0, 0.0) };
  __complex__ double *q = &p.value;
  *q = __builtin_complex(5.0, 12.0);
  return p.tag == 7 && __builtin_cabs(p.value) == 13.0 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "complex-pointer-assignment-to-struct-field-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCComplexPointerCompoundAssignmentToStructFieldExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

struct pair {
  int tag;
  __complex__ double value;
};

int main(void)
{
  struct pair p = { 7, __builtin_complex(1.0, 4.0) };
  __complex__ double *q = &p.value;
  *q += 2.0;
  return p.tag == 7 && __builtin_cabs(p.value) == 5.0 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "complex-pointer-compound-assignment-to-struct-field-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCComplexPointerDereferenceReturnExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

struct pair {
  int tag;
  __complex__ double value;
};

__complex__ double load(__complex__ double *q)
{
  return *q;
}

int main(void)
{
  struct pair p = { 7, __builtin_complex(5.0, 12.0) };
  return __builtin_cabs(load(&p.value)) == 13.0 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "complex-pointer-dereference-return-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCComplexPointerDereferenceInitializerExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

struct pair {
  int tag;
  __complex__ double value;
};

int main(void)
{
  struct pair p = { 7, __builtin_complex(5.0, 12.0) };
  __complex__ double *q = &p.value;
  __complex__ double z = *q;
  return __builtin_cabs(z) == 13.0 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "complex-pointer-dereference-initializer-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCComplexPointerDereferenceAbsExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

struct pair {
  int tag;
  __complex__ double value;
};

int main(void)
{
  struct pair p = { 7, __builtin_complex(5.0, 12.0) };
  __complex__ double *q = &p.value;
  return __builtin_cabs(*q) == 13.0 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "complex-pointer-dereference-abs-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCComplexPointerDereferenceArgumentExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

struct pair {
  int tag;
  __complex__ double value;
};

double norm(__complex__ double z)
{
  return __builtin_cabs(z);
}

int main(void)
{
  struct pair p = { 7, __builtin_complex(5.0, 12.0) };
  __complex__ double *q = &p.value;
  return norm(*q) == 13.0 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "complex-pointer-dereference-argument-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestTgmathComplexPointerDereferenceExpExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <tgmath.h>

struct pair {
  int tag;
  complex double value;
};

int main(void)
{
  struct pair p = { 7, __builtin_complex(0.0, 0.0) };
  complex double *q = &p.value;
  complex double r = exp(*q);
  return __builtin_cabs(r) == 1.0 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "tgmath-complex-pointer-dereference-exp-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestTgmathComplexPointerDereferencePowExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <tgmath.h>

struct pair {
  int tag;
  complex double value;
};

int main(void)
{
  struct pair p = { 7, __builtin_complex(2.0, 0.0) };
  complex double *q = &p.value;
  complex double r = pow(*q, 3.0);
  return __builtin_cabs(r) == 8.0 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "tgmath-complex-pointer-dereference-pow-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCComplexPointerDereferenceBinaryExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

struct pair {
  int tag;
  __complex__ double value;
};

int main(void)
{
  struct pair p = { 7, __builtin_complex(3.0, 4.0) };
  __complex__ double *q = &p.value;
  __complex__ double z = *q + __builtin_complex(0.0, 0.0);
  return __builtin_cabs(z) == 5.0 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "complex-pointer-dereference-binary-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCComplexPointerDereferenceCompareExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

struct pair {
  int tag;
  __complex__ double value;
};

int main(void)
{
  struct pair p = { 7, __builtin_complex(3.0, 4.0) };
  __complex__ double *q = &p.value;
  return *q == __builtin_complex(3.0, 4.0) ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "complex-pointer-dereference-compare-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCComplexExplicitCastNarrowsThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

int main(void)
{
  __complex__ double z = __builtin_complex(3.0, 4.0);
  __complex__ float f = (__complex__ float)z;
  return __builtin_cabsf(f) == 5.0f ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "complex-explicit-cast-narrow-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCComplexCompoundAssignRealScalarExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

int main(void)
{
  __complex__ double z = __builtin_complex(1.0, 4.0);
  z += 2.0;
  z -= 0.0;
  z *= 2.0;
  z /= 2.0;
  return __builtin_cabs(z) == 5.0 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "complex-compound-real-scalar-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCComplexAssignmentExpressionExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

int main(void)
{
  __complex__ double z = __builtin_complex(0.0, 0.0);
  if (__builtin_cabs(z = __builtin_complex(3.0, 4.0)) != 5.0)
    return 1;
  return __builtin_cabs(z) == 5.0 ? 0 : 2;
}
`
	st := runGCCExecFixture(t, "complex-assignment-expression-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCComplexCompoundAssignmentExpressionExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

int main(void)
{
  __complex__ double z = __builtin_complex(1.0, 4.0);
  if (__builtin_cabs(z += 2.0) != 5.0)
    return 1;
  return __builtin_cabs(z) == 5.0 ? 0 : 2;
}
`
	st := runGCCExecFixture(t, "complex-compound-assignment-expression-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCComplexCommaExpressionExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

int main(void)
{
  __complex__ double z = __builtin_complex(0.0, 0.0);
  __complex__ double w = (z = __builtin_complex(3.0, 4.0), __builtin_complex(5.0, 12.0));
  if (__builtin_cabs(z) != 5.0)
    return 1;
  return __builtin_cabs(w) == 13.0 ? 0 : 2;
}
`
	st := runGCCExecFixture(t, "complex-comma-expression-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCScalarReturnConvertsToComplexExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

__complex__ double make(void)
{
  return 3.0;
}

int main(void)
{
  __complex__ double z = make();
  return __builtin_cabs(z - __builtin_complex(3.0, 0.0)) == 0.0 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "scalar-return-converts-to-complex-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCComplexConstantExpressionInitializerExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

#define I (__extension__ 1.0iF)

int main(void)
{
  const __complex__ double z = 3.0 + 4.0 * I;
  return __builtin_cabs(z) == 5.0 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "complex-constant-expression-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCComplexConditionalReturnExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

__complex__ double choose(int flag)
{
  return flag ? __builtin_complex(3.0, 4.0) : __builtin_complex(6.0, 8.0);
}

int main(void)
{
  if (__builtin_cabs(choose(1)) != 5.0)
    return 1;
  if (__builtin_cabs(choose(0)) != 10.0)
    return 2;
  return 0;
}
`
	st := runGCCExecFixture(t, "complex-conditional-return-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCComplexConditionalInitializerExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

int main(void)
{
  int flag = 0;
  __complex__ double z = flag ? __builtin_complex(3.0, 4.0) : __builtin_complex(6.0, 8.0);
  return __builtin_cabs(z) == 10.0 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "complex-conditional-initializer-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCStaticComplexConstantExpressionInitializerExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

#define I (__extension__ 1.0iF)

const __complex__ double z = 3.0 + 4.0 * I;

int main(void)
{
  return __builtin_cabs(z) == 5.0 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "static-complex-constant-expression-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCStaticComplexConditionalConstantInitializerExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

const __complex__ double z = 1 ? __builtin_complex(3.0, 4.0) : __builtin_complex(6.0, 8.0);

int main(void)
{
  return __builtin_cabs(z) == 5.0 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "static-complex-conditional-constant-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCBuiltinCabslExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

int main(void)
{
  __complex__ long double z = __builtin_complex(3.0L, 4.0L);
  return __builtin_cabsl(z) == 5.0L ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "builtin-cabsl-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestTgmathComplexLongDoubleExpExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <tgmath.h>

int main(void)
{
  complex long double z = __builtin_complex(0.0L, 0.0L);
  complex long double r = exp(z);
  return __builtin_cabsl(r) == 1.0L ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "tgmath-complex-long-double-exp.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestTgmathComplexLongDoublePowExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <tgmath.h>

int main(void)
{
  complex long double z = __builtin_complex(2.0L, 0.0L);
  complex long double r = pow(z, 3.0L);
  return __builtin_cabsl(r) == 8.0L ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "tgmath-complex-long-double-pow.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func parseGCCExecManifest(t *testing.T, content string) []gccExecCase {
	t.Helper()
	cases, err := parseGCCExecManifestContent(content)
	if err != nil {
		t.Fatal(err)
	}
	return cases
}

func parseGCCExecManifestContent(content string) ([]gccExecCase, error) {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "path\texit\tcategory\treason" {
		return nil, fmt.Errorf("manifest header is missing or malformed: %q", firstManifestLine(content))
	}
	var cases []gccExecCase
	seen := map[string]bool{}
	for lineNo, line := range lines[1:] {
		if strings.TrimSpace(line) == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) != 4 {
			return nil, fmt.Errorf("manifest line %d malformed: %q", lineNo+2, line)
		}
		if seen[fields[0]] {
			return nil, fmt.Errorf("manifest line %d duplicates path %s", lineNo+2, fields[0])
		}
		seen[fields[0]] = true
		if !isAllowedGCCExecPath(fields[0]) {
			return nil, fmt.Errorf("manifest line %d path is outside supported GCC accept fixture roots: %s", lineNo+2, fields[0])
		}
		exitCode, err := strconv.Atoi(fields[1])
		if err != nil {
			return nil, fmt.Errorf("manifest line %d has invalid exit code %q: %v", lineNo+2, fields[1], err)
		}
		if strings.TrimSpace(fields[2]) == "" {
			return nil, fmt.Errorf("manifest line %d has empty category", lineNo+2)
		}
		if strings.TrimSpace(fields[3]) == "" {
			return nil, fmt.Errorf("manifest line %d has empty reason", lineNo+2)
		}
		cases = append(cases, gccExecCase{path: fields[0], exitCode: exitCode, category: fields[2], reason: fields[3]})
	}
	return cases, nil
}

func runGCCExecFixture(t *testing.T, path, source string) ExitStatus {
	t.Helper()
	const gccExecStepLimit = 100000

	src := stripGCCDirectives(source)
	pp, err := preprocessor.PreprocessSource(path, src, preprocessor.Options{})
	if err != nil {
		t.Fatalf("%s preprocess: %v", path, err)
	}
	if parserTokenCount(pp.Tokens) == 0 {
		t.Fatalf("%s preprocess produced no parser tokens", path)
	}
	candidates, err := parser.NewParser(pp.Tokens).Parse()
	if err != nil {
		t.Fatalf("%s parse: %v", path, err)
	}
	prog, err := sema.AnalyzeWithOptions(candidates, gccSemaOptions(source))
	if err != nil {
		t.Fatalf("%s sema: %v", path, err)
	}
	mod, err := codegen.Generate(prog)
	if err != nil {
		t.Fatalf("%s codegen: %v", path, err)
	}
	var encoded bytes.Buffer
	if err := bytecode.EncodeModule(&encoded, mod); err != nil {
		t.Fatalf("%s EncodeModule: %v", path, err)
	}
	p, err := Load(bytes.NewReader(encoded.Bytes()), LoadOptions{})
	if err != nil {
		t.Fatalf("%s Load: %v", path, err)
	}
	st, err := Run(context.Background(), p, RunOptions{StepLimit: gccExecStepLimit})
	if err != nil {
		t.Fatalf("%s Run: %v", path, err)
	}
	return st
}

func gccSemaOptions(src string) sema.SemaOptions {
	return sema.SemaOptions{
		PedanticErrors:                  gccPedanticErrors(src),
		GNUExtensions:                   gccGNUExtensions(src),
		Permissive:                      gccPermissive(src),
		WErrorDeclarationAfterStatement: gccWErrorDeclarationAfterStatement(src),
	}
}

func parserTokenCount(tokens []entity.Token) int {
	count := 0
	for _, tok := range tokens {
		if tok.Typ != entity.EOF {
			count++
		}
	}
	return count
}

func gccPedanticErrors(src string) bool {
	for _, line := range strings.Split(src, "\n") {
		if strings.Contains(line, "dg-options") && strings.Contains(line, "-pedantic-errors") {
			return true
		}
	}
	return false
}

func gccGNUExtensions(src string) bool {
	for _, line := range strings.Split(src, "\n") {
		if strings.Contains(line, "dg-options") && strings.Contains(line, "-std=gnu") {
			return true
		}
	}
	if strings.Contains(src, "empty initializer braces") && strings.Contains(src, "dg-warning") {
		return true
	}
	return false
}

func gccPermissive(src string) bool {
	for _, line := range strings.Split(src, "\n") {
		if strings.Contains(line, "dg-options") && strings.Contains(line, "-fpermissive") {
			return true
		}
	}
	return false
}

func gccWErrorDeclarationAfterStatement(src string) bool {
	for _, line := range strings.Split(src, "\n") {
		if strings.Contains(line, "dg-options") && strings.Contains(line, "-Werror=declaration-after-statement") {
			return true
		}
	}
	return false
}

func stripGCCDirectives(src string) string {
	var b strings.Builder
	inBlockComment := false
	for _, line := range strings.SplitAfter(src, "\n") {
		body := strings.TrimSuffix(line, "\n")
		newline := ""
		if strings.HasSuffix(line, "\n") {
			newline = "\n"
		}
		if isDejaGNULine(body) {
			if inBlockComment && strings.Contains(body, "*/") {
				b.WriteString("*/")
			}
			b.WriteString(newline)
			inBlockComment = updateBlockCommentState(body, inBlockComment)
			continue
		}
		b.WriteString(line)
		inBlockComment = updateBlockCommentState(body, inBlockComment)
	}
	return b.String()
}

func updateBlockCommentState(line string, inBlock bool) bool {
	for i := 0; i < len(line); i++ {
		if inBlock {
			if i+1 < len(line) && line[i] == '*' && line[i+1] == '/' {
				inBlock = false
				i++
			}
			continue
		}
		if i+1 < len(line) && line[i] == '/' && line[i+1] == '*' {
			inBlock = true
			i++
		}
	}
	return inBlock
}

func isDejaGNULine(line string) bool {
	trim := strings.TrimSpace(line)
	if strings.HasPrefix(trim, "/*") && strings.HasSuffix(trim, "*/") && strings.Contains(trim, "{ dg-") {
		return true
	}
	return strings.HasPrefix(trim, "//") && strings.Contains(trim, "{ dg-")
}

func hasGCCRunDirective(source string) bool {
	return strings.Contains(source, "{ dg-do run") || strings.Contains(source, "{ dg-require-effective-target c99_runtime")
}

func isAllowedGCCExecPath(manifestPath string) bool {
	if strings.Contains(manifestPath, "\\") {
		return false
	}
	if path.IsAbs(manifestPath) {
		return false
	}
	if cleaned := path.Clean(manifestPath); cleaned != manifestPath {
		return false
	}
	if !strings.HasSuffix(manifestPath, ".c") {
		return false
	}
	roots := []string{
		"sema/testdata/gcc-c99/accept/",
		"sema/testdata/gcc-c99-extra/accept/",
		"sema/testdata/gcc-c90-as-c99/accept/",
	}
	for _, root := range roots {
		if strings.HasPrefix(manifestPath, root) {
			return true
		}
	}
	return false
}

func firstManifestLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}
