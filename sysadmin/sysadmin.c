#include <stdio.h>
#include <stdlib.h>

int main() {
    char *Address;
    char *PolyPass;

    // Get the Address to Polygon Server
    printf("Address to access =>");
    scanf("%s", &Address);
    printf("Password to Polygon Database =>");
    scanf("%s", &PolyPass);

    free(Address);
    free(PolyPass);
}
