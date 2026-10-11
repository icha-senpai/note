import {strict as assert} from "node:assert";
import {test} from "node:test";
import {createSettingSaveQueue} from "./saveQueue";

const tick = () => new Promise(resolve => setTimeout(resolve, 0));

test("related settings are sent together in one request", async () => {
    const requests: Array<{mode: number, modeOS: boolean}> = [];
    const queue = createSettingSaveQueue({
        read: () => ({mode: 0, modeOS: false}),
        async save(data) { requests.push(data); return data; },
        apply() {},
        status() {},
    });
    queue.patchMany([{path: "mode", value: 1}, {path: "modeOS", value: true}]);
    await tick();
    assert.deepEqual(requests, [{mode: 1, modeOS: true}]);
});

test("rapid edits wait for acknowledgement and preserve earlier changes", async () => {
    let config = {spell: false, paste: false, nested: {size: 1}};
    const requests: typeof config[] = [];
    let release: (value: typeof config) => void;
    const queue = createSettingSaveQueue({
        read: () => config,
        save: async data => {
            requests.push(data);
            if (requests.length === 1) {
                return new Promise<typeof config>(resolve => { release = resolve; });
            }
            return data;
        },
        apply: data => { config = data; },
        status() {},
    });
    queue.patch("spell", true);
    queue.patch("paste", true);
    queue.patch("nested.size", 2);
    assert.equal(requests.length, 1);
    release(requests[0]);
    await tick();
    assert.deepEqual(config, {spell: true, paste: true, nested: {size: 2}});
    assert.equal(requests.length, 2);
});

test("failed edits remain queued and can be retried with newer edits", async () => {
    let config = {first: false, second: false};
    let fail = true;
    const states: string[] = [];
    const queue = createSettingSaveQueue({
        read: () => config,
        async save(data) { if (fail) { throw new Error("offline"); } return data; },
        apply: data => { config = data; },
        status: state => states.push(state),
    });
    queue.patch("first", true);
    await tick();
    assert.equal(states.at(-1), "failed");
    assert.deepEqual(config, {first: false, second: false});
    queue.patch("second", true);
    await tick();
    fail = false;
    queue.retry();
    await tick();
    assert.deepEqual(config, {first: true, second: true});
    assert.equal(states.at(-1), "saved");
});
