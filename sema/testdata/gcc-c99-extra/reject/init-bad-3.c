void f(void);
void g(void) = f;

void h(a)
     int a = 1;
{
  struct s x = { 0 };


}

char s[1] = "x";
char s1[1] = { "x" };
char t[1] = "xy";
char t1[1] = { "xy" };
char u[1] = { "x", "x" };


int j = { 1 };

int k = { 1, 2 };


int a1[1] = { [1] = 0 };

int a2[1] = { [-1] = 0 };
