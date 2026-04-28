struct foo
{
  int foo;
  int bar;
};

union u
{
  int color;
  int shape;
};



struct foo old_style_f = {
 foa: 1,






 this_does_not_match: 3





};

union u old_style_u = { colour: 3 };








struct foo c99_style_f = {
  .foa = 1,






  .this_does_not_match = 3




};

union u c99_style_u = { .colour=3 };
