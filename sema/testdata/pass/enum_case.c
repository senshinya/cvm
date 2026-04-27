enum Color { RED, GREEN, BLUE };
int main(void) {
    int c = GREEN;
    switch (c) {
        case RED: return 0;
        case GREEN: return 1;
        case BLUE: return 2;
    }
    return -1;
}
