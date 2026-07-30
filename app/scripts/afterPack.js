const fsPromises = require("fs").promises;
const path = require("path");
const { trimChangelogs } = require("./trimChangelogs");

module.exports = async function afterPack(context) {
  const {appOutDir, electronPlatformName, packager} = context;
  await removeLanguagePacks(appOutDir, electronPlatformName);
  await trimPackagedChangelogs(appOutDir, packager);
};

// Keep only the packaged app version's changelog.
async function trimPackagedChangelogs(appOutDir, packager) {
  const changelogsDir = path.join(appOutDir, "resources", "changelogs");

  try {
    const result = await trimChangelogs(changelogsDir, packager.appInfo.version);
    if (!result.ok) {
      console.error(`trimChangelogs: ${result.reason}`);
      return;
    }
    if (result.path) {
      console.log(`trimChangelogs: ${result.path}`);
    }
  } catch (error) {
    console.error("Failed to trim changelogs:", error.message);
  }
}

async function removeLanguagePacks(appOutDir, platform) {
  if (platform !== "win32") {
    return;
  }

  const wantedLanguages = ["ar", "de", "en", "es", "fr", "he", "hi", "id", "it", "ja", "ko", "nl", "pl", "pt-BR", "ru", "sk", "th", "tr", "uk", "zh-TW", "zh-CN"];
  const keepPrefixes = new Set(wantedLanguages.map(lang => lang.substring(0, 2)));
  const resourcePath = path.join(appOutDir, "locales");
  const fileExtension = ".pak";

  try {
    const entries = await fsPromises.readdir(resourcePath);
    const targetFiles = entries.filter(file => file.endsWith(fileExtension));

    if (targetFiles.length === 0) {
      return;
    }

    let deletedCount = 0;
    let deletedSize = 0;
    const deletePromises = entries.map(async (file) => {
      if (!file.endsWith(fileExtension)) return;

      const languageName = file.replace(new RegExp(`\\${fileExtension}$`), "");
      const langPrefix = languageName.substring(0, 2);

      if (keepPrefixes.has(langPrefix)) {
        return;
      }

      const filePath = path.join(resourcePath, file);

      const stats = await fsPromises.stat(filePath);
      const fileSize = stats.size;

      await fsPromises.rm(filePath, {
        force: true,
      });

      deletedCount++;
      deletedSize += fileSize;
    });

    await Promise.all(deletePromises);

    if (deletedCount > 0) {
      console.log(`Removed ${deletedCount}/${targetFiles.length} language packs, saved ${formatBytes(deletedSize)}`);
    }
  } catch (error) {
    console.error("Failed to remove language packs:", error.message);
  }
}

function formatBytes(bytes) {
  if (bytes === 0) {
    return "0 B";
  }

  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB", "TB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  const size = bytes / Math.pow(k, i);

  const formattedSize = size % 1 === 0 ? size.toString() : size.toFixed(1);

  return `${formattedSize} ${sizes[i]}`;
}

