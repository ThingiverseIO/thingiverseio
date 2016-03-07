all: library-archive test

test:
	mkdir -p _test
	gcc shared_library/test.c build/archive/tvio.a -Ibuild/archive -lpthread -lzmq -o _test/test
	./_test/test
	rm -rf _test

test-shared:
	mkdir -p _test
	gcc shared_library/test_shared.c -Ishared_library -Lbuild -lpthread -lzmq -lthingiverse -o _test/test
	./_test/test
	rm -rf _test

library-archive:
	go build --buildmode="c-archive" -o build/archive/tvio.a shared_library/input.go shared_library/output.go shared_library/main.go

library-shared:
	gcc -c shared_library/thingiverseio.c -Lbuild/archive -l:tvio.a -Ibuild/archive -fPIC -lpthread -lzmq -o build/tvio.o
	gcc -shared -fPIC -o build/libthingiverse.so build/tvio.o

tool:
	go build tool/main.go -o bin/tvio

clean:
	rm -rf build/
