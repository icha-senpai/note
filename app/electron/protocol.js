"use strict";

const PRIMARY_PROTOCOL = "scribli";

const hasAppProtocol = (value) => {
    if (typeof value !== "string") {
        return false;
    }
    return value.startsWith(PRIMARY_PROTOCOL + "://");
};

const normalizeAppProtocolURL = (value) => {
    if (typeof value !== "string") {
        return "";
    }
    if (value.startsWith(PRIMARY_PROTOCOL + "://")) {
        return value;
    }
    return "";
};

module.exports = {
    PRIMARY_PROTOCOL,
    hasAppProtocol,
    normalizeAppProtocolURL,
};
