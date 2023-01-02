#ifndef clox_chunk_h
#define clox_chunk_h

#include "common.h"
#include "value.h"

#define OP_CODES \
    X(OP_CONSTANT) \
    X(OP_CONSTANT_LONG) \
    X(OP_NIL) \
    X(OP_TRUE) \
    X(OP_FALSE) \
    X(OP_EQUAL) \
    X(OP_GREATER) \
    X(OP_LESS) \
    X(OP_ADD) \
    X(OP_SUBTRACT) \
    X(OP_MULTIPLY) \
    X(OP_DIVIDE) \
    X(OP_NOT) \
    X(OP_NEGATE) \
    X(OP_PRINT) \
    X(OP_POP) \
    X(OP_GET_GLOBAL) \
    X(OP_DEFINE_GLOBAL) \
    X(OP_SET_GLOBAL) \
    X(OP_RETURN) \

typedef enum {
    #define X(name) name,
        OP_CODES
    #undef X
    NUM_OP_CODES,
} OpCode;

typedef struct {
    int count;
    int capacity;
    uint8_t *code;
    int *lines;
    ValueArray constants;
} Chunk;

void initChunk(Chunk *chunk);
void freeChunk(Chunk *chunk);
void writeChunk(Chunk *chunk, uint8_t byte, int line);

int addConstant(Chunk *chunk, Value value);
void writeConstant(Chunk *chunk, Value value, int line);

#endif
