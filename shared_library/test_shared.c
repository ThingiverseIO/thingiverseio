#include<stdio.h>
#include "thingiverseio.h"

char * const DESCRIPTOR = "\
functions:\n\
  - name: SayHello\n\
    input:\n\
      - name: Greeting\n\
        type: string\n\
    output:\n\
      - name: Answer\n\
        type: string\n\
";

int main() {

	printf("Testing Input Creation...\n");

	int input = tvio_new_input(DESCRIPTOR);

	if (input == -1) {
		printf("FAIL\n");
		return 1;
	};
	printf("SUCCES\n");
}
