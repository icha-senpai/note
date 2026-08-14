const fs = require("fs");
const os = require("os");
const path = require("path");

const requirePlaywrightTest = () => {
    const runnerArg = process.argv.find((entry) => /node_modules[\\/]+(?:@playwright[\\/]test|playwright)[\\/]+/.test(entry));
    if (runnerArg) {
        const nodeModules = runnerArg.slice(0, runnerArg.search(/node_modules[\\/]+(?:@playwright[\\/]test|playwright)[\\/]+/) + "node_modules".length);
        const candidate = path.join(nodeModules, "@playwright", "test");
        if (fs.existsSync(path.join(candidate, "index.js"))) {
            return require(candidate);
        }
        const playwrightTest = path.join(nodeModules, "playwright", "test.js");
        if (fs.existsSync(playwrightTest)) {
            return require(playwrightTest);
        }
    }
    try {
        return require("@playwright/test");
    } catch (error) {
        const cacheRoot = process.env.npm_config_cache || path.join(os.homedir(), "AppData", "Local", "npm-cache");
        const npxRoot = path.join(cacheRoot, "_npx");
        const packageRoots = fs.existsSync(npxRoot)
            ? fs.readdirSync(npxRoot).map((entry) => path.join(npxRoot, entry, "node_modules", "@playwright", "test"))
            : [];
        const packageRoot = packageRoots.find((candidate) => fs.existsSync(path.join(candidate, "index.js")));
        if (!packageRoot) {
            throw error;
        }
        return require(packageRoot);
    }
};

module.exports = requirePlaywrightTest();
