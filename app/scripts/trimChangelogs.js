const fsPromises = require("fs").promises;
const path = require("path");

// Trim changelogs, keeping only the top-level changelogs/v{version}/ directory. Prerelease builds delete the whole changelogs directory.
// The version must match package.json and kernel/util/working.go Ver because the kernel reads this path.
async function trimChangelogs(changelogsDir, version) {
  let exists = true;
  try {
    await fsPromises.access(changelogsDir);
  } catch {
    exists = false;
  }

  if (version.includes("-")) {
    // Prerelease build
    if (exists) {
      await fsPromises.rm(changelogsDir, {recursive: true, force: true});
    }
    return {ok: true};
  }

  if (!exists) {
    return {ok: false, reason: "directory not found"};
  }

  const verName = `v${version}`;
  const destDir = path.join(changelogsDir, verName);
  const entries = await fsPromises.readdir(changelogsDir, {withFileTypes: true});
  let hasTarget = false;
  const obsolete = [];

  for (const entry of entries) {
    if (!entry.isDirectory()) {
      continue;
    }
    if (entry.name === verName) {
      hasTarget = true;
    } else {
      obsolete.push(entry.name);
    }
  }

  if (!hasTarget) {
    return {ok: false, reason: `changelog directory not found: ${destDir}`};
  }

  await Promise.all(
    obsolete.map((name) => fsPromises.rm(path.join(changelogsDir, name), {recursive: true, force: true}))
  );
  return {ok: true, path: destDir};
}

module.exports = { trimChangelogs };

// CLI: node scripts/trimChangelogs.js, used by package builds.
if (require.main === module) {
  const changelogsDir = path.resolve(process.cwd(), "changelogs");
  const version = require(path.join(__dirname, "..", "package.json")).version;
  trimChangelogs(changelogsDir, version).then((result) => {
    if (!result.ok) {
      console.error(`trimChangelogs: ${result.reason}`);
      return;
    }
    if (result.path) {
      console.log(`trimChangelogs: ${result.path}`);
    }
  }).catch((error) => {
    console.error("trimChangelogs failed:", error.message);
  });
}
