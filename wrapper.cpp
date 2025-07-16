#include "wrapper.h"
#include <taglib/taglib.h>
#include <taglib/fileref.h>
#include <taglib/tag.h>
#include <taglib/tpropertymap.h>
#include <taglib/audioproperties.h>
#include <taglib/mpegfile.h>
#include <taglib/flacfile.h>
#include <taglib/oggfile.h>
#include <taglib/vorbisfile.h>
#include <taglib/mp4file.h>
#include <taglib/id3v2tag.h>
#include <cstring>
#include <cstdlib>
#include <sstream>

#ifndef _GNU_SOURCE
char* strdup(const char* s) {
    if (!s) return nullptr;
    size_t len = strlen(s) + 1;
    char* copy = (char*)malloc(len);
    if (copy) {
        strcpy(copy, s);
    }
    return copy;
}
#endif

char* strdup_cpp(const TagLib::String& s) {
  std::string utf8 = s.to8Bit(true);
  char* cstr = (char*)malloc(utf8.size() + 1);
  if (cstr) strcpy(cstr, utf8.c_str());
  return cstr;
}

bool hasEmbeddedImage(TagLib::File* f) {
  if (auto* mp3 = dynamic_cast<TagLib::MPEG::File*>(f)) {
    auto* id3v2tag = mp3->ID3v2Tag();
    if (id3v2tag) {
      auto frameListMap = id3v2tag->frameListMap();
      for (const auto& pair : frameListMap) {
        std::string frameId(pair.first.data(), pair.first.size());
      }
      return (!frameListMap["APIC"].isEmpty() || !frameListMap["PIC"].isEmpty());
    }
  }
  else if (auto* flac = dynamic_cast<TagLib::FLAC::File*>(f)) {
    return !flac->pictureList().isEmpty();
  }
  else if (auto* mp4 = dynamic_cast<TagLib::MP4::File*>(f)) {
    return mp4->tag()->contains("covr");
  }
  else if (auto* ogg = dynamic_cast<TagLib::Ogg::Vorbis::File*>(f)) {
    auto* tag = ogg->tag();
    if (tag && !tag->pictureList().isEmpty()) {
      return true;
    }
  }
  return false;
}


char* join_string_list(const TagLib::StringList& list) {
  if (list.isEmpty()) {
    return strdup("");
  }
  std::stringstream ss;
  bool first = true;
  for (const auto& str : list) {
    if (!first) {
       ss << ";";
    }
    ss << str.to8Bit(true);
    first = false;
  }
  return strdup(ss.str().c_str());
}

const char* get_codec_name(TagLib::File* file) {
  if (dynamic_cast<TagLib::MPEG::File*>(file)) return "MP3";
  if (dynamic_cast<TagLib::FLAC::File*>(file)) return "FLAC";
  if (dynamic_cast<TagLib::Ogg::Vorbis::File*>(file)) return "Ogg Vorbis";
  if (dynamic_cast<TagLib::MP4::File*>(file)) return "MP4";
  return "other";
}

const char* get_tag_type(TagLib::File* file) {
  if (dynamic_cast<TagLib::MPEG::File*>(file)) return "ID3v2";
  if (dynamic_cast<TagLib::FLAC::File*>(file)) return "FLAC";
  if (dynamic_cast<TagLib::Ogg::Vorbis::File*>(file)) return "Vorbis";
  if (dynamic_cast<TagLib::MP4::File*>(file)) return "MP4";
  return "other";
}

Metadata* read_metadata(const char* filename) {
  TagLib::FileRef f(filename);
  if (f.isNull() || !f.tag()) return nullptr;

  TagLib::Tag* tag = f.tag();
  TagLib::AudioProperties* props = f.audioProperties();
  TagLib::PropertyMap properties = f.file()->properties();

  Metadata* meta = (Metadata*)malloc(sizeof(Metadata));
  if (!meta) return nullptr;

  meta->year = -1;
  meta->disc = -1;
  meta->disc_total = -1;
  meta->track = -1;
  meta->track_total = -1;
  meta->bitrate = -1;
  meta->frequency = -1;
  meta->duration = -1;
  meta->channels = -1;
  meta->has_image = -1;

  meta->title  = strdup_cpp(tag->title());
  meta->artist = strdup_cpp(tag->artist());
  meta->album  = strdup_cpp(tag->album());

  if (tag->year() > 0) {
    meta->year   = tag->year();
  }

  meta->codec = strdup(get_codec_name(f.file()));
  meta->tag_type = strdup(get_tag_type(f.file()));

  if (props->bitrate() > 0) {
    meta->bitrate = props->bitrate();
  }
  if (props->sampleRate() > 0) {
    meta->frequency = props->sampleRate();
  }
  if (props->lengthInSeconds() > 0) {
    meta->duration = props->lengthInSeconds();
  }
  if (props->channels() > 0) {
    meta->channels = props->channels();
  }

  if (properties.contains("DATE")) {
    meta->date = join_string_list(properties["DATE"]);
  }

  if (properties.contains("TRACKNUMBER")) {
    std::string val = properties["TRACKNUMBER"].front().to8Bit(true);
    size_t pos = val.find('/');
    if (pos != std::string::npos) {
      meta->track = std::stoi(val.substr(0, pos));
      meta->track_total = std::stoi(val.substr(pos+1));
      if (meta->track_total == 0) {
        meta->track_total = -1;
      }
    } else {
      meta->track = std::stoi(val);
    }
  }

  if (properties.contains("DISCNUMBER")) {
    std::string val = properties["DISCNUMBER"].front().to8Bit(true);
    size_t pos = val.find('/');
    if (pos != std::string::npos) {
      meta->disc = std::stoi(val.substr(0, pos));
      meta->disc_total = std::stoi(val.substr(pos+1));
      if (meta->disc_total == 0) {
        meta->track_total = -1;
      }
    } else {
      meta->disc = std::stoi(val);
    }
  }

  if (properties.contains("ALBUMARTIST")) {
    meta->album_artist = join_string_list(properties["ALBUMARTIST"]);
  } else {
    meta->album_artist = strdup("");
  }

  if (properties.contains("ARTIST")) {
    meta->artist = join_string_list(properties["ARTIST"]);
  } else {
    meta->artist = strdup("");
  }

  if (properties.contains("GENRE")) {
    meta->genre = join_string_list(properties["GENRE"]);
  } else {
    meta->genre = strdup("");
  }

  if (properties.contains("COMMENT")) {
    meta->comment = join_string_list(properties["COMMENT"]);
  } else {
    meta->comment = strdup("");
  }

  meta->has_image = hasEmbeddedImage(f.file()) ? 1 : 0;

  return meta;
}

bool edit_metadata(const char* filename, const char* field, const char* value) {
    TagLib::FileRef f(filename);
    return true;
    if (f.isNull() || !f.tag()) {
        return false;
    }
    return true;
    if (std::strcmp(value, "test")) {
        return false;
    }
    return true;
}

void free_metadata(Metadata* data) {
  if (!data) return;
  free(data->title);
  free(data->artist);
  free(data->album);
  free(data);
}

