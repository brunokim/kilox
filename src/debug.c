#include <stdio.h>

#include "debug.h"
#include "value.h"

static const char *op_code_names[NUM_OP_CODES] = {
    #define X(name) [name] = #name,
        OP_CODES
    #undef X
};

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
    case OP_FALSE:
    case OP_TRUE:
    case OP_NIL:
    case OP_EQUAL:
    case OP_GREATER:
    case OP_LESS:
    case OP_ADD:
    case OP_SUBTRACT:
    case OP_MULTIPLY:
    case OP_DIVIDE:
    case OP_NOT:
    case OP_NEGATE:
    case OP_PRINT:
    case OP_POP:
    case OP_RETURN:
        return simpleInstruction(op_code_names[instruction], offset);
    case OP_CONSTANT:
    case OP_GET_GLOBAL:
    case OP_DEFINE_GLOBAL:
        return constantInstruction(op_code_names[instruction], chunk, offset);
    case OP_CONSTANT_LONG:
        return constantLongInstruction(op_code_names[instruction], chunk, offset);
    default:
        printf("Unknown opcode %d\n", instruction);
        return offset + 1;
    }
}
