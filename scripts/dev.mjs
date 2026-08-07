import { spawn, spawnSync } from "node:child_process";
import { mkdirSync } from "node:fs";
import { resolve } from "node:path";
import process from "node:process";

const isWindows = process.platform === "win32";
const executableExtension = isWindows ? ".exe" : "";

const rootDirectory = process.cwd();
const temporaryDirectory = resolve(rootDirectory, "tmp");

const apiExecutable = resolve(
  temporaryDirectory,
  `aksescheck-api-dev${executableExtension}`,
);

const workerExecutable = resolve(
  temporaryDirectory,
  `aksescheck-worker-dev${executableExtension}`,
);

const nextExecutable = resolve(
  rootDirectory,
  "apps",
  "web",
  "node_modules",
  "next",
  "dist",
  "bin",
  "next",
);

const developmentEnvironment = {
  ...process.env,
  API_ADDR: "127.0.0.1:8080",
  WEB_ORIGIN: "http://127.0.0.1:3000",
  NEXT_PUBLIC_API_URL: "http://127.0.0.1:8080/api/v1",
  NEXT_PUBLIC_CSRF_COOKIE_NAME: "aksesibilitaswebsite_session_csrf",
};

mkdirSync(temporaryDirectory, {
  recursive: true,
});

function build(name, output, packagePath) {
  console.log(`[DEV] Building ${name}...`);

  const result = spawnSync("go", ["build", "-o", output, packagePath], {
    cwd: rootDirectory,
    stdio: "inherit",
    shell: false,
    env: developmentEnvironment,
  });

  if (result.error) {
    console.error(`[DEV] Gagal menjalankan Go untuk ${name}:`, result.error);
    process.exit(1);
  }

  if (result.status !== 0) {
    process.exit(result.status ?? 1);
  }
}

build("API", apiExecutable, "./services/api/cmd/api");

build("worker", workerExecutable, "./services/worker/cmd/worker");

const processes = new Map();

let shuttingDown = false;
let exitCode = 0;
let forceShutdownTimer;

function startProcess(name, command, args) {
  const child = spawn(command, args, {
    cwd: rootDirectory,
    stdio: "inherit",
    shell: false,
    windowsHide: false,
    env: developmentEnvironment,
  });

  processes.set(name, child);

  child.on("error", (error) => {
    console.error(`[${name}] gagal dijalankan:`, error);

    exitCode = 1;

    if (!shuttingDown) {
      shutdown(false);
    }
  });

  child.on("exit", (code, signal) => {
    processes.delete(name);

    if (!shuttingDown) {
      if (code !== 0 && code !== null) {
        console.error(`[${name}] berhenti dengan exit code ${code}`);
        exitCode = code;
      } else if (signal) {
        console.error(`[${name}] berhenti karena signal ${signal}`);
      } else {
        console.log(`[${name}] berhenti`);
      }

      shutdown(false);
      return;
    }

    if (processes.size === 0) {
      finish();
    }
  });

  return child;
}

function killChildren() {
  for (const child of processes.values()) {
    if (child.exitCode !== null || child.signalCode !== null) {
      continue;
    }

    try {
      child.kill("SIGTERM");
    } catch {}
  }
}

function forceKillChildren() {
  for (const child of processes.values()) {
    if (child.exitCode !== null || child.signalCode !== null) {
      continue;
    }

    try {
      child.kill("SIGKILL");
    } catch {}
  }
}

function shutdown(fromConsoleInterrupt) {
  if (shuttingDown) {
    return;
  }

  shuttingDown = true;

  console.log("\n[DEV] Menghentikan semua service...");

  if (!isWindows || !fromConsoleInterrupt) {
    killChildren();
  }

  forceShutdownTimer = setTimeout(() => {
    if (processes.size > 0) {
      console.log("[DEV] Memaksa service yang belum berhenti...");
      forceKillChildren();
    }
  }, 8000);

  forceShutdownTimer.unref();

  if (processes.size === 0) {
    finish();
  }
}

function finish() {
  if (forceShutdownTimer) {
    clearTimeout(forceShutdownTimer);
  }

  console.log("[DEV] Semua service sudah berhenti.");

  process.exit(exitCode);
}

process.on("SIGINT", () => {
  shutdown(true);
});

process.on("SIGTERM", () => {
  shutdown(false);
});

process.on("uncaughtException", (error) => {
  console.error(error);
  exitCode = 1;
  shutdown(false);
});

process.on("unhandledRejection", (error) => {
  console.error(error);
  exitCode = 1;
  shutdown(false);
});

startProcess("API", apiExecutable, []);

startProcess("WORKER", workerExecutable, []);

startProcess("WEB", process.execPath, [
  nextExecutable,
  "dev",
  "apps/web",
  "--hostname",
  "127.0.0.1",
]);
