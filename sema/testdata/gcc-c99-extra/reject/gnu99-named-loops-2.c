void
foo (int x)
{
 label1:
  for (int i = 0; i < 16; ++i)
   another_label1:
    for (int j = 0; j < 16; ++j)
      break label2;
  for (int i = 0; i < 16; ++i)
    break label3;
 label4:
  switch (x)
    {
    case 0:
      for (int i = 0; i < 16; ++i)
	continue label5;
      break label4;
    case 1:
      for (int i = 0; i < 16; ++i)
	continue label4;
    }
 label6:
  for (int i = 0; i < 16; ++i)
    continue label7;
 label2:
  for (int i = 0; i < 16; ++i)
    ;
 label8:;
  for (int i = 0; i < 16; ++i)
    break label8;
 label9:;
  for (int i = 0; i < 16; ++i)
    continue label9;
 label10:
  ;
  switch (x)
    {
    case 0:
      break label10;
    }
}
