#pragma once

#include <cstdio>

#if defined(__ANDROID__)
#  include <android/log.h>
#  define GLUE_LOG(fmt, ...) \
       __android_log_print(ANDROID_LOG_INFO, "interlock-glue", fmt, ##__VA_ARGS__)
#elif defined(__APPLE__)
// TODO: os_log is preferred on Apple platforms but requires an OS_LOG_DEFAULT object;
// TODO: use stderr for now so any iOS/macOS host can redirect it with its own bridge.
#  define GLUE_LOG(fmt, ...) \
       fprintf(stderr, "[interlock-glue] " fmt "\n", ##__VA_ARGS__)
#else
#  define GLUE_LOG(fmt, ...) \
       fprintf(stderr, "[interlock-glue] " fmt "\n", ##__VA_ARGS__)
#endif
