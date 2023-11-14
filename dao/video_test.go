package dao

import (
	"testing"
	"time"
)

func TestGetFeedList(t *testing.T) {
	_, _, err := GetFeedList(12182603931062272, time.Now(), 10)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
}

func TestSaveVideo(t *testing.T) {
	err := SaveVideo(12182603931062272, "test.mp4", "test.jpg", "test")
	if err != nil {
		t.Log(err)
		t.Fail()
	}
}

func TestGetPublishList(t *testing.T) {
	_, err := GetPublishList(12182603931062272, 12182603931062272)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
}

func TestGetVideoList(t *testing.T) {
	_, err := GetVideoList(12182603931062272, []int64{10760595804524544, 12795765776715776})
	if err != nil {
		t.Log(err)
		t.Fail()
	}
}
