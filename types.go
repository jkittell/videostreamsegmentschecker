package main

import (
	"github.com/google/uuid"
)

type Job struct {
	Id       uuid.UUID
	URL      string
	Segments []Segment
}

type Segment struct {
	PlaylistURL    string `json:"playlist_url"`
	StreamName     string `json:"stream_name"`
	StreamURL      string `json:"stream_url"`
	SegmentName    string `json:"segment_name"`
	SegmentURL     string `json:"segment_url"`
	ByteRangeStart int    `json:"byte_range_start"`
	ByteRangeSize  int    `json:"byte_range_size"`
}
