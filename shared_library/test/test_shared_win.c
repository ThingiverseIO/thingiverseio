#include<stdio.h>
#include "thingiverseio.h"
#include <windows.h>

  char * const DESCRIPTOR = "func SayHello(Greeting string) (Answer string)  ";

  int main() {

    printf("Testing Input Creation...\n");

    int input = tvio_new_input(DESCRIPTOR);

	if (input == -1) {
		printf("FAIL\n");
		return 1;
	};
	printf("SUCCES\n");

	printf("Testing Output Creation...\n");

	int output = tvio_new_output(DESCRIPTOR);

	if (output == -1) {
		printf("FAIL\n");
		return 1;
	};

	sleep(5);
	int is;
	int err = tvio_connected(input, &is);
	if (err != 0) {
		printf("FAIL, err not 0\n");
		return 1;
	};
	if (is != 1) {
		printf("FAIL, input did not connect \n");
		return 1;
	};

	printf("SUCCES\n");

	printf("Testing Call...\n");

	char * uuid;
	int uuid_size;

	char * fun = "Hello";
	char * params = "HELLO";
	int params_size = 5;

	err = tvio_call(input, fun,params,params_size, &uuid, &uuid_size);
	if (err != 0) {
		printf("FAIL, err not 0\n");
		return 1;
	};
	if (uuid_size != 36) {
		printf("FAIL, uuid_size is %d, want 36\n");
		return 1;
	};

	sleep(5);

	char * req_uuid;
	int req_uuid_size;
	err = tvio_get_next_request_id(output, &req_uuid, &req_uuid_size);
	if (err != 0) {
		printf("FAIL, err not 0\n");
		return 1;
	};
	if (req_uuid_size != 36) {
		printf("FAIL, req_uuid_size is %d, want 36\n", req_uuid_size);
		return 1;
	};

	char * rfun;
	int rfun_size;
	err = tvio_retrieve_request_function(output, uuid, &rfun, &rfun_size);
	if (err != 0) {
		printf("FAIL, err not 0\n");
		return 1;
	};
	if (rfun_size == 0) {
		printf("FAIL, fun_size is 0\n");
		return 1;
	};
	char * rparams;
	int rparams_size;
	err = tvio_retrieve_request_params(output, uuid, &rparams, &rparams_size);
	if (err != 0) {
		printf("FAIL, err not 0\n");
		return 1;
	};
	if (rparams_size != 5) {
		printf("FAIL, rparams_size is 0\n");
		return 1;
	};

	char * resparams = "HELLO_BACK";
	int resparams_size = 10;

	err = tvio_reply(output, uuid, resparams, resparams_size);
	if (err != 0) {
		printf("FAIL, err not 0\n");
		return 1;
	};

	Sleep(5);

	int ready;
	err = tvio_result_ready(input, uuid, &ready);
	if (err != 0) {
		printf("FAIL, err not 0\n");
		return 1;
	};
	if (ready != 1) {
		printf("FAIL, result hasnt arrived\n");
		return 1;
	}

	char * resultparams;
	int resultparams_size;
	err = tvio_retrieve_result_params(input, uuid, &resultparams, &resultparams_size);
	if (err != 0) {
		printf("FAIL, err not 0\n");
		return 1;
	};
	if (resultparams_size != 10) {
		printf("FAIL, rparams_size is 0\n");
		return 1;
	};

	printf("SUCCES\n");

	printf("Testing Trigger...\n");

	err = tvio_start_listen(input, fun);
	if (err != 0) {
		printf("FAIL, tvio_start_listen err not 0\n");
		return 1;
	};

	Sleep(5);

	err = tvio_trigger(input, fun,params,params_size);
	if (err != 0) {
		printf("FAIL, tvio_trigger err not 0\n");
		return 1;
	};
	Sleep(5);

	err = tvio_get_next_request_id(output, &req_uuid, &req_uuid_size);
	if (err != 0) {
		printf("FAIL, get_gext_req err not 0\n");
		return 1;
	};
	if (req_uuid_size == 0) {
		printf("FAIL, req_uuid_size is 0\n");
		return 1;
	};

	err = tvio_reply(output, req_uuid, resparams, resparams_size);
	if (err != 0) {
		printf("FAIL, tvio_reply err not 0, is \n");
		return 1;
	};

	Sleep(5*1000);

	err = tvio_listen_result_available(input, &ready);
	if (err != 0) {
		printf("FAIL,tvio_listen_result_available err not 0\n");
		return 1;
	};
	if (ready != 1) {
		printf("FAIL,tvio_listen_result_available result hasnt arrived\n");
		return 1;
	}

	err = tvio_retrieve_listen_result_params(input, &resultparams, &resultparams_size);
	if (err != 0) {
		printf("FAIL,tvio_retrieve_listen_result_params err not 0\n");
		return 1;
	};
	if (resultparams_size != 10) {
		printf("FAIL, rparams_size is 0\n");
		return 1;
	};

	err = tvio_remove_input(input);
	if (err != 0) {
		printf("FAIL, tvio_remove_input err not 0\n");
		return 1;
	};

	err = tvio_remove_output(output);
	if (err != 0) {
		printf("FAIL, tvio_remove_input err not 0\n");
		return 1;
	};
	printf("SUCCES\n");

	return 0;
}
