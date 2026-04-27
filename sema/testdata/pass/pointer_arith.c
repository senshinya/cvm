int a[3];
int main(void) {
    int *p = a;
    int v = *(p + 1);
    return v;
}
