#include <stdarg.h>
#include <stdio.h>
#include <string.h>

#include "chunk.h"
#include "common.h"
#include "compiler.h"
#include "debug.h"
#include "memory.h"
#include "object.h"
#include "value.h"
#include "vm.h"

VM vm;

static void resetStack() {
    freeValueArray(&vm.stack);
}

static void runtimeError(const char *format, ...) {
    va_list args;
    va_start(args, format);
    vfprintf(stderr, format, args);
    va_end(args);
    fputs("\n", stderr);

    size_t instruction = vm.ip - vm.chunk->code - 1;
    int line = vm.chunk->lines[instruction];
    fprintf(stderr, "[line %d] in script\n", line);
    resetStack();
}

void initVM() {
    initValueArray(&vm.stack);
    vm.objects = NULL;
    initTable(&vm.strings);
    initValueArray(&vm.globals);
}

void freeVM() {
    freeValueArray(&vm.globals);
    freeTable(&vm.strings);
    freeObjects();
    freeValueArray(&vm.stack);
}

void push(Value value) {
    writeValueArray(&vm.stack, value);
}

Value pop() {
    vm.stack.count--;
    return vm.stack.values[vm.stack.count];
}

Value peek(int distance) {
    return vm.stack.values[vm.stack.count-1-distance];
}

static bool isFalsey(Value value) {
    return IS_NIL(value) || IS_BOOL(value) && !AS_BOOL(value);
}

static void concatenate() {
    ObjString *b = AS_STRING(pop());
    ObjString *a = AS_STRING(pop());

    int length = a->length + b->length;
    char *chars = ALLOCATE(char, length + 1);
    memcpy(chars, a->chars, a->length);
    memcpy(chars + a->length, b->chars, b->length);
    chars[length] = '\0';

    ObjString *result = takeString(chars, length);
    push(OBJ_VAL(result));
}

static InterpretResult run() {
#define READ_BYTE() (*vm.ip++)
#define READ_UINT24() \
    (vm.ip += 3, \
     (uint32_t)(vm.ip[-3] << 0 | vm.ip[-2] << 8 | vm.ip[-1] << 16))
#define READ_CONSTANT() (vm.chunk->constants.values[READ_UINT24()])
#define READ_STRING() AS_STRING(READ_CONSTANT())
#define BINARY_OP(valueType, op) \
    do { \
        if (!IS_NUMBER(peek(0)) || !IS_NUMBER(peek(1))) { \
            runtimeError("Operands must be numbers."); \
            return INTERPRET_RUNTIME_ERROR; \
        } \
        double b = AS_NUMBER(pop()); \
        double a = AS_NUMBER(pop()); \
        push(valueType(a op b)); \
    } while (false)

    for (;;) {
#ifdef DEBUG_TRACE_EXECUTION
        printf("        ");
        for (int i = 0; i < vm.stack.count; i++) {
            printf("[ ");
            printValue(vm.stack.values[i]);
            printf(" ]");
        }
        printf("\n");
        disassembleInstruction(
            vm.chunk,
            (int)(vm.ip - vm.chunk->code));
#endif
        uint8_t instruction;
        switch (instruction = READ_BYTE()) {
            case OP_CONSTANT: {
                Value constant = READ_CONSTANT();
                push(constant);
                break;
            }
            case OP_FALSE: push(BOOL_VAL(false)); break;
            case OP_TRUE: push(BOOL_VAL(true)); break;
            case OP_NIL: push(NIL_VAL); break;
            case OP_EQUAL: {
                Value b = pop();
                Value a = pop();
                push(BOOL_VAL(valuesEqual(a, b)));
                break;
            }
            case OP_GREATER:
                BINARY_OP(BOOL_VAL, >);
                break;
            case OP_LESS:
                BINARY_OP(BOOL_VAL, <);
                break;
            case OP_ADD:
                if (IS_STRING(peek(0)) && IS_STRING(peek(1))) {
                    concatenate();
                } else if (IS_NUMBER(peek(0)) && IS_NUMBER(peek(1))) {
                    double b = AS_NUMBER(pop());
                    double a = AS_NUMBER(pop());
                    push(NUMBER_VAL(a + b));
                } else {
                    runtimeError(
                        "Operands must be two numbers or two strings.");
                    return INTERPRET_RUNTIME_ERROR;
                }
                break;
            case OP_SUBTRACT:
                BINARY_OP(NUMBER_VAL, -);
                break;
            case OP_MULTIPLY:
                BINARY_OP(NUMBER_VAL, *);
                break;
            case OP_DIVIDE:
                BINARY_OP(NUMBER_VAL, /);
                break;
            case OP_NOT:
                push(BOOL_VAL(isFalsey(pop())));
                break;
            case OP_NEGATE:
                if (!IS_NUMBER(peek(0))) {
                    runtimeError("Operand must be a number.");
                    return INTERPRET_RUNTIME_ERROR;
                }
                push(NUMBER_VAL(-AS_NUMBER(pop())));
                break;
            case OP_PRINT:
                printValue(pop());
                printf("\n");
                break;
            case OP_POP:
                pop();
                break;
            case OP_GET_GLOBAL: {
                uint32_t index = READ_UINT24();
                Value value = vm.globals.values[index];
                if (IS_INVALID(value)) {
                    Value name = vm.chunk->constants.values[index];
                    runtimeError("Undefined variable '%s'.", AS_CSTRING(name));
                    return INTERPRET_RUNTIME_ERROR;
                }
                push(value);
                break;
            }
            case OP_DEFINE_GLOBAL: {
                uint32_t index = READ_UINT24();
                vm.globals.values[index] = peek(0);
                pop();
                break;
            }
            case OP_SET_GLOBAL: {
                uint32_t index = READ_UINT24();
                Value value = vm.globals.values[index];
                if (IS_INVALID(value)) {
                    Value name = vm.chunk->constants.values[index];
                    runtimeError("Undefined variable '%s'.", AS_CSTRING(name));
                    return INTERPRET_RUNTIME_ERROR;
                }
                vm.globals.values[index] = peek(0);
                break;
            }
            case OP_GET_LOCAL: {
                uint32_t slot = READ_UINT24();
                push(vm.stack.values[slot]);
                break;
            }
            case OP_SET_LOCAL: {
                uint32_t slot = READ_UINT24();
                vm.stack.values[slot] = peek(0);
                break;
            }
            case OP_RETURN:
                // Exit interpreter.
                return INTERPRET_OK;
        }
    }

#undef BINARY_OP
#undef READ_STRING
#undef READ_UINT24
#undef READ_CONSTANT
#undef READ_BYTE
}

InterpretResult interpret(const char *source) {
    Chunk chunk;
    initChunk(&chunk);

    if (!compile(source, &chunk)) {
        freeChunk(&chunk);
        return INTERPRET_COMPILE_ERROR;
    }

    vm.chunk = &chunk;
    vm.ip = vm.chunk->code;
    growValueArray(&vm.globals, chunk.constants.count);

    InterpretResult result = run();

    freeChunk(&chunk);
    return result;
}
