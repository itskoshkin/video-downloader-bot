package telegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/spf13/viper"

	"video-downloader-bot/internal/config"
	"video-downloader-bot/internal/logger"
	"video-downloader-bot/internal/models"
	"video-downloader-bot/internal/telegram/helpers/errors"
	"video-downloader-bot/internal/telegram/helpers/keyboards"
	"video-downloader-bot/internal/telegram/helpers/storage"
	req "video-downloader-bot/internal/telegram/middlewares/requests"
	s "video-downloader-bot/internal/telegram/strings"
	"video-downloader-bot/internal/utils/errs"
	"video-downloader-bot/internal/utils/links"
)

// A job row lives while a link is being processed, so a request that never finished can be offered a retry:
// - the handler itself stops at processingTimeout and shows the error with a retry button
// - a row still unfinished after processingTimeout+jobGrace means its process died, and the sweeper attaches the button
const (
	retryPrefix        = "retry:"
	jobGrace           = time.Minute      // Lets a live handler hit its own deadline first, so the sweeper only picks up dead ones
	jobSweepInterval   = time.Minute      // How often the sweeper looks for stuck jobs
	jobRetention       = 48 * time.Hour   // Stale rows are kept this long so their retry buttons keep working
	jobDBTimeout       = 5 * time.Second  // Job bookkeeping must not block on a slow DB, and must work after the request deadline
	defaultProcTimeout = 10 * time.Minute // Used when app.telegram.bot.processing_timeout is not set
)

func retryData(id uint) string { return fmt.Sprintf("%s%d", retryPrefix, id) }

func processingTimeout() time.Duration {
	if sec := viper.GetInt(config.TelegramBotProcessingTimeout); sec > 0 {
		return time.Duration(sec) * time.Second
	}
	return defaultProcTimeout
}

// startJob records a link that is about to be processed; returns nil when the row can't be saved (processing goes on without a retry button)
func (b *Bot) startJob(reqCtx context.Context, job *models.Job) *models.Job {
	ctx, cancel := context.WithTimeout(context.Background(), jobDBTimeout)
	defer cancel()
	if err := b.jobs.Create(ctx, job); err != nil {
		logger.WarnWithID(reqCtx, "Failed to record job, no retry button for it: %v", err)
		return nil
	}
	return job
}

// endJob removes the row once the result or a regular error is shown; a stale row stays for its retry button
func (b *Bot) endJob(reqCtx context.Context, job *models.Job) {
	if job == nil || job.StaleAt != nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), jobDBTimeout)
	defer cancel()
	if _, err := b.jobs.Delete(ctx, job.ID); err != nil {
		logger.WarnWithID(reqCtx, "Failed to remove finished job: %v", err)
	}
}

func (b *Bot) markStale(reqCtx context.Context, job *models.Job) {
	now := time.Now()
	job.StaleAt = &now
	ctx, cancel := context.WithTimeout(context.Background(), jobDBTimeout)
	defer cancel()
	if err := b.jobs.MarkStale(ctx, job.ID); err != nil {
		logger.WarnWithID(reqCtx, "Failed to mark job stale: %v", err)
	}
}

// fail reports a processing error; when the processing deadline was hit it says so and adds the retry button
func (b *Bot) fail(bot *gotgbot.Bot, ctx *ext.Context, lang string, reqCtx context.Context, job *models.Job, err error) error {
	if job == nil || reqCtx.Err() != context.DeadlineExceeded {
		return errors.HandleError(bot, ctx, lang, s.Lang(lang).FailedToProcessLink, err)
	}
	b.markStale(reqCtx, job)
	ctx.Data["retry_data"] = retryData(job.ID)
	// Flatten the error: the killed tool's partial stderr (download progress and such) is not worth showing to the user
	return errors.HandleError(bot, ctx, lang, s.Lang(lang).ProcessingTimedOut, fmt.Errorf("processing timed out after %v: %s", processingTimeout(), errs.ShortError(err)))
}

// StartJobSweeper attaches the retry button to requests left hanging by a crash or restart; runs one sweep now, then every jobSweepInterval
func (b *Bot) StartJobSweeper() {
	go func() {
		b.sweepJobs()
		ticker := time.NewTicker(jobSweepInterval)
		defer ticker.Stop()
		for range ticker.C {
			b.sweepJobs()
		}
	}()
}

func (b *Bot) sweepJobs() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	stuck, err := b.jobs.ListStuck(ctx, time.Now().Add(-processingTimeout()-jobGrace))
	if err != nil {
		logger.Warn("jobs: %v", err)
		return
	}
	for _, job := range stuck {
		opts := &gotgbot.EditMessageReplyMarkupOpts{
			ReplyMarkup: gotgbot.InlineKeyboardMarkup{InlineKeyboard: [][]gotgbot.InlineKeyboardButton{keyboards.GetRetryButton(job.Lang, retryData(job.ID))}},
		}
		if job.InlineMessageID != "" {
			opts.InlineMessageId = job.InlineMessageID
		} else {
			opts.ChatId, opts.MessageId = job.ChatID, job.MessageID
		}
		if _, _, err = b.bot.EditMessageReplyMarkup(opts); err != nil {
			logger.Warn("jobs: failed to attach retry button to job %d (%s): %v", job.ID, job.Link, err) // Message deleted or already replaced
		} else {
			logger.Info("jobs: attached retry button to job %d (%s), its processing never finished", job.ID, job.Link)
		}
		if err = b.jobs.MarkStale(ctx, job.ID); err != nil { // Mark either way, so a message that can't be edited isn't retried every minute
			logger.Warn("jobs: %v", err)
		}
	}

	if err = b.jobs.DeleteOlderThan(ctx, time.Now().Add(-jobRetention)); err != nil {
		logger.Warn("jobs: %v", err)
	}
}

// RetryCallback handles the "Didn't work? Retry" button: it resets the message to "downloading" and processes the same link again in place.
// Callback data is "retry:<job id>".
func (b *Bot) RetryCallback(bot *gotgbot.Bot, ctx *ext.Context) error {
	reqCtx := req.FromExtContext(ctx)
	callback := ctx.CallbackQuery

	lang := s.GetUserLanguageCode(ctx)
	settings, err := b.settings.GetOrCreate(reqCtx, callback.From.Id)
	if err == nil {
		lang = settings.Language
	}

	id, err := strconv.ParseUint(strings.TrimPrefix(callback.Data, retryPrefix), 10, 64)
	if err != nil {
		_, _ = callback.Answer(bot, nil)
		return nil
	}
	job, err := b.jobs.Get(reqCtx, uint(id))
	if err != nil {
		_, _ = callback.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{Text: s.Lang(lang).RetryExpired, ShowAlert: true})
		return nil
	}

	if !b.rateLimiter.Allow(callback.From.Id) {
		logger.WarnWithID(reqCtx, "Rate limited %s", storage.GetUserString(&callback.From))
		_, _ = callback.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{Text: s.Lang(lang).RateLimited})
		return nil
	}

	// Delete before starting: only the click that actually removed the row gets to retry, so a double tap can't run it twice
	if deleted, dErr := b.jobs.Delete(reqCtx, job.ID); dErr != nil || !deleted {
		_, _ = callback.Answer(bot, nil)
		return nil
	}
	_, _ = callback.Answer(bot, nil)

	logger.InfoWithID(reqCtx, "Retrying \"%s\" for %s...", job.Link, storage.GetUserString(&callback.From))

	if callback.InlineMessageId != "" {
		ctx.Data["inline_message_id"] = callback.InlineMessageId
		ctx.Data["inline_link"] = links.DetrackLink(job.Link)
		if _, _, err = bot.EditMessageText(s.Lang(lang).InlineProcessingNotice+"\n\n"+links.DetrackLink(job.Link), &gotgbot.EditMessageTextOpts{
			InlineMessageId: callback.InlineMessageId,
			ReplyMarkup:     *keyboards.GetInlinePlaceholderButton(lang),
		}); err != nil {
			logger.WarnWithID(reqCtx, "Failed to reset the inline placeholder: %v", err)
		}
		return b.processInlineLink(bot, ctx, job.Link, callback.InlineMessageId, &callback.From)
	}

	status := ctx.EffectiveMessage // The status/error message carrying the button
	if status == nil {
		return nil
	}
	if _, _, err = bot.EditMessageText(s.Lang(lang).Downloading, &gotgbot.EditMessageTextOpts{ChatId: status.Chat.Id, MessageId: status.MessageId}); err != nil {
		logger.WarnWithID(reqCtx, "Failed to reset the status message: %v", err)
	}
	ctx.Data["status_message_id"] = status.MessageId
	ctx.Data["status_chat_id"] = status.Chat.Id
	if status.ReplyToMessage != nil {
		ctx.EffectiveMessage = status.ReplyToMessage // The user's original message: reactions and the video go there, as on the first try
	}
	return b.processChatLink(bot, ctx, lang, settings, job.Link)
}
