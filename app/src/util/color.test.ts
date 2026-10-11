import * as assert from "node:assert/strict";
import {test} from "node:test";
import {canvasColor, contrastColor, normalizeColor, sanitizeColorLibrary, selectColorBackground, selectColorStyle, selectColorValue} from "./color";

test("custom colors normalize while unsafe and out-of-range values are rejected", () => {
    assert.equal(normalizeColor(" #AbC "), "#aabbcc");
    assert.equal(normalizeColor("rgb(255, 0, 128)"), "#ff0080");
    for (const value of ["rgb(256,0,0)", "rgb(-1,0,0)", "red;display:none", "url(x)", "#12345678"]) { assert.equal(normalizeColor(value), undefined); }
    assert.equal(contrastColor("#ffffff"), "#000000");
    assert.equal(contrastColor("#000000"), "#ffffff");
});

test("category rendering preserves palette IDs and never interpolates invalid stored colors", () => {
    assert.equal(selectColorStyle("14"), "background-color:var(--b3-font-background14);color:var(--b3-font-color14)");
    assert.equal(selectColorStyle("#FFFFFF"), "background-color:#ffffff;color:#000000");
    assert.equal(selectColorBackground("#123456"), "color-mix(in srgb, #123456 16%, var(--b3-theme-background))");
    assert.equal(selectColorStyle("1);color:red"), selectColorStyle("1"));
    assert.equal(selectColorValue("rgb(243, 154, 199)", "rgb(0, 0, 0)"), "#f39ac7");
    assert.equal(selectColorValue("var(--b3-font-background14)", "var(--b3-font-color14)"), "14");
});

test("saved colors tolerate corrupt storage and deduplicate normalized colors", () => {
    assert.deepEqual(sanitizeColorLibrary(null), {recent: [], palettes: []});
    assert.deepEqual(sanitizeColorLibrary({recent: ["#fff", "#FFFFFF", null, "bad"], palettes: [{id: "a", name: " Pink ", colors: ["#F0C", "bad"]}]}),
        {recent: ["#ffffff"], palettes: [{id: "a", name: "Pink", colors: ["#ff00cc"]}]});
});

test("Canvas accepts custom colors and the six JSON Canvas preset IDs", () => {
    assert.equal(canvasColor("#F0C"), "#ff00cc");
    for (const id of ["1", "2", "3", "4", "5", "6"]) { assert.match(canvasColor(id), /^#[\da-f]{6}$/); }
    assert.equal(canvasColor("7"), undefined);
    assert.equal(canvasColor("constructor"), undefined);
    assert.equal(canvasColor("red;display:none"), undefined);
});
