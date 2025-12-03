package s3

// ErrorCode represents an S3 error code
type ErrorCode string

const (
	AccessDenied        ErrorCode = "AccessDenied"
	BucketAlreadyExists ErrorCode = "BucketAlreadyExists"
	BucketNotEmpty      ErrorCode = "BucketNotEmpty"
	EntityTooLarge      ErrorCode = "EntityTooLarge"
	EntityTooSmall      ErrorCode = "EntityTooSmall"
	InternalError       ErrorCode = "InternalError"
	InvalidArgument     ErrorCode = "InvalidArgument"
	InvalidBucketName   ErrorCode = "InvalidBucketName"
	InvalidPart         ErrorCode = "InvalidPart"
	InvalidPartOrder    ErrorCode = "InvalidPartOrder"
	InvalidRange        ErrorCode = "InvalidRange"
	NoSuchBucket        ErrorCode = "NoSuchBucket"
	NoSuchKey           ErrorCode = "NoSuchKey"
	NoSuchUpload        ErrorCode = "NoSuchUpload"
	NoSuchVersion       ErrorCode = "NoSuchVersion"
	PreconditionFailed  ErrorCode = "PreconditionFailed"
)

// ErrorResponse represents an S3 error response
type ErrorResponse struct {
	Code      ErrorCode `xml:"Code"`
	Message   string    `xml:"Message"`
	Resource  string    `xml:"Resource"`
	RequestID string    `xml:"RequestId"`
}
