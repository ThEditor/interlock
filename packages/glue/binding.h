#pragma once

static JSValue js_console_log(JSContext *ctx,
                              JSValueConst this_val,
                              int argc,
                              JSValueConst *argv);

void render(JSContext *ctx,
            JSValueConst this_val,
            int argc,
            JSValueConst *argv);

void bind_js(JSContext *ctx);
