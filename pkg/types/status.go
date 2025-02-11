package types

import "github.com/bazilio91/sferra-cloud/pkg/proto"

// IsTerminalState checks if the given status is a terminal state
func IsTerminalState(status proto.Status) bool {
	switch status {
	case proto.Status_STATUS_PROCESSING_COMPLETED,
		proto.Status_STATUS_IMAGES_FAILED_QUOTA,
		proto.Status_STATUS_IMAGES_FAILED_TIMEOUT,
		proto.Status_STATUS_IMAGES_FAILED_PROCESSING,
		proto.Status_STATUS_RECOGNITION_FAILED_TIMEOUT,
		proto.Status_STATUS_RECOGNITION_FAILED_PROCESSING,
		proto.Status_STATUS_RECOGNITION_FAILED_QUOTA:
		return true
	default:
		return false
	}
}
