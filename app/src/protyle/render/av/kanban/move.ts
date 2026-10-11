import {fetchSyncPost} from "../../../../util/fetch";
import {Menu} from "../../../../plugin/Menu";
import {Constants} from "../../../../constants";
import {getSavedDatabase} from "../undo";
import {transaction} from "../../../wysiwyg/transaction";
import {showMessage} from "../../../../dialog/message";
import * as dayjs from "dayjs";

export const openMoveToColumnMenu = async (protyle: IProtyle, block: HTMLElement, cards: HTMLElement[], position: IPosition) => {
    const avID = block.dataset.avId;
    const response = await fetchSyncPost("/api/av/renderAttributeView", {
        id: avID, blockID: block.dataset.nodeId,
        viewID: block.getAttribute(Constants.CUSTOM_SY_AV_VIEW), ignoreRows: true,
    });
    if (response.code !== 0 || protyle.disabled) { return; }
    const data = response.data as IAV;
    const keyID = data.view.group?.field;
    const saved = await getSavedDatabase(avID);
    const column = saved.keyValues.find(item => item.key.id === keyID);
    if (!column || !["select", "mSelect"].includes(column.key.type)) { return; }
    const rows = cards.map(card => ({
        id: card.dataset.id,
        group: data.view.groups.find(group => group.id === card.closest<HTMLElement>(".av__body")?.dataset.groupId)?.groupValue.mSelect?.[0]?.content,
    }));
    const move = async (name?: string) => {
        if (protyle.disabled) { return; }
        const current = await getSavedDatabase(avID);
        const values = current.keyValues.find(item => item.key.id === keyID);
        if (!values || protyle.disabled) { return; }
        const option = name ? values.key.options.find(item => item.name === name) : null;
        if (name && !option) { return; }
        const doOperations: IOperation[] = [];
        const undoOperations: IOperation[] = [];
        const visited = new Set<string>();
        rows.forEach(row => {
            if (visited.has(row.id)) { return; }
            visited.add(row.id);
            const value = values.values?.find(item => item.blockID === row.id);
            const oldValue: IAVCellValue = value || {id: Lute.NewNodeID(), keyID, blockID: row.id, type: values.key.type, mSelect: []};
            const selected = values.key.type === "mSelect" && option ? (oldValue.mSelect || []).filter(item => item.content !== row.group && item.content !== name) : [];
            const nextValue = {...oldValue, mSelect: option ? [...selected, {content: option.name, color: option.color}] : []};
            doOperations.push({action: "updateAttrViewCell", avID, id: oldValue.id, keyID, rowID: row.id, data: nextValue});
            undoOperations.push({action: "updateAttrViewCell", avID, id: oldValue.id, keyID, rowID: row.id, data: oldValue});
        });
        if (doOperations.length === 0) { return; }
        const updated = dayjs().format("YYYYMMDDHHmmss");
        transaction(protyle, [...doOperations, {action: "doUpdateUpdated", id: block.dataset.nodeId, data: updated}], [...undoOperations, {action: "doUpdateUpdated", id: block.dataset.nodeId, data: block.getAttribute("updated") || updated}]);
    };
    const menu = new Menu("av-kanban-move");
    (column.key.options || []).forEach(option => {
        menu.addItem({label: option.name, icon: "iconRight", click() { void move(option.name).catch(error => showMessage(error.message)); }});
    });
    menu.addSeparator();
    menu.addItem({label: window.scribli.languages.kanbanUncategorized, icon: "iconRemove", click() { void move().catch(error => showMessage(error.message)); }});
    menu.open({x: position.x, y: position.y});
};
