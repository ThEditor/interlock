// Android JNI bridge
// TODO: On iOS, call runtime_start() from an Objective-C/Swift bridge instead.
#include <jni.h>
#include "runtime.h"

extern "C" JNIEXPORT void JNICALL
Java_xyz_theditor_interlocktest_MainActivity_startRuntime(JNIEnv* /* env */, jobject /* thiz */) {
	runtime_start();
}
