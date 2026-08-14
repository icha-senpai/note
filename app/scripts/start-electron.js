// Scribli - Refactor your thinking
// Copyright (c) 2020-present Scribli
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

const childProcess = require("child_process");
const electron = require("electron");

const env = {...process.env, NODE_ENV: "development"};
delete env.ELECTRON_RUN_AS_NODE;

const child = childProcess.spawn(electron, ["./electron/main.js"], {
    cwd: process.cwd(),
    env,
    stdio: "inherit",
});

child.on("exit", (code, signal) => {
    if (signal) {
        process.kill(process.pid, signal);
        return;
    }
    process.exit(code ?? 0);
});

child.on("error", (error) => {
    console.error(error);
    process.exit(1);
});
