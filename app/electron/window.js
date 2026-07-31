// Shared script for desktop pre-boot windows (init.html / workspace.html).
// Included by each HTML file through <script src="window.js"></script>; depends on nodeIntegration: true.
// External support, community, and download links are disabled.
"use strict";

// Parse URL query parameters.
const getSearch = (key) => {
    if (window.location.search.indexOf("?") === -1) {
        return "";
    }
    let value = "";
    const data = window.location.search.split("?")[1].split("&");
    data.find(item => {
        const keyValue = item.split("=");
        if (keyValue[0] === key) {
            value = keyValue[1];
            return true;
        }
    });
    return value;
};

// English copy shared by all pre-boot windows.
const I18N_BASE = {
    en: {
        title: "Scribli",
        crashTip: "⚠️ A renderer process previously exited unexpectedly. This may be related to plugins, code snippets, or a custom theme and icon. Starting in safe mode is recommended. Safe mode disables all plugins and code snippets and switches to the default theme and icon. Related content is not deleted, but these settings must be restored manually after startup.",
        safeModeBtn: "🛡️ Start in safe mode",
        normalBtn: "Start normally",
        slogan: "Refactor your thinking",
        wsTitle: "Workspaces",
        missingTip: "⚠️ The last opened workspace path could not be found",
        emptyHint: "No available workspaces. Please select a new workspace path.",
        selectPath: "🗂️ Select a new workspace path",
        selectPathDesc: "The workspace stores your data. You can switch workspaces later from the top bar menu.",
        workspace: "🗂️ Workspace",
        workspaceDesc: "The workspace stores your data. You can switch workspaces later from the top bar menu.",
        notice: "⚠️ Do not use a third-party sync disk to sync data. This can cause abnormal behavior and data damage.",
        open: "Open",
        selectBtn: "Select",
        lang: "🌐 Language",
        langDesc: "Scribli currently ships with English interface text only.",
        msgPartitionRoot: "⚠️ Do not create the workspace at the partition root. Create a new folder for the workspace.",
        msgNotEmpty: "⚠️ This folder contains other files. Create a new folder for the workspace.",
        msgICloud: "⚠️ This folder is under the iCloud sync path. Choose another path.",
        msgCloudDrive: "⚠️ The folder path cannot contain a third-party sync disk name such as onedrive, dropbox, google drive, or pcloud. Choose another path.",
        msgConfirm: "⚠️ Please confirm that the workspace is not set under a third-party sync disk path, otherwise it can cause data damage. Continue?",
    },
};;

// Current interface language, set by each HTML file.
let currentLang = "en";

// Apply the specified language copy to the DOM and return the current language object.
const applyLang = () => {
    const langData = I18N_BASE.en;
    document.title = `${langData.title} v${getSearch("v")}`;
    document.querySelectorAll("[data-i18n]").forEach(item => {
        const key = item.getAttribute("data-i18n");
        if (langData[key]) {
            item.textContent = langData[key];
        }
    });
    // Copy containing HTML tags such as <kbd> uses innerHTML.
    document.querySelectorAll("[data-i18n-html]").forEach(item => {
        const key = item.getAttribute("data-i18n-html");
        if (langData[key]) {
            item.innerHTML = langData[key];
        }
    });
    currentLang = "en";
    return langData;
};

// Workspace path validation helpers.
const isPartitionRootPath = (absPath) => {
    const path = require("path");
    return path.parse(absPath).root === absPath;
};

const isEmptyDir = (absPath) => {
    const fs = require("fs");
    let files;
    try {
        files = fs.readdirSync(absPath).filter(file => file !== ".DS_Store");
    } catch (err) {
        return false;
    }
    return 0 === files.length;
};

const isWorkspaceDir = (absPath) => {
    const path = require("path");
    const fs = require("fs");
    const conf = path.join(absPath, "conf", "conf.json");
    let data;
    try {
        data = fs.readFileSync(conf, "utf8");
    } catch (err) {
        return false;
    }
    return data.includes("kernelVersion");
};

const isCloudDrivePath = (absPath) => {
    const absPathLower = absPath.toLowerCase();
    return -1 < absPathLower.indexOf("onedrive") || -1 < absPathLower.indexOf("dropbox") ||
        -1 < absPathLower.indexOf("google drive") || -1 < absPathLower.indexOf("pcloud");
};

// Check whether a macOS workspace is placed under an iCloud path.
// 
const isICloudPath = (absPath) => {
    const os = require("os");
    if ("darwin" !== os.platform()) {
        return false;
    }
    const path = require("path");
    const homePath = decodeURIComponent(getSearch("home"));
    const absPathLower = absPath.toLowerCase();
    const iCloudRoot = path.join(homePath, "Library", "Mobile Documents");
    if (!simpleCheckIcloudPath(absPath, homePath)) {
        // Fall back to a deeper check when the simple check cannot rule it out.
        const allFiles = walk(iCloudRoot);
        for (const file of allFiles) {
            if (-1 < absPathLower.indexOf(file.toLowerCase())) {
                return true;
            }
        }
    }
    return false;
};

// Simple iCloud sync directory check.
// Reject Desktop, Documents, iCloud folders, and symbolic links.
const simpleCheckIcloudPath = (absPath, homePath) => {
    const fs = require("fs");
    const path = require("path");
    let stat = fs.lstatSync(absPath);
    if (stat.isSymbolicLink()) {
        return false;
    }
    const absPathLower = absPath.toLowerCase();
    const iCloudRoot = path.join(homePath, "Library", "Mobile Documents");
    if (absPathLower.startsWith(iCloudRoot.toLowerCase())) {
        return false;
    }
    const documentsRoot = path.join(homePath, "Documents");
    if (absPathLower.startsWith(documentsRoot.toLowerCase())) {
        return false;
    }
    const desktopRoot = path.join(homePath, "Desktop");
    if (absPathLower.startsWith(desktopRoot.toLowerCase())) {
        return false;
    }
    return true;
};

const walk = (dir, files = []) => {
    const fs = require("fs");
    const path = require("path");
    let dirFiles;
    try {
        if (!fs.existsSync(dir)) {
            console.log("dir [" + dir + "] not exists");
            return files;
        }
        dirFiles = fs.readdirSync(dir);
    } catch (e) {
        console.error("read dir [" + dir + "] failed: ", e);
        return files;
    }
    for (const f of dirFiles) {
        let stat = fs.lstatSync(dir + path.sep + f);
        if (stat.isSymbolicLink()) {
            files.push(fs.readlinkSync(dir + path.sep + f));
            continue;
        }
        if (stat.isDirectory()) {
            // Skip directories that have already been walked.
            if (files.includes(dir + path.sep + f)) {
                continue;
            }
            files.push(dir + path.sep + f);
            walk(dir + path.sep + f, files);
        }
    }
    return files;
};

// Choose a workspace directory and validate it; return null on cancel or validation failure.
const chooseWorkspacePath = async (langData) => {
    const path = require("path");
    const fs = require("fs");
    const {ipcRenderer} = require("electron");

    let defaultWorkspace = path.join(decodeURIComponent(getSearch("home")), "Scribli");
    if ("darwin" === process.platform) {
        // Change the initial workspace path to ~/Library/Application Support/Scribli on macOS 
        defaultWorkspace = path.join(decodeURIComponent(getSearch("home")), "Library", "Application Support", "Scribli");
    }
    if (!fs.existsSync(defaultWorkspace)) {
        fs.mkdirSync(defaultWorkspace, {mode: 0o755, recursive: true});
    }

    const result = await ipcRenderer.invoke("scribli-get", {
        cmd: "showOpenDialog",
        defaultPath: defaultWorkspace,
        properties: ["openDirectory", "createDirectory"],
    });

    if (result.canceled) {
        return null;
    }
    const initPath = result.filePaths[0];

    if (isPartitionRootPath(initPath)) {
        alert(langData.msgPartitionRoot);
        return null;
    }
    if (!isWorkspaceDir(initPath) && !isEmptyDir(initPath)) {
        alert(langData.msgNotEmpty);
        return null;
    }
    if (isICloudPath(initPath)) {
        alert(langData.msgICloud);
        return null;
    }
    if (isCloudDrivePath(initPath)) {
        alert(langData.msgCloudDrive);
        return null;
    }
    if (!confirm(langData.msgConfirm)) {
        return null;
    }
    if (!fs.existsSync(initPath)) {
        fs.mkdirSync(initPath, {mode: 0o755, recursive: true});
    }
    return initPath;
};

// Shared window initialization: macOS body class plus close/minimize button IPC.
const initWindowChrome = () => {
    const {ipcRenderer} = require("electron");
    if ("darwin" === process.platform) {
        document.body.classList.add("darwin");
    }
    document.getElementById("close").addEventListener("click", () => {
        ipcRenderer.send("scribli-first-quit");
    });
    document.getElementById("min").addEventListener("click", () => {
        ipcRenderer.send("scribli-cmd", "minimize");
    });
};

window.getSearch = getSearch;
window.I18N_BASE = I18N_BASE;
window.applyLang = applyLang;
window.currentLang = currentLang;
window.isPartitionRootPath = isPartitionRootPath;
window.isEmptyDir = isEmptyDir;
window.isWorkspaceDir = isWorkspaceDir;
window.isCloudDrivePath = isCloudDrivePath;
window.isICloudPath = isICloudPath;
window.simpleCheckIcloudPath = simpleCheckIcloudPath;
window.walk = walk;
window.chooseWorkspacePath = chooseWorkspacePath;
window.initWindowChrome = initWindowChrome;

