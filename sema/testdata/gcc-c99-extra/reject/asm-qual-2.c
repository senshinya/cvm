void
f (void)
{
  asm volatile goto ("" :::: lab);
  asm volatile inline ("" :::);
  asm inline volatile ("" :::);
  asm inline goto ("" :::: lab);
  asm goto volatile ("" :::: lab);
  asm goto inline ("" :::: lab);

  asm volatile inline goto ("" :::: lab);
  asm volatile goto inline ("" :::: lab);
  asm inline volatile goto ("" :::: lab);
  asm inline goto volatile ("" :::: lab);
  asm goto volatile inline ("" :::: lab);
  asm goto inline volatile ("" :::: lab);


  asm goto volatile volatile ("" :::: lab);
  asm volatile goto volatile ("" :::: lab);
  asm volatile volatile goto ("" :::: lab);
  asm goto goto volatile ("" :::: lab);
  asm goto volatile goto ("" :::: lab);
  asm volatile goto goto ("" :::: lab);

  asm inline volatile volatile ("" :::);
  asm volatile inline volatile ("" :::);
  asm volatile volatile inline ("" :::);
  asm inline inline volatile ("" :::);
  asm inline volatile inline ("" :::);
  asm volatile inline inline ("" :::);

  asm goto inline inline ("" :::: lab);
  asm inline goto inline ("" :::: lab);
  asm inline inline goto ("" :::: lab);
  asm goto goto inline ("" :::: lab);
  asm goto inline goto ("" :::: lab);
  asm inline goto goto ("" :::: lab);

lab:
  ;
}
