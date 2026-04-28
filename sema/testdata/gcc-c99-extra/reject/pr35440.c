struct A {};
struct B { int i; char j[2]; };

void foo (void)
{
  (struct A){}();
  (struct B){ .i = 2, .j[1] = 1 }();
}
