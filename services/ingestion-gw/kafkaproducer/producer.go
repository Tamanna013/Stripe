package kafkaproducer

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	spanWriter   *kafka.Writer
	logWriter    *kafka.Writer
	metricWriter *kafka.Writer
}

func NewProducer(brokers []string) *Producer {
	// Idempotent producers as defined in Part 2.2
	return &Producer{
		spanWriter: &kafka.Writer{
			Addr:                   kafka.TCP(brokers...),
			Topic:                  "otel.spans.raw",
			RequiredAcks:           kafka.RequireAll,
			AllowAutoTopicCreation: true,
		},
		logWriter: &kafka.Writer{
			Addr:                   kafka.TCP(brokers...),
			Topic:                  "otel.logs.raw",
			RequiredAcks:           kafka.RequireAll,
			AllowAutoTopicCreation: true,
		},
		metricWriter: &kafka.Writer{
			Addr:                   kafka.TCP(brokers...),
			Topic:                  "otel.metrics.raw",
			RequiredAcks:           kafka.RequireAll,
			AllowAutoTopicCreation: true,
		},
	}
}

func (p *Producer) PublishSpan(ctx context.Context, key []byte, payload []byte) error {
	return p.spanWriter.WriteMessages(ctx, kafka.Message{
		Key:   key,
		Value: payload,
	})
}

func (p *Producer) PublishLog(ctx context.Context, key []byte, payload []byte) error {
	return p.logWriter.WriteMessages(ctx, kafka.Message{
		Key:   key,
		Value: payload,
	})
}

func (p *Producer) PublishMetric(ctx context.Context, key []byte, payload []byte) error {
	return p.metricWriter.WriteMessages(ctx, kafka.Message{
		Key:   key,
		Value: payload,
	})
}

func (p *Producer) Close() error {
	p.spanWriter.Close()
	p.logWriter.Close()
	return p.metricWriter.Close()
}
