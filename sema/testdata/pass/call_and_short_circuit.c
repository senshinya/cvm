int inc(int x) {
	return x + 1;
}

int choose(int a, int b) {
	return (a && inc(b)) || (!a && b);
}
