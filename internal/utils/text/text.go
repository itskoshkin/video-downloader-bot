package text

import (
	"video-downloader-bot/internal/utils/colors"
)

func Red(text string) string { return colors.Red() + text + colors.Reset() }

func Orange(text string) string { return colors.Orange() + text + colors.Reset() }

func Yellow(text string) string { return colors.Yellow() + text + colors.Reset() }

func Green(text string) string { return colors.Green() + text + colors.Reset() }

func Cyan(text string) string { return colors.Cyan() + text + colors.Reset() }

func Sky(text string) string { return colors.Sky() + text + colors.Reset() }

func Blue(text string) string { return colors.Blue() + text + colors.Reset() }

func Purple(text string) string { return colors.Purple() + text + colors.Reset() }

func Magenta(text string) string { return colors.Magenta() + text + colors.Reset() }

func White(text string) string { return colors.White() + text + colors.Reset() }

func Gray(text string) string { return colors.Gray() + text + colors.Reset() }

func Black(text string) string { return colors.Black() + text + colors.Reset() }

func Bold(text string) string { return colors.Bold() + text + colors.Reset() }

func Italic(text string) string { return colors.Italic() + text + colors.Reset() }

func Underline(text string) string { return colors.Underline() + text + colors.Reset() }

func Strikethrough(text string) string { return colors.Strikethrough() + text + colors.Reset() }

func Background(text string) string { return colors.Background() + text + colors.Reset() }
