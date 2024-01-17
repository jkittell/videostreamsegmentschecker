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

func check(db database.PosgresDB[*Result], stream Job, jobDone chan bool) {
	var total int
	var ok int

	for _, seg := range stream.Segments {
		total++
		resp, _ := client.Head(seg.SegmentURL)
		if resp != nil {
			if resp.StatusCode > 200 || resp.StatusCode < 300 {
				if resp.ContentLength > 0 {
					ok++
				}
			}
		}
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
	_, err := db.Create(context.TODO(), &result)
	if err != nil {
		log.Println(err)
	}
	jobDone <- true
}

func checkSegments(streamsToCheck chan Job, jobDone chan bool) {
	db, err := database.NewPostgresDB[*Result]("segment_checks", func() *Result {
		return &Result{}
	})
	failOnError(err, "unable to connect to postgres")

	for stream := range streamsToCheck {
		check(db, stream, jobDone)
	}
}
