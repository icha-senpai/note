let luteInstance: Lute | undefined;

/**
 *
 */
export const getLute = (options: ILuteOptions): Lute => {
    if (!luteInstance) {
        luteInstance = setLute(options);
    }
    return luteInstance;
};

/**
 */
export const getLuteInstance = (): Lute | undefined => {
    return luteInstance;
};

/**
 *
 */
export const getAgentLute = (options: ILuteOptions): Lute => {
    const lute: Lute = Lute.New();
    lute.SetSpellcheck(false);
    lute.SetProtyleMarkNetImg(false);
    lute.SetFileAnnotationRef(true);
    lute.SetHTMLTag2TextMark(true);
    lute.SetTextMark(true);
    lute.SetHeadingID(false);
    lute.SetYamlFrontMatter(false);
    lute.PutEmojis(options.emojis);
    lute.SetEmojiSite(options.emojiSite);
    lute.SetHeadingAnchor(options.headingAnchor);
    lute.SetInlineMathAllowDigitAfterOpenMarker(true);
    lute.SetToC(false);
    lute.SetIndentCodeBlock(false);
    lute.SetParagraphBeginningSpace(true);
    lute.SetSetext(false);
    lute.SetFootnotes(false);
    lute.SetLinkRef(false);
    lute.SetSanitize(options.sanitize);
    lute.SetChineseParagraphBeginningSpace(options.paragraphBeginningSpace);
    lute.SetRenderListStyle(options.listStyle);
    lute.SetImgPathAllowSpace(true);
    lute.SetKramdownIAL(true);
    lute.SetSuperBlock(true);
    lute.SetCallout(true);
    lute.SetInlineAsterisk(true);
    lute.SetInlineUnderscore(true);
    lute.SetSup(true);
    lute.SetSub(true);
    lute.SetTag(true);
    lute.SetInlineMath(true);
    lute.SetGFMStrikethrough1(false);
    lute.SetGFMStrikethrough(true);
    lute.SetMark(true);
    lute.SetSpin(true);
    lute.SetProtyleWYSIWYG(true);
    if (options.lazyLoadImage) {
        lute.SetImageLazyLoading(options.lazyLoadImage);
    }
    lute.SetBlockRef(true);
    lute.SetUnorderedListMarker("-");
    lute.SetDataTask(true);
    lute.SetExportNormalizeTaskListMarker(true);
    lute.SetArbitraryTaskListItemMarker(true);
    lute.SetEnsureListItemParagraph(true);
    return lute;
};

/**
 */
const setLute = (options: ILuteOptions) => {
    const lute: Lute = Lute.New();
    lute.SetSpellcheck(window.scribli.config.editor.spellcheck);
    lute.SetProtyleMarkNetImg(window.scribli.config.editor.displayNetImgMark);
    lute.SetFileAnnotationRef(true);
    lute.SetHTMLTag2TextMark(true);
    lute.SetTextMark(true);
    lute.SetHeadingID(false);
    lute.SetYamlFrontMatter(false);
    lute.PutEmojis(options.emojis);
    lute.SetEmojiSite(options.emojiSite);
    lute.SetHeadingAnchor(options.headingAnchor);
    lute.SetInlineMathAllowDigitAfterOpenMarker(true);
    lute.SetToC(false);
    lute.SetIndentCodeBlock(false);
    lute.SetParagraphBeginningSpace(true);
    lute.SetSetext(false);
    lute.SetFootnotes(false);
    lute.SetLinkRef(false);
    lute.SetSanitize(options.sanitize);
    lute.SetChineseParagraphBeginningSpace(options.paragraphBeginningSpace);
    lute.SetRenderListStyle(options.listStyle);
    lute.SetImgPathAllowSpace(true);
    lute.SetKramdownIAL(true);
    lute.SetTag(true);
    lute.SetSuperBlock(true);
    lute.SetCallout(true);
    lute.SetInlineAsterisk(window.scribli.config.editor.markdown.inlineAsterisk);
    lute.SetInlineUnderscore(window.scribli.config.editor.markdown.inlineUnderscore);
    lute.SetSup(window.scribli.config.editor.markdown.inlineSup);
    lute.SetSub(window.scribli.config.editor.markdown.inlineSub);
    lute.SetTag(window.scribli.config.editor.markdown.inlineTag);
    lute.SetInlineMath(window.scribli.config.editor.markdown.inlineMath);
    lute.SetGFMStrikethrough1(false);
    lute.SetGFMStrikethrough(window.scribli.config.editor.markdown.inlineStrikethrough);
    lute.SetMark(window.scribli.config.editor.markdown.inlineMark);
    lute.SetSpin(true);
    lute.SetProtyleWYSIWYG(true);
    if (options.lazyLoadImage) {
        lute.SetImageLazyLoading(options.lazyLoadImage);
    }
    lute.SetBlockRef(true);
    if (window.scribli.emojis[0].items.length > 0) {
        const emojis: IObject = {};
        window.scribli.emojis[0].items.forEach(item => {
            emojis[item.keywords] = options.emojiSite + "/" + item.unicode;
        });
        lute.PutEmojis(emojis);
    }
    lute.SetUnorderedListMarker("-");
    lute.SetDataTask(true);
    lute.SetExportNormalizeTaskListMarker(true);
    lute.SetArbitraryTaskListItemMarker(true);
    lute.SetEnsureListItemParagraph(true);
    return lute;
};
