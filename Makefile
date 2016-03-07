all: library-archive test

test:
	mkdir -p _test
	gcc shared_library/test.c build/archive/tvio.a -Ibuild/archive -lpthread -lzmq -o _test/test
	./_test/test
	rm -rf _test

library-archive:
	go build --buildmode="c-archive" -o build/archive/tvio.a shared_library/input.go shared_library/output.go shared_library/main.go

library-shared:
	gcc -fPIC -shared -Lbuild/archive -Ibuild/archive shared_library/thingiverseio.c -l:tvio.a -o build/thingiverseio.so

tool:
	go build tool/main.go -o bin/tvio

clean:
	rm -rf build/
