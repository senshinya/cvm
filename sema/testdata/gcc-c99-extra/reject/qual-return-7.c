void test_local (void)
{
  auto int foo ();

  const int foo () { return 0; }

  auto void bar (void);
  volatile void bar () { }

  auto volatile void baz (void);
  void baz () { }
}
