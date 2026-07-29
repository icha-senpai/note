
export const getDefaultType = () => {
    return {
        audioBlock: window.scribli.config.search.audioBlock,
        videoBlock: window.scribli.config.search.videoBlock,
        iframeBlock: window.scribli.config.search.iframeBlock,
        widgetBlock: window.scribli.config.search.widgetBlock,
        document: window.scribli.config.search.document,
        heading: window.scribli.config.search.heading,
        list: window.scribli.config.search.list,
        listItem: window.scribli.config.search.listItem,
        codeBlock: window.scribli.config.search.codeBlock,
        htmlBlock: window.scribli.config.search.htmlBlock,
        mathBlock: window.scribli.config.search.mathBlock,
        table: window.scribli.config.search.table,
        blockquote: window.scribli.config.search.blockquote,
        callout: window.scribli.config.search.callout,
        superBlock: window.scribli.config.search.superBlock,
        paragraph: window.scribli.config.search.paragraph,
        embedBlock: window.scribli.config.search.embedBlock,
        databaseBlock: window.scribli.config.search.databaseBlock,
    };
};

export const getDefaultSubType = (): Config.IUILayoutTabSearchConfigSubTypes => {
    return {
        h1: false, h2: false, h3: false, h4: false, h5: false, h6: false,
        o: false, u: false, t: false,
    };
};
