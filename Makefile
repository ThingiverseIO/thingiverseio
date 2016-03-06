all: library-archive library-archive test

test:
	mkdir -p _test
	gcc -Wall shared_library/test.c build/archive/libthingiverseio.a -Ibuild/archive -lpthread -lzmq -o _test/test
	./_test/test
	rm -rf _test

library-archive:
	go build --buildmode="c-archive" -o build/archive/tvio.a shared_library/input.go shared_library/output.go shared_library/main.go

library-shared:
	go build --buildmode="c-shared" -o build/shared/libthingiverseio.so shared_library/input.go shared_library/output.go shared_library/main.go
	mkdir -p build/include
	mv build/shared/libthingiverseio.h build/include/

tvio-cfg:
	go build  tvio-cfg/main.go

clean:
	rm -rf build/
