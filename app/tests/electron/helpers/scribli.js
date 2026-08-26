const fs = require("fs");
const childProcess = require("child_process");
const net = require("net");
const path = require("path");
const {chromium, expect} = require("./playwright");

const appRoot = path.resolve(__dirname, "../../..");
const repoRoot = path.resolve(appRoot, "..");
const artifactRoot = path.join(repoRoot, "tmp", "playwright-electron");
const actionTimeout = 15000;
const cdpTimeout = 15000;
const mainWindowTimeout = 55000;
const readyTimeout = 20000;

const sanitizeName = (value) => value.replace(/[^a-z0-9_-]+/gi, "-").replace(/^-|-$/g, "").slice(0, 80) || "scribli";

const getFreePort = () => new Promise((resolve, reject) => {
    const server = net.createServer();
    server.unref();
    server.on("error", reject);
    server.listen(0, "127.0.0.1", () => {
        const address = server.address();
        server.close(() => resolve(address.port));
    });
});

const readTail = (filePath) => {
    try {
        const content = fs.readFileSync(filePath, "utf8");
        return content.slice(-4000);
    } catch (error) {
        return "";
    }
};

const waitForExit = (child, timeout = 3000) => new Promise((resolve) => {
    if (!child || child.exitCode !== null || child.killed) {
        resolve();
        return;
    }
    const timer = setTimeout(resolve, timeout);
    child.once("exit", () => {
        clearTimeout(timer);
        resolve();
    });
});

const removeRunRoot = async (runRoot) => {
    await fs.promises.rm(runRoot, {recursive: true, force: true, maxRetries: 3, retryDelay: 100});
};

const killProcessTree = async (child) => {
    if (!child || child.exitCode !== null || child.killed) {
        return;
    }
    if (process.platform === "win32") {
        await new Promise((resolve) => {
            const killer = childProcess.spawn("taskkill.exe", ["/PID", String(child.pid), "/T", "/F"], {
                stdio: "ignore",
                windowsHide: true,
            });
            killer.once("exit", resolve);
            killer.once("error", resolve);
        });
        return;
    }
    child.kill();
    await waitForExit(child);
};

const waitForCDP = async (port, electronProcess, errLogPath) => {
    const deadline = Date.now() + cdpTimeout;
    while (Date.now() < deadline) {
        if (electronProcess.exitCode !== null) {
            throw new Error(`Electron exited before CDP opened. ${readTail(errLogPath)}`);
        }
        try {
            const response = await fetch(`http://127.0.0.1:${port}/json/version`);
            if (response.ok) {
                return;
            }
        } catch (error) {
            // Keep polling until Electron opens the debugging endpoint.
        }
        await new Promise((resolve) => setTimeout(resolve, 250));
    }
    throw new Error(`Timed out waiting for Electron CDP port ${port}`);
};

const waitForMainPage = async (browser) => {
    const deadline = Date.now() + mainWindowTimeout;
    while (Date.now() < deadline) {
        for (const context of browser.contexts()) {
            for (const page of context.pages()) {
                if (page.url().includes("/stage/build/app/")) {
                    return page;
                }
            }
        }
        await new Promise((resolve) => setTimeout(resolve, 500));
    }
    throw new Error("Timed out waiting for Scribli main window");
};

const launchScribli = async (testInfo, options = {}) => {
    const titlePath = Array.isArray(testInfo.titlePath) ? testInfo.titlePath : [testInfo.title || "scribli"];
    const testSlug = sanitizeName(titlePath.join("-"));
    const runRoot = path.join(artifactRoot, `${Date.now()}-${process.pid}-${testSlug}`);
    const workspace = path.join(runRoot, "workspace");
    const appData = path.join(runRoot, "appdata");
    const configDir = path.join(runRoot, "config");
    fs.mkdirSync(workspace, {recursive: true});
    fs.mkdirSync(appData, {recursive: true});
    fs.mkdirSync(configDir, {recursive: true});
    if (options.workspaceArg === false) {
        fs.writeFileSync(path.join(configDir, "workspace.json"), JSON.stringify([workspace]));
    }

    const port = await getFreePort();
    const cdpPort = await getFreePort();
    const electronPath = require("electron");
    const outLogPath = path.join(runRoot, "electron.out.log");
    const errLogPath = path.join(runRoot, "electron.err.log");
    const outLog = fs.openSync(outLogPath, "a");
    const errLog = fs.openSync(errLogPath, "a");
    const env = {
        ...process.env,
        APPDATA: appData,
        NODE_ENV: "development",
        SCRIBLI_CONFIG_DIR: configDir,
        SCRIBLI_WORKSPACE_PATH: workspace,
    };
    delete env.ELECTRON_RUN_AS_NODE;

    const args = [
        "./electron/main.js",
        `--remote-debugging-port=${cdpPort}`,
        `--port=${port}`,
    ];
    if (options.workspaceArg !== false) {
        args.push(`--workspace=${workspace}`);
    }

    const electronProcess = childProcess.spawn(electronPath, args, {
        cwd: appRoot,
        env,
        stdio: ["ignore", outLog, errLog],
        windowsHide: false,
    });

    const pageErrors = [];
    const consoleErrors = [];
    let browser;
    let page;
    const cleanup = async (cleanupOptions = {}) => {
        const preserveArtifacts = Boolean(cleanupOptions.preserveArtifacts);
        if (preserveArtifacts && (pageErrors.length || consoleErrors.length)) {
            fs.writeFileSync(path.join(runRoot, "browser-errors.json"), JSON.stringify({pageErrors, consoleErrors}, undefined, 2));
        }
        if (browser) {
            await browser.close().catch(() => undefined);
        }
        await killProcessTree(electronProcess);
        for (const fd of [outLog, errLog]) {
            try {
                fs.closeSync(fd);
            } catch (error) {
                // Ignore duplicate close errors during failed launches.
            }
        }
        if (!preserveArtifacts) {
            await removeRunRoot(runRoot);
        }
    };

    try {
        await waitForCDP(cdpPort, electronProcess, errLogPath);
        browser = await chromium.connectOverCDP(`http://127.0.0.1:${cdpPort}`);
        page = await waitForMainPage(browser);
        page.setDefaultTimeout(actionTimeout);
        page.on("pageerror", (error) => pageErrors.push(error.stack || error.message));
        page.on("console", (message) => {
            if (message.type() === "error") {
                consoleErrors.push(message.text());
            }
        });
        await page.waitForFunction(() => Boolean(window.scribli && window.openFileByURL), null, {timeout: readyTimeout});
    } catch (error) {
        await cleanup({preserveArtifacts: true});
        error.message = `${error.message}\nElectron logs: ${outLogPath}\n${errLogPath}`;
        throw error;
    }

    const api = async (url, body = {}) => {
        const response = await page.evaluate(async ({url, body}) => {
            const result = await fetch(url, {
                method: "POST",
                headers: {"Content-Type": "application/json"},
                body: JSON.stringify(body),
            });
            return result.json();
        }, {url, body});
        expect(response.code, `${url} failed: ${response.msg || JSON.stringify(response)}`).toBe(0);
        return response.data;
    };

    return {api, appData, artifactDir: runRoot, close: cleanup, page, port, process: electronProcess, workspace};
};

const createDoc = async (scribli, title, markdown) => {
    const notebooksData = await scribli.api("/api/notebook/lsNotebooks");
    const notebook = notebooksData.notebooks?.[0]?.id;
    expect(notebook, "expected at least one notebook in isolated workspace").toBeTruthy();
    return scribli.api("/api/filetree/createDocWithMd", {
        notebook,
        path: `/Playwright Electron QA/${title}`,
        markdown,
    });
};

const newNodeID = async (scribli) => scribli.page.evaluate(() => window.Lute.NewNodeID());

const insertBlock = async (scribli, data, options = {}) => {
    return scribli.api("/api/block/insertBlock", {
        data,
        dataType: options.dataType || "dom",
        nextID: options.nextID || "",
        parentID: options.parentID || "",
        previousID: options.previousID || "",
    });
};

const updateBlock = async (scribli, id, markdown) => {
    return scribli.api("/api/block/updateBlock", {
        data: markdown,
        dataType: "markdown",
        id,
    });
};

const openDoc = async (scribli, id, marker) => {
    await scribli.page.evaluate((id) => window.openFileByURL(`scribli://blocks/${id}?focus=1`), id);
    await scribli.page.waitForFunction((marker) => {
        const active = document.querySelector(".layout__wnd--active .protyle-wysiwyg");
        return Boolean(active && ((active.textContent || "").includes(marker) || document.title.includes(marker)));
    }, marker, {timeout: actionTimeout});
};

const getCanvas = async (scribli, id) => {
    const data = await scribli.api("/api/canvas/call", {action: "get", id});
    return data.structuredContent.canvas;
};

const getBlockKramdown = async (scribli, id) => {
    return scribli.api("/api/block/getBlockKramdown", {id, mode: "md"});
};

const createAttributeView = async (scribli, id, blockID) => {
    const data = await scribli.api("/api/av/renderAttributeView", {id, blockID, createIfNotExist: true, page: 1, pageSize: 10});
    expect(data.id || data.av?.id, `expected database ${id} to render`).toBeTruthy();
    return data;
};

const checkBlockExists = async (scribli, id) => {
    return scribli.api("/api/block/checkBlockExist", {id});
};

const removeDocByID = async (scribli, id) => {
    return scribli.api("/api/filetree/removeDocByID", {id});
};

module.exports = {appRoot, artifactRoot, checkBlockExists, createAttributeView, createDoc, getBlockKramdown, getCanvas, insertBlock, launchScribli, newNodeID, openDoc, removeDocByID, repoRoot, updateBlock};
