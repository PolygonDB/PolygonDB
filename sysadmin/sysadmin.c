#define _GNU_SOURCE
#define __USE_GNU
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/stat.h>
#include <sys/types.h>
#include <unistd.h>


void help(){
    printf("====Polygon Terminal====\n");
    printf("help\t\t\t\t\t\tThis displays all the possible executable lines for Polygon\n");
    printf("create_database (name) (password)\t\tThis will create a database for you with name and password\n");
    printf("========================\n");
}

void datacreate(char *name, char *pass) {

    char path[50];
    sprintf(path, "databases/%s", name);

    #ifdef _WIN32
        // Code for Windows
        mkdir(path);
    #elif defined __linux__
        // Code for Linux
        mkdir(path, 0777);
    #elif defined __APPLE__
        // Code for MacOS
        mkdir(path, 0777);
    #endif

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

    //char* output = (char*) malloc(sizeof(char) * 32);
    printf("File has been created.\n");
}