import {fetchSyncPost} from "../../../util/fetch";
import {transaction} from "../../wysiwyg/transaction";
import * as dayjs from "dayjs";

interface SavedDatabase {
    views: Array<Record<string, unknown> & {id: string, filters?: IAVFilter[]}>;
    keyValues: Array<{key: IAVColumn, values?: IAVCellValue[]}>;
    newItemTemplates?: IAVNewItemTemplate[];
}

export const getSavedDatabase = async (avID: string): Promise<SavedDatabase> => {
    const response = await fetchSyncPost("/api/av/getAttributeView", {id: avID});
    if (response.code !== 0 || !response.data?.av) {
        throw new Error(response.msg || window.scribli.languages.kanbanColumnUnavailable);
    }
    return response.data.av;
};

export const getViewBlockIDs = async (avID: string, viewID: string) => {
    const response = await fetchSyncPost("/api/av/getMirrorDatabaseBlocks", {avID});
    if (response.code !== 0) { throw new Error(response.msg); }
    const ids = (response.data.refDefs || []).map((item: {refID: string}) => item.refID) as string[];
    const attributes = await Promise.all(ids.map(id => fetchSyncPost("/api/attr/getBlockAttrs", {id})));
    if (attributes.some(item => item.code !== 0)) { throw new Error(attributes.find(item => item.code !== 0).msg); }
    return ids.filter((id, index) => attributes[index].data["custom-sy-av-view"] === viewID);
};

export const deleteDatabaseOption = async (protyle: IProtyle, avID: string, blockID: string, keyID: string, name: string) => {
    const saved = await getSavedDatabase(avID);
    const column = saved.keyValues.find(item => item.key.id === keyID);
    if (protyle.disabled || !column?.key.options?.some(option => option.name === name)) { return false; }
    const updated = dayjs().format("YYYYMMDDHHmmss");
    const block = protyle.wysiwyg.element.querySelector(`[data-node-id="${blockID}"]`);
    transaction(protyle, [{action: "removeAttrViewColOption", avID, id: keyID, data: name}, {
        action: "doUpdateUpdated", id: blockID, data: updated,
    }], [{
        action: "restoreAttrViewColOption", avID, blockID,
        data: {
            keyID,
            options: column.key.options,
            values: (column.values || []).filter(value => value.mSelect?.some(option => option.content === name)),
            filters: Object.fromEntries(saved.views.map(view => [view.id, view.filters || []])),
            templates: saved.newItemTemplates || [],
        },
    }, {action: "doUpdateUpdated", id: blockID, data: block?.getAttribute("updated") || updated}]);
    return true;
};
