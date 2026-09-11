// Run the repository's pinned pnpm without Corepack or a global pnpm install.
const { spawnSync } = require("node:child_process");
const { existsSync } = require("node:fs");
const path = require("node:path");
const { packageManager } = require("../../package.json");

const pin =
  typeof packageManager === "string" &&
  /^pnpm@(\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?)(?:\+sha\d+\.[0-9a-f]+)?$/.exec(
    packageManager
  );
if (!pin) {
  throw new Error(
    "The root package.json must pin packageManager to an exact pnpm version."
  );
}

const args = [
  "exec",
  "--yes",
  `--package=pnpm@${pin[1]}`,
  "--",
  "pnpm",
  ...process.argv.slice(2)
];
let command = "npm";

// Windows cannot spawn npm.cmd without a shell. Run npm's JavaScript entry
// point with Node instead, so paths and arguments never need shell escaping.
if (process.platform === "win32") {
  const directories = [
    ...(process.env.PATH || "").split(path.delimiter).filter(Boolean),
    path.dirname(process.execPath)
  ];
  const candidates = directories.map((directory) =>
    path.join(
      directory.replace(/^"|"$/g, ""),
      "node_modules",
      "npm",
      "bin",
      "npm-cli.js"
    )
  );
  if (process.env.npm_execpath?.endsWith("npm-cli.js")) {
    candidates.unshift(process.env.npm_execpath);
  }
  const npmCli = candidates.find((candidate) => existsSync(candidate));
  if (!npmCli) {
    throw new Error(
      "Cannot find npm-cli.js. Install Node.js with npm and add it to PATH."
    );
  }
  command = process.execPath;
  args.unshift(npmCli);
}

const result = spawnSync(command, args, { stdio: "inherit" });
if (result.error) {
  throw result.error;
}
if (result.signal) {
  process.kill(process.pid, result.signal);
}
process.exitCode = result.status ?? 1;
