cmake_minimum_required(VERSION 3.15)
project(mmmbed_sql_example C)
find_program(MMMBED_EXECUTABLE mmmbed REQUIRED)
find_package(SQLite3 3.35 REQUIRED)
file(GLOB_RECURSE QUERY_FILES CONFIGURE_DEPENDS "${CMAKE_CURRENT_SOURCE_DIR}/queries/*.sql")
set(GENERATED_QUERIES "${CMAKE_CURRENT_BINARY_DIR}/queries_generated.h")
add_custom_command(OUTPUT "${GENERATED_QUERIES}"
  COMMAND "${MMMBED_EXECUTABLE}" sql -l c -i "${CMAKE_CURRENT_SOURCE_DIR}/queries" -o "${GENERATED_QUERIES}"
  DEPENDS "${MMMBED_EXECUTABLE}" ${QUERY_FILES}
  VERBATIM)
add_executable(sql_example main.c "${GENERATED_QUERIES}")
target_include_directories(sql_example PRIVATE "${CMAKE_CURRENT_BINARY_DIR}")
if(TARGET SQLite3::SQLite3)
  target_link_libraries(sql_example PRIVATE SQLite3::SQLite3)
else()
  target_link_libraries(sql_example PRIVATE SQLite::SQLite3)
endif()
