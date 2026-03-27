package reactions

import (
	"fmt"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"

	"video-downloader-bot/internal/logger"
)

type Reaction string

const ( // https://core.telegram.org/bots/api#reactiontypeemoji
	ReactionHeart          Reaction = "❤"   // ❤️
	ReactionThumbsUp       Reaction = "👍"   // 👍
	ReactionThumbsDown     Reaction = "👎"   // 👎
	ReactionFire           Reaction = "🔥"   // 🔥
	ReactionSmilingHearts  Reaction = "🥰"   // 🥰
	ReactionClap           Reaction = "👏"   // 👏
	ReactionGrin           Reaction = "😁"   // 😁
	ReactionThinking       Reaction = "🤔"   // 🤔
	ReactionMindBlown      Reaction = "🤯"   // 🤯
	ReactionScream         Reaction = "😱"   // 😱
	ReactionSwearing       Reaction = "🤬"   // 🤬
	ReactionCry            Reaction = "😢"   // 😢
	ReactionParty          Reaction = "🎉"   // 🎉
	ReactionStarStruck     Reaction = "🤩"   // 🤩
	ReactionVomit          Reaction = "🤮"   // 🤮
	ReactionPoop           Reaction = "💩"   // 💩
	ReactionPray           Reaction = "🙏"   // 🙏
	ReactionOK             Reaction = "👌"   // 👌
	ReactionDove           Reaction = "🕊"   // 🕊️
	ReactionClown          Reaction = "🤡"   // 🤡
	ReactionYawn           Reaction = "🥱"   // 🥱
	ReactionWoozy          Reaction = "🥴"   // 🥴
	ReactionHeartEyes      Reaction = "😍"   // 😍
	ReactionWhale          Reaction = "🐳"   // 🐳
	ReactionHeartOnFire    Reaction = "❤‍🔥" // ❤️‍🔥
	ReactionMoonFace       Reaction = "🌚"   // 🌚
	ReactionHotDog         Reaction = "🌭"   // 🌭
	ReactionHundred        Reaction = "💯"   // 💯
	ReactionRollingLaugh   Reaction = "🤣"   // 🤣
	ReactionLightning      Reaction = "⚡"   // ⚡
	ReactionBanana         Reaction = "🍌"   // 🍌
	ReactionTrophy         Reaction = "🏆"   // 🏆
	ReactionBrokenHeart    Reaction = "💔"   // 💔
	ReactionRaisedEyebrow  Reaction = "🤨"   // 🤨
	ReactionNeutralFace    Reaction = "😐"   // 😐
	ReactionStrawberry     Reaction = "🍓"   // 🍓
	ReactionChampagne      Reaction = "🍾"   // 🍾
	ReactionKiss           Reaction = "💋"   // 💋
	ReactionMiddleFinger   Reaction = "🖕"   // 🖕
	ReactionSmilingDevil   Reaction = "😈"   // 😈
	ReactionSleep          Reaction = "😴"   // 😴
	ReactionLoudCry        Reaction = "😭"   // 😭
	ReactionNerd           Reaction = "🤓"   // 🤓
	ReactionGhost          Reaction = "👻"   // 👻
	ReactionTechnologist   Reaction = "👨‍💻" // 👨‍💻
	ReactionEyes           Reaction = "👀"   // 👀
	ReactionPumpkin        Reaction = "🎃"   // 🎃
	ReactionSeeNoEvil      Reaction = "🙈"   // 🙈
	ReactionInnocent       Reaction = "😇"   // 😇
	ReactionFear           Reaction = "😨"   // 😨
	ReactionHandshake      Reaction = "🤝"   // 🤝
	ReactionWritingHand    Reaction = "✍"   // ✍️
	ReactionHug            Reaction = "🤗"   // 🤗
	ReactionSalute         Reaction = "🫡"   // 🫡
	ReactionSanta          Reaction = "🎅"   // 🎅
	ReactionChristmasTree  Reaction = "🎄"   // 🎄
	ReactionSnowman        Reaction = "☃"   // ☃️
	ReactionNailPolish     Reaction = "💅"   // 💅
	ReactionZany           Reaction = "🤪"   // 🤪
	ReactionMoai           Reaction = "🗿"   // 🗿
	ReactionCool           Reaction = "🆒"   // 🆒
	ReactionHeartWithArrow Reaction = "💘"   // 💘
	ReactionHearNoEvil     Reaction = "🙉"   // 🙉
	ReactionUnicorn        Reaction = "🦄"   // 🦄
	ReactionBlowKiss       Reaction = "😘"   // 😘
	ReactionPill           Reaction = "💊"   // 💊
	ReactionSpeakNoEvil    Reaction = "🙊"   // 🙊
	ReactionSunglasses     Reaction = "😎"   // 😎
	ReactionAlienMonster   Reaction = "👾"   // 👾
	ReactionShrugMan       Reaction = "🤷‍♂" // 🤷‍♂️
	ReactionShrug          Reaction = "🤷"   // 🤷
	ReactionShrugWoman     Reaction = "🤷‍♀" // 🤷‍♀️
	ReactionAngry          Reaction = "😡"   // 😡
)

func (r Reaction) String() string {
	return string(r)
}

func MessageReactionOpts(reaction Reaction) *gotgbot.SetMessageReactionOpts {
	return &gotgbot.SetMessageReactionOpts{Reaction: []gotgbot.ReactionType{gotgbot.ReactionTypeEmoji{Emoji: reaction.String()}}}
}

func React(bot *gotgbot.Bot, ctx *ext.Context, reaction Reaction) error {
	if ctx.EffectiveMessage == nil {
		return nil
	}

	_, err := bot.SetMessageReaction(ctx.EffectiveMessage.Chat.Id, ctx.EffectiveMessage.MessageId, MessageReactionOpts(reaction))
	if err != nil {
		return fmt.Errorf("failed to set reaction: %w", err)
	}

	return nil
}

func ReactWithTimer(bot *gotgbot.Bot, ctx *ext.Context, reaction Reaction, sleep time.Duration) error {
	if ctx.EffectiveMessage == nil {
		return nil
	}

	_, err := bot.SetMessageReaction(ctx.EffectiveMessage.Chat.Id, ctx.EffectiveMessage.MessageId, MessageReactionOpts(reaction))
	if err != nil {
		logger.Error("failed to set reaction: %v", err)
		return nil
	}
	go func() {
		time.Sleep(sleep * time.Second)
		err = RemoveReaction(bot, ctx)
		if err != nil {
			logger.Error("failed to remove reaction: %v", err)
		}
	}()

	return nil
}

func RemoveReaction(bot *gotgbot.Bot, ctx *ext.Context) error {
	if ctx.EffectiveMessage == nil {
		return nil
	}

	_, err := bot.SetMessageReaction(ctx.EffectiveMessage.Chat.Id, ctx.EffectiveMessage.MessageId, &gotgbot.SetMessageReactionOpts{})
	if err != nil {
		return fmt.Errorf("failed to remove reaction: %w", err)
	}

	return nil
}
