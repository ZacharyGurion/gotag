#ifndef TAGLIB_WRAPPER_H
#define TAGLIB_WRAPPER_H

#ifdef __cplusplus
extern "C" {
#endif

// Metadata structure to return to Go
typedef struct {
    char* title;
    char* artist;
    char* album;
    int year;
} Metadata;

// Reads metadata from file. Returns NULL if failed.
Metadata* read_metadata(const char* filename);

// Frees metadata
void free_metadata(Metadata* data);

#ifdef __cplusplus
}
#endif

#endif // TAGLIB_WRAPPER_H

