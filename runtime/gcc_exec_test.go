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

func TestStdioFileOperationStubsExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <stdio.h>

int main(void)
{
  if (remove("missing.tmp") != -1)
    return 1;
  return rename("old.tmp", "new.tmp") == -1 ? 0 : 2;
}
`
	st := runGCCExecFixture(t, "stdio-file-operation-stubs-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestStdioOpenStubsExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <stdio.h>

int main(void)
{
  if (fopen("missing.tmp", "r") != 0)
    return 1;
  if (freopen("missing.tmp", "r", stdout) != 0)
    return 2;
  return tmpfile() == 0 ? 0 : 3;
}
`
	st := runGCCExecFixture(t, "stdio-open-stubs-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestStdioTmpnamExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <stdio.h>

int main(void)
{
  if (L_tmpnam < 1 || TMP_MAX < 1)
    return 1;
  return tmpnam(0) == 0 ? 0 : 2;
}
`
	st := runGCCExecFixture(t, "stdio-tmpnam-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
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

func TestStdioPositionStubsExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <stdio.h>

int main(void)
{
  if (fseek(stdout, 0, SEEK_SET) != -1)
    return 1;
  if (ftell(stdout) != -1L)
    return 2;
  rewind(stdin);
  return SEEK_SET == 0 && SEEK_CUR == 1 && SEEK_END == 2 ? 0 : 3;
}
`
	st := runGCCExecFixture(t, "stdio-position-stubs-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
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

func TestTgmathAsinExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <tgmath.h>

int main(void)
{
  if (asin(0.0f) != 0.0f)
    return 1;
  return asin(0.0L) == 0.0L ? 0 : 2;
}
`
	st := runGCCExecFixture(t, "tgmath-asin-real.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestTgmathAcosExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <tgmath.h>

int main(void)
{
  if (acos(1.0f) != 0.0f)
    return 1;
  return acos(1.0L) == 0.0L ? 0 : 2;
}
`
	st := runGCCExecFixture(t, "tgmath-acos-real.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestTgmathAtanExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <tgmath.h>

int main(void)
{
  if (atan(0.0f) != 0.0f)
    return 1;
  return atan(0.0L) == 0.0L ? 0 : 2;
}
`
	st := runGCCExecFixture(t, "tgmath-atan-real.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestTgmathAsinhExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <tgmath.h>

int main(void)
{
  if (asinh(0.0f) != 0.0f)
    return 1;
  return asinh(0.0L) == 0.0L ? 0 : 2;
}
`
	st := runGCCExecFixture(t, "tgmath-asinh-real.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestTgmathAcoshExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <tgmath.h>

int main(void)
{
  if (acosh(1.0f) != 0.0f)
    return 1;
  return acosh(1.0L) == 0.0L ? 0 : 2;
}
`
	st := runGCCExecFixture(t, "tgmath-acosh-real.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestTgmathAtanhExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <tgmath.h>

int main(void)
{
  if (atanh(0.0f) != 0.0f)
    return 1;
  return atanh(0.0L) == 0.0L ? 0 : 2;
}
`
	st := runGCCExecFixture(t, "tgmath-atanh-real.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestTgmathAtan2ExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <tgmath.h>

int main(void)
{
  if (atan2(0.0f, 1.0f) != 0.0f)
    return 1;
  return atan2(0.0L, 1.0L) == 0.0L ? 0 : 2;
}
`
	st := runGCCExecFixture(t, "tgmath-atan2-real.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestTgmathHypotExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <tgmath.h>

int main(void)
{
  if (hypot(3.0f, 4.0f) != 5.0f)
    return 1;
  return hypot(3.0L, 4.0L) == 5.0L ? 0 : 2;
}
`
	st := runGCCExecFixture(t, "tgmath-hypot-real.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestTgmathCbrtExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <tgmath.h>

int main(void)
{
  if (cbrt(8.0f) != 2.0f)
    return 1;
  return cbrt(8.0L) == 2.0L ? 0 : 2;
}
`
	st := runGCCExecFixture(t, "tgmath-cbrt-real.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestTgmathRoundingExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <tgmath.h>

int main(void)
{
  if (ceil(1.25f) != 2.0f)
    return 1;
  if (floor(1.75L) != 1.0L)
    return 2;
  if (trunc(-1.75f) != -1.0f)
    return 3;
  return round(-1.5L) == -2.0L ? 0 : 4;
}
`
	st := runGCCExecFixture(t, "tgmath-rounding-real.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestTgmathExponentialLogHelpersExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <tgmath.h>

int main(void)
{
  if (exp2(3.0f) != 8.0f)
    return 1;
  if (expm1(0.0L) != 0.0L)
    return 2;
  if (log10(1.0f) != 0.0f)
    return 3;
  if (log1p(0.0L) != 0.0L)
    return 4;
  return log2(8.0f) == 3.0f ? 0 : 5;
}
`
	st := runGCCExecFixture(t, "tgmath-exp-log-helpers-real.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestTgmathSpecialUnaryHelpersExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <tgmath.h>

int main(void)
{
  if (erf(0.0f) != 0.0f)
    return 1;
  if (erfc(0.0L) != 1.0L)
    return 2;
  if (tgamma(1.0f) != 1.0f)
    return 3;
  return lgamma(1.0L) == 0.0L ? 0 : 4;
}
`
	st := runGCCExecFixture(t, "tgmath-special-unary-helpers.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestTgmathRintHelpersExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <tgmath.h>

int main(void)
{
  if (nearbyint(2.0f) != 2.0f)
    return 1;
  return rint(3.0L) == 3.0L ? 0 : 2;
}
`
	st := runGCCExecFixture(t, "tgmath-rint-helpers.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestTgmathLogbExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <tgmath.h>

int main(void)
{
  return logb(8.0f) == 3.0f ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "tgmath-logb.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestTgmathIlogbExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <tgmath.h>

int main(void)
{
  return ilogb(8.0f) == 3 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "tgmath-ilogb.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestTgmathLongRoundingHelpersExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <tgmath.h>

int main(void)
{
  long a = lrint(3.0f);
  long b = lround(4.0L);
  long long c = llrint(5.0);
  long long d = llround(6.0f);
  return a + b + c + d == 18 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "tgmath-long-rounding.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestTgmathFrexpExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <tgmath.h>

int main(void)
{
  int exp = 0;
  double frac = frexp(8.0, &exp);
  if (frac != 0.5)
    return 1;
  return exp == 4 ? 0 : 2;
}
`
	st := runGCCExecFixture(t, "tgmath-frexp.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestTgmathLdexpExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <tgmath.h>

int main(void)
{
  return ldexp(2.0f, 3) == 16.0f ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "tgmath-ldexp.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestTgmathRemquoExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <tgmath.h>

int main(void)
{
  int quo = 0;
  float rem = remquo(4.0f, 2.0f, &quo);
  if (rem != 0.0f)
    return 1;
  return quo == 2 ? 0 : 2;
}
`
	st := runGCCExecFixture(t, "tgmath-remquo.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestBuiltinPowExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

int main(void)
{
  return __builtin_pow(2.0, 3.0) == 8.0 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "builtin-pow-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestBuiltinFloatingConstantsExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

double nan(const char *);

int main(void)
{
  float hf = __builtin_huge_valf();
  double h = __builtin_huge_val();
  long double hl = __builtin_huge_vall();
  double n = __builtin_nan("");
  double m = nan("");

  if (!(hf > 1e30f))
    return 1;
  if (!(h > 1e300))
    return 2;
  if (!(hl > 1e300L))
    return 3;
  if (!(n != n))
    return 4;
  return m != m ? 0 : 5;
}
`
	st := runGCCExecFixture(t, "builtin-floating-constants-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestBuiltinAllocationExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

int main(void)
{
  char *p = __builtin_malloc(4);
  if (p == 0)
    return 1;
  p[0] = 'a';
  p[1] = 0;
  if (p[0] != 'a')
    return 2;

  char *q = __builtin_calloc(3, 2);
  if (q == 0)
    return 3;
  if (q[0] != 0 || q[5] != 0)
    return 4;
  q[5] = 'z';
  if (q[5] != 'z')
    return 5;

  char *r = __builtin_strdup("hi");
  if (r == 0)
    return 6;
  return r[0] == 'h' && r[1] == 'i' && r[2] == 0 ? 0 : 7;
}
`
	st := runGCCExecFixture(t, "builtin-allocation-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestPlainAllocationExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <stdlib.h>

int main(void)
{
  char *p = malloc(4);
  if (p == 0)
    return 1;
  p[0] = 'a';
  p[1] = 0;
  if (p[0] != 'a')
    return 2;

  char *q = calloc(3, 2);
  if (q == 0)
    return 3;
  if (q[0] != 0 || q[5] != 0)
    return 4;
  q[5] = 'z';
  if (q[5] != 'z')
    return 5;

  char *r = strdup("hi");
  if (r == 0)
    return 6;
  if (r[0] != 'h' || r[1] != 'i' || r[2] != 0)
    return 7;

  free(p);
  free(q);
  free(r);
  free(0);
  return 0;
}
`
	st := runGCCExecFixture(t, "plain-allocation-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestStdlibReallocExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <stdlib.h>

int main(void)
{
  char *p = malloc(3);
  char *q;
  if (p == 0)
    return 1;
  p[0] = 'a';
  p[1] = 'b';
  p[2] = 0;

  q = realloc(p, 6);
  if (q == 0)
    return 2;
  if (q[0] != 'a' || q[1] != 'b' || q[2] != 0)
    return 3;
  q[3] = 'c';
  q[4] = 0;

  p = realloc(q, 2);
  if (p == 0)
    return 4;
  if (p[0] != 'a' || p[1] != 'b')
    return 5;
  if (realloc(p, 0) != 0)
    return 6;

  p = realloc(0, 4);
  if (p == 0)
    return 7;
  free(p);
  return 0;
}
`
	st := runGCCExecFixture(t, "stdlib-realloc-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestStdlibRandExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <stdlib.h>

int main(void)
{
  int a;
  int b;
  int c;

  if (RAND_MAX < 32767)
    return 1;
  srand(123);
  a = rand();
  b = rand();
  if (a < 0 || b < 0 || a > RAND_MAX || b > RAND_MAX)
    return 2;
  if (a == b)
    return 3;
  srand(123);
  c = rand();
  return c == a ? 0 : 4;
}
`
	st := runGCCExecFixture(t, "stdlib-rand-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestStdlibGetenvExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <stdlib.h>

int main(void)
{
  if (getenv("PATH") != 0)
    return 1;
  return getenv("") == 0 ? 0 : 2;
}
`
	st := runGCCExecFixture(t, "stdlib-getenv-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestStdlibSystemExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <stdlib.h>

int main(void)
{
  if (system(0) != 0)
    return 1;
  return system("echo nope") == -1 ? 0 : 2;
}
`
	st := runGCCExecFixture(t, "stdlib-system-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestStdlibAtexitExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <stdlib.h>

static void cleanup(void)
{
}

int main(void)
{
  return atexit(cleanup) == 0 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "stdlib-atexit-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestLocaleSetlocaleExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <locale.h>
#include <string.h>

int main(void)
{
  char *current = setlocale(LC_ALL, 0);
  if (current == 0 || strcmp(current, "C") != 0)
    return 1;
  if (setlocale(LC_NUMERIC, "C") == 0)
    return 2;
  return setlocale(LC_TIME, "ja_JP.UTF-8") == 0 ? 0 : 3;
}
`
	st := runGCCExecFixture(t, "locale-setlocale-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestTimeHeaderExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <time.h>

int main(void)
{
  time_t now = 99;
  if (CLOCKS_PER_SEC != 1000000L)
    return 1;
  if (time(0) != 0)
    return 2;
  if (time(&now) != 0 || now != 0)
    return 3;
  if (difftime(7, 2) != 5.0)
    return 4;
  return clock() == 0 ? 0 : 5;
}
`
	st := runGCCExecFixture(t, "time-header-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestErrnoHeaderExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <errno.h>

int main(void)
{
  if (errno != 0)
    return 1;
  errno = ERANGE;
  if (errno != ERANGE)
    return 2;
  if (EDOM == ERANGE || EILSEQ == 0)
    return 3;
  return 0;
}
`
	st := runGCCExecFixture(t, "errno-header-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestAssertHeaderExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <assert.h>

int main(void)
{
  assert(1);
  return 0;
}
`
	st := runGCCExecFixture(t, "assert-header-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestStdlibDivExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <stdlib.h>

int main(void)
{
  div_t d = div(-7, 3);
  ldiv_t ld = ldiv(-9L, 4L);
  lldiv_t lld = lldiv(10LL, -4LL);

  if (d.quot != -2 || d.rem != -1)
    return 1;
  if (ld.quot != -2 || ld.rem != -1)
    return 2;
  return lld.quot == -2 && lld.rem == 2 ? 0 : 3;
}
`
	st := runGCCExecFixture(t, "stdlib-div-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestStdlibExitExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <stdlib.h>

int main(void)
{
  exit(23);
  return 1;
}
`
	st := runGCCExecFixture(t, "stdlib-exit-runtime.c", source)
	if st.Code != 23 {
		t.Fatalf("exit code = %d, want 23", st.Code)
	}
}

func TestStdlibImmediateExitExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <stdlib.h>

int main(void)
{
  _Exit(24);
  return 1;
}
`
	st := runGCCExecFixture(t, "stdlib-immediate-exit-runtime.c", source)
	if st.Code != 24 {
		t.Fatalf("exit code = %d, want 24", st.Code)
	}
}

func TestBuiltinObjectSizeExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

int main(void)
{
  char *p = __builtin_malloc(4);
  if (__builtin_object_size(p, 0) != (unsigned long)-1)
    return 1;
  if (__builtin_object_size(p, 2) != 0)
    return 2;
  if (__builtin_dynamic_object_size(p, 1) != (unsigned long)-1)
    return 3;
  return __builtin_dynamic_object_size(p, 3) == 0 ? 0 : 4;
}
`
	st := runGCCExecFixture(t, "builtin-object-size-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestBuiltinSprintfExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

int strcmp(const char *, const char *);

int main(void)
{
  char buf[32];
  int n = __builtin_sprintf(buf, "x=%d %s %c %%", -7, "ok", '!');
  if (n != 11)
    return 1;
  if (strcmp(buf, "x=-7 ok ! %") != 0)
    return 2;

  char small[5];
  n = __builtin_snprintf(small, 5, "%s-%u", "abcdef", 3U);
  if (n != 8)
    return 3;
  return strcmp(small, "abcd") == 0 ? 0 : 4;
}
`
	st := runGCCExecFixture(t, "builtin-sprintf-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestBuiltinCheckedSprintfExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

int strcmp(const char *, const char *);

int main(void)
{
  char buf[32];
  int n = __builtin___sprintf_chk(buf, 0, 32, "x=%d %s", -7, "ok");
  if (n != 7)
    return 1;
  if (strcmp(buf, "x=-7 ok") != 0)
    return 2;

  char small[5];
  n = __builtin___snprintf_chk(small, 5, 0, 5, "%s-%u", "abcdef", 3U);
  if (n != 8)
    return 3;
  return strcmp(small, "abcd") == 0 ? 0 : 4;
}
`
	st := runGCCExecFixture(t, "builtin-checked-sprintf-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestStdioPutcharExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <stdio.h>

int main(void)
{
  return putchar('A') == 'A' ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "stdio-putchar-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestStdioPutsUnlockedExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <stdio.h>

int main(void)
{
  return puts_unlocked("unlocked") == 9 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "stdio-puts-unlocked-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestStdioFputcExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <stdio.h>

int main(void)
{
  return fputc('B', stdout) == 'B' ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "stdio-fputc-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestStdioFflushExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <stdio.h>

int main(void)
{
  if (fputc('C', stdout) != 'C')
    return 1;
  if (fflush(stdout) != 0)
    return 2;
  return fflush_unlocked(stdout) == 0 ? 0 : 3;
}
`
	st := runGCCExecFixture(t, "stdio-fflush-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestStdioFcloseExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <stdio.h>

int main(void)
{
  return fclose(stdout) == 0 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "stdio-fclose-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestStdioFilenoExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <stdio.h>

int main(void)
{
  if (fileno(stdin) != 0)
    return 1;
  if (fileno(stdout) != 1)
    return 2;
  if (fileno_unlocked(stderr) != 2)
    return 3;
  return 0;
}
`
	st := runGCCExecFixture(t, "stdio-fileno-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestStdioBufferControlsExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <stdio.h>

int main(void)
{
  char buf[BUFSIZ];
  setbuf(stdout, buf);
  if (setvbuf(stdout, 0, _IONBF, 0) != 0)
    return 1;
  return fputc('S', stdout) == 'S' ? 0 : 2;
}
`
	st := runGCCExecFixture(t, "stdio-buffer-controls-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestStdioStreamLockControlsExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <stdio.h>

int main(void)
{
  flockfile(stdout);
  if (ftrylockfile(stdout) != 0)
    return 1;
  funlockfile(stdout);
  return fputc('L', stdout) == 'L' ? 0 : 2;
}
`
	st := runGCCExecFixture(t, "stdio-stream-lock-controls-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestStdioPerrorExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <stdio.h>

int main(void)
{
  perror("phase");
  perror("");
  perror(0);
  return 0;
}
`
	st := runGCCExecFixture(t, "stdio-perror-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestStdlibAbsExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <stdlib.h>

int main(void)
{
  if (abs(-7) != 7)
    return 1;
  if (labs(-9L) != 9L)
    return 2;
  if (llabs(-11LL) != 11LL)
    return 3;
  return 0;
}
`
	st := runGCCExecFixture(t, "stdlib-abs-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestStdlibAtoiExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <stdlib.h>

int main(void)
{
  if (atoi(" \t-42tail") != -42)
    return 1;
  if (atol("\n12345x") != 12345L)
    return 2;
  if (atoll("  -9876543210") != -9876543210LL)
    return 3;
  if (atoi("nondigit") != 0)
    return 4;
  return 0;
}
`
	st := runGCCExecFixture(t, "stdlib-atoi-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestStdlibStrtolExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <stdlib.h>

int main(void)
{
  char *end = 0;

  if (strtol(" \t-0x2a tail", &end, 0) != -42L)
    return 1;
  if (*end != ' ')
    return 2;
  if (strtoul("077z", &end, 0) != 63UL)
    return 3;
  if (*end != 'z')
    return 4;
  if (strtoul("7f!", 0, 16) != 127UL)
    return 5;
  if (strtol("xyz", &end, 10) != 0L)
    return 6;
  return *end == 'x' ? 0 : 7;
}
`
	st := runGCCExecFixture(t, "stdlib-strtol-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestStdlibStrtollExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <stdlib.h>

int main(void)
{
  char *end = 0;

  if (strtoll("-9876543210x", &end, 10) != -9876543210LL)
    return 1;
  if (*end != 'x')
    return 2;
  if (strtoull("0X100000000z", &end, 0) != 4294967296ULL)
    return 3;
  return *end == 'z' ? 0 : 4;
}
`
	st := runGCCExecFixture(t, "stdlib-strtoll-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestStdlibFloatParserExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <stdlib.h>

int main(void)
{
  char *end = 0;

  if (atof(" \t3.25tail") != 3.25)
    return 1;
  if (strtod(" -12.5e1x", &end) != -125.0)
    return 2;
  if (*end != 'x')
    return 3;
  if (strtod("0x1.8p+2!", &end) != 6.0)
    return 4;
  if (*end != '!')
    return 5;
  if (strtod("word", &end) != 0.0)
    return 6;
  return *end == 'w' ? 0 : 7;
}
`
	st := runGCCExecFixture(t, "stdlib-float-parser-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestStdlibMoreFloatParserExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <stdlib.h>

int main(void)
{
  char *end = 0;

  if (strtof(" 1.5!", &end) != 1.5f)
    return 1;
  if (*end != '!')
    return 2;
  if (strtold("-0x1.4p+2z", &end) != -5.0L)
    return 3;
  return *end == 'z' ? 0 : 4;
}
`
	st := runGCCExecFixture(t, "stdlib-more-float-parser-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestCtypeClassificationExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <ctype.h>

int main(void)
{
  if (!isdigit('5') || isdigit('x'))
    return 1;
  if (!isalpha('Z') || !isalnum('7'))
    return 2;
  if (!isspace('\n') || !islower('q') || !isupper('Q'))
    return 3;
  if (!isxdigit('f') || isxdigit('G'))
    return 4;
  if (!isprint(' ') || isprint(127))
    return 5;
  return 0;
}
`
	st := runGCCExecFixture(t, "ctype-classification-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestCtypeExtraClassificationExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <ctype.h>

int main(void)
{
  if (!isblank(' ') || !isblank('\t') || isblank('\n'))
    return 1;
  if (!iscntrl(31) || !iscntrl(127) || iscntrl('A'))
    return 2;
  if (!isgraph('!') || isgraph(' '))
    return 3;
  if (!ispunct('!') || ispunct('A') || ispunct(' '))
    return 4;
  return 0;
}
`
	st := runGCCExecFixture(t, "ctype-extra-classification-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestCtypeCaseConversionExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <ctype.h>

int main(void)
{
  if (tolower('A') != 'a')
    return 1;
  if (tolower('z') != 'z')
    return 2;
  if (toupper('q') != 'Q')
    return 3;
  if (toupper('!') != '!')
    return 4;
  return 0;
}
`
	st := runGCCExecFixture(t, "ctype-case-conversion-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestStringHeaderReadOnlyHelpersExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <string.h>

int main(void)
{
  const char *text = "abcdef";
  if (strlen(text) != 6)
    return 1;
  if (strchr(text, 'd') != text + 3)
    return 2;
  if (strchr(text, 'z') != 0)
    return 3;
  if (strstr(text, "cd") != text + 2)
    return 4;
  if (strcmp(text, "abcdef") != 0)
    return 5;
  return memcmp(text, "abcxef", 4) < 0 ? 0 : 6;
}
`
	st := runGCCExecFixture(t, "string-read-only-header-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestStringBoundedCompareSearchExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <string.h>

int main(void)
{
  const char *left = "abcdef";
  const char *right = "abcxyz";
  char data[5] = {1, 2, 3, 2, 0};

  if (strncmp(left, right, 3) != 0)
    return 1;
  if (strncmp(left, right, 4) >= 0)
    return 2;
  if (strncmp(left, "abbzzz", 3) <= 0)
    return 3;
  if (memchr(data, 2, 4) != data + 1)
    return 4;
  return memchr(data, 9, 4) == 0 ? 0 : 5;
}
`
	st := runGCCExecFixture(t, "string-bounded-compare-search-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestStringReverseAndSetSearchExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <string.h>

int main(void)
{
  const char *text = "abacad";

  if (strrchr(text, 'a') != text + 4)
    return 1;
  if (strrchr(text, 0) != text + 6)
    return 2;
  if (strrchr(text, 'z') != 0)
    return 3;
  if (strpbrk(text, "xyc") != text + 3)
    return 4;
  return strpbrk(text, "xyz") == 0 ? 0 : 5;
}
`
	st := runGCCExecFixture(t, "string-reverse-set-search-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestStringSpanExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <string.h>

int main(void)
{
  const char *text = "abcde312";

  if (strspn(text, "abc") != 3)
    return 1;
  if (strspn(text, "xyz") != 0)
    return 2;
  if (strcspn(text, "de") != 3)
    return 3;
  return strcspn(text, "xyz") == 8 ? 0 : 4;
}
`
	st := runGCCExecFixture(t, "string-span-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestStringNLengthExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <string.h>

int main(void)
{
  if (strnlen("abc", 2) != 2)
    return 1;
  if (strnlen("hello", 10) != 5)
    return 2;
  return strnlen("hello", 3) == 3 ? 0 : 3;
}
`
	st := runGCCExecFixture(t, "string-strnlen-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestStringStrerrorExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <string.h>

int main(void)
{
  char *a = strerror(1);
  char *b = strerror(2);

  if (strcmp(a, "error") != 0)
    return 1;
  return a == b ? 0 : 2;
}
`
	st := runGCCExecFixture(t, "string-strerror-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestPlainMemoryOperationsExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <string.h>

int main(void)
{
  char dst[8] = "abcdef";
  char src[4] = "XYZ";

  if (memcpy(dst, src, 3) != dst)
    return 1;
  if (dst[0] != 'X' || dst[1] != 'Y' || dst[2] != 'Z')
    return 2;

  if (memmove(dst + 1, dst, 4) != dst + 1)
    return 3;
  if (dst[0] != 'X' || dst[1] != 'X' || dst[2] != 'Y' || dst[3] != 'Z' || dst[4] != 'd')
    return 4;

  if (mempcpy(dst + 2, "pq", 2) != dst + 4)
    return 5;
  if (dst[2] != 'p' || dst[3] != 'q')
    return 6;

  if (memset(dst + 4, 'r', 2) != dst + 4)
    return 7;
  if (dst[4] != 'r' || dst[5] != 'r')
    return 8;

  bzero(dst + 5, 2);
  return dst[5] == 0 && dst[6] == 0 ? 0 : 9;
}
`
	st := runGCCExecFixture(t, "plain-memory-ops-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestStringsBSDMemoryExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <strings.h>

int main(void)
{
  const char src[5] = "abcd";
  char dst[5] = "xxxx";

  if (bcmp(src, "abce", 4) >= 0)
    return 1;
  bcopy(src, dst, 4);
  if (bcmp(dst, "abcd", 4) != 0)
    return 2;
  bzero(dst + 2, 2);
  return dst[2] == 0 && dst[3] == 0 ? 0 : 3;
}
`
	st := runGCCExecFixture(t, "strings-bsd-memory-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestPlainStringWritesExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <string.h>

int main(void)
{
  char buf[12];

  if (strcpy(buf, "ab") != buf)
    return 1;
  if (buf[0] != 'a' || buf[1] != 'b' || buf[2] != 0)
    return 2;

  if (stpcpy(buf + 2, "cd") != buf + 4)
    return 3;
  if (buf[0] != 'a' || buf[1] != 'b' || buf[2] != 'c' || buf[3] != 'd' || buf[4] != 0)
    return 4;

  if (strcat(buf, "ef") != buf)
    return 5;
  if (buf[4] != 'e' || buf[5] != 'f' || buf[6] != 0)
    return 6;

  if (strncpy(buf, "xy", 4) != buf)
    return 7;
  if (buf[0] != 'x' || buf[1] != 'y' || buf[2] != 0 || buf[3] != 0)
    return 8;

  if (stpncpy(buf, "pqrs", 2) != buf + 2)
    return 9;
  if (buf[0] != 'p' || buf[1] != 'q')
    return 10;

  buf[0] = 0;
  if (strncat(buf, "uvwx", 2) != buf)
    return 11;
  return buf[0] == 'u' && buf[1] == 'v' && buf[2] == 0 ? 0 : 12;
}
`
	st := runGCCExecFixture(t, "plain-string-writes-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestStringStrtokExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <string.h>

int main(void)
{
  char text[24];
  char *tok;

  strcpy(text, ",alpha,beta;gamma");
  tok = strtok(text, ",;");
  if (tok != text + 1 || strcmp(tok, "alpha") != 0)
    return 1;
  tok = strtok(0, ",;");
  if (tok != text + 7 || strcmp(tok, "beta") != 0)
    return 2;
  tok = strtok(0, ",;");
  if (tok != text + 12 || strcmp(tok, "gamma") != 0)
    return 3;
  return strtok(0, ",;") == 0 ? 0 : 4;
}
`
	st := runGCCExecFixture(t, "string-strtok-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestStringCollateTransformExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <string.h>

int main(void)
{
  char buf[4];

  if (strcoll("abc", "abd") >= 0)
    return 1;
  if (strxfrm(buf, "abcdef", sizeof buf) != 6)
    return 2;
  return strcmp(buf, "abc") == 0 ? 0 : 3;
}
`
	st := runGCCExecFixture(t, "string-collate-transform-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestStdioStatusFunctionsExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <stdio.h>

int main(void)
{
  if (ferror(stdout) != 0)
    return 1;
  clearerr(stdout);
  if (feof(stdout) != 0)
    return 2;
  return 0;
}
`
	st := runGCCExecFixture(t, "stdio-status-functions-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestStdioFwriteExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <stdio.h>

int main(void)
{
  char data[4] = { 'D', 'E', 'F', 'G' };
  return fwrite(data, 2, 2, stdout) == 2 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "stdio-fwrite-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestStdioFreadEmptyInputExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <stdio.h>

int main(void)
{
  char buf[2] = { 'x', 'y' };
  if (fread(buf, 1, 2, stdin) != 0)
    return 1;
  return buf[0] == 'x' && buf[1] == 'y' ? 0 : 2;
}
`
	st := runGCCExecFixture(t, "stdio-fread-empty-input-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestStdioFreadPushbackExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <stdio.h>

int main(void)
{
  char buf[4] = { 0 };
  ungetc('C', stdin);
  ungetc('B', stdin);
  ungetc('A', stdin);
  if (fread(buf, 1, 3, stdin) != 3)
    return 1;
  if (buf[0] != 'A' || buf[1] != 'B' || buf[2] != 'C' || buf[3] != 0)
    return 2;
  if (fread(buf, 1, 1, stdin) != 0)
    return 3;
  if (feof(stdin) == 0)
    return 4;
  return 0;
}
`
	st := runGCCExecFixture(t, "stdio-fread-pushback-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestStdioUnlockedBlockIOAliasesExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <stdio.h>

int main(void)
{
  char buf[4] = { 0 };
  ungetc('Y', stdin);
  ungetc('X', stdin);
  if (fread_unlocked(buf, 1, 2, stdin) != 2)
    return 1;
  if (buf[0] != 'X' || buf[1] != 'Y')
    return 2;
  if (fwrite_unlocked(buf, 1, 2, stdout) != 2)
    return 3;
  ungetc('\n', stdin);
  ungetc('Z', stdin);
  if (fgets_unlocked(buf, sizeof buf, stdin) != buf)
    return 4;
  if (buf[0] != 'Z' || buf[1] != '\n' || buf[2] != 0)
    return 5;
  return 0;
}
`
	st := runGCCExecFixture(t, "stdio-unlocked-block-io-aliases-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestStdioGetcharFgetcEmptyInputExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <stdio.h>

int main(void)
{
  if (fgetc(stdin) != EOF)
    return 1;
  if (getchar() != EOF)
    return 2;
  return 0;
}
`
	st := runGCCExecFixture(t, "stdio-getchar-fgetc-empty-input-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestStdioUngetcExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <stdio.h>

int main(void)
{
  if (ungetc('R', stdin) != 'R')
    return 1;
  if (fgetc(stdin) != 'R')
    return 2;
  if (fgetc(stdin) != EOF)
    return 3;
  if (ungetc('S', stdin) != 'S')
    return 4;
  if (getchar() != 'S')
    return 5;
  return 0;
}
`
	st := runGCCExecFixture(t, "stdio-ungetc-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestStdioFgetsPushbackExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <stdio.h>

int main(void)
{
  char buf[4] = { 0 };
  ungetc('\n', stdin);
  ungetc('i', stdin);
  ungetc('H', stdin);
  if (fgets(buf, sizeof buf, stdin) != buf)
    return 1;
  if (buf[0] != 'H' || buf[1] != 'i' || buf[2] != '\n' || buf[3] != 0)
    return 2;
  if (fgets(buf, sizeof buf, stdin) != 0)
    return 3;
  return 0;
}
`
	st := runGCCExecFixture(t, "stdio-fgets-pushback-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestStdioEOFStatusExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <stdio.h>

int main(void)
{
  if (feof(stdin) != 0)
    return 1;
  if (fgetc(stdin) != EOF)
    return 2;
  if (feof(stdin) == 0)
    return 3;
  clearerr(stdin);
  if (feof(stdin) != 0)
    return 4;
  if (fgetc(stdin) != EOF)
    return 5;
  if (ungetc('A', stdin) != 'A')
    return 6;
  if (feof(stdin) != 0)
    return 7;
  if (fgetc(stdin) != 'A')
    return 8;
  return 0;
}
`
	st := runGCCExecFixture(t, "stdio-eof-status-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestStdioUnlockedStatusAliasesExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <stdio.h>

int main(void)
{
  if (fgetc(stdin) != EOF)
    return 1;
  if (feof_unlocked(stdin) == 0)
    return 2;
  if (ferror_unlocked(stdin) != 0)
    return 3;
  clearerr_unlocked(stdin);
  if (feof_unlocked(stdin) != 0)
    return 4;
  return 0;
}
`
	st := runGCCExecFixture(t, "stdio-unlocked-status-aliases-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestStdioInputUnlockedAliasesExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <stdio.h>

int main(void)
{
  ungetc('G', stdin);
  if (getc(stdin) != 'G')
    return 1;
  ungetc('F', stdin);
  if (fgetc_unlocked(stdin) != 'F')
    return 2;
  ungetc('U', stdin);
  if (getc_unlocked(stdin) != 'U')
    return 3;
  ungetc('H', stdin);
  if (getchar_unlocked() != 'H')
    return 4;
  return 0;
}
`
	st := runGCCExecFixture(t, "stdio-input-unlocked-aliases-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestStdioOutputUnlockedAliasesExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <stdio.h>

int main(void)
{
  if (putc('P', stdout) != 'P')
    return 1;
  if (putc_unlocked('U', stdout) != 'U')
    return 2;
  if (putchar_unlocked('H') != 'H')
    return 3;
  return 0;
}
`
	st := runGCCExecFixture(t, "stdio-output-unlocked-aliases-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestBuiltinMemoryOpsExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

int main(void)
{
  char dst[8] = "abcdef";
  char src[4] = "XYZ";

  if (__builtin_memcpy(dst, src, 3) != dst)
    return 1;
  if (dst[0] != 'X' || dst[1] != 'Y' || dst[2] != 'Z')
    return 2;

  if (__builtin_memmove(dst + 1, dst, 4) != dst + 1)
    return 3;
  if (dst[0] != 'X' || dst[1] != 'X' || dst[2] != 'Y' || dst[3] != 'Z' || dst[4] != 'd')
    return 4;

  if (__builtin_mempcpy(dst + 2, "pq", 2) != dst + 4)
    return 5;
  if (dst[2] != 'p' || dst[3] != 'q')
    return 6;

  if (__builtin_memset(dst + 4, 'r', 2) != dst + 4)
    return 7;
  if (dst[4] != 'r' || dst[5] != 'r')
    return 8;

  __builtin_bzero(dst + 5, 2);
  return dst[5] == 0 && dst[6] == 0 ? 0 : 9;
}
`
	st := runGCCExecFixture(t, "builtin-memory-ops-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestBuiltinCheckedMemoryOpsExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

int main(void)
{
  char dst[10];
  char src[5] = "abcd";

  if (__builtin___memcpy_chk(dst, src, 5, 10) != dst)
    return 1;
  if (dst[0] != 'a' || dst[3] != 'd' || dst[4] != 0)
    return 2;

  if (__builtin___memmove_chk(dst + 1, dst, 4, 9) != dst + 1)
    return 3;
  if (dst[0] != 'a' || dst[1] != 'a' || dst[2] != 'b' || dst[3] != 'c')
    return 4;

  if (__builtin___mempcpy_chk(dst + 4, "xy", 2, 6) != dst + 6)
    return 5;
  if (dst[4] != 'x' || dst[5] != 'y')
    return 6;

  if (__builtin___memset_chk(dst + 6, 'z', 2, 4) != dst + 6)
    return 7;
  return dst[6] == 'z' && dst[7] == 'z' ? 0 : 8;
}
`
	st := runGCCExecFixture(t, "builtin-checked-memory-ops-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestBuiltinStringSearchExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

int main(void)
{
  char text[8] = "abcdef";

  if (__builtin_strlen(text) != 6)
    return 1;
  if (__builtin_strchr(text, 'd') != text + 3)
    return 2;
  if (__builtin_strchr(text, 'z') != 0)
    return 3;
  if (__builtin_strchr(text, 0) != text + 6)
    return 4;
  if (__builtin_strstr(text, "cd") != text + 2)
    return 5;
  return __builtin_strstr(text, "gh") == 0 ? 0 : 6;
}
`
	st := runGCCExecFixture(t, "builtin-string-search-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestBuiltinStringWritesExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

int main(void)
{
  char buf[12];

  if (__builtin_strcpy(buf, "ab") != buf)
    return 1;
  if (buf[0] != 'a' || buf[1] != 'b' || buf[2] != 0)
    return 2;

  if (__builtin_stpcpy(buf + 2, "cd") != buf + 4)
    return 3;
  if (buf[0] != 'a' || buf[1] != 'b' || buf[2] != 'c' || buf[3] != 'd' || buf[4] != 0)
    return 4;

  if (__builtin_strcat(buf, "ef") != buf)
    return 5;
  if (buf[4] != 'e' || buf[5] != 'f' || buf[6] != 0)
    return 6;

  if (__builtin_strncpy(buf, "xy", 4) != buf)
    return 7;
  if (buf[0] != 'x' || buf[1] != 'y' || buf[2] != 0 || buf[3] != 0)
    return 8;

  if (__builtin_stpncpy(buf, "pqrs", 2) != buf + 2)
    return 9;
  if (buf[0] != 'p' || buf[1] != 'q')
    return 10;

  buf[0] = 0;
  if (__builtin_strncat(buf, "uvwx", 2) != buf)
    return 11;
  return buf[0] == 'u' && buf[1] == 'v' && buf[2] == 0 ? 0 : 12;
}
`
	st := runGCCExecFixture(t, "builtin-string-writes-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestBuiltinCheckedStringWritesExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */

int main(void)
{
  char buf[12];

  if (__builtin___strcpy_chk(buf, "ab", 12) != buf)
    return 1;
  if (buf[0] != 'a' || buf[1] != 'b' || buf[2] != 0)
    return 2;

  if (__builtin___stpcpy_chk(buf + 2, "cd", 10) != buf + 4)
    return 3;
  if (buf[2] != 'c' || buf[3] != 'd' || buf[4] != 0)
    return 4;

  if (__builtin___strcat_chk(buf, "ef", 12) != buf)
    return 5;
  if (buf[4] != 'e' || buf[5] != 'f' || buf[6] != 0)
    return 6;

  if (__builtin___strncpy_chk(buf, "xy", 4, 12) != buf)
    return 7;
  if (buf[0] != 'x' || buf[1] != 'y' || buf[2] != 0 || buf[3] != 0)
    return 8;

  if (__builtin___stpncpy_chk(buf, "pqrs", 2, 12) != buf + 2)
    return 9;
  if (buf[0] != 'p' || buf[1] != 'q')
    return 10;

  buf[0] = 0;
  if (__builtin___strncat_chk(buf, "uvwx", 2, 12) != buf)
    return 11;
  return buf[0] == 'u' && buf[1] == 'v' && buf[2] == 0 ? 0 : 12;
}
`
	st := runGCCExecFixture(t, "builtin-checked-string-writes-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestTgmathFabsExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <tgmath.h>

int main(void)
{
  if (fabs(-3.0f) != 3.0f)
    return 1;
  return fabs(__builtin_complex(3.0, 4.0)) == 5.0 ? 0 : 2;
}
`
	st := runGCCExecFixture(t, "tgmath-fabs.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestTgmathComplexProjectionExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <tgmath.h>

int main(void)
{
  complex float z = __builtin_complex(3.0f, 4.0f);
  if (creal(z) != 3.0f)
    return 1;
  return cimag(z) == 4.0f ? 0 : 2;
}
`
	st := runGCCExecFixture(t, "tgmath-complex-projection.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestTgmathConjExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <tgmath.h>

int main(void)
{
  complex float z = __builtin_complex(3.0f, 4.0f);
  complex float r = conj(z);
  if (creal(r) != 3.0f)
    return 1;
  return cimag(r) == -4.0f ? 0 : 2;
}
`
	st := runGCCExecFixture(t, "tgmath-conj.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestTgmathCargExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <tgmath.h>

int main(void)
{
  complex float z = __builtin_complex(1.0f, 0.0f);
  return carg(z) == 0.0f ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "tgmath-carg.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestTgmathCprojExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <tgmath.h>

int main(void)
{
  complex float z = __builtin_complex(3.0f, 4.0f);
  complex float r = cproj(z);
  if (creal(r) != 3.0f)
    return 1;
  return cimag(r) == 4.0f ? 0 : 2;
}
`
	st := runGCCExecFixture(t, "tgmath-cproj.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestTgmathBinaryRealHelpersExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <tgmath.h>

int main(void)
{
  if (fdim(5.0f, 3.0f) != 2.0f)
    return 1;
  if (fmax(2.0L, 3.0L) != 3.0L)
    return 2;
  if (fmin(2.0f, 3.0f) != 2.0f)
    return 3;
  if (fmod(5.0L, 2.0L) != 1.0L)
    return 4;
  if (remainder(4.0f, 2.0f) != 0.0f)
    return 5;
  return copysign(2.0L, -1.0L) == -2.0L ? 0 : 6;
}
`
	st := runGCCExecFixture(t, "tgmath-binary-real-helpers.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestTgmathFMAExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <tgmath.h>

int main(void)
{
  if (fma(2.0f, 3.0f, 4.0f) != 10.0f)
    return 1;
  return fma(2.0L, 3.0L, -4.0L) == 2.0L ? 0 : 2;
}
`
	st := runGCCExecFixture(t, "tgmath-fma-real.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestTgmathNextafterExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <tgmath.h>

int main(void)
{
  if (nextafter(1.0f, 1.0f) != 1.0f)
    return 1;
  return nextafter(1.0L, 1.0L) == 1.0L ? 0 : 2;
}
`
	st := runGCCExecFixture(t, "tgmath-nextafter-real.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestTgmathNexttowardExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <tgmath.h>

int main(void)
{
  return nexttoward(1.0f, 1.0L) == 1.0f ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "tgmath-nexttoward-real.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestTgmathScalbnHelpersExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <tgmath.h>

int main(void)
{
  if (scalbn(2.0f, 3) != 16.0f)
    return 1;
  return scalbln(3.0L, 2L) == 12.0L ? 0 : 2;
}
`
	st := runGCCExecFixture(t, "tgmath-scalbn-real.c", source)
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

func TestTgmathComplexAtanhExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <tgmath.h>

int main(void)
{
  complex double z = __builtin_complex(0.0, 0.0);
  complex double r = atanh(z);
  return __builtin_cabs(r) == 0.0 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "tgmath-complex-atanh.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestTgmathComplexAcoshExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <tgmath.h>

int main(void)
{
  complex double z = __builtin_complex(1.0, 0.0);
  complex double r = acosh(z);
  return __builtin_cabs(r) == 0.0 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "tgmath-complex-acosh.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestTgmathComplexAsinhExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <tgmath.h>

int main(void)
{
  complex double z = __builtin_complex(0.0, 0.0);
  complex double r = asinh(z);
  return __builtin_cabs(r) == 0.0 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "tgmath-complex-asinh.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestTgmathComplexAtanExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <tgmath.h>

int main(void)
{
  complex double z = __builtin_complex(0.0, 0.0);
  complex double r = atan(z);
  return __builtin_cabs(r) == 0.0 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "tgmath-complex-atan.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestTgmathComplexAcosExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <tgmath.h>

int main(void)
{
  complex double z = __builtin_complex(1.0, 0.0);
  complex double r = acos(z);
  return __builtin_cabs(r) == 0.0 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "tgmath-complex-acos.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestTgmathComplexAsinExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <tgmath.h>

int main(void)
{
  complex double z = __builtin_complex(0.0, 0.0);
  complex double r = asin(z);
  return __builtin_cabs(r) == 0.0 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "tgmath-complex-asin.c", source)
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

func TestGCCBuiltinVaStartEndExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <stdarg.h>

int first(int n, ...)
{
  va_list ap;
  va_start(ap, n);
  va_end(ap);
  return n;
}

int main(void)
{
  return first(7, 9) == 7 ? 0 : 1;
}
`
	st := runGCCExecFixture(t, "builtin-va-start-end-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestGCCVFormatNoArgumentExecutesThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <stdarg.h>
#include <stdio.h>

int fill(int n, ...)
{
  char buf[16];
  va_list ap = 0;
  va_start(ap, n);
  int a = vsprintf(buf, "ok%%", ap);
  int b = vsnprintf(buf + 3, 4, "xy", ap);
  va_end(ap);

  if (a != 3 || b != 2)
    return 1;
  if (buf[0] != 'o' || buf[1] != 'k' || buf[2] != '%')
    return 2;
  if (buf[3] != 'x' || buf[4] != 'y' || buf[5] != 0)
    return 3;
  return 0;
}

int main(void)
{
  return fill(1, 2);
}
`
	st := runGCCExecFixture(t, "vformat-no-argument-runtime.c", source)
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}

func TestStdioVFormatUnlockedAliasesExecuteThroughRuntime(t *testing.T) {
	source := `/* { dg-do run } */
#include <stdarg.h>
#include <stdio.h>

int emit(int n, ...)
{
  va_list ap = 0;
  va_start(ap, n);
  int a = vprintf_unlocked("vu", ap);
  int b = vfprintf_unlocked(stderr, "vf", ap);
  va_end(ap);
  return a == 2 && b == 2 ? 0 : 1;
}

int main(void)
{
  return emit(1, 2);
}
`
	st := runGCCExecFixture(t, "stdio-vformat-unlocked-aliases-runtime.c", source)
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
