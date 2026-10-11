import * as assert from "node:assert/strict";
import {test} from "node:test";
import {calendarDays, localDateKey, moveCalendarDate} from "./calendarDate";

test("calendar covers leap February and leading/trailing weeks", () => {
    const days = calendarDays("2024-02");
    assert.equal(days.length, 35);
    assert.equal(days[0], "2024-01-29");
    assert.equal(days[34], "2024-03-03");
    assert.ok(days.includes("2024-02-29"));
    assert.deepEqual(calendarDays("2026-13"), []);
});

test("moving a scheduled range preserves local time and range length across DST", () => {
    const start = new Date(2026, 2, 7, 14, 30);
    const end = new Date(2026, 2, 9, 15, 45);
    const moved = moveCalendarDate({content: start.getTime(), isNotEmpty: true, hasEndDate: true, content2: end.getTime(), isNotEmpty2: true, isNotTime: false}, "2026-03-08");
    assert.equal(localDateKey(new Date(moved.content)), "2026-03-08");
    assert.equal(localDateKey(new Date(moved.content2)), "2026-03-10");
    assert.equal(new Date(moved.content).getHours(), 14);
    assert.equal(new Date(moved.content2).getMinutes(), 45);
    assert.equal(moved.isNotTime, false);
    const scheduled = moveCalendarDate({isNotEmpty: false}, "2026-10-10");
    assert.equal(localDateKey(new Date(scheduled.content)), "2026-10-10");
    assert.equal(scheduled.isNotEmpty, true);
    assert.equal(scheduled.isNotTime, true);
});
