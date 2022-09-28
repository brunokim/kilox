#include "common.h"
#include "chunk.h"
#include "debug.h"

#include <stdio.h>

int main(int argc, const char *argv[]) {
    printf("hello clox\n");

    Chunk chunk;
    initChunk(&chunk);
    writeChunk(&chunk, OP_RETURN, 123);

    for (int i = 0; i < 260; i++) {
        writeConstant(&chunk, 1.0 * i + 0.5, 123);
    }

    disassembleChunk(&chunk, "test chunk");
    freeChunk(&chunk);

    return 0;
}
