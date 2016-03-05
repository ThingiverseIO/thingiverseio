all: library-shared library-archive test

test:
	mkdir _test
	gcc -Wall lib/test.c build/archive/libthingiverseio.a -Ibuild/archive -lpthread -lzmq -o _test/test
	./_test/test
	rm -rf _test

library-archive:
	go build --buildmode="c-archive" -o build/archive/libthingiverseio.a lib/input.go lib/output.go lib/main.go

library-shared:
	go build --buildmode="c-shared" -o build/shared/libthingiverseio.so lib/input.go lib/output.go lib/main.go
	mkdir -p build/include
	mv build/shared/libthingiverseio.h build/include/

tvio-cfg:
	go build  tvio-cfg/main.go

clean:
	rm -rf build/
