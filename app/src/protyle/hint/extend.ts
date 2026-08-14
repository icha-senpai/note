import {fetchPost} from "../../util/fetch";
import {insertHTML} from "../util/insertHTML";
import {getIconByType} from "../../editor/getIcon";
import {updateHotkeyTip} from "../util/compatibility";
import {blockRender} from "../render/blockRender";
import {Constants} from "../../constants";
import {processRender} from "../util/processCode";
import {highlightRender} from "../render/highlightRender";
import {focusBlock, focusByRange, getEditorRange} from "../util/selection";
import {hasClosestBlock, hasClosestByClassName} from "../util/hasClosest";
import {getContenteditableElement, getTopAloneElement} from "../wysiwyg/getBlock";
import {replaceFileName} from "../../editor/rename";
import {transaction} from "../wysiwyg/transaction";
import {getAssetName, getDisplayName, isEncryptedBox, pathPosix} from "../../util/pathName";
import {genEmptyElement} from "../../block/util";
import {updateListOrder} from "../wysiwyg/list";
import {escapeHtml, stripSearchMark} from "../../util/escape";
import {zoomOut} from "../../menus/protyle";
import {hideElements} from "../ui/hideElements";
import {genAssetHTML} from "../../asset/renderAssets";
import {unicode2Emoji} from "../../emoji";
import {avRender} from "../render/av/render";

const getHotkeyOrMarker = (hotkey: string, marker: string) => {
    if (hotkey) {
        return `<span class="b3-menu__accelerator b3-menu__accelerator--hotkey">${updateHotkeyTip(hotkey)}</span>`;
    } else if (marker) {
        return `<span class="b3-list-item__meta">${marker}</span>`;
    }
    return "";
};

export const hintSlash = (key: string, protyle: IProtyle) => {
    const allList: IHintData[] = [{
        filter: [window.scribli.languages.template, "template", "moban", "muban", "mb"],
        id: "template",
        value: Constants.ZWSP,
        html: `<div class="b3-list-item__first"><svg class="b3-list-item__graphic"><use xlink:href="#iconMarkdown"></use></svg><span class="b3-list-item__text">${window.scribli.languages.template}</span></div>`,
    }, {
        filter: [window.scribli.languages.widget, "widget", "guajian", "gj"],
        id: "widget",
        value: Constants.ZWSP + 1,
        html: `<div class="b3-list-item__first"><svg class="b3-list-item__graphic"><use xlink:href="#iconBoth"></use></svg><span class="b3-list-item__text">${window.scribli.languages.widget}</span></div>`,
    }, {
        filter: [window.scribli.languages.assets, "assets", "ziyuan", "zy"],
        id: "assets",
        value: Constants.ZWSP + 2,
        html: `<div class="b3-list-item__first"><svg class="b3-list-item__graphic"><use xlink:href="#iconImage"></use></svg><span class="b3-list-item__text">${window.scribli.languages.assets}</span></div>`,
    }, {
        filter: [window.scribli.languages.ref, "block reference", "kuaiyinyong", "kyy"],
        id: "ref",
        value: "((",
        html: `<div class="b3-list-item__first"><svg class="b3-list-item__graphic"><use xlink:href="#iconRef"></use></svg><span class="b3-list-item__text">${window.scribli.languages.ref}</span><span class="b3-list-item__meta">((</span></div>`,
    }, {
        filter: [window.scribli.languages.blockEmbed, "embed block", "qianrukuai", "qrk"],
        id: "blockEmbed",
        value: "{{",
        html: `<div class="b3-list-item__first"><svg class="b3-list-item__graphic"><use xlink:href="#iconSQL"></use></svg><span class="b3-list-item__text">${window.scribli.languages.blockEmbed}</span><span class="b3-list-item__meta">{{</span></div>`,
    }, {
        filter: [window.scribli.languages.aiWriting, "ai writing", "aibianxie", "aibx", "rengongzhineng", "rgzn"],
        id: "aiWriting",
        value: Constants.ZWSP + 5,
        html: `<div class="b3-list-item__first"><svg class="b3-list-item__graphic"><use xlink:href="#iconSparkles"></use></svg><span class="b3-list-item__text">${window.scribli.languages.aiWriting}</span>${getHotkeyOrMarker(window.scribli.config.keymap.editor.general.aiWriting.custom, "")}</div>`,
    }, {
        filter: [window.scribli.languages.database, "database", "db", "shujuku", "sjk", "view"],
        id: "database",
        value: '<div data-type="NodeAttributeView" data-av-type="table"></div>',
        html: `<div class="b3-list-item__first"><svg class="b3-list-item__graphic"><use xlink:href="#iconDatabase"></use></svg><span class="b3-list-item__text">${window.scribli.languages.database}</span></div>`,
    }, {
        filter: [window.scribli.languages.newFileRef, "create new doc with reference", "xinjianwendangbingyinyong", "xjwdbyy"],
        id: "newFileRef",
        value: Constants.ZWSP + 4,
        html: `<div class="b3-list-item__first"><svg class="b3-list-item__graphic"><use xlink:href="#iconFile"></use></svg><span class="b3-list-item__text">${window.scribli.languages.newFileRef}</span></div>`,
    }, {
        filter: [window.scribli.languages.newSubDocRef, "create sub doc with reference", "xinjianziwendangbingyinyong", "xjzwdbyy"],
        id: "newSubDocRef",
        value: Constants.ZWSP + 6,
        html: `<div class="b3-list-item__first"><svg class="b3-list-item__graphic"><use xlink:href="#iconFile"></use></svg><span class="b3-list-item__text">${window.scribli.languages.newSubDocRef}</span></div>`,
    }, {
        value: "",
        id: "separator_1",
        html: "separator",
    }, {
        filter: [window.scribli.languages.heading1, "heading1", "h1", "yijibiaoti", "yjbt"],
        id: "heading1",
        value: "# " + Lute.Caret,
        html: `<div class="b3-list-item__first"><svg class="b3-list-item__graphic"><use xlink:href="#iconH1"></use></svg><span class="b3-list-item__text">${window.scribli.languages.heading1}</span>${getHotkeyOrMarker(window.scribli.config.keymap.editor.heading.heading1.custom, "# ")}</div>`,
    }, {
        filter: [window.scribli.languages.heading2, "heading2", "h2", "erjibiaoti", "ejbt"],
        id: "heading2",
        value: "## " + Lute.Caret,
        html: `<div class="b3-list-item__first"><svg class="b3-list-item__graphic"><use xlink:href="#iconH2"></use></svg><span class="b3-list-item__text">${window.scribli.languages.heading2}</span>${getHotkeyOrMarker(window.scribli.config.keymap.editor.heading.heading2.custom, "## ")}</div>`,
    }, {
        filter: [window.scribli.languages.heading3, "heading3", "h3", "sanjibiaoti", "sjbt"],
        id: "heading3",
        value: "### " + Lute.Caret,
        html: `<div class="b3-list-item__first"><svg class="b3-list-item__graphic"><use xlink:href="#iconH3"></use></svg><span class="b3-list-item__text">${window.scribli.languages.heading3}</span>${getHotkeyOrMarker(window.scribli.config.keymap.editor.heading.heading3.custom, "### ")}</div>`,
    }, {
        filter: [window.scribli.languages.heading4, "heading4", "h4", "sijibiaoti", "sjbt"],
        id: "heading4",
        value: "#### " + Lute.Caret,
        html: `<div class="b3-list-item__first"><svg class="b3-list-item__graphic"><use xlink:href="#iconH4"></use></svg><span class="b3-list-item__text">${window.scribli.languages.heading4}</span>${getHotkeyOrMarker(window.scribli.config.keymap.editor.heading.heading4.custom, "#### ")}</div>`,
    }, {
        filter: [window.scribli.languages.heading5, "heading5", "h5", "wujibiaoti", "wjbt"],
        id: "heading5",
        value: "##### " + Lute.Caret,
        html: `<div class="b3-list-item__first"><svg class="b3-list-item__graphic"><use xlink:href="#iconH5"></use></svg><span class="b3-list-item__text">${window.scribli.languages.heading5}</span>${getHotkeyOrMarker(window.scribli.config.keymap.editor.heading.heading5.custom, "##### ")}</div>`,
    }, {
        filter: [window.scribli.languages.heading6, "heading6", "h6", "liujibiaoti", "ljbt"],
        id: "heading6",
        value: "###### " + Lute.Caret,
        html: `<div class="b3-list-item__first"><svg class="b3-list-item__graphic"><use xlink:href="#iconH6"></use></svg><span class="b3-list-item__text">${window.scribli.languages.heading6}</span>${getHotkeyOrMarker(window.scribli.config.keymap.editor.heading.heading6.custom, "###### ")}</div>`,
    }, {
        filter: [window.scribli.languages.list, "unordered list", "wuxvliebiao", "wuxuliebiao", "wxlb"],
        id: "list",
        value: "- " + Lute.Caret,
        html: `<div class="b3-list-item__first"><svg class="b3-list-item__graphic"><use xlink:href="#iconList"></use></svg><span class="b3-list-item__text">${window.scribli.languages.list}</span>${getHotkeyOrMarker(window.scribli.config.keymap.editor.insert.list.custom, "- ")}</div>`,
    }, {
        filter: [window.scribli.languages["ordered-list"], "order list", "ordered list", "youxvliebiao", "youxuliebiao", "yxlb"],
        id: "orderedList",
        value: "1. " + Lute.Caret,
        html: `<div class="b3-list-item__first"><svg class="b3-list-item__graphic"><use xlink:href="#iconOrderedList"></use></svg><span class="b3-list-item__text">${window.scribli.languages["ordered-list"]}</span>${getHotkeyOrMarker(window.scribli.config.keymap.editor.insert["ordered-list"].custom, "1. ")}</div>`,
    }, {
        filter: [window.scribli.languages.check, "task list", "todo list", "renwuliebiao", "rwlb"],
        id: "check",
        value: "- [ ] " + Lute.Caret,
        html: `<div class="b3-list-item__first"><svg class="b3-list-item__graphic"><use xlink:href="#iconCheck"></use></svg><span class="b3-list-item__text">${window.scribli.languages.check}</span>${getHotkeyOrMarker(window.scribli.config.keymap.editor.insert.check.custom, "[]")}</div>`,
    }, {
        filter: [window.scribli.languages.quote, "blockquote", "bq", "yinshu", "ys"],
        id: "quote",
        value: "> " + Lute.Caret,
        html: `<div class="b3-list-item__first"><svg class="b3-list-item__graphic"><use xlink:href="#iconQuote"></use></svg><span class="b3-list-item__text">${window.scribli.languages.quote}</span>${getHotkeyOrMarker(window.scribli.config.keymap.editor.insert.quote.custom, ">")}</div>`,
    }, {
        filter: [window.scribli.languages.callout, "callout", "ts", "tishi", "note"],
        id: "calloutNote",
        value: `> [!NOTE]\n> ${Lute.Caret}`,
        html: `<div class="b3-list-item__first"><span class="b3-list-item__graphic">✏️</span><span class="b3-list-item__text">${window.scribli.languages.callout} - <span style="color: var(--b3-callout-note)">Note</span></span></div>`,
    }, {
        filter: [window.scribli.languages.callout, "callout", "ts", "tishi", "tip"],
        id: "calloutTip",
        value: `> [!TIP]\n> ${Lute.Caret}`,
        html: `<div class="b3-list-item__first"><span class="b3-list-item__graphic">💡</span><span class="b3-list-item__text">${window.scribli.languages.callout} - <span style="color: var(--b3-callout-tip)">Tip</span></span></div>`,
    }, {
        filter: [window.scribli.languages.callout, "callout", "ts", "tishi", "important"],
        id: "calloutImportant",
        value: `> [!IMPORTANT]\n> ${Lute.Caret}`,
        html: `<div class="b3-list-item__first"><span class="b3-list-item__graphic">❗</span><span class="b3-list-item__text">${window.scribli.languages.callout} - <span style="color: var(--b3-callout-important)">Important</span></span></div>`,
    }, {
        filter: [window.scribli.languages.callout, "callout", "ts", "tishi", "warning"],
        id: "calloutWarning",
        value: `> [!WARNING]\n> ${Lute.Caret}`,
        html: `<div class="b3-list-item__first"><span class="b3-list-item__graphic">⚠️</span><span class="b3-list-item__text">${window.scribli.languages.callout} - <span style="color: var(--b3-callout-warning)">Warning</span></span></div>`,
    }, {
        filter: [window.scribli.languages.callout, "callout", "ts", "tishi", "caution"],
        id: "calloutCaution",
        value: `> [!CAUTION]\n> ${Lute.Caret}`,
        html: `<div class="b3-list-item__first"><span class="b3-list-item__graphic">🚨</span><span class="b3-list-item__text">${window.scribli.languages.callout} - <span style="color: var(--b3-callout-caution)">Caution</span></span></div>`,
    }, {
        filter: [window.scribli.languages.code, "code block", "daimakuai", "dmk"],
        id: "code",
        value: "```",
        html: `<div class="b3-list-item__first"><svg class="b3-list-item__graphic"><use xlink:href="#iconCode"></use></svg><span class="b3-list-item__text">${window.scribli.languages.code}</span>${getHotkeyOrMarker(window.scribli.config.keymap.editor.insert.code.custom, "```" + window.scribli.languages.enterKey)}</div>`,
    }, {
        filter: [window.scribli.languages.canvas, "canvas", "visual board", "board"],
        id: "canvas",
        value: "```scribli-canvas\n" + Lute.Caret + "\n```",
        html: `<div class="b3-list-item__first"><svg class="b3-list-item__graphic"><use xlink:href="#iconBoth"></use></svg><span class="b3-list-item__text">${window.scribli.languages.canvas}</span></div>`,
    }, {
        filter: [window.scribli.languages.table, "table", "biaoge", "bg"],
        id: "table",
        value: `| ${Lute.Caret} |  |  |\n| --- | --- | --- |\n|  |  |  |\n|  |  |  |`,
        html: `<div class="b3-list-item__first"><svg class="b3-list-item__graphic"><use xlink:href="#iconTable"></use></svg><span class="b3-list-item__text">${window.scribli.languages.table}</span><span class="b3-menu__accelerator b3-menu__accelerator--hotkey">${updateHotkeyTip((window.scribli.config.keymap.editor.insert.table.custom))}</span></div>`,
    }, {
        filter: [window.scribli.languages.line, "thematic break", "divider", "fengexian", "fgx"],
        id: "line",
        value: "---",
        html: `<div class="b3-list-item__first"><svg class="b3-list-item__graphic"><use xlink:href="#iconLine"></use></svg><span class="b3-list-item__text">${window.scribli.languages.line}</span><span class="b3-list-item__meta">---</span></div>`,
    }, {
        filter: [window.scribli.languages.math, "formulas block", "math block", "shuxuegongshikuai", "sxgsk"],
        id: "math",
        value: "$$",
        html: `<div class="b3-list-item__first"><svg class="b3-list-item__graphic"><use xlink:href="#iconMath"></use></svg><span class="b3-list-item__text">${window.scribli.languages.math}</span><span class="b3-list-item__meta">$$</span></div>`,
    }, {
        filter: ["html"],
        id: "html",
        value: "<div>",
        html: '<div class="b3-list-item__first"><svg class="b3-list-item__graphic"><use xlink:href="#iconHTML5"></use></svg><span class="b3-list-item__text">HTML</span></div>',
    }, {
        value: "",
        id: "separator_2",
        html: "separator",
    }, {
        filter: [window.scribli.languages.emoji, "emoji", "biaoqing", "bq"],
        id: "emoji",
        value: "emoji",
        html: `<div class="b3-list-item__first"><svg class="b3-list-item__graphic"><use xlink:href="#iconEmoji"></use></svg><span class="b3-list-item__text">${window.scribli.languages.emoji}</span><span class="b3-list-item__meta">:</span></div>`,
    }, {
        filter: [window.scribli.languages.link, "link", "a", "lianjie", "lj"],
        id: "link",
        value: "a",
        html: `<div class="b3-list-item__first"><svg class="b3-list-item__graphic"><use xlink:href="#iconLink"></use></svg><span class="b3-list-item__text">${window.scribli.languages.link}</span><span class="b3-menu__accelerator b3-menu__accelerator--hotkey">${updateHotkeyTip((window.scribli.config.keymap.editor.insert.link.custom))}</span></div>`,
    }, {
        filter: [window.scribli.languages.bold, "bold", "strong", "cuti", "ct", "jiacu", "jc"],
        id: "bold",
        value: "strong",
        html: `<div class="b3-list-item__first"><svg class="b3-list-item__graphic"><use xlink:href="#iconBold"></use></svg><span class="b3-list-item__text">${window.scribli.languages.bold}</span><span class="b3-menu__accelerator b3-menu__accelerator--hotkey">${updateHotkeyTip((window.scribli.config.keymap.editor.insert.bold.custom))}</span></div>`,
    }, {
        filter: [window.scribli.languages.italic, "italic", "em", "xieti", "xt"],
        id: "italic",
        value: "em",
        html: `<div class="b3-list-item__first"><svg class="b3-list-item__graphic"><use xlink:href="#iconItalic"></use></svg><span class="b3-list-item__text">${window.scribli.languages.italic}</span><span class="b3-menu__accelerator b3-menu__accelerator--hotkey">${updateHotkeyTip((window.scribli.config.keymap.editor.insert.italic.custom))}</span></div>`,
    }, {
        filter: [window.scribli.languages.underline, "underline", "xiahuaxian", "xhx"],
        id: "underline",
        value: "u",
        html: `<div class="b3-list-item__first"><svg class="b3-list-item__graphic"><use xlink:href="#iconUnderline"></use></svg><span class="b3-list-item__text">${window.scribli.languages.underline}</span><span class="b3-menu__accelerator b3-menu__accelerator--hotkey">${updateHotkeyTip((window.scribli.config.keymap.editor.insert.underline.custom))}</span></div>`,
    }, {
        filter: [window.scribli.languages.strike, "strike", "delete", "shanchuxian", "scx"],
        id: "strike",
        value: "s",
        html: `<div class="b3-list-item__first"><svg class="b3-list-item__graphic"><use xlink:href="#iconStrike"></use></svg><span class="b3-list-item__text">${window.scribli.languages.strike}</span><span class="b3-menu__accelerator b3-menu__accelerator--hotkey">${updateHotkeyTip((window.scribli.config.keymap.editor.insert.strike.custom))}</span></div>`,
    }, {
        filter: [window.scribli.languages.mark, "mark", "biaoji", "bj", "gaoliang", "gl"],
        id: "mark",
        value: "mark",
        html: `<div class="b3-list-item__first"><svg class="b3-list-item__graphic"><use xlink:href="#iconMark"></use></svg><span class="b3-list-item__text">${window.scribli.languages.mark}</span><span class="b3-menu__accelerator b3-menu__accelerator--hotkey">${updateHotkeyTip((window.scribli.config.keymap.editor.insert.mark.custom))}</span></div>`,
    }, {
        filter: [window.scribli.languages.sup, "superscript", "shangbiao", "sb"],
        id: "sup",
        value: "sup",
        html: `<div class="b3-list-item__first"><svg class="b3-list-item__graphic"><use xlink:href="#iconSup"></use></svg><span class="b3-list-item__text">${window.scribli.languages.sup}</span><span class="b3-menu__accelerator b3-menu__accelerator--hotkey">${updateHotkeyTip((window.scribli.config.keymap.editor.insert.sup.custom))}</span></div>`,
    }, {
        filter: [window.scribli.languages.sub, "subscript", "xiaobiao", "xb"],
        id: "sub",
        value: "sub",
        html: `<div class="b3-list-item__first"><svg class="b3-list-item__graphic"><use xlink:href="#iconSub"></use></svg><span class="b3-list-item__text">${window.scribli.languages.sub}</span><span class="b3-menu__accelerator b3-menu__accelerator--hotkey">${updateHotkeyTip((window.scribli.config.keymap.editor.insert.sub.custom))}</span></div>`,
    }, {
        filter: [window.scribli.languages["inline-code"], "inline code", "hangjidaima", "hjdm"],
        id: "inlineCode",
        value: "code",
        html: `<div class="b3-list-item__first"><svg class="b3-list-item__graphic"><use xlink:href="#iconInlineCode"></use></svg><span class="b3-list-item__text">${window.scribli.languages["inline-code"]}</span><span class="b3-menu__accelerator b3-menu__accelerator--hotkey">${updateHotkeyTip((window.scribli.config.keymap.editor.insert["inline-code"].custom))}</span></div>`,
    }, {
        filter: [window.scribli.languages.kbd, "kbd", "jianpan", "jp"],
        id: "kbd",
        value: "kbd",
        html: `<div class="b3-list-item__first"><svg class="b3-list-item__graphic"><use xlink:href="#iconKeymap"></use></svg><span class="b3-list-item__text">${window.scribli.languages.kbd}</span><span class="b3-menu__accelerator b3-menu__accelerator--hotkey">${updateHotkeyTip((window.scribli.config.keymap.editor.insert.kbd.custom))}</span></div>`,
    }, {
        filter: [window.scribli.languages.tag, "tags", "biaoqian", "bq"],
        id: "tag",
        value: "tag",
        html: `<div class="b3-list-item__first"><svg class="b3-list-item__graphic"><use xlink:href="#iconTag"></use></svg><span class="b3-list-item__text">${window.scribli.languages.tag}</span><span class="b3-menu__accelerator b3-menu__accelerator--hotkey">${updateHotkeyTip((window.scribli.config.keymap.editor.insert.tag.custom))}</span></div>`,
    }, {
        filter: [window.scribli.languages["inline-math"], "inline formulas", "inline math", "hangjigongshi", "hjgs", "hangjishuxvegongshi", "hangjishuxuegongshi", "hjsxgs"],
        id: "inlineMath",
        value: "inline-math",
        html: `<div class="b3-list-item__first"><svg class="b3-list-item__graphic"><use xlink:href="#iconMath"></use></svg><span class="b3-list-item__text">${window.scribli.languages["inline-math"]}</span><span class="b3-menu__accelerator b3-menu__accelerator--hotkey">${updateHotkeyTip((window.scribli.config.keymap.editor.insert["inline-math"].custom))}</span></div>`,
    }, {
        value: "",
        id: "separator_3",
        html: "separator",
    }, {
        filter: [window.scribli.languages.insertAsset, "insert image or file", "upload", "charutupianhuowenjian", "crtphwj", "sc"],
        id: "insertAsset",
        value: Constants.ZWSP + 3,
        html: `<div class="b3-list-item__first"><svg class="b3-list-item__graphic"><use xlink:href="#iconDownload"></use></svg><span class="b3-list-item__text">${window.scribli.languages.insertAsset}</span>
<input class="b3-form__upload" type="file" ${protyle.options.upload.accept ? 'multiple="' + protyle.options.upload.accept + '"' : ""}></div>`,
    }, {
        filter: [window.scribli.languages.insertIframeURL, "insert iframe link", "charuiframelianjie", "criframelj"],
        id: "insertIframeURL",
        value: '<iframe sandbox="allow-forms allow-presentation allow-same-origin allow-scripts allow-modals allow-popups allow-storage-access-by-user-activation" src="" border="0" frameborder="no" framespacing="0" allowfullscreen="true"></iframe>',
        html: `<div class="b3-list-item__first"><svg class="b3-list-item__graphic"><use xlink:href="#iconGlobe"></use></svg><span class="b3-list-item__text">${window.scribli.languages.insertIframeURL}</span></div>`,
    }, {
        filter: [window.scribli.languages.insertImgURL, "insert image link", "image", "img", "charutupianlianjie", "crtplj"],
        id: "insertImgURL",
        value: "![]()",
        html: `<div class="b3-list-item__first"><svg class="b3-list-item__graphic"><use xlink:href="#iconImage"></use></svg><span class="b3-list-item__text">${window.scribli.languages.insertImgURL}</span></div>`,
    }, {
        filter: [window.scribli.languages.insertVideoURL, "insert video link", "charushipinlianjie", "crsplj"],
        id: "insertVideoURL",
        value: '<video controls="controls" src=""></video>',
        html: `<div class="b3-list-item__first"><svg class="b3-list-item__graphic"><use xlink:href="#iconVideo"></use></svg><span class="b3-list-item__text">${window.scribli.languages.insertVideoURL}</span></div>`,
    }, {
        filter: [window.scribli.languages.insertAudioURL, "insert audio link", "charuyinpinlianjie", "cryplj"],
        id: "insertAudioURL",
        value: '<audio controls="controls" src=""></audio>',
        html: `<div class="b3-list-item__first"><svg class="b3-list-item__graphic"><use xlink:href="#iconRecord"></use></svg><span class="b3-list-item__text">${window.scribli.languages.insertAudioURL}</span></div>`,
    }, {
        value: "",
        id: "separator_4",
        html: "separator",
    }, {
        filter: [window.scribli.languages.staff, "staff", "wuxianpu", "wxp"],
        id: "staff",
        value: "```abc\n```",
        html: `<div class="b3-list-item__first"><span class="b3-list-item__text">ABC</span><span class="b3-list-item__meta">${window.scribli.languages.staff}</span></div>`,
    }, {
        filter: [window.scribli.languages.chart, "chart", "tubiao", "tb"],
        id: "chart",
        value: "```echarts\n```",
        html: `<div class="b3-list-item__first"><span class="b3-list-item__text">Chart</span><span class="b3-list-item__meta">${window.scribli.languages.chart}</span></div>`,
    }, {
        filter: ["flowchart", "flow chart", "liuchengtu", "lct"],
        id: "flowChart",
        value: "```flowchart\n```",
        html: '<div class="b3-list-item__first"><span class="b3-list-item__text">FlowChart</span><span class="b3-list-item__meta">Flow Chart</span></div>',
    }, {
        filter: ["graphviz", "zhuangtaitu", "ztt"],
        id: "graph",
        value: "```graphviz\n```",
        html: '<div class="b3-list-item__first"><span class="b3-list-item__text">Graphviz</span><span class="b3-list-item__meta">Graph</span></div>',
    }, {
        filter: ["mermaid", "diagram", "tubiao", "tb"],
        id: "mermaid",
        value: "```mermaid\n```",
        html: '<div class="b3-list-item__first"><span class="b3-list-item__text">Mermaid</span><span class="b3-list-item__meta">Mermaid</span></div>',
    }, {
        filter: [window.scribli.languages.mindmap, "mindmap", "naotu", "nt"],
        id: "mindmap",
        value: "```mindmap\n```",
        html: `<div class="b3-list-item__first"><span class="b3-list-item__text">Mind map</span><span class="b3-list-item__meta">${window.scribli.languages.mindmap}</span></div>`,
    }, {
        filter: ["plantuml", "jianmoyuyan", "jmyy"],
        id: "UML",
        value: "```plantuml\n```",
        html: '<div class="b3-list-item__first"><span class="b3-list-item__text">PlantUML</span><span class="b3-list-item__meta">UML</span></div>',
    }, {
        value: "",
        id: "separator_5",
        html: "separator",
    }, {
        filter: [window.scribli.languages.infoStyle, "info style", "xinxiyangshi", "xxys"],
        id: "infoStyle",
        value: `style${Constants.ZWSP}color: var(--b3-card-info-color);background-color: var(--b3-card-info-background);`,
        html: `<div class="b3-list-item__first"><div style="color: var(--b3-card-info-color);background-color: var(--b3-card-info-background);" class="color__square color__square--list">A</div><span class="b3-list-item__text">${window.scribli.languages.infoStyle}</span></div>`,
    }, {
        filter: [window.scribli.languages.successStyle, "success style", "chenggongyangshi", "cgys"],
        id: "successStyle",
        value: `style${Constants.ZWSP}color: var(--b3-card-success-color);background-color: var(--b3-card-success-background);`,
        html: `<div class="b3-list-item__first"><div style="color: var(--b3-card-success-color);background-color: var(--b3-card-success-background);" class="color__square color__square--list">A</div><span class="b3-list-item__text">${window.scribli.languages.successStyle}</span></div>`,
    }, {
        filter: [window.scribli.languages.warningStyle, "warning style", "jinggaoyangshi", "jgys"],
        id: "warningStyle",
        value: `style${Constants.ZWSP}color: var(--b3-card-warning-color);background-color: var(--b3-card-warning-background);`,
        html: `<div class="b3-list-item__first"><div style="color: var(--b3-card-warning-color);background-color: var(--b3-card-warning-background);" class="color__square color__square--list">A</div><span class="b3-list-item__text">${window.scribli.languages.warningStyle}</span></div>`,
    }, {
        filter: [window.scribli.languages.errorStyle, "error style", "cuowuyangshi", "cwys"],
        id: "errorStyle",
        value: `style${Constants.ZWSP}color: var(--b3-card-error-color);background-color: var(--b3-card-error-background);`,
        html: `<div class="b3-list-item__first"><div style="color: var(--b3-card-error-color);background-color: var(--b3-card-error-background);" class="color__square color__square--list">A</div><span class="b3-list-item__text">${window.scribli.languages.errorStyle}</span></div>`,
    }, {
        filter: [window.scribli.languages.clearFontStyle, "clear style", "qingchuyangshi", "qcys"],
        id: "clearFontStyle",
        value: `style${Constants.ZWSP}`,
        html: `<div class="b3-list-item__first"><div class="color__square color__square--list">A</div><span class="b3-list-item__text">${window.scribli.languages.clearFontStyle}</span></div>`,
    }, {
        value: "",
        id: "separator_6",
        html: "separator",
    }];
    let hasPlugin = false;
    protyle.app.plugins.forEach((plugin) => {
        plugin.protyleSlash.forEach(slash => {
            allList.push({
                filter: slash.filter,
                id: slash.id,
                value: `plugin${Constants.ZWSP}${plugin.name}${Constants.ZWSP}${slash.id}`,
                html: slash.html
            });
            hasPlugin = true;
        });
    });
    if (!hasPlugin) {
        allList.pop();
    }
    if (key === "") {
        return allList;
    }
    return allList.filter((item) => {
        if (!item.filter) {
            return false;
        }
        const match = item.filter.find((filter) => {
            if (filter.toLowerCase().indexOf(key.toLowerCase()) > -1) {
                return true;
            }
        });
        if (match) {
            return true;
        } else {
            return false;
        }
    });
};

export const hintTag = (key: string, protyle: IProtyle): IHintData[] => {
    protyle.hint.genLoading(protyle);
    fetchPost("/api/search/searchTag", {
        k: key,
    }, (response) => {
        if (protyle.hint.element.classList.contains("fn__none")) {
            return;
        }
        const dataList: IHintData[] = [];
        let hasKey = false;
        response.data.tags.forEach((item: string) => {
            const value = item.replace(/<mark>/g, "").replace(/<\/mark>/g, "");
            dataList.push({
                value: `<span data-type="tag">${value}</span>`,
                html: `<div class="b3-list-item__text">${item}</div>`,
            });
            if (value === response.data.k) {
                hasKey = true;
            }
        });
        if (response.data.k && !hasKey) {
            dataList.splice(0, 0, {
                value: `<span data-type="tag">${response.data.k}</span>`,
                html: `<div class="b3-list-item__text">${window.scribli.languages.newTag} <mark>${escapeHtml(response.data.k)}</mark></div>`,
            });
            if (dataList.length > 1) {
                dataList[1].focus = true;
            }
        }
        protyle.hint.genHTML(dataList, protyle, true, "hint");
    });

    return [];
};

export const genHintItemHTML = (item: IBlock) => {
    let iconHTML;
    if (item.type === "NodeDocument" && item.ial.icon) {
        iconHTML = unicode2Emoji(item.ial.icon, "b3-list-item__graphic popover__block", true);
        iconHTML = iconHTML.replace('popover__block"', `popover__block" data-id="${item.id}"`);
    } else {
        iconHTML = `<svg class="b3-list-item__graphic popover__block" data-id="${item.id}"><use xlink:href="#${getIconByType(item.type)}"></use></svg>`;
    }
    let attrHTML = "";
    if (item.name) {
        attrHTML += `<span class="fn__flex"><svg class="b3-list-item__hinticon"><use xlink:href="#iconN"></use></svg><span>${item.name}</span></span><span class="fn__space"></span>`;
    }
    if (item.alias) {
        attrHTML += `<span class="fn__flex"><svg class="b3-list-item__hinticon"><use xlink:href="#iconA"></use></svg><span>${item.alias}</span></span><span class="fn__space"></span>`;
    }
    if (item.memo) {
        attrHTML += `<span class="fn__flex"><svg class="b3-list-item__hinticon"><use xlink:href="#iconM"></use></svg><span>${item.memo}</span></span>`;
    }
    if (attrHTML) {
        attrHTML = `<div class="fn__flex b3-list-item__meta b3-list-item__showall">${attrHTML}</div>`;
    }
    let countHTML = "";
    if (item.refCount) {
        countHTML = `<span class="popover__block counter b3-tooltips b3-tooltips__w" aria-label="${window.scribli.languages.ref}">${item.refCount}</span>`;
    }
    return `${attrHTML}<div class="b3-list-item__first" data-node-id="${item.id}">
    ${iconHTML}
    <span class="b3-list-item__text">${item.content}</span>${countHTML}
</div>
<div class="b3-list-item__meta b3-list-item__showall">${item.hPath}</div>`;
};

export const hintRef = (key: string, protyle: IProtyle, source: THintSource): IHintData[] => {
    const nodeElement = hasClosestBlock(getEditorRange(protyle.wysiwyg.element).startContainer);
    protyle.hint.genLoading(protyle);
    let refParam: IObject;
    if (protyle.lite) {
        refParam = {k: key, id: "", rootID: "", beforeLen: 48, isDatabase: false, isSquareBrackets: true};
    } else {
        refParam = {
            k: key,
            id: nodeElement ? nodeElement.getAttribute("data-node-id") : protyle.block.parentID,
            beforeLen: Math.floor((Math.max(protyle.element.clientWidth / 2, 320) - 58) / 28.8),
            rootID: source === "av" ? "" : protyle.block.rootID,
            isDatabase: source === "av",
            isSquareBrackets: ["[[", "【【"].includes(protyle.hint.splitChar)
        };
        if (isEncryptedBox(protyle.notebookId)) {
            refParam.notebook = protyle.notebookId;
        }
    }
    fetchPost("/api/search/searchRefBlock", refParam, (response) => {
        const dataList: IHintData[] = [];
        if (response.data.newDoc) {
            const newFileName = Lute.UnEscapeHTMLStr(replaceFileName(response.data.k));
            dataList.push({
                value: `((newFile "${newFileName}"${Constants.ZWSP}'${newFileName}${Lute.Caret}'))`,
                html: `<div class="b3-list-item__first"><svg class="b3-list-item__graphic"><use xlink:href="#iconFile"></use></svg>
<span class="b3-list-item__text">${window.scribli.languages.newFile} <mark>${response.data.k}</mark></span></div>`,
            });
        }
        response.data.blocks.forEach((item: IBlock) => {
            const name = item.name ? stripSearchMark(item.name) : item.refText.replace(new RegExp(Constants.ZWSP, "g"), "");
            let value = `<span data-type="block-ref" data-id="${item.id}" data-subtype="d">${name}</span>`;
            if (source === "search") {
                value = `<span data-type="block-ref" data-id="${item.id}" data-subtype="s">${key}${Constants.ZWSP}${name}</span>`;
            } else if (source === "av") {
                let refText = name;
                if (nodeElement) {
                    refText = item.ial["custom-sy-av-s-text-" + nodeElement.getAttribute("data-av-id")] || refText;
                }
                value = `<span data-type="block-ref" data-id="${item.id}" data-subtype="s">${refText}</span>`;
            }
            dataList.push({
                value,
                html: genHintItemHTML(item),
            });
        });
        if (source === "search") {
            protyle.hint.splitChar = "((";
            protyle.hint.lastIndex = -1;
        }
        if (dataList.length === 0) {
            dataList.push({
                value: "",
                html: window.scribli.languages.emptyContent,
            });
        } else if (response.data.newDoc && dataList.length > 1) {
            dataList[1].focus = true;
        }
        protyle.hint.genHTML(dataList, protyle, true, source);
    });
    return [];
};

export const hintEmbed = (key: string, protyle: IProtyle): IHintData[] => {
    if (key.endsWith("}}") || key.endsWith("」」")) {
        return [];
    }
    protyle.hint.genLoading(protyle);
    const nodeElement = hasClosestBlock(getEditorRange(protyle.wysiwyg.element).startContainer);
    const embedParam: IObject = {
        k: key,
        isDatabase: false,
        beforeLen: Math.floor((Math.max(protyle.element.clientWidth / 2, 320) - 58) / 28.8),
        id: nodeElement ? nodeElement.getAttribute("data-node-id") : protyle.block.parentID,
        rootID: protyle.block.rootID,
    };
    if (isEncryptedBox(protyle.notebookId)) {
        embedParam.notebook = protyle.notebookId;
    }
    fetchPost("/api/search/searchRefBlock", embedParam, (response) => {
        const dataList: IHintData[] = [];
        response.data.blocks.forEach((item: IBlock) => {
            dataList.push({
                value: `{{select * from blocks where id='${item.id}'}}`,
                html: genHintItemHTML(item),
            });
        });
        if (dataList.length === 0) {
            dataList.push({
                value: "",
                html: window.scribli.languages.emptyContent,
            });
        }
        protyle.hint.genHTML(dataList, protyle, true, "hint");
    });
    return [];
};

export const hintRenderTemplate = (value: string, protyle: IProtyle, nodeElement: Element) => {
    fetchPost("/api/template/render", {
        id: protyle.block.parentID,
        path: value
    }, (response) => {
        focusByRange(protyle.toolbar.range);
        const editElement = getContenteditableElement(nodeElement);
        if (editElement && editElement.textContent.trim() === "") {
            insertHTML(response.data.content, protyle, true);
        } else {
            insertHTML(response.data.content, protyle);
        }
        // 
        protyle.wysiwyg.element.querySelectorAll('[status="temp"]').forEach(item => {
            item.remove();
        });
        blockRender(protyle, protyle.wysiwyg.element);
        processRender(protyle.wysiwyg.element);
        highlightRender(protyle.wysiwyg.element);
        avRender(protyle.wysiwyg.element, protyle);
        hideElements(["util"], protyle);
    });
};

export const hintRenderWidget = (value: string, protyle: IProtyle) => {
    focusByRange(protyle.toolbar.range);
    // Use the path ending with `/` when loading the widget 
    insertHTML(protyle.lute.SpinBlockDOM(`<iframe src="/widgets/${value}/" data-subtype="widget" border="0" frameborder="no" framespacing="0" allowfullscreen="true"></iframe>`), protyle, true);
    hideElements(["util"], protyle);
};

export const hintRenderAssets = (value: string, protyle: IProtyle) => {
    focusByRange(protyle.toolbar.range);
    const type = pathPosix().extname(value).toLowerCase();
    const filename = value.startsWith("assets/") ? getAssetName(value) : value;
    insertHTML(genAssetHTML(type, value, filename, value.startsWith("assets/") ? filename + type : value), protyle);
    hideElements(["util"], protyle);
};

export const hintMoveBlock = (pathString: string, sourceElements: Element[], protyle: IProtyle) => {
    if (pathString === "/") {
        return;
    }
    const parentID = getDisplayName(pathString, true, true);
    if (protyle.block.rootID === parentID) {
        return;
    }
    const doOperations: IOperation[] = [];
    let topSourceElement: Element;
    const parentElement = sourceElements[0].parentElement;
    let sideElement;
    sourceElements.forEach((item, index) => {
        if (index === sourceElements.length - 1 &&
            item.parentElement) {
            topSourceElement = getTopAloneElement(item);
            sideElement = topSourceElement.nextElementSibling || topSourceElement.previousElementSibling;
            if (topSourceElement === item) {
                topSourceElement = undefined;
            }
        }
        doOperations.push({
            action: "append",
            id: item.getAttribute("data-node-id"),
            parentID,
        });
        item.remove();
    });
    if (topSourceElement) {
        doOperations.push({
            action: "delete",
            id: topSourceElement.getAttribute("data-node-id"),
        });
        topSourceElement.remove();
    } else if (parentElement.classList.contains("list") && parentElement.getAttribute("data-subtype") === "o" &&
        parentElement.childElementCount > 1) {
        updateListOrder(parentElement, 1);
        Array.from(parentElement.children).forEach((item) => {
            if (item.classList.contains("protyle-attr")) {
                return;
            }
            item.setAttribute(Constants.ATTRIBUTE_EDITING, "true");
            doOperations.push({
                action: "update",
                id: item.getAttribute("data-node-id"),
                data: item.outerHTML
            });
        });
    } else if (protyle.block.showAll && parentElement.classList.contains("protyle-wysiwyg") && parentElement.childElementCount === 0) {
        setTimeout(() => {
            zoomOut({protyle, id: protyle.block.parent2ID, focusId: protyle.block.parent2ID});
        }, Constants.TIMEOUT_INPUT * 2 + 100);
    } else if (parentElement.classList.contains("protyle-wysiwyg") && parentElement.innerHTML === "" &&
        !hasClosestByClassName(parentElement, "block__edit", true) &&
        protyle.block.id === protyle.block.rootID) {
        const newId = Lute.NewNodeID();
        const newElement = genEmptyElement(false, false, newId);
        doOperations.splice(0, 0, {
            action: "insert",
            id: newId,
            data: newElement.outerHTML,
            parentID: protyle.block.parentID
        });
        parentElement.innerHTML = newElement.outerHTML;
        focusBlock(newElement);
    } else if (sideElement) {
        focusBlock(sideElement);
    }
    transaction(protyle, doOperations);
};
