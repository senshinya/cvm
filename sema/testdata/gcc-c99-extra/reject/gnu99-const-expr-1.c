int n;

void
f (void)
{
  int i = 0;
  int a[n];
  enum e1 {


    E1 = (1 ? 0 : ({ 0; })),



    E2 = __real__ (1 ? 0 : i++),
    E3 = __real__ 0,
    E4 = __imag__ (1 ? 0 : i++),
    E5 = __imag__ 0,

    E6 = __alignof__ (int[n]),
    E7 = __alignof__ (a),

    E8 = __extension__ (1 ? 0 : i++),
    E9 = __extension__ 0,


    E10 = (1 ? : i++),

    E11 = (1 ? : 0)
  };
  enum e2 {


    F1 = (int) (_Complex int) 2i,


    F2 = (int) +2i,

    F3 = (int) (1 + 2i),

    F4 = (int) 2i
  };
  static double dr = __real__ (1.0 + 2.0i);

  static double di = __imag__ (1.0 + 2.0i);



  static int j = (1 ? 0 : ({ 0; }));

}
