-- Missing comma between table elements in CREATE TABLE.
-- SQLite rejects this with "near 'data_inserimento': syntax error", but the
-- rqlite/sql parser used by the checker loops over column definitions without
-- requiring the separating comma (parseColumnDefinitions), so a parse-only
-- check would report it as valid. The checker re-scans the source to flag it.
-- Note: a comma is only required once a column carries a constraint or before a
-- table constraint; "a INT b INT" is a legitimate multi-word SQLite type name.
CREATE TABLE bilanci_calcolati (
    id_bilanci_calcolati     INTEGER PRIMARY KEY,
    importo                  NUMERIC NOT NULL,
    numero_righe             INTEGER NOT NULL   -- <-- missing comma before next column
    data_inserimento         INTEGER NULL
);
