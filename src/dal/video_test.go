package dal

import (
	"context"
	"testing"
	"time"
)

var id int64 = 12182603931062272

func TestGetFeedList(t *testing.T) {
	_, _, err := GetFeedList(context.Background(), &id, time.Now(), 10)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
}

func TestSaveVideo(t *testing.T) {
	err := SaveVideo(context.Background(), id, "test.mp4", "test.jpg", "test")
	if err != nil {
		t.Log(err)
		t.Fail()
	}
}

func TestGetPublishList(t *testing.T) {
	_, err := GetPublishList(context.Background(), &id, 12182603931062272)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
}

func TestGetVideoList(t *testing.T) {
	_, err := GetVideoList(context.Background(), &id, []int64{10760595804524544, 12795765776715776})
	if err != nil {
		t.Log(err)
		t.Fail()
	}
}
