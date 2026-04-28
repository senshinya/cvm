struct s { char c[1]; } x;
struct s f (void) { return x; }

void
g (void)
{
  char c[1];
  c = f ().c;
}

void
h (void)
{
  char c[1] = f ().c;
}
