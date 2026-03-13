cmake_minimum_required(VERSION 3.18.1)
project(glue LANGUAGES C CXX)

set(GLUE_TARGET glue)

file(
	GLOB GLUE_SOURCES CONFIGURE_DEPENDS
	"${CMAKE_CURRENT_SOURCE_DIR}/*.cpp"
	"${CMAKE_CURRENT_SOURCE_DIR}/*.cc"
	"${CMAKE_CURRENT_SOURCE_DIR}/*.cxx"
)

if(NOT GLUE_SOURCES)
	message(FATAL_ERROR "No C++ source files found in ${CMAKE_CURRENT_SOURCE_DIR}")
endif()

add_library(glue SHARED
  ${GLUE_SOURCES}

  ../quickjs/quickjs.c
  ../quickjs/libregexp.c
  ../quickjs/libunicode.c
  ../quickjs/dtoa.c
)

target_include_directories(glue PRIVATE
  ${CMAKE_CURRENT_SOURCE_DIR}/../quickjs
)

target_compile_features(${GLUE_TARGET} PRIVATE cxx_std_17)

if(ANDROID)
	find_library(LOG_LIB log)
	target_link_libraries(${GLUE_TARGET} PRIVATE ${LOG_LIB})
endif()
