static int f0(void);
void g0(void) { __alignof__(f0()); }


static int f1(void);
void g1(void) { __typeof__(f1()) x; }


static int f2(void);
void g2(void) { __typeof__(int [f2()]) x; }


static int f3(void);
void g3(void) { __typeof__(int (*)[f3()]) x; }


static int f4(void);
void g4(void) { sizeof(__typeof__(int (*)[f3()])); }
