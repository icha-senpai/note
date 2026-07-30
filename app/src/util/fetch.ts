import {Constants} from "../constants";
/// #if !BROWSER
import {ipcRenderer} from "electron";
/// #endif
import {processMessage} from "./processMessage";
import {kernelError} from "./kernelFault";

const getFetchErrorResponse = (response: Response): IWebSocketData => ({
    data: null,
    msg: response.statusText,
    code: -response.status,
});

export const fetchPost = (
    url: string,
    data?: any,
    cb?: (response: IWebSocketData) => void,
    headers?: Record<string, string>,
    failCallback?: (response: IWebSocketData) => void,
    signal?: AbortSignal) => {
    const init: RequestInit = {
        method: "POST",
    };
    if (data) {
        if (["/api/search/searchRefBlock", "/api/graph/getGraph", "/api/graph/getLocalGraph",
            "/api/block/getRecentUpdatedBlocks", "/api/search/fullTextSearchBlock"].includes(url)) {
            window.scribli.reqIds[url] = Date.now();
            if (data.type === "local" && url === "/api/graph/getLocalGraph") {
                // Intentionally empty.
            } else {
                data.reqId = window.scribli.reqIds[url];
            }
        }
        if (url === "/api/transactions") {
            data.reqId = Date.now();
        }
        if (data instanceof FormData) {
            init.body = data;
        } else {
            init.body = JSON.stringify(data);
        }
    }
    if (headers) {
        init.headers = headers;
    }
    if (signal) {
        init.signal = signal;
    }
    let isGetFile202 = false;
    return fetch(url, init).then((response) => {
        switch (response.status) {
            case 403:
            case 404:
                return getFetchErrorResponse(response);
            case 401:
                setTimeout(() => {
                    window.location.reload();
                }, 3000);
                return getFetchErrorResponse(response);
            default:
                if (response.status === 202 && url === "/api/file/getFile") {
                    isGetFile202 = true;
                }
                if (response.headers.get("content-type")?.indexOf("application/json") > -1) {
                    return response.json();
                } else {
                    return response.text();
                }
        }
    }).then((response: IWebSocketData) => {
        if (failCallback && url === "/api/file/getFile" && isGetFile202) {
            failCallback(response);
            return;
        }
        if (typeof response === "string") {
            if (cb) {
                cb(response);
            }
            return;
        }
        if (["/api/search/searchRefBlock", "/api/graph/getGraph", "/api/graph/getLocalGraph",
            "/api/block/getRecentUpdatedBlocks", "/api/search/fullTextSearchBlock"].includes(url)) {
            if (response.data.reqId && window.scribli.reqIds[url] && window.scribli.reqIds[url] > response.data.reqId) {
                return;
            }
        }
        if (typeof response === "object" && typeof response.msg === "string" && typeof response.code === "number") {
            if (processMessage(response) && cb) {
                cb(response);
            }
        } else if (cb) {
            cb(response);
        }
    }).catch((e) => {
        if (e?.name === "AbortError") {
            return;
        }
        if (failCallback && url === "/api/file/getFile") {
            failCallback({
                data: null,
                msg: e.message,
                code: 400,
            });
            return;
        }
        console.warn("fetch post failed [" + e + "], url [" + url + "]");
        if (url === "/api/transactions" && (e.message === "Failed to fetch" || e.message === "Unexpected end of JSON input")) {
            kernelError();
            return;
        }
        /// #if !BROWSER
        if (url === "/api/system/exit" || url === "/api/system/setWorkspaceDir" || (
            ["/api/system/setUILayout"].includes(url) && data.errorExit
        )) {
            ipcRenderer.send(Constants.SCRIBLI_QUIT, location.port);
        }
        /// #endif
    });
};

export const fetchSyncPost = async (url: string, data?: any, headers?: Record<string, string>) => {
    const init: RequestInit = {
        method: "POST",
    };
    if (headers) {
        init.headers = headers;
    }
    if (data) {
        if (data instanceof FormData) {
            init.body = data;
        } else {
            init.body = JSON.stringify(data);
        }
    }
    const res = await fetch(url, init);
    const res2 = await res.json() as IWebSocketData;
    processMessage(res2);
    return res2;
};

export const fetchGet = <T extends IWebSocketData | IObject | string = IWebSocketData | IObject | string>(url: string, cb: (response: T) => void) => {
    fetch(url).then((response) => {
        if (response.headers.get("content-type")?.indexOf("application/json") > -1) {
            return response.json();
        } else {
            return response.text();
        }
    }).then((response) => {
        cb(response as T);
    });
};
