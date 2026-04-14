#include "quickjs.h"
#include "log.h"
#include <jni.h>
#include "runtime.h"

struct ScopedJniEnv {
  JNIEnv* env;
  bool did_attach;

  ScopedJniEnv() : env(nullptr), did_attach(false) {
    env = get_jni_env_for_current_thread(&did_attach);
  }

  ~ScopedJniEnv() {
    release_jni_env_for_current_thread(did_attach);
  }
};

static JSValue js_console_log(JSContext* ctx,
                              JSValueConst this_val,
                              int argc,
                              JSValueConst* argv)
{
	if (argc < 1) {
		GLUE_LOG("<empty>");
		return JS_UNDEFINED;
	}

	const char* msg = JS_ToCString(ctx, argv[0]);
	if (msg == nullptr) {
		GLUE_LOG("<non-string>");
		return JS_UNDEFINED;
	}

    GLUE_LOG("%s", msg);

    JS_FreeCString(ctx, msg);
    return JS_UNDEFINED;
}

static JSValue render(JSContext* ctx,
                              JSValueConst this_val,
                              int argc,
                              JSValueConst* argv)
{
  if (argc < 1) {
    return JS_ThrowTypeError(ctx, "render expects one argument");
  }

  if (!JS_IsArray(argv[0]))
    return JS_ThrowTypeError(ctx, "Argument must be an array");

  JSValue len_val = JS_GetPropertyStr(ctx, argv[0], "length");
  uint32_t len; // length of array
  JS_ToUint32(ctx, &len, len_val);
  JS_FreeValue(ctx, len_val);

  ScopedJniEnv jni;
  JNIEnv* env = jni.env;
  if (env == nullptr) {
    return JS_ThrowInternalError(ctx, "JNIEnv is unavailable on this thread");
  }

  jclass stringClass = env->FindClass("java/lang/String");
  if (stringClass == nullptr) {
    if (env->ExceptionCheck()) env->ExceptionClear();
    return JS_ThrowInternalError(ctx, "Failed to resolve java.lang.String class");
  }

  jobjectArray typeArray = env->NewObjectArray(len, stringClass, nullptr);
  if (typeArray == nullptr) {
    if (env->ExceptionCheck()) env->ExceptionClear();
    env->DeleteLocalRef(stringClass);
    return JS_ThrowInternalError(ctx, "Failed to allocate types array");
  }

  jobjectArray textArray = env->NewObjectArray(len, stringClass, nullptr);
  if (textArray == nullptr) {
    if (env->ExceptionCheck()) env->ExceptionClear();
    env->DeleteLocalRef(typeArray);
    env->DeleteLocalRef(stringClass);
    return JS_ThrowInternalError(ctx, "Failed to allocate texts array");
  }

  for (uint32_t i = 0; i < len; i++) {
    JSValue obj = JS_GetPropertyUint32(ctx, argv[0], i);

    JSValue type_val = JS_GetPropertyStr(ctx, obj, "type");
    const char *type = JS_ToCString(ctx, type_val);
    if (type == nullptr) {
      JS_FreeValue(ctx, type_val);
      JS_FreeValue(ctx, obj);
      env->DeleteLocalRef(textArray);
      env->DeleteLocalRef(typeArray);
      env->DeleteLocalRef(stringClass);
      return JS_ThrowTypeError(ctx, "render item 'type' must be a string");
    }
    jstring jtype = env->NewStringUTF(type);
    if (jtype == nullptr) {
      if (env->ExceptionCheck()) env->ExceptionClear();
      JS_FreeCString(ctx, type);
      JS_FreeValue(ctx, type_val);
      JS_FreeValue(ctx, obj);
      env->DeleteLocalRef(textArray);
      env->DeleteLocalRef(typeArray);
      env->DeleteLocalRef(stringClass);
      return JS_ThrowInternalError(ctx, "OOM creating jstring for type");
    }

    JSValue text_val = JS_GetPropertyStr(ctx, obj, "text");
    const char *text = JS_ToCString(ctx, text_val);
    if (text == nullptr) {
      env->DeleteLocalRef(jtype);
      JS_FreeCString(ctx, type);
      JS_FreeValue(ctx, type_val);
      JS_FreeValue(ctx, text_val);
      JS_FreeValue(ctx, obj);
      env->DeleteLocalRef(textArray);
      env->DeleteLocalRef(typeArray);
      env->DeleteLocalRef(stringClass);
      return JS_ThrowTypeError(ctx, "render item 'text' must be a string");
    }
    jstring jtext = env->NewStringUTF(text);
    if (jtext == nullptr) {
      if (env->ExceptionCheck()) env->ExceptionClear();
      env->DeleteLocalRef(jtype);
      JS_FreeCString(ctx, text);
      JS_FreeValue(ctx, text_val);
      JS_FreeCString(ctx, type);
      JS_FreeValue(ctx, type_val);
      JS_FreeValue(ctx, obj);
      env->DeleteLocalRef(textArray);
      env->DeleteLocalRef(typeArray);
      env->DeleteLocalRef(stringClass);
      return JS_ThrowInternalError(ctx, "OOM creating jstring for text");
    }

    env->SetObjectArrayElement(typeArray, i, jtype);
    env->SetObjectArrayElement(textArray, i, jtext);

    env->DeleteLocalRef(jtext);
    env->DeleteLocalRef(jtype);
    
    JS_FreeCString(ctx, text);
    JS_FreeValue(ctx, text_val);
    JS_FreeValue(ctx, obj);

    JS_FreeCString(ctx, type);
    JS_FreeValue(ctx, type_val);
  }

  // Call MainActivity.createViews with the array of Pair objects
  RuntimeContext* rctx = reinterpret_cast<RuntimeContext*>(JS_GetContextOpaque(ctx));
  if (rctx == nullptr) {
    env->DeleteLocalRef(textArray);
    env->DeleteLocalRef(typeArray);
    env->DeleteLocalRef(stringClass);
    return JS_ThrowInternalError(ctx, "Activity context is unavailable");
  }
  jobject activity = rctx->activity;
  
  if (activity == nullptr) {
    env->DeleteLocalRef(textArray);
    env->DeleteLocalRef(typeArray);
    env->DeleteLocalRef(stringClass);
    return JS_ThrowInternalError(ctx, "Activity reference is unavailable");
  }

  jclass activityClass = env->GetObjectClass(activity);
  if (activityClass == nullptr) {
    if (env->ExceptionCheck()) env->ExceptionClear();
    env->DeleteLocalRef(textArray);
    env->DeleteLocalRef(typeArray);
    env->DeleteLocalRef(stringClass);
    return JS_ThrowInternalError(ctx, "Failed to resolve activity class");
  }

  jmethodID createViewsMethod = env->GetMethodID(activityClass, "enqueueCreateViews", "([Ljava/lang/String;[Ljava/lang/String;)V");
  if (createViewsMethod == nullptr) {
    if (env->ExceptionCheck()) env->ExceptionClear();
    env->DeleteLocalRef(activityClass);
    env->DeleteLocalRef(textArray);
    env->DeleteLocalRef(typeArray);
    env->DeleteLocalRef(stringClass);
    return JS_ThrowInternalError(ctx, "Failed to resolve MainActivity.enqueueCreateViews");
  }

  env->CallVoidMethod(activity, createViewsMethod, typeArray, textArray);
  if (env->ExceptionCheck()) {
    env->ExceptionDescribe();
    env->ExceptionClear();
    env->DeleteLocalRef(activityClass);
    env->DeleteLocalRef(textArray);
    env->DeleteLocalRef(typeArray);
    env->DeleteLocalRef(stringClass);
    return JS_ThrowInternalError(ctx, "Exception while calling MainActivity.enqueueCreateViews");
  }
  env->DeleteLocalRef(activityClass);
  env->DeleteLocalRef(textArray);
  env->DeleteLocalRef(typeArray);
  env->DeleteLocalRef(stringClass);

  return JS_UNDEFINED;
}

void bind_js(JSContext *ctx) {
	JSValue global = JS_GetGlobalObject(ctx);

	JS_SetPropertyStr(
    ctx,
    global,
    "consoleLog",
    JS_NewCFunction(ctx, js_console_log, "consoleLog", 1)
	);

	JS_SetPropertyStr(
    ctx,
    global,
    "render",
    JS_NewCFunction(ctx, render, "render", 1)
	);

  JS_FreeValue(ctx, global);
}
