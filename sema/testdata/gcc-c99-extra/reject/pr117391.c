int foo(int n, char (*buf)[*]);
int bar(int n, char (*buf)[n]);

void test()
{
	(1 ? foo : bar)(0);
	(0 ? bar : foo)(0);
	(0 ? foo : bar)(0);
	(1 ? bar : foo)(0);
}
