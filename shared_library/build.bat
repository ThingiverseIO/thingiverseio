del /F /Q bin

go build -a --buildmode="c-archive" -o tvio.a src/input.go src/output.go src/error.go src/main.go

mv lib/tvio.h include/

gcc -c src/thingiverseio.c -L. -I. -fPIC -lpthread -lzmq -l:tvio.a -o tvio.o

mkdir bin

gcc -shared -fPIC tvio.o tvio.a -o bin/libthingiverseio.dll -lpthread -lzmq -lws2_32 -lntdll -Wl,--output-def,testdll.def,--out-implib,libtestdll.a

del /F /Q tvio.*


