all: library-archive test

test:
	mkdir -p _test
	gcc shared_library/test.c build/archive/tvio.a -Ibuild/archive -lpthread -lzmq -o _test/test
	./_test/test
	rm -rf _test

test-shared:
	mkdir -p _test
	gcc shared_library/test_shared.c -Ibuild/archive -Ishared_library -Lbuild -lpthread -lzmq -lthingiverse -o _test/test
	./_test/test
	rm -rf _test

library-archive:
	go build --buildmode="c-archive" -o build/archive/tvio.a shared_library/input.go shared_library/output.go shared_library/main.go

library-shared:
	gcc -c shared_library/thingiverseio.c -Lbuild/archive -Ibuild/archive -fPIC -lpthread -lzmq -l:tvio.a -o build/tvio.o
	gcc -shared -fPIC build/tvio.o build/archive/tvio.a -o build/libthingiverse.dll -lpthread -lzmq -lws2_32 -lntdll

tool:
	go build tool/main.go -o bin/tvio

clean:
	rm -rf build/
