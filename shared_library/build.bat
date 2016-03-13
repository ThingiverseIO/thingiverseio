rm -f bin

go build -a --buildmode="c-archive" -o lib/tvio.a src/input.go src/output.go src/main.go

gcc -c src/thingiverseio.c -L./lib -I./lib -I./include -fPIC -lpthread -lzmq -l:tvio.a -o tvio.o

mkdir bin

gcc -shared -fPIC tvio.o lib/tvio.a -o bin/libthingiverseio.dll -lpthread -lzmq -lws2_32 -lntdll

rm tvio.*


