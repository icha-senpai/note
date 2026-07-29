"use strict";

const assert = require("node:assert/strict");
const {describe, it} = require("node:test");
const {
    PRIMARY_PROTOCOL,
    hasAppProtocol,
    normalizeAppProtocolURL,
} = require("./protocol");

describe("app protocol handling", () => {
    it("keeps scribli links as the primary protocol", () => {
        assert.equal(PRIMARY_PROTOCOL, "scribli");
        assert.equal(normalizeAppProtocolURL("scribli://blocks/20260729000000-abcdefg?focus=1"), "scribli://blocks/20260729000000-abcdefg?focus=1");
        assert.equal(hasAppProtocol("scribli://plugins/sample"), true);
    });

    it("ignores non-application URLs", () => {
        assert.equal(normalizeAppProtocolURL("other://blocks/20260729000000-abcdefg"), "");
        assert.equal(hasAppProtocol("other://blocks/20260729000000-abcdefg"), false);
        assert.equal(normalizeAppProtocolURL("https://example.com"), "");
        assert.equal(hasAppProtocol("https://example.com"), false);
        assert.equal(hasAppProtocol(undefined), false);
    });
});
