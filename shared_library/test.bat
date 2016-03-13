mkdir _test
gcc test/test_shared.c -Iinclude -Lbin -lpthread -lzmq -lthingiverseio -o _test/test.exe
cp bin/libthingiverseio.dll _test/
.\_test\test.exe