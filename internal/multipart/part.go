package multipart

// Part represents a part of a multipart upload
type Part struct {
	PartNumber int    `json:"part_number"`
	ETag       string `json:"etag"`
	Size       int64  `json:"size"`
	Checksum   string `json:"checksum"`
}
