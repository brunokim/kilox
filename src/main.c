#include "common.h"
#include "chunk.h"
#include "debug.h"

#include <stdio.h>

int main(int argc, const char *argv[]) {
    printf("hello clox\n");

    Chunk chunk;
    initChunk(&chunk);
    writeChunk(&chunk, OP_RETURN, 123);

    int index = addConstant(&chunk, 1.2);
    writeChunk(&chunk, OP_CONSTANT, 123);
    writeChunk(&chunk, index, 123);

    disassembleChunk(&chunk, "test chunk");
    freeChunk(&chunk);

    return 0;
}
