struct s { int a; } sv;
union u { int a; } uv;
int i;
long l;
char c;
void *p;
float fv;

void
f (void)
{
  (int []) p;
  (int ()) p;
  (struct s) sv;
  (union u) uv;
  (struct s) i;
  (union u) i;
  (union u) l;
  (int) sv;
  (int) uv;
  (float) sv;
  (float) uv;
  (_Complex double) sv;
  (_Complex double) uv;
  (void *) sv;
  (void *) uv;
  (_Bool) sv;
  (_Bool) uv;
  (void) sv;
  (const void) uv;
  (void *) c;
  (void *) (char) 1;
  (char) p;
  (char) (void *) 1;
}
