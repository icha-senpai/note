declare const SCRIBLI_VERSION: string;
declare const NODE_ENV: string;

const _SCRIBLI_VERSION = SCRIBLI_VERSION;
const _NODE_ENV = NODE_ENV;

const getFunctionKey = () => {
    const fData: { [key: number]: string } = {};
    for (let i = 1; i <= 32; i++) {
        fData[i + 111] = "F" + i;
    }
    return fData;
};

export abstract class Constants {
    public static readonly SCRIBLI_VERSION: string = _SCRIBLI_VERSION;
    public static readonly NODE_ENV: string = _NODE_ENV;
    public static readonly SCRIBLI_APPID: string = Math.random().toString(36).substring(8);

    public static readonly ASSETS_ADDRESS: string = "";
    public static readonly PROTYLE_CDN: string = "/stage/protyle";
    public static readonly UPLOAD_ADDRESS: string = "/upload";
    public static readonly SERVICE_WORKER_PATH: string = "/service-worker.js";

    public static readonly SCRIBLI_DROP_FILE: string = "application/scribli-file";
    public static readonly SCRIBLI_DROP_GUTTER: string = "application/scribli-gutter";
    public static readonly SCRIBLI_DROP_BLOCK_REF: string = "application/scribli-block-ref";
    public static readonly SCRIBLI_DROP_TAB: string = "application/scribli-tab";
    public static readonly SCRIBLI_DROP_EDITOR: string = "application/scribli-editor";

    public static readonly SCRIBLI_CMD: string = "scribli-cmd";
    public static readonly SCRIBLI_GET: string = "scribli-get";
    public static readonly SCRIBLI_EVENT: string = "scribli-event";

    public static readonly SCRIBLI_CONFIG_TRAY: string = "scribli-config-tray";
    public static readonly SCRIBLI_QUIT: string = "scribli-quit";
    public static readonly SCRIBLI_INSTALL_UPDATE: string = "scribli-install-update";
    public static readonly SCRIBLI_HOTKEY: string = "scribli-hotkey";
    public static readonly SCRIBLI_INIT: string = "scribli-init";
    public static readonly SCRIBLI_READY_TO_SHOW: string = "scribli-ready-to-show";
    public static readonly SCRIBLI_SEND_WINDOWS: string = "scribli-send-windows";
    public static readonly SCRIBLI_SAVE_CLOSE: string = "scribli-save-close";
    public static readonly SCRIBLI_AUTO_LAUNCH: string = "scribli-auto-launch";

    public static readonly SCRIBLI_OPEN_WORKSPACE: string = "scribli-open-workspace";
    public static readonly SCRIBLI_OPEN_URL: string = "scribli-open-url";
    public static readonly SCRIBLI_OPEN_WINDOW: string = "scribli-open-window";
    public static readonly SCRIBLI_OPEN_FILE: string = "scribli-open-file";

    public static readonly SCRIBLI_EXPORT_PDF: string = "scribli-export-pdf";
    public static readonly SCRIBLI_EXPORT_NEWWINDOW: string = "scribli-export-newwindow";

    public static readonly SCRIBLI_CONTEXT_MENU: string = "scribli-context-menu";
    public static readonly SCRIBLI_CONFIRM_DIALOG: string = "scribli-confirm-dialog";
    public static readonly SCRIBLI_ALERT_DIALOG: string = "scribli-alert-dialog";

    public static readonly SCRIBLI_SHOW_WINDOW: string = "scribli-show-window";

    // custom
    public static readonly CUSTOM_RIFF_DECKS: string = "custom-riff-decks";
    public static readonly CUSTOM_SY_READONLY: string = "custom-sy-readonly";
    public static readonly CUSTOM_SY_FULLWIDTH: string = "custom-sy-fullwidth";
    public static readonly CUSTOM_SY_AV_VIEW: string = "custom-sy-av-view";
    public static readonly CUSTOM_SY_TITLE_EMPTY: string = "custom-sy-title-empty";

    public static readonly ATTRIBUTE_EDITING = "data-editing";
    public static readonly ATTRIBUTE_V_SCROLL = "data-v-scroll";
    public static readonly ATTRIBUTE_DOCK_WIDTH = "data-dock-width";
    public static readonly ATTRIBUTE_MENU_KEYMAP = "data-menu-keymap";

    // size
    public static readonly SIZE_DATABASE_MAZ_SIZE: number = 102400;
    public static readonly SIZE_UPLOAD_TIP_SIZE: number = 268435456; // 256 M
    public static readonly SIZE_DRAG_THRESHOLD: number = 5;
    public static readonly SIZE_SCROLL_TB: number = 24;
    public static readonly SIZE_SCROLL_STEP: number = 256;
    public static readonly SIZE_LINK_TEXT_MAX: number = 64;
    public static readonly SIZE_GET_MAX = 102400;
    public static readonly SIZE_UNDO = 64;
    public static readonly SIZE_TITLE = 512;
    public static readonly SIZE_EDITOR_WIDTH = 760;
    public static readonly SIZE_ZOOM = [
        {
            zoom: 0.67,
            position: {x: 5, y: 2}
        },
        {
            zoom: 0.75,
            position: {x: 5, y: 4}
        }, {
            zoom: 0.8,
            position: {x: 6, y: 4}
        }, {
            zoom: 0.9,
            position: {x: 7, y: 6}
        }, {
            zoom: 1,
            position: {x: 8, y: 8}
        }, {
            zoom: 1.1,
            position: {x: 12, y: 9}
        }, {
            zoom: 1.25,
            position: {x: 18, y: 12}
        }, {
            zoom: 1.5,
            position: {x: 27, y: 16}
        }, {
            zoom: 1.75,
            position: {x: 36, y: 20}
        }, {
            zoom: 2,
            position: {x: 45, y: 23}
        }, {
            zoom: 2.5,
            position: {x: 63, y: 31}
        }, {
            zoom: 3,
            position: {x: 80, y: 39}
        }];

    // ws callback
    public static readonly CB_MOVE_NOLIST = "cb-move-nolist";
    public static readonly CB_GET_APPEND = "cb-get-append";
    public static readonly CB_GET_BEFORE = "cb-get-before";
    public static readonly CB_GET_UNCHANGEID = "cb-get-unchangeid";
    public static readonly CB_GET_HL = "cb-get-hl";
    public static readonly CB_GET_FOCUS = "cb-get-focus";
    public static readonly CB_GET_FOCUSFIRST = "cb-get-focusfirst";
    public static readonly CB_GET_SETID = "cb-get-setid";
    public static readonly CB_GET_OUTLINE = "cb-get-outline";
    public static readonly CB_GET_ALL = "cb-get-all";
    public static readonly CB_GET_BACKLINK = "cb-get-backlink";
    public static readonly CB_GET_UNUNDO = "cb-get-unundo";
    public static readonly CB_GET_SCROLL = "cb-get-scroll";
    public static readonly CB_GET_SEARCH = "cb-get-search";
    public static readonly CB_GET_CONTEXT = "cb-get-context";
    public static readonly CB_GET_ROOTSCROLL = "cb-get-rootscroll";
    public static readonly CB_GET_HTML = "cb-get-html";
    public static readonly CB_GET_HISTORY = "cb-get-history";
    public static readonly CB_GET_OPENNEW = "cb-get-opennew";
    public static readonly CB_GET_AV_NO_CREATE = "cb-get-av-no-create";

    // localstorage
    public static readonly LOCAL_ZOOM = "local-zoom";
    public static readonly LOCAL_SEARCHDATA = "local-searchdata";
    public static readonly LOCAL_SEARCHKEYS = "local-searchkeys";
    public static readonly LOCAL_SEARCHASSET = "local-searchasset";
    public static readonly LOCAL_SEARCHUNREF = "local-searchunref";
    public static readonly LOCAL_DOCINFO = "local-docinfo"; // only mobile
    public static readonly LOCAL_DAILYNOTEID = "local-dailynoteid"; // string
    public static readonly LOCAL_HISTORY = "local-history";
    public static readonly LOCAL_CODELANG = "local-codelang"; // string
    public static readonly LOCAL_FONTSTYLES = "local-fontstyles";
    public static readonly LOCAL_EXPORTPDF = "local-exportpdf";
    public static readonly LOCAL_EXPORTWORD = "local-exportword";
    public static readonly LOCAL_EXPORTIMG = "local-exportimg";
    public static readonly LOCAL_PDFTHEME = "local-pdftheme";
    public static readonly LOCAL_LAYOUTS = "local-layouts";
    public static readonly LOCAL_AI = "local-ai";
    public static readonly LOCAL_PLUGINTOPUNPIN = "local-plugintopunpin";
    public static readonly LOCAL_FLASHCARD = "local-flashcard";
    public static readonly LOCAL_FILEPOSITION = "local-fileposition";
    public static readonly LOCAL_FILESPATHS = "local-filespaths";
    public static readonly LOCAL_DIALOGPOSITION = "local-dialogposition";
    public static readonly LOCAL_SESSION_FIRSTLOAD = "local-session-firstload";
    public static readonly LOCAL_OUTLINE = "local-outline";
    public static readonly LOCAL_PLUGIN_DOCKS = "local-plugin-docks";
    public static readonly LOCAL_IMAGES = "local-images";
    public static readonly LOCAL_EMOJIS = "local-emojis";
    public static readonly LOCAL_MOVE_PATH = "local-move-path";
    public static readonly LOCAL_RECENT_DOCS = "local-recent-docs";
    public static readonly LOCAL_CLOSED_TABS = "local-closed-tabs";

    // dialog
    public static readonly DIALOG_CONFIRM = "dialog-confirm";
    public static readonly DIALOG_OPENCARD = "dialog-opencard";
    public static readonly DIALOG_MAKECARD = "dialog-makecard";
    public static readonly DIALOG_VIEWCARDS = "dialog-viewcards";
    public static readonly DIALOG_DIALYNOTE = "dialog-dialynote";
    public static readonly DIALOG_RECENTDOCS = "dialog-recentdocs";
    public static readonly DIALOG_SWITCHTAB = "dialog-switchtab";
    public static readonly DIALOG_SEARCH = "dialog-search";
    public static readonly DIALOG_REPLACE = "dialog-replace";
    public static readonly DIALOG_GLOBALSEARCH = "dialog-globalsearch";
    public static readonly DIALOG_HISTORYCOMPARE = "dialog-historycompare";

    public static readonly DIALOG_ACCESSAUTHCODE = "dialog-accessauthcode";
    public static readonly DIALOG_AICUSTOMACTION = "dialog-aicustomaction";
    public static readonly DIALOG_AIUPDATECUSTOMACTION = "dialog-aiupdatecustomaction";
    public static readonly DIALOG_AIPROVIDER = "dialog-aiprovider";
    public static readonly DIALOG_AIMODEL = "dialog-aimodel";
    public static readonly DIALOG_AIMCPSERVER = "dialog-aimcpserver";
    public static readonly DIALOG_BACKGROUNDLINK = "dialog-backgroundlink";
    public static readonly DIALOG_BACKGROUNDRANDOM = "dialog-backgroundrandom";
    public static readonly DIALOG_CHANGELOG = "dialog-changelog";
    public static readonly DIALOG_COMMANDPANEL = "dialog-commandpanel";
    public static readonly DIALOG_DEACTIVATEUSER = "dialog-deactivateuser";
    public static readonly DIALOG_EMOJIS = "dialog-emojis";
    public static readonly DIALOG_EXPORTIMAGE = "dialog-exportimage";
    public static readonly DIALOG_EXPORTTEMPLATE = "dialog-exporttemplate";
    public static readonly DIALOG_EXPORTWORD = "dialog-exportword";
    public static readonly DIALOG_EXPORTMARKDOWN = "dialog-exportmarkdown";
    public static readonly DIALOG_HISTORY = "dialog-history";
    public static readonly DIALOG_HISTORYDOC = "dialog-historydoc";
    public static readonly DIALOG_MOVEPATHTO = "dialog-movepathto";
    public static readonly DIALOG_RENAME = "dialog-rename";
    public static readonly DIALOG_RENAMEASSETS = "dialog-renameassets";
    public static readonly DIALOG_RENAMEBOOKMARK = "dialog-renamebookmark";
    public static readonly DIALOG_RENAMETAG = "dialog-renametag";
    public static readonly DIALOG_REPLACETYPE = "dialog-replacetype";
    public static readonly DIALOG_SAVECRITERION = "dialog-savecriterion";
    public static readonly DIALOG_SEARCHTYPE = "dialog-searchtype";
    public static readonly DIALOG_SEARCHASSETSTYPE = "dialog-searchassetstype";
    public static readonly DIALOG_SETTING = "dialog-setting";
    public static readonly DIALOG_SNAPSHOTTAG = "dialog-snapshottag";
    public static readonly DIALOG_SNAPSHOTMEMO = "dialog-snapshotmemo";
    public static readonly DIALOG_SNIPPETS = "dialog-snippets";
    public static readonly DIALOG_SYNCADDCLOUDDIR = "dialog-syncaddclouddir";
    public static readonly DIALOG_SYNCCHOOSEDIR = "dialog-syncchoosedir";
    public static readonly DIALOG_SYNCCHOOSEDIRECTION = "dialog-syncchoosedirection";
    public static readonly DIALOG_TRANSFERBLOCKREF = "dialog-transferblockref";
    public static readonly DIALOG_PASSWORD = "dialog-password";
    public static readonly DIALOG_SETPASSWORD = "dialog-setpassword";
    public static readonly DIALOG_BOOTSYNCFAILED = "dialog-bootsyncfailed";
    public static readonly DIALOG_KERNELFAULT = "dialog-kernelfault";
    public static readonly DIALOG_STATEEXCEPTED = "dialog-stateexcepted";
    public static readonly DIALOG_ATTR = "dialog-attr";
    public static readonly DIALOG_SETCUSTOMATTR = "dialog-setcustomattr";
    public static readonly DIALOG_CREATENOTEBOOK = "dialog-createnotebook";
    public static readonly DIALOG_NOTEBOOKCONF = "dialog-notebookconf";
    public static readonly DIALOG_CREATEWORKSPACE = "dialog-createworkspace";
    public static readonly DIALOG_OPENWORKSPACE = "dialog-openworkspace";
    public static readonly DIALOG_SAVEWORKSPACE = "dialog-saveworkspace";

    // menu
    public static readonly MENU_BAR_WORKSPACE = "barWorkspace";
    public static readonly MENU_BAR_PLUGIN = "topBarPlugin";
    public static readonly MENU_BAR_ZOOM = "barZoom";
    public static readonly MENU_BAR_MODE = "barmode";
    public static readonly MENU_BAR_MORE = "barmore";
    public static readonly MENU_STATUS_BACKGROUND_TASK = "statusBackgroundTask";
    public static readonly MENU_DOCK = "menu-dock";
    public static readonly MENU_DOCK_MOBILE = "dockMobileMenu";

    public static readonly MENU_BLOCK_SINGLE = "block-single";
    public static readonly MENU_BLOCK_MULTI = "block-multi";
    public static readonly MENU_TITLE = "titleMenu";
    public static readonly MENU_FROM_TITLE_PROTYLE = "title-protyle";
    public static readonly MENU_FROM_TITLE_BREADCRUMB = "title-breadcrumb";
    public static readonly MENU_BREADCRUMB_MORE = "breadcrumbMore";
    public static readonly MENU_BREADCRUMB_MOBILE_PATH = "breadcrumb-mobile-path";

    public static readonly MENU_DOC_TREE_MORE = "docTreeMore";
    public static readonly MENU_FROM_DOC_TREE_MORE_NOTEBOOK = "tree-notebook";
    public static readonly MENU_FROM_DOC_TREE_MORE_DOC = "tree-doc";
    public static readonly MENU_FROM_DOC_TREE_MORE_ITEMS = "tree-items";
    public static readonly MENU_TAG = "tagMenu";
    public static readonly MENU_BOOKMARK = "bookmarkMenu";
    public static readonly MENU_OUTLINE_CONTEXT = "outline-context";
    public static readonly MENU_OUTLINE_EXPAND_LEVEL = "outline-expand-level";

    public static readonly MENU_AV_VIEW = "av-view";
    public static readonly MENU_AV_HEADER_CELL = "av-header-cell";
    public static readonly MENU_AV_HEADER_ADD = "av-header-add";
    public static readonly MENU_AV_ADD_FILTER = "av-add-filter";
    public static readonly MENU_AV_ADD_SORT = "av-add-sort";
    public static readonly MENU_AV_COL_OPTION = "av-col-option";
    public static readonly MENU_AV_COL_FORMAT_NUMBER = "av-col-format-number";
    public static readonly MENU_AV_GROUP_DATE = "avGroupDate";
    public static readonly MENU_AV_GROUP_SORT = "avGroupSort";
    public static readonly MENU_AV_ASSET_EDIT = "av-asset-edit";
    public static readonly MENU_AV_CALC = "av-calc";
    public static readonly MENU_AV_PAGE_SIZE = "av-page-size";

    public static readonly MENU_SEARCH_MORE = "searchMore";
    public static readonly MENU_SEARCH_METHOD = "searchMethod";
    public static readonly MENU_SEARCH_ASSET_MORE = "searchAssetMore";
    public static readonly MENU_SEARCH_ASSET_METHOD = "searchAssetMethod";
    public static readonly MENU_SEARCH_UNREF_MORE = "searchUnRefMore";
    public static readonly MENU_SEARCH_HISTORY = "search-history";
    public static readonly MENU_SEARCH_REPLACE_HISTORY = "search-replace-history";
    public static readonly MENU_SEARCH_ASSET_HISTORY = "search-asset-history";
    public static readonly MENU_MOVE_PATH_HISTORY = "move-path-history";
    public static readonly MENU_CALLOUT_SELECT = "callout-select";

    public static readonly MENU_BACKGROUND_ASSET = "background-asset";
    public static readonly MENU_AI = "ai";
    public static readonly MENU_TAB = "tab";
    public static readonly MENU_TAB_LIST = "tabList";

    public static readonly MENU_INLINE_CONTEXT = "inline-context";
    public static readonly MENU_INLINE_IMG = "inline-img";
    public static readonly MENU_INLINE_FILE_ANNOTATION_REF = "inline-file-annotation-ref";
    public static readonly MENU_INLINE_REF = "inline-block-ref";
    public static readonly MENU_INLINE_A = "inline-a";
    public static readonly MENU_INLINE_TAG = "inline-tag";
    public static readonly MENU_INLINE_MATH = "inline-math";

    // timeout
    public static readonly TIMEOUT_OPENDIALOG = 50;
    public static readonly TIMEOUT_DBLCLICK = 190;
    public static readonly TIMEOUT_RESIZE = 200;
    public static readonly TIMEOUT_INPUT = 256;
    public static readonly TIMEOUT_LOAD = 300;
    public static readonly TIMEOUT_LONGPRESS = 400;
    public static readonly TIMEOUT_MOUSE_DRAG_DELAY = 150;
    public static readonly TIMEOUT_MULTIPLE_SELECT = 1500;
    public static readonly TIMEOUT_TRANSITION = 300;
    public static readonly TIMEOUT_COUNT = 1000;

    public static readonly QUICK_DECK_ID = "20230218211946-2kw8jgx";

    public static KEYCODELIST: { [key: number]: string } = Object.assign(getFunctionKey(), {
        8: "⌫",
        9: "⇥",
        13: "↩",
        16: "⇧",
        17: "⌃",
        18: "⌥",
        19: "Pause",
        20: "CapsLock",
        27: "Escape",
        32: " ",
        33: "PageUp",
        34: "PageDown",
        35: "End",
        36: "Home",
        37: "←",
        38: "↑",
        39: "→",
        40: "↓",
        44: "PrintScreen",
        45: "Insert",
        46: "⌦",
        48: "0",
        49: "1",
        50: "2",
        51: "3",
        52: "4",
        53: "5",
        54: "6",
        55: "7",
        56: "8",
        57: "9",
        65: "A",
        66: "B",
        67: "C",
        68: "D",
        69: "E",
        70: "F",
        71: "G",
        72: "H",
        73: "I",
        74: "J",
        75: "K",
        76: "L",
        77: "M",
        78: "N",
        79: "O",
        80: "P",
        81: "Q",
        82: "R",
        83: "S",
        84: "T",
        85: "U",
        86: "V",
        87: "W",
        88: "X",
        89: "Y",
        90: "Z",
        91: "⌘",
        92: "⌘",
        93: "ContextMenu",
        96: "0",
        97: "1",
        98: "2",
        99: "3",
        100: "4",
        101: "5",
        102: "6",
        103: "7",
        104: "8",
        105: "9",
        106: "*",
        107: "+",
        109: "-",
        110: ".",
        111: "/",
        144: "NumLock",
        145: "ScrollLock",
        182: "MyComputer",
        183: "MyCalculator",
        186: ";",
        187: "=",
        188: ",",
        189: "-",
        190: ".",
        191: "/",
        192: "`",
        219: "[",
        220: "\\",
        221: "]",
        222: "'",
    });
    // "⌘", "⇧", "⌥", "⌃"
    // "⌘A", "⌘X", "⌘C", "⌘V", "⌘-", "⌘=", "⌘0", "⇧⌘V", "⌘/", "⇧↑", "⇧↓", "⇧→", "⇧←", "⇧⇥", "⌃D", "⇧⌘→", "⇧⌘←",
    public static readonly SCRIBLI_KEYMAP: Config.IKeymap = {
        general: {
            mainMenu: {default: "⌥\\", custom: "⌥\\"},
            commandPanel: {default: "⌥⇧P", custom: "⌥⇧P"},
            editReadonly: {default: "⇧⌘G", custom: "⇧⌘G"},
            syncNow: {default: "F9", custom: "F9"},
            enterBack: {default: "⌥←", custom: "⌥←"},
            enter: {default: "⌥→", custom: "⌥→"},
            goForward: {default: "⌘]", custom: "⌘]"},
            goBack: {default: "⌘[", custom: "⌘["},
            newFile: {default: "⌘N", custom: "⌘N"},
            search: {default: "⌘F", custom: "⌘F"},
            globalSearch: {default: "⌘P", custom: "⌘P"},
            stickSearch: {default: "⇧⌘F", custom: "⇧⌘F"},
            replace: {default: "⌘R", custom: "⌘R"},
            closeTab: {default: "⌘W", custom: "⌘W"},
            fileTree: {default: "⌃1", custom: "⌃1"},
            outline: {default: "⌃2", custom: "⌃2"},
            bookmark: {default: "⌃3", custom: "⌃3"},
            tag: {default: "⌃4", custom: "⌃4"},
            dailyNote: {default: "⌃5", custom: "⌃5"},
            inbox: {default: "⌃6", custom: "⌃6"},
            backlinks: {default: "⌃7", custom: "⌃7"},
            graphView: {default: "⌃8", custom: "⌃8"},
            globalGraph: {default: "⌃9", custom: "⌃9"},
            riffCard: {default: "⌃0", custom: "⌃0"},
            config: {default: "⌥P", custom: "⌥P"},
            dataHistory: {default: "⌥H", custom: "⌥H"},
            toggleWin: {default: "⌥M", custom: "⌥M"},
            lockScreen: {default: "⌥N", custom: "⌥N"},
            recentDocs: {default: "⌘E", custom: "⌘E"},
            goToTab1: {default: "⌘1", custom: "⌘1"},
            goToTab2: {default: "⌘2", custom: "⌘2"},
            goToTab3: {default: "⌘3", custom: "⌘3"},
            goToTab4: {default: "⌘4", custom: "⌘4"},
            goToTab5: {default: "⌘5", custom: "⌘5"},
            goToTab6: {default: "⌘6", custom: "⌘6"},
            goToTab7: {default: "⌘7", custom: "⌘7"},
            goToTab8: {default: "⌘8", custom: "⌘8"},
            goToTab9: {default: "⌘9", custom: "⌘9"},
            goToTabNext: {default: "⇧⌘]", custom: "⇧⌘]"},
            goToTabPrev: {default: "⇧⌘[", custom: "⇧⌘["},
            goToEditTabNext: {default: "⌃⇥", custom: "⌃⇥"},
            goToEditTabPrev: {default: "⌃⇧⇥", custom: "⌃⇧⇥"},
            recentClosed: {default: "⇧⌘T", custom: "⇧⌘T"},
            move: {default: "", custom: ""},
            selectOpen1: {default: "", custom: ""},
            switchLeftDock: {default: "", custom: ""},
            switchRightDock: {default: "", custom: ""},
            switchBottomDock: {default: "", custom: ""},
            toggleDock: {default: "", custom: ""},
            splitLR: {default: "", custom: ""},
            splitMoveR: {default: "", custom: ""},
            splitTB: {default: "", custom: ""},
            splitMoveB: {default: "", custom: ""},
            closeOthers: {default: "", custom: ""},
            closeAll: {default: "", custom: ""},
            closeUnmodified: {default: "", custom: ""},
            closeLeft: {default: "", custom: ""},
            closeRight: {default: "", custom: ""},
            tabToWindow: {default: "", custom: ""},
            addToDatabase: {default: "", custom: ""},
            unsplit: {default: "", custom: ""},
            unsplitAll: {default: "", custom: ""},
        },
        editor: {
            general: {
                duplicate: {default: "⌘D", custom: "⌘D"},
                expandDown: {default: "⌥⇧↓", custom: "⌥⇧↓"},
                expandUp: {default: "⌥⇧↑", custom: "⌥⇧↑"},
                expand: {default: "⌘↓", custom: "⌘↓"},
                collapse: {default: "⌘↑", custom: "⌘↑"},
                foldRecursive: {default: "⌥⌘↑", custom: "⌥⌘↑"},
                insertBottom: {default: "⌥⌘.", custom: "⌥⌘."},
                refTab: {default: "⇧⌘.", custom: "⇧⌘."},
                openBy: {default: "⌥,", custom: "⌥,"},
                insertRight: {default: "⌥.", custom: "⌥."},
                attr: {default: "⌥⌘A", custom: "⌥⌘A"},
                quickMakeCard: {default: "⌥⌘F", custom: "⌥⌘F"},
                refresh: {default: "F5", custom: "F5"},
                copyBlockRef: {default: "⇧⌘C", custom: "⇧⌘C"},
                copyProtocol: {default: "⇧⌘H", custom: "⇧⌘H"},
                copyBlockEmbed: {default: "⇧⌘E", custom: "⇧⌘E"},
                copyHPath: {default: "⇧⌘P", custom: "⇧⌘P"},
                undo: {default: "⌘Z", custom: "⌘Z"},
                redo: {default: "⇧⌘Z", custom: "⇧⌘Z"},
                rename: {default: "F2", custom: "F2"},
                newNameFile: {default: "F3", custom: "F3"},
                newContentFile: {default: "F4", custom: "F4"},
                newNameSettingFile: {default: "⌘F3", custom: "⌘F3"},
                showInFolder: {default: "⌥A", custom: "⌥A"},
                outline: {default: "⌥O", custom: "⌥O"},
                backlinks: {default: "⌥B", custom: "⌥B"},
                graphView: {default: "⌥G", custom: "⌥G"},
                spaceRepetition: {default: "⌥F", custom: "⌥F"},
                fullscreen: {default: "⌥Y", custom: "⌥Y"},
                alignLeft: {default: "⌥L", custom: "⌥L"},
                alignCenter: {default: "⌥C", custom: "⌥C"},
                alignRight: {default: "⌥R", custom: "⌥R"},
                wysiwyg: {default: "⌥⌘7", custom: "⌥⌘7"},
                preview: {default: "⌥⌘9", custom: "⌥⌘9"},
                insertBefore: {default: "⇧⌘B", custom: "⇧⌘B"},
                insertAfter: {default: "⇧⌘A", custom: "⇧⌘A"},
                jumpToParentNext: {default: "⇧⌘N", custom: "⇧⌘N"},
                jumpToParentPrev: {default: "⇧⌘M", custom: "⇧⌘M"},
                jumpToParent: {default: "⇧⌘J", custom: "⇧⌘J"},
                moveToUp: {default: "⇧⌘↑", custom: "⇧⌘↑"},
                moveToDown: {default: "⇧⌘↓", custom: "⇧⌘↓"},
                duplicateCompletely: {default: "", custom: ""},
                copyPlainText: {default: "", custom: ""},
                copyID: {default: "", custom: ""},
                copyProtocolInMd: {default: "", custom: ""},
                netImg2LocalAsset: {default: "", custom: ""},
                netAssets2LocalAssets: {default: "", custom: ""},
                hLayout: {default: "", custom: ""},
                vLayout: {default: "", custom: ""},
                refPopover: {default: "", custom: ""},
                copyText: {default: "", custom: ""},
                exitFocus: {default: "", custom: ""},
                ai: {default: "", custom: ""},
                switchReadonly: {default: "", custom: ""},
                switchAdjust: {default: "", custom: ""},
                rtl: {default: "", custom: ""},
                ltr: {default: "", custom: ""},
                aiWriting: {default: "", custom: ""},
                openInNewTab: {default: "", custom: ""},
            },
            insert: {
                appearance: {default: "⌥⌘X", custom: "⌥⌘X"},
                lastUsed: {default: "⌥X", custom: "⌥X"},
                ref: {default: "⌥[", custom: "⌥["},
                kbd: {default: "⌘'", custom: "⌘'"},
                sup: {default: "⌘H", custom: "⌘H"},
                sub: {default: "⌘J", custom: "⌘J"},
                bold: {default: "⌘B", custom: "⌘B"},
                "inline-math": {default: "⌘M", custom: "⌘M"},
                memo: {default: "⌥⌘M", custom: "⌥⌘M"},
                underline: {default: "⌘U", custom: "⌘U"},
                italic: {default: "⌘I", custom: "⌘I"},
                mark: {default: "⌥D", custom: "⌥D"},
                tag: {default: "⌘T", custom: "⌘T"},
                strike: {default: "⇧⌘S", custom: "⇧⌘S"},
                "inline-code": {default: "⌘G", custom: "⌘G"},
                link: {default: "⌘K", custom: "⌘K"},
                check: {default: "⌘L", custom: "⌘L"},
                "ordered-list": {default: "", custom: ""},
                list: {default: "", custom: ""},
                table: {default: "⌘O", custom: "⌘O"},
                code: {default: "⇧⌘K", custom: "⇧⌘K"},
                quote: {default: "", custom: ""},
                clearInline: {default: "⌘\\", custom: "⌘\\"},
            },
            heading: {
                paragraph: {default: "⌥⌘0", custom: "⌥⌘0"},
                heading1: {default: "⌥⌘1", custom: "⌥⌘1"},
                heading2: {default: "⌥⌘2", custom: "⌥⌘2"},
                heading3: {default: "⌥⌘3", custom: "⌥⌘3"},
                heading4: {default: "⌥⌘4", custom: "⌥⌘4"},
                heading5: {default: "⌥⌘5", custom: "⌥⌘5"},
                heading6: {default: "⌥⌘6", custom: "⌥⌘6"},
            },
            list: {
                indent: {default: "⇥", custom: "⇥"},
                outdent: {default: "⇧⇥", custom: "⇧⇥"},
                checkToggle: {default: "⌘↩", custom: "⌘↩"},
            },
            table: {
                insertRowAbove: {default: "", custom: ""},
                insertRowBelow: {default: "", custom: ""},
                insertColumnLeft: {default: "", custom: ""},
                insertColumnRight: {default: "", custom: ""},
                moveToUp: {default: "⌥⌘T", custom: "⌥⌘T"},
                moveToDown: {default: "⌥⌘B", custom: "⌥⌘B"},
                moveToLeft: {default: "⌥⌘L", custom: "⌥⌘L"},
                moveToRight: {default: "⌥⌘R", custom: "⌥⌘R"},
                "delete-row": {default: "⌘-", custom: "⌘-"},
                "delete-column": {default: "⇧⌘-", custom: "⇧⌘-"}
            }
        },
        plugin: {},
    };

    public static readonly SCRIBLI_EMPTY_LAYOUT: Config.IUiLayout = {
        hideDock: false,
        layout: {
            "direction": "tb",
            "size": "0px",
            "type": "normal",
            "instance": "Layout",
            "children": [{
                "direction": "lr",
                "size": "auto",
                "type": "normal",
                "instance": "Layout",
                "children": [{
                    "direction": "tb",
                    "size": "0px",
                    "type": "left",
                    "instance": "Layout",
                    "children": [{
                        "instance": "Wnd",
                        "children": []
                    }, {
                        "instance": "Wnd",
                        "resize": "tb",
                        "children": []
                    }]
                }, {
                    "direction": "lr",
                    "resize": "lr",
                    "size": "auto",
                    "type": "center",
                    "instance": "Layout",
                    "children": [{
                        "instance": "Wnd",
                        "children": [{
                            "instance": "Tab",
                            "children": []
                        }]
                    }]
                }, {
                    "direction": "tb",
                    "size": "0px",
                    "resize": "lr",
                    "type": "right",
                    "instance": "Layout",
                    "children": [{
                        "instance": "Wnd",
                        "children": []
                    }, {
                        "instance": "Wnd",
                        "resize": "tb",
                        "children": []
                    }]
                }]
            }, {
                "direction": "lr",
                "size": "0px",
                "resize": "tb",
                "type": "bottom",
                "instance": "Layout",
                "children": [{
                    "instance": "Wnd",
                    "children": []
                }, {
                    "instance": "Wnd",
                    "resize": "lr",
                    "children": []
                }]
            }]
        },
        bottom: {
            pin: true,
            data: []
        },
        left: {
            pin: true,
            data: [
                [{
                    type: "file",
                    size: {width: 232, height: 0},
                    show: true,
                    icon: "iconFiles",
                    hotkeyLangId: "fileTree",
                }, {
                    type: "outline",
                    size: {width: 232, height: 0},
                    show: false,
                    icon: "iconOutline",
                    hotkeyLangId: "outline",
                }, {
                    type: "inbox",
                    size: {width: 320, height: 0},
                    show: false,
                    icon: "iconInbox",
                    hotkeyLangId: "inbox",
                }], [{
                    type: "bookmark",
                    size: {width: 232, height: 0},
                    show: false,
                    icon: "iconBookmark",
                    hotkeyLangId: "bookmark",
                }, {
                    type: "tag",
                    size: {width: 232, height: 0},
                    show: false,
                    icon: "iconTag",
                    hotkeyLangId: "tag",
                }]
            ]
        },
        right: {
            pin: true,
            data: [
                [{
                    type: "agentChat",
                    size: {width: 320, height: 0},
                    show: false,
                    icon: "iconSparkles",
                    hotkeyLangId: "agentChat",
                }, {
                    type: "graph",
                    size: {width: 320, height: 0},
                    show: false,
                    icon: "iconGraph",
                    hotkeyLangId: "graphView",
                }, {
                    type: "globalGraph",
                    size: {width: 320, height: 0},
                    show: false,
                    icon: "iconGlobalGraph",
                    hotkeyLangId: "globalGraph",
                }], [{
                    type: "backlink",
                    size: {width: 320, height: 0},
                    show: false,
                    icon: "iconLink",
                    hotkeyLangId: "backlinks",
                }]
            ]
        }
    };

    public static readonly SCRIBLI_DEFAULT_REPLACETYPES: Required<Config.IUILayoutTabSearchConfigReplaceTypes> = {
        "text": true,
        "imgText": true,
        "imgTitle": true,
        "imgSrc": true,
        "aText": true,
        "aTitle": true,
        "aHref": true,
        "code": true,
        "em": true,
        "strong": true,
        "inlineMath": true,
        "inlineMemo": true,
        "blockRef": true,
        "fileAnnotationRef": true,
        "kbd": true,
        "mark": true,
        "s": true,
        "sub": true,
        "sup": true,
        "tag": true,
        "u": true,
        "docTitle": true,
        "codeBlock": true,
        "mathBlock": true,
        "htmlBlock": true
    };

    // image
    public static readonly SCRIBLI_IMAGE_SPONSOR: string = `<svg version="1.1" xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 32 32">
<path fill="#ffe43c" d="M6.4 0h19.2c4.268 0 6.4 2.132 6.4 6.4v19.2c0 4.268-2.132 6.4-6.4 6.4h-19.2c-4.268 0-6.4-2.132-6.4-6.4v-19.2c0-4.268 2.135-6.4 6.4-6.4z"></path>
<path fill="#00f5d4" d="M25.6 0h-8.903c-7.762 1.894-14.043 7.579-16.697 15.113v10.487c0 3.533 2.867 6.4 6.4 6.4h19.2c3.533 0 6.4-2.867 6.4-6.4v-19.2c0-3.537-2.863-6.4-6.4-6.4z"></path>
<path fill="#01beff" d="M25.6 0h-0.119c-12.739 2.754-20.833 15.316-18.079 28.054 0.293 1.35 0.702 2.667 1.224 3.946h16.974c3.533 0 6.4-2.867 6.4-6.4v-19.2c0-3.537-2.863-6.4-6.4-6.4z"></path>
<path fill="#9a5ce5" d="M31.005 2.966c-0.457-0.722-1.060-1.353-1.784-1.849-8.342 3.865-13.683 12.223-13.679 21.416-0.003 3.256 0.67 6.481 1.978 9.463h8.081c0.602 0 1.185-0.084 1.736-0.238-2.1-3.189-3.401-7.624-3.401-12.526 0-7.337 2.921-13.628 7.070-16.266z"></path>
<path fill="#f15bb5" d="M32 25.6v-19.2c0-1.234-0.354-2.419-0.998-3.43-4.149 2.638-7.067 8.928-7.067 16.266 0 4.902 1.301 9.334 3.401 12.526 2.693-0.757 4.664-3.231 4.664-6.162z"></path>
<path fill="#fff" opacity="0.2" d="M26.972 22.415c-2.889 0.815-4.297 2.21-6.281 3.182 1.552 0.348 3.105 0.461 4.902 0.461 2.644 0 5.363-1.449 6.406-2.519v-1.085c-1.598-0.399-2.664-0.705-5.028-0.039zM4.773 21.612c-0.003 0-0.006-0.003-0.006-0.003-1.726-0.863-3.382-1.205-4.767-1.301v2.487c0.779-0.341 2.396-0.921 4.773-1.182zM17.158 26.599c1.472-0.158 2.57-0.531 3.533-1.002-1.063-0.238-2.126-0.583-3.269-1.079-2.767-1.205-5.63-3.092-10.491-3.034-0.779 0.010-1.495 0.058-2.158 0.132 4.503 2.248 7.882 5.463 12.384 4.983z"></path>
<path fill="#fff" opacity="0.2" d="M20.691 25.594c-0.963 0.47-2.061 0.844-3.533 1.002-4.503 0.483-7.882-2.731-12.381-4.983-2.38 0.261-3.994 0.841-4.773 1.179v2.809c0 4.268 2.132 6.4 6.4 6.4h19.197c4.268 0 6.4-2.132 6.4-6.4v-2.065c-1.044 1.069-3.762 2.519-6.406 2.519-1.797 0-3.35-0.113-4.902-0.461z"></path>
<path fill="#fff" opacity="0.5" d="M3.479 19.123c0 0.334 0.271 0.606 0.606 0.606s0.606-0.271 0.606-0.606v0c0-0.334-0.271-0.606-0.606-0.606s-0.606 0.271-0.606 0.606v0z"></path>
<path fill="#fff" opacity="0.5" d="M29.027 14.266c0 0.334 0.271 0.606 0.606 0.606s0.606-0.271 0.606-0.606v0c0-0.334-0.271-0.606-0.606-0.606s-0.606 0.271-0.606 0.606v0z"></path>
<path fill="#fff" d="M9.904 1.688c0 0.167 0.136 0.303 0.303 0.303s0.303-0.136 0.303-0.303v0c0-0.167-0.136-0.303-0.303-0.303s-0.303 0.136-0.303 0.303v0z"></path>
<path fill="#fff" d="M2.673 10.468c0 0.167 0.136 0.303 0.303 0.303s0.303-0.136 0.303-0.303v0c0-0.167-0.136-0.303-0.303-0.303s-0.303 0.136-0.303 0.303v0z"></path>
<path fill="#fff" opacity="0.6" d="M30.702 9.376c0 0.167 0.136 0.303 0.303 0.303s0.303-0.136 0.303-0.303v0c0-0.167-0.136-0.303-0.303-0.303s-0.303 0.136-0.303 0.303v0z"></path>
<path fill="#fff" opacity="0.8" d="M29.236 20.881c0 0.276 0.224 0.499 0.499 0.499s0.499-0.224 0.499-0.499v0c0-0.276-0.224-0.499-0.499-0.499s-0.499 0.224-0.499 0.499v0z"></path>
<path fill="#fff" opacity="0.8" d="M15.38 1.591c0.047 0.016 0.101 0.026 0.158 0.026 0.276 0 0.499-0.224 0.499-0.499 0-0.219-0.141-0.406-0.338-0.473l-0.004-0.001c-0.047-0.016-0.101-0.026-0.158-0.026-0.276 0-0.499 0.224-0.499 0.499 0 0.219 0.141 0.406 0.338 0.473l0.004 0.001z"></path>
<path fill="#ffdeeb" d="M25.732 8.268c-2.393-2.371-6.249-2.371-8.642 0l-1.089 1.085-1.079-1.089c-2.38-2.39-6.249-2.393-8.639-0.013s-2.393 6.249-0.013 8.639l2.158 2.158 6.474 6.464c0.596 0.593 1.562 0.593 2.158 0l6.474-6.464 2.193-2.158c2.384-2.383 2.384-6.242 0.003-8.622z"></path>
<path fill="#fff" d="M17.081 8.268l-1.079 1.085-1.079-1.089c-2.38-2.39-6.249-2.393-8.639-0.013s-2.393 6.249-0.013 8.639l2.158 2.158 2.548 2.487c4.097-1.044 7.627-3.646 9.837-7.254 1.424-2.271 2.284-4.848 2.503-7.518-2.193-0.715-4.606-0.132-6.236 1.504z"></path>
</svg>`;

    // assets
    public static readonly SCRIBLI_ASSETS_IMAGE: string[] = [".apng", ".ico", ".cur", ".jpg", ".jpe", ".jpeg", ".jfif", ".pjp", ".pjpeg", ".png", ".gif", ".webp", ".bmp", ".svg", ".avif", ".tiff", ".tif"];
    public static readonly SCRIBLI_ASSETS_AUDIO: string[] = [".mp3", ".wav", ".ogg", ".m4a", ".aac", ".flac"];
    public static readonly SCRIBLI_ASSETS_VIDEO: string[] = [".mov", ".weba", ".mkv", ".mp4", ".webm"];
    public static readonly SCRIBLI_ASSETS_EXTS: string[] = [".pdf"].concat(Constants.SCRIBLI_ASSETS_IMAGE, Constants.SCRIBLI_ASSETS_AUDIO, Constants.SCRIBLI_ASSETS_VIDEO);
    public static readonly SCRIBLI_ASSETS_SEARCH: string[] = [".txt", ".md", ".markdown", ".docx", ".xlsx", ".pptx", ".pdf", ".json", ".log", ".sql", ".html", ".xml", ".java", ".h", ".c",
        ".cpp", ".go", ".rs", ".swift", ".kt", ".py", ".php", ".js", ".css", ".ts", ".sh", ".bat", ".cmd", ".ini", ".yaml",
        ".rst", ".adoc", ".textile", ".opml", ".org", ".wiki", ".epub", ".cs"];

    // protyle
    public static readonly SCRIBLI_CONFIG_APPEARANCE_DARK_CODE: string[] = ["a11y-dark", "agate", "an-old-hope", "androidstudio",
        "arta", "atom-one-dark", "atom-one-dark-reasonable", "base16/3024", "base16/apathy", "base16/apprentice", "base16/ashes",
        "base16/atelier-cave", "base16/atelier-dune", "base16/atelier-estuary", "base16/atelier-forest", "base16/atelier-heath",
        "base16/atelier-lakeside", "base16/atelier-plateau", "base16/atelier-savanna", "base16/atelier-seaside", "base16/atelier-sulphurpool",
        "base16/atlas", "base16/bespin", "base16/black-metal", "base16/black-metal-bathory", "base16/black-metal-burzum",
        "base16/black-metal-dark-funeral", "base16/black-metal-gorgoroth", "base16/black-metal-immortal", "base16/black-metal-khold",
        "base16/black-metal-marduk", "base16/black-metal-mayhem", "base16/black-metal-nile", "base16/black-metal-venom",
        "base16/brewer", "base16/bright", "base16/brogrammer", "base16/brush-trees-dark", "base16/chalk", "base16/circus",
        "base16/classic-dark", "base16/codeschool", "base16/colors", "base16/danqing", "base16/darcula", "base16/dark-violet",
        "base16/darkmoss", "base16/darktooth", "base16/decaf", "base16/default-dark", "base16/dracula", "base16/edge-dark",
        "base16/eighties", "base16/embers", "base16/equilibrium-dark", "base16/equilibrium-gray-dark", "base16/espresso",
        "base16/eva", "base16/eva-dim", "base16/flat", "base16/framer", "base16/gigavolt", "base16/google-dark", "base16/grayscale-dark",
        "base16/green-screen", "base16/gruvbox-dark-hard", "base16/gruvbox-dark-medium", "base16/gruvbox-dark-pale", "base16/gruvbox-dark-soft",
        "base16/hardcore", "base16/harmonic16-dark", "base16/heetch-dark", "base16/helios", "base16/hopscotch", "base16/horizon-dark",
        "base16/humanoid-dark", "base16/ia-dark", "base16/icy-dark", "base16/ir-black", "base16/isotope", "base16/kimber",
        "base16/london-tube", "base16/macintosh", "base16/marrakesh", "base16/materia", "base16/material", "base16/material-darker",
        "base16/material-palenight", "base16/material-vivid", "base16/mellow-purple", "base16/mocha", "base16/monokai",
        "base16/nebula", "base16/nord", "base16/nova", "base16/ocean", "base16/oceanicnext", "base16/onedark", "base16/outrun-dark",
        "base16/papercolor-dark", "base16/paraiso", "base16/pasque", "base16/phd", "base16/pico", "base16/pop", "base16/porple",
        "base16/qualia", "base16/railscasts", "base16/rebecca", "base16/ros-pine", "base16/ros-pine-moon", "base16/sandcastle",
        "base16/seti-ui", "base16/silk-dark", "base16/snazzy", "base16/solar-flare", "base16/solarized-dark", "base16/spacemacs",
        "base16/summercamp", "base16/summerfruit-dark", "base16/synth-midnight-terminal-dark", "base16/tango", "base16/tender",
        "base16/tomorrow-night", "base16/twilight", "base16/unikitty-dark", "base16/vulcan", "base16/windows-10", "base16/windows-95",
        "base16/windows-high-contrast", "base16/windows-nt", "base16/woodland", "base16/xcode-dusk", "base16/zenburn", "codepen-embed",
        "cybertopia-cherry", "cybertopia-dimmer", "cybertopia-icecap", "cybertopia-saturated", "dark", "devibeans", "far",
        "felipec", "github-dark", "github-dark-dimmed", "gml", "gradient-dark", "hybrid", "ir-black", "isbl-editor-dark",
        "kimbie-dark", "lioshi", "monokai", "monokai-sublime", "night-owl", "nnfx-dark", "nord", "obsidian", "panda-syntax-dark",
        "paraiso-dark", "pojoaque", "qtcreator-dark", "rainbow", "rose-pine", "rose-pine-moon", "shades-of-purple", "srcery",
        "stackoverflow-dark", "sunburst", "tomorrow-night-blue", "tomorrow-night-bright", "tokyo-night-dark", "vs-dark", "vs2015", "xt256"
    ];
    public static readonly SCRIBLI_CONFIG_APPEARANCE_LIGHT_CODE: string[] = ["ant-design",
        "1c-light", "a11y-light", "arduino-light", "ascetic", "atom-one-light", "base16/atelier-cave-light", "base16/atelier-dune-light",
        "base16/atelier-estuary-light", "base16/atelier-forest-light", "base16/atelier-heath-light", "base16/atelier-lakeside-light",
        "base16/atelier-plateau-light", "base16/atelier-savanna-light", "base16/atelier-seaside-light", "base16/atelier-sulphurpool-light",
        "base16/brush-trees", "base16/classic-light", "base16/cupcake", "base16/cupertino", "base16/default-light", "base16/dirtysea",
        "base16/edge-light", "base16/equilibrium-gray-light", "base16/equilibrium-light", "base16/fruit-soda", "base16/github",
        "base16/google-light", "base16/grayscale-light", "base16/gruvbox-light-hard", "base16/gruvbox-light-medium",
        "base16/gruvbox-light-soft", "base16/harmonic16-light", "base16/heetch-light", "base16/humanoid-light", "base16/horizon-light",
        "base16/ia-light", "base16/material-lighter", "base16/mexico-light", "base16/one-light", "base16/papercolor-light",
        "base16/ros-pine-dawn", "base16/sagelight", "base16/shapeshifter", "base16/silk-light", "base16/solar-flare-light",
        "base16/solarized-light", "base16/summerfruit-light", "base16/synth-midnight-terminal-light", "base16/tomorrow",
        "base16/unikitty-light", "base16/windows-10-light", "base16/windows-95-light", "base16/windows-high-contrast-light",
        "brown-paper", "base16/windows-nt-light", "color-brewer", "docco", "foundation", "github", "googlecode", "gradient-light",
        "grayscale", "idea", "intellij-light", "isbl-editor-light", "kimbie-light", "lightfair", "magula", "mono-blue",
        "nnfx-light", "panda-syntax-light", "paraiso-light", "purebasic", "qtcreator-light", "rose-pine-dawn", "routeros",
        "school-book", "stackoverflow-light", "tokyo-night-light", "vs", "xcode", "default"];
    public static readonly ZWSP: string = "\u200b";
    public static readonly INLINE_TYPE: string[] = ["block-ref", "kbd", "text", "file-annotation-ref", "a", "strong", "em", "u", "s", "mark", "sup", "sub", "tag", "code", "inline-math", "inline-memo", "clear"];
    public static readonly BLOCK_HINT_KEYS: string[] = ["((", "[[", "（（", "【【"];
    public static readonly BLOCK_HINT_CLOSE_KEYS: Record<string, string> = {"((": "))", "[[": "]]", "（（": "））", "【【": "】】"};
    // common: "bash", "c", "csharp", "cpp", "css", "diff", "go", "xml", "json", "java", "javascript", "kotlin", "less", "lua", "makefile", "markdown", "objectivec", "php", "php-template", "perl", "plaintext", "python", "python-repl", "r", "ruby", "rust", "scss", "sql", "shell", "swift", "ini", "typescript", "vbnet", "yaml", "properties", "1c", "armasm", "avrasm", "actionscript", "ada", "angelscript", "accesslog", "apache", "applescript", "arcade", "arduino", "asciidoc", "aspectj", "abnf", "autohotkey", "autoit", "awk", "basic", "bnf", "dos", "brainfuck", "cal", "cmake", "csp", "cos", "capnproto", "ceylon", "clean", "clojure", "clojure-repl", "coffeescript", "coq", "crystal", "d", "dns", "dart", "delphi", "dts", "django", "dockerfile", "dust", "erb", "elixir", "elm", "erlang", "erlang-repl", "excel", "ebnf", "fsharp", "fix", "flix", "fortran", "gcode", "gams", "gauss", "glsl", "gml", "gherkin", "golo", "gradle", "groovy", "haml", "hsp", "http", "handlebars", "haskell", "haxe", "hy", "irpf90", "isbl", "inform7", "x86asm", "jboss-cli", "julia", "julia-repl", "ldif", "llvm", "lsl", "latex", "lasso", "leaf", "lisp", "livecodeserver", "livescript", "mel", "mipsasm", "matlab", "maxima", "mercury", "axapta", "routeros", "mizar", "mojolicious", "monkey", "moonscript", "n1ql", "nsis", "nestedtext", "nginx", "nim", "nix", "node-repl", "ocaml", "openscad", "ruleslanguage", "oxygene", "pf", "parser3", "pony", "pgsql", "powershell", "processing", "prolog", "protobuf", "puppet", "purebasic", "profile", "q", "qml", "reasonml", "rib", "rsl", "roboconf", "sas", "sml", "sqf", "step21", "scala", "scheme", "scilab", "smali", "smalltalk", "stan", "stata", "stylus", "subunit", "tp", "taggerscript", "tcl", "tap", "thrift", "twig", "vbscript", "vbscript-html", "vhdl", "vala", "verilog", "vim", "wasm", "mathematica", "wren", "xl", "xquery", "zephir", "crmsh", "dsconfig", "graphql",
    // third: "yul", "solidity", "abap", "hlsl", "gdscript", "moonbit", "mlir"
    public static readonly ALIAS_CODE_LANGUAGES: string[] = [
        "js", "ts", "html", "toml", "c#", "bat"
    ];
    public static readonly SCRIBLI_RENDER_CODE_LANGUAGES: string[] = [
        "abc", "plantuml", "mermaid", "flowchart", "echarts", "mindmap", "graphviz", "math"
    ];
}
