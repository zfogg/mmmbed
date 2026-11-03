cmake_minimum_required(VERSION 3.15)
project(mmmbed_example C)
find_program(MMMBED_EXECUTABLE mmmbed REQUIRED)
set(GENERATED_ASSET "${CMAKE_CURRENT_BINARY_DIR}/asset.c")
add_custom_command(OUTPUT "${GENERATED_ASSET}"
  COMMAND "${MMMBED_EXECUTABLE}" -l c --omit-parent --variable greeting
          "${CMAKE_CURRENT_SOURCE_DIR}/greeting.txt" "${GENERATED_ASSET}"
  DEPENDS "${CMAKE_CURRENT_SOURCE_DIR}/greeting.txt"
  VERBATIM)
add_executable(app main.c "${GENERATED_ASSET}")
target_include_directories(app PRIVATE "${CMAKE_CURRENT_BINARY_DIR}")
