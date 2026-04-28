extern int printf (const char *, ...);
struct foo { int i; };

int
main ()
{
  struct foo f = (struct foo_typo) { };
  printf ("%d\n", f.i);
  return 0;
}
