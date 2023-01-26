#include <stdio.h>
#include <stdlib.h>

int main() {
    char *Input, *Arg1, *Arg2;

    // Get the Address to Polygon Server
    scanf("%s", &Address);
    printf("Password to Polygon Database =>");
    scanf("%s", &PolyPass);

    reach(Address, Password)
    free(Address);
    free(PolyPass);
}

char reach(char Address, char password) {
    CURL *curl;
    CURLcode res;

    curl = curl_easy_init();
    if(curl) {
        curl_easy_setopt(curl, CURLOPT_URL, "http://example.com");
        curl_easy_setopt(curl, CURLOPT_POSTFIELDS, "password create_database test test");

        res = curl_easy_perform(curl);
        if(res != CURLE_OK)
            fprintf(stderr, "curl_easy_perform() failed: %s\n", curl_easy_strerror(res));

        curl_easy_cleanup(curl);
    }
}