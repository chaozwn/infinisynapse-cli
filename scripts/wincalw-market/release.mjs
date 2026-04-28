#!/usr/bin/env node

import fs from "node:fs/promises";
import { createReadStream } from "node:fs";
import crypto from "node:crypto";
import os from "node:os";
import path from "node:path";
import process from "node:process";
import { spawnSync } from "node:child_process";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const projectDir = path.dirname(__filename);
const repoRoot = path.resolve(projectDir, "..", "..");
const versionFile = path.join(projectDir, "VERSION");
const modulePath = "github.com/chaozwn/infinisynapse-cli";

const platforms = [
  { name: "windows", goos: "windows", goarch: "amd64", os: "windows", arch: "x64", archive: "zip" },
  { name: "windows-arm64", goos: "windows", goarch: "arm64", os: "windows", arch: "arm64", archive: "zip" },
  { name: "mac-arm64", goos: "darwin", goarch: "arm64", os: "mac", arch: "arm64", archive: "tar.gz" },
  { name: "mac-x64", goos: "darwin", goarch: "amd64", os: "mac", arch: "x64", archive: "tar.gz" },
  { name: "linux", goos: "linux", goarch: "amd64", os: "linux", arch: "x64", archive: "tar.gz" },
  { name: "linux-arm64", goos: "linux", goarch: "arm64", os: "linux", arch: "arm64", archive: "tar.gz" },
];

function parseArgs(argv) {
  const opts = {
    version: "",
    releasesDir: path.join(projectDir, "releases"),
    artifactPrefix: "agent_infini",
    appName: "agent_infini",
    publish: false,
    provider: "all",
    uploadPrefix: "tools",
    uploadPrefixExplicit: false,
    dryRun: false,
    skipBuild: false,
    bump: "patch",
    versionExplicit: false,
    envFiles: [],
  };

  for (let i = 0; i < argv.length; i++) {
    const arg = argv[i];
    if (arg === "--help" || arg === "-h") {
      printHelp();
      process.exit(0);
    }
    if (arg === "--publish") {
      opts.publish = true;
      continue;
    }
    if (arg === "--dry-run") {
      opts.dryRun = true;
      opts.publish = true;
      continue;
    }
    if (arg === "--skip-build") {
      opts.skipBuild = true;
      continue;
    }
    if (arg === "--no-bump") {
      opts.bump = "none";
      continue;
    }

    const valueFlags = new Set([
      "--version",
      "--releases-dir",
      "--artifact-prefix",
      "--app-name",
      "--provider",
      "--upload-prefix",
      "--prefix",
      "--bump",
      "--env-file",
    ]);
    if (valueFlags.has(arg)) {
      if (i + 1 >= argv.length) throw new Error(`${arg} requires a value`);
      setOption(opts, arg, argv[++i]);
      continue;
    }

    const eq = arg.indexOf("=");
    if (eq > 0) {
      const key = arg.slice(0, eq);
      const value = arg.slice(eq + 1);
      if (valueFlags.has(key)) {
        setOption(opts, key, value);
        continue;
      }
    }

    throw new Error(`Unknown option: ${arg}`);
  }

  opts.provider = normalizeProvider(opts.provider || process.env.STORAGE_TYPE || "all");
  opts.releasesDir = path.resolve(opts.releasesDir);
  return opts;
}

function setOption(opts, key, value) {
  switch (key) {
    case "--version":
      opts.version = value;
      opts.versionExplicit = true;
      break;
    case "--releases-dir":
      opts.releasesDir = value;
      break;
    case "--artifact-prefix":
      opts.artifactPrefix = value;
      break;
    case "--app-name":
      opts.appName = value;
      break;
    case "--provider":
      opts.provider = value;
      break;
    case "--upload-prefix":
    case "--prefix":
      opts.uploadPrefix = value;
      opts.uploadPrefixExplicit = true;
      break;
    case "--bump":
      opts.bump = normalizeBump(value);
      break;
    case "--env-file":
      opts.envFiles.push(value);
      break;
  }
}

function printHelp() {
  console.log(`release.mjs - Build and optionally publish agent_infini Go CLI.

Usage:
  node release.mjs [options]
  node release.mjs
  node release.mjs --publish
  node release.mjs --publish --provider oss --bump minor

Build options:
  --version VERSION           Manual version. Disables automatic bump.
  --bump patch|minor|major|none
                              Auto bump type when --version is not set. Default: patch.
  --no-bump                   Alias of --bump none.
  --releases-dir DIR          Output directory. Default: ./releases
  --artifact-prefix NAME      Archive file prefix. Default: agent_infini
  --app-name NAME             Binary name. Default: agent_infini
  --skip-build                Reuse existing releases directory.
  --env-file FILE             Load env file before defaults. Can be repeated.

Publish options:
  --publish                   Upload binaries after build. Default uploads to Blob and OSS.
  --provider all|blob|oss     Upload provider. Default: $STORAGE_TYPE or all.
  --upload-prefix PREFIX      Tool object prefix. Default: tools
  --prefix PREFIX             Alias of --upload-prefix.
  --dry-run                   Publish dry-run; implies --publish.

Vercel Blob env:
  BLOB_READ_WRITE_TOKEN

Aliyun OSS env:
  OSS_BUCKET
  OSS_REGION or OSS_ENDPOINT
  OSS_ACCESS_KEY_ID
  OSS_ACCESS_KEY_SECRET
  OSS_PREFIX                  Optional extra prefix. Default: ac-aibot
  OSS_PUBLIC_BASE_URL         Optional public URL base.
`);
}

function normalizeProvider(provider) {
  const value = String(provider || "").toLowerCase();
  if (value === "vercel_blob" || value === "vercel-blob") return "blob";
  if (value === "aliyun-oss" || value === "ali-oss") return "oss";
  if (value === "both") return "all";
  if (value === "all" || value === "blob" || value === "oss") return value;
  throw new Error(`Unsupported provider: ${provider}`);
}

function normalizePrefix(prefix) {
  return String(prefix || "")
    .replace(/^\/+/, "")
    .replace(/\/+$/, "");
}

function normalizeBump(value) {
  const normalized = String(value || "").toLowerCase();
  if (["patch", "minor", "major", "none"].includes(normalized)) return normalized;
  throw new Error(`Unsupported bump type: ${value}`);
}

function expandHomePath(filePath) {
  const value = String(filePath || "");
  if (value === "~") return os.homedir();
  if (value.startsWith("~/")) return path.join(os.homedir(), value.slice(2));
  return value;
}

async function loadEnvFiles(extraFiles = []) {
  const candidates = [
    ...extraFiles,
    process.env.AGENT_INFINI_RELEASE_ENV_FILE,
    process.env.ACCHAT_RELEASE_ENV_FILE,
    "~/projects/ac-aibot.com/.env.china.local",
    "~/projects/ac-aibot.com/.env.local",
    "~/projects/ac-aibot.com/.env.global.local",
    path.join(repoRoot, ".env.local"),
    path.join(projectDir, ".env.local"),
  ].filter(Boolean);
  for (const filePath of candidates) {
    await loadEnvFile(expandHomePath(filePath));
  }
}

async function loadEnvFile(filePath) {
  let text;
  try {
    text = await fs.readFile(filePath, "utf8");
  } catch (err) {
    if (err?.code === "ENOENT") return;
    throw err;
  }

  for (const rawLine of text.split(/\r?\n/)) {
    const line = rawLine.trim();
    if (!line || line.startsWith("#")) continue;
    const match = line.match(/^(?:export\s+)?([A-Za-z_][A-Za-z0-9_]*)\s*=\s*(.*)$/);
    if (!match) continue;
    const [, key, rawValue] = match;
    if (process.env[key] !== undefined) continue;
    process.env[key] = parseEnvValue(rawValue);
  }
}

function parseEnvValue(value) {
  let v = value.trim();
  if ((v.startsWith('"') && v.endsWith('"')) || (v.startsWith("'") && v.endsWith("'"))) {
    v = v.slice(1, -1);
  }
  return v.replace(/\\n/g, "\n");
}

function run(command, args, options = {}) {
  const result = spawnSync(command, args, {
    cwd: options.cwd || repoRoot,
    stdio: options.stdio || "inherit",
    env: options.env || process.env,
  });
  if (result.error) throw result.error;
  if (result.status !== 0) {
    throw new Error(`${command} ${args.join(" ")} failed with exit code ${result.status}`);
  }
  return result;
}

function commandOutput(command, args, options = {}) {
  const result = spawnSync(command, args, {
    cwd: options.cwd || repoRoot,
    encoding: "utf8",
    stdio: ["ignore", "pipe", "pipe"],
  });
  if (result.status !== 0) return "";
  return String(result.stdout || "").trim();
}

async function resolveVersion(opts) {
  if (opts.versionExplicit) {
    opts.version = normalizeVersion(opts.version);
  } else if (opts.skipBuild) {
    opts.version = await readCurrentVersion();
  } else {
    const current = await readCurrentVersion();
    opts.currentVersion = current;
    opts.version = opts.bump === "none" ? current : bumpVersion(current, opts.bump);

    if (opts.dryRun) {
      console.log(`Version preview : ${current} -> ${opts.version} (dry-run, VERSION not updated)`);
    } else if (opts.version !== current) {
      console.log(`Version pending : ${current} -> ${opts.version}`);
    } else {
      console.log(`Version         : ${opts.version} (no bump)`);
    }
  }

  if (!opts.uploadPrefixExplicit) opts.uploadPrefix = "tools";
  opts.uploadPrefix = normalizePrefix(opts.uploadPrefix);
}

async function finalizeVersion(opts) {
  if (opts.dryRun || opts.versionExplicit || opts.skipBuild) return;
  if (!opts.currentVersion || opts.version === opts.currentVersion) return;
  await fs.writeFile(versionFile, `${opts.version}\n`);
  console.log(`Version bumped  : ${opts.currentVersion} -> ${opts.version}`);
}

async function readCurrentVersion() {
  try {
    const text = await fs.readFile(versionFile, "utf8");
    return normalizeVersion(text.trim());
  } catch (err) {
    if (err?.code === "ENOENT") return "v0.9.0";
    throw err;
  }
}

function normalizeVersion(version) {
  const value = String(version || "").trim();
  if (!/^v?\d+\.\d+\.\d+$/.test(value)) {
    throw new Error(`Invalid version "${version}". Expected vMAJOR.MINOR.PATCH, for example v0.9.0`);
  }
  return value.startsWith("v") ? value : `v${value}`;
}

function bumpVersion(version, bump) {
  const normalized = normalizeVersion(version);
  const match = normalized.match(/^v(\d+)\.(\d+)\.(\d+)$/);
  if (!match) throw new Error(`Invalid version: ${version}`);

  let major = Number(match[1]);
  let minor = Number(match[2]);
  let patch = Number(match[3]);

  switch (bump) {
    case "major":
      major += 1;
      minor = 0;
      patch = 0;
      break;
    case "minor":
      minor += 1;
      patch = 0;
      break;
    case "patch":
      patch += 1;
      break;
    case "none":
      break;
    default:
      throw new Error(`Unsupported bump type: ${bump}`);
  }

  return `v${major}.${minor}.${patch}`;
}

async function build(opts) {
  const commit = commandOutput("git", ["rev-parse", "--short", "HEAD"]) || "none";
  const buildDate = new Date().toISOString().replace(/\.\d{3}Z$/, "Z");

  console.log("\x1b[0;34magent_infini WinClaw market build\x1b[0m");
  console.log(`Go version      : ${commandOutput("go", ["version"]) || "<unknown>"}`);
  console.log(`Version         : ${opts.version}`);
  console.log(`Commit          : ${commit}`);
  console.log(`Build date      : ${buildDate}`);
  console.log(`Output dir      : ${opts.releasesDir}`);
  console.log(`Artifact prefix : ${opts.artifactPrefix}`);
  console.log("");

  await fs.rm(opts.releasesDir, { recursive: true, force: true });
  await fs.mkdir(path.join(opts.releasesDir, "packages"), { recursive: true });

  const checksums = [];
  for (const platform of platforms) {
    await buildPlatform(opts, platform, checksums, commit, buildDate);
  }

  await fs.writeFile(
    path.join(opts.releasesDir, "checksums.txt"),
    checksums.map((item) => `${item.sha256}  ${item.relativePath}`).join("\n") + "\n",
  );
  await writeManifest(opts, checksums);

  console.log("");
  console.log(`\x1b[0;32m[ok] Build completed: ${opts.releasesDir}\x1b[0m`);
  console.log("Artifacts:");
  for (const item of checksums) {
    console.log(`  - ${item.relativePath} (${formatBytes(item.sizeBytes)})`);
  }
  console.log("");
}

async function buildPlatform(opts, platform, checksums, commit, buildDate) {
  const outputName = platform.goos === "windows" ? `${opts.appName}.exe` : opts.appName;
  const platformDir = path.join(opts.releasesDir, platform.name);
  await fs.mkdir(platformDir, { recursive: true });

  process.stdout.write(
    `  Building ${platform.name.padEnd(14)} (${platform.goos}/${platform.goarch}) ... `,
  );
  run(
    "go",
    [
      "build",
      "-trimpath",
      "-ldflags",
      [
        "-s",
        "-w",
        `-X ${modulePath}/cmd.Version=${opts.version}`,
        `-X ${modulePath}/cmd.Commit=${commit}`,
        `-X ${modulePath}/cmd.BuildDate=${buildDate}`,
      ].join(" "),
      "-o",
      path.join(platformDir, outputName),
      ".",
    ],
    {
      env: {
        ...process.env,
        GOOS: platform.goos,
        GOARCH: platform.goarch,
        CGO_ENABLED: "0",
      },
      stdio: ["ignore", "ignore", "inherit"],
    },
  );
  process.stdout.write("\x1b[0;32mdone\x1b[0m\n");

  const archiveName = `${opts.artifactPrefix}_${opts.version}_${platform.name}`;
  const archivePath =
    platform.archive === "zip"
      ? path.join(opts.releasesDir, "packages", `${archiveName}.zip`)
      : path.join(opts.releasesDir, "packages", `${archiveName}.tar.gz`);

  if (platform.archive === "zip") {
    run("zip", ["-q", "-9", archivePath, outputName], { cwd: platformDir });
  } else {
    run("tar", ["-C", platformDir, "-czf", archivePath, outputName]);
  }

  const stat = await fs.stat(archivePath);
  const binaryPath = path.join(platformDir, outputName);
  const binaryStat = await fs.stat(binaryPath);
  checksums.push({
    file: path.basename(archivePath),
    relativePath: path.posix.join("packages", path.basename(archivePath)),
    fullPath: archivePath,
    binaryFile: outputName,
    binaryPath,
    toolKey: path.posix.join(platform.os, platform.arch, outputName),
    sha256: await sha256File(archivePath),
    sizeBytes: stat.size,
    binarySizeBytes: binaryStat.size,
  });
}

async function sha256File(filePath) {
  const data = await fs.readFile(filePath);
  return crypto.createHash("sha256").update(data).digest("hex");
}

async function writeManifest(opts, checksums) {
  const manifest = {
    name: opts.appName,
    version: opts.version,
    builtAt: new Date().toISOString(),
    artifacts: checksums.map((item) => ({
      file: item.file,
      binaryFile: item.binaryFile,
      toolKey: item.toolKey,
      sha256: item.sha256,
      sizeBytes: item.sizeBytes,
      binarySizeBytes: item.binarySizeBytes,
    })),
  };
  await fs.writeFile(
    path.join(opts.releasesDir, "manifest.json"),
    JSON.stringify(manifest, null, 2) + "\n",
  );
}

function formatBytes(bytes) {
  if (bytes < 1024) return `${bytes}B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)}K`;
  return `${(bytes / 1024 / 1024).toFixed(1)}M`;
}

async function collectFiles(opts) {
  const files = [];
  const manifestArtifacts = await readManifestArtifacts(opts.releasesDir);

  if (manifestArtifacts.length > 0) {
    for (const artifact of manifestArtifacts) {
      const binaryFile = path.basename(artifact.binaryFile || "");
      const toolKey = String(artifact.toolKey || "");
      if (!binaryFile || !toolKey) continue;
      const fullPath = binaryPathForToolKey(opts.releasesDir, toolKey, binaryFile);
      const stat = await fs.stat(fullPath);
      if (!stat.isFile()) continue;
      files.push({
        fullPath,
        relativePath: toolKey,
        contentType: "application/octet-stream",
        size: stat.size,
      });
    }
  } else {
    for (const platform of platforms) {
      const outputName = platform.goos === "windows" ? `${opts.appName}.exe` : opts.appName;
      const fullPath = path.join(opts.releasesDir, platform.name, outputName);
      try {
        const stat = await fs.stat(fullPath);
        if (!stat.isFile()) continue;
        files.push({
          fullPath,
          relativePath: path.posix.join(platform.os, platform.arch, outputName),
          contentType: "application/octet-stream",
          size: stat.size,
        });
      } catch (err) {
        if (err?.code !== "ENOENT") throw err;
      }
    }
  }

  if (files.length === 0) throw new Error(`No release files found under ${opts.releasesDir}`);
  return files;
}

function binaryPathForToolKey(releasesDir, toolKey, binaryFile) {
  const parts = toolKey.split("/");
  if (parts.length !== 3) throw new Error(`Invalid tool key in manifest: ${toolKey}`);
  const [os, arch] = parts;
  const platform = platforms.find((p) => p.os === os && p.arch === arch);
  if (!platform) throw new Error(`Unknown platform in manifest: ${toolKey}`);
  return path.join(releasesDir, platform.name, binaryFile);
}

async function readManifestArtifacts(releasesDir) {
  try {
    const text = await fs.readFile(path.join(releasesDir, "manifest.json"), "utf8");
    const manifest = JSON.parse(text);
    if (!Array.isArray(manifest.artifacts)) return [];
    return manifest.artifacts.filter(
      (artifact) => artifact && typeof artifact.file === "string",
    );
  } catch (err) {
    if (err?.code === "ENOENT") return [];
    throw err;
  }
}

function objectKey(prefix, relativePath) {
  const key = prefix ? `${prefix}/${relativePath}` : relativePath;
  return key.replaceAll(path.sep, "/");
}

const TOOL_MANIFEST_KEY = "tools/manifest.json";

async function publish(opts) {
  const files = await collectFiles(opts);
  console.log(`Found ${files.length} release file(s):`);
  for (const file of files) console.log(`  - ${file.relativePath}`);
  console.log("");

  const uploaded = [];
  for (const provider of providersFor(opts.provider)) {
    console.log(`Publishing to ${provider}...`);
    const providerUploaded =
      provider === "blob" ? await uploadBlob(files, opts) : await uploadOSS(files, opts);
    uploaded.push(...providerUploaded.map((item) => ({ ...item, provider })));
    console.log("");
  }

  if (uploaded.length > 0) {
    console.log("Uploaded objects:");
    for (const item of uploaded) {
      console.log(`  - [${item.provider}] ${item.key}${item.url ? ` (${item.url})` : ""}`);
    }
  }

  await updateToolManifest(opts);
}

function providersFor(provider) {
  return provider === "all" ? ["blob", "oss"] : [provider];
}

function validatePublishConfig(opts) {
  if (!opts.publish || opts.dryRun) return;
  for (const provider of providersFor(opts.provider)) {
    if (provider === "blob" && !process.env.BLOB_READ_WRITE_TOKEN) {
      throw new Error("BLOB_READ_WRITE_TOKEN is required for --provider blob/all");
    }
    if (provider === "oss") {
      const required = ["OSS_BUCKET", "OSS_ACCESS_KEY_ID", "OSS_ACCESS_KEY_SECRET"];
      for (const name of required) {
        if (!process.env[name]) throw new Error(`${name} is required for --provider oss/all`);
      }
    }
  }
}

async function uploadBlob(files, opts) {
  if (!opts.dryRun && !process.env.BLOB_READ_WRITE_TOKEN) {
    throw new Error("BLOB_READ_WRITE_TOKEN is required for --provider blob");
  }

  const { put } = opts.dryRun ? { put: null } : await import("@vercel/blob");
  const uploaded = [];
  for (const file of files) {
    const key = objectKey(opts.uploadPrefix, file.relativePath);
    if (opts.dryRun) {
      console.log(`[dry-run] blob put ${file.fullPath} -> ${key}`);
      uploaded.push({ key, url: "" });
      continue;
    }
    const data = await fs.readFile(file.fullPath);
    const result = await put(key, data, {
      access: "public",
      addRandomSuffix: false,
      contentType: file.contentType,
      token: process.env.BLOB_READ_WRITE_TOKEN,
      allowOverwrite: true,
    });
    console.log(`[blob] ${file.relativePath} -> ${result.url}`);
    uploaded.push({ key, url: result.url });
  }
  return uploaded;
}

async function updateToolManifest(opts) {
  const toolName = opts.appName;
  const version = opts.version.replace(/^v/, "");
  console.log(`Updating tool manifest: ${toolName} -> v${version}`);
  for (const provider of providersFor(opts.provider)) {
    if (provider === "blob") {
      await updateBlobManifest(toolName, version, opts);
    } else {
      await updateOSSManifest(toolName, version, opts);
    }
  }
}

function mergeToolManifest(manifest, toolName, version) {
  const updated = {
    version: manifest?.version || 1,
    updated_at: new Date().toISOString(),
    tools: { ...(manifest?.tools || {}) },
  };
  updated.tools[toolName] = { version };
  return updated;
}

async function updateBlobManifest(toolName, version, opts) {
  let manifest = { version: 1, updated_at: new Date().toISOString(), tools: {} };
  if (!opts.dryRun && process.env.BLOB_READ_WRITE_TOKEN) {
    try {
      const resp = await fetch(`https://blob.vercel-storage.com/${TOOL_MANIFEST_KEY}`, {
        headers: { Authorization: `Bearer ${process.env.BLOB_READ_WRITE_TOKEN}` },
      });
      if (resp.ok) manifest = await resp.json();
    } catch {
      // Missing or unreadable manifest: create a fresh one.
    }
  }
  const next = mergeToolManifest(manifest, toolName, version);
  if (opts.dryRun) {
    console.log(`[dry-run] blob put ${TOOL_MANIFEST_KEY}: ${JSON.stringify(next)}`);
    return;
  }
  const { put } = await import("@vercel/blob");
  await put(TOOL_MANIFEST_KEY, Buffer.from(JSON.stringify(next, null, 2)), {
    access: "public",
    token: process.env.BLOB_READ_WRITE_TOKEN,
    contentType: "application/json",
    addRandomSuffix: false,
    allowOverwrite: true,
  });
  console.log(`[blob] ${TOOL_MANIFEST_KEY} updated`);
}

async function createOSSClient() {
  const { default: OSS } = await import("ali-oss");
  const endpoint = process.env.OSS_ENDPOINT || undefined;
  const region =
    process.env.OSS_REGION ||
    inferOssRegionFromEndpoint(endpoint) ||
    "oss-cn-hangzhou";

  const required = {
    OSS_BUCKET: process.env.OSS_BUCKET,
    OSS_ACCESS_KEY_ID: process.env.OSS_ACCESS_KEY_ID,
    OSS_ACCESS_KEY_SECRET: process.env.OSS_ACCESS_KEY_SECRET,
  };
  for (const [name, value] of Object.entries(required)) {
    if (!value) throw new Error(`${name} is required for --provider oss`);
  }

  return new OSS({
    bucket: process.env.OSS_BUCKET,
    region,
    endpoint,
    accessKeyId: process.env.OSS_ACCESS_KEY_ID,
    accessKeySecret: process.env.OSS_ACCESS_KEY_SECRET,
  });
}

function inferOssRegionFromEndpoint(endpoint) {
  if (!endpoint) return undefined;
  const match = String(endpoint).match(/(oss-[a-z0-9-]+)\.aliyuncs\.com/i);
  return match?.[1];
}

function ossPublicURL(key) {
  const publicBase = process.env.OSS_PUBLIC_BASE_URL;
  if (publicBase) return `${publicBase.replace(/\/+$/, "")}/${key}`;

  const endpoint = process.env.OSS_ENDPOINT;
  const bucket = process.env.OSS_BUCKET;
  if (endpoint && bucket) {
    const normalized = endpoint.startsWith("http") ? endpoint : `https://${endpoint}`;
    try {
      const url = new URL(normalized);
      return `${url.protocol}//${bucket}.${url.host}/${key}`;
    } catch {
      return "";
    }
  }
  return "";
}

async function uploadOSS(files, opts) {
  const client = opts.dryRun ? null : await createOSSClient();
  const basePrefix = normalizePrefix(
    [process.env.OSS_PREFIX || "ac-aibot", opts.uploadPrefix].filter(Boolean).join("/"),
  );

  const uploaded = [];
  for (const file of files) {
    const key = objectKey(basePrefix, file.relativePath);
    if (opts.dryRun) {
      console.log(`[dry-run] oss put ${file.fullPath} -> ${key}`);
      uploaded.push({ key, url: ossPublicURL(key) });
      continue;
    }
    await client.put(key, createReadStream(file.fullPath), {
      headers: { "Content-Type": file.contentType },
    });
    const url = ossPublicURL(key);
    console.log(`[oss] ${file.relativePath} -> ${url || key}`);
    uploaded.push({ key, url });
  }
  return uploaded;
}

async function updateOSSManifest(toolName, version, opts) {
  const client = opts.dryRun ? null : await createOSSClient();
  const ossPrefix = normalizePrefix(process.env.OSS_PREFIX || "ac-aibot");
  const key = objectKey(ossPrefix, TOOL_MANIFEST_KEY);
  let manifest = { version: 1, updated_at: new Date().toISOString(), tools: {} };

  if (!opts.dryRun) {
    try {
      const result = await client.get(key);
      manifest = JSON.parse(result.content.toString("utf8"));
    } catch {
      // Missing or unreadable manifest: create a fresh one.
    }
  }

  const next = mergeToolManifest(manifest, toolName, version);
  if (opts.dryRun) {
    console.log(`[dry-run] oss put ${key}: ${JSON.stringify(next)}`);
    return;
  }

  await client.put(key, Buffer.from(JSON.stringify(next, null, 2)), {
    headers: { "Content-Type": "application/json" },
  });
  console.log(`[oss] ${key} updated`);
}

async function main() {
  const opts = parseArgs(process.argv.slice(2));
  await loadEnvFiles(opts.envFiles);
  await resolveVersion(opts);
  validatePublishConfig(opts);

  if (opts.skipBuild) {
    console.log("Skipping build (--skip-build).");
  } else {
    await build(opts);
  }

  if (opts.publish) {
    await publish(opts);
    console.log("");
    console.log("\x1b[0;32m[ok] Publish completed\x1b[0m");
  }
  await finalizeVersion(opts);
}

main().catch((err) => {
  console.error(`[error] ${err instanceof Error ? err.message : String(err)}`);
  process.exit(1);
});
