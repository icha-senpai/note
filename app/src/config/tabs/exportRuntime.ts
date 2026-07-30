import {createConfigNamespaceApi} from "../util/namespaceApi";

const applyExportConfig = (data: Config.IExport) => {
    window.scribli.config.export = data;
    const pathDisplay = document.getElementById("pandocBinPathDisplay");
    if (pathDisplay) {
        pathDisplay.textContent = data.pandocBin;
    }
};

export const exportConfigApi = createConfigNamespaceApi<Config.IExport>({
    namespace: "export",
    getConfig: () => window.scribli.config.export,
    setConfig: applyExportConfig,
    apiPath: "/api/setting/setExport",
});
