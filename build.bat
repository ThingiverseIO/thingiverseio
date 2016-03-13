
go build -a --buildmode="c-archive" -o tvio.a shared_library/input.go shared_library/output.go shared_library/main.go

gcc -c shared_library/thingiverseio.c -L. -I. -fPIC -lpthread -lzmq -l:tvio.a -o tvio.o
mkdir build\lib
gcc -shared -fPIC tvio.o tvio.a -o build/lib/libthingiverseio.dll -lpthread -lzmq -lws2_32 -lntdll
rm tvio.*
mkdir build\include
cp shared_library/thingiverseio.h build/include/
mkdir _test
gcc shared_library/test_shared.c -Ibuild/include -Lbuild/lib -lpthread -lzmq -lthingiverseio -o _test/test.exe
cp build/lib/libthingiverseio.dll _test/
.\_test\test.exe


