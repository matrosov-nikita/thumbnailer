#pragma once
#include "ffmpeg.h"

struct Buffer {
    uint8_t* data;
    size_t size;
    unsigned long width, height;
};

// Writes RGBA thumbnail buffer to img
int extract_image(struct Buffer* img, AVFormatContext* avfc,
    AVCodecContext* avcc, const int stream);
