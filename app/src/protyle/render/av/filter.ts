import {Menu} from "../../../plugin/Menu";
import {transaction} from "../../wysiwyg/transaction";
import {escapeAttr, escapeHtml} from "../../../util/escape";
import {getColIconByType} from "./col";
import {setPosition} from "../../../util/setPosition";
import {genCellValue} from "./cell";
import * as dayjs from "dayjs";
import {unicode2Emoji} from "../../../emoji";
import {fetchPost, fetchSyncPost} from "../../../util/fetch";
import {getFieldsByData} from "./view";
import {Constants} from "../../../constants";

export const getDefaultOperatorByType = (type: TAVCol) => {
    if (["select", "number", "date", "created", "updated"].includes(type)) {
        return "=";
    }
    if (["checkbox"].includes(type)) {
        return "=";
    }
    if (["rollup", "relation", "mAsset", "text", "mSelect", "url", "block", "email", "phone", "template"].includes(type)) {
        return "Contains";
    }
};

export const getEditableFilters = (data: IAV): IAVFilter[] => {
    if (data.view.filters.length === 1 && (data.view.filters[0].filters || data.view.filters[0].combination)) {
        if (!data.view.filters[0].filters) {
            data.view.filters[0].filters = [];
        }
        return data.view.filters[0].filters;
    }
    return data.view.filters;
};

const getRootFilters = (data: IAV): IAVFilter[] => getEditableFilters(data);

export const getFilterByPath = (nodes: IAVFilter[], path: string): IAVFilter => {
    if (!path || "" === path) {
        return undefined;
    }
    const indices = path.split(",").map(i => parseInt(i, 10));
    let current: IAVFilter;
    let list = nodes;
    for (let i = 0; i < indices.length; i++) {
        const idx = indices[i];
        if (!list || isNaN(idx) || idx < 0 || idx >= list.length) {
            return undefined;
        }
        current = list[idx];
        list = current.filters;
    }
    return current;
};

export const getParentByPath = (nodes: IAVFilter[], path: string): { parent: IAVFilter[], index: number } => {
    if (!path || "" === path) {
        return {parent: nodes, index: -1};
    }
    const indices = path.split(",").map(i => parseInt(i, 10));
    const lastIndex = indices.pop();
    if (isNaN(lastIndex) || lastIndex < 0) {
        return {parent: null, index: -1};
    }
    let list = nodes;
    for (const idx of indices) {
        if (!list || isNaN(idx) || idx < 0 || idx >= list.length) {
            return {parent: null, index: -1};
        }
        list = list[idx].filters || (list[idx].filters = []);
    }
    return {parent: list, index: lastIndex};
};

export const removeFilterByPath = (nodes: IAVFilter[], path: string): boolean => {
    const {parent, index} = getParentByPath(nodes, path);
    if (!parent || index < 0 || index >= parent.length) {
        return false;
    }
    parent.splice(index, 1);
    if (parent.length === 0 && path.includes(",")) {
        const groupPath = path.substring(0, path.lastIndexOf(","));
        removeFilterByPath(nodes, groupPath);
    }
    return true;
};

export const removeFiltersByColumn = (filters: IAVFilter[], column: string): IAVFilter[] => {
    const ret: IAVFilter[] = [];
    filters.forEach(f => {
        if (f.filters) {
            const children = removeFiltersByColumn(f.filters, column);
            if (children.length > 0) {
                ret.push({...f, filters: children});
            }
        } else if (f.column !== column) {
            ret.push(f);
        }
    });
    return ret;
};

export const hasFilterForColumn = (filters: IAVFilter[], column: string): boolean => {
    for (const f of filters) {
        if (f.filters) {
            if (hasFilterForColumn(f.filters, column)) {
                return true;
            }
        } else if (f.column === column) {
            return true;
        }
    }
    return false;
};

export const addFilterGroup = (data: IAV, path: string) => {
    let target: IAVFilter[];
    if ("" === path) {
        target = getEditableFilters(data);
    } else {
        const node = getFilterByPath(getRootFilters(data), path);
        if (!node) {
            target = getEditableFilters(data);
        } else {
            target = node.filters || (node.filters = []);
        }
    }
    target.push({combination: "and", filters: []});
};

export const addFilter = (options: {
    data: IAV,
    rect: DOMRect,
    menuElement: HTMLElement,
    tabRect: DOMRect,
    avId: string,
    protyle: IProtyle
    blockElement: Element,
    parentPath?: string
}) => {
    const menu = new Menu(Constants.MENU_AV_ADD_FILTER);
    let targetGroupFilters: IAVFilter[];
    if (options.parentPath && options.parentPath !== "") {
        const node = getFilterByPath(getRootFilters(options.data), options.parentPath);
        targetGroupFilters = node && node.filters ? node.filters : getEditableFilters(options.data);
    } else {
        targetGroupFilters = getEditableFilters(options.data);
    }
    getFieldsByData(options.data).forEach((column) => {
        if (column.type !== "lineNumber") {
            menu.addItem({
                label: column.name,
                iconHTML: column.icon ? unicode2Emoji(column.icon, "b3-menu__icon", true) : `<svg class="b3-menu__icon"><use xlink:href="#${getColIconByType(column.type)}"></use></svg>`,
                click: () => {
                    const {operator, value} = genEmptyFilterValue(column);
                    const filter: IAVFilter = {
                        column: column.id,
                        operator,
                        value,
                    };
                    const oldFilters = JSON.parse(JSON.stringify(options.data.view.filters));
                    targetGroupFilters.push(filter);
                    const blockID = options.blockElement.getAttribute("data-node-id");
                    transaction(options.protyle, [{
                        action: "setAttrViewFilters",
                        avID: options.avId,
                        data: JSON.parse(JSON.stringify(options.data.view.filters)),
                        blockID
                    }], [{
                        action: "setAttrViewFilters",
                        avID: options.avId,
                        data: oldFilters,
                        blockID
                    }]);
                    options.menuElement.innerHTML = getFiltersHTML(options.data);
                    setPosition(options.menuElement, options.tabRect.right - options.menuElement.clientWidth, options.tabRect.bottom, options.tabRect.height, 0, true);
                }
            });
        }
    });
    menu.open({
        x: options.rect.left,
        y: options.rect.bottom,
        h: options.rect.height,
    });
};

export const getFiltersHTML = (data: IAV) => {
    let html = "";
    const fields = getFieldsByData(data);
    const measureEl = document.createElement("span");
    measureEl.style.cssText = "position:absolute;visibility:hidden;font-size:14px;white-space:nowrap;";
    document.body.appendChild(measureEl);
    let andOrTextWidth = 0;
    [window.scribli.languages.filterWhen, window.scribli.languages.filterCombinationAnd, window.scribli.languages.filterCombinationOr].forEach(t => {
        measureEl.textContent = t;
        andOrTextWidth = Math.max(andOrTextWidth, measureEl.offsetWidth);
    });
    document.body.removeChild(measureEl);
    const andOrControlWidth = andOrTextWidth + 36;
    const genAndOrSelect = (groupPath: string, combination: string) =>
        `<select class="b3-select" data-type="toggleCombination" data-path="${groupPath}" style="width:${andOrControlWidth}px;"><option value="and" ${combination === "and" ? "selected" : ""}>${window.scribli.languages.filterCombinationAnd}</option><option value="or" ${combination === "or" ? "selected" : ""}>${window.scribli.languages.filterCombinationOr}</option></select>`;

    const genWhenLabel = () =>
        `<span class="av__filter-label ft__on-surface" style="width:${andOrControlWidth}px;">${window.scribli.languages.filterWhen}</span>`;

    const genAndOrLabel = (combination: string) =>
        `<span class="av__filter-label ft__on-surface" style="width:${andOrControlWidth}px;">${combination === "or" ? window.scribli.languages.filterCombinationOr : window.scribli.languages.filterCombinationAnd}</span>`;

    const genNodeHTML = (node: IAVFilter, path: string, depth: number, groupPath: string, groupCombination: string, index: number = 0): string => {
        if (!node) {
            return "";
        }
        if (node.filters) {
            const isRoot = 0 === depth;
            const combination = node.combination === "or" ? "or" : "and";
            let childrenHTML = "";
            node.filters.forEach((child, index) => {
                const childPath = path ? `${path},${index}` : `${index}`;
                childrenHTML += genNodeHTML(child, childPath, depth + 1, path, combination, index);
            });

            if (isRoot) {
                return childrenHTML;
            }

            const depthClass = `av__filter-group-children--depth${Math.min(depth, 3)}`;
            const addConditionBtn = depth >= 3
                ? `<span class="block__icon block__icon--text ariaLabel" data-position="4north" data-type="addFilter" data-path="${path}" aria-label="${window.scribli.languages.addFilterCondition}"><svg><use xlink:href="#iconAdd"></use></svg>${window.scribli.languages.addFilterCondition}</span>`
                : `<span class="block__icon block__icon--text ariaLabel" data-position="4north" data-type="addFilterCondition" data-path="${path}" data-depth="${depth}" aria-label="${window.scribli.languages.addFilterCondition}"><svg><use xlink:href="#iconAdd"></use></svg>${window.scribli.languages.addFilterCondition}<svg><use xlink:href="#iconDown"></use></svg></span>`;

            const andOrHTML = 0 === index ? genWhenLabel() : 1 === index ? genAndOrSelect(groupPath, groupCombination) : genAndOrLabel(groupCombination);
            return `<div class="av__filter-group-item" data-path="${path}">
    <span class="av__filter-group-left">
        ${andOrHTML}
    </span>
    <div class="av__filter-group-children ${depthClass}" data-children="${path}">
        ${childrenHTML}
        <div class="av__filter-group-actions">${addConditionBtn}</div>
    </div>
    <svg class="b3-menu__action ariaLabel" data-position="4west" data-type="moreFilter" data-path="${path}" aria-label="${window.scribli.languages.more}"><use xlink:href="#iconMore"></use></svg>
</div>`;
        }

        let colData: IAVColumn;
        fields.find((column: IAVColumn) => {
            if (column.id === node.column) {
                colData = column;
                return true;
            }
        });
        if (!colData) {
            return "";
        }
        const iconHTML = colData.icon
            ? unicode2Emoji(colData.icon, "b3-menu__icon", true)
            : `<svg class="b3-menu__icon"><use xlink:href="#${getColIconByType(colData.type)}"></use></svg>`;
        const fieldOptions = fields.filter((f: IAVColumn) => f.type !== "lineNumber").map((f: IAVColumn) =>
            `<option value="${f.id}" ${f.id === node.column ? "selected" : ""}>${escapeHtml(f.name)}</option>`
        ).join("");
        const fieldSelect = `<select class="b3-select fn__flex-1 av__filter-field" data-type="fieldSelect" data-path="${path}">${fieldOptions}</select>`;
        const fieldWrapper = `<span class="av__field-wrapper ariaLabel" data-position="4west" aria-label="${escapeAttr(colData.name)}">${iconHTML}${fieldSelect}</span>`;
        const inlineHTML = genInlineFilterHTML(node, colData, path);
        const leafAndOrHTML = 0 === index ? genWhenLabel() : 1 === index ? genAndOrSelect(groupPath, groupCombination) : genAndOrLabel(groupCombination);
        return `<div class="b3-menu__item av__filter-row" data-path="${path}" data-column="${node.column}">${leafAndOrHTML}<div class="fn__flex-1 av__filter-rowinner">${fieldWrapper}${inlineHTML}</div><svg class="b3-menu__action ariaLabel" data-position="4west" data-type="moreFilter" data-path="${path}" aria-label="${window.scribli.languages.more}"><use xlink:href="#iconMore"></use></svg></div>`;
    };

    const isRootGroup = data.view.filters.length === 1 && (data.view.filters[0].filters || data.view.filters[0].combination);
    const root = isRootGroup ? data.view.filters[0] : {filters: data.view.filters} as IAVFilter;
    const rootCombination = isRootGroup
        ? (data.view.filters[0].combination === "or" ? "or" : "and")
        : "and";
    html = genNodeHTML(root, "", 0, "", rootCombination);

    const countLeaves = (nodes: IAVFilter[]): number => nodes.reduce((sum, n) => sum + (n.filters ? countLeaves(n.filters) : 1), 0);
    const leafCount = countLeaves(root.filters || []);

    return `<div class="b3-menu__items">
<button class="b3-menu__item" data-type="nobg">
    <span class="block__icon" style="padding: 8px;margin-left: -4px;" data-type="go-config">
        <svg><use xlink:href="#iconLeft"></use></svg>
    </span>
    <span class="b3-menu__label ft__center">${window.scribli.languages.filter}</span>
</button>
<button class="b3-menu__separator"></button>
${html}
<button class="b3-menu__item" data-type="addFilterCondition" data-path="" data-depth="0">
    <svg class="b3-menu__icon"><use xlink:href="#iconAdd"></use></svg>
    <span class="b3-menu__label av__filter-add-label">${window.scribli.languages.addFilterCondition}</span>
    <svg class="av__filter-arrow"><use xlink:href="#iconDown"></use></svg>
</button>
<button class="b3-menu__item b3-menu__item--warning${leafCount > 0 ? "" : " fn__none"}" data-type="removeFilters">
    <svg class="b3-menu__icon"><use xlink:href="#iconTrashcan"></use></svg>
    <span class="b3-menu__label">${window.scribli.languages.removeFilters}</span>
</button>
</div>`;
};

export const duplicateFilterByPath = (nodes: IAVFilter[], path: string): boolean => {
    const {parent, index} = getParentByPath(nodes, path);
    if (!parent || index < 0 || index >= parent.length) {
        return false;
    }
    const clone = JSON.parse(JSON.stringify(parent[index]));
    parent.splice(index + 1, 0, clone);
    return true;
};

export const convertFilterToGroup = (nodes: IAVFilter[], path: string): boolean => {
    const {parent, index} = getParentByPath(nodes, path);
    if (!parent || index < 0 || index >= parent.length) {
        return false;
    }
    const node = parent[index];
    if (node.filters) {
        return false;
    }
    const group: IAVFilter = {
        combination: "and",
        filters: [node],
    };
    parent.splice(index, 1, group);
    return true;
};

export const convertGroupToFilter = (nodes: IAVFilter[], path: string): boolean => {
    const {parent, index} = getParentByPath(nodes, path);
    if (!parent || index < 0 || index >= parent.length) {
        return false;
    }
    const node = parent[index];
    if (!node.filters || 1 !== node.filters.length) {
        return false;
    }
    parent.splice(index, 1, node.filters[0]);
    return true;
};


const getOperatorSelectByType = (type: TAVCol, currentOperator: string): string => {
    const opt = (value: string, label: string) => `<option ${value === currentOperator ? "selected" : ""} value="${value}">${label}</option>`;
    switch (type) {
        case "checkbox":
            return opt("=", window.scribli.languages.filterOperatorIs) + opt("!=", window.scribli.languages.filterOperatorIsNot);
        case "block":
        case "mAsset":
        case "text":
        case "url":
        case "phone":
        case "email":
            return opt("=", window.scribli.languages.filterOperatorIs) + opt("!=", window.scribli.languages.filterOperatorIsNot) +
                opt("Contains", window.scribli.languages.filterOperatorContains) + opt("Does not contains", window.scribli.languages.filterOperatorDoesNotContain) +
                opt("Starts with", window.scribli.languages.filterOperatorStartsWith) + opt("Ends with", window.scribli.languages.filterOperatorEndsWith) +
                opt("Is empty", window.scribli.languages.filterOperatorIsEmpty) + opt("Is not empty", window.scribli.languages.filterOperatorIsNotEmpty);
        case "template":
            return opt("=", window.scribli.languages.filterOperatorIs) + opt("!=", window.scribli.languages.filterOperatorIsNot) +
                opt("Contains", window.scribli.languages.filterOperatorContains) + opt("Does not contains", window.scribli.languages.filterOperatorDoesNotContain) +
                opt("Starts with", window.scribli.languages.filterOperatorStartsWith) + opt("Ends with", window.scribli.languages.filterOperatorEndsWith) +
                opt("Is empty", window.scribli.languages.filterOperatorIsEmpty) + opt("Is not empty", window.scribli.languages.filterOperatorIsNotEmpty) +
                opt(">", "&gt;") + opt("<", "&lt;") + opt(">=", "&GreaterEqual;") + opt("<=", "&le;");
        case "date":
        case "created":
        case "updated":
            return opt("=", window.scribli.languages.filterOperatorIs) + opt(">", window.scribli.languages.filterOperatorIsAfter) +
                opt("<", window.scribli.languages.filterOperatorIsBefore) + opt(">=", window.scribli.languages.filterOperatorIsOnOrAfter) +
                opt("<=", window.scribli.languages.filterOperatorIsOnOrBefore) + opt("Is between", window.scribli.languages.filterOperatorIsBetween) +
                opt("Is empty", window.scribli.languages.filterOperatorIsEmpty) + opt("Is not empty", window.scribli.languages.filterOperatorIsNotEmpty);
        case "number":
            return opt("=", "=") + opt("!=", "!=") + opt(">", "&gt;") + opt("<", "&lt;") +
                opt(">=", "&GreaterEqual;") + opt("<=", "&le;") +
                opt("Is empty", window.scribli.languages.filterOperatorIsEmpty) + opt("Is not empty", window.scribli.languages.filterOperatorIsNotEmpty);
        case "mSelect":
        case "relation":
            return opt("Contains", window.scribli.languages.filterOperatorContains) + opt("Does not contains", window.scribli.languages.filterOperatorDoesNotContain) +
                opt("Is empty", window.scribli.languages.filterOperatorIsEmpty) + opt("Is not empty", window.scribli.languages.filterOperatorIsNotEmpty);
        case "select":
            return opt("=", window.scribli.languages.filterOperatorIs) + opt("!=", window.scribli.languages.filterOperatorIsNot) +
                opt("Is empty", window.scribli.languages.filterOperatorIsEmpty) + opt("Is not empty", window.scribli.languages.filterOperatorIsNotEmpty);
        default:
            return "";
    }
};

const rollupTargetColumns = new WeakMap<IAVColumn, IAVColumn>();

export const prepareFilterColumns = async (data: IAV) => {
    const fields = getFieldsByData(data);
    const avRequests = new Map<string, Promise<IAVColumn[]>>();
    const tasks = fields.filter((column) => column.type === "rollup" && column.rollup?.relationKeyID && column.rollup?.keyID).map(async (column) => {
        const relationColumn = fields.find((item) => item.id === column.rollup.relationKeyID);
        const targetAVID = relationColumn?.relation?.avID;
        if (!targetAVID) {
            return;
        }
        let request = avRequests.get(targetAVID);
        if (!request) {
            request = fetchSyncPost("/api/av/getAttributeView", {id: targetAVID}).then((response) => {
                return (response.data?.av?.keyValues || []).map((item: { key: IAVColumn }) => item.key);
            }).catch(() => []);
            avRequests.set(targetAVID, request);
        }
        const targetColumns = await request;
        const targetColumn = targetColumns.find((item) => item.id === column.rollup.keyID);
        if (targetColumn) {
            rollupTargetColumns.set(column, targetColumn);
        }
    });
    await Promise.all(tasks);
};

const resolveFilterValueType = (filter: IAVFilter, colData: IAVColumn): { type: TAVCol, colData: IAVColumn, isRollup: boolean } => {
    const valueType = filter.value?.type as TAVCol;
    if (valueType !== "rollup") {
        return {type: valueType, colData, isRollup: false};
    }
    const targetColumn = rollupTargetColumns.get(colData);
    const rollup = filter.value?.rollup;
    const contentType = rollup?.contents?.[0]?.type as TAVCol;
    const calcOperator = colData.rollup?.calc?.operator;
    const numberOperators = [
        "Count all", "Count values", "Count unique values", "Count empty", "Count not empty",
        "Percent empty", "Percent not empty", "Percent unique values", "Sum", "Average", "Median", "Min", "Max",
        "Checked", "Unchecked", "Percent checked", "Percent unchecked",
    ];
    const resolvedType = numberOperators.includes(calcOperator)
        ? "number"
        : targetColumn?.type || contentType || "text";
    return {type: resolvedType, colData: targetColumn || colData, isRollup: true};
};

const getFilterCellValue = (filter: IAVFilter) => filter.value?.type === "rollup"
    ? filter.value.rollup?.contents?.[0]
    : filter.value;

const escapeFilterValue = (value: string) => escapeAttr(escapeHtml(value));

const genEmptyCellValue = (type: TAVCol): IAVCellValue => type === "checkbox"
    ? genCellValue(type, {checked: undefined})
    : {type} as IAVCellValue;

const genEmptyFilterValue = (column: IAVColumn): { operator: TAVFilterOperator, value: IAVCellValue } => {
    if (column.type !== "rollup") {
        return {
            operator: getDefaultOperatorByType(column.type),
            value: genEmptyCellValue(column.type),
        };
    }
    const emptyRollup = {type: "rollup", rollup: {contents: []}} as IAVCellValue;
    const {type} = resolveFilterValueType({value: emptyRollup} as IAVFilter, column);
    return {
        operator: getDefaultOperatorByType(type),
        value: {
            type: "rollup",
            rollup: {contents: [genEmptyCellValue(type)]},
        } as IAVCellValue,
    };
};

const genInlineFilterHTML = (filter: IAVFilter, colData: IAVColumn, path: string): string => {
    const {type: valueType, colData: valueColumn, isRollup} = resolveFilterValueType(filter, colData);
    const operator = filter.operator;
    const isEmptyOp = operator === "Is empty" || operator === "Is not empty";
    const valueHidden = isEmptyOp ? " fn__none" : "";

    const operatorSelect = `<select class="b3-select" data-type="operation" data-path="${path}">${getOperatorSelectByType(valueType, operator)}</select>`;

    const quantifierSelect = (isRollup || valueType === "mAsset")
        ? `<select class="b3-select" data-type="quantifier" data-path="${path}">
<option ${(!filter.quantifier || filter.quantifier === "Any") ? "selected" : ""} value="Any">${window.scribli.languages.filterQuantifierAny}</option>
<option ${filter.quantifier === "All" ? "selected" : ""} value="All">${window.scribli.languages.filterQuantifierAll}</option>
<option ${filter.quantifier === "None" ? "selected" : ""} value="None">${window.scribli.languages.filterQuantifierNone}</option>
</select>`
        : "";

    let valueHTML = "";
    let extraHTML = "";
    const filterValue = getFilterCellValue(filter);
    if (["text", "url", "block", "email", "phone", "template"].includes(valueType)) {
        const content = filterValue?.[valueType as "text"]?.content || "";
        valueHTML = `<input class="b3-text-field b3-text-field--text fn__flex-1" value="${escapeFilterValue(content)}" data-type="filterValue" data-path="${path}">`;
    } else if (valueType === "mAsset") {
        const content = filterValue?.mAsset?.[0]?.content || "";
        valueHTML = `<input class="b3-text-field b3-text-field--text fn__flex-1" value="${escapeFilterValue(content)}" data-type="filterValue" data-path="${path}">`;
    } else if (valueType === "number") {
        const content = filterValue?.number?.isNotEmpty ? filterValue.number.content : "";
        valueHTML = `<input class="b3-text-field b3-text-field--text av__filter-num" value="${content}" data-type="filterValue" data-path="${path}">`;
    } else if (valueType === "checkbox") {
        const isChecked = filterValue?.checkbox?.checked;
        valueHTML = `<select class="b3-select" data-type="filterValue" data-path="${path}"><option value="true" ${isChecked ? "selected" : ""}>${window.scribli.languages.checked}</option><option value="false" ${!isChecked ? "selected" : ""}>${window.scribli.languages.unchecked}</option></select>`;
    } else if (["date", "created", "updated"].includes(valueType)) {
        valueHTML = genInlineDateHTML(filter, valueType, path);
    } else if (valueType === "select" || valueType === "mSelect") {
        const {trigger, dropdown} = genInlineSelectHTML(filter, valueColumn, path, valueType);
        valueHTML = trigger;
        extraHTML = dropdown;
    } else if (valueType === "relation") {
        const content = filterValue?.relation?.blockIDs?.[0] || "";
        valueHTML = `<input class="b3-text-field b3-text-field--text fn__flex-1" value="${escapeFilterValue(content)}" data-type="filterValue" data-type-rel="relation" data-path="${path}">`;
    }

    return `${quantifierSelect}${operatorSelect}<span class="av__filter-value${valueHidden}" data-type="valueContainer" data-path="${path}">${valueHTML}</span>${extraHTML}`;
};

const genInlineDateHTML = (filter: IAVFilter, valueType: TAVCol, path: string): string => {
    const dateValue = getFilterCellValue(filter)?.[valueType as "date"];
    const showToday1 = !filter.relativeDate?.direction;
    const showToday2 = !filter.relativeDate2?.direction;
    const isBetween = filter.operator === "Is between";

    const formatAbsDate = (timestamp: any): string => {
        if (!timestamp) {
            return "";
        }
        const dayObj = dayjs(timestamp);
        return dayObj.isValid() ? dayObj.format("YYYY-MM-DD") : "";
    };

    const dateBlock = (suffix: "" | "2", relativeDate: IAVRelativeDate, dateVal: any, showToday: boolean): string => {
        const dateTypeSel = `<select class="b3-select" data-type="dateType${suffix}" data-path="${path}">
<option value="time"${!relativeDate ? " selected" : ""}>${window.scribli.languages.includeTime}</option>
<option value="custom"${relativeDate ? " selected" : ""}>${window.scribli.languages.relativeToToday}</option>
</select>`;
        const absDate = `<input value="${(dateVal && (dateVal.isNotEmpty || (suffix === "2" ? dateVal.isNotEmpty2 : valueType !== "date"))) ? formatAbsDate(suffix === "2" ? dateVal.content2 : dateVal.content) : ""}" type="date" max="9999-12-31" class="b3-text-field b3-text-field--text" data-type="absDate${suffix}" data-path="${path}" style="${relativeDate ? "display:none;" : ""}">`;
        const relDir = `<select class="b3-select" data-type="dataDirection${suffix}" data-path="${path}" style="${!relativeDate ? "display:none;" : ""}">
<option value="-1"${relativeDate?.direction === -1 ? " selected" : ""}>${window.scribli.languages.pastDate}</option>
<option value="1"${relativeDate?.direction === 1 ? " selected" : ""}>${window.scribli.languages.nextDate}</option>
<option value="0"${showToday ? " selected" : ""}>${window.scribli.languages.current}</option>
</select>`;
        const relCount = `<input type="number" min="1" step="1" value="${relativeDate?.count || 1}" class="b3-text-field b3-text-field--text av__filter-num" data-type="relCount${suffix}" data-path="${path}" style="${(!relativeDate || showToday) ? "display:none;" : ""}">`;
        const relUnit = `<select class="b3-select" data-type="relUnit${suffix}" data-path="${path}" style="${!relativeDate ? "display:none;" : ""}">
<option value="0"${relativeDate?.unit === 0 ? " selected" : ""}>${window.scribli.languages.day}</option>
<option value="1"${(!relativeDate || relativeDate?.unit === 1) ? " selected" : ""}>${window.scribli.languages.week}</option>
<option value="2"${relativeDate?.unit === 2 ? " selected" : ""}>${window.scribli.languages.month}</option>
<option value="3"${relativeDate?.unit === 3 ? " selected" : ""}>${window.scribli.languages.year}</option>
</select>`;
        return `<span class="av__filter-date-row">${dateTypeSel}${absDate}${relDir}${relCount}${relUnit}</span>`;
    };

    const filter1 = dateBlock("", filter.relativeDate, dateValue, showToday1);
    const filter2 = dateBlock("2", filter.relativeDate2, dateValue, showToday2);
    return `<span class="av__filter-date-col">${filter1}<span data-type="filter2Wrap" data-path="${path}" style="${isBetween ? "" : "display:none;"}">${filter2}</span></span>`;
};

const genInlineSelectHTML = (filter: IAVFilter, colData: IAVColumn, path: string, valueType: TAVCol): { trigger: string, dropdown: string } => {
    const isSingle = valueType === "select";
    const options = colData.options || [];
    const selectedValues = (getFilterCellValue(filter)?.mSelect || []).filter((s: IAVCellSelectValue) => s.content);
    const placeholder = isSingle ? window.scribli.languages.select : window.scribli.languages.multiSelect;

    const selectedChips = selectedValues.map((item: IAVCellSelectValue) => {
        return `<span class="b3-chip b3-chip--middle av__select-chip" style="background-color:var(--b3-font-background${item.color});color:var(--b3-font-color${item.color})">${escapeHtml(item.content)}</span>`;
    }).join("");
    const triggerContent = selectedChips || `<span class="ft__on-surface fn__ellipsis">${placeholder}</span>`;
    const trigger = `<span class="av__select-trigger" data-type="selectTrigger" data-path="${path}">${triggerContent}<svg class="av__select-trigger-arrow"><use xlink:href="#iconDown"></use></svg></span>`;

    const searchInput = options.length > 5
        ? `<input class="b3-text-field" placeholder="${window.scribli.languages.search}" data-type="filterSearch" data-path="${path}">`
        : "";
    const chips = options.map((option: { name: string; color: string; desc?: string }) => {
        const selected = selectedValues.some((s: IAVCellSelectValue) => s.content === option.name);
        return `<span class="b3-chip b3-chip--middle${selected ? " b3-chip--primary" : ""} av__select-option" data-name="${escapeAttr(option.name)}" data-color="${option.color}" data-type="selectOption" data-path="${path}" style="background-color:var(--b3-font-background${option.color});color:var(--b3-font-color${option.color})">
<svg class="icon"><use xlink:href="#${selected ? "iconCheck" : "iconUncheck"}"></use></svg>
<span class="fn__ellipsis">${escapeHtml(option.name)}</span>
</span>`;
    }).join("");
    const dropdown = `<div class="av__select-dropdown" data-type="selectDropdown" data-path="${path}" data-single="${isSingle ? "true" : "false"}" style="display:none;">
${searchInput}<div class="av__select-options" data-type="selectOptions" data-path="${path}">${chips}</div>
</div>`;
    return {trigger, dropdown};
};

const readInlineValue = (rowElement: HTMLElement, valueType: TAVCol, operator: string, filter: IAVFilter): { newValue: IAVCellValue, relativeDate: IAVRelativeDate, relativeDate2: IAVRelativeDate } => {
    let newValue: IAVCellValue = filter.value;
    let relativeDate: IAVRelativeDate = filter.relativeDate;
    let relativeDate2: IAVRelativeDate = filter.relativeDate2;

    if (operator === "Is empty" || operator === "Is not empty") {
        newValue = genEmptyCellValue(valueType);
        relativeDate = undefined;
        relativeDate2 = undefined;
    } else if (valueType === "checkbox") {
        const select = rowElement.querySelector('[data-type="filterValue"]') as HTMLSelectElement;
        const isChecked = select?.value !== "false";
        newValue = genCellValue("checkbox", {checked: isChecked});
    } else if (valueType === "relation") {
        const input = rowElement.querySelector('[data-type="filterValue"]') as HTMLInputElement;
        newValue = input?.value ? genCellValue("relation", input.value) : genEmptyCellValue("relation");
    } else if (["text", "url", "block", "email", "phone", "template", "mAsset", "number"].includes(valueType)) {
        const input = rowElement.querySelector('[data-type="filterValue"]') as HTMLInputElement;
        const val = input?.value || "";
        newValue = val ? genCellValue(valueType, val) : genEmptyCellValue(valueType);
    } else if (["date", "created", "updated"].includes(valueType)) {
        const dateTypeSel = rowElement.querySelector('[data-type="dateType"]') as HTMLSelectElement;
        const isRelative = dateTypeSel?.value === "custom";
        if (isRelative) {
            relativeDate = readRelativeDate(rowElement, "");
            if (operator === "Is between") {
                relativeDate2 = readRelativeDate(rowElement, "2");
            } else {
                relativeDate2 = undefined;
            }
            newValue = {type: valueType} as IAVCellValue;
        } else {
            const absDate1 = rowElement.querySelector('[data-type="absDate"]') as HTMLInputElement;
            const dateStr1 = absDate1?.value || "";
            const content1 = dateStr1 ? new Date(dateStr1 + " 00:00").getTime() : 0;
            const isNotEmpty = !!dateStr1;
            let content2 = 0;
            let isNotEmpty2 = false;
            let dateStr2 = "";
            if (operator === "Is between") {
                const absDate2 = rowElement.querySelector('[data-type="absDate2"]') as HTMLInputElement;
                dateStr2 = absDate2?.value || "";
                content2 = dateStr2 ? new Date(dateStr2 + " 00:00").getTime() : 0;
                isNotEmpty2 = !!dateStr2;
            }
            newValue = {
                type: valueType,
                [valueType]: {
                    content: content1,
                    isNotEmpty,
                    content2,
                    isNotEmpty2,
                    hasEndDate: operator === "Is between" && isNotEmpty2,
                    isNotTime: true,
                },
            } as IAVCellValue;
            relativeDate = undefined;
            relativeDate2 = undefined;
        }
    } else if (valueType === "select" || valueType === "mSelect") {
        const path = rowElement.dataset.path;
        const mSelect: IAVCellSelectValue[] = [];
        const dropdown = document.querySelector(`[data-type="selectDropdown"][data-path="${path}"]`);
        const searchRoot = dropdown || rowElement;
        searchRoot.querySelectorAll('[data-type="selectOption"]').forEach((chip: HTMLElement) => {
            const useEl = chip.querySelector("use");
            if (useEl && useEl.getAttribute("xlink:href") === "#iconCheck") {
                mSelect.push({content: chip.dataset.name, color: chip.dataset.color});
            }
        });
        newValue = mSelect.length > 0 ? genCellValue(valueType, mSelect) : genEmptyCellValue(valueType);
    }

    if (filter.value?.type === "rollup") {
        newValue = {type: "rollup", rollup: {contents: [newValue]}} as IAVCellValue;
    }

    return {newValue, relativeDate, relativeDate2};
};

const readRelativeDate = (rowElement: HTMLElement, suffix: string): IAVRelativeDate => {
    const dirSel = rowElement.querySelector(`[data-type="dataDirection${suffix}"]`) as HTMLSelectElement;
    const countInput = rowElement.querySelector(`[data-type="relCount${suffix}"]`) as HTMLInputElement;
    const unitSel = rowElement.querySelector(`[data-type="relUnit${suffix}"]`) as HTMLSelectElement;
    const direction = parseInt(dirSel?.value || "0", 10);
    return {
        count: parseInt(countInput?.value || "1", 10),
        unit: parseInt(unitSel?.value || "0", 10) as 0 | 1 | 2 | 3,
        direction: direction as -1 | 0 | 1,
    };
};

export const commitFilter = (data: IAV, path: string, newFilter: IAVFilter, protyle: IProtyle, blockID: string, avID: string, menuElement: HTMLElement, reRender: boolean) => {
    const editable = getEditableFilters(data);
    const {parent, index} = getParentByPath(editable, path);
    if (!parent || index < 0 || index >= parent.length) {
        return;
    }
    const oldFilters = JSON.parse(JSON.stringify(data.view.filters));
    parent[index] = newFilter;

    transaction(protyle, [{
        action: "setAttrViewFilters",
        avID,
        data: JSON.parse(JSON.stringify(data.view.filters)),
        blockID
    }], [{
        action: "setAttrViewFilters",
        avID,
        data: oldFilters,
        blockID
    }]);

    if (reRender && menuElement) {
        menuElement.innerHTML = getFiltersHTML(data);
    }
};

export const bindInlineFilterEvents = (panelElement: HTMLElement, data: IAV, protyle: IProtyle, blockID: string, avID: string) => {
    if (panelElement.dataset.filterEventsBound === "true") {
        return;
    }
    panelElement.dataset.filterEventsBound = "true";
    const menuElement = panelElement.querySelector(".b3-menu") as HTMLElement;
    const fields = getFieldsByData(data);

    const getRow = (target: HTMLElement): HTMLElement => {
        const path = target.dataset.path;
        if (!path) return null;
        return menuElement.querySelector(`[data-path="${path}"]`) as HTMLElement;
    };

    const findColData = (path: string): IAVColumn => {
        const filter = getFilterByPath(getEditableFilters(data), path);
        if (!filter) return null;
        let colData: IAVColumn;
        fields.find((column: IAVColumn) => {
            if (column.id === filter.column) {
                colData = column;
                return true;
            }
        });
        return colData;
    };

    const saveRow = (rowElement: HTMLElement, path: string, reRender: boolean) => {
        const filter = getFilterByPath(getEditableFilters(data), path);
        const colData = findColData(path);
        if (!filter || !colData) return;
        const {type: valueType} = resolveFilterValueType(filter, colData);
        const operatorSel = rowElement.querySelector('[data-type="operation"]') as HTMLSelectElement;
        const operator = (operatorSel?.value || filter.operator) as TAVFilterOperator;
        const {newValue, relativeDate, relativeDate2} = readInlineValue(rowElement, valueType, operator, filter);
        const quantifierSel = rowElement.querySelector('[data-type="quantifier"]') as HTMLSelectElement;
        const newFilter: IAVFilter = {
            column: filter.column,
            operator,
            value: newValue,
            relativeDate,
            relativeDate2,
        };
        if (quantifierSel) {
            newFilter.quantifier = quantifierSel.value;
        }
        commitFilter(data, path, newFilter, protyle, blockID, avID, menuElement, reRender);
    };

    panelElement.addEventListener("change", (event: Event) => {
        const target = event.target as HTMLElement;
        const type = target.dataset.type;
        if (!type) return;
        const path = target.dataset.path;
        if (!path) return;
        const row = getRow(target);
        if (!row) return;

        if (type === "fieldSelect") {
            const newColId = (target as HTMLSelectElement).value;
            const newColData = fields.find((f: IAVColumn) => f.id === newColId);
            if (newColData) {
                const {operator, value} = genEmptyFilterValue(newColData);
                const newFilter: IAVFilter = {
                    column: newColId,
                    operator,
                    value,
                };
                commitFilter(data, path, newFilter, protyle, blockID, avID, menuElement, true);
            }
        } else if (type === "operation") {
            const filter = getFilterByPath(getEditableFilters(data), path);
            const colData = findColData(path);
            const {type: valueType} = resolveFilterValueType(filter, colData);
            const newOp = (target as HTMLSelectElement).value;
            const oldOp = filter.operator;
            const structureChange = (["date", "created", "updated"].includes(valueType) &&
                ((newOp === "Is between") !== (oldOp === "Is between"))) ||
                ((newOp === "Is empty" || newOp === "Is not empty") !== (oldOp === "Is empty" || oldOp === "Is not empty"));
            saveRow(row, path, structureChange);
        } else if (type === "quantifier" || type?.startsWith("dataDirection") || type?.startsWith("dateType")) {
            if (type === "dateType" || type === "dateType2" || type?.startsWith("dataDirection")) {
                saveRow(row, path, true);
            } else {
                saveRow(row, path, false);
            }
        } else if (type === "relUnit" || type === "relUnit2") {
            saveRow(row, path, false);
        } else if (type === "filterValue") {
            saveRow(row, path, false);
        }
    });

    panelElement.addEventListener("blur", (event: Event) => {
        const target = event.target as HTMLElement;
        if (target.dataset.type === "filterValue" || target.dataset.type?.startsWith("absDate") || target.dataset.type?.startsWith("relCount")) {
            const path = target.dataset.path;
            const row = getRow(target);
            if (path && row) saveRow(row, path, false);
        }
    }, true);

    panelElement.addEventListener("keydown", (event: KeyboardEvent) => {
        const target = event.target as HTMLElement;
        if (event.key !== "Enter" || event.isComposing) return;
        if (target.dataset.type === "filterValue") {
            const path = target.dataset.path;
            const row = getRow(target);
            if (path && row) {
                saveRow(row, path, false);
                event.preventDefault();
            }
        }
    });

    panelElement.addEventListener("click", (event: MouseEvent) => {
        const target = event.target as HTMLElement;
        const trigger = target.closest('[data-type="selectTrigger"]') as HTMLElement;
        if (trigger) {
            const path = trigger.dataset.path;
            const dropdown = menuElement.querySelector(`[data-type="selectDropdown"][data-path="${path}"]`) as HTMLElement;
            if (dropdown) {
                menuElement.querySelectorAll('[data-type="selectDropdown"]').forEach((el: HTMLElement) => {
                    if (el !== dropdown) el.style.display = "none";
                });
                if (dropdown.style.display === "none") {
                    const rect = trigger.getBoundingClientRect();
                    dropdown.style.zIndex = (++window.scribli.zIndex).toString();
                    dropdown.style.left = rect.left + "px";
                    dropdown.style.width = Math.max(rect.width, 120) + "px";
                    dropdown.style.visibility = "hidden";
                    dropdown.style.display = "block";
                    const dropdownHeight = dropdown.offsetHeight;
                    dropdown.style.visibility = "";
                    const spaceBelow = window.innerHeight - rect.bottom;
                    if (spaceBelow < dropdownHeight + 8 && rect.top > dropdownHeight + 8) {
                        dropdown.style.top = (rect.top - dropdownHeight - 4) + "px";
                    } else {
                        dropdown.style.top = (rect.bottom + 4) + "px";
                    }
                } else {
                    dropdown.style.display = "none";
                }
            }
            event.stopImmediatePropagation();
            return;
        }
        const chip = target.closest('[data-type="selectOption"]') as HTMLElement;
        if (!chip) return;
        const path = chip.dataset.path;
        const row = getRow(chip);
        if (!path || !row) return;
        const dropdown = menuElement.querySelector(`[data-type="selectDropdown"][data-path="${path}"]`) as HTMLElement;
        const isSingle = dropdown?.dataset.single === "true";
        const useEl = chip.querySelector("use");
        const isCheck = useEl.getAttribute("xlink:href") === "#iconCheck";
        if (isSingle && !isCheck) {
            dropdown.querySelectorAll('[data-type="selectOption"]').forEach((c: HTMLElement) => {
                if (c !== chip) {
                    const u = c.querySelector("use");
                    if (u && u.getAttribute("xlink:href") === "#iconCheck") {
                        u.setAttribute("xlink:href", "#iconUncheck");
                        c.classList.remove("b3-chip--primary");
                    }
                }
            });
        }
        useEl.setAttribute("xlink:href", isCheck ? "#iconUncheck" : "#iconCheck");
        chip.classList.toggle("b3-chip--primary", !isCheck);
        const triggerEl = menuElement.querySelector(`[data-type="selectTrigger"][data-path="${path}"]`) as HTMLElement;
        if (triggerEl && dropdown) {
            const isSingleSel = dropdown.dataset.single === "true";
            const placeholderStr = isSingleSel ? window.scribli.languages.select : window.scribli.languages.multiSelect;
            const selectedChips: string[] = [];
            dropdown.querySelectorAll('[data-type="selectOption"]').forEach((c: HTMLElement) => {
                const u = c.querySelector("use");
                if (u && u.getAttribute("xlink:href") === "#iconCheck") {
                    const name = c.dataset.name;
                    const color = c.dataset.color;
                    selectedChips.push(`<span class="b3-chip b3-chip--middle av__select-chip" style="background-color:var(--b3-font-background${color});color:var(--b3-font-color${color})">${escapeHtml(name)}</span>`);
                }
            });
            const contentHTML = selectedChips.join("") || `<span class="ft__on-surface fn__ellipsis">${placeholderStr}</span>`;
            triggerEl.innerHTML = contentHTML;
        }
        saveRow(row, path, false);
        event.stopImmediatePropagation();
    });

    panelElement.addEventListener("click", (event: MouseEvent) => {
        const target = event.target as HTMLElement;
        if (!target.closest('[data-type="selectTrigger"]') && !target.closest('[data-type="selectDropdown"]')) {
            menuElement.querySelectorAll('[data-type="selectDropdown"]').forEach((el: HTMLElement) => {
                el.style.display = "none";
            });
        }
        if (!target.closest('[data-type-rel="relation"]') && !target.closest('[data-type="relList"]')) {
            menuElement.querySelectorAll('[data-type="relList"]').forEach((el: HTMLElement) => {
                el.style.display = "none";
            });
        }
    }, true);

    panelElement.addEventListener("input", (event: InputEvent) => {
        const target = event.target as HTMLElement;
        if (target.dataset.type === "filterSearch") {
            const path = target.dataset.path;
            const dropdown = menuElement.querySelector(`[data-type="selectDropdown"][data-path="${path}"]`);
            if (!dropdown) return;
            const key = (target as HTMLInputElement).value.toLowerCase();
            dropdown.querySelectorAll('[data-type="selectOption"]').forEach((chip: HTMLElement) => {
                const name = (chip.dataset.name || "").toLowerCase();
                chip.style.display = (!key || name.indexOf(key) > -1 || key.indexOf(name) > -1) ? "" : "none";
            });
        } else if (target.dataset.type === "filterValue" && target.dataset.typeRel === "relation") {
            const path = target.dataset.path;
            const filter = getFilterByPath(getEditableFilters(data), path);
            const sourceColumn = findColData(path);
            const colData = filter && sourceColumn ? resolveFilterValueType(filter, sourceColumn).colData : sourceColumn;
            if (!colData?.relation?.avID) return;
            const keyword = (target as HTMLInputElement).value;
            fetchPost("/api/av/getAttributeViewPrimaryKeyValues", {
                id: colData.relation.avID,
                keyword,
            }, response => {
                if ((target as HTMLInputElement).value !== keyword) {
                    return;
                }
                const row = getRow(target);
                if (!row) return;
                let listEl = menuElement.querySelector(`[data-type="relList"][data-path="${path}"]`) as HTMLElement;
                if (!listEl) {
                    listEl = document.createElement("div");
                    listEl.setAttribute("data-type", "relList");
                    listEl.setAttribute("data-path", path);
                    listEl.className = "av__select-dropdown b3-list b3-list--background";
                    menuElement.appendChild(listEl);
                }
                let html = "";
                (response.data.rows.values as IAVCellValue[] || []).forEach((item, index) => {
                    const content = item.block?.content || window.scribli.languages.untitled;
                    html += `<div class="b3-list-item${index === 0 ? " b3-list-item--focus" : ""}" data-path="${path}" data-name="${escapeAttr(content)}">${escapeHtml(content)}</div>`;
                });
                listEl.innerHTML = html;
                if (!html) {
                    listEl.style.display = "none";
                    return;
                }
                const rect = target.getBoundingClientRect();
                listEl.style.zIndex = (++window.scribli.zIndex).toString();
                listEl.style.left = rect.left + "px";
                listEl.style.width = rect.width + "px";
                listEl.style.visibility = "hidden";
                listEl.style.display = "block";
                const listHeight = listEl.offsetHeight;
                listEl.style.visibility = "";
                listEl.style.top = window.innerHeight - rect.bottom < listHeight + 8 && rect.top > listHeight + 8
                    ? rect.top - listHeight - 4 + "px"
                    : rect.bottom + 4 + "px";
            });
        }
    });

    panelElement.addEventListener("click", (event: MouseEvent) => {
        const target = event.target as HTMLElement;
        const item = target.closest('[data-type="relList"] .b3-list-item') as HTMLElement;
        if (!item) return;
        const listEl = item.closest('[data-type="relList"]') as HTMLElement;
        const path = listEl.dataset.path;
        const row = menuElement.querySelector(`.av__filter-row[data-path="${path}"]`) as HTMLElement;
        if (!path || !row) return;
        const input = row.querySelector('[data-type="filterValue"]') as HTMLInputElement;
        if (input) {
            input.value = item.dataset.name || "";
        }
        listEl.style.display = "none";
        saveRow(row, path, false);
    }, true);
};
