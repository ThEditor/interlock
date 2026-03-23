// Android JNI bridge
// TODO: On iOS, call runtime_start() from an Objective-C/Swift bridge instead.
#include <jni.h>
#include <android/asset_manager.h>
#include <android/asset_manager_jni.h>
#include <string>
#include "log.h"
#include "runtime.h"

static std::string load_bundle_from_assets(JNIEnv* env, jobject asset_manager)
{
	AAssetManager* native_asset_manager = AAssetManager_fromJava(env, asset_manager);
	if (native_asset_manager == nullptr) {
		GLUE_LOG("AssetManager is null, using fallback bundle");
		return {};
	}

	AAsset* bundle_asset = AAssetManager_open(native_asset_manager, "bundle.js", AASSET_MODE_BUFFER);
	if (bundle_asset == nullptr) {
		GLUE_LOG("Could not open assets/bundle.js, using fallback bundle");
		return {};
	}

	const off_t length = AAsset_getLength(bundle_asset);
	if (length <= 0) {
		GLUE_LOG("assets/bundle.js is empty, using fallback bundle");
		AAsset_close(bundle_asset);
		return {};
	}

	std::string source(static_cast<size_t>(length), '\0');
	const int64_t bytes_read = AAsset_read(bundle_asset, source.data(), source.size());
	AAsset_close(bundle_asset);

	if (bytes_read < 0) {
		GLUE_LOG("Failed reading assets/bundle.js, using fallback bundle");
		return {};
	}

	if (static_cast<size_t>(bytes_read) < source.size()) {
		source.resize(static_cast<size_t>(bytes_read));
	}

	GLUE_LOG("Loaded assets/bundle.js (%lld bytes)", static_cast<long long>(bytes_read));
	return source;
}

extern "C" JNIEXPORT void JNICALL
Java_xyz_theditor_interlocktest_MainActivity_startRuntime(JNIEnv* env, jobject /* thiz */, jobject asset_manager) {
	runtime_start(load_bundle_from_assets(env, asset_manager));
}
