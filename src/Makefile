
FILES=chunk compiler debug object scanner table value memory vm
SOURCE_FILES=$(patsubst %,%.c,$(FILES))
SOURCE_HEADERS=$(patsubst %,%.h,$(FILES))

clox: $(SOURCE_FILES) $(SOURCE_HEADERS) main.c
	$(CC) $(SOURCE_FILES) main.c -g -o $@

test_table: $(SOURCE_FILES) $(SOURCE_HEADERS) test_table.c
	$(CC) $(SOURCE_FILES) test_table.c -g -o $@

build_test: test_table

build: clox build_test

test: build_test
	./test_table
