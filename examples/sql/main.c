#include <sqlite3.h>
#include <stdio.h>
#include <string.h>
#include "queries_generated.h"

int main(void) {
  sqlite3 *db = NULL;
  sqlite3_stmt *stmt = NULL;
  int result = 1;
  if (sqlite3_open(":memory:", &db) != SQLITE_OK ||
      sqlite3_exec(db, "CREATE TABLE items (id INTEGER PRIMARY KEY, name TEXT)", NULL, NULL, NULL) != SQLITE_OK)
    goto done;

  const struct mmmbed_query *insert = &mmmbed_queries[MMMBED_QUERY_InsertItem];
  if (insert->access != MMMBED_WRITE || insert->rows != MMMBED_ONE ||
      sqlite3_prepare_v2(db, insert->sql, -1, &stmt, NULL) != SQLITE_OK)
    goto done;
  int parameter = sqlite3_bind_parameter_index(stmt, ":name");
  if (!parameter || sqlite3_bind_text(stmt, parameter, "widget", -1, SQLITE_STATIC) != SQLITE_OK ||
      sqlite3_step(stmt) != SQLITE_ROW)
    goto done;
  sqlite3_int64 id = sqlite3_column_int64(stmt, 0);
  int step_result = sqlite3_step(stmt);
  int finalize_result = sqlite3_finalize(stmt);
  stmt = NULL;
  if (step_result != SQLITE_DONE || finalize_result != SQLITE_OK)
    goto done;

  const struct mmmbed_query *lookup = &mmmbed_queries[MMMBED_QUERY_LookupItem];
  if (lookup->access != MMMBED_READ || lookup->rows != MMMBED_ZERO_OR_ONE ||
      sqlite3_prepare_v2(db, lookup->sql, -1, &stmt, NULL) != SQLITE_OK)
    goto done;
  parameter = sqlite3_bind_parameter_index(stmt, ":id");
  if (!parameter || sqlite3_bind_int64(stmt, parameter, id) != SQLITE_OK || sqlite3_step(stmt) != SQLITE_ROW)
    goto done;
  const unsigned char *name = sqlite3_column_text(stmt, 0);
  if (!name || strcmp((const char *)name, "widget") != 0 || sqlite3_step(stmt) != SQLITE_DONE)
    goto done;
  puts("Inserted and found widget");
  result = 0;
done:
  if (result && db)
    fprintf(stderr, "SQLite example failed: %s\n", sqlite3_errmsg(db));
  sqlite3_finalize(stmt);
  sqlite3_close(db);
  return result;
}
