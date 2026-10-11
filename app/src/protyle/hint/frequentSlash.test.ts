import * as assert from "node:assert/strict";
import {test} from "node:test";
import {frequentSlashItems, normalizeSlashUsage} from "./frequentSlash";

test("frequent commands preserve ties, filter unavailable commands and cap the section", () => {
    const items = Array.from({length: 8}, (_, i) => ({id: `command${i}`, filter: ["command"], html: "command"}));
    assert.deepEqual(frequentSlashItems(items, {command7: 9, command4: 3, command0: 3, command1: 1, command2: 1, command3: 1, removed: 10}).map(item => item.id),
        ["command7", "command0", "command4", "command1", "command2"]);
    assert.deepEqual(frequentSlashItems([{id: "command0", html: "reference"}], {command0: 5}), []);
});

test("usage rejects corrupt values and never ranks prototype keys", () => {
    assert.deepEqual(normalizeSlashUsage({good: 3, negative: -1, fraction: 2.5, huge: Infinity, content: "note text"}), {good: 3});
    assert.deepEqual(frequentSlashItems([{id: "toString", filter: ["x"], html: "x"}], {}), []);
});
