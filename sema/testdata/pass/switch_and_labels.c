int route(int x) {
	int y = 0;
start:
	switch (x) {
	case 0:
		y = 10;
		break;
	case 1:
		x = x - 1;
		goto start;
	default:
		y = x;
	}
	return y;
}
