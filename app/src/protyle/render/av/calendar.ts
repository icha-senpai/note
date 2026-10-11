import {Dialog} from "../../../dialog";
import {escapeHtml, escapeAttr} from "../../../util/escape";
import {Constants} from "../../../constants";
import {transaction} from "../../wysiwyg/transaction";
import {genTabHeaderHTML, updateSearch} from "./render";
import {bindAvSearch} from "./search";
import {calendarDays, localDateKey, calendarDateValue, moveCalendarDate} from "./calendarDate";
import {openDatabaseRowByData} from "./openDatabaseRow";
import {addCol} from "./col";
import {finishAVLocate, getAVLocateRequest} from "./locate";
import * as dayjs from "dayjs";

export const renderCalendar = (blockElement: HTMLElement, protyle: IProtyle, data: IAV, cb?: (data: IAV) => void) => {
    const view = data.view as IAVTable;
    const editable = !protyle.disabled && !protyle.options.history && !window.scribli.config.readonly;
    const dateFields = view.columns.filter(field => ["date", "created", "updated"].includes(field.type));
    const dateField = dateFields.find(field => field.id === view.calendar?.dateKeyID) || dateFields[0];
    const oldSearch = blockElement.querySelector<HTMLElement>('[data-type="av-search"]');
    const query = oldSearch?.textContent || "";
    const isSearching = oldSearch && document.activeElement === oldSearch;
    const storageKey = `scribliCalendar:${window.scribli.config.system.workspaceDir}:${data.id}:${data.viewID}`;
    let month = localDateKey(new Date()).slice(0, 7);
    try { const saved = localStorage.getItem(storageKey); if (calendarDays(saved || "").length) { month = saved; } } catch { /* Optional viewport state. */ }
    let undatedLimit = 50;
    const updateCalendar = (operations: IOperation[], undo: IOperation[]) => {
        const updated = dayjs().format("YYYYMMDDHHmmss");
        transaction(protyle, [...operations, {action: "doUpdateUpdated", id: blockElement.dataset.nodeId, data: updated}],
            [...undo, {action: "doUpdateUpdated", id: blockElement.dataset.nodeId, data: blockElement.getAttribute("updated")}]);
        blockElement.setAttribute("updated", updated);
    };
    const locateRow = view.rows.find(row => row.id === getAVLocateRequest(blockElement)?.itemID);
    if (locateRow && dateField) {
        const date = calendarDateValue(locateRow.cells[view.columns.findIndex(field => field.id === dateField.id)]);
        if (date.isNotEmpty) { month = localDateKey(new Date(date.content)).slice(0, 7); }
        else { undatedLimit = view.rows.length; }
    }

    const changeDate = (row: IAVRow, target: string) => {
        if (!editable || dateField?.type !== "date") { return; }
        const cell = row.cells.find(cell => cell.value?.keyID === dateField.id || view.columns[row.cells.indexOf(cell)]?.id === dateField.id);
        const original = cell?.value || {type: "date", date: calendarDateValue()};
        const date = target ? moveCalendarDate(calendarDateValue(cell), target) : {...calendarDateValue(cell), isNotEmpty: false, isNotEmpty2: false};
        const next: IAVCellValue = {...original, type: "date", date};
        const operation = {action: "updateAttrViewCell" as const, avID: data.id, blockID: blockElement.dataset.nodeId, keyID: dateField.id, rowID: row.id, id: cell?.id || Lute.NewNodeID()};
        updateCalendar([{...operation, data: next}], [{...operation, data: original}]);
    };
    const schedule = (row: IAVRow) => {
        const cell = row.cells[view.columns.findIndex(field => field.id === dateField.id)];
        const value = calendarDateValue(cell);
        const dialog = new Dialog({title: window.scribli.languages.calendarSchedule, width: "min(90vw, 360px)", content: `<div class="b3-dialog__content">
<input type="date" class="b3-text-field fn__block" aria-label="${window.scribli.languages.date}" value="${value.isNotEmpty ? localDateKey(new Date(value.content)) : ""}">
<div class="fn__hr"></div><div class="fn__flex"><button type="button" class="b3-button" data-calendar-action="save">${window.scribli.languages.save}</button><span class="fn__space"></span><button type="button" class="b3-button b3-button--text" data-calendar-action="clear">${window.scribli.languages.clear}</button></div></div>`});
        const input = dialog.element.querySelector<HTMLInputElement>("input");
        const save = () => { if (input.value) { changeDate(row, input.value); dialog.destroy(); } };
        dialog.element.querySelector('[data-calendar-action="save"]').addEventListener("click", save);
        dialog.element.querySelector('[data-calendar-action="clear"]').addEventListener("click", () => { changeDate(row, ""); dialog.destroy(); });
        input.addEventListener("keydown", event => { if (event.key === "Enter") { save(); } });
        input.focus();
    };
    const renderBody = () => {
        const days = calendarDays(month);
        const rowsByDay = new Map(days.map(day => [day, [] as IAVRow[]]));
        const undated: IAVRow[] = [];
        const dateIndex = dateField ? view.columns.findIndex(field => field.id === dateField.id) : -1;
        for (const row of view.rows) {
            const date = calendarDateValue(row.cells[dateIndex]);
            if (!date.isNotEmpty) { undated.push(row); continue; }
            const start = localDateKey(new Date(date.content));
            const end = date.hasEndDate && date.isNotEmpty2 ? localDateKey(new Date(date.content2)) : start;
            for (const day of days) { if (day >= start && day <= end) { rowsByDay.get(day).push(row); } }
        }
        const item = (row: IAVRow) => {
            const title = row.cells.find(cell => cell.valueType === "block")?.value?.block?.content || window.scribli.languages.untitled;
            const color = view.itemColors?.[row.id];
            return `<div class="av__calendar-item${color ? " av__rule-colored" : ""}" data-calendar-row="${row.id}"${color && /^#[0-9a-f]{6}$/i.test(color) ? ` style="--av-rule-color:${color}"` : ""}>
${editable && dateField?.type === "date" ? `<button type="button" draggable="true" class="block__icon" data-calendar-action="drag" aria-label="${window.scribli.languages.kanbanDragCard}"><svg><use xlink:href="#iconDrag"></use></svg></button>` : ""}
<button type="button" class="av__calendar-title" data-calendar-action="open" title="${escapeAttr(escapeHtml(title))}">${escapeHtml(title)}</button>
${editable && dateField?.type === "date" ? `<button type="button" class="block__icon" data-calendar-action="schedule" aria-label="${window.scribli.languages.calendarSchedule}"><svg><use xlink:href="#iconCalendar"></use></svg></button>` : ""}</div>`;
        };
        const body = blockElement.querySelector<HTMLElement>(".av__calendar");
        body.innerHTML = `<div class="av__calendar-toolbar"><button type="button" class="b3-button b3-button--text" data-calendar-action="previous" aria-label="${window.scribli.languages.previous}">‹</button>
<input class="b3-text-field" type="month" value="${month}" aria-label="${window.scribli.languages.calendarMonth}">
<button type="button" class="b3-button b3-button--text" data-calendar-action="next" aria-label="${window.scribli.languages.next}">›</button>
<button type="button" class="b3-button b3-button--text" data-calendar-action="today">${window.scribli.languages.today}</button>
<select class="b3-select" aria-label="${window.scribli.languages.calendarDateField}" ${!editable ? "disabled" : ""}>
${dateFields.length ? dateFields.map(field => `<option value="${field.id}"${dateField?.id === field.id ? " selected" : ""}>${escapeHtml(field.name || window.scribli.languages.date)}</option>`).join("") : `<option>${window.scribli.languages.calendarChooseDate}</option>`}</select>
${editable ? `<button type="button" class="b3-button b3-button--text" data-calendar-action="field">${window.scribli.languages.addCol}</button>` : ""}</div>
<div class="av__calendar-scroll"><div class="av__calendar-grid">
${days.slice(0, 7).map(day => `<div class="av__calendar-weekday">${escapeHtml(new Date(`${day}T12:00:00`).toLocaleDateString(undefined, {weekday: "short"}))}</div>`).join("")}
${days.map(day => `<div data-calendar-day="${day}" class="av__calendar-day${day.slice(0, 7) !== month ? " av__calendar-day--outside" : ""}${day === localDateKey(new Date()) ? " av__calendar-day--today" : ""}"><div class="av__calendar-number">${Number(day.slice(8))}</div>${rowsByDay.get(day).map(item).join("")}</div>`).join("")}
</div></div><details class="av__calendar-undated"${undated.length && !dateField ? " open" : ""}><summary>${window.scribli.languages.calendarUndated} (${undated.length})</summary><div class="av__calendar-inbox" data-calendar-day="">${undated.slice(0, undatedLimit).map(item).join("")}</div>${undated.length > undatedLimit ? `<button type="button" class="b3-button b3-button--text" data-calendar-action="more">${window.scribli.languages.loadMore}</button>` : ""}</details>`;
        const monthInput = body.querySelector<HTMLInputElement>('input[type="month"]');
        const setMonth = (next: string) => { if (!calendarDays(next).length) { return; } month = next; try { localStorage.setItem(storageKey, month); } catch { /* Optional viewport state. */ } renderBody(); };
        monthInput.addEventListener("change", () => setMonth(monthInput.value));
        body.querySelector("select").addEventListener("change", event => {
            const oldConfig = view.calendar || {dateKeyID: ""};
            updateCalendar([{action: "setAttrViewCalendar", avID: data.id, blockID: blockElement.dataset.nodeId, viewID: data.viewID, data: {dateKeyID: (event.target as HTMLSelectElement).value}}],
                [{action: "setAttrViewCalendar", avID: data.id, blockID: blockElement.dataset.nodeId, viewID: data.viewID, data: oldConfig}]);
        });
        body.onclick = event => {
            const target = (event.target as HTMLElement).closest<HTMLElement>("[data-calendar-action]");
            if (!target) { return; }
            event.stopPropagation(); event.preventDefault();
            const action = target.dataset.calendarAction;
            if (action === "previous" || action === "next") { const date = new Date(`${month}-01T12:00:00`); date.setMonth(date.getMonth() + (action === "next" ? 1 : -1)); setMonth(localDateKey(date).slice(0, 7)); }
            else if (action === "today") { setMonth(localDateKey(new Date()).slice(0, 7)); }
            else if (action === "more") { undatedLimit += 50; renderBody(); }
            else if (action === "field") { const rect = target.getBoundingClientRect(); addCol(protyle, blockElement).open({x: rect.left, y: rect.bottom}); }
            else {
                const row = view.rows.find(row => row.id === target.closest<HTMLElement>("[data-calendar-row]")?.dataset.calendarRow);
                if (!row) { return; }
                if (action === "schedule") { schedule(row); }
                if (action === "open") {
                    const title = row.cells.find(cell => cell.valueType === "block");
                    if (title) { openDatabaseRowByData(protyle, {avID: data.id, databaseBlockID: blockElement.dataset.nodeId, notebookID: protyle.notebookId, itemID: row.id, valueID: title.id,
                        title: title.value?.block?.content || "", boundBlockID: title.value?.block?.id, isDetached: !!title.value?.isDetached || !title.value?.block?.id}); }
                }
            }
        };
        body.ondragstart = event => {
            const target = (event.target as HTMLElement).closest<HTMLElement>('[data-calendar-action="drag"]');
            if (!target || !editable) { event.preventDefault(); return; }
            event.stopPropagation();
            event.dataTransfer.setData("application/x-scribli-calendar", `${data.id}:${target.closest<HTMLElement>("[data-calendar-row]").dataset.calendarRow}`);
            event.dataTransfer.effectAllowed = "move";
        };
        body.ondragover = event => { if (editable && event.dataTransfer.types.includes("application/x-scribli-calendar") && (event.target as HTMLElement).closest("[data-calendar-day]")) { event.preventDefault(); event.stopPropagation(); event.dataTransfer.dropEffect = "move"; } };
        body.ondrop = event => {
            const target = (event.target as HTMLElement).closest<HTMLElement>("[data-calendar-day]");
            const [avID, id] = event.dataTransfer.getData("application/x-scribli-calendar").split(":");
            const row = view.rows.find(row => row.id === id);
            if (!target || avID !== data.id || !row || !editable) { return; }
            event.preventDefault(); event.stopPropagation(); changeDate(row, target.dataset.calendarDay);
        };
    };
    // Replace the shell on database refresh; body-only navigation keeps the header search intact.
    blockElement.firstElementChild.outerHTML = `<div class="av__container fn__block">${genTabHeaderHTML(data, isSearching || !!query, editable)}<div class="av__calendar" contenteditable="false"></div><div class="av__cursor" contenteditable="true">${Constants.ZWSP}</div></div>`;
    blockElement.dataset.avType = "calendar";
    blockElement.setAttribute("data-render", "true");
    blockElement.removeAttribute(Constants.ATTRIBUTE_V_SCROLL);
    renderBody();
    if (locateRow) { blockElement.querySelector<HTMLDetailsElement>("details").open = true; }
    bindAvSearch({blockElement, query, isSearching, onChange: () => updateSearch(blockElement, protyle)});
    if (cb) { cb(data); }
    finishAVLocate(blockElement, protyle, data);
};
