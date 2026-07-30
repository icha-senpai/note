type TPluginDockPosition = "LeftTop" | "LeftBottom" | "RightTop" | "RightBottom" | "BottomLeft" | "BottomRight"
type TDockPosition = "Left" | "Right" | "Bottom"
type TWS = "main" | "filetree" | "protyle" | "backlink" | "bookmark" | "graph" | "outline" | "tag" | "agentChat"
type TDock = "file" | "outline" | "inbox" | "bookmark" | "tag" | "graph" | "globalGraph" | "backlink" | "agentChat"
type TTab = "Outline" | "Graph" | "Backlink" | "Asset" | "Editor" | "Search" | "scribli-card"
type TOperation =
    "insert"
    | "restoreCreatedDoc"
    | "removeCreatedDoc"
    | "update"
    | "delete"
    | "move"
    | "foldHeading"
    | "unfoldHeading"
    | "setAttrs"
    | "updateAttrs"
    | "append"
    | "insertAttrViewBlock"
    | "removeAttrViewBlock"
    | "addAttrViewCol"
    | "removeAttrViewCol"
    | "addFlashcards"
    | "removeFlashcards"
    | "updateAttrViewCell"
    | "updateAttrViewCol"
    | "updateAttrViewColTemplate"
    | "sortAttrViewRow"
    | "sortAttrViewCol"
    | "sortAttrViewKey"
    | "setAttrViewColPin"
    | "setAttrViewColHidden"
    | "setAttrViewColWrap"
    | "setAttrViewColWidth"
    | "setAttrViewColAlign"
    | "updateAttrViewColOptions"
    | "removeAttrViewColOption"
    | "updateAttrViewColOption"
    | "setAttrViewName"
    | "setAttrViewNewItemTemplates"
    | "doUpdateUpdated"
    | "duplicateAttrViewKey"
    | "setAttrViewColIcon"
    | "setAttrViewFilters"
    | "setAttrViewSorts"
    | "setAttrViewColCalc"
    | "updateAttrViewColNumberFormat"
    | "replaceAttrViewBlock"
    | "addAttrViewView"
    | "setAttrViewViewName"
    | "removeAttrViewView"
    | "setAttrViewViewIcon"
    | "duplicateAttrViewView"
    | "duplicateAttrViewRow"
    | "sortAttrViewView"
    | "setAttrViewPageSize"
    | "updateAttrViewColRelation"
    | "moveOutlineHeading"
    | "updateAttrViewColRollup"
    | "hideAttrViewName"
    | "setAttrViewCardSize"
    | "setAttrViewCardAspectRatio"
    | "setAttrViewCoverFrom"
    | "setAttrViewCoverFromAssetKeyID"
    | "setAttrViewFitImage"
    | "setAttrViewShowIcon"
    | "setAttrViewWrapField"
    | "setAttrViewColDateFillCreated"
    | "setAttrViewColDateFillSpecificTime"
    | "setAttrViewViewDesc"
    | "setAttrViewColDesc"
    | "setAttrViewBlockView"
    | "setAttrViewGroup"
    | "removeAttrViewGroup"
    | "hideAttrViewAllGroups"
    | "syncAttrViewTableColWidth"
    | "hideAttrViewGroup"
    | "sortAttrViewGroup"
    | "foldAttrViewGroup"
    | "setAttrViewDisplayFieldName"
    | "setAttrViewFillColBackgroundColor"
    | "setAttrViewUpdatedIncludeTime"
    | "setAttrViewCreatedIncludeTime"
type TCardType = "doc" | "notebook" | "all"
type TEventBus = "ws-main" | "sync-start" | "sync-end" | "sync-fail" |
    "click-blockicon" | "click-editorcontent" | "click-pdf" | "click-editortitleicon" | "click-flashcard-action" |
    "open-noneditableblock" |
    "open-menu-blockref" | "open-menu-fileannotationref" | "open-menu-tag" | "open-menu-link" | "open-menu-image" |
    "open-menu-av" | "open-menu-content" | "open-menu-breadcrumbmore" | "open-menu-doctree" | "open-menu-inbox" |
    "open-scribli-url-plugin" | "open-scribli-url-block" | "opened-notebook" |
    "closed-notebook" |
    "paste" |
    "input-search" |
    "loaded-protyle-dynamic" | "loaded-protyle-static" |
    "switch-protyle" | "switch-protyle-mode" |
    "destroy-protyle" |
    "lock-screen" |
    "code-language-update" | "code-language-change" |
    "kernel-plugin-state-change"
type TAVView = "table" | "gallery" | "kanban"
type TAVAlign = "" | "left" | "center" | "right"
type TAVCol =
    "text"
    | "date"
    | "number"
    | "relation"
    | "rollup"
    | "select"
    | "block"
    | "mSelect"
    | "url"
    | "email"
    | "phone"
    | "mAsset"
    | "template"
    | "created"
    | "updated"
    | "checkbox"
    | "lineNumber"
type TAVFilterOperator =
    "="
    | "!="
    | ">"
    | ">="
    | "<"
    | "<="
    | "Contains"
    | "Does not contains"
    | "Is empty"
    | "Is not empty"
    | "Starts with"
    | "Ends with"
    | "Is between"
    | "Is relative to today"
    | "Is true"
    | "Is false"

type TRecentDocsSort = "viewedAt" | "closedAt" | "openAt" | "updated"
type TPublishAccessLevel = "public" | "protected" | "hidden" | "private" | "forbidden";

/**
 */
type TKernelPluginState = -1 | 0 | 1 | 2 | 3 | 4 | 5

type TJsonRpcId = string | number;
type TJsonRpcMethod = string;
type TJsonRpcPositionalParams = any[];
type TJsonRpcNamedParams = Record<string, any>;
type TJsonRpcParams = TJsonRpcPositionalParams | TJsonRpcNamedParams | undefined;
type TJsonRpcMethodParams = TJsonRpcPositionalParams | [TJsonRpcNamedParams] | [];
type TJsonRpcHandler<T = any> = (...args: TJsonRpcMethodParams) => Promise<T> | T;

declare module "blueimp-md5"

declare class Highlight {
    constructor(...range: Range[]);

    add(range: Range): void

    clear(): void

    forEach(callbackfn: (value: Range, key: number) => void): void;
}

declare namespace CSS {
    const highlights: Map<string, Highlight>;
}

interface CSSStyleDeclarationElectron extends CSSStyleDeclaration {
    WebkitAppRegion: string;
}

declare module "*.scss";

interface Window {
    DOMPurify: {
        sanitize(dirty: string, options?: any): string;
    };
    echarts: {
        init(element: Element, theme?: string, options?: {
            width: number
        }): {
            setOption(option: any): void;
            getZr(): any;
            on(name: string, event: (e: any) => void): any;
            containPixel(name: string, position: number[]): any;
            resize(): void;
        };
        dispose(element: Element): void;
        getInstanceById(id: string): {
            resize: () => void
            clear: () => void
            getOption: () => { series: { type: string }[] }
        };
    };
    ABCJS: {
        renderAbc(element: Element, text: string, options: {
            responsive: string
        }): void;
    };
    MathJax: {
        svg: {
            fontCache: string
        }
        startup?: {
            promise: Promise<void>
        }
        tex2svg?(math: string, options: { display: boolean }): HTMLElement
    };
    hljs: {
        listLanguages(): string[];
        highlight(text: string, options: {
            language?: string,
            ignoreIllegals: boolean
        }): {
            value: string
        };
        getLanguage(text: string): {
            name: string
        };
    };
    katex: {
        renderToString(math: string, option: {
            displayMode: boolean;
            output: string;
            macros: IObject;
            trust: boolean;
            strict: (errorCode: string) => "ignore" | "warn";
        }): string;
    };
    zenuml: object,
    mermaid: {
        initialize(options: any): void,
        render(id: string, text: string): { svg: string },
        registerExternalDiagrams(ex: object[]): void,
        registerIconPacks(options: {
            name: string,
            loader(): Promise<Response>
        }[]): void
    };
    plantumlEncoder: {
        encode(options: string): string,
    };
    pdfjsLib: any;
    htmlToImage: {
        toCanvas: (element: Element, options?: IHtmlToImageOptions) => Promise<HTMLCanvasElement>
        toBlob: (element: Element, options?: IHtmlToImageOptions) => Promise<Blob>
    };
    scribli: IScribli;

    Protyle: import("../protyle/method").default;

    lockscreenByMode(): void;

    goBack(): void;

    showMessage(message: string, timeout: number, type: string, messageId?: string): void;

    reconnectWebSocket(): void;

    openFileByURL(URL: string): boolean;

    destroyTheme(): Promise<void>;
}

interface ILocalFiles {
    path: string,
    size: number
}

interface IClipboardData {
    textHTML?: string,
    textPlain?: string,
    scribliHTML?: string,
    files?: File[],
    localFiles?: ILocalFiles[],
}

interface IRefDefs {
    refID: string,
    defIDs?: string[],
    avItemID?: string,
    avViewID?: string,
    avGroupID?: string,
}

interface IFilesPath {
    notebookId: string,
    openPaths: string[]
}

interface IPosition {
    x: number,
    y: number,
    w?: number,
    h?: number,
    isLeft?: boolean
}

interface ISaveLayout {
    name: string,
    layout: IObject
    time: number
    filesPaths: IFilesPath[]
}

interface IWorkspace {
    path: string;
    closed: boolean;
}

interface ICardPackage {
    id: string;
    updated: string;
    name: string;
    size: number;
}

interface ICard {
    deckID: string;
    cardID: string;
    blockID: string;
    nextDues: Record<string, string>;
    lapses: number;
    lastReview: number;
    reps: number;
    state: number;
}

interface ICardData {
    cards: ICard[],
    unreviewedCount: number
    unreviewedNewCardCount: number
    unreviewedOldCardCount: number
}

interface IPluginSettingOption {
    title: string;
    description?: string;
    actionElement?: HTMLElement;
    direction?: "column" | "row";

    createActionElement?(): HTMLElement;
}

interface ISearchAssetOption {
    keys: string[],
    col: string,
    row: string,
    layout: number,
    method: number,
    types: {
        ".txt": boolean,
        ".md": boolean,
        ".docx": boolean,
        ".xlsx": boolean,
        ".pptx": boolean,
    },
    sort: number,
    k: string,
}

interface ITextOption {
    color?: string,
    type: string
}

interface ISnippet {
    id?: string;
    name: string;
    type: string;
    enabled: boolean;
    content: string;
    disabledInPublish: boolean;
}

interface IInbox {
    oId: string;
    shorthandContent: string;
    shorthandMd: string;
    shorthandDesc: string;
    shorthandFrom: number;
    shorthandTitle: string;
    shorthandURL: string;
    hCreated: string;
}

interface IPdfAnno {
    pages?: {
        index: number
        positions: number[]
    }[]
    index?: number,
    color: string,
    type: string,   // border, text
    content: string,    // rect, text
    mode: string,
    id?: string,
    coords?: number[]
    ids?: string[]
}

interface IBackStack {
    id: string,
    data?: {
        startId: string,
        endId: string
        path: string
        notebookId: string
    },
    scrollTop?: number,
    callback?: TProtyleAction[],
    position?: {
        start: number,
        end: number
    }
    protyle?: IProtyle,
    zoomId?: string
}

interface IEmojiItem {
    unicode: string,
    description: string,
    keywords: string
}

interface IEmoji {
    id: string,
    title: string,
    items: IEmojiItem[]
}

interface INotebook {
    name: string;
    id: string;
    closed: boolean;
    icon: string;
    sort: number;
    subFileCount: number;
    dueFlashcardCount?: string;
    newFlashcardCount?: string;
    flashcardCount?: string;
    sortMode: number;
    encrypted?: boolean;
}

interface IScribli {
    zIndex: number
    storage?: {
        [key: string]: any
    },
    closedTabs?: ILayoutJSON[]
    reqIds: {
        [key: string]: number
    },
    editorIsFullscreen?: boolean,
    hideBreadcrumb?: boolean,
    notebooks?: INotebook[],
    emojis?: IEmoji[],
    backStack?: IBackStack[],
    dragElement?: HTMLElement,
    dragTitle?: string,
    currentDragOverTabHeadersElement?: HTMLElement
    touchDragActive?: boolean,
    touchDragGhost?: HTMLElement | null,
    layout?: {
        layout?: import("../layout").Layout,
        centerLayout?: import("../layout").Layout,
        leftDock?: import("../layout/dock").Dock,
        rightDock?: import("../layout/dock").Dock,
        bottomDock?: import("../layout/dock").Dock,
    }
    config?: Config.IConf;
    ws: import("../layout/Model").Model,
    ctrlIsPressed?: boolean,
    altIsPressed?: boolean,
    shiftIsPressed?: boolean,
    coordinates?: {
        pageX: number,
        pageY: number,
        clientX: number,
        clientY: number,
        screenX: number,
        screenY: number,
    },
    menus?: import("../menus").Menus
    languages?: {
        [key: string]: any;
    }
    bookmarkLabel?: string[]
    blockPanels: import("../block/Panel").BlockPanel[],
    dialogs: import("../dialog").Dialog[],
    viewer?: Viewer,
    /**
     */
    isPublish?: boolean;
}

interface IOperation {
    action: TOperation,
    id?: string,
    context?: Record<string, string>,  // focusId, message, ignoreProcess, setRange
    blockID?: string,
    isTwoWay?: boolean,
    backRelationKeyID?: string,
    avID?: string,  // av
    format?: string
    keyID?: string
    rowID?: string
    data?: any,
    parentID?: string
    previousID?: string
    retData?: any
    nextID?: string
    isDetached?: boolean
    srcIDs?: string[]
    srcs?: IOperationSrcs[]
    ignoreDefaultFill?: boolean
    viewID?: string
    name?: string
    type?: TAVCol
    deckID?: string
    blockIDs?: string[]
    removeDest?: boolean
    layout?: string
    groupID?: string
    targetGroupID?: string
}

interface IOperationSrcs {
    itemID: string,
    id: string,
    content?: string,
    isDetached: boolean
}

interface IObject {
    [key: string]: string | number | boolean;
}

interface IHtmlToImageOptions {
    [key: string]: unknown;
    imagePlaceholder?: string;
    onImageErrorHandler?: (event: Event) => void;
}

interface ILayoutJSON extends ILayoutOptions {
    scrollAttr?: IScrollAttr,
    instance?: string,
    width?: string,
    height?: string,
    title?: string,
    lang?: string
    docIcon?: string
    page?: string
    path?: string
    blockId?: string
    mode?: TEditorMode
    action?: TProtyleAction
    icon?: string
    rootId?: string
    databaseRowId?: string
    active?: boolean
    pin?: boolean
    isPreview?: boolean
    customModelData?: any
    customModelType?: string
    config?: Config.IUILayoutTabSearchConfig
    children?: ILayoutJSON[] | ILayoutJSON
}

interface ICommand {
    langKey: string,
    langText?: string,
    hotkey?: string,
    customHotkey?: string,
    callback?: () => void
    globalCallback?: () => void
    fileTreeCallback?: (file: import("../layout/dock/Files").Files) => void
    editorCallback?: (protyle: IProtyle) => void
    dockCallback?: (element: HTMLElement) => void
}

interface IPluginData {
    displayName: string,
    name: string,
    js: string,
    css: string,
    i18n: Record<string, string>
}

interface IPluginDockTab {
    position: TPluginDockPosition,
    size: Config.IUILayoutDockPanelSize,
    icon: string,
    hotkey?: string,
    title: string,
    index?: number
    show?: boolean
}

interface IExportOptions {
    type: string,
    id: string,
}

interface IOpenFileOptions {
    app: import("../index").App,
    searchData?: Config.IUILayoutTabSearchConfig,
    custom?: {
        title: string,
        icon: string,
        data?: any
        id: string,
        fn?: (options: {
            tab: import("../layout/Tab").Tab,
            data: any,
        }) => import("../layout/Model").Model,
    }
    scrollPosition?: ScrollLogicalPosition,
    assetPath?: string,
    fileName?: string,
    rootTitleEmpty?: boolean,
    rootIcon?: string,
    id?: string,
    rootID?: string,
    position?: string,
    page?: number | string, // asset
    mode?: TEditorMode // file
    action?: TProtyleAction[]
    keepCursor?: boolean
    zoomIn?: boolean
    removeCurrentTab?: boolean
    openNewTab?: boolean
    afterOpen?: (model?: import("../layout/Model").Model) => void
}

interface ILayoutOptions {
    direction?: Config.TUILayoutDirection;
    size?: string;
    resize?: Config.TUILayoutDirection;
    type?: Config.TUILayoutType;
    element?: HTMLElement;
}

interface ITab {
    icon?: string;
    docIcon?: string;
    title?: string;
    panel?: string;
    callback?: (tab: import("../layout/Tab").Tab) => void;
}

interface IWebSocketData {
    cmd?: string;
    callback?: string;
    data?: any;
    msg: string;
    code: number;
    sid?: string;
    context?: any;
}

interface IGraphCommon {
    d3: {
        centerStrength: number
        collideRadius: number
        collideStrength: number
        lineOpacity: number
        linkDistance: number
        linkWidth: number
        nodeSize: number
        arrow: boolean
    };
    type: {
        blockquote: boolean
        callout: boolean
        code: boolean
        heading: boolean
        list: boolean
        listItem: boolean
        math: boolean
        paragraph: boolean
        super: boolean
        table: boolean
        tag: boolean
    };
}

interface IKeymapItem {
    default: string,
    custom: string
}

interface IFile {
    icon: string;
    name1: string;
    alias: string;
    memo: string;
    bookmark: string;
    path: string;
    name: string;
    titleEmpty?: boolean;
    hMtime: string;
    hCtime: string;
    hSize: string;
    dueFlashcardCount?: string;
    newFlashcardCount?: string;
    flashcardCount?: string;
    id: string;
    count: number;
    subFileCount: number;
}

interface IBlockTree {
    box: string,
    nodeType: string,
    hPath: string,
    subType: string,
    name: string,
    type: string,
    depth: number,
    url?: string,
    label?: string,
    id?: string,
    blocks?: IBlock[],
    count: number,
    children?: IBlockTree[]
}

interface IBlock {
    riffCard?: IRiffCard,
    depth?: number,
    box?: string;
    path?: string;
    hPath?: string;
    id?: string;
    rootID?: string;
    type?: string;
    content?: string;
    def?: IBlock;
    defID?: string
    defPath?: string
    refText?: string;
    name?: string;
    memo?: string;
    alias?: string;
    tag?: string;
    refs?: IBlock[];
    children?: IBlock[]
    length?: number
    ial: Record<string, string>
    refCount?: number
}

interface IRiffCard {
    due?: string;
    reps?: number;
}

interface IModels {
    editor: import("../editor").Editor[],
    graph: import("../layout/dock/Graph").Graph[],
    outline: import("../layout/dock/Outline").Outline[]
    backlink: import("../layout/dock/Backlink").Backlink[]
    inbox: import("../layout/dock/Inbox").Inbox[]
    files: import("../layout/dock/Files").Files[]
    bookmark: import("../layout/dock/Bookmark").Bookmark[]
    tag: import("../layout/dock/Tag").Tag[]
    asset: import("../asset").Asset[]
    search: import("../search").Search[]
    custom: import("../layout/dock/Custom").Custom[]
}

interface IMenu {
    checked?: boolean,
    iconClass?: string,
    label?: string,
    click?: (element: HTMLElement, event: MouseEvent) => boolean | void | Promise<boolean | void>
    type?: "separator" | "submenu" | "readonly" | "empty",
    accelerator?: string,
    action?: string,
    id?: string,
    submenu?: IMenu[]
    disabled?: boolean
    icon?: string
    iconHTML?: string
    current?: boolean
    bind?: (element: HTMLElement) => void
    index?: number
    element?: HTMLElement
    ignore?: boolean
    warning?: boolean
}

interface IAV {
    id: string;
    name: string;
    view: IAVTable | IAVGallery;
    viewID: string;
    viewType: TAVView;
    views: IAVView[];
    isMirror?: boolean;
    newItemTemplates?: IAVNewItemTemplate[];
    defaultTemplateID?: string;
    target?: IAVRenderTarget;
}

interface IAVRenderTarget {
    status: "visible" | "filtered" | "itemNotFound" | "viewNotFound" | "groupHidden";
    itemID: string;
    groupID?: string;
    index: number;
    offset: number;
    pageSize: number;
}

type TAVNewItemTarget = "detached" | "document";
type TAVNewItemFieldValueMode = "static" | "currentTime";

interface IAVNewItemSaveLocation {
    boxID?: string;
    pathTemplate: string;
}

interface IAVNewItemFieldValue {
    mode: TAVNewItemFieldValueMode;
    value?: IAVCellValue;
}

interface IAVNewItemTemplate {
    id: string;
    name: string;
    icon?: string;
    targetType: TAVNewItemTarget;
    primaryKeyTemplate?: string;
    fieldValues?: Record<string, IAVNewItemFieldValue>;
    saveLocation?: IAVNewItemSaveLocation;
    contentTemplatePath?: string;
}

interface IAVView {
    name: string;
    desc: string;
    id: string;
    type: TAVView;
    icon: string;
    hideAttrViewName: boolean;
    pageSize: number;
    showIcon: boolean;
    wrapField: boolean;
    groupHidden?: number,
    groupFolded?: boolean,
    filters: IAVFilter[],
    sorts: IAVSort[],
    groups: IAVView[]
    group: IAVGroup
    groupKey: IAVColumn
    groupValue: IAVCellValue
}

interface IAVTable extends IAVView {
    columns: IAVColumn[],
    rows: IAVRow[],
    rowCount: number,
}

interface IAVVirtualData {
    renderedStart: number;
    renderedEnd: number;
    topSpacerHeight: number;
    rowOffset?: number;
    locate?: boolean;
}

interface IAVGallery extends IAVView {
    coverFrom: number;
    coverFromAssetKeyID?: string;
    cardSize: number;
    cardAspectRatio: number;
    displayFieldName: boolean;
    fitImage: boolean;
    cards: IAVGalleryItem[],
    desc: string
    fields: IAVColumn[]
    cardCount: number,
}

interface IAVKanban extends IAVView {
    coverFrom: number;
    coverFromAssetKeyID?: string;
    cardSize: number;
    cardAspectRatio: number;
    displayFieldName: boolean;
    fitImage: boolean;
    cards: IAVGalleryItem[],
    desc: string
    fields: IAVColumn[]
    cardCount: number,
    fillColBackgroundColor: boolean
}

interface IAVFilter {
    column?: string,
    operator?: TAVFilterOperator,
    quantifier?: string,
    value?: IAVCellValue,
    relativeDate?: IAVRelativeDate,
    relativeDate2?: IAVRelativeDate,
    combination?: "and" | "or",
    filters?: IAVFilter[],
}

interface IAVRelativeDate {
    count: number;
    unit: number;
    direction: number;
}

interface IAVGroup {
    field: string,
    method?: number
    range?: {
        numStart: number
        numEnd: number
        numStep: number
    }
    hideEmpty?: boolean
    order?: number
}

interface IAVSort {
    column: string,
    order: "ASC" | "DESC"
}

interface IAVColumn {
    width: string,
    align: TAVAlign,
    icon: string,
    id: string,
    name: string,
    desc: string,
    wrap: boolean,
    pin: boolean,
    hidden: boolean,
    type: TAVCol,
    numberFormat: string,
    template: string,
    calc: IAVCalc,
    updated?: {
        includeTime: boolean
    }
    created?: {
        includeTime: boolean
    }
    date?: {
        autoFillNow: boolean,
        fillSpecificTime: boolean,
    }
    options?: {
        name: string,
        color: string,
        desc?: string,
    }[],
    relation?: IAVColumnRelation,
    rollup?: IAVCellRollupValue
}

interface IAVRow {
    id: string,
    cells: IAVCell[]
}

interface IAVGalleryItem {
    coverURL?: string;
    coverContent?: string;
    id: string;
    values: IAVCell[];
}

interface IAVCell {
    id: string,
    color: string,
    bgColor: string,
    value: IAVCellValue,
    valueType: TAVCol,
}

interface IAVCellValue {
    keyID?: string,
    id?: string,
    blockID?: string
    type: TAVCol,
    isDetached?: boolean,
    text?: {
        content: string
    },
    number?: {
        content?: number,
        isNotEmpty: boolean,
        format?: string,
        formattedContent?: string
    },
    mSelect?: IAVCellSelectValue[]
    mAsset?: IAVCellAssetValue[]
    block?: {
        content: string,
        id?: string,
        icon?: string
    }
    url?: {
        content: string
    }
    phone?: {
        content: string
    }
    email?: {
        content: string
    }
    template?: {
        content: string
    },
    checkbox?: {
        checked: boolean,
        content?: string,
    }
    relation?: IAVCellRelationValue
    rollup?: {
        contents?: IAVCellValue[]
    }
    date?: IAVCellDateValue
    created?: IAVCellDateValue
    updated?: IAVCellDateValue
}

interface IAVCellRelationValue {
    blockIDs: string[];
    contents?: IAVCellValue[];
}

interface IAVCellDateValue {
    content?: number,
    isNotEmpty?: boolean
    content2?: number,
    isNotEmpty2?: boolean
    hasEndDate?: boolean
    formattedContent?: string,
    isNotTime?: boolean
}

interface IAVCellSelectValue {
    content: string,
    color: string
}

interface IAVCellAssetValue {
    content: string,
    name: string,
    type: "file" | "image"
}

interface IAVColumnRelation {
    avID?: string;
    backKeyID?: string;
    isTwoWay?: boolean;
}

interface IAVCellRollupValue {
    relationKeyID?: string;
    keyID?: string;
    calc?: IAVCalc;
}

interface IAVCalc {
    operator?: string,
    template?: string,
    result?: IAVCellValue
}

interface IPublishAccessItem {
    id: string,
    visible: boolean,
    password: string,
    disable: boolean
    iconHTML?: string
}

interface IKernelPlugin {
    /**
     */
    state: IKernelPluginState;

    /**
     */
    rpc: IKernelPluginRpc;
}

interface IKernelPluginState {
    /**
     */
    code: TKernelPluginState;

    /**
     */
    description: string;
}

interface IKernelPluginRpcCall {
    /**
     */
    method: TJsonRpcMethod;

    /**
     *
     */
    id?: TJsonRpcId;

    /**
     */
    params?: any[] | Record<string, any>;

    /**
     * @defaultValue false
     */
    notification?: boolean;
}

interface IKernelPluginRpcRequest extends IKernelPluginRpcCall {
    jsonrpc: "2.0";
}

interface IKernelPluginRpcBaseResponse {
    jsonrpc: "2.0";
}

interface IKernelPluginRpcResultResponse extends IKernelPluginRpcBaseResponse {
    id: TJsonRpcId;
    result?: any;
}

interface IKernelPluginRpcErrorResponse extends IKernelPluginRpcBaseResponse {
    id: TJsonRpcId | null;
    error?: any;
}

interface IKernelPluginRpcError {
    code: number;
    message: string;
    data?: any;
}

interface IKernelPluginRpc {
    /**
     */
    call: Record<TJsonRpcMethod, (...args: TJsonRpcMethodParams) => Promise<any>>;

    /**
     */
    notify: Record<TJsonRpcMethod, (...args: TJsonRpcMethodParams) => void>;

    /**
     */
    batch: (...calls: IKernelPluginRpcCall[]) => Promise<IKernelPluginRpcError | (IKernelPluginRpcResultResponse | IKernelPluginRpcErrorResponse)[]>;

    /**
     */
    bind: (method: TJsonRpcMethod, handler: TJsonRpcHandler<void>) => void;

    /**
     */
    unbind: (method: TJsonRpcMethod, handler: TJsonRpcHandler<void>) => void;
}

/**
 */
interface IScribliUriBlockInfo {
    /**
     */
    id: string;

    /**
     * 
     * @defaultValue false
     */
    focus: boolean;

    /**
     * 
     * @defaultValue false
     */
    fullscreen: boolean;
    avItemID?: string;
    avViewID?: string;
    avGroupID?: string;
}
