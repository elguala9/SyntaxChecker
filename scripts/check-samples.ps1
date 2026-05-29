<#
.SYNOPSIS
  Runs the checker over every file in test-samples/ and verifies the outcome.

.DESCRIPTION
  Convention: a file whose name contains "_not_correct" MUST fail (exit 1);
  every other file MUST pass (exit 0). Type/dialect is inferred from the
  extension and file name. File types without a checker yet are reported SKIP.

  Exit code: 0 if all checks pass, 1 if any check fails.

.EXAMPLE
  pwsh scripts/check-samples.ps1
  pwsh scripts/check-samples.ps1 -Exe dist/syntax-checker.exe -SamplesDir test-samples
#>
param(
  [string]$Exe        = "$PSScriptRoot\..\dist\syntax-checker.exe",
  [string]$SamplesDir = "$PSScriptRoot\..\test-samples"
)

$ErrorActionPreference = "Stop"

if (-not (Test-Path $Exe))        { Write-Error "Executable not found: $Exe (build it first: make build)"; exit 2 }
if (-not (Test-Path $SamplesDir)) { Write-Error "Samples dir not found: $SamplesDir"; exit 2 }

# Maps a sample file to a checker --type, or $null when no checker exists yet.
function Resolve-Type([System.IO.FileInfo]$file) {
  $name = $file.Name.ToLower()
  switch ($file.Extension.ToLower()) {
    ".json" { return "json" }
    ".sql"  {
      if ($name -match "postgres|postgresql|pg")   { return "sql:postgres" }
      if ($name -match "mysql|tidb")               { return "sql:mysql" }
      if ($name -match "sqlite")                   { return "sql:sqlite" }
      if ($name -match "mssql|tsql|sqlserver")     { return "sql:mssql" }
      if ($name -match "oracle|plsql")             { return "sql:oracle" }
      return "sql:ansi"
    }
    ".xml"   { return "xml" }
    ".xhtml" { return "xml" }
    ".yaml"  { return "yaml" }
    ".yml"   { return "yaml" }
    ".toml"  { return "toml" }
    ".ini"   { return "ini" }
    ".cfg"   { return "ini" }
    ".csv"   { return "csv" }
    ".tsv"   { return "tsv" }
    ".hcl"   { return "hcl" }
    ".tf"    { return "hcl" }
    ".md"       { return "markdown" }
    ".markdown" { return "markdown" }
    ".env"        { return "env" }
    ".properties" { return "properties" }
    default { return $null }  # no checker for this extension yet
  }
}

$pass = 0; $fail = 0; $skip = 0
$failures = @()

Get-ChildItem -Path $SamplesDir -Recurse -File | Sort-Object FullName | ForEach-Object {
  $file = $_
  $type = Resolve-Type $file
  $rel  = $file.FullName.Substring((Resolve-Path $SamplesDir).Path.Length + 1)

  if ($null -eq $type) {
    "  SKIP  {0,-45} (no checker for {1})" -f $rel, $file.Extension
    $skip++
    return
  }

  # Always use --strict so JSON duplicate keys count as errors.
  $cliArgs = @("-f", $file.FullName, "-t", $type, "--strict")

  # Schema-validation samples live in test-samples/schema: a document
  # "<topic>.<variant>.<ext>" is validated against its sibling
  # "<topic>.schema.json" via --schema. The schema files themselves are
  # only syntax-checked (plain JSON).
  if (($file.Directory.Name -eq "schema") -and ($file.Name -notlike "*.schema.json")) {
    $topic      = ($file.Name -split '\.')[0]
    $schemaPath = Join-Path $file.DirectoryName "$topic.schema.json"
    if (-not (Test-Path $schemaPath)) {
      "  FAIL  {0,-45} (schema '{1}.schema.json' not found)" -f $rel, $topic
      $fail++; $failures += $rel
      return
    }
    $cliArgs += @("--schema", $schemaPath)
  }

  $output  = & $Exe @cliArgs
  $code    = $LASTEXITCODE

  $expectFail = $file.Name -match "_not_correct"
  $expected   = if ($expectFail) { 1 } else { 0 }
  $ok         = ($code -eq $expected)

  if ($ok) {
    $status = if ($expectFail) { "PASS (error detected)" } else { "PASS (valid)" }
    "  PASS  {0,-45} [{1}] exit={2}" -f $rel, $type, $code
    $pass++
  } else {
    "  FAIL  {0,-45} [{1}] expected exit={2}, got={3}" -f $rel, $type, $expected, $code
    foreach ($line in $output) { "          > $line" }
    $fail++
    $failures += $rel
  }
}

""
"Result: $pass PASS, $fail FAIL, $skip SKIP"
if ($fail -gt 0) {
  "Failed files:"
  $failures | ForEach-Object { "  - $_" }
  exit 1
}
exit 0
