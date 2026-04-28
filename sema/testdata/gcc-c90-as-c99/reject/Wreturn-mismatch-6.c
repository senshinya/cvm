void f1 (void);

int
f2 (void)
{
  f1 ();
}

static __inline__ int
f3 (void)
{
  f1 ();
}

void
f4 (void)
{
  return 1;
}

void
f5 (void)
{
  return f1 ();
}

int
f6 (void)
{
  return;
}

int
f7 (void)
{
  return f1 ();
}
