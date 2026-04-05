#pragma once

#include <string>
#include <jni.h>

JNIEnv* get_jni_env_for_current_thread(bool* did_attach);
void release_jni_env_for_current_thread(bool did_attach);
jobject get_activity();

void runtime_start(JNIEnv* env, jobject activity, std::string bundle_source);
void run_js_runtime(std::string bundle_source);
