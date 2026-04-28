__complex__ unsigned int i;
int j;
char k;
__complex__ double l;
double m;
float n;

void
foo ()
{
  ((__complex__ int)i)();
  ((__complex__ int)j)();
  ((__complex__ int)k)();
  ((__complex__ long double)l)();
  ((__complex__ long double)m)();
  ((__complex__ long double)n)();
}
