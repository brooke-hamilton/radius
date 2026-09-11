const assert = require("node:assert/strict");
const { spawnSync } = require("node:child_process");
const { readFileSync } = require("node:fs");
const path = require("node:path");
const { test } = require("node:test");
const { runInNewContext } = require("node:vm");

const launcher = path.join(__dirname, "pnpm.cjs");
const source = readFileSync(launcher, "utf8");
const root = path.resolve(__dirname, "../..");
const pinnedVersion = require("../../package.json")
  .packageManager.split("@")[1]
  .split("+")[0];

function simulate({
  platform = "linux",
  packageManager = "pnpm@1.2.3",
  env = {},
  files = [],
  args = [],
  result = { status: 0 },
  execPath = "C:\\Program Files\\nodejs\\node.exe"
} = {}) {
  const calls = [];
  const signals = [];
  const process = {
    platform,
    env,
    execPath,
    argv: ["node", launcher, ...args],
    pid: 123,
    kill: (...args) => signals.push(args)
  };
  const modules = {
    "node:child_process": {
      spawnSync: (command, args, options) => {
        calls.push({ command, args: [...args], options: { ...options } });
        return result;
      }
    },
    "node:fs": { existsSync: (file) => files.includes(file) },
    "node:path": platform === "win32" ? path.win32 : path.posix,
    "../../package.json": { packageManager }
  };
  runInNewContext(source, {
    process,
    require: (name) => {
      assert.ok(Object.hasOwn(modules, name), `Unexpected module: ${name}`);
      return modules[name];
    }
  });
  return { calls, signals, exitCode: process.exitCode };
}

for (const platform of ["darwin", "linux"]) {
  test(`${platform}: forwards the exact pin and arguments without a shell`, () => {
    const args = [
      "exec",
      "tsp",
      "format",
      "**/*.tsp",
      "path with spaces",
      "a&b"
    ];
    const { calls, exitCode } = simulate({ platform, args });
    assert.deepEqual(calls, [
      {
        command: "npm",
        args: ["exec", "--yes", "--package=pnpm@1.2.3", "--", "pnpm", ...args],
        options: { stdio: "inherit" }
      }
    ]);
    assert.equal(exitCode, 0);
  });
}

for (const packageManager of [
  "pnpm@2.3.4",
  "pnpm@2.3.4-rc.1",
  "pnpm@2.3.4+sha512.abcdef"
]) {
  test(`uses the manifest pin ${packageManager}`, () => {
    const { calls } = simulate({ packageManager });
    assert.equal(calls[0].args[2], `--package=${packageManager.split("+")[0]}`);
  });
}

for (const packageManager of [
  null,
  {},
  "",
  "npm@1.2.3",
  "pnpm",
  "pnpm@latest",
  "pnpm@^1.2.3",
  "pnpm@1",
  "pnpm@1.2.3 && echo unexpected"
]) {
  test(`rejects an unpinned packageManager: ${JSON.stringify(packageManager)}`, () => {
    assert.throws(
      () => simulate({ packageManager }),
      /root package.json must pin packageManager/
    );
  });
}

for (const { name, env, npmCli } of [
  {
    name: "bundled npm",
    env: {},
    npmCli: "C:\\Program Files\\nodejs\\node_modules\\npm\\bin\\npm-cli.js"
  },
  {
    name: "npm on PATH before bundled npm",
    env: { PATH: '"C:\\Users\\Developer Name\\npm";C:\\Program Files\\nodejs' },
    npmCli: "C:\\Users\\Developer Name\\npm\\node_modules\\npm\\bin\\npm-cli.js"
  },
  {
    name: "npm lifecycle entry point",
    env: { npm_execpath: "C:\\custom npm\\npm-cli.js" },
    npmCli: "C:\\custom npm\\npm-cli.js"
  }
]) {
  test(`Windows: uses ${name} without invoking npm.cmd`, () => {
    const args = ["exec", "node", "a file.js", "--flag=a&b", "**/*.tsp"];
    const { calls } = simulate({
      platform: "win32",
      env,
      args,
      files: [
        npmCli,
        "C:\\Program Files\\nodejs\\node_modules\\npm\\bin\\npm-cli.js"
      ]
    });
    assert.deepEqual(calls, [
      {
        command: "C:\\Program Files\\nodejs\\node.exe",
        args: [
          npmCli,
          "exec",
          "--yes",
          "--package=pnpm@1.2.3",
          "--",
          "pnpm",
          ...args
        ],
        options: { stdio: "inherit" }
      }
    ]);
  });
}

test("Windows: reports missing npm instead of falling back to a shell", () => {
  assert.throws(
    () => simulate({ platform: "win32", env: { npm_execpath: "pnpm.cjs" } }),
    /Cannot find npm-cli.js/
  );
});

test("propagates npm exit codes, spawn failures and termination signals", () => {
  assert.equal(simulate({ result: { status: 23 } }).exitCode, 23);
  const error = new Error("npm was not found");
  assert.throws(
    () => simulate({ result: { error } }),
    (value) => value === error
  );
  const { signals, exitCode } = simulate({
    result: { status: null, signal: "SIGTERM" }
  });
  assert.deepEqual(signals, [[123, "SIGTERM"]]);
  assert.equal(exitCode, 1);
});

function run(args, cwd = root) {
  const result = spawnSync(process.execPath, [launcher, ...args], {
    cwd,
    encoding: "utf8",
    timeout: 120_000
  });
  assert.ifError(result.error);
  return result;
}

test("runs the pinned pnpm from the root and a nested TypeSpec directory", () => {
  for (const cwd of [root, path.join(root, "typespec", "Applications.Core")]) {
    const result = run(["--version"], cwd);
    assert.equal(result.status, 0, result.stderr);
    assert.equal(result.stdout.trim(), pinnedVersion);
  }
});

test("preserves the working directory, -C, and literal arguments", () => {
  const args = [
    "two words",
    "**/*.tsp",
    "a&b",
    'a"b',
    "a'b",
    "%PATH%",
    "$HOME"
  ];
  const program =
    "console.log(JSON.stringify({cwd:process.cwd(),args:process.argv.slice(1)}))";
  const typespec = path.join(root, "typespec");
  const service = path.join(typespec, "Applications.Core");
  for (const { cwd, prefix, expectedCwd } of [
    { cwd: service, prefix: [], expectedCwd: service },
    { cwd: root, prefix: ["-C", typespec], expectedCwd: typespec }
  ]) {
    const result = run(
      [
        "--reporter=silent",
        ...prefix,
        "exec",
        "node",
        "-e",
        program,
        "--",
        ...args
      ],
      cwd
    );
    assert.equal(result.status, 0, result.stderr || result.stdout);
    assert.deepEqual(JSON.parse(result.stdout), { cwd: expectedCwd, args });
  }
});

test("returns a failing pnpm child command's exit status", () => {
  const result = run(["exec", "node", "-e", "process.exit(23)"]);
  assert.equal(result.status, 23, result.stderr);
});
