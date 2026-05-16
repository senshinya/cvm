int g = 3 + 4;
int arr[3] = { 1, 2, 3 };
int *gp = &g;

int read_globals(void) {
	return arr[1] + *gp;
}
