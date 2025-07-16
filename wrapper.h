#ifndef TAGLIB_WRAPPER_H
#define TAGLIB_WRAPPER_H

#ifdef __cplusplus
extern "C" {
#endif

// Metadata structure to return to Go
typedef struct {
  char* title;
  char* artist;
  char* album_artist;
  char* album;
  char* genre;
  char* comment;
  char* codec;
  char* tag_type;
  char* date;
  int year;
  int duration;
  int track;
  int track_total;
  int disc;
  int disc_total;
  int bitrate;
  int frequency;
  int channels;
  int has_image;
} Metadata;

// Reads metadata from file. Returns NULL if failed.
Metadata* read_metadata(const char* filename);

// Frees metadata
void free_metadata(Metadata* data);

bool edit_metadata(const char* filename, const char* field, const char* value);

#ifdef __cplusplus
}
#endif

#endif // TAGLIB_WRAPPER_H

