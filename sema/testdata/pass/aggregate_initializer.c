struct pair {
	int left;
	int right;
};

int sum_pair(void) {
	struct pair p = { .left = 3, .right = 4 };
	return p.left + p.right;
}
