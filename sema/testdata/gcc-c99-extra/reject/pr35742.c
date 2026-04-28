void
foo ()
{
  for (;;)
    ({break;})();
  for (;;)
    ({continue;})();
}
