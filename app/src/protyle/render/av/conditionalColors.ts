import {Constants} from "../../../constants";
import {Dialog} from "../../../dialog";
import {escapeHtml} from "../../../util/escape";
import {fetchSyncPost} from "../../../util/fetch";
import {transaction} from "../../wysiwyg/transaction";
import {getFieldsByData} from "./view";
import {openColorPicker} from "../../../util/colorPicker";

const supportedTypes = ["block", "text", "url", "email", "phone", "number", "select", "mSelect", "checkbox", "date", "created", "updated"];
const dateTypes = ["date", "created", "updated"];
const textValue = (filter: IAVFilter) => {
    const value = filter.value;
    if (!value) { return ""; }
    if (["select", "mSelect"].includes(value.type)) { return value.mSelect?.[0]?.content || ""; }
    if (value.type === "checkbox") { return value.checkbox?.checked ? "true" : "false"; }
    if (value.type === "number") { return value.number?.isNotEmpty ? String(value.number.content) : ""; }
    if (dateTypes.includes(value.type)) {
        const date = value[value.type as "date"];
        if (!date?.isNotEmpty) { return ""; }
        const d = new Date(date.content);
        return `${d.getFullYear()}-${(d.getMonth() + 1).toString().padStart(2, "0")}-${d.getDate().toString().padStart(2, "0")}`;
    }
    return value.type === "block" ? value.block?.content || "" : value[value.type as "text"]?.content || "";
};

export const colorFilterValue = (type: TAVCol, text: string): IAVCellValue => {
    const value: IAVCellValue = {type};
    if (type === "number") { value.number = {content: Number(text) || 0, isNotEmpty: text.trim() !== ""}; }
    else if (["select", "mSelect"].includes(type)) { value.mSelect = text ? [{content: text, color: ""}] : []; }
    else if (type === "checkbox") { value.checkbox = {checked: text === "true"}; }
    else if (dateTypes.includes(type)) { value[type as "date"] = {content: text ? new Date(`${text}T00:00:00`).getTime() : 0, isNotEmpty: !!text, content2: 0, isNotEmpty2: false, hasEndDate: false, isNotTime: true}; }
    else if (type === "block") { value.block = {content: text}; }
    else { value[type as "text"] = {content: text}; }
    return value;
};

export const openConditionalColors = async (protyle: IProtyle, blockElement: HTMLElement) => {
    if (protyle.disabled || protyle.options.history || window.scribli.config.readonly) { return; }
    const response = await fetchSyncPost("/api/av/renderAttributeView", {id: blockElement.dataset.avId, blockID: blockElement.dataset.nodeId,
        viewID: blockElement.getAttribute(Constants.CUSTOM_SY_AV_VIEW) || "", page: 1});
    if (response.code !== 0 || !response.data) { return; }
    const data = response.data as IAV;
    const fields = getFieldsByData(data).filter(field => supportedTypes.includes(field.type));
    const original = JSON.parse(JSON.stringify(data.view.conditionalColors || [])) as IAVConditionalColor[];
    const rules = JSON.parse(JSON.stringify(original)) as IAVConditionalColor[];
    const dialog = new Dialog({title: window.scribli.languages.conditionalColors, width: "min(95vw, 800px)", content: `<div class="b3-dialog__content">
<div class="av__color-rules"></div><div class="fn__hr"></div><div class="fn__flex"><button type="button" class="b3-button b3-button--outline" data-color-action="add">${window.scribli.languages.addColorRule}</button><span class="fn__flex-1"></span><button type="button" class="b3-button" data-color-action="save">${window.scribli.languages.save}</button></div></div>`});
    const list = dialog.element.querySelector<HTMLElement>(".av__color-rules");
    const render = () => {
        list.innerHTML = rules.length ? rules.map((rule, index) => `<div class="av__color-rule" data-index="${index}">
<input type="checkbox" class="b3-switch" data-color-control="enabled" aria-label="${window.scribli.languages.enable}"${rule.enabled ? " checked" : ""}>
<select class="b3-select" data-color-control="field" aria-label="${window.scribli.languages.field}">${!fields.some(field => field.id === rule.filter.column) ? `<option value="" selected disabled>${window.scribli.languages.missingColorField}</option>` : ""}${fields.map(field => `<option value="${field.id}"${field.id === rule.filter.column ? " selected" : ""}>${escapeHtml(field.name)}</option>`).join("")}</select>
<select class="b3-select" data-color-control="operator" aria-label="${window.scribli.languages.filter}"></select><span class="av__color-value"></span>
<button type="button" class="scribli-color-picker__swatch" value="${/^#[0-9a-f]{6}$/i.test(rule.color) ? rule.color : "#ea6565"}" style="background-color:${/^#[0-9a-f]{6}$/i.test(rule.color) ? rule.color : "#ea6565"}" data-color-control="color" aria-label="${window.scribli.languages.color}"></button>
<button type="button" class="block__icon block__icon--show" data-color-action="up" aria-label="${window.scribli.languages.moveUp}"${index === 0 ? " disabled" : ""}><svg><use xlink:href="#iconUp"></use></svg></button>
<button type="button" class="block__icon block__icon--show" data-color-action="down" aria-label="${window.scribli.languages.moveDown}"${index === rules.length - 1 ? " disabled" : ""}><svg><use xlink:href="#iconDown"></use></svg></button>
<button type="button" class="block__icon block__icon--show" data-color-action="remove" aria-label="${window.scribli.languages.remove}"><svg><use xlink:href="#iconTrashcan"></use></svg></button></div>`).join("") : `<div class="ft__on-surface">${window.scribli.languages.noColorRules}</div>`;
        list.querySelectorAll<HTMLElement>("[data-index]").forEach(element => {
            const rule = rules[Number(element.dataset.index)];
            const field = fields.find(field => field.id === rule.filter.column);
            const operators: Array<[string, string]> = [["=", "="], ["!=", "≠"], ["Is empty", window.scribli.languages.conditionEmpty], ["Is not empty", window.scribli.languages.conditionNotEmpty]];
            if (field?.type === "number" || dateTypes.includes(field?.type)) { operators.push(["<", "<"], [">", ">"], ["<=", "≤"], [">=", "≥"]); }
            else if (field?.type !== "checkbox") { operators.push(["Contains", window.scribli.languages.conditionContains]); }
            if (dateTypes.includes(field?.type)) { operators.push(["beforeToday", window.scribli.languages.beforeToday], ["onToday", window.scribli.languages.today], ["afterToday", window.scribli.languages.afterToday]); }
            const operator = element.querySelector<HTMLSelectElement>('[data-color-control="operator"]');
            const relativeOperators: Record<string, string> = {"<": "beforeToday", "=": "onToday", ">": "afterToday"};
            const selected = rule.filter.relativeDate ? (relativeOperators[rule.filter.operator] || rule.filter.operator) : rule.filter.operator;
            operator.innerHTML = operators.map(([value, label]) => `<option value="${value}"${selected === value ? " selected" : ""}>${escapeHtml(label)}</option>`).join("");
            const wrapper = element.querySelector<HTMLElement>(".av__color-value");
            const relativeDate = rule.filter.relativeDate;
            const preserve = !!rule.filter.filters?.length || !!rule.filter.quantifier ||
                !!rule.filter.relativeDate2 || (relativeDate && (relativeDate.count !== 0 || relativeDate.unit !== 0 || relativeDate.direction !== 0)) ||
                (field && !operators.some(([value]) => value === selected)) ||
                (rule.filter.value?.mSelect?.length || 0) > 1 || !!rule.filter.value?.date?.hasEndDate;
            element.dataset.colorPreserve = preserve ? "true" : "false";
            if (preserve) {
                element.querySelector<HTMLSelectElement>('[data-color-control="field"]').disabled = true;
                element.querySelector<HTMLSelectElement>('[data-color-control="field"]').classList.add("fn__none");
                operator.disabled = true;
                operator.classList.add("fn__none");
                wrapper.textContent = window.scribli.languages.advancedColorCondition;
                return;
            }
            if (["Is empty", "Is not empty", "beforeToday", "onToday", "afterToday"].includes(selected)) { return; }
            let input: HTMLInputElement | HTMLSelectElement;
            if (field?.type === "checkbox" || ["select", "mSelect"].includes(field?.type)) {
                input = document.createElement("select"); input.className = "b3-select";
                const options = field.type === "checkbox" ? [{name: window.scribli.languages.checked, value: "true"}, {name: window.scribli.languages.unchecked, value: "false"}] : (field.options || []).map(option => ({name: option.name, value: option.name}));
                options.forEach(option => { const item = document.createElement("option"); item.value = option.value; item.textContent = option.name; input.appendChild(item); });
            } else { input = document.createElement("input"); input.type = dateTypes.includes(field?.type) ? "date" : field?.type === "number" ? "number" : "text"; input.className = "b3-text-field"; }
            input.setAttribute("data-color-control", "value"); input.setAttribute("aria-label", window.scribli.languages.value);
            input.value = textValue(rule.filter); wrapper.appendChild(input);
        });
    };
    const readRule = (element: HTMLElement) => {
        const rule = rules[Number(element.dataset.index)];
        rule.enabled = element.querySelector<HTMLInputElement>('[data-color-control="enabled"]').checked;
        rule.color = element.querySelector<HTMLInputElement>('[data-color-control="color"]').value;
        if (element.dataset.colorPreserve === "true") { return; }
        const field = fields.find(field => field.id === element.querySelector<HTMLSelectElement>('[data-color-control="field"]').value);
        if (!field) { return; }
        const operator = element.querySelector<HTMLSelectElement>('[data-color-control="operator"]').value;
        const relative = ["beforeToday", "onToday", "afterToday"].includes(operator);
        rule.filter = {column: field.id, operator: relative ? ({beforeToday: "<", onToday: "=", afterToday: ">"}[operator] as IAVFilter["operator"]) : operator as IAVFilter["operator"],
            value: colorFilterValue(field.type, element.querySelector<HTMLInputElement | HTMLSelectElement>('[data-color-control="value"]')?.value || "")};
        if (relative) { rule.filter.relativeDate = {count: 0, unit: 0, direction: 0}; }
    };
    list.addEventListener("change", event => {
        const element = (event.target as HTMLElement).closest<HTMLElement>("[data-index]");
        if (!element) { return; }
        if ((event.target as HTMLElement).dataset.colorControl === "field") {
            const rule = rules[Number(element.dataset.index)];
            const field = fields.find(field => field.id === (event.target as HTMLSelectElement).value);
            rule.filter = {column: field.id, operator: "Is not empty", value: colorFilterValue(field.type, "")};
        } else { readRule(element); }
        if (["field", "operator"].includes((event.target as HTMLElement).dataset.colorControl)) { render(); }
    });
    dialog.element.addEventListener("click", event => {
        const colorButton = (event.target as HTMLElement).closest<HTMLButtonElement>('[data-color-control="color"]');
        if (colorButton) {
            const element = colorButton.closest<HTMLElement>("[data-index]");
            openColorPicker({color: colorButton.value, onApply(color) {
                colorButton.value = color;
                colorButton.style.backgroundColor = color;
                readRule(element);
            }});
            return;
        }
        const target = (event.target as HTMLElement).closest<HTMLElement>("[data-color-action]");
        if (!target) { return; }
        const action = target.dataset.colorAction;
        if (action === "save") {
            list.querySelectorAll<HTMLElement>("[data-index]").forEach(readRule);
            transaction(protyle, [{action: "setAttrViewConditionalColors", avID: data.id, blockID: blockElement.dataset.nodeId, viewID: data.viewID, data: rules}],
                [{action: "setAttrViewConditionalColors", avID: data.id, blockID: blockElement.dataset.nodeId, viewID: data.viewID, data: original}]);
            dialog.destroy(); return;
        }
        if (action === "add") {
            if (!fields.length || rules.length >= 50) { return; }
            const field = fields.find(field => ["select", "mSelect", "checkbox", "date"].includes(field.type)) || fields[0];
            rules.push({id: Lute.NewNodeID(), enabled: true, color: "#ea6565", filter: {column: field.id, operator: "Is not empty", value: colorFilterValue(field.type, field.options?.[0]?.name || "")}});
        } else {
            const index = Number(target.closest<HTMLElement>("[data-index]").dataset.index);
            if (action === "remove") { rules.splice(index, 1); }
            if (action === "up" && index > 0) { [rules[index], rules[index - 1]] = [rules[index - 1], rules[index]]; }
            if (action === "down" && index < rules.length - 1) { [rules[index], rules[index + 1]] = [rules[index + 1], rules[index]]; }
        }
        render();
    });
    render();
};
