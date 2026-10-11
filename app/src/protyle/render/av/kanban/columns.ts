import {Menu} from "../../../../plugin/Menu";
import {confirmDialog} from "../../../../dialog/confirmDialog";
import {showMessage} from "../../../../dialog/message";
import {escapeAttr, escapeHtml} from "../../../../util/escape";
import {fetchSyncPost} from "../../../../util/fetch";
import {Constants} from "../../../../constants";
import {transaction} from "../../../wysiwyg/transaction";
import * as dayjs from "dayjs";
import {deleteDatabaseOption} from "../undo";

type ColumnOption = NonNullable<IAVColumn["options"]>[number];

interface ColumnContext {
    data: IAV;
    key: IAVColumn;
    values: IAVCellValue[];
}

const changeColumns = (protyle: IProtyle, blockElement: HTMLElement, doOperations: IOperation[], undoOperations: IOperation[]) => {
    const updated = dayjs().format("YYYYMMDDHHmmss");
    transaction(protyle, [...doOperations, {
        action: "doUpdateUpdated", id: blockElement.dataset.nodeId, data: updated,
    }], [...undoOperations, {
        action: "doUpdateUpdated", id: blockElement.dataset.nodeId, data: blockElement.getAttribute("updated") || updated,
    }]);
};

const getColumnContext = async (blockElement: HTMLElement, includeValues = false): Promise<ColumnContext> => {
    const id = blockElement.dataset.avId;
    const responses = await Promise.all([
        fetchSyncPost("/api/av/renderAttributeView", {
            id,
            blockID: blockElement.dataset.nodeId,
            viewID: blockElement.getAttribute(Constants.CUSTOM_SY_AV_VIEW),
            ignoreRows: true,
        }),
        includeValues ? fetchSyncPost("/api/av/getAttributeView", {id}) :
            fetchSyncPost("/api/av/getAttributeViewKeysByAvID", {avID: id}),
    ]);
    if (responses.some(response => response.code !== 0)) {
        throw new Error(responses.find(response => response.code !== 0).msg);
    }
    const data = responses[0].data as IAV;
    const keyValues: Array<{key: IAVColumn, values?: IAVCellValue[]}> = includeValues ? responses[1].data.av.keyValues :
        (responses[1].data as IAVColumn[]).map(key => ({key}));
    const column = keyValues.find(item => item.key.id === data.view.group?.field);
    if (!column || !["select", "mSelect"].includes(column.key.type)) {
        throw new Error(window.scribli.languages.kanbanColumnUnavailable);
    }
    return {data, key: column.key, values: column.values || []};
};

const optionOperation = (context: ColumnContext, option: ColumnOption, name: string, color = option.color): IOperation => ({
    action: "updateAttrViewColOption",
    avID: context.data.id,
    id: context.key.id,
    data: {oldName: option.name, newName: name, newColor: color, newDesc: option.desc || ""},
});

const addOptions = (protyle: IProtyle, blockElement: HTMLElement, context: ColumnContext, added: ColumnOption[]) => {
    const existing = context.key.options || [];
    const doOperations: IOperation[] = [{
        action: "updateAttrViewColOptions",
        avID: context.data.id,
        id: context.key.id,
        data: [...existing, ...added],
    }, {
        action: "setAttrViewGroup",
        avID: context.data.id,
        blockID: blockElement.dataset.nodeId,
        data: {...context.data.view.group, hideEmpty: false},
    }];
    const undoOperations: IOperation[] = added.map(option => ({
        action: "removeAttrViewColOption",
        avID: context.data.id,
        id: context.key.id,
        data: option.name,
    }));
    undoOperations.push({
        action: "setAttrViewGroup",
        avID: context.data.id,
        blockID: blockElement.dataset.nodeId,
        data: context.data.view.group,
    });
    changeColumns(protyle, blockElement, doOperations, undoOperations);
};

const deleteColumn = async (protyle: IProtyle, blockElement: HTMLElement, keyID: string, name: string) => {
    await deleteDatabaseOption(protyle, blockElement.dataset.avId, blockElement.dataset.nodeId, keyID, name);
};

const editColumnName = (protyle: IProtyle, blockElement: HTMLElement, button: HTMLElement, oldName?: string) => {
    if (!button.isConnected) {
        return;
    }
    const input = document.createElement("input");
    input.className = "b3-text-field av__kanban-column-input";
    input.setAttribute("aria-label", window.scribli.languages.kanbanColumnName);
    input.placeholder = window.scribli.languages.kanbanColumnName;
    input.value = oldName || "";
    button.replaceWith(input);
    input.focus();
    input.select();
    let saving = false;
    let cancelled = false;
    const restore = () => {
        cancelled = true;
        input.replaceWith(button);
    };
    const save = async () => {
        if (saving || cancelled) {
            return;
        }
        const name = input.value.trim();
        if (!name || name === oldName) {
            restore();
            return;
        }
        saving = true;
        try {
            const context = await getColumnContext(blockElement);
            if (cancelled || protyle.disabled || !input.isConnected) {
                restore();
                return;
            }
            if (/^_@.*@_$/.test(name) || context.key.options?.some(option => option.name === name)) {
                showMessage(window.scribli.languages.kanbanColumnNameInvalid);
                input.focus();
                return;
            }
            if (oldName !== undefined) {
                const option = context.key.options?.find(item => item.name === oldName);
                if (!option) {
                    restore();
                    return;
                }
                restore();
                changeColumns(protyle, blockElement, [optionOperation(context, option, name)], [
                    optionOperation(context, {...option, name}, oldName),
                ]);
            } else {
                restore();
                addOptions(protyle, blockElement, context, [{name, color: String((context.key.options?.length || 0) % 14 + 1)}]);
            }
        } catch (error) {
            showMessage(error.message);
            input.focus();
        } finally {
            saving = false;
        }
    };
    input.addEventListener("blur", () => { void save(); });
    input.addEventListener("click", event => event.stopPropagation());
    input.addEventListener("keydown", event => {
        event.stopPropagation();
        if (event.isComposing) {
            return;
        }
        if (event.key === "Escape") {
            event.preventDefault();
            restore();
            button.focus();
        } else if (event.key === "Enter") {
            event.preventDefault();
            void save();
        }
    });
};

const openColumnMenu = async (protyle: IProtyle, blockElement: HTMLElement, button: HTMLElement) => {
    const context = await getColumnContext(blockElement);
    if (protyle.disabled || !button.isConnected) {
        return;
    }
    const name = button.dataset.optionName;
    const option = context.key.options?.find(item => item.name === name);
    if (!option) {
        return;
    }
    const title = button.parentElement.querySelector<HTMLElement>('[data-type="av-kanban-rename-column"]');
    const menu = new Menu("av-kanban-column");
    menu.addItem({
        icon: "iconEdit",
        label: window.scribli.languages.rename,
        click() { editColumnName(protyle, blockElement, title, name); },
    });
    menu.addItem({
        icon: "iconPaintBucket",
        label: window.scribli.languages.color,
        submenu: Array.from({length: 14}, (_, index) => ({
            label: window.scribli.languages.kanbanColumnColor.replace("${x}", String(index + 1)),
            iconHTML: `<span class="color__square" style="color:var(--b3-font-color${index + 1});background-color:var(--b3-font-background${index + 1})">A</span>`,
            checked: option.color === String(index + 1),
            click() {
                changeColumns(protyle, blockElement, [optionOperation(context, option, name, String(index + 1))], [optionOperation(context, option, name)]);
            },
        })),
    });
    const groups = context.data.view.groups;
    const index = groups.findIndex(group => group.id === button.dataset.groupId);
    const visibleGroups = groups.filter(group => group.groupHidden !== 2 && ["select", "mSelect"].includes(group.groupValue.type));
    const visibleIndex = visibleGroups.findIndex(group => group.id === button.dataset.groupId);
    const moveColumn = (offset: number) => {
        const neighbour = visibleGroups[visibleIndex + offset];
        const neighbourIndex = groups.findIndex(group => group.id === neighbour.id);
        const previousID = offset > 0 ? neighbour.id : (groups[neighbourIndex - 1]?.id || "");
        changeColumns(protyle, blockElement, [{
            action: "sortAttrViewGroup",
            avID: context.data.id,
            blockID: blockElement.dataset.nodeId,
            id: button.dataset.groupId,
            previousID,
        }], [{
            action: "sortAttrViewGroup",
            avID: context.data.id,
            blockID: blockElement.dataset.nodeId,
            id: button.dataset.groupId,
            previousID: groups[index - 1]?.id || "",
        }, {
            action: "setAttrViewGroup",
            avID: context.data.id,
            blockID: blockElement.dataset.nodeId,
            data: context.data.view.group,
        }]);
    };
    menu.addItem({icon: "iconLeft", label: window.scribli.languages.kanbanMoveLeft, disabled: visibleIndex <= 0, click() { moveColumn(-1); }});
    menu.addItem({icon: "iconRight", label: window.scribli.languages.kanbanMoveRight, disabled: visibleIndex < 0 || visibleIndex >= visibleGroups.length - 1, click() { moveColumn(1); }});
    menu.addSeparator();
    menu.addItem({
        icon: "iconTrashcan",
        label: window.scribli.languages.kanbanDeleteColumn,
        warning: true,
        click() {
            confirmDialog(window.scribli.languages.kanbanDeleteColumn,
                window.scribli.languages.kanbanDeleteColumnConfirm.replace("${x}", escapeHtml(name)), () => {
                    void deleteColumn(protyle, blockElement, context.key.id, name).catch(error => showMessage(error.message));
                }, undefined, true);
        },
    });
    const rect = button.getBoundingClientRect();
    menu.open({x: rect.left, y: rect.bottom, h: rect.height});
};

export const getKanbanColumnActionsHTML = (data: IAV, editable: boolean) => {
    const key = (data.view as IAVKanban).fields.find(field => field.id === data.view.group?.field);
    if (!editable || !key || !["select", "mSelect"].includes(key.type)) {
        return "";
    }
    const hasColumns = data.view.groups.some(group => ["select", "mSelect"].includes(group.groupValue.type));
    return `<div class="av__kanban-column-actions">
    <button type="button" class="b3-button b3-button--outline" data-type="av-kanban-add-column"><svg><use xlink:href="#iconAdd"></use></svg>${window.scribli.languages.kanbanAddColumn}</button>
    ${hasColumns ? "" : `<button type="button" class="b3-button b3-button--outline" data-type="av-kanban-starter">${window.scribli.languages.kanbanStarter}</button>`}
</div>`;
};

export const getKanbanColumnTitleHTML = (group: IAVView, nameHTML: string, editable: boolean) => {
    const option = group.groupValue.mSelect?.[0];
    if (!editable || !["select", "mSelect"].includes(group.groupValue.type) || !option) {
        return {title: `<span class="av__group-name fn__ellipsis">${nameHTML}</span>`, menu: ""};
    }
    return {
        title: `<span class="av__group-name av__kanban-column-label fn__ellipsis">${nameHTML}</span><button type="button" class="av__group-name av__kanban-column-name fn__ellipsis" data-type="av-kanban-rename-column" data-option-name="${escapeAttr(escapeHtml(option.content))}" aria-label="${escapeAttr(escapeHtml(window.scribli.languages.kanbanRenameColumn.replace("${x}", option.content)))}">${nameHTML}</button>`,
        menu: `<button type="button" class="av__group-icon av__kanban-column-menu" data-type="av-kanban-column-menu" data-group-id="${escapeAttr(group.id)}" data-option-name="${escapeAttr(escapeHtml(option.content))}" aria-label="${escapeAttr(escapeHtml(window.scribli.languages.kanbanColumnMenu.replace("${x}", option.content)))}"><svg><use xlink:href="#iconMore"></use></svg></button>`,
    };
};

export const kanbanColumnAction = (protyle: IProtyle, blockElement: HTMLElement, target: HTMLElement, type: string) => {
    if (type === "av-kanban-add-column" || type === "av-kanban-rename-column") {
        editColumnName(protyle, blockElement, target, type === "av-kanban-rename-column" ? target.dataset.optionName : undefined);
    } else if (type === "av-kanban-column-menu") {
        void openColumnMenu(protyle, blockElement, target).catch(error => showMessage(error.message));
    } else if (type === "av-kanban-starter") {
        void getColumnContext(blockElement).then(context => {
            if (protyle.disabled || context.key.options?.length) {
                return;
            }
            addOptions(protyle, blockElement, context, [
                {name: window.scribli.languages.kanbanTodo, color: "1"},
                {name: window.scribli.languages.kanbanInProgress, color: "2"},
                {name: window.scribli.languages.kanbanDone, color: "3"},
            ]);
        }).catch(error => showMessage(error.message));
    }
};
