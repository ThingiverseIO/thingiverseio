#ifdef __cplusplus
extern "C" {
#endif


extern int tvio_new_input(char* descriptor);

extern int tvio_remove_input(int input);

extern int tvio_connected(int input, int* is);

extern int tvio_start_listen(int input, char* function);

extern int tvio_stop_listen(int input, char* function);

extern int tvio_call(int input, char* function, void* params, int params_size, char** id, int* id_size);

extern int tvio_call_all(int input, char* function, void* params, int params_size, char** id, int* id_size);

extern int tvio_trigger(int input, char* function, void* params, int params_size);

extern int tvio_trigger_all(int input, char* function, void* params, int params_size);

extern int tvio_result_ready(int input, char* id, int* ready);

extern int tvio_retrieve_result_params(int input, char* id, void** params, int* params_size);

extern int tvio_listen_result_available(int input, int* is);

extern int tvio_retrieve_listen_result_id(int input, char** id, int* id_size);

extern int tvio_retrieve_listen_result_function(int input, char** function, int* function_size);

extern int tvio_retrieve_listen_result_params(int input, void** params, int* params_size);

extern int tvio_retrieve_next_call_all_result_params(int input, char* id, void** params, int* params_size);

extern int tvio_new_output(char* descriptor);

extern int tvio_remove_output(int output);

extern int tvio_request_available(int output, int* is);

extern int tvio_get_next_request_id(int output, char** id, int* id_size);

extern int tvio_retrieve_request_function(int output, char* id, char** function, int* function_size);

extern int tvio_retrieve_request_params(int output, char* id, void** params, int* params_size);

extern int tvio_reply(int output, char* id, void* params, int params_size);

extern int tvio_emit(int output, char* function, void* in_params, int in_params_size, void* params, int params_size);

#ifdef __cplusplus
}
#endif
