struct A {
  int B;
  short C[2];
};
int a[10] = { 10, [4] = 15 };
struct A b = { .B = 2 };
struct A c[] = { [3].C[1] = 1 };
struct A d[] = { [4 ... 6].C[0 ... 1] = 2 };
int e[] = { [2] 2 };
struct A f = { C: { 0, 1 } };
int g;

void foo (int *);

void bar (void)
{
  int x[] = { g++, 2 };

  foo (x);
}
