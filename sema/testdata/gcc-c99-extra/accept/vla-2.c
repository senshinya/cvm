static inline void A (int i)
{
  struct S { int ar[1][i]; } s;

  s.ar[0][0] = 0;
}

void B(void)
{
  A(23);
}

static inline void C (int i)
{
  union U { int ar[1][i]; } u;

  u.ar[0][0] = 0;
}

void D(void)
{
  C(23);
}
