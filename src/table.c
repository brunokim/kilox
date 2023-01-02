#include <stdlib.h>
#include <string.h>

#include "memory.h"
#include "object.h"
#include "table.h"
#include "value.h"

#define TABLE_MAX_LOAD 0.75

#define IS_UNUSED(entry)    (IS_OBJ((entry)->key) && AS_OBJ((entry)->key) == NULL)
#define IS_EMPTY(entry)     (IS_UNUSED(entry) && IS_NIL((entry)->value))
#define IS_TOMBSTONE(entry) (IS_UNUSED(entry) && IS_BOOL((entry)->value))

void initTable(Table *table) {
    table->count = 0;
    table->capacity = 0;
    table->entries = NULL;
}

void freeTable(Table *table) {
    FREE_ARRAY(Entry, table->entries, table->capacity);
    initTable(table);
}

static Entry *findEntry(Entry *entries, int capacity, Value key) {
    uint32_t index = valueHash(key) % capacity;
    Entry *tombstone = NULL;

    for (;;) {
        Entry *entry = &entries[index];
        if (IS_UNUSED(entry)) {
            if (IS_EMPTY(entry)) {
                // Empty entry.
                return tombstone != NULL ? tombstone : entry;
            }
            // Entry is a tombstone.
            if (tombstone == NULL) {
                tombstone = entry;
            }
        } else if (valuesEqual(entry->key, key)) {
            return entry;
        }
        index = (index + 1) % capacity;
    }
}

bool tableGet(Table *table, Value key, Value *value) {
    if (table->count == 0) {
        return false;
    }

    Entry *entry = findEntry(table->entries, table->capacity, key);
    if (IS_UNUSED(entry)) {
        return false;
    }

    *value = entry->value;
    return true;
}

static void adjustCapacity(Table *table, int capacity) {
    Entry *entries = ALLOCATE(Entry, capacity);
    for (int i = 0; i < capacity; i++) {
        // Empty entry: NULL key, nil value.
        entries[i].key = OBJ_VAL(NULL);
        entries[i].value = NIL_VAL;
    }

    // Reset table count to discard any tombstones.
    table->count = 0;
    for (int i = 0; i < table->capacity; i++) {
        Entry *entry = &table->entries[i];
        if (IS_UNUSED(entry)) {
            continue;
        }
        Entry *dest = findEntry(entries, capacity, entry->key);
        dest->key = entry->key;
        dest->value = entry->value;
        table->count++;
    }

    FREE_ARRAY(Entry, table->entries, table->capacity);
    table->entries = entries;
    table->capacity = capacity;
}

bool tableSet(Table *table, Value key, Value value) {
    if (table->count + 1 > table->capacity * TABLE_MAX_LOAD) {
        int capacity = GROW_CAPACITY(table->capacity);
        adjustCapacity(table, capacity);
    }
    Entry *entry = findEntry(table->entries, table->capacity, key);
    bool isNewKey = IS_UNUSED(entry);
    if (IS_EMPTY(entry)) {
        // Tombstones are not accounted when set or deleted.
        table->count++;
    }
    entry->key = key;
    entry->value = value;
    return isNewKey;
}

bool tableDelete(Table *table, Value key) {
    if (table->count == 0) {
        return false;
    }

    Entry *entry = findEntry(table->entries, table->capacity, key);
    if (IS_UNUSED(entry)) {
        return false;
    }
    entry->key = OBJ_VAL(NULL);
    entry->value = BOOL_VAL(true);
    return true;
}

void tableAddAll(Table *from, Table *to) {
    for (int i = 0; i < from->capacity; i++) {
        Entry *entry = &from->entries[i];
        if (!IS_UNUSED(entry)) {
            tableSet(to, entry->key, entry->value);
        }
    }
}

ObjString *tableFindString(Table *table, const char *chars, int length, uint32_t hash) {
    if (table->count == 0) {
        return NULL;
    }

    uint32_t index = hash % table->capacity;
    for (;;) {
        Entry *entry = &table->entries[index];
        if (IS_UNUSED(entry)) {
            if (IS_EMPTY(entry)) {
                // Stop if we find an empty non-tombstone entry.
                return NULL;
            }
        } else if (
                IS_STRING(entry->key) &&
                AS_STRING(entry->key)->length == length &&
                AS_STRING(entry->key)->hash == hash &&
                memcmp(AS_STRING(entry->key)->chars, chars, length) == 0) {
            return AS_STRING(entry->key);
        }
        index = (index + 1) % table->capacity;
    }
}
