package main

import (
	"encoding/json"
	"fmt"

	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/extractors"
	"github.com/iawia002/annie/extractors/types"
	"github.com/urfave/cli/v2"
)

func Download(c *cli.Context, videoURL string) error {
	data, err := extractors.Extract(videoURL, types.Options{
		Playlist:         c.Bool("playlist"),
		Items:            c.String("items"),
		ItemStart:        int(c.Uint("start")),
		ItemEnd:          int(c.Uint("end")),
		ThreadNumber:     int(c.Uint("thread")),
		EpisodeTitleOnly: c.Bool("episode-title-only"),
		Cookie:           c.String("cookie"),
		YoukuCcode:       c.String("youku-ccode"),
		YoukuCkey:        c.String("youku-ckey"),
		YoukuPassword:    c.String("youku-password"),
	})
	if err != nil {
		// if this error occurs, it means that an error occurred before actually starting to extract data
		// (there is an error in the preparation step), and the data list is empty.
		return err
	}

	if c.Bool("json") {
		jsonData, err := json.MarshalIndent(data, "", "\t")
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", jsonData)
		return nil
	}

	defaultDownloader := downloader.New(downloader.Options{
		InfoOnly:   c.Bool("info"),
		Stream:     c.String("stream-format"),
		Refer:      c.String("refer"),
		OutputPath: DestFolder,
		// TODO: Avoid the dirty approach (looks so stupid now)
		OutputName:     c.String("output-name"),
		FileNameLength: int(c.Uint("file-name-length")),
		Caption:        c.Bool("caption"),
		MultiThread:    c.Bool("multi-thread"),
		ThreadNumber:   int(c.Uint("thread")),
		RetryTimes:     int(c.Uint("retry")),
		ChunkSizeMB:    int(c.Uint("chunk-size")),
		UseAria2RPC:    c.Bool("aria2"),
		Aria2Token:     c.String("aria2-token"),
		Aria2Method:    c.String("aria2-method"),
		Aria2Addr:      c.String("aria2-addr"),
	})
	errors := make([]error, 0)
	for _, item := range data {
		if item.Err != nil {
			// if this error occurs, the preparation step is normal, but the data extraction is wrong.
			// the data is an empty struct.
			errors = append(errors, item.Err)
			continue
		}
		if err = defaultDownloader.Download(item); err != nil {
			errors = append(errors, err)
		}
	}
	if len(errors) != 0 {
		return errors[0]
	}
	return nil
}
