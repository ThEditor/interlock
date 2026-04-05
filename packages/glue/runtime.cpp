#include "runtime.h"
#include "log.h"
#include <thread>
#include <string>
#include "quickjs.h"
#include "binding.h"

static JSRuntime *rt = nullptr;
static JSContext *ctx = nullptr;

static JavaVM *g_jvm = nullptr;
static jobject g_activity = nullptr;

JNIEnv *get_jni_env_for_current_thread(bool *did_attach)
{
	if (did_attach != nullptr)
	{
		*did_attach = false;
	}

	if (g_jvm == nullptr)
	{
		return nullptr;
	}

	JNIEnv *env = nullptr;
	const jint get_env_result = g_jvm->GetEnv(reinterpret_cast<void **>(&env), JNI_VERSION_1_6);
	if (get_env_result == JNI_OK)
	{
		return env;
	}

	if (get_env_result != JNI_EDETACHED)
	{
		return nullptr;
	}

	if (g_jvm->AttachCurrentThread(&env, nullptr) != JNI_OK)
	{
		return nullptr;
	}

	if (did_attach != nullptr)
	{
		*did_attach = true;
	}

	return env;
}

void release_jni_env_for_current_thread(bool did_attach)
{
	if (did_attach && g_jvm != nullptr)
	{
		g_jvm->DetachCurrentThread();
	}
}

jobject get_activity()
{
	return g_activity;
}

void runtime_start(JNIEnv *env, jobject activity, std::string bundle_source)
{
	if (bundle_source.empty())
	{
		GLUE_LOG("No JS source loaded, exiting...");
		return;
	}

	if (env == nullptr || activity == nullptr)
	{
		GLUE_LOG("Invalid JNI input: env or activity is null");
		return;
	}

	if (env->GetJavaVM(&g_jvm) != JNI_OK)
	{
		GLUE_LOG("Failed to obtain JavaVM");
		return;
	}

	if (g_activity != nullptr)
	{
		env->DeleteGlobalRef(g_activity);
		g_activity = nullptr;
	}

	g_activity = env->NewGlobalRef(activity);
	if (g_activity == nullptr)
	{
		GLUE_LOG("Failed to create global reference for activity");
		return;
	}

	GLUE_LOG("Initiating");
	std::thread jsThread(run_js_runtime, std::move(bundle_source));
	jsThread.detach();
}

void run_js_runtime(std::string bundle_source)
{
	GLUE_LOG("Starting js runtime");
	rt = JS_NewRuntime();
	ctx = JS_NewContext(rt);

	bind_js(ctx);

	JSValue eval_result = JS_Eval(
			ctx,
			bundle_source.c_str(),
			bundle_source.size(),
			"bundle.js",
			JS_EVAL_TYPE_GLOBAL);

	if (JS_IsException(eval_result))
	{
		JSValue exception = JS_GetException(ctx);
		const char *error = JS_ToCString(ctx, exception);
		if (error != nullptr)
		{
			GLUE_LOG("QuickJS error: %s", error);
			JS_FreeCString(ctx, error);
		}
		else
		{
			GLUE_LOG("QuickJS error: <unable to stringify exception>");
		}
		JS_FreeValue(ctx, exception);
	}
	else
	{
		GLUE_LOG("Bundle executed successfully");
	}

	JS_FreeValue(ctx, eval_result);

	if (ctx != nullptr)
	{
		JS_FreeContext(ctx);
		ctx = nullptr;
	}
	if (rt != nullptr)
	{
		JS_FreeRuntime(rt);
		rt = nullptr;
	}

	bool did_attach = false;
	JNIEnv *env = get_jni_env_for_current_thread(&did_attach);
	if (env != nullptr && g_activity != nullptr)
	{
		env->DeleteGlobalRef(g_activity);
		g_activity = nullptr;
	}
	release_jni_env_for_current_thread(did_attach);
}
