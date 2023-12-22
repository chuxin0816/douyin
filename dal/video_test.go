package dal

import (
	"testing"
	"time"
)

var id int64 = 12182603931062272

func TestGetFeedList(t *testing.T) {
	_, _, err := GetFeedList(&id, time.Now(), 10)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
}

func TestSaveVideo(t *testing.T) {
	err := SaveVideo(id, "test.mp4", "test.jpg", "test")
	if err != nil {
		t.Log(err)
		t.Fail()
	}
}

func TestGetPublishList(t *testing.T) {
	_, err := GetPublishList(&id, 12182603931062272)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
}

func TestGetVideoList(t *testing.T) {
	_, err := GetVideoList(&id, []int64{10760595804524544, 12795765776715776})
	if err != nil {
		t.Log(err)
		t.Fail()
	}
}
