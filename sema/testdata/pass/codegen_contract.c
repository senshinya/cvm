int g;

int sum(int n) {
	int a[n];
	int total = 0;
	for (int i = 0; i < n; i = i + 1) {
		total = total + a[i];
	}
	return total + g;
}
