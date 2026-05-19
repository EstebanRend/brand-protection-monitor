#!/usr/bin/env node

import { existsSync, readFileSync } from "node:fs";
import { resolve } from "node:path";
import { spawnSync } from "node:child_process";

const repoRoot = resolve(process.cwd());
const migrationsDir = "migrations";
const migrationsPathArg = "./migrations";
const backendEnvPath = resolve(repoRoot, "backend", ".env");

function printHelp() {
  console.log(
    [
      "Migration helper",
      "",
      "Usage:",
      "  npm run mig:new -- <name>      Create new up/down migration files",
      "  npm run mig:up                  Apply pending migrations",
      "  npm run mig:down -- [steps]     Roll back migrations (default: 1)",
      "  npm run mig:version             Show current migration version",
      "  npm run mig:force -- <version>  Force migration version",
      "",
      "Notes:",
      "  - Requires `migrate` CLI installed and available in PATH.",
      "  - DATABASE_URL is read from process env first, then backend/.env.",
    ].join("\n")
  );
}

function parseEnvFile(path) {
  if (!existsSync(path)) return {};
  const content = readFileSync(path, "utf8");
  const result = {};

  for (const rawLine of content.split(/\r?\n/)) {
    const line = rawLine.trim();
    if (!line || line.startsWith("#")) continue;
    const idx = line.indexOf("=");
    if (idx <= 0) continue;
    const key = line.slice(0, idx).trim();
    let value = line.slice(idx + 1).trim();
    if (
      (value.startsWith('"') && value.endsWith('"')) ||
      (value.startsWith("'") && value.endsWith("'"))
    ) {
      value = value.slice(1, -1);
    }
    result[key] = value;
  }

  return result;
}

function getDatabaseURL() {
  if (process.env.DATABASE_URL) return process.env.DATABASE_URL;
  const envVars = parseEnvFile(backendEnvPath);
  return envVars.DATABASE_URL;
}

function runMigrate(args) {
  const result = spawnSync("migrate", args, {
    stdio: "inherit",
    shell: process.platform === "win32",
    cwd: repoRoot,
  });

  if (result.error) {
    if (result.error.code === "ENOENT") {
      console.error("`migrate` CLI is not installed or not in PATH.");
      console.error(
        "Install with: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"
      );
    } else {
      console.error(result.error.message);
    }
    process.exit(1);
  }

  process.exit(result.status ?? 1);
}

const command = process.argv[2];
const arg = process.argv[3];

if (!command || command === "help") {
  printHelp();
  process.exit(0);
}

if (command === "new") {
  if (!arg) {
    console.error("Missing migration name. Example: npm run mig:new -- add_domain_index");
    process.exit(1);
  }
  runMigrate(["create", "-ext", "sql", "-dir", migrationsDir, "-seq", arg]);
}

if (command === "up") {
  const databaseURL = getDatabaseURL();
  if (!databaseURL) {
    console.error("DATABASE_URL not found in environment or backend/.env.");
    process.exit(1);
  }
  runMigrate(["-path", migrationsPathArg, "-database", databaseURL, "up"]);
}

if (command === "down") {
  const databaseURL = getDatabaseURL();
  if (!databaseURL) {
    console.error("DATABASE_URL not found in environment or backend/.env.");
    process.exit(1);
  }
  const steps = arg || "1";
  runMigrate(["-path", migrationsPathArg, "-database", databaseURL, "down", steps]);
}

if (command === "version") {
  const databaseURL = getDatabaseURL();
  if (!databaseURL) {
    console.error("DATABASE_URL not found in environment or backend/.env.");
    process.exit(1);
  }
  runMigrate(["-path", migrationsPathArg, "-database", databaseURL, "version"]);
}

if (command === "force") {
  const databaseURL = getDatabaseURL();
  if (!databaseURL) {
    console.error("DATABASE_URL not found in environment or backend/.env.");
    process.exit(1);
  }
  if (!arg) {
    console.error("Missing version. Example: npm run mig:force -- 2");
    process.exit(1);
  }
  runMigrate(["-path", migrationsPathArg, "-database", databaseURL, "force", arg]);
}

console.error(`Unknown command: ${command}`);
printHelp();
process.exit(1);
