#include <stdio.h>
#include <string.h>

#include "memory.h"
#include "object.h"
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

static ObjString *allocateString(int length) {
    ObjString *string = (ObjString *)allocateObject(
        sizeof(ObjString) + (length + 1) * sizeof(char),
        OBJ_STRING);
    string->length = length;
    return string;
}

ObjString *takeString(char *chars, int length) {
    ObjString *string = allocateString(length);
    memcpy(string->chars, chars, length + 1);
    FREE(char, chars);
    return string;
}

ObjString *copyString(const char *chars, int length) {
    ObjString *string = allocateString(length);
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

