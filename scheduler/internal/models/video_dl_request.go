package models

import (
	"fmt"
	"time"

	"github.com/go-redsync/redsync"
	"github.com/jmoiron/sqlx"
)

type VideoDLRequest struct {
	Redsync     *redsync.Redsync
	Db          *sqlx.DB
	VideoID     string // Foreign ID
	ID          int    // Domestic ID
	DownloaddID int
	URL         string
	ParentURL   string
}

func (v *VideoDLRequest) SetDownloaded() error {
	website, err := GetWebsiteFromURL(v.URL)
	if err != nil {
		return err
	}

	sql := "UPDATE videos SET download_id = $1, upload_time =  Now() WHERE video_ID = $2 AND id IN (select videos.id FROM videos " +
		"INNER JOIN downloads_to_videos ON videos.id = downloads_to_videos.video_id INNER JOIN downloads ON downloads_to_videos.download_id = downloads.id " +
		"WHERE videos.website = $3)"
	_, err = v.Db.Exec(sql, v.DownloaddID, v.ID, website)
	if err != nil {
		return err
	}

	return v.SetDownloadSucceeded()
}

func (v *VideoDLRequest) SetDownloadSucceeded() error {
	sql := "UPDATE videos SET dlStatus = 1 WHERE id = $1"
	_, err := v.Db.Exec(sql, v.ID)
	return err
}

func (v *VideoDLRequest) SetDownloadFailed() error {
	sql := "UPDATE videos SET dlStatus = 2 WHERE id = $1"
	_, err := v.Db.Exec(sql, v.ID)
	return err
}

func (v *VideoDLRequest) AcquireLockForVideo() error {
	mut := v.Redsync.NewMutex(v.VideoID, redsync.SetExpiry(time.Minute*30))
	return mut.Lock()
}

type event string

const (
	Scheduled  event = "Video %s from %s has been scheduled for download"
	Error      event = "Video %s from %s could not be downloaded, failed with an error. "
	Downloaded event = "Video %s from %s has been downloaded successfully, and uploaded to videoservice"
)

func (v *VideoDLRequest) RecordEvent(inpEvent event, additionalErrorMsg string) error {
	website, err := GetWebsiteFromURL(v.ParentURL)
	if err != nil {
		return err
	}

	formattedMsg := fmt.Sprintf(string(inpEvent), v.VideoID, website)

	if additionalErrorMsg != "" {
		formattedMsg += fmt.Sprintf("\n\nError log: %s", additionalErrorMsg)
	}

	sql := "insert into archival_events (video_url, download_id, parent_url, event_message, event_time) VALUES ($1, $2, $3, $4, Now())"
	_, err = v.Db.Exec(sql, v.URL, v.DownloaddID, v.ParentURL, formattedMsg)
	return err
}
