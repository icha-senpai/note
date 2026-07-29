"use strict";

const PRIMARY_PROTOCOL = "scribli";
const LEGACY_PROTOCOL = "siyuan";
const LEGACY_PROTOCOL_REMOVAL_TARGET = "4.0.0";

const hasAppProtocol = (value) => {
    if (typeof value !== "string") {
        return false;
    }
    return value.startsWith(PRIMARY_PROTOCOL + "://") || value.startsWith(LEGACY_PROTOCOL + "://");
};

const normalizeAppProtocolURL = (value, onDeprecated) => {
    if (typeof value !== "string") {
        return "";
    }
    if (value.startsWith(PRIMARY_PROTOCOL + "://")) {
        return value;
    }
    if (!value.startsWith(LEGACY_PROTOCOL + "://")) {
        return "";
    }

    const normalized = PRIMARY_PROTOCOL + value.substring(LEGACY_PROTOCOL.length);
    if (typeof onDeprecated === "function") {
        onDeprecated(value, normalized);
    }
    return normalized;
};

module.exports = {
    PRIMARY_PROTOCOL,
    LEGACY_PROTOCOL,
    LEGACY_PROTOCOL_REMOVAL_TARGET,
    hasAppProtocol,
    normalizeAppProtocolURL,
};
