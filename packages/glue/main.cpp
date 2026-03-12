#include <jni.h>

#if defined(__ANDROID__)
#include <android/log.h>
#endif

namespace {
constexpr const char* kLogTag = "interlock-glue";
}

extern "C" JNIEXPORT void JNICALL
Java_xyz_theditor_interlocktest_MainActivity_startRuntime(JNIEnv* env, jobject /* thiz */) {
	(void)env;

#if defined(__ANDROID__)
	__android_log_print(ANDROID_LOG_INFO, kLogTag, "Native runtime started");
#endif
}
