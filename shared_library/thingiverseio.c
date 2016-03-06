#include <tvio.h>

int tvio_new_input(char* descriptor){
	return new_input(descriptor);
}

int tvio_start_listen(int input, char* function){
	return start_listen(input, function);
}

int tvio_stop_listen(int input, char* function){
	return stop_listen(input, function);
}

int tvio_call(int input, char* function, void* params, int params_size, char** id, int* id_size){
	return call(input, function, params, params_size, id, id_size);
}

int tvio_call_all(int input, char* function, void* params, int params_size, char** id, int* id_size){
	return call_all(input, function, params, params_size, id, id_size);
}

int tvio_trigger(int input, char* function, void* params, int params_size) {
	return trigger(input, function, params, params_size);
}

int tvio_trigger_all(int input, char* function, void* params, int params_size){
	return trigger_all(input, function, params, params_size);
}

int tvio_result_ready(int input, char* id, int* ready){
	return result_ready(input, id, ready);
}

int tvio_retrieve_result_params(int input, char* id, void** params, int* params_size){
	return retrieve_result_params(input, id, params, params_size);
}

int tvio_listen_result_available(int input, int* is){
	return listen_result_available(input, is);
}

int tvio_retrieve_listen_result_id(int input, char** id, int* id_size) {
	return retrieve_listen_result_id(input, id, id_size);
}

int tvio_retrieve_listen_result_function(int input, char** function, int* function_size){
	return retrieve_listen_result_function(input, function, function_size);
}

int tvio_retrieve_listen_result_request_params(int input, void** params, int* params_size){
	return retrieve_listen_result_request_params(input, params, params_size);
}

int tvio_retrieve_listen_result_params(int input, void** params, int* params_size){
	return retrieve_listen_result_params(input, params, params_size);
}

int tvio_retrieve_next_call_all_result_params(int input, char* id, void** params, int* params_size){
	return retrieve_next_call_all_result_params(input, id , params, params_size);
}

int tvio_new_output(char* descriptor){
	return new_output(descriptor);
}

int tvio_get_next_request_id(int output, char** id, int* id_size){
	return get_next_request_id(output, id, id_size);
}

int tvio_retrieve_request_function(int output, char* id, char** function, int* function_size){
	return retrieve_request_function(output, id ,function, function_size);
}

int tvio_retrieve_request_params(int output, char* id, void** params, int* params_size){
	return retrieve_request_params(output, id ,params, params_size);
}

int tvio_reply(int output, char* id, void* params, int params_size){
	return reply(output, id , params, params_size);
}

int tvio_emit(int output, char* function, void* in_params, int in_params_size, void* params, int params_size){
	return emit(output, function, in_params, in_params_size, params, params_size);
}

int main(){}

