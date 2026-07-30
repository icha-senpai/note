// Scribli - Refactor your thinking
// Copyright (c) 2020-present Scribli
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

// Hide Electron security checklist console messages in development.
// https://www.electronjs.org/docs/latest/tutorial/security
process.env.ELECTRON_DISABLE_SECURITY_WARNINGS = "true";

const {
    net,
    app,
    BrowserWindow,
    Notification,
    shell,
    session,
    Menu,
    MenuItem,
    screen,
    ipcMain,
    clipboard,
    globalShortcut,
    Tray,
    dialog,
    systemPreferences,
    powerMonitor
} = require("electron");
const path = require("path");
const fs = require("fs");
const gNet = require("net");
const childProcess = require("child_process");
const remote = require("@electron/remote/main");
const {
    PRIMARY_PROTOCOL,
    hasAppProtocol,
    normalizeAppProtocolURL,
} = require("./protocol");

process.noAsar = true;
const appDir = path.dirname(app.getAppPath());
const isDevEnv = process.env.NODE_ENV === "development";
const appVer = app.getVersion();
const confDir = path.join(app.getPath("home"), ".config", "scribli");
const windowStatePath = path.join(confDir, "windowState.json");
const appCrashLogPath = path.join(confDir, "app.crash.log");
const appCrashMarkerPath = path.join(confDir, "app.crash.json");
const systemShutdownNone = 0;
const systemShutdownEnding = 1;
const systemShutdownForced = 2;
const systemShutdownExitTimeout = 30000;
const safeModeReasons = new Set(["abnormal-exit", "killed", "crashed", "oom", "memory-eviction"]);
const noSafeModeReasons = new Set(["clean-exit", "launch-failed", "integrity-failure"]);
const expectedRendererExitIds = new Set();
const expectedKernelExitPorts = new Set();
const handledCrashWebContents = new Set();
const kernelProcesses = new Map();
let bootWindow;
let latestActiveWindow;
let firstOpen = false;
let workspaces = []; // workspaceDir, id, port, webContentsId, browserWindow, tray, hideShortcut
let kernelPort = 6806;
let resetWindowStateOnRestart = false;
let openAsHidden = false;
let systemShutdownState = systemShutdownNone;
let gracefulSystemShutdownPromise;
let keepAppOpenDuringSystemShutdown = false;
const openDialogSingletons = new Set();
const isOpenAsHidden = function () {
    return 1 === workspaces.length && openAsHidden;
};

remote.initialize();

// Store Electron data under `Scribli-Electron`.
// app.getPath("userData") creates an empty Scribli directory, so use app.getPath("appData") instead.
// 
app.setPath("userData", path.join(app.getPath("appData"), app.getName() + "-Electron"));

if (process.platform === "win32") {
    // Windows needs AppUserModelId to show the app name and icon correctly.
    // 
    app.setAppUserModelId("app.scribli.desktop");
}

if (!app.requestSingleInstanceLock()) {
    app.quit();
    return;
}

const registerDefaultProtocolClient = (protocol) => {
    if (isDevEnv && process.defaultApp && process.argv.length >= 2) {
        // In development on Windows, pass the Electron executable and main.js explicitly so protocol URLs are not
        // treated as relative paths.
        const mainScript = path.resolve(process.argv[1]);
        if (process.platform === "win32") {
            app.removeAsDefaultProtocolClient(protocol, process.execPath, [mainScript]);
            app.setAsDefaultProtocolClient(protocol, process.execPath, [mainScript]);
        } else {
            app.setAsDefaultProtocolClient(protocol);
        }
        return;
    }

    app.setAsDefaultProtocolClient(protocol);
};

const normalizeIncomingProtocolURL = (url) => {
    return normalizeAppProtocolURL(url);
};

registerDefaultProtocolClient(PRIMARY_PROTOCOL);
writeLog("registered primary protocol [" + PRIMARY_PROTOCOL + "://]");

app.commandLine.appendSwitch("disable-web-security");
app.commandLine.appendSwitch("auto-detect", "false");
app.commandLine.appendSwitch("no-proxy-server");
app.commandLine.appendSwitch("enable-features", "PlatformHEVCDecoderSupport");
app.commandLine.appendSwitch("xdg-portal-required-version", "4");
// Do not auto-upgrade HTTP linked images loaded from local HTTPS pages to HTTPS.
app.commandLine.appendSwitch("disable-features", "AutoupgradeMixedContent");

// Support set Chromium command line arguments on the desktop 
writeLog("app is packaged [" + app.isPackaged + "], command line args [" + process.argv.join(", ") + "]");
let argStart = 1;
if (!app.isPackaged) {
    argStart = 2;
}

for (let i = argStart; i < process.argv.length; i++) {
    let arg = process.argv[i];
    if (arg.startsWith("--workspace=") || arg.startsWith("--openAsHidden") || arg.startsWith("--port=") || arg.startsWith("--safe-mode=") || arg.startsWith("--lang=") || hasAppProtocol(arg)) {
        // Skip built-in arguments.
        if (arg.startsWith("--openAsHidden")) {
            openAsHidden = true;
            writeLog("open as hidden");
        }
        continue;
    }

    app.commandLine.appendSwitch(arg);
    writeLog("command line switch [" + arg + "]");
}

try {
    firstOpen = !fs.existsSync(path.join(confDir, "workspace.json"));
    if (!fs.existsSync(confDir)) {
        fs.mkdirSync(confDir, {mode: 0o755, recursive: true});
    }
} catch (e) {
    console.error(e);
    require("electron").dialog.showErrorBox("Failed to create config directory", "Scribli needs to create a configuration folder (~/.config/scribli) in the user's home directory. Please make sure that the path has write permissions.");
    app.exit();
}

// Parse command-line arguments passed as `name=value`.
// 
const getArg = (name) => {
    for (let i = 0; i < process.argv.length; i++) {
        if (process.argv[i].startsWith(name)) {
            return process.argv[i].split("=")[1];
        }
    }
};

// Detect whether the last opened workspace is missing.
// 
let lastWorkspaceMissing = false;
let missingWorkspacePath = "";
let availableWorkspaces = [];
if (!firstOpen && !getArg("--workspace")) {
    // Respect an explicit command-line workspace and skip the missing-workspace check.
    try {
        const wsFile = path.join(confDir, "workspace.json");
        if (fs.existsSync(wsFile)) {
            const wsList = JSON.parse(fs.readFileSync(wsFile, "utf8"));
            if (Array.isArray(wsList) && 0 < wsList.length) {
                const last = wsList[wsList.length - 1];
                if (!fs.existsSync(last) || !fs.statSync(last).isDirectory()) {
                    lastWorkspaceMissing = true;
                    missingWorkspacePath = last;
                    availableWorkspaces = wsList.slice(0, -1).filter(p =>
                        fs.existsSync(p) && fs.statSync(p).isDirectory());
                }
            }
        }
    } catch (e) {
        writeLog("check missing workspace failed: " + e);
    }
}

// Read the last opened workspace path so crash recovery can select it by default.
let lastWorkspacePath = "";
if (!firstOpen && !getArg("--workspace")) {
    try {
        const wsFile = path.join(confDir, "workspace.json");
        if (fs.existsSync(wsFile)) {
            const wsList = JSON.parse(fs.readFileSync(wsFile, "utf8"));
            if (Array.isArray(wsList) && 0 < wsList.length) {
                lastWorkspacePath = wsList[wsList.length - 1];
            }
        }
    } catch (e) {
        writeLog("read last workspace path failed: " + e);
    }
}

const windowNavigate = (currentWindow, windowType) => {
    currentWindow.webContents.on("will-navigate", (event) => {
        const url = event.url;
        if (url.startsWith(localServer)) {
            try {
                const pathname = new URL(url).pathname;
                if (windowType === "app" && ["/", "/stage/build/app/", "/check-auth"].includes(pathname) ||
                    (windowType === "window" && ["/stage/build/app/window.html", "/check-auth"].includes(pathname)) ||
                    (windowType === "export" && pathname.startsWith("/export/temp/"))) {
                    return;
                }
            } catch (e) {
                return;
            }
        }
        // Open all other links in the browser.
        event.preventDefault();
        shell.openExternal(url);
    });
};

const setProxy = (proxyURL, webContents) => {
    if (proxyURL.startsWith("://")) {
        console.log("network proxy [system]");
        return webContents.session.setProxy({mode: "system"});
    }
    console.log("network proxy [" + proxyURL + "]");
    return webContents.session.setProxy({proxyRules: proxyURL});
};

const hotKey2Electron = (key) => {
    if (!key) {
        return key;
    }
    let electronKey = "";
    if (key.indexOf("⌘") > -1) {
        electronKey += "CommandOrControl+";
    }
    if (key.indexOf("⌃") > -1) {
        electronKey += "Control+";
    }
    if (key.indexOf("⇧") > -1) {
        electronKey += "Shift+";
    }
    if (key.indexOf("⌥") > -1) {
        electronKey += "Alt+";
    }
    return electronKey + key.replace("⌘", "").replace("⇧", "").replace("⌥", "").replace("⌃", "")
        .replace("←", "Left").replace("→", "Right").replace("↑", "Up").replace("↓", "Down").replace(" ", "Space")
        .replace("+", "Plus").replace("⇥", "Tab").replace("⌫", "Backspace").replace("⌦", "Delete").replace("↩", "Return");
};

/**
 * Scribli currently ships English interface text only.
 * @returns {string} App-supported language code.
 */
const resolveAppLanguage = () => {
    return "en";
};

const markExpectedRendererExit = (window) => {
    if (window && !window.isDestroyed()) {
        expectedRendererExitIds.add(window.webContents.id);
    }
};

const exitApp = (port, errorWindowId) => {
    const workspaceIndex = workspaces.findIndex((item) => port.toString() === item.port.toString());
    const workspace = -1 < workspaceIndex ? workspaces[workspaceIndex] : undefined;
    const mainWindow = workspace ? workspace.browserWindow : undefined;
    const tray = workspace ? workspace.tray : undefined;

    // Close all non-main windows using the same port.
    BrowserWindow.getAllWindows().forEach((item) => {
        try {
            const currentURL = new URL(item.getURL());
            if (port.toString() === currentURL.port.toString()) {
                if (!mainWindow || mainWindow.id !== item.id) {
                    item.destroy();
                }
            }
        } catch (e) {
            // load file is not a url
        }
    });
    if (workspace) {
        if (workspaces.length > 1 && mainWindow && !mainWindow.isDestroyed()) {
            markExpectedRendererExit(mainWindow);
            mainWindow.destroy();
        }
        workspaces.splice(workspaceIndex, 1);
    }
    if (tray && ("win32" === process.platform || "linux" === process.platform)) {
        tray.destroy();
    }
    if (workspaces.length === 0 && mainWindow) {
        try {
            if (resetWindowStateOnRestart) {
                fs.writeFileSync(windowStatePath, "{}");
            } else {
                // Save window state for next launch. isMaximized records whether it was maximized at close.
                // x/y/width/height must use getNormalBounds because it returns the restored rectangle for every
                // window state. getBounds returns fullscreen size while maximized, which can pin the restored window
                // to an edge.
                // 
                // https://www.electronjs.org/docs/latest/api/browser-window#wingetnormalbounds
                const bounds = mainWindow.getNormalBounds();
                fs.writeFileSync(windowStatePath, JSON.stringify({
                    isMaximized: mainWindow.isMaximized(),
                    fullscreen: mainWindow.isFullScreen(),
                    isDevToolsOpened: mainWindow.webContents.isDevToolsOpened(),
                    x: bounds.x,
                    y: bounds.y,
                    width: bounds.width,
                    height: bounds.height,
                }));
            }
        } catch (e) {
            writeLog(e);
        }

        if (errorWindowId) {
            markExpectedRendererExit(mainWindow);
            BrowserWindow.getAllWindows().forEach((item) => {
                if (errorWindowId !== item.id) {
                    item.destroy();
                }
            });
        } else {
            markExpectedRendererExit(mainWindow);
            if (keepAppOpenDuringSystemShutdown) {
                mainWindow.destroy();
            } else {
                app.exit();
            }
        }
        globalShortcut.unregisterAll();
        writeLog("exited ui");
    }
};

const localServer = "https://127.0.0.1";

const getServer = (port = kernelPort) => {
    return localServer + ":" + port;
};

const requestKernelExit = (port, options = {}, signal) => {
    if (!port) {
        return Promise.resolve();
    }

    const exitOptions = Object.assign({
        force: false,
        setCurrentWorkspace: true,
        execInstallPkg: 1,
    }, options);
    return net.fetch(getServer(port) + "/api/system/exit", {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
        },
        body: JSON.stringify(exitOptions),
        signal,
    }).catch((error) => {
        writeLog("shutdown kernel failed [port=" + port + "]: " + error);
    });
};

const getSystemShutdownPorts = () => {
    const ports = new Set();
    workspaces.forEach((workspaceItem) => {
        if (workspaceItem.port) {
            ports.add(workspaceItem.port);
        }
    });
    if (bootWindow && !bootWindow.isDestroyed() && kernelPort) {
        ports.add(kernelPort);
    }
    return Array.from(ports);
};

const requestGracefulKernelExit = async (port) => {
    const abortController = new AbortController();
    const timeout = setTimeout(() => abortController.abort(), systemShutdownExitTimeout);
    try {
        const response = await requestKernelExit(port, {
            force: false,
            setCurrentWorkspace: false,
            execInstallPkg: 1,
        }, abortController.signal);
        if (!response) {
            return false;
        }

        const apiData = await response.json();
        if (apiData.code !== 0) {
            writeLog("graceful system shutdown failed [port=" + port + ", code=" + apiData.code + "]");
            return false;
        }
        writeLog("graceful system shutdown succeeded [port=" + port + "]");
        return true;
    } catch (error) {
        writeLog("parse graceful system shutdown response failed [port=" + port + "]: " + error);
        return false;
    } finally {
        clearTimeout(timeout);
    }
};

const resetSystemShutdown = (ports) => {
    if (systemShutdownState === systemShutdownForced) {
        return;
    }

    systemShutdownState = systemShutdownNone;
    gracefulSystemShutdownPromise = undefined;
    keepAppOpenDuringSystemShutdown = false;
    writeLog("system shutdown canceled because Scribli failed to exit gracefully [ports=" + ports.join(",") + "]");
    ports.forEach((port) => {
        const workspace = workspaces.find((item) => port.toString() === item.port.toString());
        if (workspace && workspace.browserWindow && !workspace.browserWindow.isDestroyed()) {
            showWindow(workspace.browserWindow);
        }
    });
    if (bootWindow && !bootWindow.isDestroyed() && ports.includes(kernelPort)) {
        showWindow(bootWindow);
    }
};

const beginGracefulSystemShutdown = () => {
    if (gracefulSystemShutdownPromise || systemShutdownState === systemShutdownForced) {
        return;
    }

    systemShutdownState = systemShutdownEnding;
    const ports = getSystemShutdownPorts();
    if (ports.length === 0) {
        app.exit();
        return;
    }

    keepAppOpenDuringSystemShutdown = true;
    gracefulSystemShutdownPromise = Promise.all(ports.map(async (port) => {
        return {
            port,
            success: await requestGracefulKernelExit(port),
        };
    })).then((results) => {
        const succeededPorts = results.filter((item) => item.success).map((item) => item.port);
        const failedPorts = results.filter((item) => !item.success).map((item) => item.port);
        succeededPorts.forEach((port) => exitApp(port));
        if (bootWindow && !bootWindow.isDestroyed() && succeededPorts.includes(kernelPort)) {
            bootWindow.destroy();
        }

        const remainingPorts = getSystemShutdownPorts();
        const incompletePorts = Array.from(new Set(failedPorts.concat(remainingPorts)));
        if (incompletePorts.length > 0) {
            resetSystemShutdown(incompletePorts);
            return;
        }
        keepAppOpenDuringSystemShutdown = false;
        app.exit();
    }).catch((error) => {
        writeLog("graceful system shutdown failed: " + error);
        resetSystemShutdown(getSystemShutdownPorts());
    });
};

const beginForcedSystemShutdown = () => {
    if (systemShutdownState === systemShutdownForced) {
        return;
    }

    systemShutdownState = systemShutdownForced;
    keepAppOpenDuringSystemShutdown = false;
    getSystemShutdownPorts().forEach((port) => {
        requestKernelExit(port, {
            force: true,
            setCurrentWorkspace: false,
        });
    });
};

if (process.platform === "win32") {
    // On Windows shutdown, restart, or logout, delay session end until the kernel exits cleanly.
    app.on("browser-window-created", (event, window) => {
        window.on("query-session-end", (sessionEvent) => {
            writeLog("query-session-end");
            sessionEvent.preventDefault();
            beginGracefulSystemShutdown();
        });
        window.on("session-end", () => {
            writeLog("session-end");
            beginForcedSystemShutdown();
        });
    });
}

const sleep = (ms) => {
    return new Promise(resolve => setTimeout(resolve, ms));
};

const showErrorWindow = (titleZh, titleEn, content, emoji = "⚠️") => {
    let errorHTMLPath = path.join(appDir, "app", "electron", "error.html");
    if (isDevEnv) {
        errorHTMLPath = path.join(appDir, "electron", "error.html");
    }
    const errWindow = new BrowserWindow({
        width: Math.floor(screen.getPrimaryDisplay().size.width * 0.5),
        height: Math.floor(screen.getPrimaryDisplay().workAreaSize.height * 0.8),
        frame: "darwin" === process.platform,
        titleBarStyle: "hidden",
        fullscreenable: false,
        icon: path.join(appDir, "stage", "icon-large.png"),
        transparent: "darwin" === process.platform, // Avoid a white flash when closing windows in dark mode.
        webPreferences: {
            nodeIntegration: true, webviewTag: true, webSecurity: false, contextIsolation: false,
        },
    });
    errWindow.loadFile(errorHTMLPath, {
        query: {
            home: app.getPath("home"),
            v: appVer,
            title: `<h2>${titleZh}</h2><h2>${titleEn}</h2>`,
            emoji,
            content,
            icon: path.join(appDir, "stage", "icon-large.png"),
        },
    });
    errWindow.show();
    return errWindow.id;
};

const initMainWindow = (currentKernelPort = kernelPort) => {
    if (!app.isReady()) {
        writeLog("initMainWindow: app not ready, skipping");
        return;
    }

    // Restore main window state.
    let oldWindowState = {};
    try {
        oldWindowState = JSON.parse(fs.readFileSync(windowStatePath, "utf8"));
    } catch (e) {
        writeLog("read window state failed: " + e);
        fs.writeFileSync(windowStatePath, "{}");
    }
    let defaultWidth;
    let defaultHeight;
    let workArea;
    try {
        defaultWidth = Math.floor(screen.getPrimaryDisplay().size.width * 0.8);
        defaultHeight = Math.floor(screen.getPrimaryDisplay().workAreaSize.height * 0.8);
        workArea = screen.getPrimaryDisplay().workArea;
    } catch (e) {
        writeLog("get screen size failed: " + e);
    }
    const windowState = Object.assign({}, {
        isMaximized: false,
        fullscreen: false,
        isDevToolsOpened: false,
        x: 0,
        y: 0,
        width: defaultWidth,
        height: defaultHeight,
    }, oldWindowState);

    writeLog("window stat [x=" + windowState.x + ", y=" + windowState.y + ", width=" + windowState.width + ", height=" + windowState.height + "], " +
        "default [x=0, y=0, width=" + defaultWidth + ", height=" + defaultHeight + "], " +
        "old [x=" + oldWindowState.x + ", y=" + oldWindowState.y + ", width=" + oldWindowState.width + ", height=" + oldWindowState.height + "]");

    let resetToCenter = false;
    let x = windowState.x;
    if (-32 < x && 0 > x) {
        x = 0;
    }
    let y = windowState.y;
    if (-32 < y && 0 > y) {
        y = 0;
    }
    if (workArea) {
        // Reset to the smaller value when the window exceeds the work area, otherwise it can hide in the lower left.
        if (windowState.width > workArea.width + 32 || windowState.height > workArea.height + 32) {
            // Window size may reset to default after restart.
            // The +32 tolerance avoids false resets when the window is only a few pixels larger than the work area.
            // 
            // 
            windowState.width = Math.min(defaultWidth, workArea.width);
            windowState.height = Math.min(defaultHeight, workArea.height);
            writeLog("reset window size [width=" + windowState.width + ", height=" + windowState.height + "]");
        }

        if (x >= workArea.width * 0.8 || y >= workArea.height * 0.8) {
            resetToCenter = true;
            writeLog("reset window to center cause x or y >= 80% of workArea");
        }
    }

    if (x < 0 || y < 0) {
        resetToCenter = true;
        writeLog("reset window to center cause x or y < 0");
    }

    if (windowState.width < 493) {
        windowState.width = 493;
        writeLog("reset window width [493]");
    }
    if (windowState.height < 376) {
        windowState.height = 376;
        writeLog("reset window height [376]");
    }

    // Create the main window.
    const currentWindow = new BrowserWindow({
        title: "Scribli",
        show: false,
        width: windowState.width,
        height: windowState.height,
        minWidth: 493,
        minHeight: 376,
        fullscreenable: true,
        fullscreen: windowState.fullscreen,
        trafficLightPosition: {x: 8, y: 8},
        webPreferences: {
            nodeIntegration: true,
            webviewTag: true,
            webSecurity: false,
            contextIsolation: false,
            autoplayPolicy: "user-gesture-required" // Disable media autoplay on desktop.
        },
        frame: "darwin" === process.platform,
        titleBarStyle: "hidden",
        icon: path.join(appDir, "stage", "icon-large.png"),
    });
    remote.enable(currentWindow.webContents);

    if (resetToCenter) {
        currentWindow.center();
    } else {
        writeLog("window position [x=" + x + ", y=" + y + "]");
        currentWindow.setPosition(x, y);
    }
    currentWindow.webContents.userAgent = "Scribli/" + appVer + " Electron " + currentWindow.webContents.userAgent;

    // Load the main UI. Wrap setProxy with a timeout because Electron's session.setProxy can stay pending on some
    // system proxy configurations, which prevents loadURL from running and leaves the main window stuck on boot.
    // Force-load the main UI after at most five seconds whether setProxy completes or not.
    const loadMainURL = () => {
        currentWindow.loadURL(getServer(currentKernelPort) + "/stage/build/app/?v=" + Date.now());
    };
    net.fetch(getServer(currentKernelPort) + "/api/system/getNetwork", {method: "POST"}).then((response) => {
        return response.json();
    }).then((response) => {
        const setProxyDone = setProxy(`${response.data.proxy.scheme}://${response.data.proxy.host}:${response.data.proxy.port}`, currentWindow.webContents);
        Promise.race([
            Promise.resolve(setProxyDone),
            new Promise((resolve) => setTimeout(resolve, 5000)), // Timeout fallback if setProxy stays pending.
        ]).then(loadMainURL).catch(() => {
            writeLog("setProxy failed, load main UI without proxy");
            loadMainURL();
        });
    }).catch((e) => {
        // Continue loading the main UI even when getNetwork fails so the main window does not stay on boot.
        writeLog("getNetwork failed, load main UI without proxy: " + e.message);
        loadMainURL();
    });

    // Bypass security policy for internet service requests.
    // 
    currentWindow.webContents.session.webRequest.onBeforeSendHeaders((details, cb) => {
        if (-1 < details.url.toLowerCase().indexOf("bili")) {
            // Keep Referer for Bilibili.
            // 
            cb({requestHeaders: details.requestHeaders});
            return;
        }

        if (-1 < details.url.toLowerCase().indexOf("douyin")) {
            // Keep Referer for Douyin because iframe login depends on Referer validation.
            // 
            cb({requestHeaders: details.requestHeaders});
            return;
        }

        if (-1 < details.url.toLowerCase().indexOf("youtube")) {
            // Set Referer handling for YouTube.
            // 
            delete details.requestHeaders["Referer"];
            cb({requestHeaders: details.requestHeaders});
            return;
        }

        for (let key in details.requestHeaders) {
            if ("referer" === key.toLowerCase()) {
                delete details.requestHeaders[key];
            }
        }
        cb({requestHeaders: details.requestHeaders});
    });
    currentWindow.webContents.session.webRequest.onHeadersReceived((details, cb) => {
        for (let key in details.responseHeaders) {
            if ("x-frame-options" === key.toLowerCase()) {
                delete details.responseHeaders[key];
            } else if ("content-security-policy" === key.toLowerCase()) {
                delete details.responseHeaders[key];
            } else if ("access-control-allow-origin" === key.toLowerCase()) {
                delete details.responseHeaders[key];
            }
        }
        cb({responseHeaders: details.responseHeaders});
    });

    currentWindow.webContents.on("did-finish-load", () => {
        let appOpenURL = "";
        process.argv.find((arg) => {
            appOpenURL = normalizeIncomingProtocolURL(arg);
            return appOpenURL !== "";
        });
        if (appOpenURL) {
            if (currentWindow.isMinimized()) {
                currentWindow.restore();
            }
            currentWindow.show();
            setTimeout(() => { // Wait for UI JavaScript to finish executing.
                writeLog(appOpenURL);
                currentWindow.webContents.send("scribli-open-url", appOpenURL);
            }, 2000);
        }
    });

    if (windowState.isDevToolsOpened) {
        currentWindow.webContents.openDevTools({mode: "bottom"});
    }

    // Menu.
    const productName = "Scribli";
    const template = [{
        label: productName, submenu: [{
            label: `About ${productName}`, role: "about",
        }, {type: "separator"}, {role: "services"}, {type: "separator"}, {
            label: `Hide ${productName}`, role: "hide",
        }, {role: "hideOthers"}, {role: "unhide"}, {type: "separator"}, {
            label: `Quit ${productName}`, role: "quit",
        },],
    }, {
        role: "editMenu", submenu: [{role: "cut"}, {role: "copy"}, {role: "paste"}, {
            role: "pasteAndMatchStyle", accelerator: "CmdOrCtrl+Shift+C"
        }, {role: "selectAll"},],
    }, {
        role: "windowMenu",
        submenu: [{role: "minimize"}, {role: "zoom"}, {role: "togglefullscreen"}, {type: "separator"}, {role: "toggledevtools"}, {type: "separator"}, {role: "front"},],
    },];
    const menu = Menu.buildFromTemplate(template);
    Menu.setApplicationMenu(menu);
    // Open links from the current page in the browser.
    windowNavigate(currentWindow, "app");
    currentWindow.on("close", (event) => {
        if (currentWindow && !currentWindow.isDestroyed()) {
            currentWindow.webContents.send("scribli-save-close", false);
        }
        event.preventDefault();
    });
    workspaces.push({
        browserWindow: currentWindow,
        webContentsId: currentWindow.webContents.id,
        port: currentKernelPort,
    });
    // Add a timeout fallback after loadURL. If the frontend app bundle fails to load or initialize,
    // scribli-ready-to-show may never arrive; destroy the boot window and show the main window instead of hanging.
    const readyToShowTimeout = setTimeout(() => {
        if (bootWindow && !bootWindow.isDestroyed()) {
            if (!currentWindow.isDestroyed()) {
                writeLog("scribli-ready-to-show timeout, force showing main window");
                currentWindow.show();
            }
            bootWindow.destroy();
        }
    }, 60000);
    ipcMain.once("scribli-ready-to-show", () => {
        clearTimeout(readyToShowTimeout); // Clear the timeout after the normal signal arrives.
        if (isOpenAsHidden()) {
            currentWindow.minimize();
        } else {
            currentWindow.show();
            if (windowState.isMaximized) {
                currentWindow.maximize();
            } else {
                currentWindow.unmaximize();
            }
        }
        if (bootWindow && !bootWindow.isDestroyed()) {
            bootWindow.destroy();
        }
    });
};

const showWindow = (wnd) => {
    if (!wnd || wnd.isDestroyed()) {
        return;
    }

    if (wnd.isMinimized()) {
        wnd.restore();
    }
    wnd.show();
};

const initKernel = (workspace, port, lang, safeMode) => {
    return new Promise(async (resolve) => {
        bootWindow = new BrowserWindow({
            show: false,
            width: Math.floor(screen.getPrimaryDisplay().size.width / 2),
            height: Math.floor(screen.getPrimaryDisplay().workAreaSize.height / 2),
            frame: false,
            backgroundColor: "#1e1e1e",
            resizable: false,
            icon: path.join(appDir, "stage", "icon-large.png"),
            webPreferences: {
                webSecurity: false,
            },
        });
        let bootIndex = path.join(appDir, "app", "electron", "boot.html");
        if (isDevEnv) {
            bootIndex = path.join(appDir, "electron", "boot.html");
        }
        bootWindow.loadFile(bootIndex, {query: {v: appVer, port: kernelPort}});
        if (openAsHidden) {
            bootWindow.minimize();
        } else {
            bootWindow.show();
        }

        const kernelName = "win32" === process.platform ? "Scribli-Kernel.exe" : "Scribli-Kernel";
        const kernelPath = path.join(appDir, "kernel", kernelName);
        if (!fs.existsSync(kernelPath)) {
            showErrorWindow("Kernel program is missing", "Kernel program is missing", `<div>The kernel program was not found. Please reinstall Scribli and add Scribli Kernel to your antivirus trust list if needed.</div><div><i>${kernelPath}</i></div>`);
            bootWindow.destroy();
            resolve(false);
            return;
        }

        if (!isDevEnv || workspaces.length > 0) {
            if (port && "" !== port) {
                kernelPort = port;
            } else {
                const getAvailablePort = () => {
                    // https://gist.github.com/mikeal/1840641
                    return new Promise((portResolve, portReject) => {
                        const server = gNet.createServer();
                        server.on("error", error => {
                            writeLog(error);
                            kernelPort = "";
                            portReject();
                        });
                        server.listen(0, () => {
                            kernelPort = server.address().port;
                            server.close(() => portResolve(kernelPort));
                        });
                    });
                };
                await getAvailablePort();
            }
        }
        writeLog("got kernel port [" + kernelPort + "]");
        if (!kernelPort) {
            bootWindow.destroy();
            resolve(false);
            return;
        }
        const currentKernelPort = kernelPort;
        const cmds = ["serve", "--port", currentKernelPort, "--wd", appDir, "--attach-ui"];
        if (isDevEnv && workspaces.length === 0) {
            cmds.push("--mode", "dev");
        }
        if (workspace && "" !== workspace) {
            cmds.push("--workspace", workspace);
        }
        if (lang && "" !== lang) {
            cmds.push("--lang", lang);
        }
        if (safeMode) {
            cmds.push("--safe-mode", "true");
        }
        let cmd = `ui version [${appVer}], booting kernel [${kernelPath} ${cmds.join(" ")}]`;
        writeLog(cmd);
        if (!isDevEnv || workspaces.length > 0) {
            const kernelProcess = childProcess.spawn(kernelPath, cmds, {
                detached: false, // Do not launch the desktop kernel process in detached mode.
                stdio: "ignore",
            },);

            const kernelPortKey = currentKernelPort.toString();
            kernelProcesses.set(kernelPortKey, kernelProcess);
            writeLog("booted kernel process [pid=" + kernelProcess.pid + ", port=" + currentKernelPort + "]");
            kernelProcess.on("close", (code, signal) => {
                if (kernelProcesses.get(kernelPortKey) === kernelProcess) {
                    kernelProcesses.delete(kernelPortKey);
                }
                const expectedExit = expectedKernelExitPorts.delete(kernelPortKey);
                writeLog(`kernel [pid=${kernelProcess.pid}, port=${currentKernelPort}] exited with code [${code}], signal [${signal}], expected [${expectedExit}]`);
                if (0 !== code && !expectedExit) {
                    let errorWindowId;
                    switch (code) {
                        case 20:
                            errorWindowId = showErrorWindow("The database is unavailable", "The database is unavailable", "<div>Cannot access the database file. Please check workspace/temp/scribli.log for detailed error information.</div>");
                            break;
                        case 21:
                            errorWindowId = showErrorWindow("Failed to listen to port " + currentKernelPort, "Failed to listen to port " + currentKernelPort, "<div>Failed to listen to port " + currentKernelPort + ". Please make sure Scribli has network permissions and is not blocked by firewalls or antivirus software.</div>");
                            break;
                        case 24: // The workspace is locked; try switching to the first open workspace.
                            if (workspaces && 0 < workspaces.length) {
                                showWindow(workspaces[0].browserWindow);
                            }

                            errorWindowId = showErrorWindow("The workspace is locked", "The workspace is locked", "<div>The workspace is being used. End the Scribli-Kernel process in Task Manager or restart the operating system, then start Scribli again.</div>");
                            break;
                        case 25:
                            errorWindowId = showErrorWindow("Failed to create workspace directory", "Failed to create workspace directory", "<div>Insufficient permissions for the workspace folder. Please check workspace/temp/scribli.log for detailed error information.</div>");
                            break;
                        case 26:
                            errorWindowId = showErrorWindow("Potential data corruption avoided", "Potential data corruption avoided", "<div>Files in the workspace are currently opened or locked by third-party software such as sync software or antivirus tools. Continuing could corrupt data, so Scribli Kernel shut down safely.</div><div>Move the workspace to another path, stop sync software from syncing the workspace, and add the workspace to your antivirus trust list if needed.</div>", "🚒");
                            break;
                        case 0:
                            break;
                        default:
                            errorWindowId = showErrorWindow("The kernel exited for unknown reasons", "The kernel exited for unknown reasons", `<div>Scribli Kernel exited for unknown reasons [code=${code}]. Try restarting the operating system, then start Scribli again. If the issue continues, check whether antivirus software is blocking Scribli Kernel.</div>`);
                            break;
                    }

                    exitApp(currentKernelPort, errorWindowId);
                    bootWindow.destroy();
                    resolve(false);
                }
            });
        }

        let apiData;
        let count = 0;
        writeLog("checking kernel version");
        for (; ;) {
            try {
                const apiResult = await net.fetch(getServer(currentKernelPort) + "/api/system/version");
                apiData = await apiResult.json();
                break;
            } catch (e) {
                writeLog("get kernel version failed: " + e.message);
                if (14 < ++count) {
                    writeLog("get kernel ver failed");
                    showErrorWindow("Failed to obtain kernel service port", "Failed to obtain kernel service port", "<div>Failed to obtain kernel service port. Please ensure Scribli has network permissions and is not blocked by firewalls or antivirus software.</div>");
                    bootWindow.destroy();
                    resolve(false);
                    return;
                }
                await sleep(500);
            }
        }

        if (0 === apiData.code) {
            writeLog("got kernel version [" + apiData.data + "]");
            if (!isDevEnv && apiData.data !== appVer) {
                writeLog(`kernel [${apiData.data}] is running, shutdown it now and then start kernel [${appVer}]`);
                requestKernelExit(currentKernelPort);
                bootWindow.destroy();
                resolve(false);
            } else {
                let progressing = false;
                const bootShowStart = Date.now();
                // Boot timeout fallback. Data sync, first full index rebuild, and full-table rebuilds triggered by
                // database version changes all happen before SetBooted(), so this loop gets a large time budget.
                const bootTimeout = 300000;
                while (!progressing) {
                    if (Date.now() - bootShowStart > bootTimeout) {
                        writeLog("boot progress timeout after " + bootTimeout + "ms, exiting boot");
                        showErrorWindow("Boot timeout", "Boot timeout",
                            "<div>Kernel boot timed out. Please check workspace/temp/scribli.log for details, or try restarting Scribli.</div>");
                        requestKernelExit(currentKernelPort);
                        bootWindow.destroy();
                        resolve(false);
                        progressing = true;
                        break;
                    }
                    try {
                        const progressResult = await net.fetch(getServer(currentKernelPort) + "/api/system/bootProgress");
                        const progressData = await progressResult.json();
                        if (progressData.data.progress >= 100) {
                            // After the kernel finishes, wait for the animation fast-forward tail (200ms).
                            await sleep(200);
                            resolve(currentKernelPort);
                            progressing = true;
                        } else {
                            await sleep(100);
                        }
                    } catch (e) {
                        writeLog("get boot progress failed: " + e.message);
                        requestKernelExit(currentKernelPort);
                        bootWindow.destroy();
                        resolve(false);
                        progressing = true;
                    }
                }
            }
        } else {
            writeLog(`get kernel version failed: ${apiData.code}, ${apiData.msg}`);
            resolve(false);
        }
    });
};

app.whenReady().then(() => {
    // Trust self-signed TLS certificates for local HTTPS server
    session.defaultSession.setCertificateVerifyProc((request, callback) => {
        if (request.hostname === "127.0.0.1" || request.hostname === "localhost") {
            callback(0); // VERIFY_OK
        } else {
            callback(-3); // default Chromium handling
        }
    });

    // Renderer crash listener; only unexpected crashes in the workspace main window trigger safe mode.
    app.on("render-process-gone", (event, webContents, details) => {
        writeLog("Render process gone [reason=" + details.reason + ", exitCode=" + details.exitCode + "]");
        if (systemShutdownState !== systemShutdownNone) {
            writeLog("ignore renderer exit during system shutdown [webContentsId=" + webContents.id + "]");
            return;
        }
        if (expectedRendererExitIds.delete(webContents.id)) {
            writeLog("ignore expected renderer exit [webContentsId=" + webContents.id + "]");
            return;
        }

        const workspace = workspaces.find((item) => item.webContentsId === webContents.id);
        if (!workspace) {
            writeLog("ignore non-workspace renderer exit [webContentsId=" + webContents.id + "]");
            return;
        }
        if (!safeModeReasons.has(details.reason)) {
            writeLog("ignore renderer exit reason [reason=" + details.reason + "]");
            return;
        }
        if (handledCrashWebContents.has(webContents.id)) {
            return;
        }

        handledCrashWebContents.add(webContents.id);
        writeAppCrashMarker(workspace, details);
        requestKernelExit(workspace.port, {
            force: true,
            setCurrentWorkspace: false,
        });
        exitApp(workspace.port); // Exit the crashed workspace; the user chooses the startup mode next launch.
    });

    const resetTrayMenu = (tray, lang, mainWindow) => {
        if (!mainWindow || mainWindow.isDestroyed()) {
            return;
        }

        const trayMenuTemplate = [{
            label: mainWindow.isVisible() ? lang.hideWindow : lang.showWindow, click: () => {
                showHideWindow(tray, lang, mainWindow);
            },
        }, {
            label: lang.resetWindow, type: "checkbox", click: v => {
                resetWindowStateOnRestart = v.checked;
                mainWindow.webContents.send("scribli-save-close", true);
            },
        }, {
            label: lang.quit, click: () => {
                mainWindow.webContents.send("scribli-save-close", true);
            },
        },];

        if ("win32" === process.platform) {
            // Support always-on-top on Windows.
            // 
            trayMenuTemplate.splice(1, 0, {
                label: mainWindow.isAlwaysOnTop() ? lang.cancelWindowTop : lang.setWindowTop, click: () => {
                    if (!mainWindow.isAlwaysOnTop()) {
                        mainWindow.setAlwaysOnTop(true);
                    } else {
                        mainWindow.setAlwaysOnTop(false);
                    }
                    resetTrayMenu(tray, lang, mainWindow);
                },
            });
        }
        const contextMenu = Menu.buildFromTemplate(trayMenuTemplate);
        tray.setContextMenu(contextMenu);
    };
    const hideWindow = (wnd) => {
        // Return focus to the previous window after minimizing with `Alt+M`.
        // 
        wnd.minimize();
        // On Mac, hidden windows cannot be shown from the Dock again.
        if ("win32" === process.platform || "linux" === process.platform) {
            wnd.hide();
        }
    };
    const showHideWindow = (tray, lang, mainWindow) => {
        if (!mainWindow || mainWindow.isDestroyed()) {
            return;
        }

        if (!mainWindow.isVisible()) {
            if (mainWindow.isMinimized()) {
                mainWindow.restore();
            }
            mainWindow.show();
        } else {
            hideWindow(mainWindow);
        }

        resetTrayMenu(tray, lang, mainWindow);
    };

    const getWindowByContentId = (id) => {
        return BrowserWindow.getAllWindows().find((win) => win.webContents.id === id);
    };
    ipcMain.on("scribli-context-menu", (event, langs) => {
        const template = [new MenuItem({
            role: "undo", label: langs.undo
        }), new MenuItem({
            role: "redo", label: langs.redo
        }), {type: "separator"}, new MenuItem({
            role: "copy", label: langs.copy
        }), new MenuItem({
            role: "cut", label: langs.cut
        }), new MenuItem({
            role: "delete", label: langs.delete
        }), new MenuItem({
            role: "paste", label: langs.paste
        }), new MenuItem({
            role: "pasteAndMatchStyle", label: langs.pasteAsPlainText
        }), new MenuItem({
            role: "selectAll", label: langs.selectAll
        })];
        const menu = Menu.buildFromTemplate(template);
        menu.popup({window: BrowserWindow.fromWebContents(event.sender)});
    });
    ipcMain.on("scribli-confirm-dialog", (event, options) => {
        event.returnValue = dialog.showMessageBoxSync(BrowserWindow.fromWebContents(event.sender) || BrowserWindow.getFocusedWindow(), options);
    });
    ipcMain.on("scribli-alert-dialog", (event, options) => {
        dialog.showMessageBoxSync(BrowserWindow.fromWebContents(event.sender) || BrowserWindow.getFocusedWindow(), options);
        event.returnValue = undefined;
    });
    ipcMain.on("scribli-first-quit", () => {
        app.exit();
    });
    ipcMain.handle("scribli-get", (event, data) => {
        if (data.cmd === "clipboardRead") {
            return clipboard.read(data.format);
        }
        if (data.cmd === "showOpenDialog") {
            if (data.singleton) {
                const singleton = `${event.sender.id}:${data.singleton}`;
                if (openDialogSingletons.has(singleton)) {
                    return {canceled: true, filePaths: []};
                }
                openDialogSingletons.add(singleton);
                const options = {...data};
                delete options.cmd;
                delete options.singleton;
                return dialog.showOpenDialog(options).finally(() => {
                    openDialogSingletons.delete(singleton);
                });
            }
            return dialog.showOpenDialog(data);
        }
        if (data.cmd === "getContentsId") {
            return event.sender.id;
        }
        if (data.cmd === "isAlwaysOnTop") {
            const wnd = getWindowByContentId(event.sender.id);
            if (!wnd) {
                return false;
            }
            return wnd.isAlwaysOnTop();
        }
        if (data.cmd === "availableSpellCheckerLanguages") {
            return event.sender.session.availableSpellCheckerLanguages;
        }
        if (data.cmd === "setProxy") {
            return setProxy(data.proxyURL, event.sender);
        }
        if (data.cmd === "showSaveDialog") {
            return dialog.showSaveDialog(data);
        }
        if (data.cmd === "isFullScreen") {
            const wnd = getWindowByContentId(event.sender.id);
            if (!wnd) {
                return false;
            }
            return wnd.isFullScreen();
        }
        if (data.cmd === "isMaximized") {
            const wnd = getWindowByContentId(event.sender.id);
            if (!wnd) {
                return false;
            }
            return wnd.isMaximized();
        }
        if (data.cmd === "getMicrophone") {
            return systemPreferences.getMediaAccessStatus("microphone");
        }
        if (data.cmd === "askMicrophone") {
            return systemPreferences.askForMediaAccess("microphone");
        }
        if (data.cmd === "printToPDF") {
            try {
                return getWindowByContentId(data.webContentsId).webContents.printToPDF(data.pdfOptions);
            } catch (e) {
                writeLog("printToPDF: ", e);
                throw e;
            }
        }
        if (data.cmd === "scribli-open-file") {
            let hasMatch = false;
            BrowserWindow.getAllWindows().find(item => {
                const url = new URL(item.webContents.getURL());
                if (item.webContents.id === event.sender.id || data.port !== url.port) {
                    return;
                }
                const ids = decodeURIComponent(url.hash.substring(1)).split("\u200b");
                const options = JSON.parse(data.options);
                if (ids.includes(options.rootID) || ids.includes(options.assetPath)) {
                    item.focus();
                    item.webContents.send("scribli-open-file", options);
                    hasMatch = true;
                    return true;
                }
            });
            return hasMatch;
        }
    });

    const initEventId = [];
    ipcMain.on("scribli-event", (event) => {
        if (initEventId.includes(event.sender.id)) {
            return;
        }
        initEventId.push(event.sender.id);
        const currentWindow = getWindowByContentId(event.sender.id);
        if (!currentWindow) {
            return;
        }
        latestActiveWindow = currentWindow;
        currentWindow.on("focus", () => {
            event.sender.send("scribli-event", "focus");
            latestActiveWindow = currentWindow;
        });
        currentWindow.on("blur", () => {
            event.sender.send("scribli-event", "blur");
        });
        if ("darwin" !== process.platform) {
            currentWindow.on("maximize", () => {
                event.sender.send("scribli-event", "maximize");
            });
            currentWindow.on("unmaximize", () => {
                event.sender.send("scribli-event", "unmaximize");
            });
        }
        currentWindow.on("enter-full-screen", () => {
            event.sender.send("scribli-event", "enter-full-screen");
        });
        currentWindow.on("leave-full-screen", () => {
            event.sender.send("scribli-event", "leave-full-screen");
        });
    });
    ipcMain.on("scribli-cmd", (event, data) => {
        let cmd = data;
        let webContentsId = event.sender.id;
        if (typeof data !== "string") {
            cmd = data.cmd;
            if (data.webContentsId) {
                webContentsId = data.webContentsId;
            }
        }
        const currentWindow = getWindowByContentId(webContentsId);
        switch (cmd) {
            case "showItemInFolder":
                shell.showItemInFolder(data.filePath);
                break;
            case "notification": {
                const n = new Notification({
                    title: data.title,
                    body: data.body,
                    timeoutType: data.timeoutType,
                });
                n.on("click", () => {
                    currentWindow.focus();
                    currentWindow.show();
                });
                n.show();
                break;
            }
            case "setSpellCheckerLanguages":
                BrowserWindow.getAllWindows().forEach(item => {
                    item.webContents.session.setSpellCheckerLanguages(data.languages);
                });
                break;
            case "openPath":
                shell.openPath(data.filePath);
                break;
            case "openDevTools":
                event.sender.openDevTools({mode: "bottom"});
                break;
            case "unregisterGlobalShortcut":
                if (data.accelerator) {
                    globalShortcut.unregister(hotKey2Electron(data.accelerator));
                }
                break;
            case "registerGlobalShortcut":
                if (data.accelerator) {
                    globalShortcut.unregister(hotKey2Electron(data.accelerator));
                    globalShortcut.register(hotKey2Electron(data.accelerator), () => {
                        BrowserWindow.getAllWindows().forEach(itemB => {
                            itemB.webContents.send("scribli-hotkey", {
                                hotkey: data.accelerator
                            });
                        });
                    });
                }
                break;
            case "setTrafficLightPosition":
                if (!currentWindow || !currentWindow.setWindowButtonPosition) {
                    return;
                }
                if (new URL(currentWindow.getURL()).pathname === "/stage/build/app/window.html") {
                    data.position.y += 5 * data.zoom;
                }
                currentWindow.setWindowButtonPosition(data.position);
                break;
            case "show":
                if (!currentWindow) {
                    return;
                }
                showWindow(currentWindow);
                break;
            case "hide":
                if (!currentWindow) {
                    return;
                }
                currentWindow.hide();
                break;
            case "minimize":
                if (!currentWindow) {
                    return;
                }
                currentWindow.minimize();
                break;
            case "maximize":
                if (!currentWindow) {
                    return;
                }
                currentWindow.maximize();
                break;
            case "restore":
                if (!currentWindow) {
                    return;
                }
                if (currentWindow.isFullScreen()) {
                    currentWindow.setFullScreen(false);
                } else {
                    currentWindow.unmaximize();
                }
                break;
            case "focus":
                if (!currentWindow) {
                    return;
                }
                currentWindow.focus();
                break;
            case "setAlwaysOnTopFalse":
                if (!currentWindow) {
                    return;
                }
                currentWindow.setAlwaysOnTop(false);
                break;
            case "setAlwaysOnTopTrue":
                if (!currentWindow) {
                    return;
                }
                currentWindow.setAlwaysOnTop(true);
                break;
            case "clearCache":
                event.sender.session.clearCache();
                break;
            case "redo":
                event.sender.redo();
                break;
            case "undo":
                event.sender.undo();
                break;
            case "destroy":
                if (!currentWindow) {
                    return;
                }
                currentWindow.destroy();
                break;
            case "writeLog":
                writeLog(data.msg);
                break;
            case "closeButtonBehavior":
                if (!currentWindow) {
                    return;
                }
                if (currentWindow.isFullScreen()) {
                    currentWindow.once("leave-full-screen", () => {
                        currentWindow.hide();
                    });
                    currentWindow.setFullScreen(false);
                } else {
                    currentWindow.hide();
                }
                break;
        }
    });
    ipcMain.on("scribli-config-tray", (event, data) => {
        workspaces.find(item => {
            if (item.browserWindow.webContents.id === event.sender.id) {
                hideWindow(item.browserWindow);
                if ("win32" === process.platform || "linux" === process.platform) {
                    resetTrayMenu(item.tray, data.languages, item.browserWindow);
                }
                return true;
            }
        });
    });
    ipcMain.on("scribli-export-pdf", (event, data) => {
        dialog.showOpenDialog({
            title: data.title, properties: ["createDirectory", "openDirectory"],
        }).then((result) => {
            if (result.canceled) {
                event.sender.destroy();
                return;
            }
            data.filePaths = result.filePaths;
            data.webContentsId = event.sender.id;
            getWindowByContentId(data.parentWindowId).send("scribli-export-pdf", data);
        });
    });
    ipcMain.on("scribli-export-newwindow", (event, data) => {
        // The PDF/Word export preview window automatically adjusts according to the size of the main window 
        const wndBounds = getWindowByContentId(event.sender.id).getBounds();
        const wndScreen = screen.getDisplayNearestPoint({x: wndBounds.x, y: wndBounds.y});
        const printWin = new BrowserWindow({
            title: "Scribli",
            show: true,
            width: Math.floor(wndScreen.size.width * 0.8),
            height: Math.floor(wndScreen.size.height * 0.8),
            resizable: true,
            frame: "darwin" === process.platform,
            icon: path.join(appDir, "stage", "icon-large.png"),
            titleBarStyle: "hidden",
            webPreferences: {
                contextIsolation: false,
                nodeIntegration: true,
                webviewTag: true,
                webSecurity: false,
                autoplayPolicy: "user-gesture-required" // Disable media autoplay on desktop.
            },
        });
        printWin.center();
        printWin.webContents.userAgent = "Scribli/" + appVer + " Electron " + printWin.webContents.userAgent;
        printWin.loadURL(data);
        windowNavigate(printWin, "export");
    });
    ipcMain.on("scribli-quit", (event, port) => {
        exitApp(port);
    });
    ipcMain.handle("scribli-install-update", () => {
        writeLog("rejected update install request because runtime updates are disabled");
        return false;
    });
    ipcMain.on("scribli-show-window", (event) => {
        const mainWindow = getWindowByContentId(event.sender.id);
        if (!mainWindow) {
            return;
        }

        if (mainWindow.isMinimized()) {
            mainWindow.restore();
        }
        mainWindow.show();
    });
    ipcMain.on("scribli-open-window", (event, data) => {
        const mainWindow = BrowserWindow.getFocusedWindow() || BrowserWindow.getAllWindows()[0];
        const mainBounds = mainWindow.getBounds();
        const mainScreen = screen.getDisplayNearestPoint({x: mainBounds.x, y: mainBounds.y});
        const win = new BrowserWindow({
            title: "Scribli",
            show: true,
            trafficLightPosition: {x: 8, y: 13},
            width: Math.floor(data.width || mainScreen.size.width * 0.7),
            height: Math.floor(data.height || mainScreen.size.height * 0.9),
            minWidth: 493,
            minHeight: 376,
            fullscreenable: true,
            frame: "darwin" === process.platform,
            icon: path.join(appDir, "stage", "icon-large.png"),
            titleBarStyle: "hidden",
            webPreferences: {
                contextIsolation: false,
                nodeIntegration: true,
                webviewTag: true,
                webSecurity: false,
                autoplayPolicy: "user-gesture-required" // Disable media autoplay on desktop.
            },
        });
        remote.enable(win.webContents);

        if (data.position) {
            win.setPosition(data.position.x, data.position.y);
        } else {
            win.center();
        }
        win.setAlwaysOnTop(data.alwaysOnTop);
        win.webContents.userAgent = "Scribli/" + appVer + " Electron " + win.webContents.userAgent;
        win.webContents.session.setSpellCheckerLanguages(["en-US"]);
        win.loadURL(data.url);
        windowNavigate(win, "window");
        win.on("close", (event) => {
            if (win && !win.isDestroyed()) {
                win.webContents.send("scribli-save-close");
            }
            event.preventDefault();
        });
        const targetScreen = screen.getDisplayNearestPoint(screen.getCursorScreenPoint());
        if (mainScreen.id !== targetScreen.id) {
            win.setBounds(targetScreen.workArea);
        }
    });
    ipcMain.on("scribli-open-workspace", (event, data) => {
        const foundWorkspace = workspaces.find((item) => {
            if (item.workspaceDir === data.workspace) {
                showWindow(item.browserWindow);
                return true;
            }
        });
        if (!foundWorkspace) {
            initKernel(data.workspace, "", "").then((startedKernelPort) => {
                if (startedKernelPort) {
                    initMainWindow(startedKernelPort);
                }
            });
        }
    });
    ipcMain.handle("scribli-init", async (event, data) => {
        const exitWS = workspaces.find(item => {
            if (event.sender.id === item.webContentsId && item.workspaceDir) {
                if (item.tray && ("win32" === process.platform || "linux" === process.platform)) {
                    // Tray menu text does not change with the appearance language 
                    resetTrayMenu(item.tray, data.languages, item.browserWindow);
                }
                return true;
            }
        });
        if (exitWS) {
            return;
        }

        const workspaceItem = workspaces.find((item) => event.sender.id === item.webContentsId);
        if (workspaceItem) {
            workspaceItem.workspaceDir = data.workspaceDir;
            let tray;
            if ("win32" === process.platform || "linux" === process.platform) {
                    // System tray.
                tray = new Tray(path.join(appDir, "stage", "icon-large.png"));
                tray.setToolTip(`${path.basename(data.workspaceDir)} - Scribli v${appVer}`);
                const mainWindow = getWindowByContentId(event.sender.id);
                if (!mainWindow || mainWindow.isDestroyed()) {
                    tray.destroy();
                    tray = undefined;
                } else {
                    resetTrayMenu(tray, data.languages, mainWindow);
                    tray.on("click", () => {
                        showHideWindow(tray, data.languages, mainWindow);
                    });
                }
            }
            workspaceItem.tray = tray;
        }
        await net.fetch(getServer(data.port) + "/api/system/uiproc?pid=" + process.pid, {method: "POST"});
    });
    ipcMain.on("scribli-hotkey", (event, data) => {
        if (!data.hotkeys || data.hotkeys.length === 0) {
            return;
        }
        workspaces.find(workspaceItem => {
            if (event.sender.id === workspaceItem.browserWindow.webContents.id) {
                workspaceItem.hotkeys = data.hotkeys;
                return true;
            }
        });
        data.hotkeys.forEach((item, index) => {
            const shortcut = hotKey2Electron(item);
            if (!shortcut) {
                return;
            }
            if (globalShortcut.isRegistered(shortcut)) {
                globalShortcut.unregister(shortcut);
            }
            if (index === 0) {
                globalShortcut.register(shortcut, () => {
                    let currentWorkspace;
                    const currentWebContentsId = (latestActiveWindow && !latestActiveWindow.isDestroyed()) ? latestActiveWindow.webContents.id : undefined;
                    workspaces.find(workspaceItem => {
                        if (currentWebContentsId === workspaceItem.browserWindow.webContents.id && workspaceItem.hotkeys[0] === item) {
                            currentWorkspace = workspaceItem;
                            return true;
                        }
                    });
                    if (!currentWorkspace) {
                        workspaces.find(workspaceItem => {
                            if (workspaceItem.hotkeys[0] === item && event.sender.id === workspaceItem.browserWindow.webContents.id) {
                                currentWorkspace = workspaceItem;
                                return true;
                            }
                        });
                    }
                    if (!currentWorkspace) {
                        return;
                    }
                    const mainWindow = currentWorkspace.browserWindow;
                    if (mainWindow.isMinimized()) {
                        mainWindow.restore();
                        mainWindow.show(); // Restoring after `Alt+M` avoids an editor focus freeze.
                    } else {
                        if (mainWindow.isVisible()) {
                            if (!mainWindow.isFocused()) {
                                mainWindow.show();
                            } else {
                                hideWindow(mainWindow);
                            }
                        } else {
                            mainWindow.show();
                        }
                    }
                    if ("win32" === process.platform || "linux" === process.platform) {
                        resetTrayMenu(currentWorkspace.tray, data.languages, mainWindow);
                    }
                });
            } else {
                globalShortcut.register(shortcut, () => {
                    BrowserWindow.getAllWindows().forEach(itemB => {
                        itemB.webContents.send("scribli-hotkey", {
                            hotkey: item
                        });
                    });
                });
            }
        });
    });
    ipcMain.on("scribli-send-windows", (event, data) => {
        BrowserWindow.getAllWindows().forEach(item => {
            item.webContents.send("scribli-send-windows", data);
        });
    });
    ipcMain.on("scribli-auto-launch", (event, data) => {
        app.setLoginItemSettings({
            openAtLogin: data.openAtLogin,
            args: data.openAsHidden ? ["--openAsHidden"] : ""
        });
    });
    const appCrashInfo = readAppCrashInfo();
    if (firstOpen) {
        const firstOpenWindow = new BrowserWindow({
            width: Math.floor(screen.getPrimaryDisplay().size.width * 0.6),
            height: Math.floor(screen.getPrimaryDisplay().workAreaSize.height * 0.8),
            frame: "darwin" === process.platform,
            titleBarStyle: "hidden",
            fullscreenable: false,
            icon: path.join(appDir, "stage", "icon-large.png"),
            transparent: "darwin" === process.platform,
            webPreferences: {
                nodeIntegration: true, webviewTag: true, webSecurity: false, contextIsolation: false,
            },
        });
        let initHTMLPath = path.join(appDir, "app", "electron", "init.html");
        if (isDevEnv) {
            initHTMLPath = path.join(appDir, "electron", "init.html");
        }

        // Improve the appearance language used during desktop initialization.
        // 
        const language = resolveAppLanguage();
        firstOpenWindow.loadFile(initHTMLPath, {
            query: {
                lang: language,
                home: app.getPath("home"),
                v: appVer,
                icon: path.join(appDir, "stage", "icon-large.png"),
            },
        });
        firstOpenWindow.show();
        // Initial startup.
        ipcMain.on("scribli-first-init", (event, data) => {
            initKernel(data.workspace, "", data.lang).then((startedKernelPort) => {
                if (startedKernelPort) {
                    initMainWindow(startedKernelPort);
                }
            });
            firstOpenWindow.destroy();
        });
    } else if (appCrashInfo) {
        // The previous workspace renderer crashed; show the safe-mode selection window.
        const safeModeWindow = new BrowserWindow({
            width: Math.floor(screen.getPrimaryDisplay().size.width * 0.55),
            height: Math.floor(screen.getPrimaryDisplay().workAreaSize.height * 0.65),
            frame: "darwin" === process.platform,
            titleBarStyle: "hidden",
            fullscreenable: false,
            icon: path.join(appDir, "stage", "icon-large.png"),
            transparent: "darwin" === process.platform,
            webPreferences: {
                nodeIntegration: true, webviewTag: true, webSecurity: false, contextIsolation: false,
            },
        });
        let safeModeHTMLPath = path.join(appDir, "app", "electron", "workspace.html");
        if (isDevEnv) {
            safeModeHTMLPath = path.join(appDir, "electron", "workspace.html");
        }

        // Improve the appearance language used during desktop initialization.
        // 
        const language = resolveAppLanguage();
        let crashWorkspace = appCrashInfo.workspaceDir || lastWorkspacePath;
        if (!appCrashInfo.workspaceDir && !isDirectory(crashWorkspace)) {
            crashWorkspace = availableWorkspaces[availableWorkspaces.length - 1] || lastWorkspacePath;
        }
        const crashWorkspaceMissing = !isDirectory(crashWorkspace);
        safeModeWindow.loadFile(safeModeHTMLPath, {
            query: {
                lang: language,
                home: app.getPath("home"),
                v: appVer,
                icon: path.join(appDir, "stage", "icon-large.png"),
                crash: "1",
                workspace: crashWorkspace,
                crashWorkspaceMissing: crashWorkspaceMissing ? "1" : "0",
                missing: crashWorkspaceMissing ? crashWorkspace : "",
                crashInfo: appCrashInfo.crashInfo,
            },
        });
        safeModeWindow.show();
        // Start the kernel after the user chooses a startup mode, and clear crash info only after successful boot.
        ipcMain.on("scribli-select-workspace", (event, data) => {
            initKernel(data.workspace, "", data.lang, data.safeMode).then((startedKernelPort) => {
                if (startedKernelPort) {
                    clearAppCrashInfo();
                    initMainWindow(startedKernelPort);
                }
            });
            safeModeWindow.destroy();
        });
    } else if (lastWorkspaceMissing) {
        // The last used workspace is missing; show the workspace selection window.
        // 
        const missingWorkspaceWindow = new BrowserWindow({
            width: Math.floor(screen.getPrimaryDisplay().size.width * 0.55),
            height: Math.floor(screen.getPrimaryDisplay().workAreaSize.height * 0.65),
            frame: "darwin" === process.platform,
            titleBarStyle: "hidden",
            fullscreenable: false,
            icon: path.join(appDir, "stage", "icon-large.png"),
            transparent: "darwin" === process.platform,
            webPreferences: {
                nodeIntegration: true, webviewTag: true, webSecurity: false, contextIsolation: false,
            },
        });
        let missingWorkspaceHTMLPath = path.join(appDir, "app", "electron", "workspace.html");
        if (isDevEnv) {
            missingWorkspaceHTMLPath = path.join(appDir, "electron", "workspace.html");
        }

        // Improve the appearance language used during desktop initialization.
        // 
        const language = resolveAppLanguage();
        missingWorkspaceWindow.loadFile(missingWorkspaceHTMLPath, {
            query: {
                lang: language,
                home: app.getPath("home"),
                v: appVer,
                icon: path.join(appDir, "stage", "icon-large.png"),
                missing: missingWorkspacePath,
                workspaces: availableWorkspaces.join("\n"),
            },
        });
        missingWorkspaceWindow.show();
        // Start the kernel after workspace selection.
        ipcMain.on("scribli-select-workspace", (event, data) => {
            initKernel(data.workspace, "", data.lang).then((startedKernelPort) => {
                if (startedKernelPort) {
                    initMainWindow(startedKernelPort);
                }
            });
            missingWorkspaceWindow.destroy();
        });
    } else {
        const workspace = getArg("--workspace");
        if (workspace) {
            writeLog("got arg [--workspace=" + workspace + "]");
        }
        const port = getArg("--port");
        if (port) {
            writeLog("got arg [--port=" + port + "]");
        }
        const safeMode = getArg("--safe-mode") === "true";
        if (safeMode) {
            writeLog("got arg [--safe-mode=true]");
        }
        const lang = getArg("--lang") || "";
        if (lang) {
            writeLog("got arg [--lang=" + lang + "]");
        }
        initKernel(workspace, port, lang, safeMode).then((startedKernelPort) => {
            if (startedKernelPort) {
                initMainWindow(startedKernelPort);
            }
        });
    }

    // Power-related events must be inside whenReady, otherwise Linux startup can fail with
    // Trace/breakpoint trap (core dumped).
    // 
    powerMonitor.on("suspend", () => {
        writeLog("system suspend");
    });
    powerMonitor.on("resume", async () => {
        // After desktop resume, check network connectivity before running data sync.
        // 
        writeLog("system resume");

        const isOnline = async () => {
            return net.isOnline();
        };
        let online = false;
        for (let i = 0; i < 7; i++) {
            if (await isOnline()) {
                online = true;
                break;
            }

            writeLog("network is offline");
            await sleep(1000);
        }

        if (!online) {
            writeLog("network is offline, do not sync after system resume");
            return;
        }

        workspaces.forEach(item => {
            const currentURL = new URL(item.browserWindow.getURL());
            const server = getServer(currentURL.port);
            writeLog("sync after system resume [" + server + "/api/sync/performSync" + "]");
            net.fetch(server + "/api/sync/performSync", {method: "POST"});
        });
    });
    powerMonitor.on("shutdown", () => {
        writeLog("system shutdown");
        beginForcedSystemShutdown();
    });
    powerMonitor.on("lock-screen", () => {
        writeLog("system lock-screen");
        BrowserWindow.getAllWindows().forEach(item => {
            item.webContents.send("scribli-send-windows", {cmd: "lockscreenByMode"});
        });
    });
});

app.on("open-url", async (event, url) => { // for macOS
    const appOpenURL = normalizeIncomingProtocolURL(url);
    if (appOpenURL) {
        event.preventDefault();
        let isBackground = true;
        if (workspaces.length === 0) {
            isBackground = false;
            let index = 0;
            while (index < 10) {
                index++;
                await sleep(500);
                if (workspaces.length > 0) {
                    break;
                }
            }
        }
        if (!isBackground) {
            await sleep(1500);
        }
        workspaces.forEach(item => {
            if (item.browserWindow && !item.browserWindow.isDestroyed()) {
                item.browserWindow.webContents.send("scribli-open-url", appOpenURL);
            }
        });
    }
});

app.on("second-instance", (event, argv) => {
    writeLog("second-instance [" + argv + "]");
    let workspace = argv.find((arg) => arg.startsWith("--workspace="));
    if (workspace) {
        workspace = workspace.split("=")[1];
        writeLog("got second-instance arg [--workspace=" + workspace + "]");
    }
    let port = argv.find((arg) => arg.startsWith("--port="));
    if (port) {
        port = port.split("=")[1];
        writeLog("got second-instance arg [--port=" + port + "]");
    } else {
        port = 0;
    }
    let lang = argv.find((arg) => arg.startsWith("--lang="));
    if (lang) {
        lang = lang.split("=")[1];
        writeLog("got second-instance arg [--lang=" + lang + "]");
    } else {
        lang = "";
    }
    const foundWorkspace = workspaces.find(item => {
        if (item.browserWindow && !item.browserWindow.isDestroyed()) {
            if (workspace && workspace === item.workspaceDir) {
                showWindow(item.browserWindow);
                return true;
            }
        }
    });
    if (foundWorkspace) {
        return;
    }
    if (workspace) {
        initKernel(workspace, port, lang).then((startedKernelPort) => {
            if (startedKernelPort) {
                initMainWindow(startedKernelPort);
            }
        });
        return;
    }

    let appOpenURL = "";
    argv.find((arg) => {
        appOpenURL = normalizeIncomingProtocolURL(arg);
        return appOpenURL !== "";
    });
    workspaces.forEach(item => {
        if (item.browserWindow && !item.browserWindow.isDestroyed() && appOpenURL) {
            item.browserWindow.webContents.send("scribli-open-url", appOpenURL);
        }
    });

    if (!appOpenURL && 0 < workspaces.length) {
        showWindow(workspaces[0].browserWindow);
    }
});

app.on("activate", () => {
    if (workspaces.length > 0) {
        const mainWindow = (latestActiveWindow && !latestActiveWindow.isDestroyed()) ? latestActiveWindow : workspaces[0].browserWindow;
        if (mainWindow && !mainWindow.isDestroyed()) {
            mainWindow.show();
        }
    }
    if (BrowserWindow.getAllWindows().length === 0) {
        initMainWindow();
    }
});

app.on("web-contents-created", (webContentsCreatedEvent, contents) => {
    contents.setWindowOpenHandler((details) => {
        // 
        if (details.url.startsWith("file:///") && details.disposition === "foreground-tab") {
            return;
        }
        // Handle links opened inside the editor, such as links from iframe blocks.
        shell.openExternal(details.url);
        return {action: "deny"};
    });
});

app.on("before-quit", (event) => {
    workspaces.forEach(item => {
        if (item.browserWindow && !item.browserWindow.isDestroyed()) {
            event.preventDefault();
            item.browserWindow.webContents.send("scribli-save-close", true);
        }
    });
});

function writeLog(out) {
    console.log(out);
    const logFile = path.join(confDir, "app.log");
    let log = "";
    const maxLogLines = 1024;
    try {
        if (fs.existsSync(logFile)) {
            log = fs.readFileSync(logFile).toString();
            let lines = log.split("\n");
            if (maxLogLines < lines.length) {
                log = lines.slice(maxLogLines / 2, maxLogLines).join("\n") + "\n";
            }
        }
        out = out.toString();
        out = new Date().toISOString().replace(/T/, " ").replace(/\..+/, "") + " " + out;
        log += out + "\n";
        fs.writeFileSync(logFile, log);
    } catch (e) {
        console.error(e);
    }
}

// Synchronously record workspace main renderer crash markers so they reach disk before the main process exits.
const writeAppCrashMarker = (workspace, details) => {
    const timestamp = new Date().toISOString();
    const marker = {
        version: 1,
        timestamp: timestamp,
        reason: details.reason,
        exitCode: details.exitCode,
        workspaceDir: workspace.workspaceDir || "",
    };

    try {
        fs.writeFileSync(appCrashMarkerPath, JSON.stringify(marker, null, 2));
    } catch (e) {
        console.error(e);
    }

    try {
        const line = timestamp.replace(/T/, " ").replace(/\..+/, "") +
            " Render process gone [reason=" + details.reason + ", exitCode=" + details.exitCode +
            ", workspace=" + JSON.stringify(marker.workspaceDir) + "]";
        let log = "";
        if (fs.existsSync(appCrashLogPath)) {
            log = fs.readFileSync(appCrashLogPath, "utf8");
        }
        const lines = (log + line).trimEnd().split("\n").slice(-20);
        fs.writeFileSync(appCrashLogPath, lines.join("\n") + "\n");
    } catch (e) {
        console.error(e);
    }
};

const isDirectory = (filePath) => {
    if (!filePath) {
        return false;
    }

    try {
        return fs.statSync(filePath).isDirectory();
    } catch (e) {
        return false;
    }
};

// Prefer structured markers while preserving compatibility with older app.crash.log files.
const readAppCrashInfo = () => {
    if (fs.existsSync(appCrashMarkerPath)) {
        try {
            const markerText = fs.readFileSync(appCrashMarkerPath, "utf8");
            const marker = JSON.parse(markerText);
            if (noSafeModeReasons.has(marker.reason)) {
                fs.unlinkSync(appCrashMarkerPath);
            } else {
                let crashInfo = markerText;
                if (fs.existsSync(appCrashLogPath)) {
                    crashInfo = fs.readFileSync(appCrashLogPath, "utf8");
                }
                return {
                    workspaceDir: typeof marker.workspaceDir === "string" ? marker.workspaceDir : "",
                    crashInfo: crashInfo,
                };
            }
        } catch (e) {
            writeLog("read crash marker failed: " + e);
            try {
                return {
                    workspaceDir: "",
                    crashInfo: fs.readFileSync(appCrashMarkerPath, "utf8"),
                };
            } catch (readError) {
                writeLog("read invalid crash marker failed: " + readError);
                return {
                    workspaceDir: "",
                    crashInfo: "Invalid renderer crash marker",
                };
            }
        }
    }

    if (!fs.existsSync(appCrashLogPath)) {
        return undefined;
    }

    try {
        const crashInfo = fs.readFileSync(appCrashLogPath, "utf8");
        const legacyLines = crashInfo.split(/\r?\n/).filter((line) => line.trim());
        const reasons = legacyLines.map((line) => {
            const match = line.match(/reason=([^,\]]+)/);
            return match ? match[1] : undefined;
        });
        if (reasons.length > 0 && reasons.every((reason) => reason && noSafeModeReasons.has(reason))) {
            fs.unlinkSync(appCrashLogPath);
            writeLog("ignored legacy crash log without safe mode reason");
            return undefined;
        }
        return {
            workspaceDir: "",
            crashInfo: crashInfo,
        };
    } catch (e) {
        writeLog("read crash log failed: " + e);
        return {
            workspaceDir: "",
            crashInfo: "Unreadable renderer crash log",
        };
    }
};

// After safe-mode selection and successful kernel boot, delete the crash info used for this recovery.
const clearAppCrashInfo = () => {
    [appCrashMarkerPath, appCrashLogPath].forEach((filePath) => {
        try {
            fs.unlinkSync(filePath);
        } catch (e) {
            // Ignore missing files and similar cleanup errors.
        }
    });
};
