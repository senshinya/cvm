package main

import (
	"testing"
)

func TestError(t *testing.T) {
	(&Compiler{}).RunSource(`int main()
{
  typedef int a;
  int a;
  return 0;
}`)
}
