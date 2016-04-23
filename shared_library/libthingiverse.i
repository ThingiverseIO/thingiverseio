
%module libthingiverseio
%include "cpointer.i"
%include "cstring.i"

/* Create some functions for working with "int *" */
%pointer_functions(int, intp);

%cstring_output_allocate_size(char **result, int *result_size, free(*$1));
%cstring_output_allocate_size(char **id, int *id_size, free(*$1));
%cstring_output_allocate_size(char **function, int *function_size, free(*$1));
%cstring_output_allocate_size(void **params, int *params_size, free(*$1));
%typemap(in) void* = char*;


%{
#define SWIG_FILE_WITH_INIT
#include "thingiverseio.h"
%}

%include "thingiverseio.h"
