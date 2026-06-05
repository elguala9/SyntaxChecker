-- Missing comma between table elements in CREATE TABLE.
-- SQL Server rejects this with "Incorrect syntax near 'data_inserimento'",
-- but the permissive bytebase T-SQL grammar treats the comma as optional
-- (column_def_table_constraints: column_def_table_constraint (','? column_def_table_constraint)*)
-- so a parse-only check currently reports it as valid.
CREATE TABLE bilanci_calcolati (
    id_bilanci_calcolati     UNIQUEIDENTIFIER PRIMARY KEY DEFAULT NEWSEQUENTIALID(),
    importo                  NUMERIC(18,2) NOT NULL,
    numero_righe             BIGINT        NOT NULL,
    CONSTRAINT uq_righe_somma UNIQUE (id_bilanci_calcolati, importo)
    data_inserimento         BIGINT NULL   -- <-- missing comma before this column
);
