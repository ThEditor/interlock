# Interlock

A runtime system that allows JavaScript code to execute inside an Android application and interact with Android APIs.

## Overview

Interlock embeds a JavaScript engine inside a C++ runtime layer, which acts as the interoperability layer between JavaScript and Android. It enables running a JS bundle, exposing Android capabilities to JS, and bidirectional communication between JS and native code.

```
Android App (Java/Kotlin)
        ↓
JNI Interface
        ↓
C++ Runtime (JS thread, host objects, Android adapter)
        ↓
JavaScript Engine
        ↓
JavaScript Bundle
```

## Key Concepts

- **C++ Runtime** — manages JS execution on a dedicated thread and registers host objects
- **Host Objects** — structured native modules accessible from JS (e.g. `Native.Device.getBattery()`)
- **JNI Layer** — bridges calls between the Android JVM and the C++ runtime
- **Engine Agnostic** — supports V8, Hermes, JavaScriptCore, QuickJS, or others

## Layout

This repository is set up as a `pnpm` monorepo.

- `apps/` for runnable applications
- `packages/` for shared libraries and utilities


## QuickStart

### Using CLI
```bash
git clone --recurse-submodules git@github.com:ThEditor/interlock.git
make

mkdir tmp && cd tmp
../bin/interlock init # start an interlock project
cd <project>
../../bin/interlock project dev # dev mode
```
### Without CLI
(only used CLI for `esbuild` bundling)

Setup
```bash
git clone --recurse-submodules git@github.com:ThEditor/interlock.git
make

mkdir tmp && cd tmp
../bin/interlock init # start an interlock project
cd <project>
../../bin/interlock project bundle # run this for each js code change
```

Compile
```bash
./gradlew clean
./gradlew assembleDebug
# or for watched changes
# ./gradlew installDebug --continuous

```

Execute
```bash
# connect an android device
# emulator
# emulator -list-avds
# emulator -avd avd_name
# wireless adb
# adb pair <ip-addr>:<port>
# adb connect <ip-addr>:<port>

# adb install not required if using installDebug
# adb install android/app/build/outputs/apk/debug/app-debug.apk

adb shell am start -n <package>/.MainActivity
adb logcat --pid=$(adb shell pidof -s <package>)
```
