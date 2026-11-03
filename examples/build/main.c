#include <stdio.h>
#include "asset.c"
int main(void) {
    fwrite(greeting, 1, greeting_size, stdout);
    return 0;
}
