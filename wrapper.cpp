#include "wrapper.h"
#include <taglib/fileref.h>
#include <taglib/tag.h>
#include <cstring>
#include <cstdlib>

Metadata* read_metadata(const char* filename) {
    TagLib::FileRef f(filename);
    if (f.isNull() || !f.tag()) return nullptr;

    TagLib::Tag* tag = f.tag();

    Metadata* meta = (Metadata*)malloc(sizeof(Metadata));
    if (!meta) return nullptr;

    auto strdup_cpp = [](const TagLib::String& s) -> char* {
        std::string utf8 = s.to8Bit(true);
        char* cstr = (char*)malloc(utf8.size() + 1);
        if (cstr) strcpy(cstr, utf8.c_str());
        return cstr;
    };

    meta->title  = strdup_cpp(tag->title());
    meta->artist = strdup_cpp(tag->artist());
    meta->album  = strdup_cpp(tag->album());
    meta->year   = tag->year();

    return meta;
}

void free_metadata(Metadata* data) {
    if (!data) return;
    free(data->title);
    free(data->artist);
    free(data->album);
    free(data);
}

