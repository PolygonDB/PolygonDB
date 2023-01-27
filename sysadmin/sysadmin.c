#define _GNU_SOURCE
#define __USE_GNU
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/stat.h>
#include <sys/types.h>
#include <unistd.h>


char* help(){
    char* path = (char*) malloc(sizeof(char) * 256);
    sprintf(path, "\n====Polygon Terminal====\nhelp\t\t\t\t\t\tThis displays all the possible executable lines for Polygon\ncreate_database (name) (password)\t\ttest\n========================\n");
    return path;
}

void datacreate(char *name, char *pass) {

    char path[50];
    sprintf(path, "databases/%s", name);
    mkdir(path, 0777);

    char conpath[50];
    sprintf(conpath, "databases/%s/config.json", name);
    char cinput[50];
    sprintf(cinput, "{\n\t\"key\": \"%s\"\n}", pass);
    FILE* cfile = fopen(conpath, "w");
    fprintf(cfile, cinput);


    char datapath[50];
    sprintf(datapath, "databases/%s/database.json", name);
    FILE* dfile = fopen(datapath, "w");
    fprintf(dfile, "{\n\t\"Example\": \"Hello world\"\n}");

    fclose(cfile);
    fclose(dfile);
}

char* term() {
    char* result;

    char input[256], arg1[256] = "", arg2[256] = "";
    char input_line[256];
    fgets(input_line, sizeof(input_line), stdin);
    sscanf(input_line, "%s %s %s", input, arg1, arg2);

    if (strcmp(input, "create_database") == 0) {
        if (strlen(arg1) > 0 && strlen(arg2) > 0) {
            datacreate(arg1, arg2);
        } else {
            printf("Database cannot be created. Proper command line: create_database name password \n");
        }
    } else if (strcmp(input, "help") == 0) {
        result = help();
    } 

    return result;
}
