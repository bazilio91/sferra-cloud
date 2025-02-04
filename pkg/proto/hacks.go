package proto

import "context"

func file_models_proto_init() {

}

//
//func file_proto_models_proto_init() {
//
//}

func file_proto_data_proto_init() {

}

func (d *DataRecognitionTaskORM) AfterToPB(ctx context.Context, task *DataRecognitionTask) error {
	node := d.RecognitionResult.Data()
	task.RecognitionResult = &node

	return nil
}
