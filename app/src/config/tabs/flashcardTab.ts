import type {SettingTabBuilder} from "../setting/builder";

const registerFlashcardCreationGroup = (tab: SettingTabBuilder) => {
    const group = tab.group("creation", window.scribli.languages.configGroupCardCreation);

    group.switch("flashcard.mark", {
        title: window.scribli.languages.flashcardMark,
        desc: window.scribli.languages.flashcardMarkTip,
    });
    group.switch("flashcard.list", {
        title: window.scribli.languages.flashcardList,
        desc: window.scribli.languages.flashcardListTip,
    });
    group.switch("flashcard.heading", {
        title: window.scribli.languages.flashcardHeading,
        desc: window.scribli.languages.flashcardHeadingTip,
    });
    group.switch("flashcard.superBlock", {
        title: window.scribli.languages.flashcardSuperBlock,
        desc: window.scribli.languages.flashcardSuperBlockTip,
    });
};

const registerFlashcardReviewGroup = (tab: SettingTabBuilder) => {
    const group = tab.group("review", window.scribli.languages.configGroupReview);

    group.select("flashcard.reviewMode", {
        title: window.scribli.languages.reviewMode,
        desc: window.scribli.languages.reviewModeTip,
        options: [
            {value: 0, label: window.scribli.languages.reviewMode0},
            {value: 1, label: window.scribli.languages.reviewMode1},
            {value: 2, label: window.scribli.languages.reviewMode2},
        ],
    });
    group.number("flashcard.newCardLimit", {
        title: window.scribli.languages.flashcardNewCardLimit,
        desc: window.scribli.languages.flashcardNewCardLimitTip,
        min: 0,
    });
    group.number("flashcard.reviewCardLimit", {
        title: window.scribli.languages.flashcardReviewCardLimit,
        desc: window.scribli.languages.flashcardReviewCardLimitTip,
        min: 0,
    });
    group.number("flashcard.requestRetention", {
        title: window.scribli.languages.flashcardFSRSParamRequestRetention,
        desc: window.scribli.languages.flashcardFSRSParamRequestRetentionTip,
        min: 0,
        max: 1,
        step: "0.01",
    });
    group.number("flashcard.maximumInterval", {
        title: window.scribli.languages.flashcardFSRSParamMaximumInterval,
        desc: window.scribli.languages.flashcardFSRSParamMaximumIntervalTip,
        min: 1,
        max: 36500,
    });
    group.textBlock("flashcard.weights", {
        title: window.scribli.languages.flashcardFSRSParamWeights,
        desc: window.scribli.languages.flashcardFSRSParamWeightsTip,
        mode: "input-text",
    });
};

const registerFlashcardOthersGroup = (tab: SettingTabBuilder) => {
    const group = tab.group("others", window.scribli.languages.configGroupOthers);

    group.switch("flashcard.deck", {
        title: window.scribli.languages.flashcardDeck,
        desc: window.scribli.languages.flashcardDeckTip,
    });
};

export const registerFlashcardTab = (tab: SettingTabBuilder) => {
    registerFlashcardCreationGroup(tab);
    registerFlashcardReviewGroup(tab);
    registerFlashcardOthersGroup(tab);
};
