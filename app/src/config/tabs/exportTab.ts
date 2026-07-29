/// #if !BROWSER
import {ipcRenderer} from "electron";
import {useShell} from "../../util/pathName";
/// #endif
import type {SettingTabBuilder} from "../setting/builder";
import {Constants} from "../../constants";
import {exportConfigApi} from "./exportRuntime";

const registerExportReferencesGroup = (tab: SettingTabBuilder) => {
    const group = tab.group("references", window.scribli.languages.configGroupReferences);

    group.switch("export.includeSubDocs", {
        title: window.scribli.languages.includeSubDocs,
        desc: window.scribli.languages.includeSubDocsTip,
    });
    group.switch("export.includeRelatedDocs", {
        title: window.scribli.languages.includeRelatedDocs,
        desc: window.scribli.languages.includeRelatedDocsTip,
    });
    group.select("export.blockRefMode", {
        title: window.scribli.languages.ref,
        desc: window.scribli.languages.export11,
        options: [
            {value: 2, label: window.scribli.languages.export2},
            {value: 3, label: window.scribli.languages.export3},
            {value: 4, label: window.scribli.languages.export4},
        ],
    });
    group.select("export.blockEmbedMode", {
        title: window.scribli.languages.blockEmbed,
        desc: window.scribli.languages.export12,
        options: [
            {value: 0, label: window.scribli.languages.export0},
            {value: 1, label: window.scribli.languages.export1},
        ],
    });
};

const registerExportFormatGroup = (tab: SettingTabBuilder) => {
    const group = tab.group("format", window.scribli.languages.configGroupFormat);

    group.switch("export.markdownYFM", {
        title: window.scribli.languages.export23,
        desc: window.scribli.languages.export24,
    });
    group.switch("export.addTitle", {
        title: window.scribli.languages.export17,
        desc: window.scribli.languages.export18,
    });
    group.switch("export.paragraphBeginningSpace", {
        title: window.scribli.languages.paragraphBeginningSpace,
        desc: window.scribli.languages.md4,
    });
    group.switch("export.removeAssetsID", {
        title: window.scribli.languages.removeAssetsID,
        desc: window.scribli.languages.removeAssetsIDTip,
    });
    group.switch("export.inlineMemo", {
        title: window.scribli.languages.export31,
        desc: window.scribli.languages.export32,
    });
    group.textPair({
        title: window.scribli.languages.export13,
        desc: window.scribli.languages.export14,
        leftId: "export.blockRefTextLeft",
        rightId: "export.blockRefTextRight",
    });
    group.textPair({
        title: window.scribli.languages.export15,
        desc: window.scribli.languages.export16,
        leftId: "export.tagOpenMarker",
        rightId: "export.tagCloseMarker",
    });
};

const registerExportPdfGroup = (tab: SettingTabBuilder) => {
    const group = tab.group("pdf", window.scribli.languages.configGroupPDF);

    group.select("export.fileAnnotationRefMode", {
        title: window.scribli.languages.export5,
        desc: window.scribli.languages.export6,
        options: [
            {value: 0, label: window.scribli.languages.export7},
            {value: 1, label: window.scribli.languages.export8},
        ],
    });
    group.text("export.pdfFooter", {
        title: window.scribli.languages.export21,
        desc: window.scribli.languages.export22,
    });
    group.stack({
        key: "pdfWatermark",
        keywords: [
            window.scribli.languages.export27,
            window.scribli.languages.export28,
            window.scribli.languages.export29,
        ],
    }, (stack) => {
        stack.title(window.scribli.languages.export27);
        stack.desc(window.scribli.languages.export28);
        stack.textBlock("export.pdfWatermarkStr", {
            mode: "input-text",
        });
        stack.desc(`<a href="https://pdfcpu.io/core/watermark#description" target="_blank">${window.scribli.languages.export29}</a>`);
        stack.textBlock("export.pdfWatermarkDesc", {
            mode: "textarea",
        });
    });
};

const registerExportImagesGroup = (tab: SettingTabBuilder) => {
    const group = tab.group("images", window.scribli.languages.configGroupImages);

    group.stack({
        key: "imageWatermark",
        keywords: [
            window.scribli.languages.export30,
            window.scribli.languages.export28,
            window.scribli.languages.export29,
            window.scribli.languages.export10,
        ],
    }, (stack) => {
        stack.title(window.scribli.languages.export30);
        stack.desc(window.scribli.languages.export28);
        stack.textBlock("export.imageWatermarkStr", {
            mode: "input-text",
        });
        stack.desc(`${window.scribli.languages.export29}<div class="fn__hr--small"></div>${window.scribli.languages.export10}`);
        stack.textBlock("export.imageWatermarkDesc", {
            mode: "textarea",
        });
    });
};

/// #if !BROWSER
const registerExportPandocGroup = (tab: SettingTabBuilder) => {
    const group = tab.group("pandoc", window.scribli.languages.configGroupPandoc);

    group.stack({
        key: "pandocBin",
        keywords: [
            window.scribli.languages.export19,
            window.scribli.languages.export20,
            window.scribli.languages.reset,
            window.scribli.languages.config,
        ],
        afterMount: mountExportPandocStack,
    }, (stack) => {
        stack.title(`${window.scribli.languages.export19}<span class="fn__space"></span><a href="javascript:void(0)" id="pandocBinPathDisplay" style="word-break: break-all">${Lute.EscapeHTMLStr(window.scribli.config.export.pandocBin)}</a>`);
        stack.button({
            id: "pandocBinReset",
            label: window.scribli.languages.reset,
            icon: "iconUndo",
        });
        stack.desc(window.scribli.languages.export20);
        stack.button({
            id: "pandocBinChooser",
            label: window.scribli.languages.config,
            icon: "iconSettings",
        });
    });
    group.textBlock("export.pandocParams", {
        title: window.scribli.languages.export25,
        desc: window.scribli.languages.export26,
        mode: "textarea",
    });
};

const mountExportPandocStack = (root: HTMLElement) => {
    root.querySelector("#pandocBinReset")?.addEventListener("click", () => {
        exportConfigApi.patch("export.pandocBin", "");
    });
    root.querySelector("#pandocBinPathDisplay")?.addEventListener("click", () => {
        if (window.scribli.config.export.pandocBin) {
            useShell("showItemInFolder", window.scribli.config.export.pandocBin);
        }
    });
    root.querySelector("#pandocBinChooser")?.addEventListener("click", async () => {
        const localPath = await ipcRenderer.invoke(Constants.SCRIBLI_GET, {
            cmd: "showOpenDialog",
            defaultPath: window.scribli.config.system.homeDir,
            properties: ["openFile", "showHiddenFiles"],
        });
        if (!localPath.filePaths.length) {
            return;
        }
        exportConfigApi.patch("export.pandocBin", localPath.filePaths[0]);
    });
};
/// #endif

export const registerExportTab = (tab: SettingTabBuilder) => {
    registerExportReferencesGroup(tab);
    registerExportFormatGroup(tab);
    registerExportPdfGroup(tab);
    registerExportImagesGroup(tab);
    /// #if !BROWSER
    registerExportPandocGroup(tab);
    /// #endif
};
