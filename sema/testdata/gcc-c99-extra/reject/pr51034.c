struct S;

int
main ()
{
  struct R { typeof (((struct W) {})) x; } r;
  struct S { typeof (((struct S) {})) x; } s;
  struct T { int x[sizeof ((struct T) {})]; } t;
  struct U { int x[sizeof((struct V){})];} u;
}
