package main

import (
	"context"
	"github.com/google/uuid"
	"github.com/jkittell/data/api/client"
	"github.com/jkittell/data/database"
	"log"
	"time"
)

type Result struct {
	Id            uuid.UUID
	URL           string
	OKSegments    int
	TotalSegments int
	OKPercent     float64
	CreatedAt     time.Time
}

func (r *Result) Primary() (string, any) {
	return "id", r.Id
}

func (r *Result) Scan(fields []string, scan database.ScanFunc) error {
	return database.Scan(map[string]any{
		"id":             &r.Id,
		"url":            &r.URL,
		"total_segments": &r.TotalSegments,
		"ok_segments":    &r.OKSegments,
		"ok_percent":     &r.OKPercent,
		"created_at":     &r.CreatedAt,
	}, fields, scan)
}

func (r *Result) Params() map[string]any {
	return map[string]any{
		"id":             r.Id,
		"url":            r.URL,
		"total_segments": &r.TotalSegments,
		"ok_segments":    &r.OKSegments,
		"ok_percent":     &r.OKPercent,
		"created_at":     &r.CreatedAt,
	}
}

type SegmentInfo struct {
	URL           string    `bson:"URL"`
	ABRStreamURL  string    `bson:"abr_stream_url" json:"abr_stream_url"`
	StatusCode    int       `bson:"status_code" json:"status_code"`
	ContentLength int64     `bson:"content_length" json:"content_length"`
	Error         string    `bson:"error" json:"error"`
	CreatedAt     time.Time `bson:"created_at" json:"created_at"`
}

type StreamInfo struct {
	Id       uuid.UUID     `bson:"_id" json:"id"`
	URL      string        `bson:"URL" json:"URL"`
	Segments []SegmentInfo `bson:"segments" json:"segments"`
}

func check(resultsDB database.PosgresDB[*Result], infoDB database.MongoDB[StreamInfo], stream Job, jobDone chan bool) {
	var total int
	var ok int

	streamInfo := StreamInfo{
		Id:       stream.Id,
		URL:      stream.URL,
		Segments: []SegmentInfo{},
	}

	for _, seg := range stream.Segments {
		total++

		segmentInfo := SegmentInfo{
			URL:           seg.SegmentURL,
			ABRStreamURL:  seg.StreamURL,
			StatusCode:    0,
			ContentLength: 0,
			Error:         "",
			CreatedAt:     time.Now(),
		}

		resp, err := client.Head(seg.SegmentURL)
		if err != nil {
			segmentInfo.Error = err.Error()
		}
		if resp != nil {
			segmentInfo.StatusCode = resp.StatusCode
			segmentInfo.ContentLength = resp.ContentLength
			if resp.StatusCode > 200 || resp.StatusCode < 300 {
				if resp.ContentLength > 0 {
					ok++
				}
			}
		}
		streamInfo.Segments = append(streamInfo.Segments, segmentInfo)
	}

	var okPercent float64
	if total > 0 {
		okPercent = (float64(ok) / float64(total)) * 100
	}

	result := Result{
		Id:            stream.Id,
		URL:           stream.URL,
		TotalSegments: total,
		OKSegments:    ok,
		OKPercent:     okPercent,
		CreatedAt:     time.Now(),
	}

	// write to database
	_, err := resultsDB.Create(context.TODO(), &result)
	if err != nil {
		log.Printf("[ %s ] error creating database entry: %s", stream.Id, err)
	}

	// save segment info to database
	err = infoDB.Insert(context.TODO(), "segment_check_info", streamInfo)
	if err != nil {
		log.Printf("[ %s ] error inserting segment check info: %s", stream.Id, err)
	}
	jobDone <- true
}

func checkSegments(streamsToCheck chan Job, jobDone chan bool) {
	resultsDB, err := database.NewPostgresDB[*Result]("segment_checks", func() *Result {
		return &Result{}
	})
	failOnError(err, "unable to connect to postgres")

	infoDB, err := database.NewMongoDB[StreamInfo]()
	failOnError(err, "unable to connect to mongo")

	for stream := range streamsToCheck {
		check(resultsDB, infoDB, stream, jobDone)
	}
}
