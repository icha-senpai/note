import {Constants} from "../../constants";
import {isBrowser, isMobile} from "../../util/functions";
import {updateHotkeyTip} from "../../protyle/util/compatibility";
import {editorConfigApi} from "./editorRuntime";
import type {SettingTabBuilder} from "../setting/builder";
/// #if !BROWSER
import {ipcRenderer} from "electron";
/// #endif

const registerEditorBehaviorGroup = (tab: SettingTabBuilder) => {
    const group = tab.group("behavior", window.scribli.languages.configGroupBehavior);
    const readOnlyKeymap = window.scribli.config.keymap.general.editReadonly.custom;
    group.switch("editor.readOnly", {
        title: isMobile()
            ? window.scribli.languages.editReadonly
            : `${window.scribli.languages.editReadonly} <code class="fn__code${readOnlyKeymap ? "" : " fn__none"}">${updateHotkeyTip(readOnlyKeymap)}</code>`,
        desc: window.scribli.languages.editReadonlyTip,
    });
    group.switch("editor.spellcheck", {
        title: window.scribli.languages.spellcheck,
        desc: isBrowser() ? window.scribli.languages.spellcheckTip : window.scribli.languages.spellcheckTip2,
        /// #if !BROWSER
        afterMount: bindSpellcheckLanguagesVisibility,
        /// #endif
    });
    /// #if !BROWSER
    group.slot({
        key: "spellcheckLanguages",
        keywords: [
            window.scribli.languages.spellcheck,
            window.scribli.languages.spellcheckTip2,
        ],
        html: () => '<div class="fn__flex b3-label config-item fn__none"><div class="b3-chips" id="editor.spellcheckLanguages"></div></div>',
        afterMount: bindSpellcheckLanguagesChips,
    });
    /// #endif
    group.range("editor.codeTabSpaces", {
        title: window.scribli.languages.md29,
        desc: window.scribli.languages.md30,
        min: 0,
        max: 8,
        step: 2,
    });
    group.switch("editor.listLogicalOutdent", {
        title: window.scribli.languages.outlineOutdent,
        desc: window.scribli.languages.outlineOutdentTip,
    });
    group.switch("editor.listItemDotNumberClickFocus", {
        title: window.scribli.languages.listItemDotNumberClickFocus,
        desc: window.scribli.languages.listItemDotNumberClickFocusTip,
    });
    group.switch("editor.pasteURLAutoConvert", {
        title: window.scribli.languages.pasteURLAutoConvert,
        desc: window.scribli.languages.pasteURLAutoConvertTip,
    });
    group.number("editor.dynamicLoadBlocks", {
        title: window.scribli.languages.dynamicLoadBlocks,
        desc: window.scribli.languages.dynamicLoadBlocksTip,
        min: 48,
    });
};

/// #if !BROWSER
const bindSpellcheckLanguagesVisibility = async (root: HTMLElement) => {
    const spellcheckSwitch = root.querySelector<HTMLInputElement>(`#${CSS.escape("editor.spellcheck")}`);
    if (!spellcheckSwitch) {
        return;
    }
    const toggleWrap = () => {
        root.querySelector(`#${CSS.escape("editor.spellcheckLanguages")}`)?.closest(".config-item")?.classList.toggle("fn__none", !spellcheckSwitch.checked);
    };
    spellcheckSwitch.addEventListener("change", toggleWrap);
    toggleWrap();
};

const bindSpellcheckLanguagesChips = async (root: HTMLElement) => {
    const el = root.querySelector<HTMLDivElement>(`#${CSS.escape("editor.spellcheckLanguages")}`);
    if (!el) {
        return;
    }
    const languages: string[] = await ipcRenderer.invoke(Constants.SCRIBLI_GET, {
        cmd: "availableSpellCheckerLanguages",
    });
    el.innerHTML = languages.map((item) =>
        `<div class="fn__pointer b3-chip b3-chip--middle${window.scribli.config.editor.spellcheckLanguages.includes(item) ? " b3-chip--current" : ""}">${item}</div>`
    ).join("");
    el.addEventListener("click", (event) => {
        const target = event.target as Element;
        if (target.classList.contains("b3-chip")) {
            target.classList.toggle("b3-chip--current");
            const selected = Array.from(el.querySelectorAll(".b3-chip--current")).map((chip) => chip.textContent || "");
            ipcRenderer.send(Constants.SCRIBLI_CMD, {
                cmd: "setSpellCheckerLanguages",
                languages: selected,
            });
            editorConfigApi.patch("spellcheckLanguages", selected);
        }
    });
};
/// #endif

const registerEditorBlockFeaturesGroup = (tab: SettingTabBuilder) => {
    const group = tab.group("blockFeatures", window.scribli.languages.configGroupBlockFeatures);
    group.switch("editor.displayNetImgMark", {
        title: window.scribli.languages.md7,
        desc: window.scribli.languages.md8,
    });
    group.switch("editor.displayBookmarkIcon", {
        title: window.scribli.languages.md12,
        desc: window.scribli.languages.md16,
    });
    group.switch("editor.embedBlockBreadcrumb", {
        title: window.scribli.languages.embedBlockBreadcrumb,
        desc: window.scribli.languages.embedBlockBreadcrumbTip,
    });
    group.select("editor.databaseAttrViewMode", {
        title: window.scribli.languages.databaseAttrViewMode,
        desc: window.scribli.languages.databaseAttrViewModeTip,
        options: [
            {value: 0, label: window.scribli.languages.expand},
            {value: 1, label: window.scribli.languages.collapse},
        ],
    });
    group.select("editor.headingEmbedMode", {
        title: window.scribli.languages.headingEmbedMode,
        desc: window.scribli.languages.headingEmbedModeTip,
        options: [
            {value: 0, label: window.scribli.languages.showHeadingWithBlocks},
            {value: 1, label: window.scribli.languages.showHeadingOnlyTitle},
            {value: 2, label: window.scribli.languages.showHeadingOnlyBlocks},
        ],
    });
    group.switch("editor.codeLineWrap", {
        title: window.scribli.languages.md31,
        desc: window.scribli.languages.md32,
    });
    group.switch("editor.codeLigatures", {
        title: window.scribli.languages.md2,
        desc: window.scribli.languages.md3,
    });
    group.switch("editor.codeSyntaxHighlightLineNum", {
        title: window.scribli.languages.md27,
        desc: window.scribli.languages.md28,
    });
};

const registerEditorBidirectionalGroup = (tab: SettingTabBuilder) => {
    const group = tab.group("bidirectional", window.scribli.languages.configGroupBidirectionalLinks);
    group.switch("editor.onlySearchForDoc", {
        title: window.scribli.languages.onlySearchForDoc,
        desc: window.scribli.languages.onlySearchForDocTip,
    });
    group.number("editor.blockRefDynamicAnchorTextMaxLen", {
        title: window.scribli.languages.md37,
        desc: window.scribli.languages.md38,
        min: 1,
        max: 5120,
    });
    group.switch("editor.virtualBlockRef", {
        title: window.scribli.languages.md33,
        desc: window.scribli.languages.md34,
    });
    group.textBlock("editor.virtualBlockRefInclude", {
        title: window.scribli.languages.md9,
        desc: window.scribli.languages.md36,
        mode: "textarea",
    });
    group.textBlock("editor.virtualBlockRefExclude", {
        title: window.scribli.languages.md35,
        desc: window.scribli.languages.md41,
        mode: "textarea",
    });
    group.switch("editor.backlinkContainChildren", {
        title: window.scribli.languages.backlinkContainChildren,
        desc: window.scribli.languages.backlinkContainChildrenTip,
    });
    group.number("editor.backlinkExpandCount", {
        title: window.scribli.languages.backlinkExpand,
        desc: window.scribli.languages.backlinkExpandTip,
        min: 0,
        max: 512,
    });
    group.number("editor.backmentionExpandCount", {
        title: window.scribli.languages.backmentionExpand,
        desc: window.scribli.languages.backmentionExpandTip,
        min: -1,
        max: 512,
    });
};

const registerEditorMarkdownInlineGroup = (tab: SettingTabBuilder) => {
    const group = tab.group("markdownInline", window.scribli.languages.configGroupMarkdownInlineSyntax);
    group.switch("editor.markdown.inlineAsterisk", {
        title: window.scribli.languages.editorMarkdownInlineAsterisk,
        desc: window.scribli.languages.editorMarkdownInlineAsteriskTip,
    });
    group.switch("editor.markdown.inlineUnderscore", {
        title: window.scribli.languages.editorMarkdownInlineUnderscore,
        desc: window.scribli.languages.editorMarkdownInlineUnderscoreTip,
    });
    group.switch("editor.markdown.inlineSup", {
        title: window.scribli.languages.editorMarkdownInlineSup,
        desc: window.scribli.languages.editorMarkdownInlineSupTip,
    });
    group.switch("editor.markdown.inlineSub", {
        title: window.scribli.languages.editorMarkdownInlineSub,
        desc: window.scribli.languages.editorMarkdownInlineSubTip,
    });
    group.switch("editor.markdown.inlineTag", {
        title: window.scribli.languages.editorMarkdownInlineTag,
        desc: window.scribli.languages.editorMarkdownInlineTagTip,
    });
    group.switch("editor.markdown.inlineMath", {
        title: window.scribli.languages.editorMarkdownInlineMath,
        desc: window.scribli.languages.editorMarkdownInlineMathTip,
    });
    group.switch("editor.markdown.inlineStrikethrough", {
        title: window.scribli.languages.editorMarkdownInlineStrikethrough,
        desc: window.scribli.languages.editorMarkdownInlineStrikethroughTip,
    });
    group.switch("editor.markdown.inlineMark", {
        title: window.scribli.languages.editorMarkdownInlineMark,
        desc: window.scribli.languages.editorMarkdownInlineMarkTip,
    });
};

const registerEditorAdvancedGroup = (tab: SettingTabBuilder) => {
    const group = tab.group("advanced", window.scribli.languages.configGroupAdvanced);
    group.text("editor.plantUMLServePath", {
        title: window.scribli.languages.md39,
        desc: window.scribli.languages.md40,
    });
    group.textBlock("editor.katexMacros", {
        title: window.scribli.languages.katexMacros,
        desc: window.scribli.languages.katexMacrosTip,
        mode: "textarea",
    });
    group.switch("editor.allowSVGScript", {
        title: window.scribli.languages.allowSVGScript,
        desc: window.scribli.languages.allowSVGScriptTip,
    });
    group.switch("editor.allowHTMLBLockScript", {
        title: window.scribli.languages.allowHTMLBLockScript,
        desc: window.scribli.languages.allowHTMLBLockScriptTip,
    });
};

export const registerEditorTab = (tab: SettingTabBuilder) => {
    registerEditorBehaviorGroup(tab);
    registerEditorBlockFeaturesGroup(tab);
    registerEditorBidirectionalGroup(tab);
    registerEditorMarkdownInlineGroup(tab);
    registerEditorAdvancedGroup(tab);
};
