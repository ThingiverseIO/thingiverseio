
#ifdef __cplusplus
extern "C" {
#endif


extern int new_input(char* p0);

extern int remove_input(int p0);

extern int connected(int p0, int* p1);

extern int start_listen(int p0, char* p1);

extern int stop_listen(int p0, char* p1);

extern int call(int p0, char* p1, void* p2, int p3, char** p4, int* p5);

extern int call_all(int p0, char* p1, void* p2, int p3, char** p4, int* p5);

extern int trigger(int p0, char* p1, void* p2, int p3);

extern int trigger_all(int p0, char* p1, void* p2, int p3);

extern int result_ready(int p0, char* p1, int* p2);

extern int retrieve_result_params(int p0, char* p1, void** p2, int* p3);

extern int listen_result_available(int p0, int* p1);

extern int retrieve_listen_result_id(int p0, char** p1, int* p2);

extern int retrieve_listen_result_function(int p0, char** p1, int* p2);

extern int retrieve_listen_result_params(int p0, void** p1, int* p2);

extern int retrieve_next_call_all_result_params(int p0, char* p1, void** p2, int* p3);

extern int new_output(char* p0);

extern int remove_output(int p0);

extern int get_next_request_id(int p0, char** p1, int* p2);

extern int request_available(int p0, int* p1);

extern int retrieve_request_function(int p0, char* p1, char** p2, int* p3);

extern int retrieve_request_params(int p0, char* p1, void** p2, int* p3);

extern int reply(int p0, char* p1, void* p2, int p3);

extern int emit(int p0, char* p1, void* p2, int p3, void* p4, int p5);

extern int version(int* p0, int* p1, int* p2);

extern void check_descriptor(char* p0, char** p1, int* p2);

#ifdef __cplusplus
}
#endif
