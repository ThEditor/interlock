#include "runtime.h"
#include "log.h"
#include <thread>
#include "quickjs.h"

static JSRuntime *rt = nullptr;
static JSContext *ctx = nullptr;

const char* code = "consoleLog('QuickJS is running');";

void runtime_start()
{
	GLUE_LOG("Initiating");
	std::thread jsThread(run_js_runtime);
	jsThread.detach();
}

static JSValue js_console_log(JSContext* ctx,
                              JSValueConst this_val,
                              int argc,
                              JSValueConst* argv)
{
    const char* msg = JS_ToCString(ctx, argv[0]);

    GLUE_LOG("%s", msg);

    JS_FreeCString(ctx, msg);
    return JS_UNDEFINED;
}

void run_js_runtime()
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

	JS_Eval(ctx,
					code,
					strlen(code),
					"bundle.js",
					JS_EVAL_TYPE_GLOBAL);
}