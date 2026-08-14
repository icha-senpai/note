const path = require("path");

const repoRoot = path.resolve(__dirname, "../../..");

module.exports = {
    testDir: __dirname,
    testMatch: "app-audit.spec.js",
    workers: 1,
    globalTimeout: 90000,
    timeout: 90000,
    expect: {
        timeout: 10000,
    },
    reporter: [["list"], ["html", {outputFolder: path.join(repoRoot, "tmp", "playwright-electron-report"), open: "never"}]],
    use: {
        screenshot: "only-on-failure",
        trace: "retain-on-failure",
        video: "retain-on-failure",
    },
    outputDir: path.join(repoRoot, "tmp", "playwright-electron-results"),
};
