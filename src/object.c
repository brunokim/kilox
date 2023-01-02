#include <stdio.h>
#include <string.h>

#include "memory.h"
#include "object.h"
#include "table.h"
#include "value.h"
#include "vm.h"

#define ALLOCATE_OBJ(type, objectType) \
    (type*)allocateObject(sizeof(type), objectType)

static Obj *allocateObject(size_t size, ObjType type) {
    Obj *object = (Obj*)reallocate(NULL, 0, size);
    object->type = type;

    // Insert this object in the head of VM's object linked list.
    object->next = vm.objects;
    vm.objects = object;
    return object;
}

static ObjString *allocateString(int length, uint32_t hash) {
    ObjString *string = (ObjString *)allocateObject(
        sizeof(ObjString) + (length + 1) * sizeof(char),
        OBJ_STRING);
    string->length = length;
    string->hash = hash;
    tableSet(&vm.strings, OBJ_VAL(string), NIL_VAL); // Intern the string.
    return string;
}

// FNV-1a hash.
static uint32_t hashString(const char *key, int length) {
    uint32_t hash = 2166136261u;
    for (int i = 0; i < length; i++) {
        hash ^= (uint8_t)key[i];
        hash *= 16777619;
    }
    return hash;
}

ObjString *takeString(char *chars, int length) {
    uint32_t hash = hashString(chars, length); 
    ObjString *interned = tableFindString(&vm.strings, chars, length, hash);
    if (interned != NULL) {
        FREE(char, chars);
        return interned;
    }
    ObjString *string = allocateString(length, hash);
    memcpy(string->chars, chars, length + 1);
    FREE(char, chars);
    return string;
}

ObjString *copyString(const char *chars, int length) {
    uint32_t hash = hashString(chars, length); 
    ObjString *interned = tableFindString(&vm.strings, chars, length, hash);
    if (interned != NULL) {
        return interned;
    }
    ObjString *string = allocateString(length, hash);
    memcpy(string->chars, chars, length);
    string->chars[length] = '\0';
    return string;
}

void printObject(Value value) {
    switch (OBJ_TYPE(value)) {
        case OBJ_STRING:
            printf("%s", AS_CSTRING(value));
            break;
    }
}

