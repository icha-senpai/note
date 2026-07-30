import {createConfigNamespaceApi} from "../util/namespaceApi";

export const flashcardConfigApi = createConfigNamespaceApi<Config.IFlashCard>({
    namespace: "flashcard",
    getConfig: () => window.scribli.config.flashcard,
    setConfig: (data) => {
        window.scribli.config.flashcard = data;
    },
    apiPath: "/api/setting/setFlashcard",
});
