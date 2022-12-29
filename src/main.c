#include "common.h"
#include "chunk.h"
#include "debug.h"
#include "vm.h"

#include <stdio.h>

int main(int argc, const char *argv[]) {
    initVM();
    printf("hello clox\n");

    Chunk chunk;
    initChunk(&chunk);

    for (int i = 0; i < 260; i++) {
        writeConstant(&chunk, 1.0 * i + 0.5, 123);
    }
    writeChunk(&chunk, OP_RETURN, 123);

    disassembleChunk(&chunk, "test chunk");
    interpret(&chunk);
    freeVM();
    freeChunk(&chunk);

    return 0;
}
