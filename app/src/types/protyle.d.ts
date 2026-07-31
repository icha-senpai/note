interface ILuteNode {
    TokensStr: () => string;
    __internal_object__: {
        Parent: {
            Type: number,
        },
        HeadingLevel: string,
    };
}

type THintSource = "search" | "av" | "hint";

type TTurnIntoOne =
    "BlocksMergeSuperBlock"
    | "Blocks2ULs"
    | "Blocks2OLs"
    | "Blocks2TLs"
    | "Blocks2Blockquote"
    | "Blocks2Callout"

type TTurnIntoOneSub = "row" | "col"

type TTurnInto = "Blocks2Ps" | "Blocks2Hs"

type TEditorMode = "preview" | "wysiwyg"

type ILuteRenderCallback = (node: ILuteNode, entering: boolean) => [string, number];

type TProtyleAction = "cb-get-append" |
    "cb-get-before" |
    "cb-get-unchangeid" |
    "cb-get-hl" |
    "cb-get-focus" |
    "cb-get-focusfirst" |
    "cb-get-setid" |
    "cb-get-outline" |
    "cb-get-all" |
    "cb-get-backlink" |
    "cb-get-unundo" |
    "cb-get-scroll" |
    "cb-get-search" |
    "cb-get-context" |
    "cb-get-rootscroll" |
    "cb-get-html" |
    "cb-get-history" |
    "cb-get-opennew" |
    "cb-get-av-no-create"

/** @link  */
interface ILuteRender {
    renderDocument?: ILuteRenderCallback;
    renderParagraph?: ILuteRenderCallback;
    renderText?: ILuteRenderCallback;
    renderCodeBlock?: ILuteRenderCallback;
    renderCodeBlockOpenMarker?: ILuteRenderCallback;
    renderCodeBlockInfoMarker?: ILuteRenderCallback;
    renderCodeBlockCode?: ILuteRenderCallback;
    renderCodeBlockCloseMarker?: ILuteRenderCallback;
    renderMathBlock?: ILuteRenderCallback;
    renderMathBlockOpenMarker?: ILuteRenderCallback;
    renderMathBlockContent?: ILuteRenderCallback;
    renderMathBlockCloseMarker?: ILuteRenderCallback;
    renderBlockquote?: ILuteRenderCallback;
    renderBlockquoteMarker?: ILuteRenderCallback;
    renderHeading?: ILuteRenderCallback;
    renderHeadingC8hMarker?: ILuteRenderCallback;
    renderList?: ILuteRenderCallback;
    renderListItem?: ILuteRenderCallback;
    renderTaskListItemMarker?: ILuteRenderCallback;
    renderThematicBreak?: ILuteRenderCallback;
    renderHTML?: ILuteRenderCallback;
    renderTable?: ILuteRenderCallback;
    renderTableHead?: ILuteRenderCallback;
    renderTableRow?: ILuteRenderCallback;
    renderTableCell?: ILuteRenderCallback;
    renderCodeSpan?: ILuteRenderCallback;
    renderCodeSpanOpenMarker?: ILuteRenderCallback;
    renderCodeSpanContent?: ILuteRenderCallback;
    renderCodeSpanCloseMarker?: ILuteRenderCallback;
    renderInlineMath?: ILuteRenderCallback;
    renderInlineMathOpenMarker?: ILuteRenderCallback;
    renderInlineMathContent?: ILuteRenderCallback;
    renderInlineMathCloseMarker?: ILuteRenderCallback;
    renderEmphasis?: ILuteRenderCallback;
    renderEmAsteriskOpenMarker?: ILuteRenderCallback;
    renderEmAsteriskCloseMarker?: ILuteRenderCallback;
    renderEmUnderscoreOpenMarker?: ILuteRenderCallback;
    renderEmUnderscoreCloseMarker?: ILuteRenderCallback;
    renderStrong?: ILuteRenderCallback;
    renderStrongA6kOpenMarker?: ILuteRenderCallback;
    renderStrongA6kCloseMarker?: ILuteRenderCallback;
    renderStrongU8eOpenMarker?: ILuteRenderCallback;
    renderStrongU8eCloseMarker?: ILuteRenderCallback;
    renderStrikethrough?: ILuteRenderCallback;
    renderStrikethrough1OpenMarker?: ILuteRenderCallback;
    renderStrikethrough1CloseMarker?: ILuteRenderCallback;
    renderStrikethrough2OpenMarker?: ILuteRenderCallback;
    renderStrikethrough2CloseMarker?: ILuteRenderCallback;
    renderHardBreak?: ILuteRenderCallback;
    renderSoftBreak?: ILuteRenderCallback;
    renderInlineHTML?: ILuteRenderCallback;
    renderLink?: ILuteRenderCallback;
    renderOpenBracket?: ILuteRenderCallback;
    renderCloseBracket?: ILuteRenderCallback;
    renderOpenParen?: ILuteRenderCallback;
    renderCloseParen?: ILuteRenderCallback;
    renderLinkText?: ILuteRenderCallback;
    renderLinkSpace?: ILuteRenderCallback;
    renderLinkDest?: ILuteRenderCallback;
    renderLinkTitle?: ILuteRenderCallback;
    renderImage?: ILuteRenderCallback;
    renderBang?: ILuteRenderCallback;
    renderEmoji?: ILuteRenderCallback;
    renderEmojiUnicode?: ILuteRenderCallback;
    renderEmojiImg?: ILuteRenderCallback;
    renderEmojiAlias?: ILuteRenderCallback;
    renderBackslash?: ILuteRenderCallback;
    renderBackslashContent?: ILuteRenderCallback;
}

interface IBreadcrumb {
    id: string,
    name: string,
    type: string,
    subType: string,
    children: []
}

interface ILuteOptions extends IMarkdownConfig {
    emojis: IObject;
    emojiSite: string;
    headingAnchor?: boolean;
    lazyLoadImage?: string;
}

declare class Viz {
    public static instance(): Promise<Viz>;

    renderSVGElement: (code: string) => SVGElement;
}

declare class Viewer {
    public destroyed: boolean;

    constructor(element: Element, options: {
        title: [number, (image: HTMLImageElement, imageData: IObject) => string],
        button: boolean,
        initialViewIndex?: number,
        transition: boolean,
        hidden: () => void,
        toolbar: {
            zoomIn: boolean,
            zoomOut: boolean,
            oneToOne: boolean,
            reset: boolean,
            prev: boolean,
            play: boolean,
            next: boolean,
            rotateLeft: boolean,
            rotateRight: boolean,
            flipHorizontal: boolean,
            flipVertical: boolean,
            close: () => void
        }
    })

    public destroy(): void

    public show(): void
}

declare class Lute {
    public static WalkStop: number;
    public static WalkSkipChildren: number;
    public static WalkContinue: number;
    public static Version: string;
    public static Caret: string;

    public static New(): Lute;

    public static EChartsMindmapStr(text: string): string;

    public static NewNodeID(): string;

    public static Sanitize(html: string): string;

    public static EscapeHTMLStr(str: string): string;

    public static UnEscapeHTMLStr(str: string): string;

    public static GetHeadingID(node: ILuteNode): string;

    public static BlockDOM2Content(html: string): string;

    private constructor();

    public BlockDOM2Content(text: string): string;

    public BlockDOM2EscapeMarkerContent(text: string): string;

    public SetSpin(enable: boolean): void;

    public SetTextMark(enable: boolean): void;

    public SetHTMLTag2TextMark(enable: boolean): void;

    public SetHeadingID(enable: boolean): void;

    public SetProtyleMarkNetImg(enable: boolean): void;

    public SetSpellcheck(enable: boolean): void;

    public SetFileAnnotationRef(enable: boolean): void;

    public SetSetext(enable: boolean): void;

    public SetYamlFrontMatter(enable: boolean): void;

    public SetRenderListStyle(enable: boolean): void;

    public SetImgPathAllowSpace(enable: boolean): void;

    public SetKramdownIAL(enable: boolean): void;

    public BlockDOM2Md(html: string): string;

    public BlockDOM2StdMd(html: string): string;

    public SetSuperBlock(enable: boolean): void;

    public SetCallout(enable: boolean): void;

    public SetTag(enable: boolean): void;

    public SetInlineMath(enable: boolean): void;

    public SetGFMStrikethrough(enable: boolean): void;

    public SetGFMStrikethrough1(enable: boolean): void;

    public SetMark(enable: boolean): void;

    public SetSub(enable: boolean): void;

    public SetSup(enable: boolean): void;

    public SetInlineAsterisk(enable: boolean): void;

    public SetInlineUnderscore(enable: boolean): void;

    public SetBlockRef(enable: boolean): void;

    public SetSanitize(enable: boolean): void;

    public SetHeadingAnchor(enable: boolean): void;

    public SetImageLazyLoading(imagePath: string): void;

    public SetInlineMathAllowDigitAfterOpenMarker(enable: boolean): void;

    public SetToC(enable: boolean): void;

    public SetIndentCodeBlock(enable: boolean): void;

    public SetParagraphBeginningSpace(enable: boolean): void;

    public SetFootnotes(enable: boolean): void;

    public SetLinkRef(enable: boolean): void;

    public SetEmojiSite(emojiSite: string): void;

    public PutEmojis(emojis: IObject): void;

    public SpinBlockDOM(html: string): string;

    public Md2BlockDOM(html: string): string;

    public Md2BlockDOMWithAutoLink(html: string): string;

    public SetProtyleWYSIWYG(wysiwyg: boolean): void;

    public MarkdownStr(name: string, md: string): string;

    public ProtylePreviewStr(name: string, md: string): string;

    public GetLinkDest(text: string): string;

    public BlockDOM2InlineBlockDOM(html: string): string;

    public BlockDOM2HTML(html: string): string;

    public HTML2Md(html: string): string;

    public HTML2BlockDOM(html: string): string;

    public SetUnorderedListMarker(marker: string): void;

    public SetDataTask(marker: boolean): void;

    public SetExportNormalizeTaskListMarker(marker: boolean): void;

    public SetArbitraryTaskListItemMarker(marker: boolean): void;

    public SetEnsureListItemParagraph(enable: boolean): void;
}

declare const webkitAudioContext: {
    prototype: AudioContext
    new(contextOptions?: AudioContextOptions): AudioContext,
};

/** @link  */
interface IUpload {
    url?: string;
    max?: number;
    linkToImgUrl?: string;
    token?: string;
    accept?: string;
    withCredentials?: boolean;
    headers?: Record<string, string>;
    extraData?: { [key: string]: string | Blob };
    fieldName?: string;

    setHeaders?(): IObject;

    success?(editor: HTMLDivElement, msg: string): void;

    error?(msg: string): void;

    filename?(name: string): string;

    validate?(files: File[]): string | boolean;

    handler?(files: File[]): string | null;

    format?(files: File[], responseText: string): string;

    file?(files: File[]): File[];

    linkToImgCallback?(responseText: string): void;
}

interface IScrollAttr {
    rootId: string,
    startId?: string,
    endId?: string
    scrollTop?: number,
    focusId?: string,
    focusStart?: number
    focusEnd?: number
    zoomInId?: string
}

/** @link  */
interface IMenuItem {
    name: string;
    tip?: string;
    lang?: string;
    icon?: string;
    hotkey?: string;
    tipPosition?: string;
    showInLite?: boolean;

    click?(protyle: import("../protyle").Protyle): void;
}

/** @link  */
interface IMarkdownConfig {
    sanitize?: boolean;
    listStyle?: boolean;
}

/** @link  */
interface IPreview {
    delay?: number;
    mode?: "both" | "editor";
    url?: string;
    /** @link  */
    markdown?: IMarkdownConfig;
    /** @link   */
    actions?: Array<IPreviewAction | IPreviewActionCustom>;

    transform?(html: string): string;
}

type IPreviewAction = "desktop" | "tablet" | "mobile";

interface IPreviewActionCustom {
    key: string;
    text: string;
    className?: string;
    click: (key: string) => void;
}

interface IHintData {
    id?: string;
    html: string;
    value: string;
    filter?: string[];
    focus?: boolean;
}

interface IHintExtend {
    key: string;

    hint?(value: string, protyle: IProtyle, source: THintSource): IHintData[];
}

/** @link  */
interface IHint {
    emojiTail?: string;
    delay?: number;
    emoji?: IObject;
    emojiPath?: string;
    extend?: IHintExtend[];
}

/** @link  */
interface IProtyleOptions {
    databaseAttr?: boolean,
    history?: {
        created?: string
        snapshot?: string
    },
    backlinkData?: {
        blockPaths: IBreadcrumb[],
        dom: string
        expand: boolean
    }[],
    action?: TProtyleAction[],
    scrollPosition?: ScrollLogicalPosition,
    mode?: TEditorMode,
    blockId?: string
    rootId?: string
    notebookId?: string
    originalRefBlockIDs?: IObject
    key?: string
    defIds?: string[]
    render?: {
        background?: boolean
        title?: boolean
        titleShowTop?: boolean
        gutter?: boolean
        scroll?: boolean
        breadcrumb?: boolean
        breadcrumbDocName?: boolean
        hideTitleOnZoom?: boolean
    }
    _lutePath?: string;
    typewriterMode?: boolean;
    toolbar?: Array<string | IMenuItem>;
    /** @link  */
    preview?: IPreview;
    /** @link  */
    hint?: IHint;
    /** @link  */
    upload?: IUpload;
    /** @link  */
    classes?: {
        preview?: string;
    };
    click?: {
        preventInsetEmptyBlock?: boolean
    }

    handleEmptyContent?(): void

    after?(protyle: import("../protyle").Protyle): void;

    lite?: boolean;
}

interface IProtyle {
    highlight: {
        mark: Highlight
        markHL: Highlight
        ranges: Range[]
        rangeIndex: number
        styleElement: HTMLStyleElement
    }
    getInstance: () => import("../protyle").Protyle,
    observerLoad?: ResizeObserver,
    observer?: ResizeObserver,
    app: import("../index").App,
    id: string,
    query?: {
        key: string,
        method: number
        types: Config.IUILayoutTabSearchConfigTypes
        subTypes: Config.IUILayoutTabSearchConfigSubTypes
    },
    block: {
        id?: string,
        scroll?: boolean
        parentID?: string,
        parent2ID?: string,
        rootID?: string,
        showAll?: boolean
        mode?: number
        blockCount?: number
        action?: TProtyleAction[]
    },
    disabled: boolean,
    lite?: boolean,
    selectElement?: HTMLElement,
    ws?: import("../layout/Model").Model,
    notebookId?: string
    path?: string
    model?: import("../../src/editor").Editor,
    updated: boolean;
    element: HTMLElement;
    scroll?: import("../protyle/scroll").Scroll,
    gutter?: import("../protyle/gutter").Gutter,
    breadcrumb?: import("../protyle/breadcrumb").Breadcrumb,
    title?: import("../protyle/header/Title").Title,
    background?: import("../protyle/header/background").Background,
    databaseAttributePanel?: import("../protyle/render/av/attributePanel").AVAttributePanel,
    contentElement?: HTMLElement,
    options: IProtyleOptions;
    lute?: Lute;
    toolbar?: import("../protyle/toolbar").Toolbar,
    preview?: import("../protyle/preview").Preview;
    hint?: import("../protyle/hint").Hint;
    upload?: import("../protyle/upload").Upload;
    undo?: import("../protyle/undo").IUndo;
    wysiwyg?: import("../protyle/wysiwyg").WYSIWYG
}
