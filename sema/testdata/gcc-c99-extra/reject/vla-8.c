int a;
struct s { void (*f)(int (*)[a]); };

static int i;
static int new_i() { i++; return i; }
static int bar1(int a[new_i()][new_i()]);

void foo(int n) {
  extern void bar(int i[n][n]);
  extern int bar1(int a[new_i()][new_i()]);
}

void foo1(int n) {
  goto A;
  void bar(int i[n][n]);
  int bar1(int a[new_i()][new_i()]);
 A:
  ;
}

void foo2(int n) {
  goto A;
  int (*(*bar2)(void))[n];
 A:
  ;
}
