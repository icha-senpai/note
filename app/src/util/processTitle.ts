import {escapeHtml} from "./escape";
import {Constants} from "../constants";
import {pathPosix} from "./pathName";

export const getWorkspaceName = () => {
    const dir = window.scribli.config.system.workspaceDir;
    // 
    return dir ? pathPosix().basename(dir.replace(/\\/g, "/")) : window.scribli.languages?.workspace;
};

export const setTitle = (title: string, showVersionTitle = false) => {
    const dragElement = document.getElementById("drag");
    const workspaceName = getWorkspaceName();
    if (showVersionTitle) {
        const versionTitle = `${workspaceName} - ${window.scribli.languages.scribliNote} v${Constants.SCRIBLI_VERSION}`;
        document.title = versionTitle;
        if (!window.scribli.config.appearance.hideToolbar && dragElement) {
            dragElement.textContent = versionTitle;
            dragElement.setAttribute("title", versionTitle);
        }
    } else {
        title = title.trim() || window.scribli.languages["_kernel"][16];
        document.title = `${title} - ${workspaceName} - ${window.scribli.languages.scribliNote} v${Constants.SCRIBLI_VERSION}`;
        if (!window.scribli.config.appearance.hideToolbar && dragElement) {
            dragElement.setAttribute("title", title);
            dragElement.innerHTML = escapeHtml(title);
        }
    }
};
