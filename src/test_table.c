#include <stdio.h>

#include "table.h"
#include "object.h"
#include "vm.h"

#define STRING_VAL(msg) OBJ_VAL(copyString(msg, sizeof(msg)/sizeof(char)))

int num_tests = 0;
int num_pass = 0;

void test_empty_map() {
    initVM();

    Table table;
    initTable(&table);

    Value keys[] = {
        NIL_VAL,
        NUMBER_VAL(123),
        BOOL_VAL(true),
        STRING_VAL("key"),
    };
    for (int i = 0; i < sizeof(keys)/sizeof(keys[0]); i++) {
        num_tests++;
        bool has_passed = true;

        Value value;
        if (tableGet(&table, keys[i], &value)) {
            has_passed = false;
            printf("test_empty_map: table[");
            printValue(keys[i]);
            printf("] = true, want false. Value: ");
            printValue(value);
            printf("\n");
        };
        if (tableDelete(&table, keys[i])) {
            has_passed = false;
            printf("test_empty_map: delete table[");
            printValue(keys[i]);
            printf("] = true, want false.\n");
        }
        if (has_passed) {
            num_pass++;
        }
    }
    num_tests++;
    if (table.capacity > 0) {
        printf("test_empty_map: table capacity = %d, want 0", table.capacity);
    } else {
        num_pass++;
    }

    freeTable(&table);
    freeVM();
}

void test_set_map() {
    initVM();

    Table table;
    initTable(&table);

    Value v1 = NIL_VAL;
    Value v2 = NUMBER_VAL(123);
    Value v3 = BOOL_VAL(true);
    Value v4 = STRING_VAL("my key");

    Value keys[] = {v1, v2, v3, v4};
    Value values[] = {v2, v3, v4, v1};
    int num_entries = sizeof(keys)/sizeof(keys[0]);

    // Test setting keys;
    for (int i = 0;  i < num_entries; i++) {
        num_tests++;
        bool has_passed = true;

        if(!tableSet(&table, keys[i], values[i])) {
            has_passed = false;
            printf("test_set_map: table[");
            printValue(keys[i]);
            printf("] = ");
            printValue(values[i]);
            printf(" => not a new key\n");
        }
        if (has_passed) {
            num_pass++;
        }
    }

    // Test getting keys.
    for (int i = 0;  i < num_entries; i++) {
        num_tests++;
        bool has_passed = true;

        Value got;
        if(!tableGet(&table, keys[i], &got)) {
            has_passed = false;
            printf("test_set_map: table[");
            printValue(keys[i]);
            printf("] failed\n");
        } else if (!valuesEqual(got, values[i])) {
            has_passed = false;
            printf("test_set_map: table[");
            printValue(keys[i]);
            printf("] = ");
            printValue(got);
            printf(" != ");
            printValue(values[i]);
            printf("\n");
        }
        if (has_passed) {
            num_pass++;
        }
    }

    // Test that capacity has increased.
    num_tests++;
    if (table.capacity < num_entries) {
        printf("test_set_map: table capacity = %d, want >=%d", table.capacity, num_entries);
    } else {
        num_pass++;
    }

    freeTable(&table);
    freeVM();
}

void test_reset_map() {
    initVM();

    Table table;
    initTable(&table);

    Value v1 = NIL_VAL;
    Value v2 = NUMBER_VAL(123);
    Value v3 = BOOL_VAL(true);
    Value v4 = STRING_VAL("my key");

    Value keys[] = {v1, v2, v3, v4};
    Value values1[] = {v2, v3, v4, v1};
    Value values2[] = {v3, v4, v1, v2};
    int num_entries = sizeof(keys)/sizeof(keys[0]);

    // Test setting keys;
    for (int i = 0;  i < num_entries; i++) {
        tableSet(&table, keys[i], values1[i]);
    }
    int prevCapacity = table.capacity;

    // Test resetting keys.
    for (int i = 0;  i < num_entries; i++) {
        num_tests++;
        bool has_passed = true;

        if(tableSet(&table, keys[i], values2[i])) {
            has_passed = false;
            printf("test_reset_map: table[");
            printValue(keys[i]);
            printf("] = ");
            printValue(values2[i]);
            printf(" => is a new key\n");
        }
        if (has_passed) {
            num_pass++;
        }
    }

    // Test getting reset keys.
    for (int i = 0;  i < num_entries; i++) {
        num_tests++;
        bool has_passed = true;

        Value got;
        if(!tableGet(&table, keys[i], &got)) {
            has_passed = false;
            printf("test_reset_map: table[");
            printValue(keys[i]);
            printf("] failed\n");
        } else if (!valuesEqual(got, values2[i])) {
            has_passed = false;
            printf("test_reset_map: table[");
            printValue(keys[i]);
            printf("] = ");
            printValue(got);
            printf(" != ");
            printValue(values2[i]);
            printf("\n");
        }
        if (has_passed) {
            num_pass++;
        }
    }

    // Test that capacity has not increased.
    num_tests++;
    if (table.capacity != prevCapacity) {
        printf("test_reset_map: table capacity = %d, want ==%d", table.capacity, prevCapacity);
    } else {
        num_pass++;
    }

    freeTable(&table);
    freeVM();
}

void test_rehash_map() {
    initVM();

    Table table;
    initTable(&table);

    num_tests++;
    bool has_passed = true;
    const int num_entries = 100;

    // Test setting keys;
    for (int i = 0;  i < num_entries; i++) {
        Value key = NUMBER_VAL(i);

        if(!tableSet(&table, key, key)) {
            has_passed = false;
            printf("test_rehash_map: table[");
            printValue(key);
            printf("] = ");
            printValue(key);
            printf(" => is not a new key\n");
        }
    }

    // Test getting keys.
    for (int i = 0;  i < num_entries; i++) {
        Value key = NUMBER_VAL(i);
        Value got;
        if(!tableGet(&table, key, &got)) {
            has_passed = false;
            printf("test_rehash_map: table[");
            printValue(key);
            printf("] failed\n");
        } else if (!valuesEqual(got, key)) {
            has_passed = false;
            printf("test_rehash_map: table[");
            printValue(key);
            printf("] = ");
            printValue(got);
            printf(" != ");
            printValue(key);
            printf("\n");
        }
    }
    if (has_passed) {
        num_pass++;
    }

    // Test that capacity has increased accordingly.
    num_tests++;
    if (table.capacity < num_entries) {
        printf("test_rehash_map: table capacity = %d, want >=%d", table.capacity, num_entries);
    } else {
        num_pass++;
    }

    freeTable(&table);
    freeVM();
}

int main() {
    test_empty_map();
    test_set_map();
    test_reset_map();
    test_rehash_map();

    printf("PASS %d/%d", num_pass, num_tests);
}
