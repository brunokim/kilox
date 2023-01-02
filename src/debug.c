#include <stdio.h>

#include "debug.h"
#include "value.h"

void disassembleChunk(Chunk *chunk, const char *name) {
    printf("== %s ==\n", name);
    for (int offset = 0; offset < chunk->count; ) {
        offset = disassembleInstruction(chunk, offset);
    }
}

static void printConstant(const char *name, Chunk *chunk, int index) {
    printf("%-16s %4d '", name, index);
    printValue(chunk->constants.values[index]);
    printf("'\n");
}

static int constantInstruction(const char *name, Chunk *chunk, int offset) {
    uint8_t index = chunk->code[offset + 1];
    printConstant(name, chunk, index);
    return offset + 2;
}

static int constantLongInstruction(const char *name, Chunk *chunk, int offset) {
    uint8_t b0 = chunk->code[offset + 1];
    uint8_t b1 = chunk->code[offset + 2];
    uint8_t b2 = chunk->code[offset + 3];
    int index = (b0 << 0) | (b1 << 8) | (b2 << 16);
    printConstant(name, chunk, index);
    return offset + 4;
}

static int simpleInstruction(const char *name, int offset) {
    printf("%s\n", name);
    return offset + 1;
}

int disassembleInstruction(Chunk *chunk, int offset) {
    // Instruction offset in chunk.
    printf("%04d ", offset);

    // Line information.
    if (offset > 0 && chunk->lines[offset] == chunk->lines[offset - 1]) {
        printf("   | ");
    } else {
        printf("%4d ", chunk->lines[offset]);
    }

    // Instruction
    uint8_t instruction = chunk->code[offset];
    switch (instruction) {
    case OP_CONSTANT:
        return constantInstruction("OP_CONSTANT", chunk, offset);
    case OP_CONSTANT_LONG:
        return constantLongInstruction("OP_CONSTANT_LONG", chunk, offset);
    case OP_FALSE:
        return simpleInstruction("OP_FALSE", offset);
    case OP_TRUE:
        return simpleInstruction("OP_TRUE", offset);
    case OP_NIL:
        return simpleInstruction("OP_NIL", offset);
    case OP_ADD:
        return simpleInstruction("OP_ADD", offset);
    case OP_SUBTRACT:
        return simpleInstruction("OP_SUBTRACT", offset);
    case OP_MULTIPLY:
        return simpleInstruction("OP_MULTIPLY", offset);
    case OP_DIVIDE:
        return simpleInstruction("OP_DIVIDE", offset);
    case OP_NOT:
        return simpleInstruction("OP_NOT", offset);
    case OP_NEGATE:
        return simpleInstruction("OP_NEGATE", offset);
    case OP_RETURN:
        return simpleInstruction("OP_RETURN", offset);
    default:
        printf("Unknown opcode %d\n", instruction);
        return offset + 1;
    }
}
