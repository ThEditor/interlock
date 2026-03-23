#include "runtime.h"
#include "log.h"
#include <thread>
#include <string>
#include "quickjs.h"

static JSRuntime *rt = nullptr;
static JSContext *ctx = nullptr;

void runtime_start(std::string bundle_source)
{
	if (bundle_source.empty()) {
		GLUE_LOG("No JS source loaded, exiting...");
		return;
	}

	GLUE_LOG("Initiating");
	std::thread jsThread(run_js_runtime, std::move(bundle_source));
	jsThread.detach();
}

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

void run_js_runtime(std::string bundle_source)
{
	GLUE_LOG("Starting js runtime");
	rt = JS_NewRuntime();
	ctx = JS_NewContext(rt);

	JSValue global = JS_GetGlobalObject(ctx);
	JS_SetPropertyStr(
    ctx,
    global,
    "consoleLog",
    JS_NewCFunction(ctx, js_console_log, "consoleLog", 1)
	);

	JSValue eval_result = JS_Eval(
		ctx,
		bundle_source.c_str(),
		bundle_source.size(),
		"bundle.js",
		JS_EVAL_TYPE_GLOBAL
	);

	if (JS_IsException(eval_result)) {
		JSValue exception = JS_GetException(ctx);
		const char* error = JS_ToCString(ctx, exception);
		if (error != nullptr) {
			GLUE_LOG("QuickJS error: %s", error);
			JS_FreeCString(ctx, error);
		} else {
			GLUE_LOG("QuickJS error: <unable to stringify exception>");
		}
		JS_FreeValue(ctx, exception);
	} else {
		GLUE_LOG("Bundle executed successfully");
	}

	JS_FreeValue(ctx, eval_result);
	JS_FreeValue(ctx, global);
}
