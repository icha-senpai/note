import {SettingTabSearchResult} from "../setting/builder";
import {getTabGroupEntries} from "../setting/item";

export const scanSettingTabSearch = (
    tabId: string,
    tabSearchTitle: string,
    keywords: string,
): SettingTabSearchResult => {
    const visibleItemIds = new Set<string>();
    const visibleGroupIds = new Set<string>();

    if (tabSearchTitle.length > 0 && tabSearchTitle.includes(keywords)) {
        for (const {group, items} of getTabGroupEntries(tabId)) {
            visibleGroupIds.add(group.id);
            for (const item of items) {
                visibleItemIds.add(item.id);
            }
        }
        return {matches: true, visibleItemIds, visibleGroupIds};
    }

    let matches = false;
    for (const {group, items} of getTabGroupEntries(tabId)) {
        if (group.searchTitle.length > 0 && group.searchTitle.includes(keywords)) {
            matches = true;
            visibleGroupIds.add(group.id);
            for (const item of items) {
                visibleItemIds.add(item.id);
            }
            continue;
        }
        for (const item of items) {
            if (item.searchIndex.some((s) => s.includes(keywords))) {
                matches = true;
                visibleItemIds.add(item.id);
                visibleGroupIds.add(group.id);
            }
        }
    }
    return {matches, visibleItemIds, visibleGroupIds};
};
