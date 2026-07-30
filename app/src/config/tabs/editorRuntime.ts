import {Constants} from "../../constants";
import {getAllEditor} from "../../layout/getAll";
import {setInlineStyle} from "../../util/assets";
import {reloadProtyle} from "../../protyle/util/reload";
import {resize} from "../../protyle/util/resize";
import {createConfigNamespaceApi} from "../util/namespaceApi";

const applyEditorConfig = (data: Config.IEditor) => {
    window.scribli.config.editor = data;
    getAllEditor().forEach((editorItem) => {
        const protyle = editorItem.protyle;
        reloadProtyle(protyle, false);
        let isFullWidth = protyle.wysiwyg.element.getAttribute(Constants.CUSTOM_SY_FULLWIDTH);
        if (!isFullWidth) {
            isFullWidth = window.scribli.config.editor.fullWidth ? "true" : "false";
        }
        if (isFullWidth === "true" && protyle.contentElement.getAttribute("data-fullwidth") === "true") {
            return;
        }
        resize(protyle);
        if (isFullWidth === "true") {
            protyle.contentElement.setAttribute("data-fullwidth", "true");
        } else {
            protyle.contentElement.removeAttribute("data-fullwidth");
        }
    });

    void setInlineStyle();
};

export const editorConfigApi = createConfigNamespaceApi<Config.IEditor>({
    namespace: "editor",
    getConfig: () => window.scribli.config.editor,
    setConfig: applyEditorConfig,
    apiPath: "/api/setting/setEditor",
});
