import type {SettingTabBuilder} from "../setting/builder";

const registerSearchQueryGroup = (tab: SettingTabBuilder) => {
    const group = tab.group("query", "");

    group.switchQuery({
        key: "blockType",
        title: window.scribli.languages.searchBlockType,
        footer: `[1] ${window.scribli.languages.containerBlockTip1}`,
        items: [
            {kind: "switch", id: "search.mathBlock", label: window.scribli.languages.math, icon: "iconMath"},
            {kind: "switch", id: "search.table", label: window.scribli.languages.table, icon: "iconTable"},
            {kind: "switch", id: "search.paragraph", label: window.scribli.languages.paragraph, icon: "iconParagraph"},
            {kind: "switch", id: "search.heading", label: window.scribli.languages.headings, icon: "iconHeadings"},
            {kind: "switch", id: "search.codeBlock", label: window.scribli.languages.code, icon: "iconCode"},
            {kind: "switch", id: "search.htmlBlock", label: "HTML", icon: "iconHTML5"},
            {kind: "switch", id: "search.databaseBlock", label: window.scribli.languages.database, icon: "iconDatabase"},
            {kind: "switch", id: "search.embedBlock", label: window.scribli.languages.embedBlock, icon: "iconSQL"},
            {kind: "switch", id: "search.videoBlock", label: window.scribli.languages.video, icon: "iconVideo"},
            {kind: "switch", id: "search.audioBlock", label: window.scribli.languages.audio, icon: "iconRecord"},
            {kind: "switch", id: "search.iframeBlock", label: "IFrame", icon: "iconGlobe"},
            {kind: "switch", id: "search.widgetBlock", label: window.scribli.languages.widget, icon: "iconBoth"},
            {kind: "switch", id: "search.blockquote", label: `${window.scribli.languages.quote} <sup>[1]</sup>`, icon: "iconQuote"},
            {kind: "switch", id: "search.callout", label: `${window.scribli.languages.callout} <sup>[1]</sup>`, icon: "iconCallout"},
            {kind: "switch", id: "search.superBlock", label: `${window.scribli.languages.superBlock} <sup>[1]</sup>`, icon: "iconSuper"},
            {kind: "switch", id: "search.list", label: `${window.scribli.languages.list1} <sup>[1]</sup>`, icon: "iconList"},
            {kind: "switch", id: "search.listItem", label: `${window.scribli.languages.listItem} <sup>[1]</sup>`, icon: "iconListItem"},
            {kind: "switch", id: "search.document", label: window.scribli.languages.doc, icon: "iconFile"},
        ],
    });
    group.switchQuery({
        key: "blockAttr",
        title: window.scribli.languages.searchBlockAttr,
        items: [
            {kind: "switch", id: "search.name", label: window.scribli.languages.name, icon: "iconN"},
            {kind: "switch", id: "search.alias", label: window.scribli.languages.alias, icon: "iconA"},
            {kind: "switch", id: "search.memo", label: window.scribli.languages.memo, icon: "iconM"},
            {kind: "switch", id: "search.ial", label: window.scribli.languages.allAttrs},
        ],
    });
    group.switchQuery({
        key: "backmention",
        title: window.scribli.languages.searchBackmention,
        items: [
            {kind: "switch", id: "search.backlinkMentionName", label: window.scribli.languages.name},
            {kind: "switch", id: "search.backlinkMentionAlias", label: window.scribli.languages.alias},
            {kind: "switch", id: "search.backlinkMentionAnchor", label: window.scribli.languages.anchor},
            {kind: "switch", id: "search.backlinkMentionDoc", label: window.scribli.languages.docName},
            {kind: "number", id: "search.backlinkMentionKeywordsLimit", label: window.scribli.languages.keywordsLimit, min: 1, max: 10240},
        ],
    });
    group.switchQuery({
        key: "virtualRef",
        title: window.scribli.languages.searchVirtualRef,
        items: [
            {kind: "switch", id: "search.virtualRefName", label: window.scribli.languages.name},
            {kind: "switch", id: "search.virtualRefAlias", label: window.scribli.languages.alias},
            {kind: "switch", id: "search.virtualRefAnchor", label: window.scribli.languages.anchor},
            {kind: "switch", id: "search.virtualRefDoc", label: window.scribli.languages.docName},
        ],
    });
    group.switchQuery({
        key: "index",
        title: window.scribli.languages.searchIndex,
        items: [
            {kind: "switch", id: "search.indexAssetPath", label: window.scribli.languages.indexAssetPath},
        ],
    });
};

const registerSearchLimitsGroup = (tab: SettingTabBuilder) => {
    const group = tab.group("limits", "");

    group.number("search.limit", {
        title: window.scribli.languages.searchLimit,
        desc: `${window.scribli.languages.searchLimit1}<br>${window.scribli.languages.searchLimit2}`,
        min: 32,
        max: 10240,
    });
    group.switch("search.caseSensitive", {
        title: window.scribli.languages.searchCaseSensitive,
        desc: window.scribli.languages.searchCaseSensitive1,
    });
};

export const registerSearchTab = (tab: SettingTabBuilder) => {
    registerSearchQueryGroup(tab);
    registerSearchLimitsGroup(tab);
};
