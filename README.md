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

Terminal 1
```bash
cd apps/interlocktest
./gradlew clean
./gradlew assembleDebug
# or for watched changes
# ./gradlew installDebug --continuous

```

Terminal 2
```bash
# connect an android device
# emulator
# emulator -list-avds
# emulator -avd avd_name
# wireless adb
# adb pair <ip-addr>:<port>
# adb connect <ip-addr>:<port>

# adb install not required if using installDebug
# adb install apps/interlocktest/app/build/outputs/apk/debug/app-debug.apk

adb shell am start -n xyz.theditor.interlocktest/.MainActivity
adb logcat --pid=$(adb shell pidof -s xyz.theditor.interlocktest)
```
