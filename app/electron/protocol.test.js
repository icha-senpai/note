"use strict";

const assert = require("node:assert/strict");
const {describe, it} = require("node:test");
const {
    LEGACY_PROTOCOL,
    LEGACY_PROTOCOL_REMOVAL_TARGET,
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

    it("converts the deprecated siyuan protocol alias", () => {
        const deprecations = [];
        const normalized = normalizeAppProtocolURL("siyuan://blocks/20260729000000-abcdefg", (oldURL, newURL) => {
            deprecations.push({oldURL, newURL});
        });

        assert.equal(LEGACY_PROTOCOL, "siyuan");
        assert.equal(LEGACY_PROTOCOL_REMOVAL_TARGET, "4.0.0");
        assert.equal(normalized, "scribli://blocks/20260729000000-abcdefg");
        assert.deepEqual(deprecations, [{
            oldURL: "siyuan://blocks/20260729000000-abcdefg",
            newURL: "scribli://blocks/20260729000000-abcdefg",
        }]);
        assert.equal(hasAppProtocol("siyuan://blocks/20260729000000-abcdefg"), true);
    });

    it("ignores non-application URLs", () => {
        assert.equal(normalizeAppProtocolURL("https://example.com"), "");
        assert.equal(hasAppProtocol("https://example.com"), false);
        assert.equal(hasAppProtocol(undefined), false);
    });
});
