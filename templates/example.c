#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdbool.h>

#define MAX_NAME_LEN 64
#define INITIAL_CAPACITY 8

/* A dynamically-sized array of strings. */
typedef struct {
    char **items;
    size_t count;
    size_t capacity;
} StringList;

// Create a new empty string list.
StringList *string_list_new(void) {
    StringList *list = malloc(sizeof(StringList));
    if (list == NULL) {
        fprintf(stderr, "Error: out of memory\n");
        exit(EXIT_FAILURE);
    }
    list->items = malloc(sizeof(char *) * INITIAL_CAPACITY);
    list->count = 0;
    list->capacity = INITIAL_CAPACITY;
    return list;
}

// Append a string to the list (copies the string).
bool string_list_push(StringList *list, const char *str) {
    if (list->count >= list->capacity) {
        size_t new_cap = list->capacity * 2;
        char **new_items = realloc(list->items, sizeof(char *) * new_cap);
        if (new_items == NULL) {
            return false;
        }
        list->items = new_items;
        list->capacity = new_cap;
    }

    size_t len = strlen(str);
    char *copy = malloc(len + 1);
    if (copy == NULL) {
        return false;
    }
    memcpy(copy, str, len + 1);
    list->items[list->count++] = copy;
    return true;
}

// Print all items in the list.
void string_list_print(const StringList *list) {
    printf("StringList (%zu items):\n", list->count);
    for (size_t i = 0; i < list->count; i++) {
        printf("  [%zu] \"%s\"\n", i, list->items[i]);
    }
}

// Free the string list and all its strings.
void string_list_free(StringList *list) {
    for (size_t i = 0; i < list->count; i++) {
        free(list->items[i]);
    }
    free(list->items);
    free(list);
}

int main(int argc, char *argv[]) {
    StringList *names = string_list_new();

    const char *defaults[] = {"Alice", "Bob", "Charlie", "Diana"};
    int num_defaults = sizeof(defaults) / sizeof(defaults[0]);

    for (int i = 0; i < num_defaults; i++) {
        if (!string_list_push(names, defaults[i])) {
            fprintf(stderr, "Failed to add: %s\n", defaults[i]);
        }
    }

    // Add command-line arguments too
    for (int i = 1; i < argc; i++) {
        if (strlen(argv[i]) < MAX_NAME_LEN) {
            string_list_push(names, argv[i]);
        }
    }

    string_list_print(names);
    printf("Total: %zu names\n", names->count);

    string_list_free(names);
    return EXIT_SUCCESS;
}
