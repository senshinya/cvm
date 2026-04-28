extern int foo1;
extern int bar1(int);

void test1(void)
{
  extern double foo1;
  extern double bar1(double);
}



void test2(void)
{
  extern double foo2;
  extern double bar2(double);
}

extern int foo2;
extern int bar2(int);




typedef float baz3;

void prime3(void)
{
  extern int foo3;
  extern int bar3(int);
  extern int baz3;
}

void test3(void)
{
  extern double foo3;
  extern double bar3(double);
  extern double baz3;
}



void prime4(void)
{
  bar4();

}

void test4(void)
{
  extern double bar4(double);

}



void prime5(void)
{
  extern double bar5(double);
}

void test5(void)
{
  bar5(1);
}



extern int test6(int);
static int test6(int x)
{ return x; }




void prime7(void)
{
  extern int test7(int);
}

static int test7(int x)
{ return x; }



void prime8(void)
{
  test8();

}

static int test8(int x)
{ return x; }
