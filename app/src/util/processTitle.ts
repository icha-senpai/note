import {escapeHtml} from "./escape";
import {Constants} from "../constants";
import {pathPosix} from "./pathName";

export const getWorkspaceName = () => {
    const dir = window.scribli.config.system.workspaceDir;
    // 浏览器环境下内核不返回工作空间绝对路径，回退到“工作空间”（Workspace）。
    // 不能用 scribliNote：setTitle 的标题模板本身已含 scribliNote，会导致“Scribli - Scribli”重复。
    // 注意：该函数可能在 languages 加载前（如 setBodyHighlight）被调用，故用可选链，
    // 此时返回 undefined，由调用方（setBodyHighlight 的 if(!name) return）跳过处理
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
