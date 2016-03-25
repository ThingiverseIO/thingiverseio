all: libthingiverseio.so

.PHONY: all test clean

test:
	mkdir -p _test
	gcc test/test_shared.c -Iinclude -Lbin -lpthread -lzmq -lthingiverseio -ltvio -o _test/test
	cp bin/lib* ./
	./_test/test
	rm -rf _test
	rm -f libtvio.so
	rm -f libthingiverseio.so

tvio.so:
	go build --buildmode="c-shared" -o bin/tvio.so src/input.go src/output.go src/main.go
	mv bin/tvio.h include/
	mv bin/tvio.so bin/libtvio.so
tvio.o: tvio.so
	gcc -c src/thingiverseio.c -L./bin/ -I./include/ -fPIC -lpthread -lzmq -ltvio -o tvio.o

libthingiverseio.so: tvio.o
	gcc -shared -fPIC tvio.o -o bin/libthingiverseio.so -lpthread -lzmq -L./bin -ltvio

install:
	cp bin/lib* /usr/lib/
	cp include/* /usr/include/
	ldconfig

clean:
	rm -f tvio.o
	rm -rf bin
	rm -f include/tvio.h
