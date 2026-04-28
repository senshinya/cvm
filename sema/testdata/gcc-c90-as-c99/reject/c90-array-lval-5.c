struct s { char c[1]; };

extern struct s foo (void);
struct s a, b, c;
int d;

void
bar (void)
{
  &((foo ()).c);
  &((d ? b : c).c);
  &((d, b).c);
  &((a = b).c);
}
