#include <stdlib.h>
#include <stdio.h>
#include <string.h>

#include "value.h"
#include "memory.h"
#include "object.h"


void initValueArray(ValueArray *array) {
    array->count = 0;
    array->capacity = 0;
    array->values = NULL;
}

void freeValueArray(ValueArray *array) {
    FREE_ARRAY(Value, array->values, array->capacity);
    initValueArray(array);
}

void growValueArray(ValueArray *array, int capacity) {
    if (array->capacity >= capacity) {
        return;
    }
    int oldCapacity = array->capacity;
    array->capacity = capacity;
    array->values = GROW_ARRAY(Value, array->values, oldCapacity, array->capacity);
    for (int i = oldCapacity; i < array->capacity; i++) {
        array->values[i] = INVALID_VAL;
    }
}

void writeValueArray(ValueArray *array, Value value) {
    if (array->capacity < array->count + 1) {
        growValueArray(array, GROW_CAPACITY(array->capacity));
    }
    array->values[array->count] = value;
    array->count++;
}

void printValue(Value value) {
    switch (value.type) {
        case VAL_INVALID:
            printf("INVALID");
            break;
        case VAL_BOOL:
            printf(AS_BOOL(value) ? "true" : "false");
            break;
        case VAL_NIL:
            printf("nil");
            break;
        case VAL_NUMBER:
            printf("%g", AS_NUMBER(value));
            break;
        case VAL_OBJ:
            printObject(value);
            break;
        default:
            printf("Unknown value type %d", value.type);
    }
}

bool valuesEqual(Value a, Value b) {
    if (a.type != b.type) {
        return false;
    }
    switch (a.type) {
        case VAL_BOOL: return AS_BOOL(a) == AS_BOOL(b);
        case VAL_NIL: return true;
        case VAL_NUMBER: return AS_NUMBER(a) == AS_NUMBER(b);
        case VAL_OBJ: return AS_OBJ(a) == AS_OBJ(b);
        default: return false; //unreachable
    }
}

static uint32_t doubleHash(double x) {
    // TODO: write efficient hash for doubles.
    // See https://github.com/python/cpython/blob/f4c03484da59049eb62a9bf7777b963e2267d187/Python/pyhash.c
    return 42u;
}

static uint32_t objectHash(Obj *obj) {
    switch (obj->type) {
        case OBJ_STRING:
            return ((ObjString*)obj)->hash;
        default: return 1000u; //unreachable
    }
}

uint32_t valueHash(Value x) {
    switch (x.type) {
        case VAL_INVALID: return -2;
        case VAL_NIL: return 0;
        case VAL_BOOL: return AS_BOOL(x) ? 1 : 2;
        case VAL_NUMBER: return doubleHash(AS_NUMBER(x));
        case VAL_OBJ: return objectHash(AS_OBJ(x));
        default: return -1; //unreachable
    }
}
