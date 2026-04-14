#pragma once

#include <string>
#include <jni.h>

struct RuntimeContext {
	jobject activity;
};

JNIEnv* get_jni_env_for_current_thread(bool* did_attach);
void release_jni_env_for_current_thread(bool did_attach);

void runtime_start(JNIEnv* env, jobject activity, std::string bundle_source);
void run_js_runtime(jobject activity, std::string bundle_source);
