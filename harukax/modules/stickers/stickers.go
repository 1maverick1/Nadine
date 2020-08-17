/*
 *    Copyright © 2020 Haruka Network Development
 *    This file is part of Haruka X.
 *
 *    Haruka X is free software: you can redistribute it and/or modify
 *    it under the terms of the Raphielscape Public License as published by
 *    the Devscapes Open Source Holding GmbH., version 1.d
 *
 *    Haruka X is distributed in the hope that it will be useful,
 *    but WITHOUT ANY WARRANTY; without even the implied warranty of
 *    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *    Devscapes Raphielscape Public License for more details.
 *
 *    You should have received a copy of the Devscapes Raphielscape Public License
 */

package stickers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/HarukaNetwork/HarukaX/harukax"
	"github.com/PaulSonOfLars/gotgbot"
	"github.com/PaulSonOfLars/gotgbot/ext"
	"github.com/PaulSonOfLars/gotgbot/handlers"
	"go.uber.org/zap"
)

func stickerId(_ ext.Bot, u *gotgbot.Update) error {
	msg := u.EffectiveMessage
	if msg.ReplyToMessage != nil && msg.ReplyToMessage.Sticker != nil {
		msg.ReplyHTMLf("Sticker  ID:\n<code>%v</code>", msg.ReplyToMessage.Sticker.FileId)
	} else {
		msg.ReplyText("Sticker ID not found.")
	}
	return nil
}

func getSticker(bot ext.Bot, u *gotgbot.Update) error {
	msg := u.EffectiveMessage
	chat := u.EffectiveChat

	if msg.ReplyToMessage != nil && msg.ReplyToMessage.Sticker != nil && msg.ReplyToMessage.Sticker.IsAnimated == false {
		fileId := msg.ReplyToMessage.Sticker.FileId

		inputFile, r, err := getInputFile(bot, fileId, "sticker.png")
		if r != nil {
			defer r.Body.Close()
		}
		if err != nil {
			return err
		}

		newDoc := bot.NewSendableDocument(chat.Id, "Sticker")
		newDoc.Document = inputFile
		newDoc.Send()
	} else if msg.ReplyToMessage != nil && msg.ReplyToMessage.Sticker != nil && msg.ReplyToMessage.Sticker.IsAnimated == true {
		fileId := msg.ReplyToMessage.Sticker.FileId

		inputFile, r, err := getInputFile(bot, fileId, "sticker.rename")
		if r != nil {
			defer r.Body.Close()
		}
		if err != nil {
			return err
		}

		newDoc := bot.NewSendableDocument(chat.Id, "Go to @Stickers bot and rename this file to .tgs then use "+
			"/newanimated or /addsticker and send this file")
		newDoc.Document = inputFile
		newDoc.Send()
	} else {
		msg.ReplyText("Please reply to a sticker for me to upload its PNG.")
	}
	return nil
}

func kangSticker(bot ext.Bot, u *gotgbot.Update) error {
	msg := u.EffectiveMessage
	user := u.EffectiveUser
	packnum := 0
	packname := fmt.Sprintf("a%v_by_%v", strconv.Itoa(user.Id), bot.UserName)
	if msg.ReplyToMessage == nil {
		msg.ReplyText("What are you trying to kang?")
		var err error
		return err
	}
	if msg.ReplyToMessage.Sticker == nil {
		msg.ReplyText("Can't kang that mate.")
		var err error
		return err
	}
	if msg.ReplyToMessage.Sticker.IsAnimated == true {
		packname = fmt.Sprintf("b%v_by_%v", strconv.Itoa(user.Id), bot.UserName)
	}
	packnameFound := 0
	maxStickers := 120
	for packnameFound == 0 {
		if msg.ReplyToMessage.Sticker.IsAnimated == true {
			stickerset, err := bot.GetStickerSet(packname)

			if err != nil {
				packnameFound = 1
				break
			}

			if len(stickerset.Stickers) >= maxStickers {
				packnum++
				packname = fmt.Sprintf("b%v_%v_by_%v", strconv.Itoa(packnum), strconv.Itoa(user.Id), bot.UserName)
			} else {
				packnameFound = 1
			}
		} else {
			stickerset, err := bot.GetStickerSet(packname)

			if err != nil {
				packnameFound = 1
				break
			}

			if len(stickerset.Stickers) >= maxStickers {
				packnum++
				packname = fmt.Sprintf("a%v_%v_by_%v", strconv.Itoa(packnum), strconv.Itoa(user.Id), bot.UserName)
			} else {
				packnameFound = 1
			}
		}
	}
	if msg.ReplyToMessage != nil {
		var fileId string
		var stickerEmoji string
		var success bool
		var err error
		animTitle := "nil"
		if msg.ReplyToMessage.Sticker != nil {
			fileId = msg.ReplyToMessage.Sticker.FileId
		} else {
			msg.ReplyText("Please reply to a sticker for me to kang.")
		}

		if msg.ReplyToMessage.Sticker != nil && msg.ReplyToMessage.Sticker.Emoji != "nil" {
			stickerEmoji = msg.ReplyToMessage.Sticker.Emoji
		} else {
			stickerEmoji = "🛡"
		}

		if msg.ReplyToMessage.Sticker.IsAnimated == true {
			inputFile, r, err := getInputFile(bot, fileId, "sticker.tgs")
			if r != nil {
				defer r.Body.Close()
			}
			if err != nil {
				return err
			}
			success, err = bot.AddTgsStickerToSet(user.Id, packname, inputFile, stickerEmoji)
			animTitle = "%v's animated pack %v"
		} else {
			inputFile, r, err := getInputFile(bot, fileId, "sticker.png")
			if r != nil {
				defer r.Body.Close()
			}
			if err != nil {
				return err
			}
			success, err = bot.AddPngStickerToSet(user.Id, packname, inputFile, stickerEmoji)
		}

		if err != nil {
			err := makeInternal(msg, user, fileId, stickerEmoji, bot, packname, packnum, animTitle)
			if err != nil {
				msg.ReplyText("Something went wrong with kanging.")
				return err
			}
		}

		if success {
			msg.ReplyMarkdownf("Sticker successfully added to [pack](t.me/addstickers/%v)\nEmoji is: %v", packname, stickerEmoji)
		}
	} else {
		msg.ReplyText("What even fam.")
	}
	return nil
}

func makeInternal(msg *ext.Message, user *ext.User, fileId string, emoji string, bot ext.Bot, packname string, packnum int, animTitle string) error {
	name := user.FirstName
	extra_version := ""
	title := "%v's pack %v"
	if packnum > 0 {
		extra_version = " " + strconv.Itoa(packnum)
	}

	if animTitle != "nil" {
		title = animTitle
	}
	newStick := bot.NewSendableCreateNewStickerSet(user.Id, packname, fmt.Sprintf(title, name, extra_version), emoji)
	if animTitle != "nil" {
		inputFile, r, err := getInputFile(bot, fileId, "sticker.tgs")
		if r != nil {
			defer r.Body.Close()
		}
		if err != nil {
			return err
		}
		newStick.TgsSticker = &inputFile
	} else {
		inputFile, r, err := getInputFile(bot, fileId, "sticker.png")
		if r != nil {
			defer r.Body.Close()
		}
		if err != nil {
			return err
		}
		newStick.PngSticker = &inputFile
	}

	success, err := newStick.Send()

	if err != nil {
		bot.Logger.Warnw("No sticker file.", zap.Error(err))
		return err
	}

	if success == true {
		msg.ReplyHTMLf(fmt.Sprintf("Successfully created pack with name: %v. Get it <a href=\"t.me/addstickers/%v\">here</a>", packname, packname))
	}

	return nil
}

func getInputFile(bot ext.Bot, fileId string, fileName string) (ext.InputFile, *http.Response, error) {
	file, err := bot.GetFile(fileId)
	var inputFile ext.InputFile
	var r *http.Response
	if err != nil {
		print("Cannot get the file!")
		return inputFile, r, err
	}

	resp, err := http.Get(fmt.Sprintf("https://api.telegram.org/file/bot%v/%v", harukax.BotConfig.ApiKey, file.FilePath))

	if err != nil {
		return inputFile, r, err
	}

	inputFile = bot.NewFileReader(fileName, io.Reader(resp.Body))
	return inputFile, resp, nil
}

func LoadStickers(u *gotgbot.Updater) {
	defer log.Println("Loading module stickers")
	u.Dispatcher.AddHandler(handlers.NewPrefixCommand("stickerid", harukax.BotConfig.Prefix, stickerId))
	u.Dispatcher.AddHandler(handlers.NewPrefixCommand("getsticker", harukax.BotConfig.Prefix, getSticker))
	u.Dispatcher.AddHandler(handlers.NewPrefixCommand("kang", harukax.BotConfig.Prefix, kangSticker))
}
