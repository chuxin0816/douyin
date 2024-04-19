package kafka

import (
	"context"
	"douyin/dal"
	"douyin/dal/model"
	"douyin/pkg/tracing"
	"encoding/json"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel/codes"
)

type videoMQ struct {
	*mq
}

var videoMQInstance *videoMQ

func initVideoMQ() {
	videoMQInstance = &videoMQ{
		mq: &mq{
			Topic:  topicVideo,
			Writer: NewWriter(topicVideo),
			Reader: NewReader(topicVideo),
		},
	}

	go videoMQInstance.consumeVideo(context.Background())
}

func (mq *videoMQ) consumeVideo(ctx context.Context) {
	for {
		m, err := mq.Reader.ReadMessage(ctx)
		if err != nil {
			klog.Error("failed to read message: ", err)
			break
		}
		video := &model.Video{}
		if err := json.Unmarshal(m.Value, video); err != nil {
			klog.Error("failed to unmarshal message: ", err)
			continue
		}
		if err := dal.UpdateVideo(ctx, video); err != nil {
			klog.Error("failed to update video: ", err)
			continue
		}
	}

}

func UpdateVideo(ctx context.Context, video *model.Video) error {
	ctx, span := tracing.Tracer.Start(ctx, "kafka.UpdateVideo")
	defer span.End()

	value, err := json.Marshal(video)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to marshal message")
		klog.Error("failed to marshal message:", err)
		return err
	}

	return videoMQInstance.Writer.WriteMessages(ctx, kafka.Message{
		Value: value,
	})
}
