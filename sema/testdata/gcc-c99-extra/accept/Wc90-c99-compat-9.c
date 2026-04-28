extern void bar (int);

void
foo (int n)
{
  for (int i = 0; i < n; i++)
    bar (i);
}
